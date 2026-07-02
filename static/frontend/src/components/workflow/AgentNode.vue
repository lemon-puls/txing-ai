<template>
  <div class="agent-node">
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Avatar /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="info-row" v-if="modelConfig?.model">
        <span class="info-label">模型:</span>
        <span class="info-value">{{ modelConfig.model }}</span>
      </div>
      <div class="info-row" v-else>
        <span class="info-label">模型:</span>
        <span class="info-value model-default">未指定</span>
      </div>
      <div class="info-row" v-if="agentConfig?.tools?.length">
        <span class="info-label">工具:</span>
        <span class="info-value">{{ agentConfig.tools.length }} 个</span>
      </div>
    </div>

    <Handle type="source" :position="Position.Right" id="output-right" />
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Avatar } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || 'Agent')
const agentConfig = computed(() => props.data?.agentConfig || {})
const modelConfig = computed(() => props.data?.modelConfig || {})
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.agent-node {
  padding: 0;
  background: white;
  border: 2px solid #ec4899;
  border-radius: 14px;
  min-width: 200px;
  box-shadow: 0 2px 12px rgba(236, 72, 153, 0.12);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(236, 72, 153, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#ec4899, 0.1) 0%, rgba(#ec4899, 0.05) 100%);
    border-bottom: 1px solid rgba(#ec4899, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #ec4899;
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

      &:last-child { margin-bottom: 0; }

      .info-label {
        flex-shrink: 0;
        width: 48px;
        color: $text-muted;
      }

      .info-value {
        flex: 1;
        color: $text-secondary;
        word-break: break-all;

        &.model-default {
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
  background: #ec4899;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
