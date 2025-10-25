<template>
  <div class="travel-container">
    <div class="travel-content">
      <el-card class="travel-card">
        <div class="two-column">
          <!-- 左侧：输入信息 -->
          <div class="left-panel">
            <div class="panel-section page-intro">
              <h2 class="intro-title">AI 旅游攻略生成</h2>
              <p class="intro-subtitle">填写目的地与偏好，右侧实时展示生成进度与结果</p>
            </div>

            <div class="panel-section">
              <h2 class="section-title">行程信息</h2>
              <el-form :model="formData" label-position="top" class="input-form">
                <el-form-item label="目的地">
                  <el-input v-model="formData.destination" placeholder="如：成都 / 东京 / 巴黎"
                            :disabled="isFormDisabled"/>
                </el-form-item>
                <el-form-item label="出行日期">
                  <div class="date-row">
                    <el-date-picker v-model="formData.startDate" type="date" placeholder="开始日期"
                                    :disabled="isFormDisabled"/>
                    <span class="date-sep">至</span>
                    <el-date-picker v-model="formData.endDate" type="date" placeholder="结束日期"
                                    :disabled="isFormDisabled"/>
                  </div>
                </el-form-item>
                <el-form-item label="预算（可选）">
                  <el-input v-model="formData.budget" placeholder="如：5000 元 / 人" :disabled="isFormDisabled"/>
                </el-form-item>
                <el-form-item label="偏好（可选）">
                  <el-input v-model="formData.preferences" type="textarea" :rows="6"
                            placeholder="如：美食、历史文化、亲子、徒步、摄影等" :disabled="isFormDisabled"/>
                </el-form-item>

                <div class="action-buttons">
                  <el-button @click="resetProcess" class="reset-button" :icon="RefreshRight" :disabled="isFormDisabled">
                    重置
                  </el-button>
                  <el-button type="primary" :disabled="isFormDisabled || !formData.destination" @click="startGenerate"
                             class="generate-button" :icon="Promotion">开始生成
                  </el-button>
                </div>
              </el-form>
            </div>
          </div>

          <!-- 右侧：过程与结果展示 -->
          <div class="right-panel">
            <div class="panel-content">
              <div v-if="!formData.destination && !isOptimizing && !isCompleted && !hasError" class="empty-state">
                <el-icon class="empty-icon">
                  <Document/>
                </el-icon>
                <p>请输入目的地，开始生成旅游攻略</p>
              </div>

              <div class="generation-progress" v-else>
                <div class="progress-header"
                     :class="{'optimizing': isOptimizing, 'completed': isCompleted, 'error': hasError}">
                  <el-icon class="status-icon loading-icon" v-if="isOptimizing">
                    <loading/>
                  </el-icon>
                  <el-icon class="status-icon success-icon" v-else-if="isCompleted">
                    <circle-check/>
                  </el-icon>
                  <el-icon class="status-icon error-icon" v-else-if="hasError">
                    <circle-close/>
                  </el-icon>
                  <el-icon class="status-icon" v-else>
                    <info-filled/>
                  </el-icon>
                  <span class="status-text" v-if="isOptimizing">AI正在生成旅游攻略...</span>
                  <span class="status-text" v-else-if="isCompleted">旅游攻略已生成！</span>
                  <span class="status-text" v-else-if="hasError">生成失败</span>
                  <span class="status-text" v-else>准备生成攻略</span>
                </div>

                <div class="process-details-container">
                  <div class="process-details" ref="processDetailsContainer">
                    <div class="agent-reasoning" v-if="agentReasoning">
                      <div class="reasoning-content">
                        <pre>{{ agentReasoning }}</pre>
                      </div>
                    </div>

                    <div class="tool-calls" v-if="toolCallsList.length > 0">
                      <div v-for="(tool, index) in toolCallsList" :key="tool.id" class="tool-call"
                           :class="{'expanded': expandedToolIds.has(tool.id) || (!tool.result && index === toolCallsList.length - 1)}">
                        <div class="tool-header" @click="toggleToolExpand(tool.id)">
                          <span class="tool-name">{{ tool.name }}</span>
                          <span class="tool-status" :class="[ tool.result ? 'completed' : 'loading' ]">
                            <template v-if="tool.result">已完成</template>
                            <template v-else>
                              <el-icon class="status-spinner"><loading/></el-icon>
                            </template>
                          </span>
                        </div>
                        <div class="tool-request-info" v-if="!expandedToolIds.has(tool.id) && (tool.result || index !== toolCallsList.length - 1)">
                          <div v-if="tool.showMsg" class="request-preview">
                            {{ getRequestPreview(tool.showMsg) }}
                          </div>
                        </div>
                        <div class="tool-message" v-show="expandedToolIds.has(tool.id) || (!tool.result && index === toolCallsList.length - 1)">
                          <div v-if="containsHtml(tool.showMsg)" v-html="tool.showMsg"></div>
                          <pre v-else>{{ tool.showMsg }}</pre>
                        </div>
                      </div>
                    </div>

                    <div v-if="isOptimizing && !hasActiveToolCall" class="center-loading">
                      <el-icon class="loading-icon"><loading/></el-icon>
                    </div>


                    <div class="optimization-result" v-if="isCompleted">
                      <div class="result-header">
                        <el-icon class="success-icon">
                          <circle-check/>
                        </el-icon>
                        <h2>旅游攻略生成完成！</h2>
                      </div>
                      <div class="result-summary" v-if="summary">
                        <h3 class="summary-title">生成总结</h3>
                        <div class="summary-content" v-html="formattedSummary"></div>
                      </div>
                      <div class="download-section">
                        <el-button type="primary" @click="downloadGuide" :disabled="!downloadUrl" class="download-button">
                          <el-icon>
                            <download/>
                          </el-icon>
                          下载旅游攻略 PDF
                        </el-button>
                      </div>
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
import {computed, onBeforeUnmount, ref} from 'vue'
import {ElMessage} from 'element-plus'
import {
  Check,
  CircleCheck,
  CircleClose,
  Close,
  Document,
  Download,
  InfoFilled,
  Loading,
  Promotion,
  RefreshRight
} from '@element-plus/icons-vue'
import {marked} from 'marked'
import fetchSSEWithAuth from '@/api/sseRequest.js'
import apiClient from "@/api/config.js";
import {defaultApi} from "@/api/index.js";
import auth from "@/api/auth.js";
import {useUserStore} from "@/stores/user.js";

// 表单数据
const formData = ref({
  destination: '',
  startDate: '',
  endDate: '',
  budget: '',
  preferences: ''
})

// 进度步骤
const steps = ref([
  {title: '收集信息', description: '整理目的地和偏好', active: false, completed: false, failed: false},
  {title: '规划路线', description: '规划每日行程与交通', active: false, completed: false, failed: false},
  {title: '生成文档', description: '生成并导出攻略 PDF', active: false, completed: false, failed: false}
])

// AI推理与工具调用
const agentReasoning = ref('')
const toolCallsMap = ref(new Map())
const toolCallsList = computed(() => Array.from(toolCallsMap.value.values()))
const processDetailsContainer = ref(null)
const expandedToolIds = ref(new Set())

// 检查是否有正在进行中的工具调用
const hasActiveToolCall = computed(() => toolCallsList.value.some(tool => !tool.result))

// 总结与下载
const summary = ref('')
const formattedSummary = computed(() => summary.value ? marked(summary.value) : '')
const downloadUrl = ref('')

// 检测工具消息是否包含 HTML（当前只检测 img 标签）
const containsHtml = (msg) => typeof msg === 'string' && /<\s*img\b/i.test(msg)

// 获取请求预览信息
const getRequestPreview = (msg) => {
  if (!msg) return ''
  // 提取【请求】后的内容，最多显示50个字符
  const requestPart = msg.split('【请求】')[1]?.split('【响应】')[0] || msg
  const preview = requestPart.trim().substring(0, 50)
  return preview + (preview.length >= 50 ? '...' : '')
}

// 切换工具调用展开/折叠状态
const toggleToolExpand = (toolId) => {
  if (expandedToolIds.value.has(toolId)) {
    expandedToolIds.value.delete(toolId)
  } else {
    expandedToolIds.value.add(toolId)
  }
}

// 滚动到最新位置
const scrollToBottom = () => {
  if (processDetailsContainer.value) {
    setTimeout(() => {
      processDetailsContainer.value.scrollTop = processDetailsContainer.value.scrollHeight
    }, 50)
  }
}

// 监听内容变化自动滚动到底部
const autoScrollToBottom = () => {
  scrollToBottom()
}

let abortController = null

// 状态派生
const isOptimizing = computed(() => steps.value.some(s => s.active))
const isCompleted = computed(() => steps.value.length > 0 && steps.value.every(s => s.completed))
const hasError = computed(() => steps.value.some(s => s.failed))
const isFormDisabled = computed(() => isOptimizing.value && !isCompleted.value && !hasError.value)

const resetProcess = () => {
  formData.value = {destination: '', startDate: '', endDate: '', budget: '', preferences: ''}
  steps.value.forEach(s => {
    s.active = false;
    s.completed = false;
    s.failed = false
  })
  agentReasoning.value = ''
  toolCallsMap.value.clear()
  summary.value = ''
  downloadUrl.value = ''
}

const startGenerate = async () => {
  if (!formData.value.destination) {
    ElMessage.error('请先填写目的地')
    return
  }

  // 重置并激活第一步
  steps.value.forEach(s => {
    s.active = false;
    s.completed = false;
    s.failed = false
  })
  agentReasoning.value = ''
  toolCallsMap.value.clear()
  summary.value = ''
  downloadUrl.value = ''
  steps.value[0].active = true

  // 构建请求内容
  const parts = []
  parts.push(`目的地：\n ${formData.value.destination}`)
  if (formData.value.startDate || formData.value.endDate) {
    parts.push(`日期：\n ${formData.value.startDate || ''} 至 ${formData.value.endDate || ''}`)
  }
  if (formData.value.budget) parts.push(`预算：\n ${formData.value.budget}`)
  if (formData.value.preferences) parts.push(`偏好：\n ${formData.value.preferences}`)
  const content = parts.join('\n') || '请生成我的旅游攻略'

  // 创建请求数据（SSE）
  const formDataObj = new FormData()
  formDataObj.append('agentType', 'travel')
  formDataObj.append('content', content)

  try {
    const url = '/api/agent/exec/stream'
    abortController = await fetchSSEWithAuth(url, formDataObj, function (msg) {
      if (!msg.startsWith('data:')) {
        throw new Error('非法的SSE消息格式')
      }
      const payload = msg.slice(5).trim()
      if (!payload) return
      try {
        const data = JSON.parse(payload)

        if (data.reasoningContent) {
          agentReasoning.value += data.reasoningContent
          scrollToBottom()
        }

        if (data.toolName && data.toolCallId) {
          if (!toolCallsMap.value.has(data.toolCallId)) {
            toolCallsMap.value.set(data.toolCallId, {
              id: data.toolCallId,
              name: data.toolName,
              showMsg: data.showMsg ? `【请求】\n${data.showMsg}` : '',
              result: false
            })
            scrollToBottom()
          } else if (data.toolResult) {
            const tool = toolCallsMap.value.get(data.toolCallId)
            if (tool) {
              const responseText = data.showMsg || data.toolResult || ''
              var isHtml = containsHtml(responseText);
              var breakLine = isHtml ? '<br>' : '\n'
              const combinedMsg = tool.showMsg ? `${tool.showMsg}${breakLine}【响应】${breakLine}${responseText}` : `【响应】${breakLine}${responseText}`
              toolCallsMap.value.set(data.toolCallId, {...tool, showMsg: combinedMsg, result: true})
              scrollToBottom()
            }
          }

          // 当调用 Markdown 转 PDF 工具时，标记前两步完成，第三步进行中
          if (data.toolName === 'markdown_to_pdf_file_tool') {
            steps.value[2].active = true
            steps.value[0].active = false;
            steps.value[0].completed = true
            steps.value[1].active = false;
            steps.value[1].completed = true
          }
        }

        if (data.content && !data.end) {
          summary.value = data.content
          // 简单的进度切换：当收到较长内容时，认为路线规划完成，进入生成文档
          if (summary.value.length > 200 && !steps.value[1].completed) {
            steps.value[0].active = false;
            steps.value[0].completed = true
            steps.value[1].active = true
          }
          scrollToBottom()
        }

        if (data.end) {
          try {
            abortController?.abort()
          } catch (e) {
          }
          abortController = null

          if (data.error) {
            const activeIdx = steps.value.findIndex(s => s.active)
            if (activeIdx !== -1) {
              steps.value[activeIdx].active = false;
              steps.value[activeIdx].failed = true
            }
            ElMessage.error('旅游攻略生成失败：' + (data.error || '未知错误'))
          } else {
            steps.value.forEach(s => {
              s.active = false;
              s.completed = true
            })
            if (data.content) downloadUrl.value = data.content
          }
        }
      } catch (e) {
        console.error('处理SSE消息时出错:', e)
      }
    }, function (error) {
      if (isCompleted.value) {
        return
      }
      try {
        abortController?.abort()
      } catch (e) {
        console.error('取消SSE请求时出错:', e)
      }
      abortController = null

      const activeIdx = steps.value.findIndex(s => s.active)
      if (activeIdx !== -1) {
        steps.value[activeIdx].active = false;
        steps.value[activeIdx].failed = true
      } else if (steps.value.length > 0) {
        steps.value[0].failed = true
      }

      const errText = typeof error === 'string' ? error : (error?.message || JSON.stringify(error))
      try {
        for (const tool of toolCallsMap.value.values()) {
          if (!tool.result) {
            toolCallsMap.value.set(tool.id, {
              ...tool,
              showMsg: tool.showMsg ? `${tool.showMsg}\n\n【错误】\n${errText}` : `【错误】\n${errText}`
            })
          }
        }
      } catch (e) {
      }

      ElMessage.error('系统繁忙，请稍后重试')
      console.error('SSE请求出错:', error)
    }, function () {
      console.log('SSE连接关闭')
    })
  } catch (error) {
    console.error('生成旅游攻略时出错:', error)
    ElMessage.error('生成旅游攻略失败，请稍后重试')
  }
}

// TODO 后续抽取公共的下载工具函数，实现 token 刷新逻辑等
const downloadGuide = async () => {
  if (!downloadUrl.value) {
    ElMessage.error('下载链接不可用')
    return
  }

  try {
    let url = downloadUrl.value

    const token = localStorage.getItem('token')
    if (!token) {
      ElMessage.error('请先登录后再下载')
      useUserStore().logout()
      return
    }
    const sep = url.includes('?') ? '&' : '?'
    url = `${url}${sep}Authorization=${encodeURIComponent(`Bearer ${token}`)}`

    window.open(url, '_blank')
  } catch (e) {
    console.error('触发下载时出错:', e)
    ElMessage.error('下载失败，请稍后重试')
  }
}

onBeforeUnmount(() => {
  abortController?.abort()
  abortController = null
})
</script>

<style scoped lang="scss">
.travel-container {
  width: 100%;
  max-width: none;
  margin: 0;
  padding-top: 5px;
  height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
  background-color: var(--el-bg-color-page, #f5f7fa);
}

.travel-card {
  //border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
  width: 100%;
  flex: 1;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  }

  :deep(.el-card__body) {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 0;
  }
}

.travel-content {
  flex: 1;
  display: flex;
  width: 100%;
  height: 100%;
}

/* 两栏布局 */
.two-column {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 0;
  height: 100%;
}

.left-panel {
  padding: 30px;
  overflow: auto;
  border-right: 1px solid var(--el-border-color-light);
  display: flex;
  flex-direction: column;
  gap: 30px;
  background-color: var(--el-bg-color);
}

.right-panel {
  //padding: 5px;
  height: 100%;
  min-height: 0; /* 允许在网格/弹性布局中收缩以显示滚动 */
  overflow-y: auto; /* 让 right-panel 自身滚动 */
  background-color: var(--el-bg-color-page, #f5f7fa);
  display: flex;
  flex-direction: column;
}

.page-intro {
  .intro-title {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 8px 0;
    background: linear-gradient(45deg, var(--el-color-primary), var(--el-color-primary-light-3));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    letter-spacing: 0.5px;
  }

  .intro-subtitle {
    font-size: 14px;
    color: var(--el-text-color-secondary);
  }
}

.panel-section {
  .section-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 20px;
    color: var(--el-text-color-primary);
    position: relative;
    padding-left: 15px;

    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 4px;
      height: 18px;
      background: var(--el-color-primary);
      border-radius: 2px;
    }
  }
}

.panel-content {
  display: flex;
  flex-direction: column;
  flex: 1; /* 占满剩余空间 */
  overflow: visible; /* 允许内容超出，由 right-panel 控制滚动 */
  min-height: 0;
  //padding: 10px 0;

  .generation-progress {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
}

.process-details-container {
  flex: 1;
  overflow: auto;
  position: relative;
}

.process-details {
  height: 100%;
  overflow-y: auto;
  padding-right: 5px;
}

.date-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.date-sep {
  color: var(--el-text-color-secondary);
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 8px;
  //margin-bottom: 12px;
  background-color: white;
  padding: 15px;
  box-sizing: border-box;
  box-shadow: 0 6px 12px -6px rgba(0, 0, 0, 0.15); /* 下边界阴影 */
  z-index: 1; /* 提升层级，避免被后续内容遮挡 */

  &.optimizing .status-text {
    color: var(--el-color-primary);
  }
  &.completed .status-text {
    color: var(--el-color-success);
  }
  &.error .status-text {
    color: var(--el-color-danger);
  }
}

.status-icon {
  font-size: 18px;
}

.loading-icon {
  color: var(--el-color-primary);
}
.success-icon {
  color: var(--el-color-success);
}
.error-icon {
  color: var(--el-color-danger);
}

.progress-steps {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

.progress-step {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.process-details-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: calc(100% - 40px);
  //margin-top: 20px;
}

.process-details {
  //flex: 1;
  overflow-y: auto;
  padding: 10px;
  //border-radius: 8px;
  background-color: var(--el-bg-color);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);

  border-top: 10px solid white;
  border-bottom: 10px solid white;
}

.agent-reasoning {
  margin-bottom: 20px;
}

.reasoning-content {
  background-color: var(--el-fill-color-lighter);
  padding: 12px;
  border-radius: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.reasoning-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--el-font-family);
  font-size: 14px;
  line-height: 1.6;
}

.tool-calls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tool-call {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.tool-call.expanded {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background-color: var(--el-fill-color-light);
  cursor: pointer;
}

.tool-name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.tool-status {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
}

.tool-status.completed {
  color: var(--el-color-success);
}

.tool-status.loading {
  color: var(--el-color-primary);
}

.status-spinner {
  animation: spin 1s linear infinite;
}

.tool-request-info {
  padding: 8px 15px;
  background-color: var(--el-fill-color-lighter);
  border-top: 1px solid var(--el-border-color-lighter);
}

.request-preview {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.5;
}

.tool-message {
  padding: 12px 15px;
  background-color: var(--el-fill-color-lighter);
  max-height: 300px;
  overflow-y: auto;
}

.tool-message pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--el-font-family);
  font-size: 14px;
  line-height: 1.6;
}

.center-loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 30px 0;
}

.center-loading .loading-icon {
  font-size: 24px;
  color: var(--el-color-primary);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.step-icon.active {
  background: var(--el-color-primary-light-5);
}

.step-icon.completed {
  background: var(--el-color-success-light-5);
}

.step-icon.failed {
  background: var(--el-color-danger-light-5);
}

.details-collapse {
  margin-top: 12px;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tool-message pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px 0;
}

.summary-content {
  background: var(--el-fill-color);
  padding: 12px;
  border-radius: 8px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
  margin-top: 20px;
}

.empty-icon {
  font-size: 34px;
}

/* 右侧详细过程与工具调用样式，保持与简历页一致 */
.details-collapse {
  margin-top: 20px;
  border-radius: 12px;
  overflow: hidden;

  :deep(.el-collapse-item__header) {
    font-weight: 500;
    font-size: 16px;
    padding: 15px;
    background-color: var(--el-bg-color);
  }

  :deep(.el-collapse-item__content) {
    padding: 20px;
    background-color: var(--el-bg-color);
  }
}

.detail-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 15px;
  color: var(--el-text-color-primary);
}

.agent-reasoning {
  margin-bottom: 25px;

  .reasoning-content {
    background-color: var(--el-fill-color-light);
    border-radius: 8px;
    padding: 15px;
    max-height: 200px;
    overflow-y: auto;

    pre {
      white-space: pre-wrap;
      word-break: break-word;
      font-family: monospace;
      margin: 0;
      font-size: 13px;
    }
  }
}

.tool-calls {
  .tool-call {
    background-color: var(--el-fill-color-light);
    border-radius: 8px;
    padding: 15px;
    margin-bottom: 15px;

    .tool-header {
      margin-bottom: 10px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .tool-name {
        font-weight: bold;
        color: var(--el-color-primary);
      }

      .tool-status {
        font-size: 12px;
        padding: 4px 10px;
        border-radius: 20px;
        background-color: var(--el-color-warning-light-9);
        color: var(--el-color-warning-dark-2);

        &.completed {
          background-color: var(--el-color-success-light-9);
          color: var(--el-color-success-dark-2);
        }

        &.loading {
          display: inline-flex;
          align-items: center;
          gap: 6px;
        }

        .status-spinner {
          font-size: 14px;
          color: var(--el-color-warning);
          animation: rotate 1s linear infinite;
        }

        .dot {
          display: inline-block;
          width: 2px;
          color: var(--el-color-warning-dark-2);
          animation: blink 1s infinite;
        }
        .dot-2 { animation-delay: 0.2s; }
        .dot-3 { animation-delay: 0.4s; }
      }
    }

    .tool-message {
      background-color: var(--el-bg-color);
      border-radius: 8px;
      padding: 12px;
      margin-bottom: 10px;
      border-left: 3px solid var(--el-color-primary-light-5);

      pre {
        white-space: pre-wrap;
        word-break: break-word;
        font-family: monospace;
        margin: 0;
        font-size: 13px;
      }
    }
  }
}

/* 结果区域样式与简历页一致 */
.optimization-result {
  margin-top: 30px;
  padding: 30px;
  background-color: var(--el-bg-color);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);

  .result-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 30px;

    .success-icon {
      font-size: 64px;
      color: var(--el-color-success);
      margin-bottom: 20px;
    }

    h2 {
      font-size: 24px;
      margin: 0;
      font-weight: 600;
    }
  }

  .result-summary {
    margin-bottom: 30px;

    .summary-title {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 15px;
      color: var(--el-text-color-primary);
      position: relative;
      padding-left: 15px;

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 4px;
        height: 18px;
        background: var(--el-color-success);
        border-radius: 2px;
      }
    }

    .summary-content {
      background-color: var(--el-fill-color-light);
      border-radius: 12px;
      padding: 25px;
      box-shadow: inset 0 0 0 1px var(--el-border-color-light);

      :deep(h2), :deep(h3) { margin-top: 0; }
      :deep(ul) { padding-left: 20px; }
      :deep(p) { margin-bottom: 12px; line-height: 1.6; }
      :deep(a) {
        color: var(--el-color-primary);
        text-decoration: none;
        &:hover { text-decoration: underline; }
      }
    }
  }

  .download-section {
    text-align: center;
    margin-top: 30px;

    .download-button {
      min-width: 200px;
      height: 48px;
      border-radius: 24px;
      font-size: 16px;
      font-weight: 500;
      background: linear-gradient(45deg, var(--el-color-success), var(--el-color-success-light-3));
      border: none;
      transition: all 0.3s;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 5px 15px rgba(var(--el-color-success-rgb), 0.3);
      }

      .el-icon { margin-right: 8px; font-size: 18px; }
    }
  }
}

/* 动画与效果 */
@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes blink {
  0% { opacity: 0; }
  50% { opacity: 1; }
  100% { opacity: 0; }
}

/* 步骤信息文字与简历页一致 */
.progress-step {
  .step-info {
    .step-title { font-weight: bold; margin-bottom: 5px; font-size: 16px; }
    .step-description { color: var(--el-text-color-secondary); font-size: 14px; }
  }
}
</style>
