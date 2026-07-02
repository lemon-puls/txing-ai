<template>
  <div class="execution-log-panel" v-if="visible">
    <div class="panel-header">
      <div class="header-left">
        <el-icon><Document /></el-icon>
        <span class="title">执行日志</span>
        <el-tag v-if="logs.length > 0" size="small" type="info">{{ logs.length }} 个节点</el-tag>
      </div>
      <div class="header-right">
        <el-button text size="small" @click="$emit('clear')">
          <el-icon><Delete /></el-icon>
          清除
        </el-button>
        <el-button text size="small" @click="$emit('close')">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="panel-body">
      <div v-if="logs.length === 0" class="empty-state">
        <el-icon><InfoFilled /></el-icon>
        <span>暂无执行日志</span>
      </div>

      <div v-else class="log-list">
        <div
          v-for="(log, index) in logs"
          :key="log.nodeId + '-' + index"
          class="log-item"
          :class="{
            'log-running': log.status === 'running',
            'log-completed': log.status === 'completed',
            'log-failed': log.status === 'failed'
          }"
          @click="toggleExpand(log.nodeId)"
        >
          <div class="log-header">
            <div class="log-status-icon">
              <el-icon v-if="log.status === 'running'" class="rotating"><Loading /></el-icon>
              <el-icon v-else-if="log.status === 'completed'" class="success"><CircleCheckFilled /></el-icon>
              <el-icon v-else-if="log.status === 'failed'" class="error"><CircleCloseFilled /></el-icon>
            </div>

            <div class="log-info">
              <div class="log-title">
                <span class="node-label">{{ log.nodeLabel }}</span>
                <el-tag :type="getNodeTypeTag(log.nodeType)" size="small" effect="plain">
                  {{ getNodeTypeLabel(log.nodeType) }}
                </el-tag>
              </div>
              <div class="log-meta">
                <span class="duration"><el-icon><Timer /></el-icon> {{ formatDuration(log.duration) }}</span>
                <span v-if="log.retry > 0" class="retry"><el-icon><RefreshRight /></el-icon> 重试 {{ log.retry }} 次</span>
                <span class="time">{{ formatTime(log.startTime) }}</span>
              </div>
            </div>

            <div class="log-expand-icon">
              <el-icon><ArrowDown v-if="!isExpanded(log.nodeId)" /><ArrowUp v-else /></el-icon>
            </div>
          </div>

          <el-collapse-transition>
            <div v-show="isExpanded(log.nodeId)" class="log-detail">
              <div v-if="log.input" class="detail-section">
                <div class="section-header">
                  <el-icon><Download /></el-icon>
                  <span>输入</span>
                  <el-button text size="small" class="copy-btn" @click.stop="copyContent(log.input)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
                <div class="section-content"><pre>{{ truncateContent(log.input, 500) }}</pre></div>
              </div>

              <div v-if="log.output" class="detail-section">
                <div class="section-header">
                  <el-icon><Upload /></el-icon>
                  <span>输出</span>
                  <el-button text size="small" class="copy-btn" @click.stop="copyContent(log.output)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
                <div class="section-content"><pre>{{ truncateContent(log.output, 500) }}</pre></div>
              </div>

              <div v-if="log.error" class="detail-section error-section">
                <div class="section-header">
                  <el-icon><WarningFilled /></el-icon>
                  <span>错误信息</span>
                </div>
                <div class="section-content error-content"><pre>{{ log.error }}</pre></div>
              </div>
            </div>
          </el-collapse-transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Document, Close, Delete, InfoFilled, Loading, CircleCheckFilled, CircleCloseFilled, Timer, RefreshRight, ArrowDown, ArrowUp, Download, Upload, WarningFilled, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  visible: { type: Boolean, default: false },
  logs: { type: Array, default: () => [] }
})

const emit = defineEmits(['close', 'clear'])
const expandedNodes = ref(new Set())

const toggleExpand = (nodeId) => {
  if (expandedNodes.value.has(nodeId)) expandedNodes.value.delete(nodeId)
  else expandedNodes.value.add(nodeId)
}

const isExpanded = (nodeId) => expandedNodes.value.has(nodeId)

const getNodeTypeTag = (type) => ({ llm: 'primary', tool: 'success', condition: 'warning', code: 'info', http: 'danger' })[type] || 'info'
const getNodeTypeLabel = (type) => ({ llm: 'LLM', tool: '工具', condition: '条件', code: '代码', http: 'HTTP' })[type] || type

const formatDuration = (ms) => { if (!ms && ms !== 0) return '-'; if (ms < 1000) return `${ms}ms`; return `${(ms / 1000).toFixed(2)}s` }
const formatTime = (timestamp) => { if (!timestamp) return '-'; return new Date(timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) }
const truncateContent = (content, maxLen) => (!content ? '' : content.length <= maxLen ? content : content.substring(0, maxLen) + '...\n[内容已截断]')

const copyContent = async (content) => {
  try { await navigator.clipboard.writeText(content); ElMessage.success('已复制到剪贴板') } catch { ElMessage.error('复制失败') }
}
</script>

<style lang="scss" scoped>
$primary-color: #3b82f6;
$primary-light: #60a5fa;
$success-color: #10b981;
$warning-color: #f59e0b;
$danger-color: #ef4444;
$bg-white: #ffffff;
$bg-light: #f8fafc;
$bg-card: #f1f5f9;
$border-color: #e2e8f0;
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.execution-log-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 45vh;
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
      .el-button {
        border-radius: 8px;
        color: $text-secondary;
        &:hover { background: rgba($primary-color, 0.08); color: $primary-color; }
      }
    }
  }

  .panel-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;

    &::-webkit-scrollbar { width: 6px; }
    &::-webkit-scrollbar-thumb { background: $border-color; border-radius: 3px; }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 48px 0;
      color: $text-muted;
      .el-icon { font-size: 40px; margin-bottom: 12px; opacity: 0.5; }
    }

    .log-list { display: flex; flex-direction: column; gap: 10px; }

    .log-item {
      border: 1px solid $border-color;
      border-radius: 12px;
      overflow: hidden;
      transition: all 0.2s ease;
      cursor: pointer;
      background: $bg-white;

      &:hover { border-color: rgba($primary-color, 0.3); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06); }

      &.log-running { border-left: 3px solid #3b82f6; background: rgba(#3b82f6, 0.02); }
      &.log-completed { border-left: 3px solid $success-color; }
      &.log-failed { border-left: 3px solid $danger-color; background: rgba($danger-color, 0.02); }

      .log-header {
        display: flex;
        align-items: center;
        padding: 14px 16px;
        background: $bg-light;

        .log-status-icon {
          margin-right: 14px;
          font-size: 20px;
          .rotating { color: #3b82f6; animation: rotate 1s linear infinite; }
          .success { color: $success-color; }
          .error { color: $danger-color; }
        }

        .log-info {
          flex: 1;
          .log-title { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
          .node-label { font-size: 14px; font-weight: 600; color: $text-primary; }
          .log-meta { display: flex; align-items: center; gap: 16px; font-size: 12px; color: $text-muted; }
        }

        .log-expand-icon { color: $text-muted; font-size: 16px; }
      }

      .log-detail {
        border-top: 1px solid $border-color;
        background: $bg-card;

        .detail-section {
          padding: 14px 16px;
          border-bottom: 1px solid rgba($border-color, 0.5);
          &:last-child { border-bottom: none; }

          .section-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 10px;
            font-size: 12px;
            font-weight: 500;
            color: $text-secondary;
            .el-icon { font-size: 14px; }
            .copy-btn {
              margin-left: auto;
              padding: 4px;
              color: $text-muted;
              border-radius: 6px;
              &:hover { color: $primary-light; background: rgba($primary-color, 0.08); }
            }
          }

          .section-content {
            background: $bg-white;
            border: 1px solid $border-color;
            border-radius: 8px;
            padding: 12px;
            max-height: 150px;
            overflow-y: auto;
            pre { margin: 0; font-family: monospace; font-size: 12px; line-height: 1.6; color: $text-secondary; white-space: pre-wrap; word-break: break-all; }
          }

          &.error-section {
            .section-header { color: $danger-color; }
            .error-content { background: rgba($danger-color, 0.05); border-color: rgba($danger-color, 0.2); pre { color: $danger-color; } }
          }
        }
      }
    }
  }
}

@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
