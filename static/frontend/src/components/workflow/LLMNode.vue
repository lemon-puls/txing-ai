<template>
  <div class="llm-node">
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><ChatDotRound /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="info-row" v-if="modelConfig?.model">
        <span class="info-label">模型:</span>
        <span class="info-value">{{ modelConfig.model }}</span>
      </div>
      <div class="info-row" v-if="modelConfig?.temperature !== undefined">
        <span class="info-label">温度:</span>
        <span class="info-value">{{ modelConfig.temperature }}</span>
      </div>
      <div class="info-row" v-if="modelConfig?.systemPrompt">
        <span class="info-label">提示词:</span>
        <span class="info-value prompt-preview">{{ promptPreview }}</span>
      </div>
    </div>

    <Handle type="source" :position="Position.Right" id="output-right" />
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { ChatDotRound } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '大模型')
const modelConfig = computed(() => props.data?.modelConfig || {})
const promptPreview = computed(() => {
  const prompt = modelConfig.value?.systemPrompt || ''
  return prompt.length > 30 ? prompt.substring(0, 30) + '...' : prompt
})
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.llm-node {
  padding: 0;
  background: white;
  border: 2px solid #3b82f6;
  border-radius: 14px;
  min-width: 220px;
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.12);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(59, 130, 246, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#3b82f6, 0.1) 0%, rgba(#3b82f6, 0.05) 100%);
    border-bottom: 1px solid rgba(#3b82f6, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #3b82f6;
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

    .info-row {
      display: flex;
      align-items: flex-start;
      margin-bottom: 8px;
      font-size: 12px;

      &:last-child {
        margin-bottom: 0;
      }

      .info-label {
        flex-shrink: 0;
        width: 48px;
        color: $text-muted;
      }

      .info-value {
        flex: 1;
        color: $text-secondary;
        word-break: break-all;

        &.prompt-preview {
          color: $text-muted;
          font-style: italic;
        }
      }
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #3b82f6;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
