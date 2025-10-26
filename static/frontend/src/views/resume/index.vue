<template>
  <div class="resume-container">
    <div class="resume-content">
      <el-card class="resume-card">
        <div class="two-column">
          <!-- 左侧：上传与输入 -->
          <div class="left-panel">
            <div class="panel-section page-intro">
              <h2 class="intro-title">AI 简历优化</h2>
              <p class="intro-subtitle">上传您的简历，并填写目标公司与岗位（可选），右侧实时展示优化进度与结果</p>
            </div>
            <div class="panel-section">
              <h2 class="section-title">上传简历</h2>
              <div class="upload-area">
                <el-upload
                  class="resume-upload"
                  drag
                  action="#"
                  :auto-upload="false"
                  :on-change="handleFileChange"
                  :limit="1"
                  :file-list="fileList"
                  accept=".pdf,.doc,.docx"
                  :disabled="isFormDisabled"
                >
                  <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                  <div class="el-upload__text">
                    拖拽文件到此处或 <em>点击上传</em>
                  </div>
                  <template #tip>
                    <div class="el-upload__tip">
                      支持 PDF、Word 格式的简历文件
                    </div>
                  </template>
                </el-upload>
              </div>
            </div>

            <div class="panel-section">
              <h2 class="section-title">优化选项</h2>
              <el-form :model="formData" label-position="top" class="input-form">
                <el-form-item label="目标公司与岗位（可选）">
                  <el-input
                    v-model="formData.target"
                    placeholder="请输入目标公司和岗位描述（可选）"
                    type="textarea"
                    :rows="10"
                    prefix-icon="Aim"
                    :disabled="isFormDisabled"
                  />
                </el-form-item>
                <el-form-item label="优化要求（可选）">
                  <el-input
                    v-model="formData.requirements"
                    type="textarea"
                    :rows="5"
                    placeholder="请输入其他优化要求，如突出某些技能、经历等"
                    :disabled="isFormDisabled"
                  />
                </el-form-item>
                <div class="action-buttons">
                  <el-button
                    @click="resetProcess"
                    class="reset-button"
                    :icon="RefreshRight"
                    :disabled="isFormDisabled"
                  >
                    重置
                  </el-button>
                  <el-button
                    type="primary"
                    :disabled="!resumeFile || isFormDisabled"
                    @click="startOptimize"
                    class="optimize-button"
                    :icon="Promotion"
                  >
                    开始优化
                  </el-button>
                </div>
              </el-form>
            </div>
          </div>

          <!-- 右侧：过程与结果展示 -->
          <div class="right-panel">
            <div class="panel-content">
              <div v-if="!resumeFile && !isOptimizing && !isCompleted && !hasError" class="empty-state">
                <el-icon class="empty-icon"><Document /></el-icon>
                <p>请上传您的简历开始优化</p>
              </div>

              <div class="optimization-progress" v-else-if="resumeFile || isOptimizing || isCompleted || hasError">
                <div class="progress-header" :class="{'optimizing': isOptimizing, 'completed': isCompleted, 'error': hasError}">
                  <el-icon class="status-icon loading-icon" v-if="isOptimizing"><loading /></el-icon>
                  <el-icon class="status-icon success-icon" v-else-if="isCompleted"><circle-check /></el-icon>
                  <el-icon class="status-icon error-icon" v-else-if="hasError"><circle-close /></el-icon>
                  <el-icon class="status-icon" v-else><info-filled /></el-icon>
                  <span class="status-text" v-if="isOptimizing">AI正在优化您的简历...</span>
                  <span class="status-text" v-else-if="isCompleted">简历优化已完成！</span>
                  <span class="status-text" v-else-if="hasError">简历优化失败</span>
                  <span class="status-text" v-else>准备优化您的简历</span>
                </div>

                <div class="progress-details">
                  <div class="progress-steps">
                    <div v-for="(step, index) in optimizationSteps" :key="index" class="progress-step">
                      <div class="step-icon" :class="{ 'active': step.active, 'completed': step.completed, 'failed': step.failed }">
                        <el-icon v-if="step.completed"><check /></el-icon>
                        <el-icon v-else-if="step.failed"><close /></el-icon>
                        <el-icon v-else-if="step.active"><loading /></el-icon>
                        <span v-else>{{ index + 1 }}</span>
                      </div>
                      <div class="step-info">
                        <div class="step-title">{{ step.title }}</div>
                        <div class="step-description">{{ step.description }}</div>
                      </div>
                    </div>
                  </div>
                </div>

                <el-collapse v-model="detailsActiveNames" v-if="agentReasoning || toolCallsList.length > 0" class="details-collapse">
                  <el-collapse-item title="查看详细过程" name="1">
                    <div class="agent-reasoning" v-if="agentReasoning">
                      <h3 class="detail-title">AI分析过程</h3>
                      <div class="reasoning-content">
                        <pre>{{ agentReasoning }}</pre>
                      </div>
                    </div>

                    <div class="tool-calls" v-if="toolCallsList.length > 0">
                      <h3 class="detail-title">工具调用</h3>
                      <div v-for="(tool, index) in toolCallsList" :key="tool.id" class="tool-call">
                        <div class="tool-header">
                          <span class="tool-name">{{ tool.name }}</span>
                          <span class="tool-status" :class="[ tool.result ? 'completed' : 'loading' ]">
                            <template v-if="tool.result">已完成</template>
                            <template v-else>
                              <el-icon class="status-spinner"><loading /></el-icon>
                              请求中<span class="dot dot-1">.</span><span class="dot dot-2">.</span><span class="dot dot-3">.</span>
                            </template>
                          </span>
                        </div>
                        <div class="tool-message">
                          <pre>{{ tool.showMsg }}</pre>
                        </div>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>

              <div class="optimization-result" v-if="isCompleted">
                <div class="result-header">
                  <el-icon class="success-icon"><circle-check /></el-icon>
                  <h2>简历优化完成！</h2>
                </div>

                <div class="result-summary" v-if="optimizationSummary">
                  <h3 class="summary-title">优化总结</h3>
                  <div class="summary-content" v-html="formattedSummary"></div>
                </div>

                <div class="download-section">
                  <el-button type="primary" @click="downloadResume" :disabled="!downloadUrl" class="download-button">
                    <el-icon><download /></el-icon>
                    下载优化后的简历
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
import { ref, computed, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import {
  UploadFilled,
  Loading,
  Check,
  CircleCheck,
  CircleClose,
  Download,
  Document,
  InfoFilled,
  Close,
  RefreshRight,
  Promotion
} from '@element-plus/icons-vue'
import { marked } from 'marked'
import fetchSSEWithAuth from "@/api/sseRequest.js";
import {useUserStore} from "@/stores/user.js";

// 文件列表
const fileList = ref([])
// 简历文件
const resumeFile = ref(null)
// 表单数据（合并公司与岗位为一个可选输入）
const formData = ref({
  target: '',
  requirements: ''
})
// 优化步骤
const optimizationSteps = ref([
  { title: '分析简历', description: '分析您的简历内容和结构', active: false, completed: false, failed: false },
  { title: '优化内容', description: '针对性调整简历内容和格式', active: false, completed: false, failed: false },
  { title: '生成文档', description: '生成优化后的简历文档', active: false, completed: false, failed: false }
])
// AI推理过程
const agentReasoning = ref('')
// 工具调用映射 (toolCallId -> 工具调用对象)
const toolCallsMap = ref(new Map())
// 用于模板渲染的工具调用数组
const toolCallsList = computed(() => Array.from(toolCallsMap.value.values()))
// 折叠面板默认展开项
const detailsActiveNames = ref(['1'])
// 优化总结
const optimizationSummary = ref('')
// 下载链接
const downloadUrl = ref('')
// SSE事件源/控制器
let abortController = null

// 格式化的总结内容（将Markdown转换为HTML）
const formattedSummary = computed(() => {
  return optimizationSummary.value ? marked(optimizationSummary.value) : ''
})

// 是否正在优化（有任一步骤处于 active）
const isOptimizing = computed(() => optimizationSteps.value.some(step => step.active))
// 是否优化完成（所有步骤 completed）
const isCompleted = computed(() => optimizationSteps.value.length > 0 && optimizationSteps.value.every(step => step.completed))
const hasError = computed(() => optimizationSteps.value.some(step => step.failed))
// 表单是否应该被禁用（在优化过程中禁用）
const isFormDisabled = computed(() => isOptimizing.value && !isCompleted.value && !hasError.value)

// 处理文件变更
const handleFileChange = (file) => {
  fileList.value = [file]
  resumeFile.value = file.raw
}

// 重置流程
const resetProcess = () => {
  fileList.value = []
  resumeFile.value = null
  formData.value = {
    target: '',
    requirements: ''
  }
  optimizationSteps.value.forEach(step => {
    step.active = false
    step.completed = false
    step.failed = false
  })
  agentReasoning.value = ''
  toolCallsMap.value.clear()
  optimizationSummary.value = ''
  downloadUrl.value = ''
}

// 开始优化（公司与岗位输入为可选）
const startOptimize = async () => {
  if (!resumeFile.value) {
    ElMessage.error('请先上传简历文件')
    return
  }

  // 先重置之前的优化状态，保留当前的文件和表单输入
  const currentFile = resumeFile.value
  const currentFormData = { ...formData.value }

  // 重置状态但保留当前文件和表单数据
  optimizationSteps.value.forEach(step => {
    step.active = false
    step.completed = false
    step.failed = false
  })
  agentReasoning.value = ''
  toolCallsMap.value.clear()
  optimizationSummary.value = ''
  downloadUrl.value = ''

  // 构建请求内容（可选项）
  const parts = []
  if (currentFormData.target) parts.push(`目标公司、岗位信息：\n ${currentFormData.target}`)
  if (currentFormData.requirements) parts.push(`用户优化要求：\n ${currentFormData.requirements}`)
  const content = parts.join('\n') || '请优化我的简历'

  // 创建表单数据
  const formDataObj = new FormData()
  formDataObj.append('agentType', 'resume')
  formDataObj.append('content', content)
  formDataObj.append('file', currentFile)

  try {
    // 设置第一个步骤为活动状态
    optimizationSteps.value[0].active = true

    // 通过POST携带FormData并解析SSE流
    const url = '/api/agent/exec/stream'

    // 使用带认证和拦截功能的SSE请求函数
    abortController = await fetchSSEWithAuth(url, formDataObj, function (msg) {
      if (!msg.startsWith('data:')) {
        throw new Error('非法的SSE消息格式')
      }
      const payload = msg.slice(5).trim()
      if (!payload) {
        return
      }
      try {
        const data = JSON.parse(payload)

        if (data.reasoningContent) {
          agentReasoning.value += data.reasoningContent
        }

        if (data.toolName && data.toolCallId) {
          // 处理工具调用
          if (!toolCallsMap.value.has(data.toolCallId)) {
            // 新的工具调用请求
            toolCallsMap.value.set(data.toolCallId, {
              id: data.toolCallId,
              name: data.toolName,
              showMsg: data.showMsg ? `【请求】\n${data.showMsg}` : '',
              result: false
            })
          } else if (data.toolResult) {
            // 工具调用响应
            const tool = toolCallsMap.value.get(data.toolCallId)
            if (tool) {
              // 创建新对象以触发响应式更新
              const responseText = data.showMsg || data.toolResult || ''
              const combinedMsg = tool.showMsg
                ? `${tool.showMsg}\n\n【响应】\n${responseText}`
                : `【响应】\n${responseText}`
              const updatedTool = {
                ...tool,
                showMsg: combinedMsg,
                result: true
              }
              toolCallsMap.value.set(data.toolCallId, updatedTool)
            }
          }

          if (data.toolName === 'markdown_to_pdf_file_tool') {
            optimizationSteps.value[2].active = true
            optimizationSteps.value[0].active = false
            optimizationSteps.value[0].completed = true
            optimizationSteps.value[1].active = false
            optimizationSteps.value[1].completed = true
          }
        }

        if (data.content && !data.end) {
          optimizationSummary.value = data.content

          if (data.content.includes('分析简历完成')) {
            optimizationSteps.value[0].active = false
            optimizationSteps.value[0].completed = true
            optimizationSteps.value[1].active = true
          }
        }

        if (data.end) {
          try {
            abortController?.abort()
          } catch (e) {
            console.error('关闭SSE连接时出错:', e)
          }
          abortController = null

          if (data.error) {
            // 错误处理：将当前活动步骤标记为失败
            const activeStepIndex = optimizationSteps.value.findIndex(step => step.active)
            if (activeStepIndex !== -1) {
              optimizationSteps.value[activeStepIndex].active = false
              optimizationSteps.value[activeStepIndex].failed = true
            }

            // 显示错误提示
            ElMessage.error('简历优化失败：' + (data.error || '未知错误'))
            console.error('简历优化失败:', data.error)
            isOptimizing.value = false
          } else {
            // 成功处理：所有步骤标记为完成
            optimizationSteps.value.forEach(step => {
              step.active = false
              step.completed = true
            })

            if (data.content) {
              downloadUrl.value = data.content
            }
          }
        }
      } catch (e) {
        console.error('处理SSE消息时出错:', e)
      }
    }, function (error) {
      if (isCompleted.value) {
        return
      }
      // 连接或网络层错误：标记流程失败并记录详细错误信息
      try {
        abortController?.abort()
      } catch (e) {
        console.error('关闭SSE连接时出错:', e)
      }
      abortController = null

      // 将当前活动步骤标记为失败
      const activeStepIndex = optimizationSteps.value.findIndex(step => step.active)
      if (activeStepIndex !== -1) {
        optimizationSteps.value[activeStepIndex].active = false
        optimizationSteps.value[activeStepIndex].failed = true
      } else if (optimizationSteps.value.length > 0) {
        // 若无活动步骤，标记第一个步骤为失败以触发错误状态展示
        optimizationSteps.value[0].failed = true
      }

      // 汇总错误信息并展示到详细过程
      const errText = typeof error === 'string' ? error : (error?.message || JSON.stringify(error))

      // 将错误信息附加到尚未完成的工具调用上，便于定位问题
      try {
        for (const tool of toolCallsMap.value.values()) {
          if (!tool.result) {
            const updatedTool = {
              ...tool,
              showMsg: tool.showMsg ? `${tool.showMsg}\n\n【错误】\n${errText}` : `【错误】\n${errText}`
            }
            toolCallsMap.value.set(tool.id, updatedTool)
          }
        }
      } catch (e) {
        console.error('更新工具调用错误信息时出错:', e)
      }

      ElMessage.error('系统繁忙，请稍后重试')
      console.error('SSE请求出错:', error)
    }, function () {
      console.log('SSE连接关闭')
    })
  } catch (error) {
    console.error('优化简历时出错:', error)
    ElMessage.error('优化简历失败，请稍后重试')
  }
}

// TODO 后续抽取公共的下载工具函数，实现 token 刷新逻辑等
const downloadResume = async () => {
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

// 组件销毁前关闭SSE连接
onBeforeUnmount(() => {
  abortController?.abort()
  abortController = null
})
</script>

<style scoped lang="scss">
.resume-container {
  width: 100%;
  max-width: none;
  margin: 0;
  padding-top: 5px;
  height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
  background-color: var(--el-bg-color-page, #f5f7fa);
}

.page-intro {
  //margin-bottom: 10px;

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

.resume-card {
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

/* 使内容区域填充剩余空间 */
.resume-content {
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

.panel-section {
  //margin-bottom: 25px;

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

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;

  .resume-upload {
    width: 100%;

    :deep(.el-upload-dragger) {
      width: 100%;
      height: 180px;
      display: flex;
      align-items: center;
      justify-content: center;
      border: 2px dashed var(--el-border-color);
      border-radius: 12px;
      transition: all 0.3s;

      &:hover {
        border-color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
      }
    }
  }

  .upload-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    .el-icon--upload {
      font-size: 48px;
      color: var(--el-color-primary);
      margin-bottom: 15px;
    }

    .el-upload__text {
      font-size: 16px;
      margin-bottom: 10px;

      em {
        color: var(--el-color-primary);
        font-style: normal;
        font-weight: bold;
      }
    }
  }

  .el-upload__tip {
    margin-top: 10px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
}

.input-form {
  :deep(.el-form-item__label) {
    font-weight: 500;
    padding-bottom: 8px;
  }

  :deep(.el-input__wrapper),
  :deep(.el-textarea__inner) {
    border-radius: 8px;
    box-shadow: 0 0 0 1px var(--el-border-color-light) inset;
    transition: all 0.3s;

    &:hover, &:focus {
      box-shadow: 0 0 0 1px var(--el-color-primary-light-5) inset;
    }
  }

  :deep(.el-input__prefix-inner) {
    color: var(--el-color-primary);
  }
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-top: 30px;

  .reset-button, .optimize-button {
    min-width: 120px;
    border-radius: 8px;
    font-weight: 500;
    transition: all 0.3s;

    .el-icon {
      margin-right: 6px;
    }
  }

  .optimize-button {
    background: linear-gradient(45deg, var(--el-color-primary), var(--el-color-primary-light-3));
    border: none;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 5px 15px rgba(var(--el-color-primary-rgb), 0.3);
    }

    &:disabled {
      background: var(--el-button-disabled-bg-color);
      transform: none;
      box-shadow: none;
    }
  }
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
    margin-bottom: 20px;
    color: var(--el-color-info-light-5);
  }

  p {
    font-size: 16px;
  }
}

.optimization-progress {
  //:deep(.el-collapse-item__header) {
  //  border-radius: initial;
  //  border-top-left-radius: 12px;
  //  border-top-right-radius: 12px;
  //}
  //.details-collapse {
  //  .el-collapse-item__header {
  //    border-radius: initial;
  //    border-top-left-radius: 12px;
  //    border-top-right-radius: 12px;
  //  }
  //}

  .progress-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    margin-bottom: 30px;
    font-size: 18px;
    padding: 20px;
    //border-radius: 12px;
    background-color: var(--el-bg-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    transition: all 0.3s;

    &.optimizing {
      background-color: var(--el-color-primary-light-9);
    }

    &.completed {
      background-color: var(--el-color-success-light-9);
    }

    &.error {
      background-color: var(--el-color-danger-light-9);
    }

    .status-icon {
      font-size: 28px;

      &.loading-icon {
        color: var(--el-color-primary);
        animation: rotate 1s linear infinite;
      }

      &.success-icon {
        color: var(--el-color-success);
      }

      &.error-icon {
        color: var(--el-color-danger);
      }
    }

    .status-text {
      font-weight: 500;
    }
  }

  .progress-details {
    margin-bottom: 30px;

    .progress-steps {
      .progress-step {
        display: flex;
        align-items: flex-start;
        margin-bottom: 20px;
        padding: 15px;
        border-radius: 12px;
        background-color: var(--el-bg-color);
        transition: all 0.3s;

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 5px 15px rgba(0, 0, 0, 0.05);
        }

        .step-icon {
          width: 36px;
          height: 36px;
          border-radius: 50%;
          background-color: var(--el-color-info-light-9);
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 15px;
          flex-shrink: 0;
          font-weight: bold;
          transition: all 0.3s;

          &.active {
            background-color: var(--el-color-primary);
            color: white;
            animation: pulse 1.5s infinite;
          }

          &.completed {
            background-color: var(--el-color-success);
            color: white;
          }

          &.failed {
            background-color: var(--el-color-danger);
            color: white;
          }
        }

        .step-info {
          .step-title {
            font-weight: bold;
            margin-bottom: 5px;
            font-size: 16px;
          }

          .step-description {
            color: var(--el-text-color-secondary);
            font-size: 14px;
          }
        }
      }
    }
  }

  .details-collapse {
    margin-top: 20px;
    border-radius: 12px;
    overflow: hidden;

    :deep(.el-collapse-item__header) {
      font-weight: 500;
      font-size: 16px;
      padding: 15px;
      background-color: var(--el-bg-color);
      //border-radius: 12px;
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
}

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

      :deep(h2), :deep(h3) {
        margin-top: 0;
      }

      :deep(ul) {
        padding-left: 20px;
      }

      :deep(p) {
        margin-bottom: 12px;
        line-height: 1.6;
      }

      :deep(a) {
        color: var(--el-color-primary);
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
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

      .el-icon {
        margin-right: 8px;
        font-size: 18px;
      }
    }
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(var(--el-color-primary-rgb), 0.4);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(var(--el-color-primary-rgb), 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(var(--el-color-primary-rgb), 0);
  }
}

@keyframes blink {
  0% { opacity: 0; }
  50% { opacity: 1; }
  100% { opacity: 0; }
}

// 响应式设计
@media screen and (max-width: 1200px) {
  .two-column {
    grid-template-columns: 1fr 1.2fr;
  }
}

@media screen and (max-width: 992px) {
  .resume-container {
    height: auto;
    min-height: calc(100vh - 64px);
  }

  .two-column {
    grid-template-columns: 1fr;
    gap: 30px;
  }

  .left-panel {
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-light);
    padding-bottom: 30px;
  }

  .right-panel {
    padding-top: 0;
  }
}

@media screen and (max-width: 768px) {
  .resume-container {
    padding: 15px 10px;
  }

  .left-panel, .right-panel {
    padding: 20px;
  }

  .action-buttons {
    flex-direction: column;

    .reset-button, .optimize-button {
      width: 100%;
    }
  }

  .optimization-result {
    padding: 20px;
  }
}

@media screen and (max-width: 480px) {
  .page-intro .intro-title {
    font-size: 20px;
  }

  .progress-step {
     flex-direction: column;

     .step-icon {
       margin-bottom: 10px;
       margin-right: 0;
     }
   }
 }
</style>
