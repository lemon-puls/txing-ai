<template>
  <div class="parallel-node">
    <!-- 输入连接点 - 顶部 -->
    <Handle type="target" :position="Position.Top" id="input" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Grid /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="branch-count">
        <el-icon><Tickets /></el-icon>
        <span>{{ branchCount }} 个分支</span>
      </div>
    </div>

    <!-- 输出连接点 - 底部 -->
    <Handle type="source" :position="Position.Bottom" id="output" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Grid, Tickets } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '并行组')
const branchCount = computed(() => {
  // 简单计算：如果有配置则使用，否则默认显示 2
  return props.data?.parallelConfig?.branchCount || 2
})
</script>

<style lang="scss" scoped>
.parallel-node {
  padding: 0;
  background: white;
  border: 2px solid #7c4dff;
  border-radius: 12px;
  min-width: 160px;
  box-shadow: 0 2px 8px rgba(124, 77, 255, 0.15);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(124, 77, 255, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #ede7f6 0%, #d1c4e9 100%);
    border-bottom: 1px solid #b39ddb;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #7c4dff;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #4527a0;
    }
  }

  .node-content {
    padding: 10px 12px;

    .branch-count {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      color: #7c4dff;
      background: rgba(124, 77, 255, 0.08);
      padding: 6px 10px;
      border-radius: 6px;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #7c4dff;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
