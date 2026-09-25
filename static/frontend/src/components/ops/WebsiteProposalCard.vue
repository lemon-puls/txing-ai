<template>
  <div class="proposal-card">
    <div class="card-header">
      <span class="card-badge">
        <el-icon :size="12"><MagicStick /></el-icon>
        网站录入提案
      </span>
      <a :href="proposal.url" target="_blank" rel="noopener noreferrer" class="card-link">
        {{ proposal.url }}
        <el-icon :size="12"><TopRight /></el-icon>
      </a>
    </div>

    <!-- 重复/校验警告 -->
    <el-alert
      v-if="proposalStatus === 'duplicate'"
      :title="proposalMessage || '该网站地址已存在，请勿重复录入'"
      type="warning"
      :closable="false"
      show-icon
      class="card-alert"
    />
    <el-alert
      v-else-if="proposalMessage"
      :title="proposalMessage"
      type="info"
      :closable="false"
      show-icon
      class="card-alert"
    />

    <div class="card-body">
      <!-- 头像预览 -->
      <div class="avatar-box">
        <img v-if="form.avatar" :src="form.avatar" alt="网站头像" @error="avatarError = true" />
        <el-icon v-else :size="22"><Link /></el-icon>
      </div>

      <div class="fields">
        <div class="field-row">
          <span class="field-label">名称</span>
          <el-input v-model="form.name" size="small" maxlength="50" placeholder="网站名称" />
        </div>
        <div class="field-row">
          <span class="field-label">描述</span>
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            size="small"
            maxlength="200"
            show-word-limit
            placeholder="网站描述"
          />
        </div>
        <div class="field-row">
          <span class="field-label">标签</span>
          <div class="tags-editor">
            <el-tag
              v-for="tag in tags"
              :key="tag"
              closable
              size="small"
              effect="light"
              round
              @close="removeTag(tag)"
            >
              {{ tag }}
            </el-tag>
            <el-input
              v-if="tagInputVisible"
              ref="tagInputRef"
              v-model="tagInputValue"
              size="small"
              style="width: 90px"
              @keyup.enter="confirmTag"
              @blur="confirmTag"
            />
            <el-button v-else size="small" link type="primary" class="add-tag-btn" @click="showTagInput">
              <el-icon :size="12"><Plus /></el-icon>
              添加标签
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <span class="footer-hint">
        <el-icon :size="12"><CircleCheck /></el-icon>
        {{ confirmed ? '该网站已录入' : '确认后才会写入数据库' }}
      </span>
      <el-button
        v-if="confirmed"
        type="success"
        size="small"
        round
        class="confirmed-btn"
        disabled
      >
        <el-icon class="btn-icon"><Check /></el-icon>
        已录入
      </el-button>
      <el-button
        v-else
        type="primary"
        size="small"
        round
        class="confirm-btn"
        :loading="submitting"
        :disabled="!canSubmit"
        @click="handleConfirm"
      >
        确认录入
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Link, TopRight, Check } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const props = defineProps({
  // 结构化提案 { type, name, description, url, avatar, tags(逗号分隔) }
  proposal: { type: Object, required: true },
  // preview 工具给出的校验状态：ok | duplicate | confirmed
  proposalStatus: { type: String, default: 'ok' },
  proposalMessage: { type: String, default: '' },
  // 提案已确认入库（历史回放时禁用确认按钮，防止重复录入）
  confirmed: { type: Boolean, default: false }
})

const emit = defineEmits(['confirmed'])

const submitting = ref(false)
const avatarError = ref(false)

// 名称/描述可编辑；头像仅展示（预签名 URL 过期后可留空人工补）
const form = ref({
  name: props.proposal.name || '',
  description: props.proposal.description || '',
  avatar: props.proposal.avatar || ''
})

// 标签拆成数组编辑，确认时再拼接
const tags = ref((props.proposal.tags || '').split(',').map(t => t.trim()).filter(Boolean))
const tagInputVisible = ref(false)
const tagInputValue = ref('')
const tagInputRef = ref(null)

const showTagInput = () => {
  tagInputVisible.value = true
  nextTick(() => tagInputRef.value?.focus())
}

const confirmTag = () => {
  const value = tagInputValue.value.trim()
  if (value && !tags.value.includes(value) && tags.value.length < 5) {
    tags.value.push(value)
  } else if (value && tags.value.length >= 5) {
    ElMessage.warning('最多 5 个标签')
  }
  tagInputVisible.value = false
  tagInputValue.value = ''
}

const removeTag = (tag) => {
  const index = tags.value.indexOf(tag)
  if (index > -1) tags.value.splice(index, 1)
}

const canSubmit = computed(() => {
  if (submitting.value) return false
  if (tags.value.length === 0) return false
  const name = form.value.name.trim()
  const description = form.value.description.trim()
  return name.length >= 2 && name.length <= 50 && description.length > 0
})

const handleConfirm = async () => {
  submitting.value = true
  try {
    const response = await defaultApi.apiAdminWebsitesPost({
      name: form.value.name.trim(),
      description: form.value.description.trim(),
      url: props.proposal.url,
      avatar: form.value.avatar,
      tags: tags.value.join(','),
      sort: 0,
      status: 1
    })
    if (response.code === 0) {
      emit('confirmed')
    } else {
      ElMessage.error(response.message || '录入失败')
    }
  } catch (error) {
    console.error('网站录入失败:', error)
    ElMessage.error('录入失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.proposal-card {
  margin-top: 10px;
  width: 100%;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: var(--el-bg-color);
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.2s ease;

  &:hover {
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08);
  }
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 12px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-color-primary-light-8));
  border-bottom: 1px solid var(--el-color-primary-light-8);

  .card-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    font-weight: 600;
    color: var(--el-color-primary);
    flex-shrink: 0;
  }

  .card-link {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    min-width: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-decoration: none;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: color 0.2s ease;

    &:hover {
      color: var(--el-color-primary);
      text-decoration: underline;
    }
  }
}

.card-alert {
  margin: 10px 12px 0;
  padding: 5px 8px;
  border-radius: 10px;
}

.card-body {
  display: flex;
  gap: 12px;
  padding: 12px;
}

.avatar-box {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  flex-shrink: 0;
  overflow: hidden;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-secondary);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.fields {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;

  .field-label {
    width: 34px;
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 24px;
    text-align: justify;
  }

  :deep(.el-input),
  :deep(.el-textarea) {
    flex: 1;
  }
}

.tags-editor {
  flex: 1;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;

  .add-tag-btn {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    font-size: 12px;
  }
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter);

  .footer-hint {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

// 确认按钮渐变强化 CTA
.confirm-btn {
  border: none;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  box-shadow: 0 2px 8px var(--el-color-primary-light-8);

  &:not(:disabled):hover {
    opacity: 0.88;
  }
}

// 已录入成功态
.confirmed-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;

  .btn-icon {
    margin-right: 2px;
  }
}

// 暗色模式补充：项目暗色主题未覆写 --el-fill-color-lighter 等变量，这里手动补齐
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.proposal-card {
  html.dark & {
    box-shadow: none;

    .card-header {
      border-bottom-color: rgba(255, 255, 255, 0.06);
    }

    .avatar-box {
      border-color: #363637;
      box-shadow: none;
    }

    .card-footer {
      background: #1c1c1c;
      border-top-color: #363637;
    }

    .confirm-btn {
      box-shadow: none;
    }
  }
}
</style>
