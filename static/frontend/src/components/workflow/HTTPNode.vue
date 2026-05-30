<template>
  <div class="http-node">
    <div class="node-header">
      <div class="node-icon">
        <el-icon><Link /></el-icon>
      </div>
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
  data: {
    type: Object,
    default: () => ({})
  }
})

const methodClass = computed(() => {
  const method = props.data?.httpConfig?.method?.toLowerCase() || 'get'
  return `method-${method}`
})

const truncatedUrl = computed(() => {
  const url = props.data?.httpConfig?.url || ''
  if (url.length > 30) {
    return url.substring(0, 30) + '...'
  }
  return url
})
</script>

<style lang="scss" scoped>
.http-node {
  background: white;
  border: 2px solid #00bcd4;
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
    background: #e0f7fa;
    border-radius: 10px 10px 0 0;
    border-bottom: 1px solid #b2ebf2;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #00bcd4;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #006064;
    }
  }

  .node-body {
    padding: 12px 16px;

    .node-info {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;

      .method-tag {
        display: inline-block;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 600;
        font-family: monospace;

        &.method-get {
          background: #e8f5e9;
          color: #2e7d32;
        }
        &.method-post {
          background: #fff3e0;
          color: #ef6c00;
        }
        &.method-put {
          background: #e3f2fd;
          color: #1565c0;
        }
        &.method-delete {
          background: #ffebee;
          color: #c62828;
        }
      }

      .url-text {
        font-size: 11px;
        color: #757575;
        font-family: monospace;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .node-hint {
      font-size: 12px;
      color: #9e9e9e;
    }
  }
}
</style>
