// 共享 Markdown 渲染工具：marked + highlight.js
// 抽离自 chat/index.vue 的内联实现，供聊天页与运营助手等场景复用。
// 渲染结果含代码块复制按钮（DOM 结构与 chat 页一致，样式由使用方按需编写）。
import { marked } from 'marked'
import hljs from 'highlight.js/lib/core'
import 'highlight.js/styles/atom-one-dark.css'

import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import cpp from 'highlight.js/lib/languages/cpp'
import csharp from 'highlight.js/lib/languages/csharp'
import go from 'highlight.js/lib/languages/go'
import rust from 'highlight.js/lib/languages/rust'
import sql from 'highlight.js/lib/languages/sql'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import scss from 'highlight.js/lib/languages/scss'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import markdown from 'highlight.js/lib/languages/markdown'
import bash from 'highlight.js/lib/languages/bash'
import shell from 'highlight.js/lib/languages/shell'
import dockerfile from 'highlight.js/lib/languages/dockerfile'

hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('java', java)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('csharp', csharp)
hljs.registerLanguage('go', go)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('scss', scss)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', shell)
hljs.registerLanguage('dockerfile', dockerfile)

marked.setOptions({
  gfm: true,
  breaks: true
})

// 代码块：语言标签 + 复制按钮（onclick 用 this.closest 定位所属块，
// 页面上多个相同内容的代码块互不串扰）
const renderer = new marked.Renderer()
renderer.code = ({ text, lang }) => {
  const codeStr = String(text || '')
  const validLang = lang && hljs.getLanguage(lang) ? lang : 'plaintext'

  let highlightedCode
  try {
    highlightedCode = hljs.highlight(codeStr, { language: validLang }).value
  } catch (err) {
    console.warn('Language highlight error:', err)
    try {
      highlightedCode = hljs.highlight(codeStr, { language: 'plaintext' }).value
    } catch (err2) {
      highlightedCode = ''
    }
  }

  const langLabel = validLang === 'plaintext' ? 'TEXT' : validLang.toUpperCase()
  return `
    <pre class="code-block">
      <div class="code-header">
        <div class="lang-info">
          <span class="code-lang">${langLabel}</span>
        </div>
        <button class="copy-button" onclick="(() => {
          const codeBlock = this.closest('.code-block');
          const code = codeBlock.querySelector('code').textContent;
          const button = codeBlock.querySelector('.copy-button');
          navigator.clipboard.writeText(code)
            .then(() => {
              button.innerHTML = '<span>已复制</span>';
              setTimeout(() => { button.innerHTML = '<span>复制</span>'; }, 2000);
            })
            .catch(() => {
              button.innerHTML = '<span>复制失败</span>';
              setTimeout(() => { button.innerHTML = '<span>复制</span>'; }, 2000);
            });
        })()">
          <span>复制</span>
        </button>
      </div>
      <code class="hljs language-${validLang}">${highlightedCode}</code>
    </pre>
  `.trim()
}

// 链接新窗口打开（后台抽屉里点击链接不应带走当前页）；
// href 仅放行安全协议，防御 javascript: 伪协议
renderer.link = function ({ href, title, tokens }) {
  let text
  try {
    text = this.parser.parseInline(tokens)
  } catch (err) {
    text = ''
  }
  const safe = /^(https?:\/\/|\/|#|mailto:)/i.test(href || '') ? href : '#'
  const titleAttr = title ? ` title="${title}"` : ''
  return `<a href="${safe}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`
}

marked.use({ renderer })

// 渲染缓存：流式更新时同一内容不会重复解析（与 chat 页同一策略）
const renderCache = new Map()
const RENDER_CACHE_MAX = 200

export function renderMarkdown(content) {
  const key = String(content || '')
  const cached = renderCache.get(key)
  if (cached !== undefined) {
    return cached
  }

  let result
  try {
    result = `<div class="markdown-body">${marked(key)}</div>`
  } catch (err) {
    console.error('Markdown rendering error:', err)
    result = key
  }

  renderCache.set(key, result)
  if (renderCache.size > RENDER_CACHE_MAX) {
    renderCache.clear()
  }
  return result
}

// 复制文本到剪贴板：优先 Clipboard API，非安全上下文（如内网 http 部署）回退 execCommand
export async function copyText(text) {
  const value = String(text ?? '')
  if (!value) return false
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
      return true
    }
  } catch (e) {
    // fall through to legacy path
  }
  try {
    const ta = document.createElement('textarea')
    ta.value = value
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch (e) {
    return false
  }
}
