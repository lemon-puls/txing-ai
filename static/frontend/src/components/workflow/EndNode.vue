<template>
  <div class="end-node">
    <!-- 输入连接点 - 左侧 -->
    <Handle type="target" :position="Position.Left" id="input-left" />
    <!-- 输入连接点 - 右侧 -->
    <Handle type="target" :position="Position.Right" id="input-right" />
    <div class="node-icon">
      <el-icon><CircleClose /></el-icon>
    </div>
    <div class="node-label">{{ label }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { CircleClose } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '结束')
</script>

<style lang="scss" scoped>
.end-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #ffebee 0%, #ffcdd2 100%);
  border: 2px solid #ef5350;
  border-radius: 12px;
  min-width: 120px;
  box-shadow: 0 2px 8px rgba(239, 83, 80, 0.2);
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 4px 12px rgba(239, 83, 80, 0.35);
    transform: translateY(-1px);
  }

  .node-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: #ef5350;
    border-radius: 50%;
    color: white;
    font-size: 14px;
  }

  .node-label {
    font-size: 14px;
    font-weight: 500;
    color: #c62828;
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #ef5350;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
