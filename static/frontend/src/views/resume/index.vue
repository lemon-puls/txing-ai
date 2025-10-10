<template>
  <div class="resume-container">
    <div class="resume-header">
      <h1 class="title">AI简历优化</h1>
      <p class="subtitle">上传您的简历，AI将为您进行针对性优化</p>
    </div>

    <div class="resume-content">
      <el-card class="resume-card">
        <div class="steps-container">
          <el-steps :active="currentStep" finish-status="success" align-center>
            <el-step title="上传简历" />
            <el-step title="填写信息" />
            <el-step title="优化中" />
            <el-step title="查看结果" />
          </el-steps>
        </div>

        <!-- 步骤1: 上传简历 -->
        <div v-if="currentStep === 0" class="step-content">
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
          <div class="step-actions">
            <el-button type="primary" :disabled="!resumeFile" @click="nextStep">下一步</el-button>
          </div>
        </div>

        <!-- 步骤2: 填写信息 -->
        <div v-if="currentStep === 1" class="step-content">
          <el-form :model="formData" label-position="top">
            <el-form-item label="目标公司">
              <el-input v-model="formData.company" placeholder="请输入目标公司名称" />
            </el-form-item>
            <el-form-item label="目标岗位">
              <el-input v-model="formData.position" placeholder="请输入目标岗位名称" />
            </el-form-item>
            <el-form-item label="其他要求（可选）">
              <el-input
                v-model="formData.requirements"
                type="textarea"
                :rows="4"
                placeholder="请输入其他优化要求，如突出某些技能、经历等"
              />
            </el-form-item>
          </el-form>
          <div class="step-actions">
            <el-button @click="prevStep">上一步</el-button>
            <el-button
              type="primary"
              :disabled="!formData.company || !formData.position"
              @click="startOptimize"
            >开始优化</el-button>
          </div>
        </div>

        <!-- 步骤3: 优化中 -->
        <div v-if="currentStep === 2" class="step-content">
          <div class="optimization-progress">
            <div class="progress-header">
              <el-icon class="loading-icon"><loading /></el-icon>
              <span>AI正在优化您的简历...</span>
            </div>

            <div class="progress-details">
              <div class="progress-steps">
                <div v-for="(step, index) in optimizationSteps" :key="index" class="progress-step">
                  <div class="step-icon" :class="{ 'active': step.active, 'completed': step.completed }">
                    <el-icon v-if="step.completed"><check /></el-icon>
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

            <div class="agent-reasoning" v-if="agentReasoning">
              <h3>AI分析过程</h3>
              <div class="reasoning-content">
                <pre>{{ agentReasoning }}</pre>
              </div>
            </div>

            <div class="tool-calls" v-if="toolCalls.length > 0">
              <h3>工具调用</h3>
              <div v-for="(tool, index) in toolCalls" :key="index" class="tool-call">
                <div class="tool-header">
                  <span class="tool-name">{{ tool.name }}</span>
                </div>
                <div class="tool-params" v-if="tool.params">
                  <pre>{{ formatJSON(tool.params) }}</pre>
                </div>
                <div class="tool-result" v-if="tool.result">
                  <pre>{{ tool.result }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 步骤4: 查看结果 -->
        <div v-if="currentStep === 3" class="step-content">
          <div class="optimization-result">
            <div class="result-header">
              <el-icon class="success-icon"><circle-check /></el-icon>
              <h2>简历优化完成！</h2>
            </div>

            <div class="result-summary" v-if="optimizationSummary">
              <h3>优化总结</h3>
              <div class="summary-content" v-html="formattedSummary"></div>
            </div>

            <div class="download-section">
              <h3>下载优化后的简历</h3>
              <el-button type="primary" @click="downloadResume" :disabled="!downloadUrl">
                <el-icon><download /></el-icon>
                下载简历
              </el-button>
            </div>

            <div class="action-buttons">
              <el-button @click="resetProcess">重新开始</el-button>
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
  Download
} from '@element-plus/icons-vue'
import {agentApi, defaultApi} from '@/api'
import { marked } from 'marked'

// 当前步骤
const currentStep = ref(0)
// 文件列表
const fileList = ref([])
// 简历文件
const resumeFile = ref(null)
// 表单数据
const formData = ref({
  company: '',
  position: '',
  requirements: ''
})
// 优化步骤
const optimizationSteps = ref([
  { title: '分析简历', description: '分析您的简历内容和结构', active: false, completed: false },
  { title: '匹配岗位', description: '根据目标岗位提取关键要求', active: false, completed: false },
  { title: '优化内容', description: '针对性调整简历内容和格式', active: false, completed: false },
  { title: '生成文档', description: '生成优化后的简历文档', active: false, completed: false }
])
// AI推理过程
const agentReasoning = ref('')
// 工具调用
const toolCalls = ref([])
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

// 处理文件变更
const handleFileChange = (file) => {
  fileList.value = [file]
  resumeFile.value = file.raw
}

// 下一步
const nextStep = () => {
  currentStep.value++
}

// 上一步
const prevStep = () => {
  currentStep.value--
}

// 重置流程
const resetProcess = () => {
  currentStep.value = 0
  fileList.value = []
  resumeFile.value = null
  formData.value = {
    company: '',
    position: '',
    requirements: ''
  }
  optimizationSteps.value.forEach(step => {
    step.active = false
    step.completed = false
  })
  agentReasoning.value = ''
  toolCalls.value = []
  optimizationSummary.value = ''
  downloadUrl.value = ''
}

// 格式化JSON
const formatJSON = (jsonString) => {
  try {
    const obj = typeof jsonString === 'string' ? JSON.parse(jsonString) : jsonString
    return JSON.stringify(obj, null, 2)
  } catch (e) {
    return jsonString
  }
}

// 开始优化
const startOptimize = async () => {
  if (!resumeFile.value) {
    ElMessage.error('请先上传简历文件')
    return
  }

  if (!formData.value.company || !formData.value.position) {
    ElMessage.error('请填写目标公司和岗位')
    return
  }

  // 更新步骤
  currentStep.value = 2

  // 构建请求内容
  const content = `目标公司：${formData.value.company}\n目标岗位：${formData.value.position}${formData.value.requirements ? '\n其他要求：' + formData.value.requirements : ''}`

  // 创建表单数据
  const formDataObj = new FormData()
  formDataObj.append('agentType', 'resume')
  formDataObj.append('content', content)
  formDataObj.append('file', resumeFile.value)

  try {
    // 设置第一个步骤为活动状态
    optimizationSteps.value[0].active = true

    // 通过POST携带FormData并解析SSE流
    const url = '/api/api/agent/exec/stream'
    abortController = new AbortController()
    const response = await fetch(url, {
      method: 'POST',
      headers: { Accept: 'text/event-stream' },
      body: formDataObj,
      signal: abortController.signal
    })

    if (!response.ok || !response.body) {
      throw new Error('连接服务器失败')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''

    const processBuffer = () => {
      const parts = buffer.split('\n\n')
      buffer = parts.pop() || ''
      for (const part of parts) {
        const lines = part.split('\n')
        for (const line of lines) {
          if (!line.startsWith('data:')) continue
          const payload = line.slice(5).trim()
          if (!payload) continue
          try {
            const data = JSON.parse(payload)

            if (data.reasoningContent) {
              agentReasoning.value += data.reasoningContent
            }

            if (data.toolName) {
              toolCalls.value.push({
                name: data.toolName,
                params: data.toolParams,
                result: data.toolResult
              })

              if (data.toolName === 'markdown_to_pdf_file_tool') {
                optimizationSteps.value[3].active = true
                optimizationSteps.value[0].active = false
                optimizationSteps.value[0].completed = true
                optimizationSteps.value[1].active = false
                optimizationSteps.value[1].completed = true
                optimizationSteps.value[2].active = false
                optimizationSteps.value[2].completed = true
              }
            }

            if (data.content && !data.end) {
              optimizationSummary.value = data.content

              if (data.content.includes('分析简历')) {
                optimizationSteps.value[0].active = true
              } else if (data.content.includes('匹配岗位')) {
                optimizationSteps.value[0].active = false
                optimizationSteps.value[0].completed = true
                optimizationSteps.value[1].active = true
              } else if (data.content.includes('优化内容')) {
                optimizationSteps.value[1].active = false
                optimizationSteps.value[1].completed = true
                optimizationSteps.value[2].active = true
              }
            }

            if (data.end) {
              try {
                abortController?.abort()
              } catch (e) {
                console.error('关闭SSE连接时出错:', e)
              }
              abortController = null

              optimizationSteps.value.forEach(step => {
                step.active = false
                step.completed = true
              })

              if (data.content) {
                downloadUrl.value = data.content
              }

              currentStep.value = 3
            }
          } catch (e) {
            console.error('处理SSE消息时出错:', e)
          }
        }
      }
    }

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      processBuffer()
    }
    processBuffer()
  } catch (error) {
    console.error('优化简历时出错:', error)
    ElMessage.error('优化简历失败，请稍后重试')
  }
}

// 下载简历
const downloadResume = () => {
  if (downloadUrl.value) {
    window.open(downloadUrl.value, '_blank')
  } else {
    ElMessage.error('下载链接不可用')
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
  max-width: 1000px;
  margin: 0 auto;
  padding: 40px 20px;
}

.resume-header {
  text-align: center;
  margin-bottom: 40px;

  .title {
    font-size: 32px;
    font-weight: bold;
    margin-bottom: 10px;
    background: linear-gradient(45deg, var(--el-color-primary), var(--el-color-primary-light-3));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .subtitle {
    font-size: 16px;
    color: var(--el-text-color-secondary);
  }
}

.resume-card {
  border-radius: 20px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.steps-container {
  margin-bottom: 30px;
}

.step-content {
  padding: 20px 0;
  min-height: 300px;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  .resume-upload {
    width: 100%;
    max-width: 500px;
  }

  .el-upload__tip {
    margin-top: 10px;
    color: var(--el-text-color-secondary);
  }
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-top: 30px;
}

.optimization-progress {
  .progress-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 30px;
    font-size: 18px;

    .loading-icon {
      font-size: 24px;
      animation: rotate 1s linear infinite;
    }
  }

  .progress-steps {
    margin-bottom: 30px;

    .progress-step {
      display: flex;
      align-items: flex-start;
      margin-bottom: 20px;

      .step-icon {
        width: 30px;
        height: 30px;
        border-radius: 50%;
        background-color: var(--el-color-info-light-9);
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 15px;
        flex-shrink: 0;

        &.active {
          background-color: var(--el-color-primary);
          color: white;
          animation: pulse 1.5s infinite;
        }

        &.completed {
          background-color: var(--el-color-success);
          color: white;
        }
      }

      .step-info {
        .step-title {
          font-weight: bold;
          margin-bottom: 5px;
        }

        .step-description {
          color: var(--el-text-color-secondary);
          font-size: 14px;
        }
      }
    }
  }

  .agent-reasoning {
    margin-top: 30px;
    border-top: 1px solid var(--el-border-color-lighter);
    padding-top: 20px;

    h3 {
      margin-bottom: 15px;
    }

    .reasoning-content {
      background-color: var(--el-color-info-light-9);
      border-radius: 8px;
      padding: 15px;
      max-height: 200px;
      overflow-y: auto;

      pre {
        white-space: pre-wrap;
        word-break: break-word;
        font-family: monospace;
        margin: 0;
      }
    }
  }

  .tool-calls {
    margin-top: 30px;
    border-top: 1px solid var(--el-border-color-lighter);
    padding-top: 20px;

    h3 {
      margin-bottom: 15px;
    }

    .tool-call {
      background-color: var(--el-color-info-light-9);
      border-radius: 8px;
      padding: 15px;
      margin-bottom: 15px;

      .tool-header {
        margin-bottom: 10px;

        .tool-name {
          font-weight: bold;
          color: var(--el-color-primary);
        }
      }

      .tool-params, .tool-result {
        background-color: var(--el-bg-color);
        border-radius: 4px;
        padding: 10px;
        margin-bottom: 10px;

        pre {
          white-space: pre-wrap;
          word-break: break-word;
          font-family: monospace;
          margin: 0;
          font-size: 12px;
        }
      }
    }
  }
}

.optimization-result {
  .result-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 30px;

    .success-icon {
      font-size: 48px;
      color: var(--el-color-success);
      margin-bottom: 15px;
    }

    h2 {
      font-size: 24px;
      margin: 0;
    }
  }

  .result-summary {
    margin-bottom: 30px;

    h3 {
      margin-bottom: 15px;
    }

    .summary-content {
      background-color: var(--el-color-info-light-9);
      border-radius: 8px;
      padding: 20px;

      :deep(h2), :deep(h3) {
        margin-top: 0;
      }

      :deep(ul) {
        padding-left: 20px;
      }
    }
  }

  .download-section {
    text-align: center;
    margin-bottom: 30px;

    h3 {
      margin-bottom: 15px;
    }
  }

  .action-buttons {
    display: flex;
    justify-content: center;
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

// 响应式设计
@media screen and (max-width: 768px) {
  .resume-container {
    padding: 20px 10px;
  }

  .resume-header {
    margin-bottom: 20px;

    .title {
      font-size: 24px;
    }

    .subtitle {
      font-size: 14px;
    }
  }

  .step-content {
    min-height: 200px;
  }
}
</style>
