<template>
  <div class="start-node">
    <div class="node-icon">
      <el-icon><VideoPlay /></el-icon>
    </div>
    <div class="node-label">{{ label }}</div>
    <Handle type="source" :position="Position.Right" id="output-right" />
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { VideoPlay } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '开始')
</script>

<style lang="scss" scoped>
$success-color: #10b981;

.start-node {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: linear-gradient(135deg, rgba($success-color, 0.1) 0%, rgba($success-color, 0.05) 100%);
  border: 2px solid $success-color;
  border-radius: 14px;
  min-width: 130px;
  box-shadow: 0 2px 12px rgba($success-color, 0.15);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 16px rgba($success-color, 0.25);
    transform: translateY(-1px);
  }

  .node-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: $success-color;
    border-radius: 50%;
    color: white;
    font-size: 15px;
  }

  .node-label {
    font-size: 14px;
    font-weight: 600;
    color: darken($success-color, 10%);
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: $success-color;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
