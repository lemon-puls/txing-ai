package domain

import (
	"regexp"
	"time"
)

// LLM Wiki 知识库数据模型
// 设计文档：docs/llmwiki_design.md（三层逻辑架构 Raw / Wiki / Schema 全部落 MySQL）

// 来源状态
const (
	WikiSourceStatusPending   = "pending"   // 已登记未 ingest
	WikiSourceStatusIngesting = "ingesting" // ingest 进行中
	WikiSourceStatusIngested  = "ingested"  // ingest 完成（草稿已产出）
	WikiSourceStatusFailed    = "failed"    // ingest 失败
)

// 来源类型
const (
	WikiSourceTypeMD  = "md"  // 管理端上传的 markdown 原文（COS 存 key）
	WikiSourceTypeURL = "url" // 网页 URL（ingest 时抓取正文）
)

// 页面状态
const (
	WikiPageStatusDraft     = "draft"     // AI 起草，待人工审核
	WikiPageStatusPublished = "published" // 已发布，问答可读
	WikiPageStatusArchived  = "archived"  // 已下线
)

// WikiSource Raw 层来源登记表
// Raw source registry; md 内容存 COS（只存 key），url 在 ingest 时抓取
type WikiSource struct {
	BaseModel
	Title       string     `gorm:"type:varchar(255);not null;comment:源标题" json:"title"`
	SourceType  string     `gorm:"type:varchar(16);not null;comment:来源类型 md/url" json:"sourceType"`
	COSKey      string     `gorm:"type:varchar(512);comment:markdown 原文的 COS key" json:"cosKey"`
	URL         string     `gorm:"type:varchar(1024);comment:网页地址(source_type=url)" json:"url"`
	ContentHash string     `gorm:"type:varchar(64);comment:正文 sha256，重复上传/变更检测" json:"contentHash"`
	Status      string     `gorm:"type:varchar(16);not null;default:'pending';index;comment:pending/ingesting/ingested/failed" json:"status"`
	ErrorMsg    string     `gorm:"type:varchar(1024);comment:最近一次 ingest 失败原因" json:"errorMsg"`
	IngestedAt  *time.Time `gorm:"type:datetime;comment:最近一次 ingest 完成时间" json:"ingestedAt"`
}

func (WikiSource) TableName() string { return "wiki_sources" }

// WikiSourceRef 页面溯源项（该页由哪个源编译而来）
// Provenance entry linking a wiki page back to its raw source
type WikiSourceRef struct {
	SourceId int64  `json:"sourceId"`
	Note     string `json:"note,omitempty"`
}

// WikiPage Wiki 层主体表（通用 wiki 页，不写死"个人"语义）
// Core wiki page table; generic by design (D1), "about me" is just content
type WikiPage struct {
	BaseModel
	Slug     string `gorm:"type:varchar(128);not null;index;comment:页面标识，跨版本稳定" json:"slug"`
	Title    string `gorm:"type:varchar(255);not null;comment:页面标题" json:"title"`
	PageType string `gorm:"type:varchar(32);not null;default:'concept';comment:页面类型 summary/entity/concept/..." json:"pageType"`
	Aliases  []string `gorm:"type:json;serializer:json;comment:别名列表(提升搜索命中)" json:"aliases"`
	Summary  string `gorm:"type:varchar(512);comment:一句话摘要(index 渲染用)" json:"summary"`
	Content  string `gorm:"type:mediumtext;comment:markdown 正文" json:"content"`
	Status   string `gorm:"type:varchar(16);not null;default:'draft';index;comment:draft/published/archived" json:"status"`
	Version  int    `gorm:"type:int;not null;default:1;comment:修订号" json:"version"`
	SourceRefs []WikiSourceRef `gorm:"type:json;serializer:json;comment:溯源[sourceId,note]" json:"sourceRefs"`
	// 草稿更新已有页时指向目标页（同一 slug 的 draft 与 published 可并存）
	TargetPageId int64 `gorm:"type:bigint;default:0;comment:草稿关联的目标页ID(更新场景)" json:"targetPageId"`
}

func (WikiPage) TableName() string { return "wiki_pages" }

// WikiLink 页面引用关系表（出入链双向查询，支撑 lint 与 read_page 链接摘要）
// Page link table; inbound/outbound queries drive lint and read_page summaries
type WikiLink struct {
	BaseModel
	FromPageId int64  `gorm:"type:bigint;not null;index;comment:出链页ID" json:"fromPageId"`
	ToSlug     string `gorm:"type:varchar(128);not null;index;comment:目标 slug(允许悬空)" json:"toSlug"`
	AnchorText string `gorm:"type:varchar(255);comment:链接文字" json:"anchorText"`
}

func (WikiLink) TableName() string { return "wiki_links" }

// WikiKV Schema 层 kv 存储（index 缓存 / 类型约定 / 提示词等低频小文本）
// Schema-layer key-value store (index cache, conventions, prompts)
type WikiKV struct {
	BaseModel
	Key   string `gorm:"type:varchar(128);not null;uniqueIndex;comment:键名" json:"key"`
	Value string `gorm:"type:mediumtext;comment:值" json:"-"`
}

func (WikiKV) TableName() string { return "wiki_kv" }

// wiki_kv 常用键名
const (
	WikiKVKeyIndexCache = "index_cache" // 已发布页 index 的渲染缓存（确认/下线时刷新）
)

// WikiToolTrace 问答过程中的工具调用轨迹
type WikiToolTrace struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Params string `json:"params"`
	Result string `json:"result"`
	Status string `json:"status"`
}

// WikiQALog 面试官问答完整留存（D10：存完整问答对，页面须挂"对话将被记录"声明）
// Full Q&A retention; the AskPanel must display a recording notice (D10)
type WikiQALog struct {
	BaseModel
	SessionId  string          `gorm:"type:varchar(64);not null;index;comment:前端会话标识(多轮聚合)" json:"sessionId"`
	Question   string          `gorm:"type:text;not null;comment:面试官提问" json:"question"`
	AnswerMd   string          `gorm:"type:mediumtext;comment:回答(markdown)" json:"answerMd"`
	ToolTraces []WikiToolTrace `gorm:"type:json;serializer:json;comment:工具调用轨迹" json:"toolTraces"`
	IpHash     string          `gorm:"type:varchar(64);comment:IP 哈希(仅限流/去重，不留明文)" json:"-"`
	DurationMs int64           `gorm:"type:bigint;default:0;comment:耗时毫秒" json:"durationMs"`
	Status     string          `gorm:"type:varchar(16);not null;default:'completed';comment:completed/failed" json:"status"`
	ErrorMsg   string          `gorm:"type:varchar(1024);comment:失败原因" json:"errorMsg"`
}

func (WikiQALog) TableName() string { return "wiki_qa_logs" }

// --- [[slug|文字]] 链接解析 ---

// wikiLinkPattern wiki 页内链接语法 [[slug]] 或 [[slug|显示文字]]，slug 限定小写字母/数字/连字符
var wikiLinkPattern = regexp.MustCompile(`\[\[([0-9a-z][0-9a-z\-]*)\|?([^\]]*)\]\]`)

// ParsedWikiLink 从页面正文中解析出的引用
type ParsedWikiLink struct {
	Slug string
	Text string
}

// ParseWikiLinks 解析正文中的 wiki 链接（按 slug 去重，保持出现顺序）
// Parse [[slug|text]] links from content; deduped by slug in order of appearance
func ParseWikiLinks(content string) []ParsedWikiLink {
	matches := wikiLinkPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]struct{}, len(matches))
	links := make([]ParsedWikiLink, 0, len(matches))
	for _, m := range matches {
		slug := m[1]
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		text := m[2]
		if text == "" {
			text = slug
		}
		links = append(links, ParsedWikiLink{Slug: slug, Text: text})
	}
	return links
}

// WikiSlugPattern 校验合法 slug（小写字母/数字/连字符，1-128）
var WikiSlugPattern = regexp.MustCompile(`^[0-9a-z][0-9a-z\-]{0,127}$`)

// IsValidWikiSlug 判断 slug 是否合法
func IsValidWikiSlug(slug string) bool { return WikiSlugPattern.MatchString(slug) }
