<template>
  <el-dialog
    v-model="dialogVisible"
    title="创建助手"
    width="640px"
    :close-on-click-modal="false"
    @close="handleClose"
    class="create-assistant-dialog"
    :class="{ 'dialog-fade-enter': dialogVisible }"
    :show-close="false"
  >
    <template #header="{ close }">
      <div class="dialog-header">
        <div class="header-content">
          <div class="title-section">
            <div class="icon-wrapper">
              <el-icon class="dialog-icon">
                <component :is="editData ? Edit : Plus" />
              </el-icon>
            </div>
            <h2 class="dialog-title">{{ editData ? '编辑助手' : '创建助手' }}</h2>
          </div>
          <el-button class="close-btn" @click="close">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
        <div class="header-decoration"></div>
      </div>
    </template>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="create-form"
    >
      <div class="form-grid">
        <el-form-item label="助手名称" prop="name" class="name-item">
          <el-input
            v-model="form.name"
            placeholder="给您的助手起个名字"
            maxlength="50"
            show-word-limit
          >
            <template #prefix>
              <el-icon><Edit /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item prop="avatar" class="avatar-item">
          <el-tooltip
            content="点击上传助手头像"
            placement="top"
            :show-after="500"
          >
            <div>
              <ImageUploader
                v-model="form.avatar"
                :circle="false"
                :width="120"
                :height="120"
                :crop-width="200"
                :crop-height="200"
                :fixed="true"
                placeholder="上传头像"
                @success="handleAvatarSuccess"
                @error="handleAvatarError"
              />
            </div>
          </el-tooltip>
        </el-form-item>

        <el-form-item label="助手描述" prop="description" class="description-item">
          <el-input
            v-model="form.description"
            type="textarea"
            placeholder="描述一下您的助手的功能和特点"
            :rows="4"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="助手类型" prop="tags" class="tag-item">
          <el-select
            v-model="form.tags"
            multiple
            collapse-tags
            collapse-tags-tooltip
            :max-collapse-tags="3"
            placeholder="选择助手类型（最多3个）"
            :multiple-limit="3"
            class="tag-select"
          >
            <el-option
              v-for="tag in tags"
              :key="tag.id"
              :label="tag.name"
              :value="tag.id"
            >
              <div class="category-option">
                <el-icon><component :is="getTagsIcon(tag.id)" /></el-icon>
                <span>{{ tag.name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="系统提示词" prop="context" class="prompt-item">
          <div class="prompt-wrapper">
            <el-input
              v-model="form.context"
              type="textarea"
              placeholder="设置助手的系统提示词，定义助手的行为和专业领域"
              :rows="6"
              maxlength="2000"
              show-word-limit
            />
            <div class="prompt-tips">
              <el-tooltip
                effect="dark"
                content="系统提示词用于定义助手的行为和专业领域，例如：'你是一个专业的前端开发专家...'"
                placement="top"
              >
                <el-icon class="tips-icon"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
        </el-form-item>
      </div>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button plain @click="handleClose" class="cancel-btn">
          <el-icon><Close /></el-icon>
          取消
        </el-button>
        <el-button type="primary" @click="submitForm(formRef)" class="submit-btn">
          <el-icon><Check /></el-icon>
          {{ editData ? '保存修改' : '创建助手' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import {
  Plus,
  Edit,
  Close,
  Check,
  QuestionFilled,
  Tools,
  Edit as EditIcon,
  Monitor,
  Reading,
  House,
  More
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ImageUploader from '../common/ImageUploader.vue'
import { defaultApi } from '@/api/index.js'

defineOptions({
  name: 'CreateAssistantDialog'
})

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  editData: {
    type: Object,
    default: null
  }
})

// 定义所有的 emits，并添加验证
const emit = defineEmits({
  'update:visible': (value) => typeof value === 'boolean',
  'created': () => true,
  'updated': () => true
})

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const formRef = ref(null)

// 表单数据
const form = reactive({
  name: '',
  avatar: '',
  description: '',
  tags: [],
  context: ''
})

// 重置表单
const resetForm = () => {
  form.name = ''
  form.avatar = ''
  form.description = ''
  form.tags = []
  form.context = ''
}

// 处理关闭
const handleClose = () => {
  dialogVisible.value = false
  resetForm()
}

// 监听编辑数据变化
watch(
  () => props.editData,
  (newVal) => {
    if (newVal) {
      // 填充表单数据
      form.name = newVal.name
      form.avatar = newVal.avatar
      form.description = newVal.description
      form.tags = newVal.tags ? newVal.tags.split(',') : []
      form.context = newVal.context
    } else {
      // 重置表单
      resetForm()
    }
  },
  { immediate: true }
)

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入助手名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入助手描述', trigger: 'blur' },
    { max: 200, message: '描述不能超过 200 个字符', trigger: 'blur' }
  ],
  tags: [
    { required: true, message: '请至少选择一个助手类型', trigger: 'change' },
    {
      validator: (rule, value, callback) => {
        if (value && value.length > 3) {
          callback(new Error('最多只能选择3个类型'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  context: [
    { required: true, message: '请输入系统提示词', trigger: 'blur' },
    { max: 2000, message: '系统提示词不能超过 2000 个字符', trigger: 'blur' }
  ]
}

// 分类数据
const tags = [
  { id: 'tools', name: '实用工具' },
  { id: 'writing', name: '文案创作' },
  { id: 'coding', name: '编码专家' },
  { id: 'learning', name: '知识学习' },
  { id: 'life', name: '生活指南' },
  { id: 'other', name: '其他' }
]

// 获取分类图标
const getTagsIcon = (tagId) => {
  const iconMap = {
    tools: Tools,
    writing: EditIcon,
    coding: Monitor,
    learning: Reading,
    life: House,
    other: More
  }
  return iconMap[tagId] || More
}

// 处理头像上传成功
const handleAvatarSuccess = async (url) => {
  form.avatar = url
}

// 处理头像上传错误
const handleAvatarError = (error) => {
  ElMessage.error('头像上传失败：' + error.message)
}

// 提交表单
const submitForm = async (formEl) => {
  if (!formEl) return

  await formEl.validate(async (valid) => {
    if (valid) {
      try {
        const params = {
          name: form.name,
          avatar: form.avatar,
          description: form.description,
          tags: form.tags.join(','),
          context: form.context
        }

        let response
        if (props.editData) {
          // 更新助手
          response = await defaultApi.apiPresetIdPut(props.editData.id, params)
        } else {
          // 创建助手
          response = await defaultApi.apiPresetPost(params)
        }

        if (response.code === 0) {
          ElMessage.success(props.editData ? '助手更新成功' : '助手创建成功')
          dialogVisible.value = false
          resetForm()
          emit(props.editData ? 'updated' : 'created')
        } else {
          ElMessage.error(response.msg || (props.editData ? '更新失败' : '创建失败'))
        }
      } catch (error) {
        console.error(props.editData ? 'Update assistant error:' : 'Create assistant error:', error)
        ElMessage.error(error.body?.msg || (props.editData ? '更新失败' : '创建失败'))
      }
    }
  })
}

</script>

<style lang="scss" scoped>
.create-assistant-dialog {
  :deep(.el-dialog) {
    border-radius: 16px;
    overflow: hidden;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    background: linear-gradient(135deg, var(--el-bg-color) 0%, var(--el-bg-color-overlay) 100%);

    .el-dialog__header {
      margin: 0;
      padding: 0;
    }

    .el-dialog__body {
      padding: 24px;
    }

    .el-dialog__footer {
      padding: 16px 24px;
      border-top: 1px solid var(--el-border-color-light);
      background: var(--el-bg-color-overlay);
    }

    .el-textarea__inner {
      width: 100%;
      min-height: 150px;
      font-family: var(--el-font-family);
      line-height: 1.6;
      resize: vertical;
    }
  }
}

.dialog-header {
  position: relative;

  .header-content {
    padding: 20px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .title-section {
      display: flex;
      align-items: center;
      gap: 12px;

      .icon-wrapper {
        width: 40px;
        height: 40px;
        border-radius: 12px;
        background: var(--el-color-primary);
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 20px;
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.3);
        animation: iconFloat 3s ease-in-out infinite;
      }

      .dialog-title {
        margin: 0;
        font-size: 18px;
        font-weight: 600;
        background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-primary-light-3) 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
      }
    }

    .close-btn {
      border: none;
      background: transparent;
      padding: 8px;
      border-radius: 8px;
      transition: all 0.3s;

      &:hover {
        background: var(--el-color-danger-light-9);
        color: var(--el-color-danger);
        transform: rotate(90deg);
      }
    }
  }

  .header-decoration {
    height: 4px;
    background: linear-gradient(90deg,
      var(--el-color-primary) 0%,
      var(--el-color-primary-light-3) 50%,
      var(--el-color-primary) 100%
    );
    opacity: 0.8;
  }
}

.form-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
  margin-top: 16px;

  .name-item {
    grid-column: 1;
  }

  .avatar-item {
    grid-column: 2;
    grid-row: 1 / span 2;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 0;

    :deep(.el-form-item__content) {
      margin: 0 !important;
      justify-content: center;
    }
  }

  .description-item {
    grid-column: 1;
  }

  .tag-item,
  .prompt-item {
    grid-column: 1 / span 2;
  }

  .prompt-wrapper {
    width: 100%;

    .el-input {
      width: 100%;
    }
  }
}

.tag-select {
  :deep(.el-select__tags) {
    max-width: calc(100% - 30px);
  }

  :deep(.el-select__tags-text) {
    max-width: 100px;
    display: inline-block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.el-tag) {
    max-width: 120px;
    display: inline-flex;
    align-items: center;

    .el-tag__content {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.category-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;

  .el-icon {
    font-size: 16px;
  }
}

.prompt-wrapper {
  position: relative;

  .prompt-tips {
    position: absolute;
    top: -24px;
    right: 0;

    .tips-icon {
      color: var(--el-color-info);
      cursor: help;
      transition: all 0.3s;

      &:hover {
        color: var(--el-color-primary);
        transform: scale(1.1);
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;

  .cancel-btn,
  .submit-btn {
    padding: 8px 24px;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.3s;

    &:hover {
      transform: translateY(-2px);
    }
  }

  .submit-btn {
    background: linear-gradient(135deg,
      var(--el-color-primary) 0%,
      var(--el-color-primary-light-3) 100%
    );
    border: none;
    box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.3);

    &:hover {
      box-shadow: 0 6px 16px rgba(var(--el-color-primary-rgb), 0.4);
    }
  }
}

@keyframes iconFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

.dialog-fade-enter {
  animation: dialogFadeIn 0.3s ease-out;
}

@keyframes dialogFadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
