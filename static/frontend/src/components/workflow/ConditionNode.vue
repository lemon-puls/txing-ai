<template>
  <div class="condition-node">
    <!-- 输入连接点 - 左侧 -->
    <Handle type="target" :position="Position.Left" id="input" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Share /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="condition-info" v-if="conditionConfig">
        <el-tag size="small" type="info">{{ conditionTypeLabel }}</el-tag>
      </div>
      <div class="no-condition" v-else>
        点击配置条件
      </div>
    </div>

    <!-- True 输出连接点 - 右上 -->
    <Handle type="source" :position="Position.Right" id="true" :style="{ top: '30%' }" />
    <div class="branch-label branch-true">是</div>

    <!-- False 输出连接点 - 右下 -->
    <Handle type="source" :position="Position.Right" id="false" :style="{ top: '70%' }" />
    <div class="branch-label branch-false">否</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Share } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
})

const label = computed(() => props.data?.label || '条件分支')
const conditionConfig = computed(() => props.data?.conditionConfig)

const conditionTypeLabel = computed(() => {
  const typeMap = {
    'expression': '表达式判断',
    'llm': 'AI判断',
    'tool_result': '工具结果'
  }
  return typeMap[conditionConfig.value?.type] || '条件判断'
})
</script>

<style lang="scss" scoped>
.condition-node {
  padding: 0;
  background: white;
  border: 2px solid #9c27b0;
  border-radius: 12px;
  min-width: 160px;
  box-shadow: 0 2px 8px rgba(156, 39, 176, 0.15);
  transition: all 0.2s ease;
  overflow: visible;
  position: relative;

  &:hover {
    box-shadow: 0 4px 16px rgba(156, 39, 176, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%);
    border-bottom: 1px solid #ce93d8;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #9c27b0;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #6a1b9a;
    }
  }

  .node-content {
    padding: 10px 12px;

    .condition-info {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    .no-condition {
      font-size: 12px;
      color: #bdbdbd;
      text-align: center;
      padding: 4px 0;
    }
  }

  .branch-label {
    position: absolute;
    right: -28px;
    font-size: 12px;
    font-weight: 500;
    padding: 2px 6px;
    border-radius: 4px;

    &.branch-true {
      top: 20%;
      background: #e8f5e9;
      color: #2e7d32;
    }

    &.branch-false {
      top: 60%;
      background: #ffebee;
      color: #c62828;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #9c27b0;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
