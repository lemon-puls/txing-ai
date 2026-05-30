<template>
  <div class="code-node">
    <div class="node-header">
      <div class="node-icon">
        <el-icon><Monitor /></el-icon>
      </div>
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
  data: {
    type: Object,
    default: () => ({})
  }
})
</script>

<style lang="scss" scoped>
.code-node {
  background: white;
  border: 2px solid #607d8b;
  border-radius: 12px;
  padding: 0;
  min-width: 180px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: #eceff1;
    border-radius: 10px 10px 0 0;
    border-bottom: 1px solid #cfd8dc;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #607d8b;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #37474f;
    }
  }

  .node-body {
    padding: 12px 16px;

    .node-info {
      margin-bottom: 8px;

      .language-tag {
        display: inline-block;
        padding: 2px 8px;
        background: #e0e0e0;
        border-radius: 4px;
        font-size: 11px;
        color: #616161;
        font-family: monospace;
      }
    }

    .node-hint {
      font-size: 12px;
      color: #9e9e9e;
    }
  }
}
</style>
