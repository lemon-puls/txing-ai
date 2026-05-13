<template>
  <div class="condition-node">
    <!-- 输入连接点 - 左侧 -->
    <Handle type="target" :position="Position.Left" id="input-left" />
    <!-- 输入连接点 - 右侧 -->
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Share /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <!-- 条件类型和配置状态 -->
      <div class="condition-info" v-if="conditionConfig">
        <div class="condition-type">
          <el-tag size="small" :type="conditionTypeTag">{{ conditionTypeLabel }}</el-tag>
        </div>
        <div class="condition-detail" v-if="conditionDetail">
          <span class="detail-text">{{ conditionDetail }}</span>
        </div>
        <div class="failure-action" v-if="failureActionLabel">
          <el-tag size="small" type="warning" effect="plain">{{ failureActionLabel }}</el-tag>
        </div>
      </div>
      <div class="no-condition" v-else>
        点击配置条件
      </div>
    </div>

    <!-- True 输出连接点 - 右上 -->
    <Handle type="source" :position="Position.Right" id="true" :style="{ top: '30%' }" />
    <div class="branch-label branch-true">是</div>

    <!-- False 输出连接点 - 右下 -->
    <Handle type="source" :position="Position.Right" id="false" :style="{ top: '70%' }" />
    <div class="branch-label branch-false">否</div>

    <!-- True 输出连接点 - 左上 -->
    <Handle type="source" :position="Position.Left" id="true-left" :style="{ top: '30%' }" />

    <!-- False 输出连接点 - 左下 -->
    <Handle type="source" :position="Position.Left" id="false-left" :style="{ top: '70%' }" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Share } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '条件分支')
const conditionConfig = computed(() => props.data?.conditionConfig)

const conditionTypeLabel = computed(() => {
  const typeMap = {
    'expression': '表达式',
    'llm': 'AI判断',
    'tool_result': '工具结果'
  }
  return typeMap[conditionConfig.value?.type] || '条件'
})

const conditionTypeTag = computed(() => {
  const tagMap = {
    'expression': 'success',
    'llm': 'primary',
    'tool_result': 'warning'
  }
  return tagMap[conditionConfig.value?.type] || 'info'
})

const conditionDetail = computed(() => {
  const config = conditionConfig.value
  if (!config) return ''

  switch (config.type) {
    case 'expression':
      if (config.expression) {
        // 截断过长的表达式
        const expr = config.expression
        if (expr.length > 20) {
          return expr.substring(0, 20) + '...'
        }
        return expr
      }
      return '未配置表达式'
    case 'llm':
      if (config.llmPrompt) {
        const prompt = config.llmPrompt
        if (prompt.length > 15) {
          return prompt.substring(0, 15) + '...'
        }
        return prompt
      }
      return '未配置提示词'
    case 'tool_result':
      if (config.toolName) {
        return config.toolName
      }
      return '未选择工具'
    default:
      return ''
  }
})

const failureActionLabel = computed(() => {
  const action = conditionConfig.value?.failureAction
  if (!action || action === 'default_false') return ''

  const actionMap = {
    'terminate': '失败终止',
    'configurable': '自定义分支'
  }
  return actionMap[action] || ''
})
</script>

<style lang="scss" scoped>
.condition-node {
  padding: 0;
  background: white;
  border: 2px solid #9c27b0;
  border-radius: 12px;
  min-width: 180px;
  box-shadow: 0 2px 8px rgba(156, 39, 176, 0.15);
  transition: all 0.2s ease;
  overflow: visible;
  position: relative;

  &:hover {
    box-shadow: 0 4px 16px rgba(156, 39, 176, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%);
    border-bottom: 1px solid #ce93d8;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #9c27b0;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #6a1b9a;
    }
  }

  .node-content {
    padding: 10px 12px;

    .condition-info {
      display: flex;
      flex-direction: column;
      gap: 6px;

      .condition-type {
        display: flex;
        align-items: center;
      }

      .condition-detail {
        .detail-text {
          font-size: 11px;
          color: #757575;
          background: #f5f5f5;
          padding: 4px 8px;
          border-radius: 4px;
          display: inline-block;
          max-width: 100%;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      .failure-action {
        margin-top: 2px;
      }
    }

    .no-condition {
      font-size: 12px;
      color: #bdbdbd;
      text-align: center;
      padding: 4px 0;
    }
  }

  .branch-label {
    position: absolute;
    right: -28px;
    font-size: 12px;
    font-weight: 500;
    padding: 2px 6px;
    border-radius: 4px;

    &.branch-true {
      top: 20%;
      background: #e8f5e9;
      color: #2e7d32;
    }

    &.branch-false {
      top: 60%;
      background: #ffebee;
      color: #c62828;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #9c27b0;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
