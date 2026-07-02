<template>
  <div class="condition-node">
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Share /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="condition-info" v-if="conditionConfig">
        <div class="condition-type">
          <el-tag size="small" :type="conditionTypeTag">{{ conditionTypeLabel }}</el-tag>
        </div>
        <div class="condition-detail" v-if="conditionDetail">
          <span class="detail-text">{{ conditionDetail }}</span>
        </div>
      </div>
      <div class="no-condition" v-else>点击配置条件</div>
    </div>

    <Handle type="source" :position="Position.Right" id="true" :style="{ top: '30%' }" />
    <div class="branch-label branch-true">是</div>
    <Handle type="source" :position="Position.Right" id="false" :style="{ top: '70%' }" />
    <div class="branch-label branch-false">否</div>
    <Handle type="source" :position="Position.Left" id="true-left" :style="{ top: '30%' }" />
    <Handle type="source" :position="Position.Left" id="false-left" :style="{ top: '70%' }" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Share } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '条件分支')
const conditionConfig = computed(() => props.data?.conditionConfig)

const conditionTypeLabel = computed(() => {
  const typeMap = { 'expression': '表达式', 'llm': 'AI判断', 'tool_result': '工具结果' }
  return typeMap[conditionConfig.value?.type] || '条件'
})

const conditionTypeTag = computed(() => {
  const tagMap = { 'expression': 'success', 'llm': 'primary', 'tool_result': 'warning' }
  return tagMap[conditionConfig.value?.type] || 'info'
})

const conditionDetail = computed(() => {
  const config = conditionConfig.value
  if (!config) return ''
  switch (config.type) {
    case 'expression': return config.expression ? (config.expression.length > 20 ? config.expression.substring(0, 20) + '...' : config.expression) : '未配置表达式'
    case 'llm': return config.llmPrompt ? (config.llmPrompt.length > 15 ? config.llmPrompt.substring(0, 15) + '...' : config.llmPrompt) : '未配置提示词'
    case 'tool_result': return config.toolName || '未选择工具'
    default: return ''
  }
})
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-muted: #94a3b8;
$success-color: #10b981;
$danger-color: #ef4444;

.condition-node {
  padding: 0;
  background: white;
  border: 2px solid #7c3aed;
  border-radius: 14px;
  min-width: 180px;
  box-shadow: 0 2px 12px rgba(168, 85, 247, 0.12);
  transition: all 0.2s ease;
  overflow: visible;
  position: relative;

  &:hover {
    box-shadow: 0 4px 16px rgba(168, 85, 247, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#7c3aed, 0.1) 0%, rgba(#7c3aed, 0.05) 100%);
    border-bottom: 1px solid rgba(#7c3aed, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #7c3aed;
      border-radius: 8px;
      color: white;
      font-size: 15px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: $text-primary;
    }
  }

  .node-content {
    padding: 12px 14px;

    .condition-info {
      display: flex;
      flex-direction: column;
      gap: 8px;

      .condition-detail .detail-text {
        font-size: 11px;
        color: $text-muted;
        background: #f8fafc;
        padding: 4px 8px;
        border-radius: 6px;
        display: inline-block;
        border: 1px solid #e2e8f0;
      }
    }

    .no-condition {
      font-size: 12px;
      color: $text-muted;
      text-align: center;
      padding: 6px 0;
    }
  }

  .branch-label {
    position: absolute;
    right: -28px;
    font-size: 11px;
    font-weight: 600;
    padding: 4px 8px;
    border-radius: 6px;
    z-index: 10;

    &.branch-true {
      top: 20%;
      background: rgba($success-color, 0.1);
      color: darken($success-color, 5%);
      border: 1px solid rgba($success-color, 0.2);
    }

    &.branch-false {
      top: 60%;
      background: rgba($danger-color, 0.1);
      color: $danger-color;
      border: 1px solid rgba($danger-color, 0.2);
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #7c3aed;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
