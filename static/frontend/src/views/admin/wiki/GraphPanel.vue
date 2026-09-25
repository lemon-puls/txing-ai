<template>
  <div class="graph-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <div class="stat-group">
          <span class="stat-chip"><i class="dot dot--summary" />概览 {{ typeCounts.summary }}</span>
          <span class="stat-chip"><i class="dot dot--entity" />实体 {{ typeCounts.entity }}</span>
          <span class="stat-chip"><i class="dot dot--concept" />概念 {{ typeCounts.concept }}</span>
          <span class="stat-chip"><i class="dot dot--dangling" />悬空 {{ typeCounts.dangling }}</span>
          <span class="stat-chip">互链 {{ data.edges.length }}</span>
          <span v-if="isolatedTotal > 0" class="stat-chip stat-chip--warn">孤立 {{ isolatedTotal }}</span>
        </div>
        <el-select
          v-model="focusSlug"
          filterable
          clearable
          placeholder="定位页面…"
          style="width: 220px"
          :filter-method="setNodeQuery"
          @change="onFocusChange"
        >
          <el-option
            v-for="n in nodeOptions"
            :key="n.slug"
            :label="`${n.title} (${n.slug})`"
            :value="n.slug"
          >
            <span>{{ n.title }}</span>
            <span class="opt-slug">{{ n.slug }}</span>
          </el-option>
        </el-select>
      </div>
      <div class="toolbar-right">
        <el-tooltip content="重新布局" placement="top">
          <el-button round circle @click="relayout">
            <el-icon><RefreshRight /></el-icon>
          </el-button>
        </el-tooltip>
        <el-button round circle :loading="loading" @click="fetchGraph">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 图谱画布 -->
    <el-empty
      v-if="!loading && data.nodes.length === 0"
      description="暂无已发布页面，先在源管理中编译并确认草稿"
    />
    <div v-show="data.nodes.length > 0" ref="chartEl" class="graph-chart"></div>

    <!-- 节点详情抽屉 -->
    <el-drawer v-model="drawerVisible" size="440px" :with-header="false" append-to-body>
      <div v-if="focusNodeInfo" class="node-drawer">
        <div class="drawer-head">
          <el-tag :type="metaOf(focusNodeInfo).tag" round effect="light">
            {{ metaOf(focusNodeInfo).label }}
          </el-tag>
          <span class="degree-chip">链入 {{ inLinks.length }} · 链出 {{ outLinks.length }}</span>
        </div>
        <h3 class="drawer-title">{{ focusNodeInfo.title }}</h3>
        <code class="drawer-slug">{{ focusNodeInfo.slug }}</code>
        <p class="drawer-summary">{{ focusNodeInfo.summary || '（无摘要）' }}</p>

        <!-- 悬空目标：尚无对应已发布页 -->
        <el-alert
          v-if="focusNodeInfo.dangling"
          type="warning"
          :closable="false"
          show-icon
          title="悬空链接目标"
          description="有页面链接到这里，但还没有对应的已发布页面。可以补充相关内容，或在页面中修正该链接。"
          class="drawer-alert"
        />

        <template v-if="outLinks.length">
          <h4 class="drawer-section">链出（{{ outLinks.length }}）</h4>
          <div class="neighbor-list">
            <el-tag
              v-for="e in outLinks"
              :key="e.target"
              round
              class="neighbor-chip"
              effect="plain"
              @click="focusBySlug(e.target)"
            >{{ nodeTitle(e.target) }} →</el-tag>
          </div>
        </template>
        <template v-if="inLinks.length">
          <h4 class="drawer-section">链入（{{ inLinks.length }}）</h4>
          <div class="neighbor-list">
            <el-tag
              v-for="e in inLinks"
              :key="e.source"
              round
              class="neighbor-chip"
              effect="plain"
              @click="focusBySlug(e.source)"
            >← {{ nodeTitle(e.source) }}</el-tag>
          </div>
        </template>

        <div v-if="!focusNodeInfo.dangling" class="drawer-actions">
          <el-button round type="primary" plain @click="viewFullContent">查看全文</el-button>
        </div>
      </div>
    </el-drawer>

    <!-- 全文弹窗 -->
    <el-dialog v-model="viewVisible" :title="viewing?.title" width="760px" top="6vh" append-to-body>
      <div class="page-view" v-html="renderedView"></div>
    </el-dialog>
  </div>
</template>

<script setup name="WikiGraphPanel">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshRight } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { GraphChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { renderWikiMarkdown } from '@/utils/wikiMarkdown'
import wikiApi from '@/api/wiki'

// 注册 ECharts 组件（按需，与 dashboard 同款范式）
use([CanvasRenderer, GraphChart, TooltipComponent, LegendComponent])

// 页面类型元数据：标签 / el-tag / 图表配色 / 节点基准尺寸
// （echarts 画布内无法读 CSS 变量，取与 Element Plus 主题对齐的固定色值）
const TYPE_META = {
  summary: { label: '概览', tag: 'warning', color: '#E6A23C', size: 30 },
  entity: { label: '实体', tag: 'success', color: '#67C23A', size: 22 },
  concept: { label: '概念', tag: 'primary', color: '#409EFF', size: 16 }
}
// 悬空是图谱合成节点（被链接但无对应已发布页），只由 Dangling 布尔标记，不是真实 pageType
const DANGLING_META = { label: '悬空', tag: 'info', color: '#909399' }
const metaOf = (n) => {
  if (n.dangling) return DANGLING_META
  return TYPE_META[n.pageType] || { label: n.pageType || '未知', tag: 'info', color: '#909399' }
}

// 图表分类（顺序即 legend 顺序，悬空固定在末位）
const CATEGORY_LIST = [
  ...Object.entries(TYPE_META).map(([type, m]) => ({ type, name: m.label, itemStyle: { color: m.color } })),
  { type: 'dangling', name: DANGLING_META.label, itemStyle: { color: DANGLING_META.color } }
]
const CATEGORY_INDEX = Object.fromEntries(CATEGORY_LIST.map((c, i) => [c.type, i]))
const catOf = (n) => (n.dangling ? CATEGORY_INDEX.dangling : CATEGORY_INDEX[n.pageType] ?? CATEGORY_INDEX.dangling)

const data = ref({ nodes: [], edges: [] })
const loading = ref(false)
const chartEl = ref(null)
let chart = null

const drawerVisible = ref(false)
const focusSlug = ref('')
const viewVisible = ref(false)
const viewing = ref(null)

// slug → 节点、slug → {链入, 链出}：tooltip / 抽屉 / 统计共用，
// 避免悬停、渲染等热路径里对 nodes/edges 反复全表扫描
const nodeBySlug = computed(() => new Map(data.value.nodes.map(n => [n.slug, n])))
const degreeBySlug = computed(() => {
  const deg = new Map()
  for (const e of data.value.edges) {
    const src = deg.get(e.source) || { in: 0, out: 0 }
    src.out++
    deg.set(e.source, src)
    const dst = deg.get(e.target) || { in: 0, out: 0 }
    dst.in++
    deg.set(e.target, dst)
  }
  return deg
})

// 概览统计全部由图谱数据派生：类型计数（悬空按布尔标记）+ 孤立页（度为 0 的已发布页）
const typeCounts = computed(() => {
  const c = { summary: 0, entity: 0, concept: 0, dangling: 0 }
  for (const n of data.value.nodes) {
    if (n.dangling) c.dangling++
    else if (n.pageType in c) c[n.pageType]++
  }
  return c
})
const isolatedTotal = computed(() =>
  data.value.nodes.filter(n => !n.dangling && !degreeBySlug.value.has(n.slug)).length
)

const nodeTitle = (slug) => nodeBySlug.value.get(slug)?.title || slug
const nodeIndex = (slug) => data.value.nodes.findIndex(n => n.slug === slug)

const outLinks = computed(() => data.value.edges.filter(e => e.source === focusSlug.value))
const inLinks = computed(() => data.value.edges.filter(e => e.target === focusSlug.value))
const focusNodeInfo = computed(() => nodeBySlug.value.get(focusSlug.value) || null)

// 定位下拉：查询词是响应式状态，候选随之派生；
// 非悬空节点预拼小写检索串（标题/slug/摘要），键入过滤时单趟匹配
const nodeQuery = ref('')
const searchPool = computed(() => data.value.nodes
  .filter(n => !n.dangling)
  .map(n => ({ node: n, hay: `${n.title}\n${n.slug}\n${n.summary || ''}`.toLowerCase() })))
const nodeOptions = computed(() => {
  const q = nodeQuery.value.trim().toLowerCase()
  if (!q) return searchPool.value.map(p => p.node)
  return searchPool.value.filter(p => p.hay.includes(q)).slice(0, 30).map(p => p.node)
})
const setNodeQuery = (q) => { nodeQuery.value = q }

const renderedView = computed(() => {
  if (!viewing.value) return ''
  try {
    return renderWikiMarkdown(viewing.value.content)
  } catch {
    return viewing.value.content
  }
})

// --- 图谱构建 ---

const buildOption = () => {
  const sizeOf = (n) => {
    if (n.dangling) return 13
    return (TYPE_META[n.pageType]?.size || 16) + Math.min(n.degree * 2, 14)
  }
  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      confine: true,
      formatter: (p) => {
        if (p.dataType === 'edge') {
          return `${escapeHtml(nodeTitle(p.data.source))} → ${escapeHtml(nodeTitle(p.data.target))}<br/>` +
            `<span style="color:#909399">${escapeHtml(p.data.anchorText || '')}</span>`
        }
        const n = data.value.nodes[p.dataIndex]
        if (!n) return ''
        const meta = metaOf(n)
        const deg = degreeBySlug.value.get(n.slug) || { in: 0, out: 0 }
        return `<b>${escapeHtml(n.title)}</b><br/>` +
          `<span style="color:${meta.color}">● ${escapeHtml(meta.label)}</span> ` +
          `<span style="color:#909399">${escapeHtml(n.slug)}</span><br/>` +
          `${escapeHtml(n.summary || '')}<br/>` +
          `<span style="color:#909399">链入 ${deg.in} · 链出 ${deg.out}</span>`
      }
    },
    legend: [{
      data: CATEGORY_LIST.map(c => c.name),
      top: 8,
      left: 12,
      icon: 'circle',
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: '#606266', fontSize: 12 }
    }],
    series: [{
      type: 'graph',
      layout: 'force',
      roam: true,
      draggable: true,
      // legend 色块取自 category 的 itemStyle（节点级 color 只影响节点本身）
      categories: CATEGORY_LIST.map(({ type, ...meta }) => meta),
      force: { repulsion: 420, gravity: 0.08, edgeLength: [90, 210], layoutAnimation: true },
      label: {
        show: true,
        position: 'bottom',
        fontSize: 11,
        color: '#606266',
        formatter: (p) => {
          const t = nodeTitle(p.name) || ''
          return t.length > 14 ? t.slice(0, 13) + '…' : t
        }
      },
      labelLayout: { hideOverlap: true },
      emphasis: { focus: 'adjacency', label: { fontWeight: 600 } },
      itemStyle: { shadowBlur: 4, shadowColor: 'rgba(0,0,0,0.12)' },
      lineStyle: { color: '#dcdfe6', width: 1.2, curveness: 0.12 },
      data: data.value.nodes.map(n => ({
        id: n.slug,
        name: n.slug,
        category: catOf(n),
        symbolSize: sizeOf(n),
        itemStyle: n.dangling
          ? { color: '#eef0f3', borderColor: '#c0c4cc', borderWidth: 1.4, opacity: 0.85 }
          : { color: metaOf(n).color }
      })),
      links: data.value.edges.map(e => ({
        source: e.source,
        target: e.target,
        anchorText: e.anchorText
      }))
    }]
  }
}

const escapeHtml = (s) =>
  String(s || '').replace(/[&<>"']/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch]))

const renderChart = () => {
  if (!chartEl.value || data.value.nodes.length === 0) return
  // el-tabs 非 lazy：本面板在隐藏 pane 中提前挂载，此时容器宽高为 0，init 会得到空画布；
  // 未可见时跳过，等 ResizeObserver 在 pane 激活后触发真正的 init
  if (chartEl.value.offsetParent === null) return
  if (!chart) {
    chart = echarts.init(chartEl.value)
    chart.on('click', (p) => {
      if (p.dataType === 'node') onFocusChange(p.data.id)
    })
  }
  chart.setOption(buildOption(), true)
}

// --- 交互 ---

const highlightNode = (slug) => {
  if (!chart) return
  chart.dispatchAction({ type: 'downplay', seriesIndex: 0 })
  const idx = nodeIndex(slug)
  if (idx >= 0) {
    chart.dispatchAction({ type: 'highlight', seriesIndex: 0, dataIndex: idx })
    chart.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: idx })
  }
}

const focusBySlug = (slug) => {
  // 抽屉里点邻居 → 切换聚焦对象（图谱随之高亮邻接）
  focusSlug.value = slug
  nodeQuery.value = ''
  highlightNode(slug)
}

const onFocusChange = (slug) => {
  if (!slug) {
    clearFocus()
    return
  }
  focusSlug.value = slug
  nodeQuery.value = '' // 选中后复位过滤词，避免下次展开下拉仍是过滤态
  drawerVisible.value = true
  highlightNode(slug)
}

const clearFocus = () => {
  focusSlug.value = ''
  drawerVisible.value = false
  if (chart) {
    chart.dispatchAction({ type: 'downplay', seriesIndex: 0 })
    chart.dispatchAction({ type: 'hideTip' })
  }
}

const relayout = () => {
  if (chart) renderChart()
}

const viewFullContent = async () => {
  const node = focusNodeInfo.value
  if (!node || node.dangling) return
  try {
    const response = await wikiApi.getPublished(node.id)
    if (response.code === 0) {
      viewing.value = response.data
      viewVisible.value = true
    } else {
      ElMessage.error(response.msg || '加载失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载页面内容失败')
  }
}

// --- 数据 ---

const fetchGraph = async () => {
  loading.value = true
  try {
    const response = await wikiApi.graph()
    if (response.code === 0) {
      data.value = response.data || { nodes: [], edges: [] }
      drawerVisible.value = false
      focusSlug.value = ''
      nodeQuery.value = ''
      nextTick(renderChart)
    } else {
      ElMessage.error(response.msg || '加载失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载知识图谱失败')
  } finally {
    loading.value = false
  }
}

let resizeObserver = null

const onResize = () => chart && chart.resize()

onMounted(() => {
  fetchGraph()
  window.addEventListener('resize', onResize)
  // 容器尺寸变化（含隐藏 pane 激活后从 0 恢复）时补一次 init/resize
  resizeObserver = new ResizeObserver(() => {
    if (!chart) {
      renderChart()
    } else {
      onResize()
    }
  })
  if (chartEl.value) resizeObserver.observe(chartEl.value)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  if (chart) {
    chart.dispose()
    chart = null
  }
})

// 供父组件在 Tab 切换时触发刷新
defineExpose({ refresh: fetchGraph })
</script>

<style scoped lang="scss">
.graph-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;

  .toolbar-left {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }

  .toolbar-right {
    display: flex;
    gap: 12px;
    margin-left: auto;
  }
}

.stat-group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.stat-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 12px;
  border-radius: 999px;
  background: var(--el-fill-color-light);
  font-size: 12.5px;
  color: var(--el-text-color-regular);

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .dot--summary { background: #e6a23c; }
  .dot--entity { background: #67c23a; }
  .dot--concept { background: #409eff; }
  .dot--dangling { background: #909399; }

  &--warn {
    background: var(--el-color-warning-light-9);
    color: var(--el-color-warning-dark-2);
  }
}

.opt-slug {
  float: right;
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.graph-chart {
  height: 640px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  background: var(--el-fill-color-blank);
}

.node-drawer {
  padding: 4px 4px 20px;

  .drawer-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;

    .degree-chip {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .drawer-title {
    margin: 0 0 6px;
    font-size: 18px;
    font-weight: 700;
    color: var(--el-text-color-primary);
  }

  .drawer-slug {
    display: inline-block;
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 12.5px;
    background: var(--el-fill-color-light);
    padding: 2px 8px;
    border-radius: 6px;
    color: var(--el-text-color-secondary);
  }

  .drawer-summary {
    margin: 12px 0;
    font-size: 13.5px;
    line-height: 1.7;
    color: var(--el-text-color-regular);
  }

  .drawer-alert {
    margin: 12px 0;
    border-radius: 10px;
  }

  .drawer-section {
    margin: 18px 0 8px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .neighbor-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;

    .neighbor-chip {
      cursor: pointer;

      &:hover {
        color: var(--el-color-primary);
        border-color: var(--el-color-primary);
      }
    }
  }

  .drawer-actions {
    margin-top: 22px;
  }
}

.page-view {
  max-height: 70vh;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.8;

  h1, h2, h3 { margin: 16px 0 8px; }
  h4, h5 { margin: 12px 0 6px; }
  p { margin: 8px 0; }
  ul, ol { margin: 4px 0 8px; padding-left: 20px; }
  blockquote {
    margin: 8px 0;
    padding: 4px 12px;
    border-left: 3px solid var(--el-color-primary-light-5);
    color: var(--el-text-color-secondary);
  }
  a { color: var(--el-color-primary); }
  code {
    background: var(--el-fill-color-light);
    padding: 1px 6px;
    border-radius: 6px;
    font-size: 13px;
  }
  pre {
    background: var(--el-fill-color-light);
    border-radius: 10px;
    padding: 12px;
    overflow-x: auto;
  }

  .wiki-cite {
    display: inline-block;
    padding: 0 8px;
    margin: 0 2px;
    border-radius: 999px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-size: 12.5px;
    line-height: 20px;
    white-space: nowrap;

    &::before { content: '🔗 '; font-size: 11px; }
  }
}
</style>
