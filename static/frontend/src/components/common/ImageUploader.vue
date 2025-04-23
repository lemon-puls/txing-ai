<template>
  <div class="image-uploader">
    <el-upload
      class="uploader"
      :show-file-list="false"
      :auto-upload="false"
      :on-change="handleFileChange"
      accept="image/*"
    >
      <div class="upload-area" :class="{ 'is-circle': circle }">
        <img v-if="modelValue" :src="modelValue" class="preview-image">
        <div v-else class="upload-placeholder">
          <el-icon class="upload-icon"><Plus /></el-icon>
          <div class="upload-text">{{ placeholder }}</div>
        </div>
      </div>
    </el-upload>

    <!-- 图片裁剪对话框 -->
    <el-dialog
      v-model="cropperVisible"
      :title="cropperTitle"
      width="600px"
      append-to-body
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="cropper-container">
        <VueCropper
          ref="cropperRef"
          :img="cropperImage"
          :outputSize="1"
          outputType="png"
          :info="true"
          :full="true"
          :canMove="true"
          :canMoveBox="true"
          :original="false"
          :autoCrop="true"
          :autoCropWidth="cropWidth"
          :autoCropHeight="cropHeight"
          :fixedBox="fixed"
          :fixed="fixed"
          :fixedNumber="aspectRatio"
          :centerBox="true"
          class="vue-cropper"
          :class="{ 'is-circle': circle }"
        />
      </div>
      <template #footer>
        <el-button @click="cropperVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCropImage" :loading="uploading">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import VueCropper from 'vue-cropper/lib/vue-cropper.vue'
import 'vue-cropper/dist/index.css'
import { defaultApi } from '@/api'

const props = defineProps({
  // v-model 绑定值，用于显示和更新图片 URL
  modelValue: {
    type: String,
    default: ''
  },
  // 是否开启裁剪功能
  enableCrop: {
    type: Boolean,
    default: true
  },
  // 是否为圆形裁剪
  circle: {
    type: Boolean,
    default: false
  },
  // 是否固定比例
  fixed: {
    type: Boolean,
    default: false
  },
  // 裁剪宽度
  cropWidth: {
    type: Number,
    default: 200
  },
  // 裁剪高度
  cropHeight: {
    type: Number,
    default: 200
  },
  // 裁剪比例 [宽, 高]
  aspectRatio: {
    type: Array,
    default: () => [1, 1]
  },
  // 占位文本
  placeholder: {
    type: String,
    default: '点击上传图片'
  },
  // 裁剪对话框标题
  cropperTitle: {
    type: String,
    default: '裁剪图片'
  },
  // 文件大小限制（MB）
  maxSize: {
    type: Number,
    default: 2
  }
})

const emit = defineEmits(['update:modelValue', 'success', 'error'])

// 裁剪相关
const cropperVisible = ref(false)
const cropperImage = ref('')
const cropperRef = ref(null)
const uploading = ref(false)

// 处理文件选择
const handleFileChange = async (file) => {
  if (!file) return

  // 验证文件类型和大小
  const isImage = file.raw.type.startsWith('image/')
  const isLtMax = file.raw.size / 1024 / 1024 < props.maxSize

  if (!isImage) {
    ElMessage.error('请上传图片文件！')
    return
  }
  if (!isLtMax) {
    ElMessage.error(`图片大小不能超过 ${props.maxSize}MB！`)
    return
  }

  if (props.enableCrop) {
    // 开启裁剪功能，显示裁剪对话框
    const reader = new FileReader()
    reader.readAsDataURL(file.raw)
    reader.onload = (e) => {
      cropperImage.value = e.target.result
      cropperVisible.value = true
    }
  } else {
    // 不开启裁剪功能，直接上传
    try {
      uploading.value = true
      
      // 生成随机文件名
      const fileExt = file.raw.name.split('.').pop() || 'png'
      const fileName = `${Date.now()}-${Math.floor(Math.random() * 1000)}.${fileExt}`

      // 获取预签名 URL
      const res = await defaultApi.apiCosPresignedUrlPost({
        type: 'upload',
        key: fileName
      })

      if (res.code !== 0) {
        throw new Error(res.msg || '获取上传地址失败')
      }

      // 上传文件到预签名 URL
      const response = await fetch(res.data.url, {
        method: 'PUT',
        body: file.raw,
        headers: {
          'Content-Type': file.raw.type
        }
      })

      if (!response.ok) {
        throw new Error('上传失败')
      }

      // 获取待签名下载 URL
      const res1 = await defaultApi.apiCosPresignedUrlPost({
        key: fileName,
        type: 'download'
      })

      if (res1.code !== 0) {
        throw new Error(res1.msg || '获取访问地址失败')
      }

      // 更新 v-model
      emit('update:modelValue', res1.data.url)
      emit('success', res1.data.url)
      ElMessage.success('上传成功')
    } catch (error) {
      console.error('上传失败:', error)
      emit('error', error)
      ElMessage.error(error.message || '上传失败，请重试')
    } finally {
      uploading.value = false
    }
  }
}

// 处理图片裁剪和上传
const handleCropImage = async () => {
  if (!cropperRef.value) return

  try {
    uploading.value = true

    // 获取裁剪后的图片数据
    const base64Data = await new Promise((resolve) => {
      cropperRef.value.getCropData((data) => resolve(data))
    })

    // 将 base64 转换为 Blob
    const blob = await fetch(base64Data).then(res => res.blob())

    // 生成随机文件名
    const fileExt = 'png'
    const fileName = `${Date.now()}-${Math.floor(Math.random() * 1000)}.${fileExt}`

    try {
      // 获取预签名 URL
      const res = await defaultApi.apiCosPresignedUrlPost({
        type: 'upload',
        key: fileName
      })

      if (res.code !== 0) {
        throw new Error(res.msg || '获取上传地址失败')
      }

      // 上传文件到预签名 URL
      const response = await fetch(res.data.url, {
        method: 'PUT',
        body: blob,
        headers: {
          'Content-Type': 'image/png'
        }
      })

      if (!response.ok) {
        throw new Error('上传失败')
      }

      // 获取待签名下载 URL
      const res1 = await defaultApi.apiCosPresignedUrlPost({
        key: fileName,
        type: 'download'
      })

      if (res1.code !== 0) {
        throw new Error(res1.msg || '获取访问地址失败')
      }

      // 处理返回的 URL（移除签名参数）
      // const urlObj = new URL(res1.data.url)
      // const finalUrl = `${urlObj.origin}${urlObj.pathname}`

      // 更新 v-model
      emit('update:modelValue', res1.data.url)
      emit('success', res1.data.url)

      // 关闭裁剪对话框
      cropperVisible.value = false
      ElMessage.success('上传成功')
    } catch (error) {
      console.error('上传失败:', error)
      emit('error', error)
      ElMessage.error(error.message || '上传失败，请重试')
    }
  } catch (error) {
    console.error('裁剪失败:', error)
    emit('error', error)
    ElMessage.error('图片裁剪失败，请重试')
  } finally {
    uploading.value = false
  }
}
</script>

<style lang="scss" scoped>
.image-uploader {
  .uploader {
    :deep(.el-upload) {
      width: 100%;
    }
  }
}

.upload-area {
  width: 200px;
  height: 200px;
  border: 2px dashed var(--el-border-color);
  border-radius: 8px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
  display: flex;
  justify-content: center;
  align-items: center;

  &.is-circle {
    border-radius: 50%;

    .preview-image {
      border-radius: 50%;
    }
  }

  &:hover {
    border-color: var(--el-color-primary);
    transform: translateY(-2px);
  }

  .preview-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .upload-placeholder {
    text-align: center;
    color: var(--el-text-color-secondary);

    .upload-icon {
      font-size: 30px;
      margin-bottom: 8px;
    }

    .upload-text {
      font-size: 14px;
    }
  }
}

.cropper-container {
  height: 500px;
  background: #262626;

  :deep(.vue-cropper) {
    height: 100%;
    width: 100%;

    &.is-circle {
      .cropper-view-box {
        border-radius: 50%;
        outline: 1px solid #fff;
      }

      .cropper-face {
        border-radius: 50%;
        background-color: transparent;
      }

      .cropper-line {
        background: none;
      }

      .cropper-point {
        width: 8px !important;
        height: 8px !important;
        opacity: 1 !important;
        background-color: #fff !important;
        border-radius: 50%;

        &.point-se {
          width: 8px !important;
          height: 8px !important;
        }
      }
    }

    .cropper-view-box {
      outline: 1px solid #fff;
    }

    .cropper-face {
      background-color: transparent !important;
    }

    .cropper-point {
      width: 8px !important;
      height: 8px !important;
      opacity: 1 !important;
      background-color: #fff !important;
      border-radius: 50%;

      &.point-se {
        width: 8px !important;
        height: 8px !important;
      }
    }

    .cropper-line {
      background: #fff;
      opacity: 0.3;
    }
  }
}
</style>
