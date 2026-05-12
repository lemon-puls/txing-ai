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
        </el-form>
      </div>

      <!-- 工具节点配置 -->
      <div class="config-section" v-if="nodeType === 'tool'">
        <div class="section-title">工具配置</div>
        <el-form label-position="top" size="small">
          <el-form-item label="选择工具">
            <el-checkbox-group v-model="localData.toolConfig.tools">
              <el-checkbox v-for="tool in toolList" :key="tool.name" :label="tool.name">
                {{ tool.displayName }}
              </el-checkbox>
            </el-checkbox-group>
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

          <!-- 表达式条件 -->
          <template v-if="localData.conditionConfig.type === 'expression'">
            <el-form-item label="条件表达式">
              <el-input
                v-model="localData.conditionConfig.expression"
                placeholder="例如: {{output}} contains '成功'"
              />
              <div class="form-tip">
                支持运算符：<br/>
                • contains - 包含子串<br/>
                • equals - 精确匹配<br/>
                • starts_with - 前缀匹配<br/>
                • ends_with - 后缀匹配<br/>
                • matches - 正则匹配<br/>
                • greater_than - 数值大于<br/>
                • less_than - 数值小于<br/>
                • contains_any - 包含任意一个（逗号分隔或JSON数组）
              </div>
            </el-form-item>
          </template>

          <!-- AI判断 -->
          <template v-if="localData.conditionConfig.type === 'llm'">
            <el-form-item label="判断提示词">
              <el-input
                v-model="localData.conditionConfig.llmPrompt"
                type="textarea"
                :rows="3"
                placeholder="请输入让AI判断的提示词，例如：请判断以下内容是否包含积极的情绪"
              />
              <div class="form-tip">AI 将返回结构化 JSON 结果 {result: true/false, reason: '判断原因'}</div>
            </el-form-item>
          </template>

          <!-- 工具结果 -->
          <template v-if="localData.conditionConfig.type === 'tool_result'">
            <el-form-item label="工具名称">
              <el-select v-model="localData.conditionConfig.toolName" placeholder="请选择工具">
                <el-option v-for="tool in toolList" :key="tool.name" :label="tool.displayName" :value="tool.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="结果字段">
              <el-input v-model="localData.conditionConfig.toolResultKey" placeholder="例如: status, code" />
            </el-form-item>
            <el-form-item label="期望值">
              <el-input v-model="localData.conditionConfig.expectedValue" placeholder="与期望值比较，例如: success, 200" />
              <div class="form-tip">如果结果等于期望值，条件为 true</div>
            </el-form-item>
          </template>

          <!-- 错误处理策略 -->
          <div class="divider"></div>
          <div class="sub-section-title">错误处理</div>
          <el-form-item label="判断失败时的处理">
            <el-radio-group v-model="localData.conditionConfig.failureAction">
              <el-radio label="default_false">
                <span>默认走 false 分支</span>
                <div class="radio-desc">条件判断出错时自动走 false 分支继续执行</div>
              </el-radio>
              <el-radio label="terminate">
                <span>终止工作流</span>
                <div class="radio-desc">条件判断出错时停止整个工作流，返回错误</div>
              </el-radio>
              <el-radio label="configurable">
                <span>自定义默认分支</span>
                <div class="radio-desc">条件判断出错时走指定的分支</div>
              </el-radio>
            </el-radio-group>
          </el-form-item>

          <!-- 自定义错误分支 -->
          <template v-if="localData.conditionConfig.failureAction === 'configurable'">
            <el-form-item label="错误时的默认分支">
              <el-radio-group v-model="localData.conditionConfig.failureBranch">
                <el-radio label="true">走 true 分支</el-radio>
                <el-radio label="false">走 false 分支</el-radio>
              </el-radio-group>
            </el-form-item>
          </template>
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

// 本地数据副本
const localData = ref({
  label: '',
  description: '',
  modelConfig: {
    model: '',
    systemPrompt: '',
    temperature: 0.7,
    maxTokens: 4096,
    contextEnabled: true
  },
  toolConfig: {
    tools: [],
    params: {}
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
  }
})

// 监听选中节点变化，更新本地数据
watch(() => props.selectedNode, (newNode) => {
  if (newNode) {
    localData.value = {
      label: newNode.data?.label || '',
      description: newNode.data?.description || '',
      modelConfig: newNode.data?.modelConfig || {
        model: '',
        systemPrompt: '',
        temperature: 0.7,
        maxTokens: 4096,
        contextEnabled: true
      },
      toolConfig: newNode.data?.toolConfig || {
        tools: [],
        params: {}
      },
      conditionConfig: newNode.data?.conditionConfig || {
        type: 'expression',
        expression: '',
        llmPrompt: '',
        toolName: '',
        toolResultKey: '',
        expectedValue: '',
        failureAction: 'default_false',
        failureBranch: 'false'
      }
    }
  }
}, { immediate: true })

const handleApply = () => {
  emit('update', {
    id: props.selectedNode.id,
    data: { ...localData.value, nodeType: nodeType.value }
  })
}
</script>

<style lang="scss" scoped>
.property-panel {
  width: 340px;
  height: 100%;
  background: white;
  border-left: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.03);

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #e8e8e8;
    background: #fafbfc;

    .header-title {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 15px;
      font-weight: 600;
      color: #1a1a1a;

      .el-icon {
        color: #1976d2;
      }
    }
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;

    .config-section {
      margin-bottom: 24px;

      .section-title {
        font-size: 13px;
        font-weight: 600;
        color: #1976d2;
        margin-bottom: 14px;
        padding-left: 10px;
        border-left: 3px solid #1976d2;
        display: flex;
        align-items: center;
      }

      :deep(.el-form-item) {
        margin-bottom: 16px;

        .el-form-item__label {
          font-weight: 500;
          color: #424242;
        }

        .el-input__wrapper,
        .el-textarea__inner {
          border-radius: 10px;
          border-color: #e0e0e0;

          &:hover, &:focus-within {
            border-color: #1976d2;
          }
        }

        .el-select {
          width: 100%;

          .el-input__wrapper {
            border-radius: 10px;
          }
        }

        .el-checkbox-group {
          display: flex;
          flex-direction: column;
          gap: 10px;

          .el-checkbox {
            margin-right: 0;
            padding: 8px 12px;
            background: #f5f5f5;
            border-radius: 8px;
            transition: all 0.2s;

            &:hover {
              background: #e3f2fd;
            }

            &.is-checked {
              background: #e3f2fd;
            }
          }
        }

        .el-radio-group {
          display: flex;
          flex-direction: column;
          gap: 10px;

          .el-radio {
            margin-right: 0;
            padding: 10px 14px;
            background: #f5f5f5;
            border-radius: 10px;
            transition: all 0.2s;

            &:hover {
              background: #f0f0f0;
            }

            &.is-checked {
              background: #e3f2fd;
              border: 1px solid #1976d2;
            }
          }
        }
      }

      .form-tip {
        font-size: 11px;
        color: #9e9e9e;
        margin-top: 6px;
        line-height: 1.4;
      }

      .divider {
        height: 1px;
        background: #e0e0e0;
        margin: 16px 0;
      }

      .sub-section-title {
        font-size: 12px;
        font-weight: 600;
        color: #757575;
        margin-bottom: 12px;
      }

      .radio-desc {
        font-size: 11px;
        color: #9e9e9e;
        margin-top: 2px;
      }
    }
  }

  .panel-footer {
    padding: 14px 20px;
    border-top: 1px solid #e8e8e8;
    background: #fafbfc;
    display: flex;
    justify-content: flex-end;

    .el-button {
      border-radius: 10px;
      padding: 10px 24px;
      font-weight: 500;
    }
  }

  &.empty {
    justify-content: center;
    align-items: center;
    background: #fafbfc;

    .empty-content {
      text-align: center;
      color: #9e9e9e;

      .el-icon {
        color: #bdbdbd;
      }

      p {
        margin-top: 16px;
        font-size: 14px;

        &.tip {
          font-size: 12px;
          color: #bdbdbd;
          margin-top: 8px;
        }
      }
    }
  }
}
</style>
