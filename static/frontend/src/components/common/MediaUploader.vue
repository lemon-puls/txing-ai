<template>
  <div class="media-uploader">
    <el-upload
      :show-file-list="false"
      :auto-upload="false"
      :on-change="handleFileChange"
      :accept="acceptString"
    >
      <slot name="trigger">
        <div
          class="upload-area"
          :style="{
            width: typeof width === 'number' ? `${width}px` : width,
            height: typeof height === 'number' ? `${height}px` : height
          }"
        >
          <template v-if="modelValue">
            <img v-if="resolvedType === 'image'" :src="modelValue" class="preview" />
            <video
              v-else
              :src="modelValue"
              class="preview"
              muted
              preload="metadata"
            />
            <div class="badge">{{ resolvedType === 'image' ? '图片' : '视频' }}</div>
          </template>
          <div v-else class="placeholder">
            <el-icon class="upload-icon"><Plus /></el-icon>
            <div class="upload-text">{{ placeholder }}</div>
            <div class="upload-tip">支持 {{ mediaType }} {{ maxSize }}MB 以内</div>
          </div>
        </div>
      </slot>
    </el-upload>

    <!-- 预览/类型选择弹窗（视频可选替换） -->
    <el-dialog
      v-model="dialogVisible"
      title="媒体管理"
      width="560px"
      append-to-body
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-if="pendingFile" class="dialog-body">
        <div v-if="pendingInfo.type === 'image'" class="preview-block">
          <img :src="pendingInfo.url" class="preview-large" />
        </div>
        <div v-else class="preview-block">
          <video
            :src="pendingInfo.url"
            class="preview-large"
            controls
            autoplay
            muted
          />
        </div>
        <el-form label-width="80px" class="dialog-form">
          <el-form-item label="媒体类型">
            <el-radio-group v-model="chosenType" @change="onTypeChange">
              <el-radio value="image">图片</el-radio>
              <el-radio value="video">视频</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="说明">
            <el-input
              v-model="pendingCaption"
              placeholder="可选：媒体说明文字"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="cancelDialog">取消</el-button>
        <el-button type="primary" @click="confirmUpload" :loading="uploading">
          确认上传
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
// MediaUploader - 通用媒体上传组件（支持图片/视频）
// Generic media uploader for the about-me admin page.
// 选择文件 → 预览 → 二次确认 → 调用 uploadFileToOSS 上传 → emit('update:modelValue', url)
// 与 ImageUploader 的区别：1) 支持视频 2) 不强制裁剪 3) v-model 是单个 URL

import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { uploadFileToOSS, isImage, isVideo } from '@/utils/ossUpload'

const props = defineProps({
  modelValue: { type: String, default: '' },
  // 已知的媒体类型；用于决定 accept 过滤和预览 badge
  mediaType: {
    type: String,
    default: 'image',
    validator: (v) => ['image', 'video', 'all'].includes(v)
  },
  width: { type: [String, Number], default: 240 },
  height: { type: [String, Number], default: 160 },
  placeholder: { type: String, default: '点击上传' },
  maxSize: { type: Number, default: 50 }, // 文件大小限制 MB
  // 如果传入 model 端的已有 type（如后端返回的 media.type），强制锁定；否则按扩展名/dom 探测
  lockedType: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'success', 'error', 'change'])

// 已确认的媒体类型（用于预览决定渲染 img 还是 video）
const resolvedType = computed(() => {
  if (props.lockedType) return props.lockedType
  const url = props.modelValue
  if (!url) return props.mediaType === 'video' ? 'video' : 'image'
  return /\.(mp4|webm|mov|avi|m4v)(\?|$)/i.test(url) ? 'video' : 'image'
})

const acceptString = computed(() => {
  if (props.mediaType === 'image') return 'image/*'
  if (props.mediaType === 'video') return 'video/*'
  return 'image/*,video/*'
})

// === 弹窗 state ===
const dialogVisible = ref(false)
const pendingFile = ref(null)        // 待上传的 File
const pendingInfo = ref({ url: '', type: '' })  // 本地预览 (base64/object URL)
const chosenType = ref('image')
const pendingCaption = ref('')       // 暂存说明文字（不会自动回填，由调用方接管）
const uploading = ref(false)

const onTypeChange = () => {
  // 切换类型不影响已选的 file，仅决定最终上报
}

const handleFileChange = (file) => {
  if (!file || !file.raw) return
  const raw = file.raw

  if (props.mediaType === 'image' && !isImage(raw)) {
    ElMessage.error('请上传图片文件')
    return
  }
  if (props.mediaType === 'video' && !isVideo(raw)) {
    ElMessage.error('请上传视频文件')
    return
  }
  if (props.mediaType === 'all' && !isImage(raw) && !isVideo(raw)) {
    ElMessage.error('仅支持图片或视频文件')
    return
  }

  if (raw.size / 1024 / 1024 > props.maxSize) {
    ElMessage.error(`文件大小不能超过 ${props.maxSize}MB`)
    return
  }

  pendingFile.value = raw
  chosenType.value = isVideo(raw) ? 'video' : 'image'
  pendingInfo.value = {
    url: isVideo(raw) ? URL.createObjectURL(raw) : '',
    type: chosenType.value
  }

  // 图片走 FileReader → base64
  if (isImage(raw)) {
    const reader = new FileReader()
    reader.readAsDataURL(raw)
    reader.onload = (e) => {
      pendingInfo.value.url = e.target.result
      dialogVisible.value = true
    }
  } else {
    dialogVisible.value = true
  }
}

const cancelDialog = () => {
  if (pendingInfo.value.url && isVideo(pendingFile.value)) {
    URL.revokeObjectURL(pendingInfo.value.url)
  }
  pendingFile.value = null
  pendingInfo.value = { url: '', type: '' }
  dialogVisible.value = false
}

const confirmUpload = async () => {
  if (!pendingFile.value) return
  uploading.value = true
  try {
    const result = await uploadFileToOSS(pendingFile.value, {
      keyPrefix: 'media',
      mimeType: pendingFile.value.type
    })
    emit('update:modelValue', result.url)
    // 同时抛出 change，让父组件可以拿到 type + caption 一并更新
    emit('change', {
      url: result.url,
      type: chosenType.value,
      caption: pendingCaption.value
    })
    emit('success', result)
    ElMessage.success('上传成功')
    cancelDialog()
  } catch (e) {
    console.error('[MediaUploader] upload failed', e)
    emit('error', e)
    ElMessage.error(e?.message || '上传失败，请重试')
  } finally {
    uploading.value = false
  }
}
</script>

<style lang="scss" scoped>
.media-uploader {
  display: inline-block;
}

.upload-area {
  border: 2px dashed var(--el-border-color);
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  background: var(--el-fill-color-light);
  transition: all 0.3s;

  &:hover {
    border-color: var(--el-color-primary);
  }

  .preview {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .badge {
    position: absolute;
    top: 6px;
    right: 6px;
    padding: 2px 8px;
    background: rgba(0, 0, 0, 0.6);
    color: #fff;
    font-size: 12px;
    border-radius: 4px;
  }

  .placeholder {
    text-align: center;
    color: var(--el-text-color-secondary);
    padding: 12px;

    .upload-icon {
      font-size: 28px;
    }

    .upload-text {
      font-size: 14px;
      margin-top: 4px;
    }

    .upload-tip {
      font-size: 12px;
      color: var(--el-text-color-secondary);
      margin-top: 2px;
    }
  }
}

.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-block {
  width: 100%;
  max-height: 360px;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;

  .preview-large {
    max-width: 100%;
    max-height: 360px;
    object-fit: contain;
  }
}

.dialog-form {
  margin-top: 4px;
}
</style>