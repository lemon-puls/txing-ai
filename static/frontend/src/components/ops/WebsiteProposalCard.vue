<template>
  <div class="proposal-card">
    <div class="card-header">
      <span class="card-badge">
        <el-icon :size="12"><Link /></el-icon>
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
            <el-button v-else size="small" link type="primary" @click="showTagInput">
              + 标签
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <span class="footer-hint">确认后将调用创建接口入库</span>
      <el-button
        type="primary"
        size="small"
        round
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
import { Link, TopRight } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const props = defineProps({
  // 结构化提案 { type, name, description, url, avatar, tags(逗号分隔) }
  proposal: { type: Object, required: true },
  // preview 工具给出的校验状态：ok | duplicate
  proposalStatus: { type: String, default: 'ok' },
  proposalMessage: { type: String, default: '' }
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
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: var(--el-bg-color);
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-color-primary-light-9);

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
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-decoration: none;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &:hover {
      color: var(--el-color-primary);
    }
  }
}

.card-alert {
  margin: 8px 12px 0;
  padding: 5px 8px;
}

.card-body {
  display: flex;
  gap: 10px;
  padding: 12px;
}

.avatar-box {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  flex-shrink: 0;
  overflow: hidden;
  background: var(--el-fill-color-light);
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
    width: 32px;
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 24px;
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
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter);

  .footer-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}
</style>
