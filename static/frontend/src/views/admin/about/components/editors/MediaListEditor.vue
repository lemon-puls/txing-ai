<template>
  <div class="media-list-editor">
    <div v-for="(m, idx) in modelValue" :key="idx" class="media-row">
      <el-select
        :model-value="m.type"
        class="type-select"
        @update:model-value="(v) => changeType(m, v)"
      >
        <el-option label="图片" value="image" />
        <el-option label="视频" value="video" />
      </el-select>
      <MediaUploader
        v-model="m.url"
        :media-type="m.type"
        :locked-type="m.type"
        :width="220"
        :height="110"
        :placeholder="`点击上传${m.type === 'video' ? '视频' : '图片'}`"
        :max-size="50"
        @change="(info) => onUploaded(m, info)"
      />
      <el-input v-model="m.caption" placeholder="说明（选填）" class="caption-input" />
      <el-button text type="danger" :icon="Delete" @click="removeRow(idx)" />
    </div>
    <el-button round size="small" :icon="Plus" @click="addRow">添加媒体</el-button>
    <p class="form-tip">支持图片/视频（50MB 内），在项目详情展开区展示；落库存 key，访问地址由后端读时签名</p>
  </div>
</template>

<script setup>
// 详情媒体列表编辑器：media[{type, key, url, caption}]
// Detail media list editor; only {type, key, caption} are persisted (key stored, URL signed on read)
import { Plus, Delete } from '@element-plus/icons-vue'
import MediaUploader from '@/components/common/MediaUploader.vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const addRow = () => {
  emit('update:modelValue', [...props.modelValue, { type: 'image', key: '', url: '', caption: '' }])
}

const removeRow = (idx) => {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== idx))
}

// 切换类型时清空旧资源，避免类型与资源不匹配
// clear the old asset when switching type
const changeType = (m, type) => {
  m.type = type
  m.url = ''
  m.key = ''
}

// 上传成功回填 key（落库值）与 url（仅预览）
// fill in key (persisted) and url (preview only) after upload
const onUploaded = (m, info) => {
  m.type = info.type
  m.key = info.key || ''
  m.url = info.url
  if (info.caption) m.caption = info.caption
}
</script>

<style lang="scss" scoped>
.media-list-editor {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  width: 100%;

  .media-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    width: 100%;
    padding: 8px 10px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;

    .type-select {
      width: 90px;
      flex-shrink: 0;
    }

    .caption-input {
      width: 180px;
    }
  }

  .form-tip {
    margin: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}
</style>
