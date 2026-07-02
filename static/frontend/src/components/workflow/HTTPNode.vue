<template>
  <div class="http-node">
    <div class="node-header">
      <div class="node-icon"><el-icon><Link /></el-icon></div>
      <div class="node-title">{{ data.label || 'HTTP 节点' }}</div>
    </div>
    <div class="node-body">
      <div class="node-info" v-if="data.httpConfig">
        <span class="method-tag" :class="methodClass">{{ data.httpConfig.method || 'GET' }}</span>
        <span class="url-text">{{ truncatedUrl }}</span>
      </div>
      <div class="node-hint">发送 HTTP 请求</div>
    </div>
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="source" :position="Position.Right" id="output-right" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Link } from '@element-plus/icons-vue'

const props = defineProps({
  data: { type: Object, default: () => ({}) }
})

const methodClass = computed(() => `method-${props.data?.httpConfig?.method?.toLowerCase() || 'get'}`)
const truncatedUrl = computed(() => {
  const url = props.data?.httpConfig?.url || ''
  return url.length > 30 ? url.substring(0, 30) + '...' : url
})
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;
$success-color: #10b981;
$warning-color: #f59e0b;
$danger-color: #ef4444;
$info-color: #3b82f6;

.http-node {
  background: white;
  border: 2px solid #06b6d4;
  border-radius: 14px;
  padding: 0;
  min-width: 180px;
  box-shadow: 0 2px 12px rgba(6, 182, 212, 0.12);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 16px rgba(6, 182, 212, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#06b6d4, 0.1) 0%, rgba(#06b6d4, 0.05) 100%);
    border-bottom: 1px solid rgba(#06b6d4, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #06b6d4;
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
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;

      .method-tag {
        display: inline-block;
        padding: 4px 8px;
        border-radius: 6px;
        font-size: 11px;
        font-weight: 600;
        font-family: monospace;

        &.method-get { background: rgba($success-color, 0.1); color: $success-color; border: 1px solid rgba($success-color, 0.2); }
        &.method-post { background: rgba($warning-color, 0.1); color: $warning-color; border: 1px solid rgba($warning-color, 0.2); }
        &.method-put { background: rgba($info-color, 0.1); color: $info-color; border: 1px solid rgba($info-color, 0.2); }
        &.method-delete { background: rgba($danger-color, 0.1); color: $danger-color; border: 1px solid rgba($danger-color, 0.2); }
      }

      .url-text {
        font-size: 11px;
        color: $text-muted;
        font-family: monospace;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
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
  background: #06b6d4;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
