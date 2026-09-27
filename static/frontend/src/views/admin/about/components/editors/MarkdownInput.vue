<template>
  <!-- 全屏时 Teleport 到 body：el-drawer 常驻 transform（rtl translateX）会把 position:fixed
       的包含块劫持为抽屉盒子，导致全屏浮层坐标错乱甚至飞出视口，必须脱离 transform 祖先 -->
  <!-- teleport to body while fullscreen: el-drawer keeps a transform which hijacks the
       fixed-position containing block and throws the fullscreen layer off-viewport -->
  <Teleport to="body" :disabled="!fullscreen">
    <div ref="rootRef" class="markdown-input" :class="{ 'is-fullscreen': fullscreen }" :style="{ '--mi-height': editorHeight }">
      <Editor
        :value="modelValue"
        :plugins="plugins"
        :locale="zhLocale"
        :placeholder="placeholder"
        mode="split"
        @change="(v) => emit('update:modelValue', v)"
      />
      <!-- 全屏时右下角操作提示 -->
      <!-- floating hint while in fullscreen -->
      <Transition name="el-fade-in">
        <span v-if="fullscreen" class="mi-fullscreen-tip">Esc 退出全屏</span>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup>
// Markdown 编辑器（bytemd）：分屏实时预览 + 工具栏 + 内置全屏 + mermaid 实时图表
// Markdown editor (bytemd): split-pane live preview + toolbar + built-in fullscreen + live mermaid
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { Editor } from '@bytemd/vue-next'
import gfm from '@bytemd/plugin-gfm'
import breaks from '@bytemd/plugin-breaks'
import 'bytemd/dist/index.css'
import 'github-markdown-css/github-markdown.css'

const props = defineProps({
  modelValue: { type: String, default: '' },
  rows: { type: Number, default: 8 },
  placeholder: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])

const rootRef = ref(null)

// ==================== mermaid 实时预览（复用前台 mermaid v12，同一动态 chunk） ====================
// live mermaid preview reuses the front page's mermaid v12 (same dynamic chunk)
let mermaidLoading = null
const loadMermaid = () => {
  if (!mermaidLoading) {
    mermaidLoading = import('mermaid').then(({ default: mermaid }) => {
      // 初始化参数与前台 views/about/index.vue loadMermaid 完全一致，保证所见即所得
      // same initialize options as the front page for WYSIWYG consistency
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        theme: 'base',
        themeVariables: {
          primaryColor: '#eef4ff',
          primaryBorderColor: '#6366f1',
          primaryTextColor: '#1f2937',
          lineColor: '#94a3b8',
          fontSize: '13px'
        }
      })
      return mermaid
    })
  }
  return mermaidLoading
}

// 同源码的渲染结果缓存，输入过程中避免重复解析（FIFO 上限 50 条）
// cache rendered SVG by source to avoid re-parsing while typing
const mermaidCache = new Map()
let mermaidSeq = 0

const mermaidPreview = () => ({
  // ⚠️ 绝不能用 async 函数：bytemd 会把 viewerEffect 的任何非 null 返回值当清理函数、
  // 在下一次预览更新时于 svelte flush 内调用；async 返回 Promise → TypeError →
  // flush 中断 → 页面上所有 bytemd 编辑器的预览/状态栏/工具栏永久冻结。
  // never declare viewerEffect as async: bytemd treats any non-null return value as a
  // cleanup fn and invokes it inside the next svelte flush; a leaked Promise throws a
  // TypeError that kills the flush and freezes every bytemd editor on the page
  viewerEffect: ({ markdownBody }) => {
    // 桥接 github-markdown-css（bytemd 预览容器不带该 class）
    // bridge github-markdown-css onto the preview container
    markdownBody.classList.add('markdown-body')
    const nodes = [...markdownBody.querySelectorAll('pre > code.language-mermaid')]
    if (!nodes.length) return
    // 异步逻辑收进内部函数，Promise 不外泄（返回 undefined = 无清理函数）
    // keep the promise inside — returning undefined means "no cleanup fn"
    renderMermaidCharts(markdownBody, nodes)
  }
})

async function renderMermaidCharts(markdownBody, nodes) {
  try {
    const mermaid = await loadMermaid()
    for (const el of nodes) {
      const pre = el.parentElement
      if (!pre || pre.dataset.mermaidDone === '1') continue
      const source = el.textContent || ''
      if (mermaidCache.has(source)) {
        pre.dataset.mermaidDone = '1'
        const holder = document.createElement('div')
        holder.className = 'mermaid-chart'
        holder.innerHTML = mermaidCache.get(source)
        pre.replaceWith(holder)
        continue
      }
      try {
        const { svg } = await mermaid.render(`admin-mmd-${Date.now()}-${mermaidSeq++}`, source)
        if (mermaidCache.size >= 50) mermaidCache.delete(mermaidCache.keys().next().value)
        mermaidCache.set(source, svg)
        pre.dataset.mermaidDone = '1'
        const holder = document.createElement('div')
        holder.className = 'mermaid-chart'
        holder.innerHTML = svg
        pre.replaceWith(holder)
      } catch (e) {
        // 语法错误：保留源码块展示，不阻塞其他块
        // keep the source block visible on syntax errors
        console.error('[MarkdownInput] mermaid render failed:', e?.message || e, e?.str || '')
      }
    }
  } catch {
    // mermaid 加载失败时预览退化为源码展示
    // degrade to source view if mermaid fails to load
  }
}

const plugins = [gfm({ locale: { strike: '删除线', strikeText: '删除文本', task: '任务列表', taskText: '任务' } }), breaks(), mermaidPreview()]

// ==================== 全屏（以 bytemd 内置按钮为唯一真值源） ====================
// bytemd 1.22 的全屏按钮只切自身 .bytemd-fullscreen class、不对外发事件（vue-next 包装层也只转发 change），
// 因此这里用 MutationObserver 跟随 class 显示 Esc 提示；Esc 退出 = 反向点击工具栏上的「退出全屏」图标。
// bytemd's fullscreen button only toggles its own .bytemd-fullscreen class and emits no event,
// so we mirror the class via MutationObserver for the Esc hint; Esc exits by clicking that icon back.
const fullscreen = ref(false)
let classObserver = null
let editorRootEl = null

// 全屏时编辑器被 Teleport 出 el-drawer，EP focus-trap 在 document 层监听 focusin/focusout，
// 会把「焦点逃出抽屉」当成逃逸立刻抢回抽屉内（focusin 同步抢回 + focusout setTimeout(0) 抢回），
// 导致进入全屏后无法打字。window capture 阶段早于 document 层监听，这里把源自/落入编辑器
// 子树的焦点事件对 trap「隐身」（聚焦本身发生在事件派发前，屏蔽事件不影响聚焦结果）。
// while fullscreen the editor teleports out of the el-drawer; EP's focus-trap listens at
// document level and treats the escaped focus as an escape, yanking it back (sync on
// focusin + setTimeout(0) on focusout), which blocks typing. Window capture runs before
// document-level listeners, so we hide focus events touching our subtree from the trap —
// focusing already happened before dispatch, so hiding the event keeps focus where it is.
const blindFocusTrap = (e) => {
  if (!fullscreen.value || !rootRef.value) return
  const related = e.relatedTarget
  if (rootRef.value.contains(e.target) || (related && rootRef.value.contains(related))) {
    e.stopPropagation()
  }
}

// 退出全屏态的 OffScreen 图标 svg 路径特征（bytemd 1.22 内置图标，用于精确定位按钮）
// distinctive path of the OffScreen icon (built-in icon in bytemd 1.22)
const OFFSCREEN_PATH = 'M33 6v9h9M15 6v9H6M15 42v-9H6M33 42v-9h8.9'

// 在 capture 阶段拦截：焦点在 CodeMirror 内时 CM5 会对 Esc stopPropagation，冒泡阶段收不到；
// 同时阻断继续传播，避免 Esc 连带触发 el-drawer 关闭
// capture-phase: CM5 stops propagation for consumed keys; also keeps el-drawer from closing
const onKeydown = (e) => {
  if (e.key !== 'Escape' || !fullscreen.value || !editorRootEl) return
  e.stopImmediatePropagation()
  const btn = [...editorRootEl.querySelectorAll('.bytemd-tippy.bytemd-tippy-right')].find((i) =>
    i.innerHTML.includes(OFFSCREEN_PATH)
  )
  btn?.click()
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown, true)
  window.addEventListener('focusin', blindFocusTrap, true)
  window.addEventListener('focusout', blindFocusTrap, true)
  editorRootEl = rootRef.value?.querySelector('.bytemd') || null
  if (editorRootEl) {
    classObserver = new MutationObserver(() => {
      fullscreen.value = editorRootEl.classList.contains('bytemd-fullscreen')
    })
    classObserver.observe(editorRootEl, { attributes: true, attributeFilter: ['class'] })
  }
  // CM5 在挂载时测量几何；编辑器若在隐藏容器中初始化（如 el-drawer 开启动画早期），
  // 测量结果全为 0，首屏左栏空白、点击后才重算渲染。等布局稳定后强制 refresh。
  // CM5 measures geometry on mount; initialized inside a hidden container (early
  // el-drawer opening phase) the measurements are 0 and the edit pane stays blank
  // until a click forces recalculation. Refresh once the layout has settled.
  nextTick(() => {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        rootRef.value?.querySelectorAll('.CodeMirror').forEach((el) => el.CodeMirror?.refresh())
      })
    })
  })
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown, true)
  window.removeEventListener('focusin', blindFocusTrap, true)
  window.removeEventListener('focusout', blindFocusTrap, true)
  classObserver?.disconnect()
})

// Teleport 进出 body 会搬移 DOM：CM5 的几何缓存随之失效需 refresh()；
// 进入全屏时 DOM 搬移会把焦点丢到 body，主动还给编辑器以便继续输入
// teleport moves the DOM: CM5 caches geometry so refresh() is required;
// the move also blurs the editor — hand focus back so typing can continue
watch(fullscreen, async (on) => {
  await nextTick()
  rootRef.value?.querySelectorAll('.CodeMirror').forEach((el) => {
    const cm = el.CodeMirror
    if (!cm) return
    cm.refresh()
    if (on) cm.focus()
  })
})

// ==================== 中文 locale ====================
// 注意：bytemd 的 locale 是整体替换不合并，缺键会导致对应按钮 title 为空，必须给全 49 键
// bytemd replaces the locale wholesale (no merge) — all keys must be provided
const zhLocale = {
  bold: '粗体',
  boldText: '粗体文本',
  cheatsheet: 'Markdown 速查',
  closeHelp: '关闭帮助',
  closeToc: '关闭目录',
  code: '代码',
  codeBlock: '代码块',
  codeLang: '语言',
  codeText: '代码',
  exitFullscreen: '退出全屏',
  exitPreviewOnly: '退出仅预览',
  exitWriteOnly: '退出仅编辑',
  fullscreen: '全屏',
  h1: '一级标题',
  h2: '二级标题',
  h3: '三级标题',
  h4: '四级标题',
  h5: '五级标题',
  h6: '六级标题',
  headingText: '标题',
  help: '帮助',
  hr: '分隔线',
  image: '图片',
  imageAlt: '替代文本',
  imageTitle: '标题',
  italic: '斜体',
  italicText: '斜体文本',
  limited: '已达最大字符数限制',
  lines: '行数',
  link: '链接',
  linkText: '链接文本',
  ol: '有序列表',
  olItem: '列表项',
  preview: '预览',
  previewOnly: '仅预览',
  quote: '引用',
  quotedText: '引用文本',
  shortcuts: '快捷键',
  source: '源码',
  sync: '同步滚动',
  toc: '目录',
  top: '回到顶部',
  ul: '无序列表',
  ulItem: '列表项',
  words: '字数',
  write: '编辑',
  writeOnly: '仅编辑'
}

// rows 换算编辑器高度（工具栏 + 代码行）；全屏时由样式接管为 100%
// map rows to editor height; fullscreen overrides it to 100%
const editorHeight = computed(() => `${Math.max(240, props.rows * 28 + 52)}px`)
</script>

<style lang="scss" scoped>
.markdown-input {
  position: relative;
  width: 100%;

  :deep(.bytemd) {
    height: var(--mi-height);
    border-radius: 8px;
    font-family: inherit;

    // bytemd 内部用 .bytemd-body{height:calc(100% - 58px)} 的魔数分配高度，但 toolbar(41px)+status(25px)
    // 实际占 66px，状态栏会被挤出容器（点击「同步滚动/回到顶部」失效）。改 flex 布局摆脱魔数。
    // bytemd hardcodes calc(100% - 58px) for the body while toolbar+status actually take 66px,
    // which pushes the status bar out of the container — flex layout removes the magic number.
    display: flex;
    flex-direction: column;

    .bytemd-toolbar {
      flex-shrink: 0;
    }

    .bytemd-body {
      flex: 1;
      height: auto !important;
      min-height: 0;
    }

    .bytemd-status {
      flex-shrink: 0;
    }

    // ---- 全屏：bytemd 内置按钮切 .bytemd-fullscreen（默认 inset:0 直角贴边），改为圆角浮层盖过 el-drawer ----
    // fullscreen: restyle the built-in state class into a rounded floating layer above el-drawer
    &.bytemd-fullscreen {
      inset: 16px;
      z-index: 3000;
      height: auto !important;
      border-radius: 12px;
      box-shadow: var(--el-box-shadow-dark);
      background: var(--el-bg-color);
      overflow: hidden;
    }

    // ---- 暗色适配：bytemd 1.22 样式为硬编码浅色，跟随主题变量覆盖主要面板 ----
    // dark theme: override bytemd's hardcoded light colors with theme tokens
    border-color: var(--el-border-color-light);
    background: var(--el-bg-color);
    color: var(--el-text-color-regular);

    .bytemd-toolbar {
      background: var(--el-bg-color);
      border-bottom-color: var(--el-border-color-lighter);

      .bytemd-toolbar-icon {
        color: var(--el-text-color-regular);

        &:hover {
          background: var(--el-fill-color);
        }

        &.bytemd-tips-right {
          border-left-color: var(--el-border-color-lighter);
        }
      }
    }

    .bytemd-status {
      background: var(--el-bg-color);
      border-top-color: var(--el-border-color-lighter);
      color: var(--el-text-color-secondary);
    }

    .bytemd-split {
      border-left-color: var(--el-border-color-lighter);
    }

    .bytemd-sidebar {
      background: var(--el-bg-color);
      border-color: var(--el-border-color-light);

      li.active {
        background: var(--el-fill-color);
      }
    }

    .bytemd-dropdown {
      background: var(--el-bg-color);
      border-color: var(--el-border-color-light);
      box-shadow: var(--el-box-shadow-light);

      .bytemd-dropdown-item:hover {
        background: var(--el-fill-color);
      }
    }

    // CodeMirror 5 编辑区跟随主题（selection/cursor 高亮保持默认）
    // CodeMirror 5 pane follows the theme tokens
    .CodeMirror {
      background: var(--el-bg-color);
      color: var(--el-text-color-regular);

      .CodeMirror-gutters {
        background: var(--el-fill-color-light);
        border-right-color: var(--el-border-color-lighter);
      }

      .CodeMirror-cursor {
        border-left-color: var(--el-color-primary);
      }

      .CodeMirror-placeholder {
        color: var(--el-text-color-placeholder);
      }
    }

    // 预览区
    .bytemd-preview {
      background: var(--el-bg-color);
      padding: 12px 20px;
      font-size: 13px;

      // github-markdown-css 用 prefers-color-scheme 媒体查询切换深浅色变量：
      // 系统深色时预览区会强制黑底，与后台主题脱节。这里把它的色彩变量改绑到 el 主题变量，
      // 让预览跟随后台明暗主题（选择器带 :deep 前缀，优先级高于媒体查询内的 .markdown-body）
      // github-markdown-css flips to dark via prefers-color-scheme which desyncs from the
      // admin theme; rebind its tokens to Element Plus variables so the preview follows the theme
      .markdown-body {
        --fgColor-default: var(--el-text-color-primary);
        --fgColor-muted: var(--el-text-color-secondary);
        --fgColor-accent: var(--el-color-primary);
        --bgColor-default: transparent;
        --bgColor-muted: var(--el-fill-color-light);
        --bgColor-inset: var(--el-fill-color);
        --bgColor-neutral-muted: var(--el-fill-color);
        --bgColor-attention-muted: var(--el-color-warning-light-9);
        --borderColor-default: var(--el-border-color);
        --borderColor-muted: var(--el-border-color-lighter);
        --borderColor-accent-emphasis: var(--el-color-primary);
        background: transparent;
        color: var(--el-text-color-primary);
      }
    }

    // mermaid 图表容器（与前台 .mermaid-chart 一致的留白）
    .mermaid-chart {
      padding: 8px 0;
      overflow-x: auto;

      svg {
        max-width: 100%;
        height: auto;
      }
    }
  }
}

.mi-fullscreen-tip {
  position: absolute;
  right: 14px;
  bottom: 10px;
  z-index: 10;
  padding: 3px 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color);
  border-radius: 999px;
  pointer-events: none;
}

// 全屏时编辑器本体已 fixed 悬浮，提示跟随视口右下角
// while fullscreen the editor is fixed — anchor the hint to the viewport corner
.markdown-input.is-fullscreen .mi-fullscreen-tip {
  position: fixed;
  right: 30px;
  bottom: 26px;
  z-index: 3001;
  background: var(--el-bg-color);
  box-shadow: var(--el-box-shadow-light);
}
</style>
