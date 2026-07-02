<template>
  <div class="code-node">
    <div class="node-header">
      <div class="node-icon"><el-icon><Monitor /></el-icon></div>
      <div class="node-title">{{ data.label || '代码节点' }}</div>
    </div>
    <div class="node-body">
      <div class="node-info" v-if="data.codeConfig">
        <span class="language-tag">{{ data.codeConfig.language || 'javascript' }}</span>
      </div>
      <div class="node-hint">执行自定义代码</div>
    </div>
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="source" :position="Position.Right" id="output-right" />
  </div>
</template>

<script setup>
import { Handle, Position } from '@vue-flow/core'
import { Monitor } from '@element-plus/icons-vue'

const props = defineProps({
  data: { type: Object, default: () => ({}) }
})
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.code-node {
  background: white;
  border: 2px solid #64748b;
  border-radius: 14px;
  padding: 0;
  min-width: 180px;
  box-shadow: 0 2px 12px rgba(100, 116, 139, 0.12);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 16px rgba(100, 116, 139, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#64748b, 0.1) 0%, rgba(#64748b, 0.05) 100%);
    border-bottom: 1px solid rgba(#64748b, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #64748b;
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

  .node-body {
    padding: 12px 14px;

    .node-info {
      margin-bottom: 8px;

      .language-tag {
        display: inline-block;
        padding: 4px 10px;
        background: rgba(#64748b, 0.1);
        border: 1px solid rgba(#64748b, 0.2);
        border-radius: 6px;
        font-size: 11px;
        color: $text-secondary;
        font-family: monospace;
        text-transform: uppercase;
      }
    }

    .node-hint {
      font-size: 12px;
      color: $text-muted;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #64748b;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
