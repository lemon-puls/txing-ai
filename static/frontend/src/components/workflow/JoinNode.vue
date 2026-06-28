<template>
  <div class="join-node">
    <!-- 输入连接点 - 顶部 -->
    <Handle type="target" :position="Position.Top" id="input" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Connection /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="strategy-badge" :class="strategy">
        <el-icon><component :is="strategyIcon" /></el-icon>
        <span>{{ strategyText }}</span>
      </div>
    </div>

    <!-- 输出连接点 - 底部 -->
    <Handle type="source" :position="Position.Bottom" id="output" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Connection, List, Check } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '汇聚')
const strategy = computed(() => props.data?.joinConfig?.strategy || 'all')

const strategyText = computed(() => {
  return strategy.value === 'any' ? '任一完成' : '全部完成'
})

const strategyIcon = computed(() => {
  return strategy.value === 'any' ? List : Check
})
</script>

<style lang="scss" scoped>
.join-node {
  padding: 0;
  background: white;
  border: 2px solid #00bfa5;
  border-radius: 12px;
  min-width: 140px;
  box-shadow: 0 2px 8px rgba(0, 191, 165, 0.15);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 191, 165, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #e0f2f1 0%, #b2dfdb 100%);
    border-bottom: 1px solid #80cbc4;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #00bfa5;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #00695c;
    }
  }

  .node-content {
    padding: 10px 12px;

    .strategy-badge {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      padding: 6px 10px;
      border-radius: 6px;

      &.all {
        background: rgba(0, 191, 165, 0.15);
        color: #00bfa5;
      }

      &.any {
        background: rgba(255, 152, 0, 0.15);
        color: #ff9800;
      }
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #00bfa5;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
