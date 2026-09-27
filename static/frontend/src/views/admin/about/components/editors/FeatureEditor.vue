<template>
  <div class="feature-editor">
    <div v-for="(f, idx) in modelValue" :key="idx" class="feature-group">
      <div class="feature-main">
        <el-input v-model="f.icon" placeholder="emoji" class="feature-icon" maxlength="4" />
        <el-input v-model="f.title" placeholder="功能标题" class="feature-title" />
        <el-input v-model="f.desc" placeholder="一句话描述" class="feature-desc" />
        <el-button text type="danger" :icon="Delete" @click="removeRow(idx)" />
      </div>
      <MarkdownInput
        v-model="f.detail"
        :rows="5"
        placeholder="详细内容（选填，支持 Markdown；```mermaid 代码块前台渲染为图表），前台点击该功能卡片展开显示"
      />
    </div>
    <el-button round size="small" :icon="Plus" @click="addRow">添加功能</el-button>
  </div>
</template>

<script setup>
// 核心功能编辑器：features[{icon, title, desc, detail(Markdown)}]
// Feature editor; detail is expanded on the front page with Markdown/Mermaid rendering
import { Plus, Delete } from '@element-plus/icons-vue'
import MarkdownInput from './MarkdownInput.vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const addRow = () => {
  emit('update:modelValue', [...props.modelValue, { icon: '', title: '', desc: '', detail: '' }])
}

const removeRow = (idx) => {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== idx))
}
</script>

<style lang="scss" scoped>
.feature-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;

  .feature-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    background: var(--el-fill-color-lighter);

    .feature-main {
      display: flex;
      align-items: center;
      gap: 8px;

      .feature-icon {
        width: 72px;
        flex-shrink: 0;
        text-align: center;
      }

      .feature-title {
        width: 200px;
        flex-shrink: 0;
      }

      .feature-desc {
        flex: 1;
      }
    }
  }
}
</style>
