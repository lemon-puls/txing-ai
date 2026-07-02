<template>
  <div class="property-panel" v-if="selectedNode">
    <div class="panel-header">
      <div class="header-title">
        <el-icon><Setting /></el-icon>
        <span>节点配置</span>
      </div>
      <el-button text @click="$emit('close')">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>

    <div class="panel-content">
      <!-- 基础配置 -->
      <div class="config-section">
        <div class="section-title">基础信息</div>
        <el-form label-position="top" size="small">
          <el-form-item label="节点名称">
            <el-input v-model="localData.label" placeholder="请输入节点名称" />
          </el-form-item>
          <el-form-item label="节点描述">
            <el-input v-model="localData.description" type="textarea" :rows="2" placeholder="请输入节点描述" />
          </el-form-item>
        </el-form>
      </div>

      <!-- LLM 节点配置 -->
      <div class="config-section" v-if="nodeType === 'llm'">
        <div class="section-title">模型配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="选择模型">
            <el-select v-model="localData.modelConfig.model" placeholder="请选择模型" style="width: 100%">
              <el-option v-for="model in modelList" :key="model.name" :label="model.name" :value="model.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="系统提示词">
            <el-input
              v-model="localData.modelConfig.systemPrompt"
              type="textarea"
              :rows="4"
              placeholder="请输入系统提示词"
            />
          </el-form-item>
          <el-form-item label="温度 (Temperature)">
            <el-slider v-model="localData.modelConfig.temperature" :min="0" :max="2" :step="0.1" show-input />
          </el-form-item>
          <el-form-item label="最大Token数">
            <el-input-number v-model="localData.modelConfig.maxTokens" :min="100" :max="32000" :step="100" style="width: 100%" />
          </el-form-item>
          <el-form-item label="启用上下文记忆">
            <el-switch v-model="localData.modelConfig.contextEnabled" />
          </el-form-item>
          <el-form-item label="绑定工具">
            <el-checkbox-group v-model="localData.modelConfig.tools">
              <el-checkbox v-for="tool in toolList" :key="tool.name" :label="tool.name">
                {{ tool.displayName }}
              </el-checkbox>
            </el-checkbox-group>
            <div class="form-tip">LLM 将通过 Function Calling 自主决定是否调用这些工具</div>
          </el-form-item>
        </el-form>
      </div>

      <!-- 工具节点配置 -->
      <div class="config-section" v-if="nodeType === 'tool'">
        <div class="section-title">工具配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="选择工具">
            <el-select v-model="localData.toolConfig.toolName" placeholder="请选择工具" style="width: 100%" clearable>
              <el-option v-for="tool in toolList" :key="tool.name" :label="tool.displayName" :value="tool.name" />
            </el-select>
            <div class="form-tip">直接执行工具，不经过大模型</div>
          </el-form-item>
          <el-form-item label="工具参数">
            <el-input
              v-model="toolParamsStr"
              type="textarea"
              :rows="4"
              placeholder='JSON 格式参数'
              style="font-family: monospace;"
              @change="parseToolParams"
            />
            <div class="form-tip">输入 JSON 格式的工具参数</div>
          </el-form-item>
        </el-form>
      </div>

      <!-- 条件节点配置 -->
      <div class="config-section" v-if="nodeType === 'condition'">
        <div class="section-title">条件配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="条件类型">
            <el-radio-group v-model="localData.conditionConfig.type">
              <el-radio label="expression">表达式判断</el-radio>
              <el-radio label="llm">AI判断</el-radio>
              <el-radio label="tool_result">工具结果</el-radio>
            </el-radio-group>
          </el-form-item>

          <template v-if="localData.conditionConfig.type === 'expression'">
            <el-form-item label="条件表达式">
              <el-input v-model="localData.conditionConfig.expression" placeholder="例如: {{output}} contains '成功'" />
              <div class="form-tip">支持运算符：contains, equals, matches 等</div>
            </el-form-item>
          </template>

          <template v-if="localData.conditionConfig.type === 'llm'">
            <el-form-item label="判断提示词">
              <el-input
                v-model="localData.conditionConfig.llmPrompt"
                type="textarea"
                :rows="3"
                placeholder="请输入让AI判断的提示词"
              />
            </el-form-item>
          </template>

          <template v-if="localData.conditionConfig.type === 'tool_result'">
            <el-form-item label="工具名称">
              <el-select v-model="localData.conditionConfig.toolName" placeholder="请选择工具">
                <el-option v-for="tool in toolList" :key="tool.name" :label="tool.displayName" :value="tool.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="期望值">
              <el-input v-model="localData.conditionConfig.expectedValue" placeholder="例如: success, 200" />
            </el-form-item>
          </template>

          <div class="divider"></div>
          <div class="sub-section-title">错误处理</div>
          <el-form-item label="判断失败时的处理">
            <el-radio-group v-model="localData.conditionConfig.failureAction">
              <el-radio label="default_false">默认走 false 分支</el-radio>
              <el-radio label="terminate">终止工作流</el-radio>
              <el-radio label="configurable">自定义默认分支</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
      </div>

      <!-- 代码节点配置 -->
      <div class="config-section" v-if="nodeType === 'code'">
        <div class="section-title">代码配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="编程语言">
            <el-select v-model="localData.codeConfig.language" placeholder="请选择语言" style="width: 100%">
              <el-option label="JavaScript" value="javascript" />
              <el-option label="Python" value="python" />
            </el-select>
          </el-form-item>
          <el-form-item label="代码内容">
            <el-input
              v-model="localData.codeConfig.code"
              type="textarea"
              :rows="10"
              placeholder="在此编写代码"
              style="font-family: monospace;"
            />
            <div class="form-tip">可用变量：input, output，使用 return 返回结果</div>
          </el-form-item>
          <el-form-item label="超时时间（秒）">
            <el-input-number v-model="localData.codeConfig.timeout" :min="1" :max="300" :step="1" style="width: 100%" />
          </el-form-item>
        </el-form>
      </div>

      <!-- HTTP 节点配置 -->
      <div class="config-section" v-if="nodeType === 'http'">
        <div class="section-title">HTTP 配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="请求方法">
            <el-select v-model="localData.httpConfig.method" placeholder="请选择方法" style="width: 100%">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
            </el-select>
          </el-form-item>
          <el-form-item label="请求 URL">
            <el-input v-model="localData.httpConfig.url" placeholder="例如: https://api.example.com/data" />
            <div class="form-tip">支持变量替换：{{input}}、{{output}}</div>
          </el-form-item>
          <el-form-item label="请求头">
            <el-input
              v-model="httpHeadersStr"
              type="textarea"
              :rows="3"
              placeholder='每行一个，格式: Key: Value'
              @change="parseHttpHeaders"
            />
          </el-form-item>
          <el-form-item label="请求体" v-if="localData.httpConfig.method !== 'GET'">
            <el-input
              v-model="localData.httpConfig.body"
              type="textarea"
              :rows="4"
              placeholder="请求体内容（支持 JSON）"
            />
          </el-form-item>
          <el-form-item label="超时时间（秒）">
            <el-input-number v-model="localData.httpConfig.timeout" :min="1" :max="300" :step="1" style="width: 100%" />
          </el-form-item>
        </el-form>
      </div>

      <!-- Agent 节点配置 -->
      <div class="config-section" v-if="nodeType === 'agent'">
        <div class="section-title">Agent 配置</div>
        <el-form label-position="top" size="small">
          <div class="sub-section-title">模型配置</div>
          <el-form-item label="选择模型">
            <el-select v-model="localData.modelConfig.model" placeholder="请选择模型" style="width: 100%" clearable>
              <el-option v-for="model in modelList" :key="model.name" :label="model.name" :value="model.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="系统提示词">
            <el-input
              v-model="localData.modelConfig.systemPrompt"
              type="textarea"
              :rows="4"
              placeholder="指导 Agent 的行为"
            />
          </el-form-item>
          <el-form-item label="温度 (Temperature)">
            <el-slider v-model="localData.modelConfig.temperature" :min="0" :max="2" :step="0.1" show-input />
          </el-form-item>
          <el-form-item label="绑定工具">
            <el-checkbox-group v-model="localData.modelConfig.tools">
              <el-checkbox v-for="tool in toolList" :key="tool.name" :label="tool.name">
                {{ tool.displayName }}
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>

          <div class="divider"></div>
          <div class="sub-section-title">Agent 行为</div>
          <el-form-item label="系统提示词 (Agent 兜底)">
            <el-input
              v-model="localData.agentConfig.systemPrompt"
              type="textarea"
              :rows="4"
              placeholder="当上方模型配置的系统提示词为空时使用"
            />
          </el-form-item>
          <el-form-item label="选择工具 (Agent 兜底)">
            <el-checkbox-group v-model="localData.agentConfig.tools">
              <el-checkbox v-for="tool in toolList" :key="tool.name" :label="tool.name">
                {{ tool.displayName }}
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="最大执行步数">
            <el-input-number v-model="localData.agentConfig.maxRunSteps" :min="1" :max="200" :step="1" style="width: 100%" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 并行组节点配置 -->
      <div class="config-section" v-if="nodeType === 'parallel'">
        <div class="section-title">并行组配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="说明">
            <div class="info-tip">
              并行组用于定义并行执行区域。将需要并行执行的节点连接到并行组的输出端。
            </div>
          </el-form-item>
          <el-form-item label="超时时间（秒）">
            <el-input-number v-model="localData.parallelConfig.timeout" :min="0" :step="10" style="width: 100%" />
            <div class="form-tip">0 表示无超时限制</div>
          </el-form-item>
        </el-form>
      </div>

      <!-- 汇聚节点配置 -->
      <div class="config-section" v-if="nodeType === 'join'">
        <div class="section-title">汇聚配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="汇聚策略">
            <el-select v-model="localData.joinConfig.strategy" placeholder="请选择策略" style="width: 100%">
              <el-option label="全部完成" value="all" />
              <el-option label="任一完成" value="any" />
            </el-select>
          </el-form-item>
          <el-form-item label="超时时间（秒）">
            <el-input-number v-model="localData.joinConfig.timeout" :min="0" :step="10" style="width: 100%" />
            <div class="form-tip">0 表示无超时限制</div>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <div class="panel-footer">
      <el-button type="primary" size="small" @click="handleApply" round>
        <el-icon><Check /></el-icon>
        应用配置
      </el-button>
    </div>
  </div>

  <!-- 未选中节点时的提示 -->
  <div class="property-panel empty" v-else>
    <div class="empty-content">
      <el-icon :size="48"><Document /></el-icon>
      <p>点击节点查看配置</p>
      <p class="tip">拖拽节点到画布开始设计工作流</p>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { Setting, Close, Document, Check } from '@element-plus/icons-vue'

const props = defineProps({
  selectedNode: {
    type: Object,
    default: null
  },
  modelList: {
    type: Array,
    default: () => []
  },
  toolList: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update', 'close'])

const nodeType = computed(() => props.selectedNode?.data?.nodeType)

const localData = ref({
  label: '',
  description: '',
  modelConfig: {
    model: '',
    systemPrompt: '',
    temperature: 0.7,
    maxTokens: 4096,
    contextEnabled: true,
    tools: [],
    maxToolRounds: 5
  },
  toolConfig: {
    toolName: '',
    params: {},
    tools: []
  },
  conditionConfig: {
    type: 'expression',
    expression: '',
    llmPrompt: '',
    toolName: '',
    toolResultKey: '',
    expectedValue: '',
    failureAction: 'default_false',
    failureBranch: 'false'
  },
  codeConfig: {
    language: 'javascript',
    code: '',
    timeout: 30
  },
  httpConfig: {
    method: 'GET',
    url: '',
    headers: {},
    body: '',
    timeout: 30
  },
  agentConfig: {
    systemPrompt: '',
    tools: [],
    maxRunSteps: 30
  },
  parallelConfig: {
    maxConcurrency: 3,
    waitStrategy: 'all',
    timeout: 60
  },
  joinConfig: {
    strategy: 'all',
    timeout: 60
  }
})

const httpHeadersStr = ref('')
const toolParamsStr = ref('')

watch(() => props.selectedNode, (newNode) => {
  if (newNode) {
    localData.value = {
      label: newNode.data?.label || '',
      description: newNode.data?.description || '',
      modelConfig: newNode.data?.modelConfig || { model: '', systemPrompt: '', temperature: 0.7, maxTokens: 4096, contextEnabled: true, tools: [], maxToolRounds: 5 },
      toolConfig: newNode.data?.toolConfig || { toolName: '', params: {}, tools: [] },
      conditionConfig: newNode.data?.conditionConfig || { type: 'expression', expression: '', llmPrompt: '', toolName: '', toolResultKey: '', expectedValue: '', failureAction: 'default_false', failureBranch: 'false' },
      codeConfig: newNode.data?.codeConfig || { language: 'javascript', code: '', timeout: 30 },
      httpConfig: newNode.data?.httpConfig || { method: 'GET', url: '', headers: {}, body: '', timeout: 30 },
      agentConfig: newNode.data?.agentConfig || { systemPrompt: '', tools: [], maxRunSteps: 30 },
      parallelConfig: newNode.data?.parallelConfig || { maxConcurrency: 3, waitStrategy: 'all', timeout: 60 },
      joinConfig: newNode.data?.joinConfig || { strategy: 'all', timeout: 60 }
    }
    const headers = localData.value.httpConfig.headers || {}
    httpHeadersStr.value = Object.entries(headers).map(([key, value]) => `${key}: ${value}`).join('\n')
    const params = localData.value.toolConfig.params || {}
    toolParamsStr.value = Object.keys(params).length > 0 ? JSON.stringify(params, null, 2) : ''
  }
}, { immediate: true })

const parseHttpHeaders = (str) => {
  const headers = {}
  if (str) {
    str.split('\n').forEach(line => {
      const colonIndex = line.indexOf(':')
      if (colonIndex > 0) {
        const key = line.substring(0, colonIndex).trim()
        const value = line.substring(colonIndex + 1).trim()
        if (key && value) {
          headers[key] = value
        }
      }
    })
  }
  localData.value.httpConfig.headers = headers
}

const parseToolParams = () => {
  try {
    const parsed = JSON.parse(toolParamsStr.value)
    localData.value.toolConfig.params = parsed
  } catch (e) {
    console.warn('工具参数 JSON 解析失败:', e.message)
  }
}

const handleApply = () => {
  emit('update', {
    id: props.selectedNode.id,
    data: { ...localData.value, nodeType: nodeType.value }
  })
}
</script>

<style lang="scss" scoped>
// 设计变量 / Design Variables (浅色主题)
$primary-color: #3b82f6;
$primary-light: #60a5fa;
$primary-dark: #2563eb;
$success-color: #10b981;
$warning-color: #f59e0b;
$danger-color: #ef4444;
$bg-white: #ffffff;
$bg-light: #f8fafc;
$bg-card: #f1f5f9;
$border-color: #e2e8f0;
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.property-panel {
  width: 360px;
  height: 100%;
  background: $bg-white;
  border-left: 1px solid $border-color;
  display: flex;
  flex-direction: column;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.04);

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 18px 20px;
    border-bottom: 1px solid $border-color;
    background: $bg-light;

    .header-title {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 15px;
      font-weight: 600;
      color: $text-primary;

      .el-icon {
        color: $primary-color;
        font-size: 18px;
      }
    }
    
    .el-button {
      border-radius: 8px;
      color: $text-secondary;
      background: transparent;
      border: none;
      
      &:hover {
        background: rgba($primary-color, 0.08);
        color: $primary-color;
      }
    }
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;

    &::-webkit-scrollbar {
      width: 6px;
    }
    
    &::-webkit-scrollbar-thumb {
      background: $border-color;
      border-radius: 3px;
    }

    .config-section {
      margin-bottom: 28px;
      padding-bottom: 20px;
      border-bottom: 1px solid $border-color;

      &:last-child {
        border-bottom: none;
        margin-bottom: 0;
      }

      .section-title {
        font-size: 12px;
        font-weight: 600;
        color: $primary-color;
        margin-bottom: 16px;
        padding-left: 12px;
        border-left: 3px solid $primary-color;
        display: flex;
        align-items: center;
        text-transform: uppercase;
        letter-spacing: 0.5px;
      }

      :deep(.el-form-item) {
        margin-bottom: 18px;

        &:last-child {
          margin-bottom: 0;
        }

        .el-form-item__label {
          font-weight: 500;
          color: $text-secondary;
          font-size: 12px;
          margin-bottom: 8px;
        }

        .el-input__wrapper,
        .el-textarea__inner {
          background: $bg-white;
          border: 1px solid $border-color;
          border-radius: 10px;
          box-shadow: none;

          &:hover {
            border-color: darken($border-color, 10%);
          }

          &:focus-within {
            border-color: $primary-color;
            box-shadow: 0 0 0 3px rgba($primary-color, 0.1);
          }
        }

        .el-select {
          width: 100%;

          :deep(.el-input__wrapper) {
            border-radius: 10px;
          }
        }

        .el-checkbox-group {
          display: flex;
          flex-direction: column;
          gap: 8px;

          .el-checkbox {
            margin-right: 0;
            padding: 10px 14px;
            background: $bg-card;
            border: 1px solid $border-color;
            border-radius: 10px;
            transition: all 0.2s ease;

            &:hover {
              background: rgba($primary-color, 0.05);
              border-color: rgba($primary-color, 0.3);
            }

            &.is-checked {
              background: rgba($primary-color, 0.08);
              border-color: rgba($primary-color, 0.4);
            }
          }
        }

        .el-radio-group {
          display: flex;
          flex-direction: column;
          gap: 10px;

          .el-radio {
            margin-right: 0;
            padding: 12px 16px;
            background: $bg-card;
            border: 1px solid $border-color;
            border-radius: 10px;
            transition: all 0.2s ease;
            height: auto !important;
            align-items: flex-start;

            :deep(.el-radio__label) {
              display: flex;
              flex-direction: column;
              width: 100%;
              padding-left: 0;
              white-space: normal;
              line-height: 1.5;
              color: $text-primary;
              font-weight: 500;
            }

            &:hover {
              background: rgba($primary-color, 0.05);
              border-color: rgba($primary-color, 0.3);
            }

            &.is-checked {
              background: rgba($primary-color, 0.08);
              border-color: rgba($primary-color, 0.5);
            }
          }
        }

        :deep(.el-slider) {
          .el-slider__runway {
            background: $border-color;
            border-radius: 4px;
          }
          
          .el-slider__bar {
            background: linear-gradient(90deg, $primary-color, $primary-light);
            border-radius: 4px;
          }
          
          .el-slider__button {
            border-color: $primary-color;
            background: $primary-color;
          }
        }
      }

      .form-tip {
        font-size: 11px;
        color: $text-muted;
        margin-top: 8px;
        line-height: 1.5;
        padding: 10px 12px;
        background: $bg-card;
        border-radius: 8px;
        border-left: 2px solid $text-muted;
      }

      .info-tip {
        font-size: 12px;
        color: $primary-color;
        background: rgba($primary-color, 0.05);
        padding: 12px 14px;
        border-radius: 10px;
        border: 1px solid rgba($primary-color, 0.1);
        line-height: 1.6;
      }

      .divider {
        height: 1px;
        background: $border-color;
        margin: 20px 0;
      }

      .sub-section-title {
        font-size: 11px;
        font-weight: 600;
        color: $text-muted;
        margin-bottom: 14px;
        text-transform: uppercase;
        letter-spacing: 1px;
      }
    }
  }

  .panel-footer {
    padding: 16px 20px;
    border-top: 1px solid $border-color;
    background: $bg-light;
    display: flex;
    justify-content: flex-end;

    .el-button {
      border-radius: 12px;
      padding: 12px 28px;
      font-weight: 600;
      background: linear-gradient(135deg, $primary-color, $primary-dark);
      border: none;
      color: white;
      
      &:hover {
        background: linear-gradient(135deg, $primary-light, $primary-color);
      }
    }
  }

  &.empty {
    justify-content: center;
    align-items: center;
    background: $bg-light;

    .empty-content {
      text-align: center;
      color: $text-muted;
      padding: 40px;

      .el-icon {
        color: $text-muted;
        opacity: 0.4;
        font-size: 56px;
        margin-bottom: 16px;
      }

      p {
        margin-top: 16px;
        font-size: 14px;
        font-weight: 500;
        color: $text-secondary;

        &.tip {
          font-size: 12px;
          color: $text-muted;
          margin-top: 10px;
        }
      }
    }
  }
}
</style>
