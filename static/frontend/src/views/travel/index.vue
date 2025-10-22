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

                <div class="progress-details">
                  <div class="progress-steps">
                    <div v-for="(step, idx) in steps" :key="idx" class="progress-step">
                      <div class="step-icon"
                           :class="{ 'active': step.active, 'completed': step.completed, 'failed': step.failed }">
                        <el-icon v-if="step.completed">
                          <check/>
                        </el-icon>
                        <el-icon v-else-if="step.failed">
                          <close/>
                        </el-icon>
                        <el-icon v-else-if="step.active">
                          <loading/>
                        </el-icon>
                        <span v-else>{{ idx + 1 }}</span>
                      </div>
                      <div class="step-info">
                        <div class="step-title">{{ step.title }}</div>
                        <div class="step-description">{{ step.description }}</div>
                      </div>
                    </div>
                  </div>
                </div>

                <el-collapse v-model="detailsActiveNames" v-if="agentReasoning || toolCallsList.length > 0"
                             class="details-collapse">
                  <el-collapse-item title="查看详细过程" name="1">
                    <div class="agent-reasoning" v-if="agentReasoning">
                      <h3 class="detail-title">AI分析过程</h3>
                      <div class="reasoning-content">
                        <pre>{{ agentReasoning }}</pre>
                      </div>
                    </div>
                    <div class="tool-calls" v-if="toolCallsList.length > 0">
                      <h3 class="detail-title">工具调用</h3>
                      <div v-for="tool in toolCallsList" :key="tool.id" class="tool-call">
                        <div class="tool-header">
                          <span class="tool-name">{{ tool.name }}</span>
                          <span class="tool-status" :class="[ tool.result ? 'completed' : 'loading' ]">
                            <template v-if="tool.result">已完成</template>
                            <template v-else>
                              <el-icon class="status-spinner"><loading/></el-icon>
                              请求中<span class="dot dot-1">.</span><span class="dot dot-2">.</span><span
                              class="dot dot-3">.</span>
                            </template>
                          </span>
                        </div>
                        <div class="tool-message">
                          <div v-if="containsHtml(tool.showMsg)" v-html="tool.showMsg"></div>
                          <pre v-else>{{ tool.showMsg }}</pre>
                        </div>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
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
const detailsActiveNames = ref(['1'])

// 总结与下载
const summary = ref('')
const formattedSummary = computed(() => summary.value ? marked(summary.value) : '')
const downloadUrl = ref('')

// 检测工具消息是否包含 HTML（当前只检测 img 标签）
const containsHtml = (msg) => typeof msg === 'string' && /<\s*img\b/i.test(msg)

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
        }

        if (data.toolName && data.toolCallId) {
          if (!toolCallsMap.value.has(data.toolCallId)) {
            toolCallsMap.value.set(data.toolCallId, {
              id: data.toolCallId,
              name: data.toolName,
              showMsg: data.showMsg ? `【请求】\n${data.showMsg}` : '',
              result: false
            })
          } else if (data.toolResult) {
            const tool = toolCallsMap.value.get(data.toolCallId)
            if (tool) {
              const responseText = data.showMsg || data.toolResult || ''
              var isHtml = containsHtml(responseText);
              var breakLine = isHtml ? '<br>' : '\n'
              const combinedMsg = tool.showMsg ? `${tool.showMsg}${breakLine}【响应】${breakLine}${responseText}` : `【响应】${breakLine}${responseText}`
              toolCallsMap.value.set(data.toolCallId, {...tool, showMsg: combinedMsg, result: true})
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

const downloadGuide = async () => {
  if (!downloadUrl.value) {
    ElMessage.error('下载链接不可用')
    return
  }
  try {
    // 兼容相对与绝对 URL，确保解析成功
    const urlObj = new URL(downloadUrl.value, window.location.origin)

    // 将被转义的查询参数解码后再传入，避免二次编码
    const queryParams = {}
    urlObj.searchParams.forEach((v, k) => {
      try {
        queryParams[k] = decodeURIComponent(v)
      } catch (_) {
        queryParams[k] = v
      }
    })

    const filePath = queryParams.filePath
    const response = await defaultApi.apiFileDownloadGet(filePath)

    // 处理文件下载（response 即为文件内容/Blob）
    if (response) {
      // 优先使用后端返回的文件名（若为 File 对象则有 name），否则使用请求的文件名或从路径推断
      const fallbackName = (typeof filePath === 'string' && filePath) ? filePath.split('/').pop() || filePath : "download.pdf"
      const filename = (response && response.name) ? response.name : fallbackName

       // 创建 Blob（response 为 Blob/File 时直接使用，否则包裹）
       const blob = response instanceof Blob ? response : new Blob([response], { type: 'application/octet-stream' })

       // 创建下载链接并触发下载
       const downloadLink = document.createElement('a')
       const objectUrl = URL.createObjectURL(blob)
       downloadLink.href = objectUrl
       downloadLink.download = filename
       document.body.appendChild(downloadLink)
       downloadLink.click()

       // 清理
       URL.revokeObjectURL(objectUrl)
       document.body.removeChild(downloadLink)

       ElMessage.success('下载成功')
     } else {
       ElMessage.error('下载失败：未获取到文件数据')
     }
  } catch (err) {
    console.warn('下载失败', err)
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
  padding: 30px;
  overflow: auto;
  background-color: var(--el-bg-color-page, #f5f7fa);
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
  height: 100%;
  display: flex;
  flex-direction: column;
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
  margin-bottom: 12px;

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

.step-icon {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-fill-color);
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
