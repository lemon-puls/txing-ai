// Package wikiagent 定义 LLM Wiki 的 agent 引擎插拔边界（设计文档 D14 / §11.1）。
//
// 契约是数据，不是动作：任何实现（内置 builtin 或将来的三方 agent 如 hermes）
// 只接收任务、产出结构化草稿/报告，DB 写入永远由 service 层完成——
// 换引擎时表结构、审核流、审计全部零改动，三方 agent 不接触数据库。
package wikiagent

import (
	"context"
	"errors"

	"txing-ai/internal/global"
)

// Port 维护类 agent 接口（Ingest / Lint）
type Port interface {
	// Ingest 将一个 Raw 源编译为 wiki 页面草稿集（只返回草稿，不落库）
	Ingest(ctx context.Context, task *IngestTask) (*IngestResult, error)
	// Lint 知识库体检，产出问题清单（P1 实现）
	Lint(ctx context.Context, scope *LintScope) (*LintReport, error)
}

// QAEngine 问答引擎接口（面试官问答，agentic retrieval：index 导航 + 读页/搜索工具）
type QAEngine interface {
	// Answer 回答一个问题；chunk 通过 callback 流式回调（复用 global.Chunk，ops 同款帧语义），
	// 返回最终完整回答。多轮对话历史由调用方折叠进 req.History，引擎自身无状态。
	Answer(ctx context.Context, req *QARequest, callback func(chunk *global.Chunk) error) (string, error)
}

// --- 任务契约 ---

// PageBrief 已发布页概览（任务上下文用，避免把全量正文塞给引擎）
type PageBrief struct {
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	PageType string   `json:"pageType"`
	Summary  string   `json:"summary"`
	Aliases  []string `json:"aliases,omitempty"`
	Version  int      `json:"version"`
}

// IngestTask 一次源编译任务
type IngestTask struct {
	TaskID    int64  `json:"taskId"`    // wiki_sources.id，贯穿日志与产物
	Title     string `json:"title"`     // 源标题
	Kind      string `json:"kind"`      // md / url（domain.WikiSourceType*）
	URL       string `json:"url,omitempty"`
	Content   string `json:"content"` // 源正文（service 层已从 COS/抓取取回）
	Existing  []PageBrief `json:"existing,omitempty"` // 现有已发布页概览，供决定新建/更新
}

// PageDraft 引擎产出的结构化草稿（审核流的唯一输入形状）
type PageDraft struct {
	Slug       string   `json:"slug"` // 跨版本稳定；更新已有页时与目标页 slug 一致
	Title      string   `json:"title"`
	PageType   string   `json:"pageType"`
	Aliases    []string `json:"aliases,omitempty"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"` // markdown，可含 [[slug|文字]] 引用
	Changes    string   `json:"changes,omitempty"` // 变更摘要（审核列表展示）
	TargetSlug string   `json:"targetSlug,omitempty"` // 非空表示更新已有页
}

// IngestResult 一次任务的产物
type IngestResult struct {
	Drafts []PageDraft `json:"drafts"`
	Notes  string      `json:"notes,omitempty"` // 引擎对本次编译的补充说明（如放弃的分支及原因）
}

// LintScope 体检范围
type LintScope struct {
	Published []PageBrief `json:"published"`
}

// Lint 问题的四类（docs §7）
const (
	LintKindOrphan    = "orphan"    // 孤儿页：无任何入链
	LintKindBroken    = "broken"    // 断链：to_slug 无对应已发布页
	LintKindUnindexed = "unindexed" // 失收：已发布页未出现在 index
	LintKindStale     = "stale"     // 受影响页：来源更新后内容可能过时
)

// LintFinding 单个体检发现
type LintFinding struct {
	Kind    string `json:"kind"`
	Slug    string `json:"slug,omitempty"`
	Detail  string `json:"detail"`
	FixHint string `json:"fixHint,omitempty"`
}

// LintReport 体检报告
type LintReport struct {
	Findings []LintFinding `json:"findings"`
}

// --- 问答契约 ---

// QAHistoryItem 多轮对话历史项（引擎无关的扁平结构，避免契约绑定具体 LLM SDK）
type QAHistoryItem struct {
	Role    string `json:"role"` // user / assistant
	Content string `json:"content"`
}

// QARequest 一次问答请求
type QARequest struct {
	Question string          `json:"question"`
	History  []QAHistoryItem `json:"history,omitempty"` // 近期对话（语言跟随提问的依据之一）
	Index    string          `json:"index"`             // 已渲染的 index 目录（路由目录，非证据）
}

// ErrNotImplemented 预留给尚未实现的引擎能力（如 P1 的 Lint）
var ErrNotImplemented = errors.New("wikiagent: not implemented")
