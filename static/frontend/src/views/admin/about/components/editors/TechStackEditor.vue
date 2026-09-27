<template>
  <div class="tech-stack-editor">
    <div v-for="(t, idx) in modelValue" :key="idx" class="tech-row">
      <el-input v-model="t.icon" placeholder="emoji" class="tech-icon" maxlength="4" />
      <el-input v-model="t.name" placeholder="技术名，如 Go" class="tech-name" />
      <el-button text type="danger" :icon="Delete" @click="removeRow(idx)" />
    </div>
    <el-button round size="small" :icon="Plus" @click="addRow">添加技术栈</el-button>
  </div>
</template>

<script setup>
// 技术栈行编辑器：{icon(emoji), name}
// Tech stack row editor: {icon(emoji), name}
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const addRow = () => {
  emit('update:modelValue', [...props.modelValue, { icon: '', name: '' }])
}

const removeRow = (idx) => {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== idx))
}
</script>

<style lang="scss" scoped>
.tech-stack-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;

  .tech-row {
    display: flex;
    align-items: center;
    gap: 8px;

    .tech-icon {
      width: 72px;
      flex-shrink: 0;
      text-align: center;
    }

    .tech-name {
      width: 220px;
    }
  }
}
</style>
