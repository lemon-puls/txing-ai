<template>
  <div class="proposal-card">
    <div class="card-header">
      <span class="card-badge">
        <el-icon :size="12"><MagicStick /></el-icon>
        {{ isUpdate ? '助手优化提案' : '助手录入提案' }}
      </span>
      <span v-if="isUpdate" class="card-link">ID: {{ proposal.id }}</span>
    </div>

    <!-- 重复/校验警告 -->
    <el-alert
      v-if="proposalStatus === 'duplicate'"
      :title="proposalMessage || '已存在同名助手，请勿重复录入'"
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
            <el-input v-model="form.name" size="small" maxlength="50" placeholder="助手名称" />
            <span v-if="originalHint('name')" class="original-hint">{{ originalHint('name') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">描述</span>
          <div class="field-input">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="2"
              size="small"
              maxlength="200"
              show-word-limit
              placeholder="助手描述"
            />
            <span v-if="originalHint('description')" class="original-hint">{{ originalHint('description') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">标签</span>
          <div class="field-input">
            <el-select v-model="tags" multiple filterable :multiple-limit="3" size="small" placeholder="最多选 3 个标签">
              <el-option v-for="opt in TAG_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
            <span v-if="originalHint('tags')" class="original-hint">{{ originalHint('tags') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">设定</span>
          <div class="field-input">
            <el-input
              v-model="form.context"
              type="textarea"
              :rows="8"
              size="small"
              maxlength="8000"
              show-word-limit
              placeholder="上下文设定（系统提示词）"
            />
            <div v-if="isUpdate && originalContext" class="original-context">
              <div class="original-toggle" role="button" @click="originalExpanded = !originalExpanded">
                <el-icon :size="12"><ArrowDown /></el-icon>
                <span>{{ originalExpanded ? '收起原设定' : '查看原设定' }}</span>
              </div>
              <div v-show="originalExpanded" class="original-content">{{ originalContext }}</div>
            </div>
          </div>
        </div>
        <div class="field-row switch-row">
          <span class="field-label">官方</span>
          <div class="switches">
            <el-switch v-model="form.official" size="small" />
            <span class="switch-text">标记为官方助手</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <span class="footer-hint">
        <el-icon :size="12"><CircleCheck /></el-icon>
        {{ confirmed ? (isUpdate ? '该助手已更新' : '该助手已录入') : '确认后才会写入数据库' }}
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
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowDown, Check } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

// 与后端 internal/tool/ops/preset_tools.go 的 PresetPresetTags、
// views/admin/preset/PresetList.vue 的 tagOptions 保持同步
const TAG_OPTIONS = [
  { value: 'popular', label: '热门推荐' },
  { value: 'tools', label: '实用工具' },
  { value: 'writing', label: '文案创作' },
  { value: 'coding', label: '编码专家' },
  { value: 'learning', label: '知识学习' },
  { value: 'life', label: '生活指南' },
  { value: 'other', label: '其他' }
]

const props = defineProps({
  // 结构化提案 { type, id, name, description, context, tags(逗号分隔), official }
  proposal: { type: Object, required: true },
  // preview 工具给出的校验状态：ok | duplicate | confirmed
  proposalStatus: { type: String, default: 'ok' },
  proposalMessage: { type: String, default: '' },
  // 提案已确认入库（历史回放时禁用确认按钮，防止重复提交）
  confirmed: { type: Boolean, default: false },
  // 优化模式下的原值（来自页面的 draft 上下文，用于展示原值对照）
  original: { type: Object, default: null }
})

const emit = defineEmits(['confirmed'])

const submitting = ref(false)
const isUpdate = computed(() => Number(props.proposal.id) > 0)

const form = ref({
  name: props.proposal.name || '',
  description: props.proposal.description || '',
  context: props.proposal.context || '',
  official: !!props.proposal.official
})

// 标签拆成数组编辑，确认时再拼接
const tags = ref((props.proposal.tags || '').split(',').map(t => t.trim()).filter(Boolean))

// 优化模式：原上下文设定折叠对照
const originalExpanded = ref(false)
const originalContext = computed(() => (isUpdate.value && props.original?.context) || '')

// 原值对照提示（仅优化模式且字段有变化时展示）
const originalHint = (key) => {
  if (!isUpdate.value || !props.original) return ''
  const before = props.original[key]
  const after = key === 'tags' ? tags.value.join(',') : form.value[key]
  return before && before !== after ? `原：${before}` : ''
}

const canSubmit = computed(() => {
  if (submitting.value) return false
  const name = form.value.name.trim()
  return name.length >= 2 && name.length <= 50 && form.value.context.trim().length > 0
})

const handleConfirm = async () => {
  submitting.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      description: form.value.description.trim(),
      context: form.value.context.trim(),
      tags: tags.value.join(','),
      official: form.value.official
    }
    // 更新时不传 avatar（后端对空 avatar 不覆盖），原头像自动保留；
    // 注意用 apiPresetIdPut（apiAdminPresetIdPut 在生成的客户端中不存在）
    const response = isUpdate.value
      ? await defaultApi.apiPresetIdPut(props.proposal.id, payload)
      : await defaultApi.apiPresetPost(payload)
    if (response.code === 0) {
      emit('confirmed')
    } else {
      ElMessage.error(response.message || (isUpdate.value ? '更新失败' : '录入失败'))
    }
  } catch (error) {
    console.error('助手提案提交失败:', error)
    ElMessage.error(error?.response?.data?.message || (isUpdate.value ? '更新失败，请稍后重试' : '录入失败，请稍后重试'))
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

.original-context {
  margin-top: 4px;

  .original-toggle {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    cursor: pointer;
    user-select: none;

    &:hover {
      color: var(--el-color-primary);
    }
  }

  .original-content {
    margin-top: 4px;
    padding: 8px 10px;
    max-height: 200px;
    overflow-y: auto;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    white-space: pre-wrap;
    word-break: break-word;
    background: var(--el-fill-color-lighter);
    border-radius: 10px;
  }
}

.switches {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;

  .switch-text {
    font-size: 12px;
    color: var(--el-text-color-regular);
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

    .original-content {
      background: #262727;
    }

    .confirm-btn {
      box-shadow: none;
    }
  }
}
</style>
