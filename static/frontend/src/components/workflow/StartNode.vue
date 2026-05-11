<template>
  <div class="start-node">
    <div class="node-icon">
      <el-icon><VideoPlay /></el-icon>
    </div>
    <div class="node-label">{{ label }}</div>
    <!-- 输出连接点 - 右侧 -->
    <Handle type="source" :position="Position.Right" id="output" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { VideoPlay } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '开始')
</script>

<style lang="scss" scoped>
.start-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%);
  border: 2px solid #4caf50;
  border-radius: 12px;
  min-width: 120px;
  box-shadow: 0 2px 8px rgba(76, 175, 80, 0.2);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 12px rgba(76, 175, 80, 0.35);
    transform: translateY(-1px);
  }

  .node-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: #4caf50;
    border-radius: 50%;
    color: white;
    font-size: 14px;
  }

  .node-label {
    font-size: 14px;
    font-weight: 500;
    color: #2e7d32;
  }
}

// Vue Flow handle 样式
:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #4caf50;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
