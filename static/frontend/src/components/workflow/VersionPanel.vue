<template>
  <div class="version-panel" v-if="visible">
    <div class="panel-header">
      <div class="header-left">
        <el-icon><Clock /></el-icon>
        <span class="title">版本管理</span>
        <el-tag v-if="versions.length > 0" size="small" type="info">{{ versions.length }} 个版本</el-tag>
      </div>
      <div class="header-right">
        <el-button type="primary" size="small" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建版本
        </el-button>
        <el-button text size="small" @click="$emit('close')">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="panel-body">
      <div v-if="loading" class="loading-state">
        <el-icon class="rotating"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <div v-else-if="versions.length === 0" class="empty-state">
        <el-icon><InfoFilled /></el-icon>
        <span>暂无版本记录</span>
        <p class="empty-hint">点击右上角"创建版本"保存当前工作流快照</p>
      </div>

      <div v-else class="version-list">
        <div v-for="ver in versions" :key="ver.id" class="version-item" :class="{ 'is-published': ver.isPublished, 'is-current': ver.version === currentVersion }">
          <div class="version-header">
            <div class="version-info">
              <div class="version-title">
                <span class="version-name">{{ ver.name }}</span>
                <el-tag v-if="ver.isPublished" type="success" size="small" effect="dark">已发布</el-tag>
                <el-tag v-if="ver.version === currentVersion" type="primary" size="small" effect="plain">当前</el-tag>
              </div>
              <div class="version-meta">
                <span class="version-number">v{{ ver.version }}</span>
                <span class="version-time">{{ formatTime(ver.createTime) }}</span>
              </div>
              <div v-if="ver.description" class="version-desc">{{ ver.description }}</div>
              <div v-if="ver.changeLog" class="version-changelog">
                <el-icon><Document /></el-icon>
                {{ ver.changeLog }}
              </div>
            </div>
          </div>

          <div class="version-actions">
            <el-button text size="small" type="primary" @click="$emit('preview', ver)">
              <el-icon><View /></el-icon>
              查看
            </el-button>
            <el-button v-if="!ver.isPublished" text size="small" type="success" @click="handlePublish(ver)">
              <el-icon><Upload /></el-icon>
              发布
            </el-button>
            <el-button text size="small" type="warning" @click="handleRollback(ver)">
              <el-icon><RefreshLeft /></el-icon>
              回滚
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showCreateDialog" title="创建版本" width="440px" :close-on-click-modal="false" class="version-dialog">
      <el-form label-position="top" :model="createForm" :rules="createRules" ref="createFormRef">
        <el-form-item label="版本名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如：v1.0 正式版" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="版本描述">
          <el-input v-model="createForm.description" type="textarea" :rows="2" placeholder="简要描述此版本的用途（可选）" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="变更日志">
          <el-input v-model="createForm.changeLog" type="textarea" :rows="3" placeholder="记录本次修改的内容（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Clock, Close, Plus, InfoFilled, Loading, Document, View, Upload, RefreshLeft } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps({
  visible: { type: Boolean, default: false },
  versions: { type: Array, default: () => [] },
  currentVersion: { type: Number, default: 0 },
  loading: { type: Boolean, default: false }
})

const emit = defineEmits(['close', 'create', 'publish', 'rollback', 'preview'])

const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref(null)
const createForm = ref({ name: '', description: '', changeLog: '' })
const createRules = {
  name: [{ required: true, message: '请输入版本名称', trigger: 'blur' }]
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate((valid) => {
    if (!valid) return
    emit('create', { ...createForm.value })
    showCreateDialog.value = false
    createForm.value = { name: '', description: '', changeLog: '' }
  })
}

const handlePublish = async (ver) => {
  try {
    await ElMessageBox.confirm(
      `确定发布版本 v${ver.version}（${ver.name}）？发布后将成为正式版本。`,
      '发布确认',
      { confirmButtonText: '发布', cancelButtonText: '取消', type: 'warning' }
    )
    emit('publish', ver)
  } catch {}
}

const handleRollback = async (ver) => {
  try {
    await ElMessageBox.confirm(
      `确定回滚到版本 v${ver.version}（${ver.name}）？当前未保存的修改将被覆盖。`,
      '回滚确认',
      { confirmButtonText: '回滚', cancelButtonText: '取消', type: 'warning' }
    )
    emit('rollback', ver)
  } catch {}
}

const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  })
}
</script>

<style lang="scss" scoped>
$primary-color: #3b82f6;
$primary-light: #60a5fa;
$success-color: #10b981;
$warning-color: #f59e0b;
$bg-white: #ffffff;
$bg-light: #f8fafc;
$bg-card: #f1f5f9;
$border-color: #e2e8f0;
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.version-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 50vh;
  background: $bg-white;
  border: 1px solid $border-color;
  border-top: 3px solid $primary-color;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  z-index: 100;
  border-radius: 16px 16px 0 0;

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px;
    border-bottom: 1px solid $border-color;
    background: $bg-light;

    .header-left {
      display: flex;
      align-items: center;
      gap: 10px;

      .el-icon { color: $primary-color; font-size: 18px; }
      .title { font-size: 14px; font-weight: 600; color: $text-primary; }
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .panel-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;

    &::-webkit-scrollbar { width: 6px; }
    &::-webkit-scrollbar-thumb { background: $border-color; border-radius: 3px; }

    .loading-state, .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 48px 0;
      color: $text-muted;

      .el-icon { font-size: 40px; margin-bottom: 12px; opacity: 0.5; }
      .empty-hint { font-size: 12px; color: $text-muted; margin-top: 4px; }
    }

    .version-list { display: flex; flex-direction: column; gap: 12px; }

    .version-item {
      border: 1px solid $border-color;
      border-radius: 12px;
      overflow: hidden;
      transition: all 0.2s ease;
      background: $bg-white;

      &:hover {
        border-color: rgba($primary-color, 0.3);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
      }

      &.is-published { border-left: 3px solid $success-color; }
      &.is-current { border-left: 3px solid $primary-color; }

      .version-header {
        padding: 14px 16px;
        background: $bg-light;

        .version-info {
          .version-title {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 6px;

            .version-name { font-size: 14px; font-weight: 600; color: $text-primary; }
          }

          .version-meta {
            display: flex;
            align-items: center;
            gap: 12px;
            font-size: 12px;
            color: $text-muted;

            .version-number { font-weight: 500; color: $text-secondary; }
          }

          .version-desc {
            font-size: 12px;
            color: $text-secondary;
            margin-top: 6px;
          }

          .version-changelog {
            display: flex;
            align-items: flex-start;
            gap: 6px;
            font-size: 12px;
            color: $text-muted;
            margin-top: 6px;
            padding: 6px 10px;
            background: $bg-card;
            border-radius: 6px;

            .el-icon { margin-top: 1px; flex-shrink: 0; }
          }
        }
      }

      .version-actions {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 16px;
        border-top: 1px solid $border-color;
        background: $bg-white;
      }
    }
  }
}

@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.rotating { animation: rotate 1s linear infinite; }
</style>
