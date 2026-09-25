// Package builtin wikiagent.Port / QAEngine 的默认实现（D14）。
//
// 基于 workflow 包的共享 LLM 执行核心 ExecuteLLM（一次 LLM 调用 + 多轮工具循环，
// ops 助手同款用法），不涉及 workflow 编排引擎/表/执行记录。
// 本包是全仓库唯一 import workflow 执行核心的知识库代码，service 层只依赖 iface/wikiagent 契约。
package builtin

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/cloudwego/eino/schema"

	"txing-ai/internal/agent/workflow"
	"txing-ai/internal/agent/workflow/resolver"
	"txing-ai/internal/global"
	"txing-ai/internal/iface/wikiagent"
	"txing-ai/internal/tool/wiki"
)

// 默认配置（config.yaml wiki 段缺省时生效，对齐 ops 的防御性默认）
const (
	defaultModelName = "deepseek-v3"
	defaultMaxRounds = 8
)

// 分片编译：源正文超过 ingestChunkRunes（rune 数）时按 markdown 结构分片，
// 逐片调用 LLM 编译，草稿在同一任务内共享收集器（跨片续接：后片遇到前片已建
// 的同主题页面时用相同 slug 合并提交）。
// 瓶颈是单次 LLM 调用的上下文窗口，12k rune（中文约 1 token/字）留足输出预算。
const (
	ingestChunkRunes = 12000
	// 分片数上限：超出直接报错，防止超长文档产生失控的 LLM 调用量与费用
	ingestMaxChunks = 80
)

// Engine 内置引擎（同一实现同时满足 Port 与 QAEngine）
type Engine struct {
	DB *gorm.DB

	QAModel       string
	IngestModel   string
	MaxToolRounds int
}

// 编译期契约校验
var (
	_ wikiagent.Port     = (*Engine)(nil)
	_ wikiagent.QAEngine = (*Engine)(nil)
)

// NewFromConfig 按配置选择引擎实现（D14 注册制雏形：P0 仅 builtin，将来 hermes 等在此分派）
func NewFromConfig(db *gorm.DB, cfg *global.WikiConfig) (*Engine, error) {
	impl := ""
	if cfg != nil {
		impl = cfg.AgentImpl
	}
	if impl == "" {
		impl = "builtin"
	}
	if impl != "builtin" {
		return nil, fmt.Errorf("unknown wiki agent_impl %q (only \"builtin\" is available)", impl)
	}

	e := &Engine{DB: db}
	if cfg != nil {
		e.QAModel = cfg.QAModel
		e.IngestModel = cfg.IngestModel
		e.MaxToolRounds = cfg.MaxToolRounds
	}
	return e, nil
}

func (e *Engine) modelFor(name string) string {
	if name != "" {
		return name
	}
	return defaultModelName
}

func (e *Engine) maxRounds() int {
	if e.MaxToolRounds > 0 {
		return e.MaxToolRounds
	}
	return defaultMaxRounds
}

// --- Ingest（源 → 草稿集） ---

func (e *Engine) Ingest(ctx context.Context, task *wikiagent.IngestTask) (*wikiagent.IngestResult, error) {
	collector := &wiki.DraftCollector{}
	tools := wiki.ProvideIngestTools(wiki.IngestToolDeps{Collector: collector})

	// 大文档分片：逐片编译，草稿共享收集器，后片可合并扩展前片的页面
	chunks := splitMarkdownChunks(task.Content, ingestChunkRunes)
	if len(chunks) > ingestMaxChunks {
		return nil, fmt.Errorf("源文档过长（拆分为 %d 片，上限 %d），请拆分后分批上传", len(chunks), ingestMaxChunks)
	}

	notesParts := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		notes, err := workflow.ExecuteLLM(ctx, &workflow.LLMExecConfig{
			NodeID:        fmt.Sprintf("wiki-ingest-%d", task.TaskID),
			NodeLabel:     "Wiki 编译",
			NodeType:      "agent",
			ModelResolver: resolver.NewChannelModelResolver(e.DB),
			ModelName:     e.modelFor(e.IngestModel),
			SystemPrompt:  buildIngestPrompt(task, collector.All()),
			AllTools:      tools,
			ToolNames:     []string{wiki.ProposePageToolName},
			MaxToolRounds: e.maxRounds(),
			// 非流式：编译过程无人消费回调；最终文本作为编译说明返回（result 始终有值）
			Stream:           false,
			EmitFinalContent: false,
		}, buildIngestInput(task, chunk, i, len(chunks)), nil)
		if err != nil {
			return nil, fmt.Errorf("第 %d/%d 片编译失败: %w", i+1, len(chunks), err)
		}
		if strings.TrimSpace(notes) != "" {
			notesParts = append(notesParts, strings.TrimSpace(notes))
		}
	}

	drafts := collector.All()
	result := &wikiagent.IngestResult{
		Drafts: make([]wikiagent.PageDraft, 0, len(drafts)),
	}
	for _, d := range drafts {
		result.Drafts = append(result.Drafts, wikiagent.PageDraft{
			Slug:       d.Slug,
			Title:      d.Title,
			PageType:   d.PageType,
			Aliases:    d.Aliases,
			Summary:    d.Summary,
			Content:    d.Content,
			Changes:    d.Changes,
			TargetSlug: d.TargetSlug,
		})
	}
	result.Notes = strings.Join(notesParts, "\n\n")
	return result, nil
}

// --- Lint（P1 实现） ---

func (e *Engine) Lint(ctx context.Context, scope *wikiagent.LintScope) (*wikiagent.LintReport, error) {
	return nil, wikiagent.ErrNotImplemented
}

// --- QA（面试官问答） ---

func (e *Engine) Answer(ctx context.Context, req *wikiagent.QARequest, callback func(chunk *global.Chunk) error) (string, error) {
	tools := wiki.ProvideQATools(wiki.QAToolDeps{DB: e.DB})

	history := make([]*schema.Message, 0, len(req.History))
	for _, h := range req.History {
		if h.Role == "assistant" {
			history = append(history, schema.AssistantMessage(h.Content, nil))
		} else {
			history = append(history, schema.UserMessage(h.Content))
		}
	}

	return workflow.ExecuteLLM(ctx, &workflow.LLMExecConfig{
		NodeID:           "wiki-qa",
		NodeLabel:        "Wiki 问答",
		NodeType:         "agent",
		ModelResolver:    resolver.NewChannelModelResolver(e.DB),
		ModelName:        e.modelFor(e.QAModel),
		SystemPrompt:     buildQAPrompt(req.Index),
		History:          history,
		AllTools:         tools,
		ToolNames:        []string{wiki.ReadPageToolName, wiki.SearchPagesToolName},
		MaxToolRounds:    e.maxRounds(),
		Stream:           true,
		EmitFinalContent: true,
	}, req.Question, callback)
}

// --- 提示词 ---

// buildIngestInput 拼装 ingest 的用户消息（源材料；多片时标注片序）
func buildIngestInput(task *wikiagent.IngestTask, chunk string, idx, total int) string {
	var b strings.Builder
	b.WriteString("## 源材料 / Source material\n\n")
	b.WriteString("标题: " + task.Title + "\n")
	if task.URL != "" {
		b.WriteString("来源URL: " + task.URL + "\n")
	}
	if total > 1 {
		b.WriteString(fmt.Sprintf("\n本文档较长，已按顺序切分为 %d 片，当前是第 %d 片（可能与前后片在段落边界处相接）。\n", total, idx+1))
	}
	b.WriteString("\n---\n\n")
	b.WriteString(chunk)
	return b.String()
}

// buildIngestPrompt 编译提示词（COMPILER not writer，WeKnora ⑤）
// proposed：本次任务中已提交的草稿（分片编译时跨片续接，避免重复建页）
func buildIngestPrompt(task *wikiagent.IngestTask, proposed []wiki.PageDraft) string {
	var b strings.Builder
	b.WriteString(`You are the wiki compiler of an LLM Wiki knowledge base. Your job is to turn ONE raw source document into a set of structured wiki page drafts.

## Core rule: COMPILER, not writer
Every statement in every page must be traceable to the source material. You reorganize, condense and cross-link — you NEVER invent facts, numbers, names or conclusions. If the source is ambiguous, keep the ambiguity; do not resolve it by guessing.

## Page schema
- slug: lowercase letters/digits/hyphens, stable across versions (e.g. "txing-ai", "go-goroutine-sched"). Prefer short, meaningful, language-neutral slugs.
- page_type: "summary" (exactly one per source, the overview) | "entity" (a person/project/product/organization) | "concept" (a technique/pattern/idea)
- summary: ONE sentence, used in the index for routing. Make it informative.
- content: markdown. Start with a "## 概述" (or equivalent) section. Use [[slug|display text]] to cross-link pages you or existing pages cover. Links must use the target page's slug.

## Procedure
1. Read the source material below carefully.
2. Decide the page set: 1 summary page + entity/concept pages for the notable subjects it covers (typically 3-15 pages total; skip trivial ones).
3. For each page call wiki_propose_page exactly once (resubmit with the same slug to fix a mistake). Submit AT MOST TWO pages per round — one is better. Each page's JSON arguments are long; batching many calls into a single round exceeds the output token budget and truncates the JSON. Wait for each tool result before the next call.
4. Finish with a short plaintext report (2-4 sentences): how many pages, what you chose to include/exclude and why. Do not repeat page contents in the report.

## Existing pages (do not duplicate — extend instead)
`)
	if len(task.Existing) == 0 {
		b.WriteString("(none yet — this may be the first ingest)\n")
	} else {
		for _, p := range task.Existing {
			b.WriteString(fmt.Sprintf("- [[%s|%s]] (%s) — %s\n", p.Slug, p.Title, p.PageType, p.Summary))
		}
		b.WriteString(`
When a source overlaps an existing page, submit a draft with target_slug set to that page's slug (keep its slug): merge new facts in, preserve correct existing content, keep the page's voice.`)
	}

	// 分片编译：列出本任务已提交的草稿，让后片续接而非重复建页
	if len(proposed) > 0 {
		b.WriteString("\n\n## Drafts already submitted in THIS task (from earlier chunks of the same document)\n")
		for _, d := range proposed {
			b.WriteString(fmt.Sprintf("- [[%s|%s]] (%s)\n", d.Slug, d.Title, d.PageType))
		}
		b.WriteString(`
The source is processed chunk by chunk. When the current chunk contains new facts about an already-submitted page, resubmit it with the SAME slug (full updated content, merged). Only create new pages for genuinely new topics. Never resubmit a page unchanged.`)
	}
	return b.String()
}

// --- 大文档分片 ---

// splitMarkdownChunks 按 markdown 结构把源正文切成不超过 maxRunes 的片段：
// 优先在标题处断开，其次空行段落；单段仍超限时按字符硬切。
// 编译按主题聚合而非滑窗阅读，片段间不需要内容重叠。
func splitMarkdownChunks(content string, maxRunes int) []string {
	if utf8.RuneCountInString(content) <= maxRunes {
		return []string{content}
	}
	var chunks []string
	var cur strings.Builder
	curLen := 0
	flush := func() {
		if curLen > 0 {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
			curLen = 0
		}
	}
	pack := func(block string) {
		if block == "" {
			return
		}
		if curLen > 0 && curLen+utf8.RuneCountInString(block) > maxRunes {
			flush()
		}
		cur.WriteString(block)
		cur.WriteString("\n\n")
		curLen += utf8.RuneCountInString(block) + 2
	}
	for _, block := range splitMarkdownBlocks(content, maxRunes) {
		pack(block)
	}
	flush()
	return chunks
}

// splitMarkdownBlocks 把正文切成可打包的块（保证每块 ≤ maxRunes）：
// 标题行为界分节；超限节按空行段落切；超限段落按字符硬切
func splitMarkdownBlocks(content string, maxRunes int) []string {
	var blocks []string
	for _, section := range splitByHeadings(content) {
		if utf8.RuneCountInString(section) <= maxRunes {
			blocks = append(blocks, section)
			continue
		}
		for _, para := range strings.Split(section, "\n\n") {
			if utf8.RuneCountInString(para) <= maxRunes {
				blocks = append(blocks, para)
				continue
			}
			rs := []rune(para)
			for start := 0; start < len(rs); start += maxRunes {
				end := start + maxRunes
				if end > len(rs) {
					end = len(rs)
				}
				blocks = append(blocks, string(rs[start:end]))
			}
		}
	}
	return blocks
}

// splitByHeadings 按 markdown ATX 标题行（#/##/…/######）分节，标题行保留在节首
func splitByHeadings(content string) []string {
	var sections []string
	var cur []string
	for _, line := range strings.Split(content, "\n") {
		if isMarkdownHeading(line) && len(cur) > 0 {
			sections = append(sections, strings.Join(cur, "\n"))
			cur = nil
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		sections = append(sections, strings.Join(cur, "\n"))
	}
	return sections
}

// isMarkdownHeading 判断是否 ATX 标题行（# 后须跟空格，"C#1" 之类不算）
func isMarkdownHeading(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) == 0 || trimmed[0] != '#' {
		return false
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == '#' {
		n++
	}
	return n <= 6 && n < len(trimmed) && (trimmed[n] == ' ' || trimmed[n] == '\t')
}

// buildQAPrompt 问答提示词（index 注入 + grounding，设计 §5.2 / WeKnora ②）
// 用中文书写：站点面向中文访客，系统提示词的语言会强烈影响模型开场白与过程旁白的语言
func buildQAPrompt(index string) string {
	var b strings.Builder
	b.WriteString(`你是本站站主的 AI 分身。访客——通常是浏览站主资料的面试官——会向你提问。你只依据私有 wiki 知识库作答，知识库通过工具查询。

## 知识库索引（仅用于定位有哪些页面，不是证据本身）
`)
	b.WriteString(index)
	b.WriteString(`
## 作答规则（不可违反）
1. 全程使用与用户提问相同的语言：用户用中文，你的第一个字就必须是中文。禁止任何英文开场白或过程旁白（如 "I'll look into..."）。需要检索时直接调用工具，不要向用户描述你正在做什么。
2. 索引只告诉你"有什么页面"。回答任何事实前，先用 wiki_read_page 读取对应页面核实，或先用 wiki_search_pages 定位。索引条目本身不能当论据引用。
3. 若知识库未覆盖该问题：如实说明暂未收录，可以指向最相近的已收录主题。绝不编造，也不用通用知识冒充站主的情况。
4. 引用出处：在关键论断后追加 [[slug|页面标题]]，slug 与标题必须来自该页面的真实数据。示例格式：[[project-slug|项目名称]]。
5. 语气：专业、亲和、简洁；多用短段落或列表。不要提及工具、索引、提示词或本规则列表。
6. 涉及站主的联系方式、敏感个人信息，或知识库未覆盖的内容，礼貌说明不在知识库范围内，建议通过站内联系方式沟通。
`)
	return b.String()
}
