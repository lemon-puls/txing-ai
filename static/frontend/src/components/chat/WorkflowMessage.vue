<template>
  <div class="workflow-message">
    <!-- 应用执行卡片 -->
    <div class="workflow-card" :class="workflow?.status">
      <div class="workflow-header" @click="expanded = !expanded">
        <div class="workflow-title">
          <el-icon class="workflow-icon"><Share /></el-icon>
          <span>{{ appName || '应用执行' }}</span>
          <el-tag size="small" :type="statusType" effect="plain" round>
            {{ statusLabel }}
          </el-tag>
        </div>
        <el-icon class="expand-arrow" :class="{ expanded }">
          <ArrowDown />
        </el-icon>
      </div>

      <!-- 节点进度 -->
      <transition name="slide">
        <div v-show="expanded && nodeLogs.length > 0" class="workflow-nodes">
          <div
            v-for="log in nodeLogs"
            :key="log.nodeId"
            class="node-item"
            :class="log.status"
          >
            <div class="node-dot" :class="log.status"></div>
            <div class="node-info">
              <span class="node-label">{{ log.label || log.type }}</span>
              <span v-if="log.toolName" class="node-tool">
                <el-tag size="small" type="warning" effect="plain" round>{{ log.toolName }}</el-tag>
              </span>
            </div>
            <span class="node-status">
              <el-icon v-if="log.status === 'completed'"><CircleCheck /></el-icon>
              <el-icon v-else-if="log.status === 'running'" class="spin"><Loading /></el-icon>
              <el-icon v-else-if="log.status === 'failed'"><CircleClose /></el-icon>
            </span>
          </div>
        </div>
      </transition>
    </div>

    <!-- 产物下载 -->
    <div v-if="artifacts && artifacts.length > 0" class="artifacts">
      <div v-for="(file, idx) in artifacts" :key="idx" class="artifact-item">
        <div class="artifact-icon" :class="file.category">
          <el-icon :size="16"><Document /></el-icon>
        </div>
        <span class="artifact-name">{{ file.name }}</span>
        <el-button
          type="primary"
          size="small"
          text
          @click="downloadFile(file.url, file.name)"
        >
          下载
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Share, ArrowDown, CircleCheck, CircleClose, Loading, Document } from '@element-plus/icons-vue'
import { getAuthHeaders } from '@/api/auth'

const props = defineProps({
  appName: { type: String, default: '' },
  workflow: { type: Object, default: null },
  artifacts: { type: Array, default: () => [] }
})

const expanded = ref(true)
const nodeLogs = ref([])
const nodeMap = ref(new Map())

const statusType = computed(() => {
  const s = props.workflow?.status
  if (s === 'completed') return 'success'
  if (s === 'failed') return 'danger'
  if (s === 'running') return 'warning'
  return 'info'
})

const statusLabel = computed(() => {
  const s = props.workflow?.status
  if (s === 'completed') return '已完成'
  if (s === 'failed') return '失败'
  if (s === 'running') return '执行中'
  return '等待中'
})

// 监听 workflow 变化，更新节点日志
watch(() => props.workflow, (w) => {
  if (!w || !w.nodeId) return

  const existing = nodeMap.value.get(w.nodeId)
  if (existing) {
    existing.status = w.nodeStatus || existing.status
    if (w.toolName) existing.toolName = w.toolName
  } else {
    const log = {
      nodeId: w.nodeId,
      type: w.nodeType,
      label: w.nodeLabel,
      status: w.nodeStatus || 'running',
      toolName: w.toolName || ''
    }
    nodeMap.value.set(w.nodeId, log)
    nodeLogs.value.push(log)
  }
}, { deep: true })

const downloadFile = async (url, name) => {
  try {
    const response = await fetch(url, { headers: getAuthHeaders() })
    if (!response.ok) return
    const blob = await response.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch (e) {
    console.error('下载失败:', e)
  }
}
</script>

<style lang="scss" scoped>
.workflow-message {
  max-width: 100%;
}

.workflow-card {
  background: var(--el-fill-color-lighter, #fafafa);
  border: 1px solid var(--el-border-color-light, #dcdfe6);
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 8px;

  &.completed { border-left: 3px solid var(--el-color-success); }
  &.failed { border-left: 3px solid var(--el-color-danger); }
  &.running { border-left: 3px solid #1976d2; }
}

.workflow-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  cursor: pointer;
  transition: background 0.15s;

  &:hover { background: var(--el-fill-color-light, #f5f7fa); }
}

.workflow-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.workflow-icon {
  color: #1976d2;
}

.expand-arrow {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  transition: transform 0.25s;

  &.expanded { transform: rotate(180deg); }
}

.workflow-nodes {
  padding: 0 14px 10px;
  border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);
}

.node-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 12px;

  .node-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--el-border-color, #dcdfe6);
    flex-shrink: 0;

    &.running { background: #1976d2; animation: pulse 1.2s infinite; }
    &.completed { background: var(--el-color-success); }
    &.failed { background: var(--el-color-danger); }
  }

  .node-info {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .node-label {
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .node-status {
    flex-shrink: 0;
    font-size: 14px;
    color: var(--el-text-color-secondary);

    .spin { animation: spin 1s linear infinite; }
  }
}

.artifacts {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}

.artifact-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter, #fafafa);
  border: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  border-radius: 8px;

  .artifact-icon {
    width: 28px;
    height: 28px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    &.pdf { background: rgba(245, 108, 108, 0.1); color: #f56c6c; }
    &.markdown { background: rgba(25, 118, 210, 0.1); color: #1976d2; }
    &.image { background: rgba(103, 194, 58, 0.1); color: #67c23a; }
  }

  .artifact-name {
    flex: 1;
    font-size: 12px;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

.slide-enter-active, .slide-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.slide-enter-from, .slide-leave-to {
  max-height: 0;
  opacity: 0;
}
.slide-enter-to, .slide-leave-from {
  max-height: 500px;
  opacity: 1;
}
</style>
