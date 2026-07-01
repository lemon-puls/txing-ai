<template>
  <div class="parallel-node">
    <Handle type="target" :position="Position.Top" id="input" />

    <div class="node-header">
      <div class="node-icon"><el-icon><Grid /></el-icon></div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="branch-count">
        <el-icon><Tickets /></el-icon>
        <span>{{ branchCount }} 个分支</span>
      </div>
    </div>

    <Handle type="source" :position="Position.Bottom" id="output" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Grid, Tickets } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '并行组')
const branchCount = computed(() => props.data?.parallelConfig?.branchCount || 2)
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;

.parallel-node {
  padding: 0;
  background: white;
  border: 2px solid #8b5cf6;
  border-radius: 14px;
  min-width: 160px;
  box-shadow: 0 2px 12px rgba(139, 92, 246, 0.12);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(139, 92, 246, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#8b5cf6, 0.1) 0%, rgba(#8b5cf6, 0.05) 100%);
    border-bottom: 1px solid rgba(#8b5cf6, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #8b5cf6;
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

    .branch-count {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 12px;
      color: #8b5cf6;
      background: rgba(#8b5cf6, 0.08);
      padding: 8px 12px;
      border-radius: 8px;
      border: 1px solid rgba(#8b5cf6, 0.15);
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #8b5cf6;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
