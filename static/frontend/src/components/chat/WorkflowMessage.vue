<template>
  <div class="workflow-message">
    <div class="workflow-card" :class="workflow?.status">
      <div class="card-status-line" :class="workflow?.status"></div>
      <div class="workflow-header" @click="expanded = !expanded">
        <div class="workflow-title">
          <div class="app-icon">
            <el-icon :size="14"><Share /></el-icon>
          </div>
          <span class="app-name">{{ appName || '应用执行' }}</span>
          <span class="status-text" :class="workflow?.status">{{ statusLabel }}</span>
        </div>
        <el-icon v-if="nodeLogsData.length > 0" class="expand-arrow" :class="{ expanded }">
          <ArrowDown />
        </el-icon>
      </div>
      <transition name="slide">
        <div v-show="expanded && nodeLogsData.length > 0" class="workflow-nodes">
          <div
            v-for="(log, index) in nodeLogsData"
            :key="log.nodeId"
            class="node-item"
            :class="log.status"
            :style="{ animationDelay: index * 60 + 'ms' }"
          >
            <div class="node-indicator">
              <el-icon v-if="log.status === 'completed'" :size="12"><Check /></el-icon>
              <el-icon v-else-if="log.status === 'running'" :size="12" class="spin"><Loading /></el-icon>
              <el-icon v-else-if="log.status === 'failed'" :size="12"><Close /></el-icon>
              <span v-else class="dot-indicator"></span>
            </div>
            <div class="node-body">
              <span class="node-label">{{ log.label || log.type }}</span>
              <div v-if="log.toolCalls && log.toolCalls.length > 0" class="tool-calls">
                <div
                  v-for="(tc, tcIdx) in log.toolCalls"
                  :key="tcIdx"
                  class="tool-chip"
                  :class="tc.status"
                >
                  <el-icon :size="10"><Tools /></el-icon>
                  <span>{{ tc.name }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </div>
    <div v-if="artifacts && artifacts.length > 0" class="artifacts">
      <div v-for="(file, idx) in artifacts" :key="idx" class="artifact-item" @click="downloadFile(file.url, file.name)">
        <div class="artifact-icon" :class="file.category">
          <el-icon :size="14"><Document /></el-icon>
        </div>
        <span class="artifact-name">{{ file.name }}</span>
        <el-icon class="download-icon" :size="14"><Download /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Share, ArrowDown, Check, Close, Loading, Document, Tools, Download } from '@element-plus/icons-vue'
import { getAuthHeaders } from '@/api/auth'

const props = defineProps({
  appName: { type: String, default: '' },
  workflow: { type: Object, default: null },
  artifacts: { type: Array, default: () => [] },
  nodeLogs: { type: Array, default: () => [] }
})

const expanded = ref(true)
const nodeLogsData = ref([])
const nodeMap = ref(new Map())

onMounted(() => {
  if (props.nodeLogs && props.nodeLogs.length > 0) {
    props.nodeLogs.forEach(log => {
      const entry = {
        nodeId: log.nodeId,
        type: log.type,
        label: log.label,
        status: log.status,
        toolCalls: log.toolCalls || []
      }
      nodeMap.value.set(log.nodeId, entry)
      nodeLogsData.value.push(entry)
    })
  }
})

const statusLabel = computed(() => {
  const s = props.workflow?.status
  if (s === 'completed') return '已完成'
  if (s === 'failed') return '失败'
  if (s === 'running') return '执行中'
  return '等待中'
})

watch(() => props.workflow, (w) => {
  if (!w || !w.nodeId) return
  const existing = nodeMap.value.get(w.nodeId)
  if (existing) {
    existing.status = w.nodeStatus || existing.status
    if (w.toolName) {
      const lastTc = existing.toolCalls[existing.toolCalls.length - 1]
      if (lastTc && lastTc.name === w.toolName) {
        lastTc.status = w.toolStatus || lastTc.status
      } else {
        existing.toolCalls.push({ name: w.toolName, status: w.toolStatus || 'running' })
      }
    }
    if (w.nodeStatus === 'completed' || w.nodeStatus === 'failed') {
      existing.toolCalls.forEach(tc => {
        if (tc.status === 'running') {
          tc.status = w.nodeStatus
        }
      })
    }
  } else {
    const log = {
      nodeId: w.nodeId,
      type: w.nodeType,
      label: w.nodeLabel,
      status: w.nodeStatus || 'running',
      toolCalls: w.toolName ? [{ name: w.toolName, status: w.toolStatus || 'running' }] : []
    }
    nodeMap.value.set(w.nodeId, log)
    nodeLogsData.value.push(log)
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
$success: #10b981;
$warning: #f59e0b;
$danger: #ef4444;
$primary: #6366f1;
$info: #94a3b8;

.workflow-message {
  max-width: 100%;
  margin: 4px 0;
}

.workflow-card {
  position: relative;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  overflow: hidden;
  width: 100%;
  min-width: 100%;
  transition: border-color 0.2s;

  &:hover {
    border-color: var(--el-border-color);
  }
}

.card-status-line {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;

  &.completed { background: $success; }
  &.failed { background: $danger; }
  &.running {
    background: $primary;
    animation: status-blink 1.5s ease-in-out infinite;
  }
  &.pending { background: $info; }
}

@keyframes status-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.workflow-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;

  &:hover {
    background: var(--el-fill-color-light);
  }
}

.workflow-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: nowrap;
  overflow: hidden;
}

.app-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: linear-gradient(135deg, $primary, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.app-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  letter-spacing: 0.2px;
  flex-shrink: 0;
}

.status-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);

  &.completed { color: $success; }
  &.failed { color: $danger; }
  &.running {
    color: $primary;
    animation: text-pulse 1.5s ease-in-out infinite;
  }
  &.pending { color: $info; }

  &::before {
    content: '·';
    margin-right: 6px;
    opacity: 0.5;
  }
}

@keyframes text-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.expand-arrow {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  transition: transform 0.2s;

  &.expanded {
    transform: rotate(180deg);
  }
}

.workflow-nodes {
  padding: 4px 16px 12px;
  border-top: 1px solid var(--el-border-color-extra-light);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 0;
  animation: node-fade-in 0.3s ease-out backwards;
  transition: opacity 0.2s;

  &.pending {
    opacity: 0.5;
  }
}

@keyframes node-fade-in {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.node-indicator {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 2px;
  transition: all 0.2s;

  .el-icon {
    color: #fff;
  }

  .dot-indicator {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--el-border-color);
  }

  &.pending {
    background: var(--el-fill-color);
    border: 1px solid var(--el-border-color-light);
  }

  &.running {
    background: $primary;
    animation: indicator-pulse 1.5s ease-in-out infinite;
  }

  &.completed {
    background: $success;
  }

  &.failed {
    background: $danger;
  }
}

@keyframes indicator-pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba($primary, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba($primary, 0);
  }
}

.node-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.node-label {
  font-size: 13px;
  color: var(--el-text-color-primary);
  line-height: 1.4;
}

.tool-calls {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tool-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  border: 1px solid var(--el-border-color-lighter);
  transition: all 0.2s;

  .el-icon {
    opacity: 0.7;
  }

  &.running {
    background: rgba($warning, 0.1);
    color: $warning;
    border-color: rgba($warning, 0.2);

    .el-icon {
      animation: spin 1s linear infinite;
      opacity: 1;
    }
  }

  &.completed {
    background: rgba($success, 0.08);
    color: $success;
    border-color: rgba($success, 0.15);
  }

  &.failed {
    background: rgba($danger, 0.08);
    color: $danger;
    border-color: rgba($danger, 0.15);
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.artifacts {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.artifact-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;

  &:hover {
    background: var(--el-fill-color);
    border-color: $primary;
    transform: translateY(-1px);
  }

  .artifact-icon {
    width: 18px;
    height: 18px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    &.pdf { background: rgba($danger, 0.12); color: $danger; }
    &.markdown { background: rgba($primary, 0.12); color: $primary; }
    &.image { background: rgba($success, 0.12); color: $success; }
    &.default { background: var(--el-fill-color); color: var(--el-text-color-secondary); }
  }

  .artifact-name {
    font-size: 12px;
    color: var(--el-text-color-primary);
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .download-icon {
    color: var(--el-text-color-secondary);
    transition: color 0.15s;
  }

  &:hover .download-icon {
    color: $primary;
  }
}

.slide-enter-active, .slide-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.slide-enter-from, .slide-leave-to {
  max-height: 0;
  opacity: 0;
}
.slide-enter-to, .slide-leave-from {
  max-height: 600px;
  opacity: 1;
}
</style>
