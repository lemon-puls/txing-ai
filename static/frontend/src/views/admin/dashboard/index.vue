<template>
  <div class="dashboard-container">
    <!-- 顶部统计卡片（由 /overview 响应驱动） -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6" v-for="item in statistics" :key="item.key">
        <el-card shadow="hover" class="statistic-card">
          <div class="stat-top">
            <div class="icon-chip" :class="iconChipClass(item.key)">
              <el-icon :size="22">
                <component :is="cardIcon(item.key)" />
              </el-icon>
            </div>
            <span class="title">{{ item.title }}</span>
          </div>
          <div class="card-content">
            <div class="value">
              {{ item.value }}<span v-if="item.unit" class="unit">{{ item.unit }}</span>
            </div>
            <div class="note-row">
              <span v-if="item.trendPercent !== null && item.trendPercent !== undefined"
                    class="trend" :class="item.trendPercent >= 0 ? 'up' : 'down'">
                {{ item.trendPercent >= 0 ? '+' : '' }}{{ item.trendPercent.toFixed(1) }}%
                <el-icon :size="12">
                  <component :is="item.trendPercent >= 0 ? 'ArrowUp' : 'ArrowDown'" />
                </el-icon>
              </span>
              <span class="note">{{ item.note }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图表 + 模型占比 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>对话趋势</span>
              <el-radio-group v-model="trendRange" size="small" @change="loadTrends">
                <el-radio-button :value="7">近 7 天</el-radio-button>
                <el-radio-button :value="30">近 30 天</el-radio-button>
                <el-radio-button :value="365">近一年</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <v-chart v-if="trendPoints.length" class="chart" :option="trendChartOption" autoresize />
            <el-empty v-else description="暂无数据" :image-size="60" />
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>模型使用占比</span>
              <span class="header-note">近 30 天 · 按会话数</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart v-if="modelUsage.length" class="pie-chart" :option="modelPieOption" autoresize />
            <el-empty v-else description="暂无数据" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 运行时监控（进程内指标） -->
    <el-row class="mb-4">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>
                运行时监控
                <el-tooltip content="进程内指标，自启动起保留 24 小时；重启后清零" placement="top">
                  <el-icon class="hint-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </span>
              <div class="header-controls">
                <el-radio-group v-model="rtMetric" size="small" @change="loadTimeseries">
                  <el-radio-button value="http">HTTP</el-radio-button>
                  <el-radio-button value="runtime">系统</el-radio-button>
                  <el-radio-button value="llm">LLM</el-radio-button>
                </el-radio-group>
                <el-radio-group v-model="rtWindow" size="small" @change="loadTimeseries">
                  <el-radio-button value="1h">近 1 小时</el-radio-button>
                  <el-radio-button value="6h">近 6 小时</el-radio-button>
                  <el-radio-button value="24h">近 24 小时</el-radio-button>
                </el-radio-group>
              </div>
            </div>
          </template>
          <el-row v-if="rtCharts.length" :gutter="20">
            <el-col :span="rtCharts.length === 3 ? 8 : 12" v-for="(chart, i) in rtCharts" :key="chart.metric">
              <div class="rt-chart-title">
                {{ chart.title }}
                <span class="rt-chart-unit">{{ chart.unit }}</span>
              </div>
              <v-chart class="rt-chart" :option="rtChartOption(chart, i)" autoresize />
            </el-col>
          </el-row>
          <el-empty v-else description="暂无运行时数据（服务刚启动或无流量）" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 渠道 + 助手 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>LLM 渠道使用</span>
              <span class="header-note">自启动累计</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart v-if="channelItems.length" class="pie-chart" :option="channelDonutOption" autoresize />
            <el-empty v-else description="暂无 LLM 调用" :image-size="60" />
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>AI 助手使用排行</span>
              <span class="header-note">近 30 天 · TOP 10</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart v-if="assistantUsage.length" class="chart" :option="assistantBarOption" autoresize />
            <el-empty v-else description="暂无数据" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row>
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
            </div>
          </template>
          <el-table :data="activities" style="width: 100%" :max-height="400">
            <el-table-column label="时间" width="180">
              <template #default="{ row }">{{ formatTime(row.time) }}</template>
            </el-table-column>
            <el-table-column prop="user" label="用户" width="180" show-overflow-tooltip />
            <el-table-column label="操作" width="140">
              <template #default="{ row }">{{ row.action }}</template>
            </el-table-column>
            <el-table-column prop="detail" label="详情" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'completed' ? 'success' : 'danger'" round>
                  {{ row.status === 'completed' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  GridComponent, TooltipComponent, LegendComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  ChatLineRound, User, Timer, DataLine, ArrowUp, ArrowDown, QuestionFilled
} from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { ElMessage } from 'element-plus'

use([CanvasRenderer, LineChart, PieChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

// ---------- dataviz 色板（light 模式，固定槽位顺序） ----------
const SERIES_COLORS = ['#2a78d6', '#eb6834', '#1baf7a', '#eda100', '#e87ba4', '#008300', '#4a3aa7', '#e34948']
const TEXT_SECONDARY = '#52514e'
const GRID_LINE = '#ececec'

// ---------- 状态 ----------
const statistics = ref([])
const trendRange = ref(7)
const trendPoints = ref([])
const trendGranularity = ref('day')
const modelUsage = ref([])
const assistantUsage = ref([])
const channelItems = ref([])
const rtMetric = ref('http')
const rtWindow = ref('1h')
const rtSeries = ref([])
const activities = ref([])
let refreshTimer = null

// ---------- 数据加载 ----------
const unwrap = (response, fallback) => {
  if (response && response.code === 0 && response.data !== undefined) return response.data
  return fallback
}

const loadOverview = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardOverviewGet(), null)
    if (data) statistics.value = data.cards || []
  } catch (e) {
    console.error('loadOverview failed', e)
  }
}

const loadTrends = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardTrendsGet({ range: trendRange.value }), null)
    if (data) {
      trendPoints.value = data.points || []
      trendGranularity.value = data.granularity || 'day'
    }
  } catch (e) {
    console.error('loadTrends failed', e)
  }
}

const loadModelUsage = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardModelUsageGet({ days: 30 }), [])
    modelUsage.value = data || []
  } catch (e) {
    console.error('loadModelUsage failed', e)
  }
}

const loadAssistantUsage = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardAssistantUsageGet({ days: 30, limit: 10 }), [])
    assistantUsage.value = data || []
  } catch (e) {
    console.error('loadAssistantUsage failed', e)
  }
}

const loadChannelUsage = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardChannelUsageGet(), null)
    if (data) channelItems.value = data.items || []
  } catch (e) {
    console.error('loadChannelUsage failed', e)
  }
}

const loadTimeseries = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardTimeseriesGet({
      metric: rtMetric.value, window: rtWindow.value
    }), null)
    if (data) rtSeries.value = data.series || []
  } catch (e) {
    console.error('loadTimeseries failed', e)
  }
}

const loadActivities = async () => {
  try {
    const data = unwrap(await defaultApi.apiAdminDashboardActivitiesGet({ limit: 20 }), [])
    activities.value = data || []
  } catch (e) {
    console.error('loadActivities failed', e)
  }
}

const loadAll = () => {
  loadOverview()
  loadTrends()
  loadModelUsage()
  loadAssistantUsage()
  loadChannelUsage()
  loadTimeseries()
  loadActivities()
}

onMounted(() => {
  loadAll()
  // 运行时监控 30s 自动刷新
  refreshTimer = setInterval(() => {
    loadOverview()
    loadTimeseries()
  }, 30_000)
})

onBeforeUnmount(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

// ---------- 卡片 ----------
const cardIconMap = {
  conversations_today: ChatLineRound,
  active_users_today: User,
  llm_latency_1h: Timer,
  tokens_since_start: DataLine
}
const cardIcon = (key) => cardIconMap[key] || DataLine

const iconChipClassMap = {
  conversations_today: 'chip-blue',
  active_users_today: 'chip-green',
  llm_latency_1h: 'chip-orange',
  tokens_since_start: 'chip-violet'
}
const iconChipClass = (key) => iconChipClassMap[key] || 'chip-blue'

// ---------- 趋势图 ----------
const trendChartOption = computed(() => ({
  color: [SERIES_COLORS[0], SERIES_COLORS[1]],
  tooltip: { trigger: 'axis', borderRadius: 10, boxShadow: '0 4px 12px rgba(0,0,0,0.08)' },
  legend: { bottom: 0, textStyle: { color: TEXT_SECONDARY } },
  grid: { left: 48, right: 16, top: 16, bottom: 40 },
  xAxis: {
    type: 'category', boundaryGap: false,
    data: trendPoints.value.map(p => p.date),
    axisLine: { lineStyle: { color: GRID_LINE } },
    axisLabel: { color: TEXT_SECONDARY, fontSize: 11 }
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: GRID_LINE } },
    axisLabel: { color: TEXT_SECONDARY, fontSize: 11 }
  },
  series: [
    {
      name: '对话数', type: 'line', smooth: true, symbol: 'circle', symbolSize: 6,
      data: trendPoints.value.map(p => p.conversations), lineStyle: { width: 2 }
    },
    {
      name: '活跃用户', type: 'line', smooth: true, symbol: 'circle', symbolSize: 6,
      data: trendPoints.value.map(p => p.activeUsers), lineStyle: { width: 2 }
    }
  ]
}))

// ---------- 饼图/环形图通用 ----------
const pieDataOf = (items, nameKey, valueKey) => {
  const list = items.map(it => ({ name: it[nameKey], value: it[valueKey] }))
  // 超过 8 个分类折叠为「其他」（固定 8 色槽位，不生成第 9 色）
  if (list.length > 8) {
    const top = list.slice(0, 7)
    const other = list.slice(7).reduce((sum, it) => sum + it.value, 0)
    top.push({ name: '其他', value: other })
    return top
  }
  return list
}

const modelPieOption = computed(() => ({
  color: SERIES_COLORS,
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', borderRadius: 10 },
  legend: { type: 'scroll', bottom: 0, textStyle: { color: TEXT_SECONDARY } },
  series: [{
    type: 'pie', radius: ['0%', '68%'], center: ['50%', '44%'],
    itemStyle: { borderColor: '#fff', borderWidth: 2 },
    label: { show: false },
    data: pieDataOf(modelUsage.value, 'name', 'count')
  }]
}))

const channelDonutOption = computed(() => ({
  color: SERIES_COLORS,
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', borderRadius: 10 },
  legend: { type: 'scroll', bottom: 0, textStyle: { color: TEXT_SECONDARY } },
  series: [{
    type: 'pie', radius: ['42%', '68%'], center: ['50%', '44%'],
    itemStyle: { borderColor: '#fff', borderWidth: 2 },
    label: { show: false },
    data: pieDataOf(channelItems.value, 'channel', 'requests')
  }]
}))

// ---------- 助手排行（单色条形） ----------
const assistantBarOption = computed(() => {
  const sorted = [...assistantUsage.value].reverse() // 升序使最大值在顶部
  return {
    tooltip: { trigger: 'item', borderRadius: 10 },
    grid: { left: 90, right: 32, top: 8, bottom: 8 },
    xAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: GRID_LINE } },
      axisLabel: { color: TEXT_SECONDARY, fontSize: 11 }
    },
    yAxis: {
      type: 'category',
      data: sorted.map(it => it.name),
      axisLabel: { color: TEXT_SECONDARY, fontSize: 11 },
      axisLine: { show: false }, axisTick: { show: false }
    },
    series: [{
      type: 'bar', barWidth: 14,
      itemStyle: { color: SERIES_COLORS[0], borderRadius: [0, 4, 4, 0] },
      data: sorted.map(it => it.count)
    }]
  }
})

// ---------- 运行时监控（小倍数图，每图单轴） ----------
const RT_META = {
  http: [
    { metric: 'http_qpm', title: '请求速率', unit: '次/分' },
    { metric: 'http_latency_ms', title: '平均时延', unit: 'ms' }
  ],
  runtime: [
    { metric: 'goroutines', title: 'Goroutines', unit: '个' },
    { metric: 'memory_mb', title: '内存占用', unit: 'MB' }
  ],
  llm: [
    { metric: 'llm_qpm', title: 'LLM 请求速率', unit: '次/分' },
    { metric: 'llm_latency_ms', title: 'LLM 平均时延', unit: 'ms' },
    { metric: 'llm_ttft_ms', title: '首 Token 时延 (TTFT)', unit: 'ms' }
  ]
}

const rtCharts = computed(() => {
  const meta = RT_META[rtMetric.value] || []
  return meta
    .map(m => ({ ...m, series: rtSeries.value.filter(s => s.metric === m.metric) }))
    .filter(c => c.series.length > 0)
})

const formatTs = (ts) => {
  const d = new Date(ts * 1000)
  const hm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  if (rtWindow.value === '24h') {
    return `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${hm}`
  }
  return hm
}

const rtChartOption = (chart, index) => {
  const points = chart.series[0]?.points || []
  return {
    color: [SERIES_COLORS[index % SERIES_COLORS.length]],
    tooltip: {
      trigger: 'axis',
      borderRadius: 10,
      valueFormatter: (v) => `${Number(v ?? 0).toFixed(1)} ${chart.unit}`
    },
    grid: { left: 52, right: 12, top: 10, bottom: 24 },
    xAxis: {
      type: 'category', boundaryGap: false,
      data: points.map(p => formatTs(p.ts)),
      axisLine: { lineStyle: { color: GRID_LINE } },
      axisLabel: { color: TEXT_SECONDARY, fontSize: 10, interval: 'auto' },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: GRID_LINE } },
      axisLabel: { color: TEXT_SECONDARY, fontSize: 10 }
    },
    series: [{
      name: chart.title, type: 'line', showSymbol: false,
      data: points.map(p => Number(p.value.toFixed(2))),
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.08 }
    }]
  }
}

// ---------- 通用 ----------
const formatTime = (t) => {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.dashboard-container {
  padding: 24px;
}

.mb-4 {
  margin-bottom: 20px;
}

/* ---- 卡片：全站后台统一圆角风格（18px + 柔和阴影 + hover 上浮） ---- */
.dashboard-container :deep(.el-card) {
  border-radius: 18px;
  border: 1px solid var(--el-border-color-lighter);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.dashboard-container :deep(.el-card:hover) {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.dashboard-container :deep(.el-card__body) {
  padding: 20px 24px;
}

/* ---- 统计卡片 ---- */
.statistic-card :deep(.el-card__body) {
  padding: 22px 24px;
}

.stat-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-chip {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.chip-blue {
  background: rgba(42, 120, 214, 0.12);
  color: #2a78d6;
}

.chip-green {
  background: rgba(27, 175, 122, 0.12);
  color: #1baf7a;
}

.chip-orange {
  background: rgba(235, 104, 52, 0.12);
  color: #eb6834;
}

.chip-violet {
  background: rgba(74, 58, 167, 0.1);
  color: #4a3aa7;
}

.stat-top .title {
  font-size: 14px;
  color: #52514e;
}

.card-content {
  margin-top: 14px;
}

.card-content .value {
  font-size: 30px;
  font-weight: 700;
  color: #1f1f1f;
  line-height: 1.2;
  letter-spacing: -0.5px;
}

.card-content .unit {
  font-size: 14px;
  font-weight: 400;
  color: #8a8984;
  margin-left: 4px;
  letter-spacing: 0;
}

.note-row {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 趋势胶囊 */
.trend {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
}

.trend.up {
  color: #e34948;
  background: rgba(227, 73, 72, 0.08);
}

.trend.down {
  color: #008300;
  background: rgba(0, 131, 0, 0.08);
}

.note {
  font-size: 12px;
  color: #8a8984;
}

/* ---- 卡片头 ---- */
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.card-header > span:first-child {
  font-size: 15px;
  font-weight: 600;
  color: #1f1f1f;
}

.header-note {
  font-size: 12px;
  color: #8a8984;
  font-weight: 400;
}

.header-controls {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.hint-icon {
  color: #a0a0a0;
  vertical-align: -2px;
  margin-left: 4px;
}

/* ---- 分段选择器（radio-button 胶囊化，贴合圆角风格） ---- */
.dashboard-container :deep(.el-radio-group) {
  background: #f2f3f5;
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
}

.dashboard-container :deep(.el-radio-group .el-radio-button__inner) {
  border: none !important;
  border-radius: 8px !important;
  padding: 5px 12px;
  font-size: 12px;
  color: #52514e;
  background: transparent;
  box-shadow: none !important;
  transition: all 0.2s ease;
}

.dashboard-container :deep(.el-radio-group .el-radio-button.is-active .el-radio-button__inner) {
  background: #ffffff;
  color: var(--el-color-primary);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1) !important;
}

/* ---- 空态 ---- */
.dashboard-container :deep(.el-empty) {
  padding: 40px 0;
}

/* ---- 图表容器 ---- */
.chart-container {
  height: 300px;
}

.chart {
  width: 100%;
  height: 100%;
}

.pie-chart {
  width: 100%;
  height: 100%;
}

/* ---- 运行时监控小倍数图 ---- */
.rt-chart-title {
  font-size: 13px;
  font-weight: 600;
  color: #52514e;
  margin-bottom: 6px;
  padding-left: 4px;
  border-left: 3px solid var(--el-color-primary);
  line-height: 1.1;
  padding-top: 1px;
  padding-bottom: 1px;
}

.rt-chart-unit {
  font-size: 12px;
  font-weight: 400;
  color: #8a8984;
  margin-left: 6px;
}

.rt-chart {
  width: 100%;
  height: 180px;
}

/* ---- 最近活动表格 ---- */
.dashboard-container :deep(.el-table) {
  border-radius: 12px;
  overflow: hidden;
}

.dashboard-container :deep(.el-table th.el-table__cell) {
  background: #fafafa;
  font-weight: 600;
  font-size: 13px;
  color: #52514e;
}

.dashboard-container :deep(.el-table .el-table__cell) {
  font-size: 13px;
}
</style>
