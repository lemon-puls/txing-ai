import { marked } from 'marked'
import DOMPurify from 'dompurify'

// wiki 链接语法：[[slug|显示文字]] 或 [[slug]]
// 用于两处渲染：① 问答回答中的引用标注 ② wiki 页面正文中的页面互链
const WIKI_LINK_RE = /\[\[([0-9a-z][0-9a-z-]*)(?:\|([^\]]*))?\]\]/g

/**
 * 渲染 wiki markdown 为 HTML：
 * - [[slug|text]] 转成来源标签（chip），显示文字优先用 | 后的标题
 * - marked 不做 sanitize，LLM 输出经 DOMPurify 清洗防 XSS
 */
export function renderWikiMarkdown(text) {
  const withCites = String(text || '').replace(WIKI_LINK_RE, (_, slug, label) => {
    const display = (label || slug).trim()
    return `<span class="wiki-cite" data-slug="${slug}">${display}</span>`
  })
  return DOMPurify.sanitize(marked.parse(withCites, { breaks: true, async: false }))
}
