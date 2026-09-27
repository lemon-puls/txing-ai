<template>
  <div class="stats-editor">
    <div v-for="(stat, idx) in modelValue" :key="idx" class="stat-row">
      <el-input v-model="stat.value" placeholder="数值，如 10+" class="stat-value" maxlength="10" />
      <el-input v-model="stat.label" placeholder="统计项名称，如 年经验" class="stat-label" />
      <el-button text type="danger" :icon="Delete" @click="removeRow(idx)" />
    </div>
    <el-button round size="small" :icon="Plus" @click="addRow">添加统计</el-button>
  </div>
</template>

<script setup>
// 统计数据行编辑器：{value, label} 列表
// Stats row editor for {value, label} entries
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const addRow = () => {
  emit('update:modelValue', [...props.modelValue, { value: '', label: '' }])
}

const removeRow = (idx) => {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== idx))
}
</script>

<style lang="scss" scoped>
.stats-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;

  .stat-row {
    display: flex;
    align-items: center;
    gap: 8px;

    .stat-value {
      width: 130px;
      flex-shrink: 0;
    }

    .stat-label {
      flex: 1;
    }
  }
}
</style>
