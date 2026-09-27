<template>
  <div class="proposal-card">
    <div class="card-header">
      <span class="card-badge">
        <el-icon :size="12"><MagicStick /></el-icon>
        {{ isUpdate ? '模型优化提案' : '模型录入提案' }}
      </span>
      <span v-if="isUpdate" class="card-link">ID: {{ proposal.id }}</span>
    </div>

    <!-- 重复/校验警告 -->
    <el-alert
      v-if="proposalStatus === 'duplicate'"
      :title="proposalMessage || '已存在同名模型，请勿重复录入'"
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
      <div class="fields">
        <div class="field-row">
          <span class="field-label">名称</span>
          <div class="field-input">
            <el-input v-model="form.name" size="small" maxlength="50" placeholder="模型名称" />
            <span v-if="originalHint('name')" class="original-hint">{{ originalHint('name') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">描述</span>
          <div class="field-input">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              size="small"
              maxlength="500"
              show-word-limit
              placeholder="模型描述"
            />
            <span v-if="originalHint('description')" class="original-hint">{{ originalHint('description') }}</span>
          </div>
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
        <!-- 快捷标签 -->
        <div class="field-row">
          <span class="field-label"></span>
          <div class="quick-tags">
            <el-tag
              v-for="preset in PRESET_MODEL_TAGS"
              :key="preset"
              size="small"
              effect="plain"
              round
              class="quick-tag"
              :class="{ active: tags.includes(preset) }"
              role="button"
              @click="togglePresetTag(preset)"
            >
              {{ preset }}
            </el-tag>
          </div>
        </div>
        <div v-if="originalHint('tag')" class="field-row">
          <span class="field-label"></span>
          <span class="original-hint">{{ originalHint('tag') }}</span>
        </div>
        <div class="field-row switch-row">
          <span class="field-label">能力</span>
          <div class="switches">
            <span class="switch-item">
              <el-switch v-model="form.high_context" size="small" />
              高上下文
            </span>
            <span class="switch-item">
              <el-switch v-model="form.multimodal" size="small" />
              多模态
            </span>
            <span class="switch-item">
              <el-switch v-model="form.default" size="small" />
              设为默认
            </span>
          </div>
        </div>
        <el-alert
          v-if="form.default"
          title="将设为全站默认模型，影响所有用户的默认对话体验"
          type="warning"
          :closable="false"
          show-icon
          class="default-alert"
        />
      </div>
    </div>

    <div class="card-footer">
      <span class="footer-hint">
        <el-icon :size="12"><CircleCheck /></el-icon>
        {{ confirmed ? (isUpdate ? '该模型已更新' : '该模型已录入') : '确认后才会写入数据库' }}
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
        {{ isUpdate ? '已更新' : '已录入' }}
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
        {{ isUpdate ? '确认更新' : '确认录入' }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

// 与后端 internal/tool/ops/tags.go 的 PresetModelTags 保持同步
const PRESET_MODEL_TAGS = ['通用', '联网搜索', '深度思考', '编程强化']

const props = defineProps({
  // 结构化提案 { type, id, name, description, default, high_context, multimodal, avatar, tags(逗号分隔) }
  proposal: { type: Object, required: true },
  // preview 工具给出的校验状态：ok | duplicate | confirmed
  proposalStatus: { type: String, default: 'ok' },
  proposalMessage: { type: String, default: '' },
  // 提案已确认入库（历史回放时禁用确认按钮，防止重复提交）
  confirmed: { type: Boolean, default: false },
  // 优化模式下的原值（来自页面的 draft 上下文，用于展示原值提示）
  original: { type: Object, default: null }
})

const emit = defineEmits(['confirmed'])

const submitting = ref(false)
const isUpdate = computed(() => Number(props.proposal.id) > 0)

const form = ref({
  name: props.proposal.name || '',
  description: props.proposal.description || '',
  high_context: !!props.proposal.high_context,
  multimodal: !!props.proposal.multimodal,
  default: !!props.proposal.default
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

const togglePresetTag = (preset) => {
  const index = tags.value.indexOf(preset)
  if (index > -1) {
    tags.value.splice(index, 1)
  } else if (tags.value.length < 5) {
    tags.value.push(preset)
  } else {
    ElMessage.warning('最多 5 个标签')
  }
}

// 原值对照提示（仅优化模式且字段有变化时展示）
const originalHint = (key) => {
  if (!isUpdate.value || !props.original) return ''
  const before = props.original[key]
  const after = key === 'tag' ? tags.value.join(',') : form.value[key]
  return before && before !== after ? `原：${before}` : ''
}

const canSubmit = computed(() => {
  if (submitting.value) return false
  const name = form.value.name.trim()
  return name.length >= 2 && name.length <= 50
})

const handleConfirm = async () => {
  submitting.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      description: form.value.description.trim(),
      default: form.value.default,
      high_context: form.value.high_context,
      multimodal: form.value.multimodal,
      tag: tags.value.join(',')
    }
    // 更新时不传 avatar（后端对空 avatar 不覆盖），原头像自动保留
    const response = isUpdate.value
      ? await defaultApi.apiAdminModelIdPut(props.proposal.id, payload)
      : await defaultApi.apiAdminModelPost(payload)
    if (response.code === 0) {
      emit('confirmed')
    } else {
      ElMessage.error(response.message || (isUpdate.value ? '更新失败' : '录入失败'))
    }
  } catch (error) {
    console.error('模型提案提交失败:', error)
    ElMessage.error(isUpdate.value ? '更新失败，请稍后重试' : '录入失败，请稍后重试')
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
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

.card-alert {
  margin: 10px 12px 0;
  padding: 5px 8px;
  border-radius: 10px;
}

.card-body {
  padding: 12px;
}

.fields {
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

.field-input {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.original-hint {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  line-height: 1.4;
  word-break: break-all;
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

.quick-tags {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;

  .quick-tag {
    cursor: pointer;
    user-select: none;
    color: var(--el-text-color-secondary);

    &.active {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary-light-5);
      background: var(--el-color-primary-light-9);
    }
  }
}

.switches {
  flex: 1;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;

  .switch-item {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--el-text-color-regular);
  }
}

.default-alert {
  margin-left: 42px;
  padding: 5px 8px;
  border-radius: 10px;
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
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致）
.proposal-card {
  html.dark & {
    box-shadow: none;

    .card-header {
      border-bottom-color: rgba(255, 255, 255, 0.06);
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
