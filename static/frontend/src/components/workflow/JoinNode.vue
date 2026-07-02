<template>
  <div class="join-node">
    <Handle type="target" :position="Position.Top" id="input" />

    <div class="node-header">
      <div class="node-icon"><el-icon><Connection /></el-icon></div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="strategy-badge" :class="strategy">
        <el-icon><component :is="strategyIcon" /></el-icon>
        <span>{{ strategyText }}</span>
      </div>
    </div>

    <Handle type="source" :position="Position.Bottom" id="output" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Connection, List, Check } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '汇聚')
const strategy = computed(() => props.data?.joinConfig?.strategy || 'all')
const strategyText = computed(() => strategy.value === 'any' ? '任一完成' : '全部完成')
const strategyIcon = computed(() => strategy.value === 'any' ? List : Check)
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$success-color: #10b981;
$warning-color: #f59e0b;

.join-node {
  padding: 0;
  background: white;
  border: 2px solid #14b8a6;
  border-radius: 14px;
  min-width: 140px;
  box-shadow: 0 2px 12px rgba(20, 184, 166, 0.12);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(20, 184, 166, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#14b8a6, 0.1) 0%, rgba(#14b8a6, 0.05) 100%);
    border-bottom: 1px solid rgba(#14b8a6, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #14b8a6;
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

    .strategy-badge {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 12px;
      padding: 8px 12px;
      border-radius: 8px;
      border: 1px solid;

      &.all {
        background: rgba($success-color, 0.08);
        color: #0d9488;
        border-color: rgba($success-color, 0.2);
      }

      &.any {
        background: rgba($warning-color, 0.08);
        color: darken($warning-color, 5%);
        border-color: rgba($warning-color, 0.2);
      }
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #14b8a6;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
