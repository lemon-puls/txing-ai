// Package wiki LLM Wiki 知识库 service 层（设计文档 docs/llmwiki_design.md）。
//
// 职责：源管理、异步 ingest 编排、草稿审核流、出入链与 index 维护、面试官问答、md 导出。
// DB 写入全部在本层完成；agent 引擎只通过 iface/wikiagent 契约产出结构化数据（D14）。
package wiki

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/config"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface/wikiagent"
	mytool "txing-ai/internal/tool"
	"txing-ai/internal/utils"
	wikibuiltin "txing-ai/internal/wiki/agent/builtin"
)

// 问答限流默认值（config wiki 段缺省时生效）
const (
	defaultRateLimitPerIPPerHour = 10
	defaultDailyQuota            = 200
	// askHistoryWindow 面试官多轮对话历史窗口
	askHistoryWindow = 10
	// ingestTimeout 单次 ingest 编译的最长耗时（大文档分片逐片调用 LLM，需放宽）
	ingestTimeout = 30 * time.Minute
	// maxStoredToolResultChars 持久化的单条工具结果最大字符数（对齐 ops）
	maxStoredToolResultChars = 4000
)

// Service 知识库服务（handler 按请求构造，DB/COS 来自 gin context，对齐 ops 范式）
type Service struct {
	db     *gorm.DB
	cos    *utils.COSClient
	engine wikiagent.Port
	qa     wikiagent.QAEngine
}

// NewService 构造服务并按配置装配 agent 引擎
func NewService(db *gorm.DB, cos *utils.COSClient) (*Service, error) {
	engine, err := wikibuiltin.NewFromConfig(db, global.LoadConfig().WikiConfig)
	if err != nil {
		return nil, err
	}
	return &Service{db: db, cos: cos, engine: engine, qa: engine}, nil
}

// --- 源管理（Raw 层） ---

// CreateMDSource 新建 markdown 源（原文经预签名 PUT 上传 COS，只存 key）
func (s *Service) CreateMDSource(ctx context.Context, title, content string) (*domain.WikiSource, error) {
	if s.cos == nil {
		return nil, fmt.Errorf("COS 未配置，无法存储 markdown 原文")
	}
	sum := sha256.Sum256([]byte(content))
	hash := hex.EncodeToString(sum[:])

	key := fmt.Sprintf("wiki/raw/%s/%s.md", time.Now().Format("200601"), hash[:16])
	uploadURL, err := s.cos.GenerateUploadPresignedURL(key)
	if err != nil {
		return nil, fmt.Errorf("生成 COS 上传凭证失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/markdown; charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传原文到 COS 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("上传原文到 COS 失败, status=%d", resp.StatusCode)
	}

	src := &domain.WikiSource{
		Title:       title,
		SourceType:  domain.WikiSourceTypeMD,
		COSKey:      key,
		ContentHash: hash,
		Status:      domain.WikiSourceStatusPending,
	}
	if err := s.db.WithContext(ctx).Create(src).Error; err != nil {
		return nil, err
	}
	return src, nil
}

// CreateURLSource 新建网页 URL 源（ingest 时抓取正文）
func (s *Service) CreateURLSource(ctx context.Context, title, url string) (*domain.WikiSource, error) {
	src := &domain.WikiSource{
		Title:      title,
		SourceType: domain.WikiSourceTypeURL,
		URL:        url,
		Status:     domain.WikiSourceStatusPending,
	}
	if err := s.db.WithContext(ctx).Create(src).Error; err != nil {
		return nil, err
	}
	return src, nil
}

// ListSources 分页列出源
func (s *Service) ListSources(ctx context.Context, page, pageSize int) ([]domain.WikiSource, int64, error) {
	var (
		list  []domain.WikiSource
		total int64
	)
	if err := s.db.WithContext(ctx).Model(&domain.WikiSource{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := s.db.WithContext(ctx).Order("update_time DESC").
		Scopes(config.Paginate(page, pageSize)).
		Find(&list).Error
	return list, total, err
}

// DeleteSource 删除源（软删；COS 原文保留作审计）
func (s *Service) DeleteSource(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&domain.WikiSource{}, id).Error
}

// --- Ingest 编排（异步） ---

// TriggerIngest 触发一次源的异步编译（ingesting → ingested/failed）
func (s *Service) TriggerIngest(sourceID int64) error {
	var src domain.WikiSource
	if err := s.db.First(&src, sourceID).Error; err != nil {
		return fmt.Errorf("源不存在: %w", err)
	}
	if src.Status == domain.WikiSourceStatusIngesting {
		return fmt.Errorf("该源正在编译中")
	}

	// 先置状态再起 goroutine（重复触发防御）
	if err := s.db.Model(&src).Update("status", domain.WikiSourceStatusIngesting).Error; err != nil {
		return err
	}

	// 后台执行：不继承请求 ctx（handler 返回后 gin ctx 会被取消）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("wiki ingest panic", zap.Int64("sourceId", src.Id), zap.Any("panic", r))
				s.markSourceFailed(src.Id, fmt.Sprintf("panic: %v", r))
			}
		}()
		s.runIngest(src)
	}()
	return nil
}

// runIngest 同步执行一次编译（取正文 → 引擎编译 → 草稿落库）
func (s *Service) runIngest(src domain.WikiSource) {
	ctx, cancel := context.WithTimeout(context.Background(), ingestTimeout)
	defer cancel()

	fail := func(msg string, err error) {
		log.Error("wiki ingest failed", zap.Int64("sourceId", src.Id), zap.Error(err))
		s.markSourceFailed(src.Id, msg+": "+err.Error())
	}

	// 1. 取正文
	content, err := s.loadSourceContent(ctx, &src)
	if err != nil {
		fail("获取源内容失败", err)
		return
	}

	// URL 源的正文可能变化：刷新 hash
	if src.SourceType == domain.WikiSourceTypeURL {
		sum := sha256.Sum256([]byte(content))
		hash := hex.EncodeToString(sum[:])
		if hash != src.ContentHash {
			if uerr := s.db.Model(&domain.WikiSource{}).Where("id = ?", src.Id).
				Update("content_hash", hash).Error; uerr == nil {
				src.ContentHash = hash
			}
		}
	}

	// 2. 现有已发布页概览（引擎据此决定新建/更新）
	briefs, err := s.listPublishedBriefs(ctx)
	if err != nil {
		fail("查询已有页面失败", err)
		return
	}

	// 3. 引擎编译（D14：只产出结构化草稿）
	result, err := s.engine.Ingest(ctx, &wikiagent.IngestTask{
		TaskID:   src.Id,
		Title:    src.Title,
		Kind:     src.SourceType,
		URL:      src.URL,
		Content:  content,
		Existing: briefs,
	})
	if err != nil {
		fail("编译失败", err)
		return
	}
	if len(result.Drafts) == 0 {
		fail("编译未产出任何草稿", fmt.Errorf("empty drafts, notes: %s", result.Notes))
		return
	}

	// 4. 草稿落库（service 层持有全部写权限）
	now := time.Now()
	for _, d := range result.Drafts {
		page := &domain.WikiPage{
			Slug:       d.Slug,
			Title:      d.Title,
			PageType:   normalizePageType(d.PageType),
			Aliases:    d.Aliases,
			Summary:    d.Summary,
			Content:    d.Content,
			Status:     domain.WikiPageStatusDraft,
			Version:    1,
			SourceRefs: []domain.WikiSourceRef{{SourceId: src.Id, Note: d.Changes}},
		}
		if d.TargetSlug != "" {
			// 更新语义：挂到同 slug 的已发布页（P2 的 diff 审核用）
			var target domain.WikiPage
			if err := s.db.Where("slug = ? AND status = ?", d.TargetSlug, domain.WikiPageStatusPublished).
				First(&target).Error; err == nil {
				page.TargetPageId = target.Id
			}
		}
		if err := s.db.Create(page).Error; err != nil {
			fail("草稿落库失败", err)
			return
		}
	}

	if err := s.db.Model(&domain.WikiSource{}).Where("id = ?", src.Id).
		Updates(map[string]interface{}{
			"status":      domain.WikiSourceStatusIngested,
			"error_msg":   "",
			"ingested_at": now,
		}).Error; err != nil {
		log.Error("wiki ingest 状态更新失败", zap.Int64("sourceId", src.Id), zap.Error(err))
	}
	log.Info("wiki ingest 完成", zap.Int64("sourceId", src.Id), zap.Int("drafts", len(result.Drafts)),
		zap.String("notes", result.Notes))
}

// loadSourceContent 按类型取源正文：md 经预签名 URL 从 COS 拉回；url 走网页抓取
func (s *Service) loadSourceContent(ctx context.Context, src *domain.WikiSource) (string, error) {
	switch src.SourceType {
	case domain.WikiSourceTypeMD:
		if s.cos == nil || src.COSKey == "" {
			return "", fmt.Errorf("COS 未配置或源缺少 cos_key")
		}
		downloadURL, err := s.cos.GenerateDownloadPresignedURL(src.COSKey)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
		if err != nil {
			return "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("下载原文失败, status=%d", resp.StatusCode)
		}
		b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		if err != nil {
			return "", err
		}
		return string(b), nil
	case domain.WikiSourceTypeURL:
		resp, err := mytool.ScrapeWebPage(ctx, &mytool.WebScrapingRequest{
			URL:           src.URL,
			MaxTextLength: 24000,
		})
		if err != nil {
			if resp.Content != "" { // 抓取器部分成功时仍返回已得内容
				return resp.Content, nil
			}
			return "", err
		}
		if resp.Truncated {
			log.Warn("wiki URL 源正文被截断", zap.Int64("sourceId", src.Id), zap.String("url", src.URL))
		}
		return resp.Content, nil
	default:
		return "", fmt.Errorf("未知来源类型: %s", src.SourceType)
	}
}

func (s *Service) markSourceFailed(id int64, msg string) {
	s.db.Model(&domain.WikiSource{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    domain.WikiSourceStatusFailed,
			"error_msg": truncateStr(msg, 1000),
		})
}

// listPublishedBriefs 已发布页概览（引擎上下文用）
func (s *Service) listPublishedBriefs(ctx context.Context) ([]wikiagent.PageBrief, error) {
	var pages []domain.WikiPage
	err := s.db.WithContext(ctx).
		Where("status = ?", domain.WikiPageStatusPublished).
		Order("update_time DESC").Limit(200).Find(&pages).Error
	if err != nil {
		return nil, err
	}
	briefs := make([]wikiagent.PageBrief, 0, len(pages))
	for _, p := range pages {
		briefs = append(briefs, wikiagent.PageBrief{
			Slug: p.Slug, Title: p.Title, PageType: p.PageType,
			Summary: p.Summary, Aliases: p.Aliases, Version: p.Version,
		})
	}
	return briefs, nil
}

// --- 草稿审核流 ---

// ListDrafts 分页列出草稿
func (s *Service) ListDrafts(ctx context.Context, page, pageSize int) ([]domain.WikiPage, int64, error) {
	var (
		list  []domain.WikiPage
		total int64
	)
	q := s.db.WithContext(ctx).Model(&domain.WikiPage{}).Where("status = ?", domain.WikiPageStatusDraft)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("update_time DESC").
		Scopes(config.Paginate(page, pageSize)).
		Find(&list).Error
	return list, total, err
}

// UpdateDraft 编辑草稿（审核人可直接改稿）
func (s *Service) UpdateDraft(ctx context.Context, id int64, req *dto.WikiUpdateDraftReq) error {
	var draft domain.WikiPage
	if err := s.db.Where("id = ? AND status = ?", id, domain.WikiPageStatusDraft).First(&draft).Error; err != nil {
		return fmt.Errorf("草稿不存在")
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.PageType != nil {
		updates["page_type"] = normalizePageType(*req.PageType)
	}
	if req.Aliases != nil {
		// json 列经 map 更新时需手动序列化（serializer 只作用于 struct 字段）
		if b, err := json.Marshal(req.Aliases); err == nil {
			updates["aliases"] = string(b)
		}
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&draft).Updates(updates).Error
}

// ConfirmDraft 确认草稿：按 slug upsert 到已发布（同 slug 即更新，version+1），重建出入链与 index
func (s *Service) ConfirmDraft(ctx context.Context, id int64) error {
	var draft domain.WikiPage
	if err := s.db.Where("id = ? AND status = ?", id, domain.WikiPageStatusDraft).First(&draft).Error; err != nil {
		return fmt.Errorf("草稿不存在")
	}

	var published domain.WikiPage
	err := s.db.Where("slug = ? AND status = ?", draft.Slug, domain.WikiPageStatusPublished).First(&published).Error
	isUpdate := err == nil

	target := &draft
	if isUpdate {
		published.Title = draft.Title
		published.PageType = normalizePageType(draft.PageType)
		published.Aliases = draft.Aliases
		published.Summary = draft.Summary
		published.Content = draft.Content
		published.SourceRefs = draft.SourceRefs
		published.Version++
		target = &published
	} else {
		draft.Status = domain.WikiPageStatusPublished
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(target).Error; err != nil {
			return err
		}
		if isUpdate {
			if err := tx.Delete(&draft).Error; err != nil { // 更新语义：草稿已合并，删除
				return err
			}
		}
		return rebuildLinks(tx, target)
	})
	if err != nil {
		return err
	}
	return s.RefreshIndex(ctx)
}

// ConfirmAllDrafts 全部确认（逐条执行，返回成功数与首个错误）
func (s *Service) ConfirmAllDrafts(ctx context.Context) (int, error) {
	var ids []int64
	if err := s.db.Model(&domain.WikiPage{}).
		Where("status = ?", domain.WikiPageStatusDraft).
		Order("update_time ASC").Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	confirmed := 0
	for _, id := range ids {
		if err := s.ConfirmDraft(ctx, id); err != nil {
			return confirmed, err
		}
		confirmed++
	}
	return confirmed, nil
}

// DeleteDraft 丢弃草稿（软删）
func (s *Service) DeleteDraft(ctx context.Context, id int64) error {
	res := s.db.WithContext(ctx).Where("id = ? AND status = ?", id, domain.WikiPageStatusDraft).
		Delete(&domain.WikiPage{})
	if res.RowsAffected == 0 {
		return fmt.Errorf("草稿不存在")
	}
	return res.Error
}

// rebuildLinks 重建某页的出入链记录（[[slug|text]] 解析，悬空允许——lint 检测用）
func rebuildLinks(tx *gorm.DB, page *domain.WikiPage) error {
	if err := tx.Where("from_page_id = ?", page.Id).Delete(&domain.WikiLink{}).Error; err != nil {
		return err
	}
	links := domain.ParseWikiLinks(page.Content)
	if len(links) == 0 {
		return nil
	}
	rows := make([]domain.WikiLink, 0, len(links))
	for _, l := range links {
		rows = append(rows, domain.WikiLink{FromPageId: page.Id, ToSlug: l.Slug, AnchorText: l.Text})
	}
	return tx.Create(&rows).Error
}

// --- 已发布页管理 ---

// ListPublished 分页列出已发布页
func (s *Service) ListPublished(ctx context.Context, page, pageSize int, keyword string) ([]domain.WikiPage, int64, error) {
	var (
		list  []domain.WikiPage
		total int64
	)
	q := s.db.WithContext(ctx).Model(&domain.WikiPage{}).Where("status = ?", domain.WikiPageStatusPublished)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR slug LIKE ? OR summary LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("update_time DESC").
		Scopes(config.Paginate(page, pageSize)).
		Find(&list).Error
	return list, total, err
}

// GetPublished 取单个已发布页
func (s *Service) GetPublished(ctx context.Context, id int64) (*domain.WikiPage, error) {
	var page domain.WikiPage
	if err := s.db.Where("id = ? AND status = ?", id, domain.WikiPageStatusPublished).First(&page).Error; err != nil {
		return nil, fmt.Errorf("页面不存在")
	}
	return &page, nil
}

// UpdatePublished 管理员直接编辑已发布页（version+1，重建链与 index）
func (s *Service) UpdatePublished(ctx context.Context, id int64, req *dto.WikiUpdateDraftReq) error {
	var page domain.WikiPage
	if err := s.db.Where("id = ? AND status = ?", id, domain.WikiPageStatusPublished).First(&page).Error; err != nil {
		return fmt.Errorf("页面不存在")
	}
	if req.Title != nil {
		page.Title = *req.Title
	}
	if req.PageType != nil {
		page.PageType = normalizePageType(*req.PageType)
	}
	if req.Aliases != nil {
		page.Aliases = req.Aliases
	}
	if req.Summary != nil {
		page.Summary = *req.Summary
	}
	if req.Content != nil {
		page.Content = *req.Content
	}
	page.Version++
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&page).Error; err != nil {
			return err
		}
		return rebuildLinks(tx, &page)
	}); err != nil {
		return err
	}
	return s.RefreshIndex(ctx)
}

// OfflinePage 下线已发布页
func (s *Service) OfflinePage(ctx context.Context, id int64) error {
	res := s.db.WithContext(ctx).Model(&domain.WikiPage{}).
		Where("id = ? AND status = ?", id, domain.WikiPageStatusPublished).
		Update("status", domain.WikiPageStatusArchived)
	if res.RowsAffected == 0 {
		return fmt.Errorf("页面不存在或已下线")
	}
	if res.Error != nil {
		return res.Error
	}
	return s.RefreshIndex(ctx)
}

// --- Index 渲染（WeKnora ② 范式） ---

// RefreshIndex 渲染已发布页 index 并写入 wiki_kv 缓存
func (s *Service) RefreshIndex(ctx context.Context) error {
	index, err := s.RenderIndex(ctx)
	if err != nil {
		return err
	}
	var kv domain.WikiKV
	err = s.db.WithContext(ctx).Where("`key` = ?", domain.WikiKVKeyIndexCache).First(&kv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.db.WithContext(ctx).Create(&domain.WikiKV{Key: domain.WikiKVKeyIndexCache, Value: index}).Error
		}
		return err
	}
	kv.Value = index
	return s.db.WithContext(ctx).Save(&kv).Error
}

// RenderIndex 按类型分组渲染 index（路由目录，非证据）
func (s *Service) RenderIndex(ctx context.Context) (string, error) {
	var pages []domain.WikiPage
	err := s.db.WithContext(ctx).
		Where("status = ?", domain.WikiPageStatusPublished).
		Order("page_type ASC, title ASC").Find(&pages).Error
	if err != nil {
		return "", err
	}
	if len(pages) == 0 {
		return "(the knowledge base is currently empty)", nil
	}
	var b strings.Builder
	lastType := ""
	for _, p := range pages {
		if p.PageType != lastType {
			b.WriteString("\n### " + p.PageType + "\n")
			lastType = p.PageType
		}
		b.WriteString(fmt.Sprintf("- [[%s|%s]]", p.Slug, p.Title))
		if p.Summary != "" {
			b.WriteString(" — " + p.Summary)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n(Use wiki_search_pages when the topic is not obviously covered above.)\n")
	return b.String(), nil
}

// GetIndexForQA 问答用 index：优先 wiki_kv 缓存，缺失时即时渲染
func (s *Service) GetIndexForQA(ctx context.Context) string {
	var kv domain.WikiKV
	if err := s.db.WithContext(ctx).Where("`key` = ?", domain.WikiKVKeyIndexCache).First(&kv).Error; err == nil && kv.Value != "" {
		return kv.Value
	}
	index, err := s.RenderIndex(ctx)
	if err != nil {
		log.Error("wiki index 渲染失败", zap.Error(err))
		return "(the knowledge base is currently empty)"
	}
	return index
}

// --- 面试官问答 ---

// AskEnabled 功能总开关
func AskEnabled() bool {
	cfg := global.LoadConfig().WikiConfig
	return cfg != nil && cfg.Enabled
}

// CheckAskLimit 问答限流：单 IP 每小时 + 全站每日（IP 只作计数，不落明文）
func CheckAskLimit(ctx context.Context, rdb *redis.Client, ip string) error {
	if rdb == nil {
		return nil // redis 不可用时放行（可用性优先，防刷靠部署层）
	}
	perIP := defaultRateLimitPerIPPerHour
	quota := defaultDailyQuota
	if cfg := global.LoadConfig().WikiConfig; cfg != nil {
		if cfg.RateLimitPerIPPerHour > 0 {
			perIP = cfg.RateLimitPerIPPerHour
		}
		if cfg.DailyQuota > 0 {
			quota = cfg.DailyQuota
		}
	}

	// 全站日配额
	dayKey := fmt.Sprintf("wiki:ask:quota:%s", time.Now().Format("20060102"))
	count, err := rdb.Incr(ctx, dayKey).Result()
	if err != nil {
		log.Error("wiki ask 日配额计数失败", zap.Error(err))
	} else {
		if count == 1 {
			rdb.Expire(ctx, dayKey, 25*time.Hour)
		}
		if count > int64(quota) {
			return fmt.Errorf("今日提问额度已用完，请明天再来")
		}
	}

	// 单 IP 小时限流
	ipKey := fmt.Sprintf("wiki:ask:ip:%s:%s", ip, time.Now().Format("2006010215"))
	ipCount, err := rdb.Incr(ctx, ipKey).Result()
	if err != nil {
		log.Error("wiki ask IP 计数失败", zap.Error(err))
		return nil
	}
	if ipCount == 1 {
		rdb.Expire(ctx, ipKey, time.Hour)
	}
	if ipCount > int64(perIP) {
		return fmt.Errorf("提问太频繁，请稍后再试")
	}
	return nil
}

// Ask 面试官问答（SSE）：index 注入 + 引擎流式回答 + 完整问答对落库（D10）
func (s *Service) Ask(ctx context.Context, req *dto.WikiAskReq, ip string, onChunk func(chunk *global.Chunk) error) error {
	// 历史窗口（从最新往前截断）
	history := req.History
	if len(history) > askHistoryWindow {
		history = history[len(history)-askHistoryWindow:]
	}

	// IP 哈希（不留明文，D10）
	ipSum := sha256.Sum256([]byte("wiki-ask:" + ip))
	ipHash := hex.EncodeToString(ipSum[:8])

	qaLog := &domain.WikiQALog{
		SessionId: req.SessionId,
		Question:  req.Question,
		IpHash:    ipHash,
		Status:    "running",
	}
	if err := s.db.Create(qaLog).Error; err != nil {
		log.Error("wiki qa log 创建失败", zap.Error(err))
	}

	start := time.Now()
	traces := make([]domain.WikiToolTrace, 0, 4)
	wrapped := func(chunk *global.Chunk) error {
		if chunk.ToolCallId != "" && chunk.ToolStatus != "running" {
			result := chunk.ToolResult
			if len([]rune(result)) > maxStoredToolResultChars {
				result = truncateStr(result, maxStoredToolResultChars) + "…（已截断）"
			}
			traces = append(traces, domain.WikiToolTrace{
				Id: chunk.ToolCallId, Name: chunk.ToolName,
				Params: chunk.ToolParams, Result: result,
				Status: chunk.ToolStatus,
			})
		}
		if onChunk != nil {
			return onChunk(chunk)
		}
		return nil
	}

	answer, err := s.qa.Answer(ctx, &wikiagent.QARequest{
		Question: req.Question,
		History:  history,
		Index:    s.GetIndexForQA(ctx),
	}, wrapped)

	// 问答留存（失败也记录，便于发现恶意使用与模型问题）
	updates := map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}
	if b, tErr := json.Marshal(traces); tErr == nil {
		updates["tool_traces"] = string(b)
	}
	if err != nil {
		updates["status"] = "failed"
		updates["error_msg"] = truncateStr(err.Error(), 1000)
	} else {
		updates["status"] = "completed"
		updates["answer_md"] = answer
	}
	if uerr := s.db.Model(&domain.WikiQALog{}).Where("id = ?", qaLog.Id).Updates(updates).Error; uerr != nil {
		log.Error("wiki qa log 更新失败", zap.Int64("id", qaLog.Id), zap.Error(uerr))
	}
	return err
}

// --- 导出（md 资产退出保险） ---

// ExportAll 打包全部已发布页为 md 压缩包（含 index.md，Obsidian 可直接打开）
func (s *Service) ExportAll(ctx context.Context) ([]byte, string, error) {
	var pages []domain.WikiPage
	if err := s.db.WithContext(ctx).
		Where("status = ?", domain.WikiPageStatusPublished).
		Order("page_type ASC, title ASC").Find(&pages).Error; err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// index.md
	if idx, err := s.RenderIndex(ctx); err == nil {
		if w, err := zw.Create("index.md"); err == nil {
			_, _ = w.Write([]byte("# Wiki Index\n" + idx))
		}
	}
	for _, p := range pages {
		w, err := zw.Create(p.Slug + ".md")
		if err != nil {
			continue
		}
		header := fmt.Sprintf("---\ntitle: %s\ntype: %s\nslug: %s\nversion: %d\n---\n\n",
			p.Title, p.PageType, p.Slug, p.Version)
		_, _ = w.Write([]byte(header + p.Content))
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("wiki-export-%s.zip", time.Now().Format("20060102")), nil
}

// --- 小工具 ---

// normalizePageType 页面类型归一（默认 concept）
func normalizePageType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if t == "" {
		return "concept"
	}
	return t
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
