<template>
  <div class="agent-node">
    <!-- 输入连接点 - 左侧 -->
    <Handle type="target" :position="Position.Left" id="input-left" />
    <!-- 输入连接点 - 右侧 -->
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Avatar /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="info-row" v-if="agentConfig?.tools?.length">
        <span class="info-label">工具:</span>
        <span class="info-value">{{ agentConfig.tools.length }} 个</span>
      </div>
      <div class="info-row" v-if="agentConfig?.maxRunSteps">
        <span class="info-label">最大步数:</span>
        <span class="info-value">{{ agentConfig.maxRunSteps }}</span>
      </div>
      <div class="info-row" v-if="agentConfig?.systemPrompt">
        <span class="info-label">提示词:</span>
        <span class="info-value prompt-preview">{{ promptPreview }}</span>
      </div>
    </div>

    <!-- 输出连接点 - 右侧 -->
    <Handle type="source" :position="Position.Right" id="output-right" />
    <!-- 输出连接点 - 左侧 -->
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Avatar } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || 'Agent')
const agentConfig = computed(() => props.data?.agentConfig || {})
const promptPreview = computed(() => {
  const prompt = agentConfig.value?.systemPrompt || ''
  return prompt.length > 30 ? prompt.substring(0, 30) + '...' : prompt
})
</script>

<style lang="scss" scoped>
.agent-node {
  padding: 0;
  background: white;
  border: 2px solid #e91e63;
  border-radius: 12px;
  min-width: 200px;
  box-shadow: 0 2px 8px rgba(233, 30, 99, 0.15);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(233, 30, 99, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #fce4ec 0%, #f8bbd0 100%);
    border-bottom: 1px solid #f48fb1;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #e91e63;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #c2185b;
    }
  }

  .node-content {
    padding: 10px 12px;

    .info-row {
      display: flex;
      align-items: flex-start;
      margin-bottom: 6px;
      font-size: 12px;

      &:last-child {
        margin-bottom: 0;
      }

      .info-label {
        flex-shrink: 0;
        width: 50px;
        color: #90a4ae;
      }

      .info-value {
        flex: 1;
        color: #455a64;
        word-break: break-all;

        &.prompt-preview {
          color: #78909c;
          font-style: italic;
        }
      }
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #e91e63;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
