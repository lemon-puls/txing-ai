<template>
  <div class="cover-media-editor">
    <div ref="listRef" class="cover-list">
      <div v-for="(m, idx) in modelValue" :key="uidOf(m)" :data-sort-id="uidOf(m)" class="cover-row">
        <span class="drag-handle" title="拖拽调整轮播顺序">
          <el-icon><Rank /></el-icon>
        </span>
        <MediaUploader
          v-model="m.url"
          media-type="all"
          :locked-type="m.type || undefined"
          :width="200"
          :height="100"
          placeholder="上传封面（图片/视频）"
          :max-size="50"
          @change="(info) => onUploaded(m, info)"
        />
        <span class="cover-order">{{ idx + 1 }}</span>
        <el-button text type="danger" :icon="Delete" @click="removeRow(idx)" />
      </div>
    </div>
    <el-button round size="small" :icon="Plus" @click="addRow">添加封面</el-button>
    <p class="form-tip">按顺序在项目卡左侧大图区轮播；未配置时前台回退为渐变+图标。拖拽手柄可调整顺序</p>
  </div>
</template>

<script setup>
// 封面轮播编辑器：coverMedia[{type, key, url}]，拖拽调序（本地重排，随表单整体保存）
// Cover carousel editor; drag reorders locally and persists with the whole form save
import { ref, computed } from 'vue'
import { Plus, Delete, Rank } from '@element-plus/icons-vue'
import MediaUploader from '@/components/common/MediaUploader.vue'
import { useDragSort, uidOf } from '@/composables/useDragSort'

const props = defineProps({
  modelValue: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const listRef = ref(null)
// get/set 计算属性：useDragSort 整体赋值时自动 emit，避免双 watcher 死循环
// computed get/set: wholesale assignment inside useDragSort emits once, no watcher loops
const items = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

// 表单内排序：仅本地重排、保存时才落库（无 persist；renumber 关闭，coverMedia 无 sort 字段）
// in-form sorting: local only, persisted on save (no persist; coverMedia has no sort field)
useDragSort(listRef, { items, getId: uidOf, renumber: false })

const addRow = () => {
  emit('update:modelValue', [...props.modelValue, { type: '', key: '', url: '' }])
}

const removeRow = (idx) => {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== idx))
}

// 上传成功回填 key（落库值）与 url（仅预览）
// fill in key (persisted) and url (preview only) after upload
const onUploaded = (m, info) => {
  m.type = info.type
  m.key = info.key || ''
  m.url = info.url
}
</script>

<style lang="scss" scoped>
.cover-media-editor {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  width: 100%;

  .cover-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    width: 100%;
  }

  .cover-row {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    background: var(--el-fill-color-lighter);

    .drag-handle {
      cursor: grab;
      color: var(--el-text-color-placeholder);
      display: flex;
      align-items: center;

      &:hover {
        color: var(--el-color-primary);
      }

      &:active {
        cursor: grabbing;
      }
    }

    .cover-order {
      min-width: 20px;
      height: 20px;
      padding: 0 6px;
      border-radius: 999px;
      background: var(--el-color-primary-light-9);
      color: var(--el-color-primary);
      font-size: 12px;
      line-height: 20px;
      text-align: center;
    }
  }

  .form-tip {
    margin: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

// 拖拽占位样式：全局复用（useDragSort 的 ghost/chosen class）
// shared drag ghost styles used by useDragSort
:deep(.drag-ghost) {
  opacity: 0.4;
  outline: 2px dashed var(--el-color-primary);
}
</style>
