<template>
  <div class="tags-input">
    <el-tag
      v-for="tag in modelValue"
      :key="tag"
      closable
      round
      size="default"
      :disable-transitions="true"
      @close="removeTag(tag)"
    >
      {{ tag }}
    </el-tag>
    <el-input
      v-if="inputVisible"
      ref="inputRef"
      v-model="inputValue"
      class="tag-editor-input"
      size="small"
      :placeholder="placeholder"
      @keyup.enter="handleInputConfirm"
      @blur="handleInputConfirm"
    />
    <el-button v-else size="small" round link type="primary" @click="showInput">
      <el-icon class="btn-icon"><Plus /></el-icon>
      {{ addText }}
    </el-button>
  </div>
</template>

<script setup>
// 标签输入：tag 胶囊 + 回车/失焦确认，与 WebsiteList 标签交互一致
// Tag-style array editor: pills + enter/blur confirm, same UX as WebsiteList
import { ref, nextTick } from 'vue'
import { Plus } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  placeholder: { type: String, default: '输入后回车' },
  addText: { type: String, default: '添加标签' }
})
const emit = defineEmits(['update:modelValue'])

const inputVisible = ref(false)
const inputValue = ref('')
const inputRef = ref(null)

const showInput = () => {
  inputVisible.value = true
  nextTick(() => inputRef.value?.focus())
}

const handleInputConfirm = () => {
  const value = inputValue.value.trim()
  if (value && !props.modelValue.includes(value)) {
    emit('update:modelValue', [...props.modelValue, value])
  }
  inputVisible.value = false
  inputValue.value = ''
}

const removeTag = (tag) => {
  emit('update:modelValue', props.modelValue.filter((t) => t !== tag))
}
</script>

<style lang="scss" scoped>
.tags-input {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  width: 100%;

  .tag-editor-input {
    width: 120px;
  }

  .btn-icon {
    margin-right: 2px;
  }
}
</style>
