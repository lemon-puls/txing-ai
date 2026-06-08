<template>
  <div class="chat-file-upload" v-if="files.length > 0">
    <div class="file-list">
      <div v-for="(file, index) in files" :key="index" class="file-item">
        <!-- 图片预览 -->
        <div v-if="file.type.startsWith('image/')" class="image-preview" @click="previewImage(file)">
          <img :src="file.preview" :alt="file.name" />
          <div class="file-overlay">
            <el-icon class="preview-icon"><ZoomIn /></el-icon>
          </div>
          <el-icon class="remove-btn" @click.stop="removeFile(index)">
            <Close />
          </el-icon>
          <div class="file-name">{{ file.name }}</div>
        </div>
        <!-- 文件预览 -->
        <div v-else class="doc-preview">
          <div class="file-icon-wrapper" :class="getFileClass(file.type)">
            <el-icon class="file-icon"><Document /></el-icon>
          </div>
          <div class="file-info">
            <div class="file-name" :title="file.name">{{ file.name }}</div>
            <div class="file-size">{{ formatFileSize(file.size) }}</div>
          </div>
          <el-icon class="remove-btn" @click="removeFile(index)">
            <Close />
          </el-icon>
        </div>
        <!-- 上传进度 -->
        <div v-if="file.uploading" class="upload-progress">
          <el-progress :percentage="file.progress || 0" :show-text="false" :stroke-width="3" />
        </div>
      </div>
    </div>

    <!-- 图片预览对话框 -->
    <el-dialog v-model="previewVisible" title="图片预览" width="80%" append-to-body>
      <div class="preview-container">
        <img :src="previewUrl" alt="预览图片" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Document, Close, ZoomIn } from '@element-plus/icons-vue'

const props = defineProps({
  files: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['remove', 'preview'])

// 图片预览
const previewVisible = ref(false)
const previewUrl = ref('')

// 移除文件
const removeFile = (index) => {
  emit('remove', index)
}

// 预览图片
const previewImage = (file) => {
  previewUrl.value = file.preview || file.url
  previewVisible.value = true
}

// 获取文件类型样式类
const getFileClass = (type) => {
  if (type.includes('pdf')) return 'file-pdf'
  if (type.includes('word') || type.includes('document')) return 'file-word'
  if (type.includes('excel') || type.includes('spreadsheet')) return 'file-excel'
  if (type.includes('text') || type.includes('markdown')) return 'file-text'
  if (type.includes('html')) return 'file-html'
  return 'file-default'
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>

<style lang="scss" scoped>
.chat-file-upload {
  margin-bottom: 8px;

  .file-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .file-item {
    position: relative;
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.2s ease;

    &:hover {
      transform: translateY(-2px);
    }
  }

  .image-preview {
    position: relative;
    width: 80px;
    height: 80px;
    cursor: pointer;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid var(--el-border-color-lighter);

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .file-overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.3);
      display: flex;
      align-items: center;
      justify-content: center;
      opacity: 0;
      transition: opacity 0.2s;

      .preview-icon {
        color: #fff;
        font-size: 20px;
      }
    }

    &:hover .file-overlay {
      opacity: 1;
    }

    .remove-btn {
      position: absolute;
      top: 2px;
      right: 2px;
      width: 18px;
      height: 18px;
      background: rgba(0, 0, 0, 0.5);
      border-radius: 50%;
      color: #fff;
      font-size: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      opacity: 0;
      transition: opacity 0.2s;
    }

    &:hover .remove-btn {
      opacity: 1;
    }

    .file-name {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 2px 4px;
      background: rgba(0, 0, 0, 0.5);
      color: #fff;
      font-size: 10px;
      text-align: center;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .doc-preview {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    min-width: 160px;
    max-width: 200px;

    .file-icon-wrapper {
      width: 36px;
      height: 36px;
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;

      .file-icon {
        font-size: 18px;
        color: #fff;
      }

      &.file-pdf {
        background: linear-gradient(135deg, #ff4757, #ff6b81);
      }

      &.file-word {
        background: linear-gradient(135deg, #2b7de9, #5f9ee9);
      }

      &.file-excel {
        background: linear-gradient(135deg, #2ed573, #7bed9f);
      }

      &.file-text {
        background: linear-gradient(135deg, #a4b0be, #ced6e0);
      }

      &.file-html {
        background: linear-gradient(135deg, #ff6348, #ff7979);
      }

      &.file-default {
        background: linear-gradient(135deg, #747d8c, #a4b0be);
      }
    }

    .file-info {
      flex: 1;
      min-width: 0;

      .file-name {
        font-size: 12px;
        font-weight: 500;
        color: var(--el-text-color-primary);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .file-size {
        font-size: 11px;
        color: var(--el-text-color-secondary);
        margin-top: 2px;
      }
    }

    .remove-btn {
      flex-shrink: 0;
      width: 18px;
      height: 18px;
      color: var(--el-text-color-secondary);
      cursor: pointer;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s;

      &:hover {
        background: var(--el-color-danger-light-9);
        color: var(--el-color-danger);
      }
    }
  }

  .upload-progress {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;

    :deep(.el-progress-bar__outer) {
      background-color: transparent;
    }
  }
}

.preview-container {
  display: flex;
  justify-content: center;
  align-items: center;
  max-height: 70vh;

  img {
    max-width: 100%;
    max-height: 70vh;
    object-fit: contain;
  }
}
</style>
