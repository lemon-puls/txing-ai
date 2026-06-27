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
          <!-- 日志头部 -->
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
                <span class="duration">
                  <el-icon><Timer /></el-icon>
                  {{ formatDuration(log.duration) }}
                </span>
                <span v-if="log.retry > 0" class="retry">
                  <el-icon><RefreshRight /></el-icon>
                  重试 {{ log.retry }} 次
                </span>
                <span class="time">{{ formatTime(log.startTime) }}</span>
              </div>
            </div>

            <div class="log-expand-icon">
              <el-icon>
                <ArrowDown v-if="!isExpanded(log.nodeId)" />
                <ArrowUp v-else />
              </el-icon>
            </div>
          </div>

          <!-- 日志详情（可展开） -->
          <el-collapse-transition>
            <div v-show="isExpanded(log.nodeId)" class="log-detail">
              <!-- 输入 -->
              <div v-if="log.input" class="detail-section">
                <div class="section-header">
                  <el-icon><Download /></el-icon>
                  <span>输入</span>
                  <el-button
                    text
                    size="small"
                    class="copy-btn"
                    @click.stop="copyContent(log.input)"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
                <div class="section-content">
                  <pre>{{ truncateContent(log.input, 500) }}</pre>
                </div>
              </div>

              <!-- 输出 -->
              <div v-if="log.output" class="detail-section">
                <div class="section-header">
                  <el-icon><Upload /></el-icon>
                  <span>输出</span>
                  <el-button
                    text
                    size="small"
                    class="copy-btn"
                    @click.stop="copyContent(log.output)"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
                <div class="section-content">
                  <pre>{{ truncateContent(log.output, 500) }}</pre>
                </div>
              </div>

              <!-- 错误 -->
              <div v-if="log.error" class="detail-section error-section">
                <div class="section-header">
                  <el-icon><WarningFilled /></el-icon>
                  <span>错误信息</span>
                </div>
                <div class="section-content error-content">
                  <pre>{{ log.error }}</pre>
                </div>
              </div>
            </div>
          </el-collapse-transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  Document, Close, Delete, InfoFilled, Loading,
  CircleCheckFilled, CircleCloseFilled, Timer, RefreshRight,
  ArrowDown, ArrowUp, Download, Upload, WarningFilled, CopyDocument
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  logs: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'clear'])

// 展开状态管理
const expandedNodes = ref(new Set())

const toggleExpand = (nodeId) => {
  if (expandedNodes.value.has(nodeId)) {
    expandedNodes.value.delete(nodeId)
  } else {
    expandedNodes.value.add(nodeId)
  }
}

const isExpanded = (nodeId) => {
  return expandedNodes.value.has(nodeId)
}

// 节点类型标签
const getNodeTypeTag = (type) => {
  const map = {
    llm: 'primary',
    tool: 'success',
    condition: 'warning',
    code: 'info',
    http: 'danger',
    subworkflow: ''
  }
  return map[type] || 'info'
}

const getNodeTypeLabel = (type) => {
  const map = {
    llm: 'LLM',
    tool: '工具',
    condition: '条件',
    code: '代码',
    http: 'HTTP',
    subworkflow: '子流程'
  }
  return map[type] || type
}

// 格式化耗时
const formatDuration = (ms) => {
  if (!ms && ms !== 0) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 截断内容
const truncateContent = (content, maxLen) => {
  if (!content) return ''
  if (content.length <= maxLen) return content
  return content.substring(0, maxLen) + '...\n[内容已截断]'
}

// 复制内容
const copyContent = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}
</script>

<style lang="scss" scoped>
.execution-log-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 45vh;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  z-index: 100;

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid #e4e7ed;
    background: #fafafa;

    .header-left {
      display: flex;
      align-items: center;
      gap: 8px;

      .el-icon {
        color: #409eff;
        font-size: 16px;
      }

      .title {
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 4px;
    }
  }

  .panel-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px;

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 40px 0;
      color: #909399;

      .el-icon {
        font-size: 32px;
        margin-bottom: 8px;
      }
    }

    .log-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .log-item {
      border: 1px solid #e4e7ed;
      border-radius: 8px;
      overflow: hidden;
      transition: all 0.2s ease;
      cursor: pointer;

      &:hover {
        border-color: #c0c4cc;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
      }

      &.log-running {
        border-left: 3px solid #409eff;
      }

      &.log-completed {
        border-left: 3px solid #67c23a;
      }

      &.log-failed {
        border-left: 3px solid #f56c6c;
      }

      .log-header {
        display: flex;
        align-items: center;
        padding: 12px 16px;
        background: #fafafa;

        .log-status-icon {
          margin-right: 12px;
          font-size: 18px;

          .rotating {
            color: #409eff;
            animation: rotate 1s linear infinite;
          }

          .success {
            color: #67c23a;
          }

          .error {
            color: #f56c6c;
          }
        }

        .log-info {
          flex: 1;

          .log-title {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 4px;

            .node-label {
              font-size: 14px;
              font-weight: 500;
              color: #303133;
            }
          }

          .log-meta {
            display: flex;
            align-items: center;
            gap: 12px;
            font-size: 12px;
            color: #909399;

            .duration,
            .retry {
              display: flex;
              align-items: center;
              gap: 4px;
            }
          }
        }

        .log-expand-icon {
          color: #909399;
          font-size: 14px;
        }
      }

      .log-detail {
        border-top: 1px solid #e4e7ed;
        background: #fff;

        .detail-section {
          padding: 12px 16px;
          border-bottom: 1px solid #f0f0f0;

          &:last-child {
            border-bottom: none;
          }

          .section-header {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-bottom: 8px;
            font-size: 12px;
            font-weight: 500;
            color: #606266;

            .el-icon {
              font-size: 14px;
            }

            .copy-btn {
              margin-left: auto;
              padding: 2px;
              color: #909399;

              &:hover {
                color: #409eff;
              }
            }
          }

          .section-content {
            background: #f5f7fa;
            border-radius: 6px;
            padding: 12px;
            max-height: 150px;
            overflow-y: auto;

            pre {
              margin: 0;
              font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
              font-size: 12px;
              line-height: 1.6;
              color: #303133;
              white-space: pre-wrap;
              word-break: break-all;
            }
          }

          &.error-section {
            .section-header {
              color: #f56c6c;
            }

            .error-content {
              background: #fef0f0;

              pre {
                color: #f56c6c;
              }
            }
          }
        }
      }
    }
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
