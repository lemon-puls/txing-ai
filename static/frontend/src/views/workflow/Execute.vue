<template>
  <div class="workflow-exec-container">
    <!-- 顶部导航条 -->
    <div class="exec-topbar">
      <el-button text class="back-button" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回市场
      </el-button>
      <div class="topbar-center">
        <span class="topbar-title">{{ workflowInfo.name || '加载中...' }}</span>
      </div>
      <div class="topbar-right"></div>
    </div>

    <div class="exec-body">
      <!-- 左侧面板 -->
      <div class="left-panel">
        <div class="panel-section intro-section">
          <h2 class="intro-title">{{ workflowInfo.name || '加载中...' }}</h2>
          <p class="intro-subtitle">{{ workflowInfo.description || '请输入内容开始执行工作流' }}</p>
        </div>

        <!-- 动态输入表单（根据 inputSchema 渲染） -->
        <template v-if="inputSchema.length > 0">
          <div class="panel-section" v-for="field in inputSchema" :key="field.name">
            <h3 class="section-title">
              <el-icon><component :is="getFieldIcon(field.type)" /></el-icon>
              {{ field.label || field.name }}
              <span class="title-hint" v-if="!field.required">（可选）</span>
              <span class="title-required" v-else>（必填）</span>
            </h3>

            <!-- 文件上传字段 -->
            <template v-if="field.type === 'file'">
              <div
                class="upload-area"
                :class="{ 'has-file': uploadedFiles[field.name], 'is-dragover': dragoverField === field.name }"
                @dragover.prevent="dragoverField = field.name"
                @dragleave.prevent="dragoverField = null"
                @drop.prevent="handleFieldDrop($event, field.name)"
              >
                <div v-if="!uploadedFiles[field.name]" class="upload-placeholder" @click="triggerFieldFileInput(field.name)">
                  <el-icon class="upload-icon"><Upload /></el-icon>
                  <p class="upload-text">{{ field.placeholder || '拖拽文件到此处，或点击选择' }}</p>
                  <p class="upload-hint" v-if="field.accept">支持格式：{{ field.accept }}</p>
                  <p class="upload-hint">最大 10MB</p>
                </div>
                <div v-else class="uploaded-file">
                  <div class="file-info">
                    <el-icon class="file-icon"><Document /></el-icon>
                    <div class="file-detail">
                      <span class="file-name">{{ uploadedFiles[field.name].name }}</span>
                      <span class="file-size">{{ formatFileSize(uploadedFiles[field.name].size) }}</span>
                    </div>
                  </div>
                  <el-button text class="remove-file" @click="removeFieldFile(field.name)" :disabled="isExecuting">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
              <input :ref="el => setFileInputRef(field.name, el)" type="file" :accept="field.accept || '.pdf,.txt,.md'" style="display:none" @change="handleFieldFileSelect($event, field.name)" />
              <p class="field-description" v-if="field.description">{{ field.description }}</p>
            </template>

            <!-- 文本域字段 -->
            <template v-else-if="field.type === 'textarea'">
              <el-input
                v-model="formFields[field.name]"
                type="textarea"
                :rows="4"
                :placeholder="field.placeholder || '请输入...'"
                :disabled="isExecuting"
                class="input-textarea"
              />
              <p class="field-description" v-if="field.description">{{ field.description }}</p>
            </template>

            <!-- 单行文本字段 -->
            <template v-else>
              <el-input
                v-model="formFields[field.name]"
                :placeholder="field.placeholder || '请输入...'"
                :disabled="isExecuting"
              />
              <p class="field-description" v-if="field.description">{{ field.description }}</p>
            </template>
          </div>
        </template>

        <!-- 无 inputSchema 时显示默认输入框 -->
        <template v-else>
          <div class="panel-section">
            <h3 class="section-title">
              <el-icon><Edit /></el-icon>
              输入内容
            </h3>
            <el-input
              v-model="formData.content"
              type="textarea"
              :rows="8"
              placeholder="请输入您想要处理的内容..."
              :disabled="isExecuting"
              class="input-textarea"
            />
          </div>

          <!-- 默认文件上传区域 -->
          <div class="panel-section">
            <h3 class="section-title">
              <el-icon><Upload /></el-icon>
              上传文件
              <span class="title-hint">（可选）</span>
            </h3>
            <div
              class="upload-area"
              :class="{ 'has-file': uploadedFile, 'is-dragover': isDragover }"
              @dragover.prevent="isDragover = true"
              @dragleave.prevent="isDragover = false"
              @drop.prevent="handleDrop"
            >
              <div v-if="!uploadedFile" class="upload-placeholder" @click="triggerFileInput">
                <el-icon class="upload-icon"><Upload /></el-icon>
                <p class="upload-text">拖拽文件到此处，或 <span class="upload-link">点击选择</span></p>
                <p class="upload-hint">支持 PDF、TXT、MD 格式，最大 10MB</p>
              </div>
              <div v-else class="uploaded-file">
                <div class="file-info">
                  <el-icon class="file-icon"><Document /></el-icon>
                  <div class="file-detail">
                    <span class="file-name">{{ uploadedFile.name }}</span>
                    <span class="file-size">{{ formatFileSize(uploadedFile.size) }}</span>
                  </div>
                </div>
                <el-button text class="remove-file" @click="removeFile" :disabled="isExecuting">
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
          <input
            ref="fileInputRef"
            type="file"
            accept=".pdf,.txt,.md"
            style="display: none"
            @change="handleFileSelect"
          />
        </template>

        <div class="action-buttons">
          <el-button @click="resetForm" class="reset-button" :icon="RefreshRight" :disabled="isExecuting">
            重置
          </el-button>
          <el-button
            type="primary"
            :disabled="isExecuting || (!canExecute)"
            @click="startExecute"
            class="execute-button"
            :icon="Promotion"
          >
            {{ isExecuting ? '执行中...' : '开始执行' }}
          </el-button>
        </div>
      </div>

      <!-- 右侧面板 -->
      <div class="right-panel">
        <div class="panel-content" ref="processDetailsContainer">
          <!-- 空状态 -->
          <div v-if="!isExecuting && !isCompleted && !hasError" class="empty-state">
            <div class="empty-icon-wrapper">
              <el-icon class="empty-icon"><Connection /></el-icon>
            </div>
            <p class="empty-title">准备就绪</p>
            <p class="empty-desc">在左侧输入内容后点击"开始执行"</p>
          </div>

          <!-- 执行进度 -->
          <div class="execution-progress" v-else>
            <!-- 状态头部 -->
            <div class="progress-header" :class="{ 'executing': isExecuting, 'completed': isCompleted, 'error': hasError }">
              <div class="header-left">
                <el-icon class="status-icon" v-if="isExecuting"><Loading /></el-icon>
                <el-icon class="status-icon" v-else-if="isCompleted"><CircleCheck /></el-icon>
                <el-icon class="status-icon" v-else-if="hasError"><CircleClose /></el-icon>
                <span class="status-text" v-if="isExecuting">工作流正在执行...</span>
                <span class="status-text" v-else-if="isCompleted">执行完成</span>
                <span class="status-text" v-else-if="hasError">执行失败</span>
              </div>
              <div class="header-right" v-if="nodeLogs.length > 0">
                <span class="node-count">{{ nodeLogs.filter(l => l.nodeStatus === 'completed').length }} / {{ nodeLogs.length }} 节点</span>
              </div>
            </div>

            <!-- 节点执行日志 -->
            <div v-if="nodeLogs.length > 0" class="node-logs">
              <div
                v-for="log in nodeLogs"
                :key="log.nodeId"
                class="node-log-item"
                :class="{ 'expanded': expandedNodeIds.has(log.nodeId), 'is-running': log.nodeStatus === 'running' }"
              >
                <div class="node-header" @click="toggleNodeExpand(log.nodeId)">
                  <div class="node-info">
                    <div class="node-icon-wrapper" :class="log.nodeType">
                      <el-icon class="node-type-icon"><component :is="getNodeIcon(log.nodeType)" /></el-icon>
                    </div>
                    <div class="node-meta">
                      <span class="node-label">{{ log.nodeLabel || log.nodeType }}</span>
                      <span class="node-duration" v-if="log.executionLog?.duration">
                        {{ formatDuration(log.executionLog.duration) }}
                      </span>
                    </div>
                  </div>
                  <div class="node-right">
                    <span class="node-status" :class="log.nodeStatus">
                      <template v-if="log.nodeStatus === 'completed'">
                        <el-icon><CircleCheck /></el-icon> 已完成
                      </template>
                      <template v-else-if="log.nodeStatus === 'running'">
                        <el-icon class="status-spinner"><Loading /></el-icon> 执行中
                      </template>
                      <template v-else-if="log.nodeStatus === 'failed'">
                        <el-icon><CircleClose /></el-icon> 失败
                      </template>
                      <template v-else>等待中</template>
                    </span>
                    <el-icon class="expand-arrow" :class="{ 'is-expanded': expandedNodeIds.has(log.nodeId) }">
                      <ArrowDown />
                    </el-icon>
                  </div>
                </div>

                <!-- 节点详情 -->
                <div class="node-detail" v-show="expandedNodeIds.has(log.nodeId)">
                  <!-- 执行细节时间线 -->
                  <div v-if="log.details && log.details.length > 0" class="node-details-timeline">
                    <div
                      v-for="(detail, idx) in log.details"
                      :key="idx"
                      class="detail-item"
                      :class="[detail.type, detail.status]"
                    >
                      <!-- 工具调用组（合并显示调用 + 结果） -->
                      <div v-if="detail.type === 'tool_call_group'" class="detail-tool-group">
                        <div
                          class="tool-group-header"
                          @click="detail.status === 'completed' && toggleToolGroup(`${log.nodeId}_${idx}`)"
                          :class="{ clickable: detail.status === 'completed' }"
                        >
                          <div class="tool-group-left">
                            <!-- 进行中：转圈动画 -->
                            <el-icon v-if="detail.status === 'running'" class="tool-spinner">
                              <Loading />
                            </el-icon>
                            <!-- 已完成：绿色勾 -->
                            <el-icon v-else class="tool-done"><CircleCheck /></el-icon>
                            <el-tag size="small" :type="detail.status === 'running' ? 'warning' : 'success'" effect="plain" round>
                              {{ detail.toolName }}
                            </el-tag>
                            <span class="tool-group-summary" v-if="detail.status === 'completed' && !isToolGroupExpanded(`${log.nodeId}_${idx}`)">
                              {{ getToolResultSummary(detail) }}
                            </span>
                          </div>
                          <el-icon
                            v-if="detail.status === 'completed'"
                            class="tool-expand-arrow"
                            :class="{ 'is-expanded': isToolGroupExpanded(`${log.nodeId}_${idx}`) }"
                          >
                            <ArrowDown />
                          </el-icon>
                        </div>
                        <!-- 展开详情 -->
                        <div
                          class="tool-group-body"
                          v-show="detail.status === 'running' || isToolGroupExpanded(`${log.nodeId}_${idx}`)"
                        >
                          <div v-if="detail.toolParams" class="tool-section">
                            <span class="tool-section-label">参数</span>
                            <pre class="detail-code">{{ formatJson(detail.toolParams) }}</pre>
                          </div>
                          <div v-if="detail.toolResult" class="tool-section">
                            <span class="tool-section-label">结果</span>
                            <pre class="detail-code result-code" :class="{ 'json-code': detail.toolResultIsJson }">{{ truncateText(detail.toolResultFormatted || detail.toolResult, 800) }}</pre>
                          </div>
                          <div v-if="detail.status === 'running' && !detail.toolResult" class="tool-running-hint">
                            <span class="tool-running-dot"></span>
                            <span>等待返回结果...</span>
                          </div>
                        </div>
                      </div>

                      <div v-else-if="detail.type === 'progress'" class="detail-progress">
                        <div class="detail-dot progress-dot"></div>
                        <span class="detail-msg">{{ detail.showMsg }}</span>
                      </div>

                      <div v-else-if="detail.type === 'content'" class="detail-content">
                        <div class="detail-header">
                          <div class="detail-dot content-dot"></div>
                          <span class="detail-title">LLM 输出</span>
                        </div>
                        <pre class="detail-code content-code">{{ truncateText(detail.content, 300) }}</pre>
                      </div>
                    </div>
                  </div>

                  <!-- 最终执行日志 -->
                  <div v-if="log.executionLog" class="exec-log">
                    <div v-if="log.executionLog.input" class="log-section">
                      <span class="log-label">输入</span>
                      <pre class="log-content">{{ log.executionLog.input }}</pre>
                    </div>
                    <div v-if="log.executionLog.output" class="log-section">
                      <span class="log-label">输出</span>
                      <pre class="log-content">{{ log.executionLog.output }}</pre>
                    </div>
                    <div v-if="log.executionLog.error" class="log-section error">
                      <span class="log-label">错误</span>
                      <pre class="log-content">{{ log.executionLog.error }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 流式输出内容 -->
            <div v-if="streamContent" class="stream-output">
              <h3 class="output-title">
                <el-icon><Document /></el-icon>
                输出结果
              </h3>
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
                    <div class="artifact-icon-wrapper" :class="file.category">
                      <el-icon :size="20"><component :is="getFileIcon(file.category)" /></el-icon>
                    </div>
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

            <!-- 加载动画 -->
            <div v-if="isExecuting && nodeLogs.length === 0" class="center-loading">
              <div class="loading-spinner"></div>
              <span>正在连接...</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, ArrowDown, CircleCheck, CircleClose, Connection, Loading,
  Promotion, RefreshRight, VideoPlay, VideoPause, Setting, Tools,
  Edit, Switch, ChatDotRound, Download, Document, Picture, Files, Upload, Close
} from '@element-plus/icons-vue'
import { marked } from 'marked'
import fetchSSEWithAuth from '@/api/sseRequest.js'
import { defaultApi } from '@/api'
import { getAuthHeaders } from '@/api/auth'

defineOptions({ name: 'WorkflowExecute' })

const route = useRoute()
const router = useRouter()
const workflowId = computed(() => route.params.id)

const workflowInfo = ref({ name: '', description: '' })
const formData = ref({ content: '' })
const isExecuting = ref(false)
const uploadedFile = ref(null)
const isDragover = ref(false)
const fileInputRef = ref(null)
const isCompleted = ref(false)
const hasError = ref(false)
const nodeLogs = ref([])
const expandedNodeIds = ref(new Set())
const streamContent = ref('')
const processDetailsContainer = ref(null)
const artifacts = ref([])

// 动态表单相关
const inputSchema = ref([])
const formFields = ref({})
const uploadedFiles = ref({})
const dragoverField = ref(null)
const fileInputRefs = {}

let abortController = null

const renderedContent = computed(() => {
  try { return marked(streamContent.value || '') } catch { return streamContent.value }
})

// 是否可以执行
const canExecute = computed(() => {
  if (inputSchema.value.length > 0) {
    // 有 inputSchema 时：检查必填字段
    for (const field of inputSchema.value) {
      if (field.required) {
        if (field.type === 'file') {
          if (!uploadedFiles.value[field.name]) return false
        } else {
          if (!formFields.value[field.name]?.trim()) return false
        }
      }
    }
    // 至少有一个字段有值
    const hasTextField = Object.values(formFields.value).some(v => v?.trim())
    const hasFile = Object.keys(uploadedFiles.value).length > 0
    return hasTextField || hasFile
  }
  // 无 inputSchema 时：需要内容或文件
  return formData.value.content.trim() || uploadedFile.value
})

const getNodeIcon = (nodeType) => {
  const map = { start: VideoPlay, end: VideoPause, llm: Edit, tool: Tools, agent: ChatDotRound, condition: Switch, code: Setting, http: Connection }
  return map[nodeType] || Setting
}

const formatDuration = (ms) => {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

const formatJson = (str) => {
  if (!str) return ''
  try { return JSON.stringify(typeof str === 'string' ? JSON.parse(str) : str, null, 2) } catch { return str }
}

const truncateText = (text, maxLen = 500) => {
  if (!text) return ''
  return text.length <= maxLen ? text : text.substring(0, maxLen) + '...'
}

// 获取工具调用结果的简短摘要（折叠状态下显示）
const getToolResultSummary = (detail) => {
  if (!detail.toolResult) return ''
  const text = detail.toolResultFormatted || detail.toolResult
  // 取第一行有意义的内容，截断到 60 字符
  const firstLine = text.split('\n').find(l => l.trim()) || ''
  const clean = firstLine.replace(/^["'{\s]+|["'}\s]+$/g, '').trim()
  return clean.length > 60 ? clean.substring(0, 60) + '...' : clean
}

// 动态表单相关
const getFieldIcon = (type) => {
  const map = { file: Upload, textarea: Edit, text: Edit }
  return map[type] || Edit
}

const setFileInputRef = (name, el) => {
  if (el) fileInputRefs[name] = el
}

const triggerFieldFileInput = (name) => {
  fileInputRefs[name]?.click()
}

const handleFieldFileSelect = (e, name) => {
  const file = e.target.files[0]
  if (file) setFieldUploadedFile(name, file)
}

const handleFieldDrop = (e, name) => {
  dragoverField.value = null
  const file = e.dataTransfer.files[0]
  if (file) setFieldUploadedFile(name, file)
}

const setFieldUploadedFile = (name, file) => {
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.warning('文件大小不能超过 10MB')
    return
  }
  uploadedFiles.value[name] = file
}

const removeFieldFile = (name) => {
  delete uploadedFiles.value[name]
  if (fileInputRefs[name]) fileInputRefs[name].value = ''
}

// 解析 inputSchema
const parseInputSchema = (topology) => {
  try {
    const topo = typeof topology === 'string' ? JSON.parse(topology) : topology
    if (topo.config && Array.isArray(topo.config.inputSchema) && topo.config.inputSchema.length > 0) {
      inputSchema.value = topo.config.inputSchema
      // 初始化 formFields
      const fields = {}
      topo.config.inputSchema.forEach(f => {
        if (f.type !== 'file') {
          fields[f.name] = f.default || ''
        }
      })
      formFields.value = fields
    } else {
      inputSchema.value = []
      formFields.value = {}
    }
  } catch {
    inputSchema.value = []
    formFields.value = {}
  }
}

const getFileIcon = (category) => {
  const map = { pdf: Document, markdown: Document, image: Picture }
  return map[category] || Files
}

const getFileTypeLabel = (category) => {
  const map = { pdf: 'PDF 文档', markdown: 'Markdown 文件', image: '图片文件' }
  return map[category] || '文件'
}

// 文件上传相关
const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = (e) => {
  const file = e.target.files[0]
  if (file) setUploadedFile(file)
}

const handleDrop = (e) => {
  isDragover.value = false
  const file = e.dataTransfer.files[0]
  if (file) setUploadedFile(file)
}

const setUploadedFile = (file) => {
  // 校验文件类型
  const allowedTypes = ['.pdf', '.txt', '.md']
  const ext = '.' + file.name.split('.').pop().toLowerCase()
  if (!allowedTypes.includes(ext)) {
    ElMessage.warning('仅支持 PDF、TXT、MD 格式')
    return
  }
  // 校验文件大小 (10MB)
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.warning('文件大小不能超过 10MB')
    return
  }
  uploadedFile.value = file
}

const removeFile = () => {
  uploadedFile.value = null
  if (fileInputRef.value) fileInputRef.value.value = ''
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const toggleNodeExpand = (nodeId) => {
  if (expandedNodeIds.value.has(nodeId)) expandedNodeIds.value.delete(nodeId)
  else expandedNodeIds.value.add(nodeId)
}

// 工具调用组的展开/折叠状态（key = nodeId_toolCallIndex）
const expandedToolGroups = ref(new Set())

const toggleToolGroup = (key) => {
  if (expandedToolGroups.value.has(key)) expandedToolGroups.value.delete(key)
  else expandedToolGroups.value.add(key)
}

const isToolGroupExpanded = (key) => expandedToolGroups.value.has(key)

const scrollToBottom = () => {
  nextTick(() => {
    if (processDetailsContainer.value) processDetailsContainer.value.scrollTop = processDetailsContainer.value.scrollHeight
  })
}

const loadWorkflowInfo = async () => {
  try {
    const res = await defaultApi.apiWorkflowPublicIdGet(workflowId.value)
    if (res.code === 0 && res.data) {
      workflowInfo.value = res.data
      // 解析 inputSchema
      if (res.data.topology) {
        parseInputSchema(res.data.topology)
      }
    }
    else ElMessage.error(res.msg || '获取工作流信息失败')
  } catch { ElMessage.error('获取工作流信息失败') }
}

const startExecute = async () => {
  isExecuting.value = true
  isCompleted.value = false
  hasError.value = false
  nodeLogs.value = []
  expandedNodeIds.value.clear()
  streamContent.value = ''
  artifacts.value = []

  // 构建请求数据
  let requestData
  if (inputSchema.value.length > 0) {
    // 有 inputSchema 时使用 FormData
    const fd = new FormData()
    // 添加文本字段
    Object.entries(formFields.value).forEach(([key, value]) => {
      if (value) fd.append(key, value)
    })
    // 添加文件字段
    Object.entries(uploadedFiles.value).forEach(([key, file]) => {
      fd.append(key, file)
    })
    requestData = fd
  } else if (uploadedFile.value) {
    // 无 inputSchema 但有文件
    const fd = new FormData()
    fd.append('file', uploadedFile.value)
    if (formData.value.content.trim()) fd.append('content', formData.value.content)
    requestData = fd
  } else {
    requestData = { content: formData.value.content }
  }

  try {
    abortController = await fetchSSEWithAuth(`/api/workflow/public/${workflowId.value}/run`, requestData, function (msg) {
      if (!msg.startsWith('data:')) return
      const payload = msg.slice(5).trim()
      if (!payload) return
      try {
        const data = JSON.parse(payload)
        if (data.nodeId) updateNodeLog(data)
        if (data.content) {
          // 防止内容重复：如果新内容与已累积内容的尾部大量重复，跳过
          const existing = streamContent.value
          const incoming = data.content
          if (existing && incoming.length > 50) {
            // 检查新内容是否是已有内容尾部的重复（取最长匹配）
            const checkLen = Math.min(incoming.length, existing.length, 2000)
            const existingTail = existing.slice(-checkLen)
            // 在 existingTail 中查找 incoming 的前缀
            const overlapIdx = existingTail.indexOf(incoming.slice(0, 100))
            if (overlapIdx >= 0) {
              // 找到重叠，检查重叠长度是否足够大（超过 100 字符视为重复）
              const overlapLen = checkLen - overlapIdx
              if (overlapLen >= 100 && incoming.length <= overlapLen + 50) {
                // 新内容基本全是重复，跳过
                return
              }
            }
          }
          streamContent.value += incoming
          scrollToBottom()
        }
        if (data.end) {
          isExecuting.value = false
          if (data.error) { hasError.value = true; ElMessage.error('执行失败：' + data.error) }
          else isCompleted.value = true
          if (data.artifacts && Array.isArray(data.artifacts)) artifacts.value = data.artifacts
          try { abortController?.abort() } catch (e) {}
          abortController = null
        }
      } catch (e) { console.error('处理SSE消息出错:', e) }
    }, function () {
      isExecuting.value = false; hasError.value = true
      try { abortController?.abort() } catch (e) {}
      abortController = null; ElMessage.error('执行出错，请稍后重试')
    }, function () { console.log('SSE连接关闭') })
  } catch { isExecuting.value = false; hasError.value = true; ElMessage.error('执行失败，请稍后重试') }
}

// 格式化工具调用结果：智能解析 JSON、截断长文本
const formatToolResult = (text) => {
  if (!text) return { formatted: '', isJson: false }
  // 尝试 JSON 格式化
  try {
    const parsed = typeof text === 'string' ? JSON.parse(text) : text
    const jsonStr = JSON.stringify(parsed, null, 2)
    return { formatted: jsonStr, isJson: true }
  } catch {
    // 非 JSON，检查是否是 URL 或普通文本
    const trimmed = text.trim()
    if (trimmed.startsWith('http')) {
      return { formatted: trimmed, isJson: false }
    }
    return { formatted: text, isJson: false }
  }
}

const updateNodeLog = (data) => {
  const existingIndex = nodeLogs.value.findIndex(log => log.nodeId === data.nodeId)
  if (existingIndex >= 0) {
    const existing = nodeLogs.value[existingIndex]
    const details = [...(existing.details || [])]

    if (data.toolName && data.toolParams && !data.toolResult) {
      // 新的工具调用开始 —— 添加一个进行中的工具调用组
      details.push({
        type: 'tool_call_group',
        toolName: data.toolName,
        toolParams: data.toolParams,
        toolResult: null,
        status: 'running', // running | completed
        showMsg: data.showMsg || '',
        timestamp: Date.now()
      })
    } else if (data.toolName && data.toolResult) {
      // 工具调用结果返回 —— 找到对应的进行中组并更新
      let groupIdx = -1
      for (let i = details.length - 1; i >= 0; i--) {
        if (details[i].type === 'tool_call_group' && details[i].toolName === data.toolName && details[i].status === 'running') {
          groupIdx = i
          break
        }
      }
      if (groupIdx >= 0) {
        const result = formatToolResult(data.toolResult)
        details[groupIdx] = {
          ...details[groupIdx],
          toolResult: data.toolResult,
          toolResultFormatted: result.formatted,
          toolResultIsJson: result.isJson,
          status: 'completed',
          showMsg: data.showMsg || details[groupIdx].showMsg
        }
        // 完成后自动折叠（不展开）
        const groupKey = `${data.nodeId}_${groupIdx}`
        expandedToolGroups.value.delete(groupKey)
      } else {
        // 没找到匹配的调用，降级为独立结果
        const result = formatToolResult(data.toolResult)
        details.push({
          type: 'tool_call_group',
          toolName: data.toolName,
          toolParams: null,
          toolResult: data.toolResult,
          toolResultFormatted: result.formatted,
          toolResultIsJson: result.isJson,
          status: 'completed',
          showMsg: data.showMsg || '',
          timestamp: Date.now()
        })
      }
    } else if (data.showMsg && !data.nodeStatus) {
      details.push({ type: 'progress', showMsg: data.showMsg, timestamp: Date.now() })
    } else if (data.content && data.nodeId) {
      details.push({ type: 'content', content: data.content, timestamp: Date.now() })
    }
    nodeLogs.value[existingIndex] = { ...existing, nodeStatus: data.nodeStatus || existing.nodeStatus, executionLog: data.execution_log || existing.executionLog, details }
  } else {
    const details = []
    if (data.toolName && data.toolParams) {
      details.push({
        type: 'tool_call_group',
        toolName: data.toolName,
        toolParams: data.toolParams,
        toolResult: null,
        status: 'running',
        showMsg: data.showMsg || '',
        timestamp: Date.now()
      })
    } else if (data.showMsg && !data.nodeStatus) {
      details.push({ type: 'progress', showMsg: data.showMsg, timestamp: Date.now() })
    }
    nodeLogs.value.push({ nodeId: data.nodeId, nodeType: data.nodeType || 'unknown', nodeLabel: data.nodeLabel || '', nodeStatus: data.nodeStatus || 'running', executionLog: data.execution_log || null, details })
    expandedNodeIds.value.add(data.nodeId)
  }
  scrollToBottom()
}

const resetForm = () => {
  formData.value.content = ''
  uploadedFile.value = null
  if (fileInputRef.value) fileInputRef.value.value = ''
  // 清除动态表单
  Object.keys(formFields.value).forEach(k => { formFields.value[k] = '' })
  Object.keys(uploadedFiles.value).forEach(k => {
    if (fileInputRefs[k]) fileInputRefs[k].value = ''
  })
  uploadedFiles.value = {}
  isExecuting.value = false; isCompleted.value = false; hasError.value = false
  nodeLogs.value = []; expandedNodeIds.value.clear(); streamContent.value = ''; artifacts.value = []
  try { abortController?.abort() } catch (e) {}
  abortController = null
}

const goBack = () => { router.push('/workflow') }

const downloadFile = async (url, name) => {
  try {
    const response = await fetch(url, { headers: getAuthHeaders() })
    if (!response.ok) { ElMessage.error('下载失败'); return }
    const blob = await response.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a'); a.href = blobUrl; a.download = name
    document.body.appendChild(a); a.click(); document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch { ElMessage.error('下载文件失败') }
}

onMounted(() => { loadWorkflowInfo() })
onBeforeUnmount(() => { try { abortController?.abort() } catch (e) {} })
</script>

<style lang="scss" scoped>
// 蓝色主题
$blue-500: #2B5EFF;
$blue-400: #4facfe;
$blue-gradient: linear-gradient(135deg, $blue-500, $blue-400);

.workflow-exec-container {
  min-height: 100vh;
  background: var(--el-bg-color-page, #f5f7fa);
  display: flex;
  flex-direction: column;
}

// ========== 顶部导航 ==========
.exec-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: var(--el-bg-color, #fff);
  border-bottom: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  flex-shrink: 0;

  .back-button {
    color: var(--el-text-color-secondary);
    border-radius: 8px;
    transition: all 0.2s ease;
    &:hover { color: $blue-500; background: rgba($blue-500, 0.06); }
  }

  .topbar-center {
    flex: 1;
    text-align: center;
  }

  .topbar-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .topbar-right { width: 80px; }
}

// ========== 主体布局 ==========
.exec-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.left-panel {
  width: 380px;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  padding: 24px;
  overflow-y: auto;
  background: var(--el-bg-color, #fff);
  display: flex;
  flex-direction: column;
}

.intro-section {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color-extra-light, #f0f0f0);
}

.intro-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 8px;
  color: var(--el-text-color-primary);
  letter-spacing: -0.3px;
}

.intro-subtitle {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin: 0;
  line-height: 1.6;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 12px;
  color: var(--el-text-color-primary);

  .el-icon { color: $blue-500; }
}

.input-textarea {
  :deep(.el-textarea__inner) {
    border-radius: 10px;
    resize: none;
    font-size: 14px;
    line-height: 1.6;
    padding: 12px 16px;
    &:focus { box-shadow: 0 0 0 2px rgba($blue-500, 0.15); }
  }
}

.title-hint {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-text-color-secondary);
}

.title-required {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-color-danger);
}

.field-description {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin: 6px 0 0;
  line-height: 1.4;
}

// ========== 文件上传 ==========
.upload-area {
  border: 2px dashed var(--el-border-color-light, #dcdfe6);
  border-radius: 12px;
  transition: all 0.25s ease;
  overflow: hidden;

  &:hover, &.is-dragover {
    border-color: $blue-500;
    background: rgba($blue-500, 0.02);
  }

  &.has-file {
    border-style: solid;
    border-color: var(--el-color-success-light-5, #e1f3d8);
  }
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  cursor: pointer;

  .upload-icon {
    font-size: 32px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }

  .upload-text {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    margin: 0 0 4px;
  }

  .upload-link {
    color: $blue-500;
    cursor: pointer;

    &:hover { text-decoration: underline; }
  }

  .upload-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin: 0;
  }
}

.uploaded-file {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--el-fill-color-lighter, #fafafa);

  .file-info {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .file-icon {
    font-size: 24px;
    color: $blue-500;
    flex-shrink: 0;
  }

  .file-detail {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .file-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-size {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .remove-file {
    color: var(--el-text-color-secondary);
    flex-shrink: 0;

    &:hover { color: var(--el-color-danger); }
  }
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: auto;
  padding-top: 20px;

  .reset-button {
    flex: 1;
    border-radius: 10px;
  }

  .execute-button {
    flex: 2;
    border-radius: 10px;
    background: $blue-gradient !important;
    border: none !important;
    box-shadow: 0 4px 16px rgba($blue-500, 0.3);
    font-weight: 500;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 6px 20px rgba($blue-500, 0.4);
      transform: translateY(-1px);
    }

    &:active { transform: translateY(0); }
  }
}

// ========== 右侧面板 ==========
.right-panel {
  flex: 1;
  overflow: hidden;
  background: var(--el-bg-color-page, #f5f7fa);
}

.panel-content {
  height: 100%;
  overflow-y: auto;
  padding: 24px;
}

// ========== 空状态 ==========
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--el-text-color-secondary);

  .empty-icon-wrapper {
    width: 80px;
    height: 80px;
    border-radius: 20px;
    background: rgba($blue-500, 0.06);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 20px;
  }

  .empty-icon { font-size: 36px; color: rgba($blue-500, 0.4); }

  .empty-title {
    font-size: 18px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin: 0 0 8px;
  }

  .empty-desc {
    font-size: 14px;
    color: var(--el-text-color-placeholder);
    margin: 0;
  }
}

// ========== 执行进度 ==========
.execution-progress { height: 100%; display: flex; flex-direction: column; }

.progress-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-radius: 12px;
  margin-bottom: 20px;
  font-weight: 500;
  font-size: 14px;

  .header-left { display: flex; align-items: center; gap: 10px; }
  .header-right { display: flex; align-items: center; }

  .node-count {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-lighter, #f5f5f5);
    padding: 4px 10px;
    border-radius: 6px;
  }

  .status-icon { font-size: 18px; }
  .status-spinner { animation: spin 1s linear infinite; }

  &.executing {
    background: rgba($blue-500, 0.06);
    color: $blue-500;
    border: 1px solid rgba($blue-500, 0.12);
  }

  &.completed {
    background: var(--el-color-success-light-9);
    color: var(--el-color-success);
    border: 1px solid var(--el-color-success-light-7);
  }

  &.error {
    background: var(--el-color-danger-light-9);
    color: var(--el-color-danger);
    border: 1px solid var(--el-color-danger-light-7);
  }
}

@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

@keyframes pulse-border {
  0%, 100% { border-color: rgba($blue-500, 0.15); box-shadow: 0 0 0 1px rgba($blue-500, 0.05); }
  50% { border-color: rgba($blue-500, 0.35); box-shadow: 0 0 0 3px rgba($blue-500, 0.08); }
}

@keyframes running-dot-pulse {
  0%, 100% { opacity: 0.4; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.2); }
}

// ========== 节点日志 ==========
.node-logs { margin-bottom: 20px; }

.node-log-item {
  background: var(--el-bg-color, #fff);
  border: 1px solid var(--el-border-color-light, #dcdfe6);
  border-radius: 12px;
  margin-bottom: 10px;
  overflow: hidden;
  transition: all 0.25s ease;

  &:hover { box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04); }

  &.is-running {
    animation: pulse-border 2s ease-in-out infinite;
  }

  &.expanded {
    border-color: rgba($blue-500, 0.25);
    box-shadow: 0 4px 16px rgba($blue-500, 0.06);
  }
}

.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  cursor: pointer;
  transition: background 0.2s ease;
  &:hover { background: var(--el-fill-color-lighter, #fafafa); }
}

.node-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.node-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  &.start { background: rgba(#43e97b, 0.1); color: #43e97b; }
  &.end { background: rgba(#909399, 0.1); color: #909399; }
  &.llm { background: rgba($blue-500, 0.1); color: $blue-500; }
  &.tool { background: rgba(#f6d365, 0.1); color: #e6a23c; }
  &.agent { background: rgba(#e91e63, 0.1); color: #e91e63; }
  &.condition { background: rgba(#f56c6c, 0.1); color: #f56c6c; }
  &.code { background: rgba(#607d8b, 0.1); color: #607d8b; }
  &.http { background: rgba(#00bcd4, 0.1); color: #00bcd4; }
}

.node-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.node-duration {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.node-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-status {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;

  &.completed { color: var(--el-color-success); }
  &.running { color: $blue-500; }
  &.failed { color: var(--el-color-danger); }
}

.expand-arrow {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  transition: transform 0.25s ease;
  &.is-expanded { transform: rotate(180deg); }
}

.node-detail {
  border-top: 1px solid var(--el-border-color-light, #dcdfe6);
  padding: 16px;
  background: var(--el-fill-color-lighter, #fafafa);
}

// ========== 执行细节时间线 ==========
.node-details-timeline {
  margin-bottom: 16px;
  padding-left: 16px;
  border-left: 2px solid var(--el-border-color, #dcdfe6);

  .detail-item {
    position: relative;
    margin-bottom: 12px;
    padding-left: 16px;

    &:last-child { margin-bottom: 0; }

    &::before {
      content: '';
      position: absolute;
      left: -21px;
      top: 8px;
      width: 10px;
      height: 10px;
      border-radius: 50%;
      background: var(--el-border-color, #dcdfe6);
    }

    &.tool_call_group.running::before { background: #e6a23c; animation: running-dot-pulse 1.2s ease-in-out infinite; }
    &.tool_call_group.completed::before { background: var(--el-color-success); }
    &.progress::before { background: $blue-500; }
    &.content::before { background: $blue-400; }
  }

  .detail-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }

  .detail-dot { display: none; } // 使用 ::before 代替

  .detail-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .detail-msg {
    font-size: 12px;
    color: var(--el-text-color-primary);
    line-height: 1.5;
  }

  .detail-code {
    font-size: 12px;
    line-height: 1.5;
    background: var(--el-bg-color, #fff);
    border: 1px solid var(--el-border-color-light, #dcdfe6);
    padding: 8px 12px;
    border-radius: 8px;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 240px;
    overflow-y: auto;
    color: var(--el-text-color-primary);

    &.result-code { border-left: 3px solid var(--el-color-success); }
    &.json-code { font-family: 'SF Mono', 'Menlo', 'Monaco', monospace; }
    &.content-code { border-left: 3px solid $blue-400; }
  }
}

// ========== 工具调用组 ==========
.detail-tool-group {
  .tool-group-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 0;

    &.clickable {
      cursor: pointer;
      border-radius: 6px;
      padding: 6px 8px;
      margin: -6px -8px;
      transition: background 0.2s ease;
      &:hover { background: rgba($blue-500, 0.04); }
    }
  }

  .tool-group-left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .tool-group-summary {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 260px;
  }

  .tool-spinner {
    color: #e6a23c;
    font-size: 16px;
    animation: spin 1s linear infinite;
    flex-shrink: 0;
  }

  .tool-done {
    color: var(--el-color-success);
    font-size: 16px;
    flex-shrink: 0;
  }

  .tool-expand-arrow {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    transition: transform 0.25s ease;
    flex-shrink: 0;
    &.is-expanded { transform: rotate(180deg); }
  }

  .tool-group-body {
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px dashed var(--el-border-color-light, #dcdfe6);
  }

  .tool-section {
    margin-bottom: 8px;
    &:last-child { margin-bottom: 0; }
  }

  .tool-section-label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    color: var(--el-text-color-placeholder);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
  }

  .tool-running-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    padding: 4px 0;
  }

  .tool-running-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #e6a23c;
    animation: running-dot-pulse 1.2s ease-in-out infinite;
  }
}

// ========== 执行日志 ==========
.exec-log {
  .log-section {
    margin-bottom: 12px;
    &.error .log-content { color: var(--el-color-danger); border-left: 3px solid var(--el-color-danger); }
  }

  .log-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    display: block;
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .log-content {
    font-size: 13px;
    line-height: 1.5;
    background: var(--el-bg-color, #fff);
    border: 1px solid var(--el-border-color-light, #dcdfe6);
    padding: 10px 14px;
    border-radius: 8px;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 200px;
    overflow-y: auto;
    color: var(--el-text-color-primary);
  }
}

// ========== 流式输出 ==========
.stream-output {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);

  .output-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;
    font-weight: 600;
    margin: 0 0 16px;
    color: var(--el-text-color-primary);

    .el-icon { color: $blue-500; }
  }

  .markdown-body {
    font-size: 14px;
    line-height: 1.8;
    color: var(--el-text-color-regular);
    background: var(--el-bg-color, #fff);
    padding: 20px;
    border-radius: 12px;
    border: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  }
}

// ========== 文件产物 ==========
.artifacts-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);

  .output-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;
    font-weight: 600;
    margin: 0 0 16px;
    color: var(--el-text-color-primary);

    .el-icon { color: $blue-500; }
  }

  .artifacts-list { display: flex; flex-direction: column; gap: 10px; }

  .artifact-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    background: var(--el-bg-color, #fff);
    border-radius: 12px;
    border: 1px solid var(--el-border-color-extra-light, #f0f0f0);
    transition: all 0.2s;

    &:hover {
      border-color: rgba($blue-500, 0.2);
      box-shadow: 0 2px 12px rgba($blue-500, 0.06);
    }
  }

  .artifact-info { display: flex; align-items: center; gap: 12px; min-width: 0; }

  .artifact-icon-wrapper {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    &.pdf { background: rgba(#f56c6c, 0.1); color: #f56c6c; }
    &.markdown { background: rgba($blue-500, 0.1); color: $blue-500; }
    &.image { background: rgba(#67c23a, 0.1); color: #67c23a; }
  }

  .artifact-detail { display: flex; flex-direction: column; min-width: 0; }

  .artifact-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .artifact-type { font-size: 12px; color: var(--el-text-color-secondary); }

  .download-btn {
    flex-shrink: 0;
    border-radius: 8px;
    background: $blue-gradient !important;
    border: none !important;
  }
}

// ========== 加载动画 ==========
.center-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  color: var(--el-text-color-secondary);
  gap: 16px;

  .loading-spinner {
    width: 36px;
    height: 36px;
    border: 3px solid var(--el-border-color-lighter, #e4e7ed);
    border-top-color: $blue-500;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
}

// ========== 响应式 ==========
@media (max-width: 768px) {
  .exec-body { flex-direction: column; }
  .left-panel { width: 100%; border-right: none; border-bottom: 1px solid var(--el-border-color-extra-light); }
}
</style>
