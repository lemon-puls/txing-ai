<template>
  <div class="proposal-card">
    <div class="card-header">
      <span class="card-badge">
        <el-icon :size="12"><MagicStick /></el-icon>
        {{ isUpdate ? '渠道优化提案' : '渠道录入提案' }}
      </span>
      <span v-if="isUpdate" class="card-link">ID: {{ proposal.id }}</span>
    </div>

    <!-- 重复/校验警告 -->
    <el-alert
      v-if="proposalStatus === 'duplicate'"
      :title="proposalMessage || '已存在同名渠道，请勿重复录入'"
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
            <el-input v-model="form.name" size="small" maxlength="50" placeholder="渠道名称" />
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">类型</span>
          <div class="field-input">
            <el-select v-model="form.channelType" size="small" placeholder="渠道类型">
              <el-option v-for="t in CHANNEL_TYPES" :key="t" :label="t" :value="t" />
            </el-select>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">地址</span>
          <div class="field-input">
            <el-input v-model="form.endpoint" size="small" placeholder="服务地址，如 https://api.example.com/v1" />
            <span v-if="originalHint('endpoint')" class="original-hint">{{ originalHint('endpoint') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">模型</span>
          <div class="field-input">
            <el-select
              v-model="form.models"
              multiple
              filterable
              allow-create
              default-first-option
              size="small"
              placeholder="支持的模型，输入可新建"
            >
              <el-option v-for="m in form.models" :key="m" :label="m" :value="m" />
            </el-select>
            <span v-if="originalHint('models')" class="original-hint">{{ originalHint('models') }}</span>
          </div>
        </div>
        <div class="field-row">
          <span class="field-label">调度</span>
          <div class="nums">
            <el-input-number v-model="form.priority" size="small" :min="0" controls-position="right" />
            <span class="num-label">优先级</span>
            <el-input-number v-model="form.weight" size="small" :min="0" controls-position="right" />
            <span class="num-label">权重</span>
            <el-input-number v-model="form.retry" size="small" :min="0" controls-position="right" />
            <span class="num-label">重试</span>
          </div>
        </div>
        <div class="field-row switch-row">
          <span class="field-label">状态</span>
          <div class="switches">
            <el-switch v-model="form.status" size="small" />
            <span class="switch-text">{{ form.status ? '启用' : '禁用' }}</span>
          </div>
        </div>

        <!-- 模型映射（只读展示，精细调整请走编辑表单） -->
        <template v-if="mappings.length">
          <div class="field-row">
            <span class="field-label">映射</span>
            <div class="mappings">
              <div v-for="(m, i) in mappings" :key="i" class="mapping-item">
                <span class="mapping-source">{{ m.sourceModel }}</span>
                <template v-for="(c, j) in m.conditions || []" :key="j">
                  <el-icon :size="10" class="mapping-arrow"><ArrowRight /></el-icon>
                  <span class="mapping-target">{{ c.targetModel }}</span>
                  <span class="mapping-conds">{{ formatConds(c.conditions) }}</span>
                </template>
              </div>
            </div>
          </div>
          <div class="field-row">
            <span class="field-label"></span>
            <span class="mapping-hint">映射为只读预览，如需精细调整请在渠道编辑表单中修改</span>
          </div>
        </template>

        <!-- 密钥区：AI 永远不生成、不传输、不展示密钥 -->
        <div v-if="!isUpdate" class="field-row">
          <span class="field-label">密钥</span>
          <div class="field-input">
            <el-input
              v-model="secretInput"
              type="password"
              show-password
              size="small"
              placeholder="密钥由管理员手动填写，AI 不会生成或看到密钥"
            />
            <span class="secret-hint">多个密钥用换行分隔，请求时随机轮询</span>
          </div>
        </div>
        <el-alert
          v-else
          title="密钥不会被修改；如需更换密钥请在渠道编辑表单中操作"
          type="info"
          :closable="false"
          show-icon
          class="secret-alert"
        />
      </div>
    </div>

    <div class="card-footer">
      <span class="footer-hint">
        <el-icon :size="12"><CircleCheck /></el-icon>
        {{ confirmed ? (isUpdate ? '该渠道已更新' : '该渠道已录入') : '确认后才会写入数据库' }}
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
import { ArrowRight, Check } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

// 与 views/admin/channel/ChannelList.vue 的 channelTypes 保持同步
const CHANNEL_TYPES = ['火星引擎', 'polo', 'OpenAI', 'Eino OpenAI']

const props = defineProps({
  // 结构化提案 { type, id, name, channelType, priority, weight, retry, models, endpoint, status, mappings }
  // 提案中永远没有密钥字段
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
  channelType: props.proposal.channelType || '',
  endpoint: props.proposal.endpoint || '',
  priority: Number(props.proposal.priority) || 0,
  weight: Number(props.proposal.weight) || 0,
  retry: Number(props.proposal.retry) || 0,
  status: !!props.proposal.status,
  models: [...(props.proposal.models || [])]
})

const mappings = computed(() => props.proposal.mappings || [])

// 创建模式：密钥必填（由管理员手动输入）
const secretInput = ref('')

// 原值对照提示（仅优化模式且字段有变化时展示）
const originalHint = (key) => {
  if (!isUpdate.value || !props.original) return ''
  const before = props.original[key]
  const after = form.value[key]
  if (!before) return ''
  const beforeText = Array.isArray(before) ? before.join(', ') : String(before)
  const afterText = Array.isArray(after) ? after.join(', ') : String(after)
  return beforeText !== afterText ? `原：${beforeText}` : ''
}

// 条件格式化：enableWeb=true, type=app（默认值不展示）
const formatConds = (conds) => {
  if (!conds) return ''
  const parts = []
  if (conds.enableWeb === true) parts.push('enableWeb=true')
  if (conds.type && conds.type !== 'model') parts.push(`type=${conds.type}`)
  return parts.length ? `(${parts.join(', ')})` : ''
}

const canSubmit = computed(() => {
  if (submitting.value) return false
  const name = form.value.name.trim()
  if (name.length < 2 || name.length > 50) return false
  if (!form.value.channelType || !form.value.endpoint.trim()) return false
  if (!form.value.models.length) return false
  // 创建模式必须填写密钥（后端 CreateChannelReq 的 secret 为 required）
  if (!isUpdate.value && !secretInput.value.trim()) return false
  return true
})

const handleConfirm = async () => {
  submitting.value = true
  try {
    let secret = secretInput.value
    if (isUpdate.value) {
      // 渠道更新接口会无条件覆盖密钥：确认时由前端取回现密钥原样并入提交，
      // 密钥全程不经过 AI 与会话记录
      const detail = await defaultApi.apiAdminChannelIdGet(props.proposal.id)
      if (detail.code !== 0 || !detail.data) {
        ElMessage.error(detail.message || '获取渠道当前密钥失败')
        return
      }
      secret = detail.data.secret || ''
    }

    const payload = {
      name: form.value.name.trim(),
      type: form.value.channelType,
      priority: form.value.priority,
      weight: form.value.weight,
      retry: form.value.retry,
      models: form.value.models,
      endpoint: form.value.endpoint.trim(),
      status: form.value.status,
      mappings: mappings.value,
      secret
    }
    const response = isUpdate.value
      ? await defaultApi.apiAdminChannelIdPut(props.proposal.id, payload)
      : await defaultApi.apiAdminChannelPost(payload)
    if (response.code === 0) {
      emit('confirmed')
    } else {
      ElMessage.error(response.message || (isUpdate.value ? '更新失败' : '录入失败'))
    }
  } catch (error) {
    console.error('渠道提案提交失败:', error)
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

.nums {
  flex: 1;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;

  .el-input-number {
    width: 90px;
  }

  .num-label {
    margin-right: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
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

.mappings {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;

  .mapping-item {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
    font-size: 12px;
    padding: 3px 8px;
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
  }

  .mapping-source {
    color: var(--el-color-primary);
    font-weight: 600;
    word-break: break-all;
  }

  .mapping-arrow {
    color: var(--el-text-color-placeholder);
    flex-shrink: 0;
  }

  .mapping-target {
    color: var(--el-text-color-primary);
    word-break: break-all;
  }

  .mapping-conds {
    color: var(--el-text-color-secondary);
    font-size: 11px;
  }
}

.mapping-hint {
  flex: 1;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.secret-hint {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.secret-alert {
  flex: 1;
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

    .mapping-item {
      background: #262727;
    }

    .confirm-btn {
      box-shadow: none;
    }
  }
}
</style>
