<template>
  <div class="workflow-exec-container">
    <div class="workflow-exec-content">
      <el-card class="exec-card">
        <div class="two-column">
          <!-- 左侧：工作流信息 + 输入 -->
          <div class="left-panel">
            <div class="panel-section page-intro">
              <el-button text class="back-button" @click="goBack">
                <el-icon><ArrowLeft /></el-icon>
                返回市场
              </el-button>
              <h2 class="intro-title">{{ workflowInfo.name || '加载中...' }}</h2>
              <p class="intro-subtitle">{{ workflowInfo.description || '请输入内容开始执行工作流' }}</p>
            </div>

            <div class="panel-section">
              <h2 class="section-title">输入内容</h2>
              <el-form :model="formData" label-position="top" class="input-form">
                <el-form-item label="内容">
                  <el-input
                    v-model="formData.content"
                    type="textarea"
                    :rows="8"
                    placeholder="请输入您想要处理的内容..."
                    :disabled="isExecuting"
                  />
                </el-form-item>

                <div class="action-buttons">
                  <el-button @click="resetForm" class="reset-button" :icon="RefreshRight" :disabled="isExecuting">
                    重置
                  </el-button>
                  <el-button
                    type="primary"
                    :disabled="isExecuting || !formData.content.trim()"
                    @click="startExecute"
                    class="execute-button"
                    :icon="Promotion"
                  >
                    开始执行
                  </el-button>
                </div>
              </el-form>
            </div>
          </div>

          <!-- 右侧：执行过程与结果 -->
          <div class="right-panel">
            <div class="panel-content">
              <div v-if="!isExecuting && !isCompleted && !hasError" class="empty-state">
                <el-icon class="empty-icon"><Connection /></el-icon>
                <p>输入内容后点击"开始执行"</p>
              </div>

              <div class="execution-progress" v-else>
                <div class="progress-header"
                     :class="{ 'executing': isExecuting, 'completed': isCompleted, 'error': hasError }">
                  <el-icon class="status-icon loading-icon" v-if="isExecuting">
                    <Loading />
                  </el-icon>
                  <el-icon class="status-icon success-icon" v-else-if="isCompleted">
                    <CircleCheck />
                  </el-icon>
                  <el-icon class="status-icon error-icon" v-else-if="hasError">
                    <CircleClose />
                  </el-icon>
                  <span class="status-text" v-if="isExecuting">工作流正在执行...</span>
                  <span class="status-text" v-else-if="isCompleted">执行完成！</span>
                  <span class="status-text" v-else-if="hasError">执行失败</span>
                </div>

                <div class="process-details-container">
                  <div class="process-details" ref="processDetailsContainer">
                    <!-- 节点执行日志 -->
                    <div v-if="nodeLogs.length > 0" class="node-logs">
                      <div
                        v-for="log in nodeLogs"
                        :key="log.nodeId"
                        class="node-log-item"
                        :class="{ 'expanded': expandedNodeIds.has(log.nodeId) }"
                      >
                        <div class="node-header" @click="toggleNodeExpand(log.nodeId)">
                          <div class="node-info">
                            <el-icon class="node-type-icon" :class="log.nodeType">
                              <component :is="getNodeIcon(log.nodeType)" />
                            </el-icon>
                            <span class="node-label">{{ log.nodeLabel || log.nodeType }}</span>
                          </div>
                          <span class="node-status" :class="log.nodeStatus">
                            <template v-if="log.nodeStatus === 'completed'">已完成</template>
                            <template v-else-if="log.nodeStatus === 'running'">
                              <el-icon class="status-spinner"><Loading /></el-icon>
                            </template>
                            <template v-else-if="log.nodeStatus === 'failed'">失败</template>
                            <template v-else>等待中</template>
                          </span>
                        </div>
                        <div class="node-detail" v-show="expandedNodeIds.has(log.nodeId)">
                          <!-- 执行细节时间线 -->
                          <div v-if="log.details && log.details.length > 0" class="node-details-timeline">
                            <div
                              v-for="(detail, idx) in log.details"
                              :key="idx"
                              class="detail-item"
                              :class="detail.type"
                            >
                              <!-- 工具调用 -->
                              <div v-if="detail.type === 'tool_call'" class="detail-tool-call">
                                <div class="detail-header">
                                  <el-icon class="detail-icon tool-call-icon"><Tools /></el-icon>
                                  <span class="detail-title">调用工具：{{ detail.toolName }}</span>
                                </div>
                                <pre class="detail-params">{{ formatJson(detail.toolParams) }}</pre>
                              </div>
                              <!-- 工具结果 -->
                              <div v-else-if="detail.type === 'tool_result'" class="detail-tool-result">
                                <div class="detail-header">
                                  <el-icon class="detail-icon tool-result-icon"><CircleCheck /></el-icon>
                                  <span class="detail-title">工具结果：{{ detail.toolName }}</span>
                                </div>
                                <pre class="detail-result">{{ truncateText(detail.toolResult, 500) }}</pre>
                              </div>
                              <!-- 进度消息 -->
                              <div v-else-if="detail.type === 'progress'" class="detail-progress">
                                <el-icon class="detail-icon progress-icon"><Loading /></el-icon>
                                <span class="detail-msg">{{ detail.showMsg }}</span>
                              </div>
                              <!-- LLM 输出片段 -->
                              <div v-else-if="detail.type === 'content'" class="detail-content">
                                <div class="detail-header">
                                  <el-icon class="detail-icon content-icon"><ChatDotRound /></el-icon>
                                  <span class="detail-title">LLM 输出</span>
                                </div>
                                <pre class="detail-content-text">{{ truncateText(detail.content, 300) }}</pre>
                              </div>
                            </div>
                          </div>

                          <!-- 最终执行日志 -->
                          <div v-if="log.executionLog" class="exec-log">
                            <div v-if="log.executionLog.input" class="log-section">
                              <span class="log-label">输入：</span>
                              <pre class="log-content">{{ log.executionLog.input }}</pre>
                            </div>
                            <div v-if="log.executionLog.output" class="log-section">
                              <span class="log-label">输出：</span>
                              <pre class="log-content">{{ log.executionLog.output }}</pre>
                            </div>
                            <div v-if="log.executionLog.error" class="log-section error">
                              <span class="log-label">错误：</span>
                              <pre class="log-content">{{ log.executionLog.error }}</pre>
                            </div>
                            <div v-if="log.executionLog.duration" class="log-meta">
                              耗时：{{ formatDuration(log.executionLog.duration) }}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <!-- 流式输出内容 -->
                    <div v-if="streamContent" class="stream-output">
                      <h3 class="output-title">输出结果</h3>
                      <div class="markdown-body" v-html="renderedContent"></div>
                    </div>

                    <!-- 文件产物下载 -->
                    <div v-if="artifacts.length > 0" class="artifacts-section">
                      <h3 class="output-title">
                        <el-icon><Files /></el-icon>
                        生成文件
                      </h3>
                      <div class="artifacts-list">
                        <div
                          v-for="(file, idx) in artifacts"
                          :key="idx"
                          class="artifact-item"
                        >
                          <div class="artifact-info">
                            <el-icon class="artifact-icon" :class="file.category">
                              <component :is="getFileIcon(file.category)" />
                            </el-icon>
                            <div class="artifact-detail">
                              <span class="artifact-name">{{ file.name }}</span>
                              <span class="artifact-type">{{ getFileTypeLabel(file.category) }}</span>
                            </div>
                          </div>
                          <el-button
                            type="primary"
                            size="small"
                            class="download-btn"
                            @click="downloadFile(file.url, file.name)"
                          >
                            <el-icon><Download /></el-icon>
                            下载
                          </el-button>
                        </div>
                      </div>
                    </div>

                    <!-- 执行中加载动画 -->
                    <div v-if="isExecuting && nodeLogs.length === 0" class="center-loading">
                      <el-icon class="loading-icon"><Loading /></el-icon>
                      <span>正在连接...</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft,
  CircleCheck,
  CircleClose,
  Connection,
  Loading,
  Promotion,
  RefreshRight,
  VideoPlay,
  VideoPause,
  Setting,
  Tools,
  Edit,
  Switch,
  ChatDotRound,
  Download,
  Document,
  Picture,
  Files
} from '@element-plus/icons-vue'
import { marked } from 'marked'
import fetchSSEWithAuth from '@/api/sseRequest.js'
import { defaultApi } from '@/api'
import { getAuthHeaders } from '@/api/auth'

defineOptions({
  name: 'WorkflowExecute'
})

const route = useRoute()
const router = useRouter()

const workflowId = computed(() => route.params.id)

// 工作流信息
const workflowInfo = ref({ name: '', description: '' })

// 表单数据
const formData = ref({ content: '' })

// 执行状态
const isExecuting = ref(false)
const isCompleted = ref(false)
const hasError = ref(false)

// 节点日志
const nodeLogs = ref([])
const expandedNodeIds = ref(new Set())

// 流式输出
const streamContent = ref('')
const processDetailsContainer = ref(null)

// 文件产物
const artifacts = ref([])

let abortController = null

// 渲染 Markdown
const renderedContent = computed(() => {
  try {
    return marked(streamContent.value || '')
  } catch {
    return streamContent.value
  }
})

// 获取节点图标
const getNodeIcon = (nodeType) => {
  const iconMap = {
    start: VideoPlay,
    end: VideoPause,
    llm: Edit,
    tool: Tools,
    condition: Switch,
    code: Setting,
    http: Connection,
    subworkflow: Connection
  }
  return iconMap[nodeType] || Setting
}

// 格式化耗时
const formatDuration = (ms) => {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

// 格式化 JSON
const formatJson = (str) => {
  if (!str) return ''
  try {
    const obj = typeof str === 'string' ? JSON.parse(str) : str
    return JSON.stringify(obj, null, 2)
  } catch {
    return str
  }
}

// 截断文本
const truncateText = (text, maxLen = 500) => {
  if (!text) return ''
  if (text.length <= maxLen) return text
  return text.substring(0, maxLen) + '...'
}

// 获取文件类型图标
const getFileIcon = (category) => {
  const iconMap = {
    pdf: Document,
    markdown: Document,
    image: Picture
  }
  return iconMap[category] || Files
}

// 获取文件类型标签
const getFileTypeLabel = (category) => {
  const labelMap = {
    pdf: 'PDF 文档',
    markdown: 'Markdown 文件',
    image: '图片文件'
  }
  return labelMap[category] || '文件'
}

// 展开/折叠节点
const toggleNodeExpand = (nodeId) => {
  if (expandedNodeIds.value.has(nodeId)) {
    expandedNodeIds.value.delete(nodeId)
  } else {
    expandedNodeIds.value.add(nodeId)
  }
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (processDetailsContainer.value) {
      processDetailsContainer.value.scrollTop = processDetailsContainer.value.scrollHeight
    }
  })
}

// 加载工作流信息
const loadWorkflowInfo = async () => {
  try {
    const res = await defaultApi.apiWorkflowPublicIdGet(workflowId.value)
    if (res.code === 0 && res.data) {
      workflowInfo.value = res.data
    } else {
      ElMessage.error(res.msg || '获取工作流信息失败')
    }
  } catch (error) {
    console.error('Load workflow info error:', error)
    ElMessage.error('获取工作流信息失败')
  }
}

// 开始执行
const startExecute = async () => {
  if (!formData.value.content.trim()) return

  // 重置状态
  isExecuting.value = true
  isCompleted.value = false
  hasError.value = false
  nodeLogs.value = []
  expandedNodeIds.value.clear()
  streamContent.value = ''
  artifacts.value = []

  try {
    const url = `/api/workflow/public/${workflowId.value}/run`
    const data = { content: formData.value.content }

    abortController = await fetchSSEWithAuth(url, data, function (msg) {
      if (!msg.startsWith('data:')) return
      const payload = msg.slice(5).trim()
      if (!payload) return

      try {
        const data = JSON.parse(payload)

        // 处理节点状态更新
        if (data.nodeId) {
          updateNodeLog(data)
        }

        // 处理流式内容
        if (data.content) {
          streamContent.value += data.content
          scrollToBottom()
        }

        // 处理结束标记
        if (data.end) {
          isExecuting.value = false
          if (data.error) {
            hasError.value = true
            ElMessage.error('执行失败：' + data.error)
          } else {
            isCompleted.value = true
          }
          // 收集文件产物
          if (data.artifacts && Array.isArray(data.artifacts)) {
            artifacts.value = data.artifacts
          }
          try { abortController?.abort() } catch (e) {}
          abortController = null
        }
      } catch (e) {
        console.error('处理SSE消息出错:', e)
      }
    }, function (error) {
      isExecuting.value = false
      hasError.value = true
      try { abortController?.abort() } catch (e) {}
      abortController = null
      ElMessage.error('执行出错，请稍后重试')
      console.error('SSE请求出错:', error)
    }, function () {
      console.log('SSE连接关闭')
    })
  } catch (error) {
    isExecuting.value = false
    hasError.value = true
    ElMessage.error('执行失败，请稍后重试')
    console.error('Execute workflow error:', error)
  }
}

// 更新节点日志
const updateNodeLog = (data) => {
  const existingIndex = nodeLogs.value.findIndex(log => log.nodeId === data.nodeId)
  if (existingIndex >= 0) {
    // 更新已有节点
    const existing = nodeLogs.value[existingIndex]
    const details = [...(existing.details || [])]

    // 收集执行细节事件
    if (data.toolName && data.toolParams && !data.toolResult) {
      // 工具调用请求
      details.push({
        type: 'tool_call',
        toolName: data.toolName,
        toolParams: data.toolParams,
        showMsg: data.showMsg || '',
        timestamp: Date.now()
      })
    } else if (data.toolName && data.toolResult) {
      // 工具执行结果
      details.push({
        type: 'tool_result',
        toolName: data.toolName,
        toolResult: data.toolResult,
        showMsg: data.showMsg || '',
        timestamp: Date.now()
      })
    } else if (data.showMsg && !data.nodeStatus) {
      // 进度消息（没有 nodeStatus 的纯 ShowMsg）
      details.push({
        type: 'progress',
        showMsg: data.showMsg,
        timestamp: Date.now()
      })
    } else if (data.content && data.nodeId) {
      // LLM 输出片段
      details.push({
        type: 'content',
        content: data.content,
        timestamp: Date.now()
      })
    }

    nodeLogs.value[existingIndex] = {
      ...existing,
      nodeStatus: data.nodeStatus || existing.nodeStatus,
      executionLog: data.execution_log || existing.executionLog,
      details
    }
  } else {
    // 添加新节点
    const details = []
    // 收集首条消息的细节
    if (data.toolName && data.toolParams) {
      details.push({
        type: 'tool_call',
        toolName: data.toolName,
        toolParams: data.toolParams,
        showMsg: data.showMsg || '',
        timestamp: Date.now()
      })
    } else if (data.showMsg && !data.nodeStatus) {
      details.push({
        type: 'progress',
        showMsg: data.showMsg,
        timestamp: Date.now()
      })
    }

    nodeLogs.value.push({
      nodeId: data.nodeId,
      nodeType: data.nodeType || 'unknown',
      nodeLabel: data.nodeLabel || '',
      nodeStatus: data.nodeStatus || 'running',
      executionLog: data.execution_log || null,
      details
    })
    // 自动展开正在运行的节点
    expandedNodeIds.value.add(data.nodeId)
  }
  scrollToBottom()
}

// 重置表单
const resetForm = () => {
  formData.value.content = ''
  isExecuting.value = false
  isCompleted.value = false
  hasError.value = false
  nodeLogs.value = []
  expandedNodeIds.value.clear()
  streamContent.value = ''
  artifacts.value = []
  try { abortController?.abort() } catch (e) {}
  abortController = null
}

// 返回市场
const goBack = () => {
  router.push('/workflow')
}

// 下载文件（带认证）
const downloadFile = async (url, name) => {
  try {
    const headers = getAuthHeaders()
    const response = await fetch(url, { headers })
    if (!response.ok) {
      ElMessage.error('下载失败：' + response.statusText)
      return
    }
    const blob = await response.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch (e) {
    console.error('下载文件失败:', e)
    ElMessage.error('下载文件失败')
  }
}

onMounted(() => {
  loadWorkflowInfo()
})

onBeforeUnmount(() => {
  try { abortController?.abort() } catch (e) {}
})
</script>

<style lang="scss" scoped>
.workflow-exec-container {
  min-height: 100vh;
  background: var(--el-bg-color);
  padding: 20px;
}

.workflow-exec-content {
  max-width: 1400px;
  margin: 0 auto;
}

.exec-card {
  border-radius: 16px;
  overflow: hidden;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.two-column {
  display: flex;
  min-height: calc(100vh - 80px);
}

.left-panel {
  width: 400px;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-light);
  padding: 24px;
  overflow-y: auto;
}

.right-panel {
  flex: 1;
  overflow: hidden;
}

.panel-section {
  margin-bottom: 24px;

  &.page-intro {
    margin-bottom: 32px;
  }
}

.back-button {
  margin-bottom: 12px;
  padding: 0;
  color: var(--el-text-color-secondary);

  &:hover {
    color: var(--el-color-primary);
  }
}

.intro-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: var(--el-text-color-primary);
}

.intro-subtitle {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin: 0;
  line-height: 1.6;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: var(--el-text-color-primary);
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: 24px;

  .reset-button {
    flex: 1;
  }

  .execute-button {
    flex: 2;
  }
}

.panel-content {
  height: 100%;
  overflow-y: auto;
  padding: 24px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--el-text-color-secondary);

  .empty-icon {
    font-size: 64px;
    margin-bottom: 16px;
    opacity: 0.3;
  }

  p {
    font-size: 16px;
  }
}

.execution-progress {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-weight: 500;

  &.executing {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
  }

  &.completed {
    background: var(--el-color-success-light-9);
    color: var(--el-color-success);
  }

  &.error {
    background: var(--el-color-danger-light-9);
    color: var(--el-color-danger);
  }

  .status-spinner {
    animation: spin 1s linear infinite;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.process-details-container {
  flex: 1;
  overflow: hidden;
}

.process-details {
  height: 100%;
  overflow-y: auto;
  padding-right: 8px;
}

.node-logs {
  margin-bottom: 20px;
}

.node-log-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  margin-bottom: 8px;
  overflow: hidden;
  transition: all 0.2s;

  &:hover {
    border-color: var(--el-border-color);
  }

  &.expanded {
    border-color: var(--el-color-primary-light-5);
  }
}

.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;

  &:hover {
    background: var(--el-fill-color-lighter);
  }
}

.node-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-type-icon {
  font-size: 16px;
  color: var(--el-color-primary);

  &.start { color: var(--el-color-success); }
  &.end { color: var(--el-color-info); }
  &.llm { color: var(--el-color-primary); }
  &.tool { color: var(--el-color-warning); }
  &.condition { color: var(--el-color-danger); }
}

.node-label {
  font-size: 14px;
  font-weight: 500;
}

.node-status {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;

  &.completed { color: var(--el-color-success); }
  &.running { color: var(--el-color-primary); }
  &.failed { color: var(--el-color-danger); }
}

.node-detail {
  border-top: 1px solid var(--el-border-color-lighter);
  padding: 12px 16px;
  background: var(--el-fill-color-blank);
}

.exec-log {
  .log-section {
    margin-bottom: 12px;

    &.error .log-content {
      color: var(--el-color-danger);
    }
  }

  .log-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--el-text-color-secondary);
    display: block;
    margin-bottom: 4px;
  }

  .log-content {
    font-size: 13px;
    line-height: 1.5;
    background: var(--el-fill-color-light);
    padding: 8px 12px;
    border-radius: 6px;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 200px;
    overflow-y: auto;
  }

  .log-meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

// 节点执行细节时间线
.node-details-timeline {
  margin-bottom: 12px;
  border-left: 2px solid var(--el-border-color-lighter);
  padding-left: 12px;

  .detail-item {
    margin-bottom: 8px;
    padding: 6px 0;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .detail-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
  }

  .detail-icon {
    font-size: 14px;
    flex-shrink: 0;

    &.tool-call-icon { color: var(--el-color-warning); }
    &.tool-result-icon { color: var(--el-color-success); }
    &.progress-icon { color: var(--el-color-primary); animation: spin 1s linear infinite; }
    &.content-icon { color: var(--el-color-primary); }
  }

  .detail-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }

  .detail-msg {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .detail-params,
  .detail-result,
  .detail-content-text {
    font-size: 12px;
    line-height: 1.4;
    background: var(--el-fill-color-lighter);
    padding: 6px 10px;
    border-radius: 4px;
    margin: 4px 0 0 20px;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 150px;
    overflow-y: auto;
    color: var(--el-text-color-regular);
  }

  .detail-tool-call {
    .detail-params {
      border-left: 2px solid var(--el-color-warning-light-3);
    }
  }

  .detail-tool-result {
    .detail-result {
      border-left: 2px solid var(--el-color-success-light-3);
    }
  }

  .detail-content {
    .detail-content-text {
      border-left: 2px solid var(--el-color-primary-light-3);
    }
  }
}

.stream-output {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 2px solid var(--el-border-color-light);

  .output-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 16px 0;
  }

  .markdown-body {
    font-size: 15px;
    line-height: 1.7;
  }
}

.artifacts-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 2px solid var(--el-border-color-light);

  .output-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 16px 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .artifacts-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .artifact-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--el-fill-color-lighter);
    border-radius: 10px;
    border: 1px solid var(--el-border-color-lighter);
    transition: all 0.2s;

    &:hover {
      border-color: var(--el-color-primary-light-5);
      background: var(--el-color-primary-light-9);
    }
  }

  .artifact-info {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }

  .artifact-icon {
    font-size: 24px;
    flex-shrink: 0;

    &.pdf { color: var(--el-color-danger); }
    &.markdown { color: var(--el-color-primary); }
    &.image { color: var(--el-color-success); }
  }

  .artifact-detail {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .artifact-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .artifact-type {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .download-btn {
    flex-shrink: 0;
    border-radius: 8px;
  }
}

.center-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--el-text-color-secondary);
  gap: 12px;

  .loading-icon {
    font-size: 32px;
    animation: spin 1s linear infinite;
  }
}

@media (max-width: 768px) {
  .two-column {
    flex-direction: column;
  }

  .left-panel {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-light);
  }
}
</style>
