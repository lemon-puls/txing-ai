// Package wiki LLM Wiki 知识库的 agent 工具集（按任务构造，非全局单例，对齐 ops 范式）。
//
// 两组工具：
//   - QA 工具（只读）：wiki_read_page / wiki_search_pages —— 面试官问答的 agentic retrieval
//   - Ingest 工具（内存暂存）：wiki_propose_page —— 只向 collector 累积结构化草稿，
//     不写库（D14：契约是数据，落库由 service 层完成）
package wiki

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"

	"txing-ai/internal/domain"
	mytool "txing-ai/internal/tool"
)

// 工具名称
const (
	ReadPageToolName    = "wiki_read_page"
	SearchPagesToolName = "wiki_search_pages"
	ProposePageToolName = "wiki_propose_page"
)

// readPageMaxRunes read_page 单次返回正文的字符预算（WeKnora ④：输出预算防上下文膨胀）
const readPageMaxRunes = 6000

// searchPageLimit search_pages 返回的最大页面数
const searchPageLimit = 8

// searchSnippetRunes 命中片段的展示长度
const searchSnippetRunes = 160

// publishedScope 裸表查询时手工对齐模型软删除与发布态过滤
func publishedScope() (string, []interface{}) {
	return "wiki_pages.status = ? AND wiki_pages.delete_time IS NULL", []interface{}{domain.WikiPageStatusPublished}
}

// --- QA 工具（只读） ---

// QAToolDeps QA 工具依赖（按请求构造）
type QAToolDeps struct {
	DB *gorm.DB
}

// ProvideQATools 构建问答只读工具集
func ProvideQATools(deps QAToolDeps) []tool.BaseTool {
	readPage, err := toolutils.InferTool(
		ReadPageToolName,
		"Read one wiki page by its slug. Returns the full markdown content plus its outbound and "+
			"inbound link summaries. Use it to verify facts before answering; index entries are routing "+
			"hints, not evidence.",
		deps.readPage)
	if err != nil {
		panic(err)
	}

	searchPages, err := toolutils.InferTool(
		SearchPagesToolName,
		"Full-text search over published wiki pages (title, content and aliases). Returns matched "+
			"pages as [[slug|title]] lines with a snippet. If it reports no hits, try different keywords "+
			"once, then honestly tell the user the topic is not covered instead of guessing.",
		deps.searchPages)
	if err != nil {
		panic(err)
	}

	// 复用统一的容错包装：工具报错时返回错误信息给 LLM 而非中断流程
	return mytool.WrapSafeTools([]tool.BaseTool{readPage, searchPages})
}

// readPageParams read_page 入参
type readPageParams struct {
	Slug string `json:"slug" binding:"required"`
}

// readPageResult read_page 出参
type readPageResult struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	PageType string `json:"pageType"`
	Summary  string `json:"summary,omitempty"`
	Content  string `json:"content"`
	// 出链/入链摘要（WeKnora ④：各附一句，帮模型顺藤摸瓜）
	OutboundLinks string `json:"outboundLinks,omitempty"`
	InboundLinks  string `json:"inboundLinks,omitempty"`
	Truncated     bool   `json:"truncated,omitempty"`
}

func (d *QAToolDeps) readPage(ctx context.Context, p *readPageParams) (readPageResult, error) {
	var page domain.WikiPage
	err := d.DB.WithContext(ctx).
		Where("slug = ? AND status = ?", p.Slug, domain.WikiPageStatusPublished).
		First(&page).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return readPageResult{}, fmt.Errorf("no published wiki page with slug %q", p.Slug)
		}
		return readPageResult{}, err
	}

	res := readPageResult{
		Slug:     page.Slug,
		Title:    page.Title,
		PageType: page.PageType,
		Summary:  page.Summary,
	}

	// 正文输出预算：超长截断并提示（WeKnora ④）
	content := page.Content
	if utf8.RuneCountInString(content) > readPageMaxRunes {
		pos := 0
		for i := 0; i < readPageMaxRunes; i++ {
			_, size := utf8.DecodeRuneInString(content[pos:])
			pos += size
		}
		content = content[:pos] + "\n\n…（内容过长已截断，如需更多细节请用 wiki_search_pages 换关键词定位）"
		res.Truncated = true
	}
	res.Content = content

	// 出链：从正文解析 [[slug|text]]（锚文本即摘要）
	if links := domain.ParseWikiLinks(page.Content); len(links) > 0 {
		parts := make([]string, 0, len(links))
		for _, l := range links {
			parts = append(parts, "[["+l.Slug+"|"+l.Text+"]]")
		}
		res.OutboundLinks = "本文引用了: " + strings.Join(parts, "、")
	}

	// 入链：哪些已发布页引用了我
	var inbound []struct {
		Slug  string
		Title string
	}
	cond, args := publishedScope()
	d.DB.WithContext(ctx).Table("wiki_links").
		Select("wiki_pages.slug as slug, wiki_pages.title as title").
		Joins("JOIN wiki_pages ON wiki_pages.id = wiki_links.from_page_id AND "+cond, args...).
		Where("wiki_links.to_slug = ? AND wiki_links.delete_time IS NULL", page.Slug).
		Limit(searchPageLimit).Scan(&inbound)
	if len(inbound) > 0 {
		parts := make([]string, 0, len(inbound))
		for _, in := range inbound {
			parts = append(parts, "[["+in.Slug+"|"+in.Title+"]]")
		}
		res.InboundLinks = "以下页面引用了本文: " + strings.Join(parts, "、")
	}

	return res, nil
}

// searchPagesParams search_pages 入参
type searchPagesParams struct {
	Keyword string `json:"keyword" binding:"required"`
}

// searchPagesResult search_pages 出参
type searchPagesResult struct {
	Keyword string   `json:"keyword"`
	Total   int      `json:"total"`
	Pages   []string `json:"pages"` // "[[slug|title]] (type) — summary | 命中片段: ..."
	Hint    string   `json:"hint,omitempty"`
}

func (d *QAToolDeps) searchPages(ctx context.Context, p *searchPagesParams) (searchPagesResult, error) {
	res := searchPagesResult{Keyword: p.Keyword, Pages: []string{}}

	type row struct {
		Slug     string
		Title    string
		PageType string
		Summary  string
		Content  string
	}

	var rows []row
	cond, args := publishedScope()
	// 优先 ngram FULLTEXT；索引未建（迁移未执行）时降级 LIKE（设计文档 §13 风险预案）
	err := d.DB.WithContext(ctx).Table("wiki_pages").
		Select("slug, title, page_type, summary, content").
		Where(cond, args...).
		Order(clause.Expr{SQL: "MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", Vars: []interface{}{p.Keyword}}).
		Limit(searchPageLimit).Scan(&rows).Error
	if err != nil {
		like := "%" + p.Keyword + "%"
		rows = nil
		d.DB.WithContext(ctx).Table("wiki_pages").
			Select("slug, title, page_type, summary, content").
			Where(cond, args...).
			Where("title LIKE ? OR content LIKE ? OR summary LIKE ? OR aliases LIKE ?",
				like, like, like, like).
			Limit(searchPageLimit).Scan(&rows)
	}

	for _, r := range rows {
		line := fmt.Sprintf("[[%s|%s]] (%s)", r.Slug, r.Title, r.PageType)
		if r.Summary != "" {
			line += " — " + r.Summary
		}
		if snippet := hitSnippet(r.Content, p.Keyword); snippet != "" {
			line += " | 命中片段: " + snippet
		}
		res.Pages = append(res.Pages, line)
	}
	res.Total = len(res.Pages)

	if res.Total == 0 {
		// 空结果引导话术（WeKnora ③）：明确告知并给下一步动作
		res.Hint = "没有命中任何页面。可以：1) 换一个关键词或同义词重试一次；" +
			"2) 如果已知页面 slug，直接用 wiki_read_page 读取；3) 如果仍无结果，" +
			"向用户说明知识库暂未收录该主题，不要编造内容。"
	}
	return res, nil
}

// hitSnippet 从正文中截取关键词附近的片段（找不到原文时返回空串）
func hitSnippet(content, keyword string) string {
	byteIdx := strings.Index(content, keyword)
	if byteIdx < 0 {
		return ""
	}
	runes := []rune(content)
	off := len([]rune(content[:byteIdx]))
	start := off - 40
	if start < 0 {
		start = 0
	}
	end := off + searchSnippetRunes
	if end > len(runes) {
		end = len(runes)
	}
	return strings.TrimSpace(string(runes[start:end]))
}

// --- Ingest 工具（内存暂存，不落库） ---

// DraftCollector ingest 草稿收集器（一次任务一个，闭包注入工具）
// 只累积结构化草稿；由 service 层在 Port.Ingest 返回后统一落库（D14）
type DraftCollector struct {
	drafts []PageDraft
}

// PageDraft 工具层草稿结构（与 iface/wikiagent.PageDraft 字段一致，由 service 层转换拷贝）
type PageDraft struct {
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	PageType   string   `json:"pageType"`
	Aliases    []string `json:"aliases,omitempty"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	Changes    string   `json:"changes,omitempty"`
	TargetSlug string   `json:"targetSlug,omitempty"`
}

// Len 已收集草稿数
func (c *DraftCollector) Len() int { return len(c.drafts) }

// All 全部草稿
func (c *DraftCollector) All() []PageDraft { return c.drafts }

// IngestToolDeps ingest 工具依赖
type IngestToolDeps struct {
	Collector *DraftCollector
}

// ProvideIngestTools 构建 ingest 工具集（草稿只进内存收集器）
func ProvideIngestTools(deps IngestToolDeps) []tool.BaseTool {
	proposePage, err := toolutils.InferTool(
		ProposePageToolName,
		"Submit one wiki page draft. Compiles (never invents) content from the provided source material. "+
			"Required: slug (lowercase letters/digits/hyphens, stable across versions), title, page_type "+
			"(summary/entity/concept), summary (one sentence), content (markdown; use [[slug|text]] to link "+
			"other pages). To revise a page you already submitted in this task, submit again with the same "+
			"slug — the previous version is replaced.",
		deps.proposePage)
	if err != nil {
		panic(err)
	}
	return mytool.WrapSafeTools([]tool.BaseTool{proposePage})
}

// proposePageParams propose_page 入参
type proposePageParams struct {
	Slug       string   `json:"slug" binding:"required"`
	Title      string   `json:"title" binding:"required"`
	PageType   string   `json:"page_type" binding:"required"`
	Aliases    []string `json:"aliases,omitempty"`
	Summary    string   `json:"summary" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Changes    string   `json:"changes,omitempty"`
	TargetSlug string   `json:"target_slug,omitempty"`
}

// proposePageResult propose_page 出参
type proposePageResult struct {
	Status      string `json:"status"` // accepted
	Slug        string `json:"slug"`
	TotalDrafts int    `json:"totalDrafts"`
	Notice      string `json:"notice,omitempty"`
}

func (d *IngestToolDeps) proposePage(_ context.Context, p *proposePageParams) (proposePageResult, error) {
	if !domain.IsValidWikiSlug(p.Slug) {
		return proposePageResult{}, fmt.Errorf("invalid slug %q: only lowercase letters, digits and hyphens allowed", p.Slug)
	}
	if strings.TrimSpace(p.Content) == "" {
		return proposePageResult{}, fmt.Errorf("content is empty")
	}
	if p.TargetSlug != "" && !domain.IsValidWikiSlug(p.TargetSlug) {
		return proposePageResult{}, fmt.Errorf("invalid target_slug %q", p.TargetSlug)
	}

	draft := PageDraft{
		Slug:       p.Slug,
		Title:      p.Title,
		PageType:   p.PageType,
		Aliases:    p.Aliases,
		Summary:    p.Summary,
		Content:    p.Content,
		Changes:    p.Changes,
		TargetSlug: p.TargetSlug,
	}

	// 同 slug 覆盖（模型迭代修订），否则追加
	replaced := false
	for i := range d.Collector.drafts {
		if d.Collector.drafts[i].Slug == draft.Slug {
			d.Collector.drafts[i] = draft
			replaced = true
			break
		}
	}
	if !replaced {
		d.Collector.drafts = append(d.Collector.drafts, draft)
	}

	notice := ""
	if replaced {
		notice = "已覆盖同 slug 的先前草稿"
	}
	return proposePageResult{
		Status:      "accepted",
		Slug:        draft.Slug,
		TotalDrafts: len(d.Collector.drafts),
		Notice:      notice,
	}, nil
}
