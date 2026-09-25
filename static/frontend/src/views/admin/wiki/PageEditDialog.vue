<template>
  <el-dialog
    :model-value="visible"
    :title="isDraft ? '编辑草稿' : '编辑页面'"
    width="760px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
  >
    <el-form :model="form" label-width="80px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="标题">
            <el-input v-model="form.title" maxlength="255" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="类型">
            <el-select v-model="form.pageType" style="width: 100%">
              <el-option label="summary 总览" value="summary" />
              <el-option label="entity 实体" value="entity" />
              <el-option label="concept 概念" value="concept" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="slug">
            <el-input v-model="form.slug" disabled />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="别名">
        <el-input
          v-model="aliasesText"
          placeholder="逗号分隔，用于提升搜索命中，如：Golang, Go 语言"
        />
      </el-form-item>
      <el-form-item label="摘要">
        <el-input
          v-model="form.summary"
          type="textarea"
          :rows="2"
          maxlength="512"
          show-word-limit
          placeholder="一句话摘要，会出现在知识库 index 中"
        />
      </el-form-item>
      <el-form-item label="正文">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="16"
          class="content-editor"
          placeholder="markdown 正文，用 [[slug|文字]] 引用其他页面"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button round @click="$emit('update:visible', false)">取消</el-button>
      <el-button round type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup name="PageEditDialog">
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import wikiApi from '@/api/wiki'

const props = defineProps({
  visible: Boolean,
  // 待编辑的草稿或已发布页记录
  page: { type: Object, default: null },
  isDraft: { type: Boolean, default: true }
})
const emit = defineEmits(['update:visible', 'saved'])

const saving = ref(false)
const form = ref({})
const aliasesText = ref('')

const isDraft = computed(() => props.isDraft)

watch(() => props.visible, (v) => {
  if (v && props.page) {
    form.value = {
      id: props.page.id,
      slug: props.page.slug,
      title: props.page.title,
      pageType: props.page.pageType || 'concept',
      summary: props.page.summary || '',
      content: props.page.content || ''
    }
    aliasesText.value = (props.page.aliases || []).join(', ')
  }
})

const save = async () => {
  if (!form.value.title?.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }
  saving.value = true
  try {
    const data = {
      title: form.value.title.trim(),
      pageType: form.value.pageType,
      summary: form.value.summary,
      content: form.value.content
    }
    const aliases = aliasesText.value.split(/[,，]/).map(s => s.trim()).filter(Boolean)
    if (aliases.length) data.aliases = aliases
    const response = isDraft.value
      ? await wikiApi.updateDraft(form.value.id, data)
      : await wikiApi.updatePublished(form.value.id, data)
    if (response.code === 0) {
      ElMessage.success('保存成功')
      emit('update:visible', false)
      emit('saved')
    } else {
      ElMessage.error(response.msg || '保存失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped lang="scss">
.content-editor {
  :deep(.el-textarea__inner) {
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 13px;
    line-height: 1.6;
  }
}
</style>
