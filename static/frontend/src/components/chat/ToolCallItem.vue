<template>
  <div class="tool-call-item">
    <!-- 工具行：状态图标 + 名称 + 展开箭头 -->
    <div
      class="tool-row"
      :class="toolCall.status"
      role="button"
      :aria-expanded="expanded"
      @click="expanded = !expanded"
    >
      <span class="tool-status-icon" :class="toolCall.status">
        <el-icon v-if="toolCall.status === 'running'" :size="12" class="spin"><Loading /></el-icon>
        <el-icon v-else-if="toolCall.status === 'completed'" :size="12"><Check /></el-icon>
        <el-icon v-else-if="toolCall.status === 'failed'" :size="12"><Close /></el-icon>
        <el-icon v-else :size="12"><Tools /></el-icon>
      </span>
      <span class="tool-name">{{ toolCall.name }}</span>
      <el-icon class="tool-expand-arrow" :size="12" :class="{ expanded }"><ArrowDown /></el-icon>
    </div>

    <!-- 展开详情：调用参数 / 执行结果（文本插值渲染，外部内容不可信，禁用 v-html） -->
    <div v-if="expanded" class="tool-detail">
      <div class="detail-section">
        <div class="detail-title">调用参数</div>
        <pre v-if="formattedArgs" class="detail-pre">{{ formattedArgs }}</pre>
        <div v-else class="detail-empty">暂无参数数据</div>
      </div>
      <div class="detail-section">
        <div class="detail-title">执行结果</div>
        <pre v-if="formattedResult" class="detail-pre">{{ formattedResult }}</pre>
        <div v-else class="detail-empty">{{ toolCall.status === 'running' ? '执行中…' : '暂无结果数据' }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ArrowDown, Check, Close, Loading, Tools } from '@element-plus/icons-vue'

const props = defineProps({
  // 归一化结构：{ id, name, status, args, result }
  toolCall: { type: Object, required: true }
})

const expanded = ref(false)

// 前端显示兜底上限（防旧数据/异常路径的大文本卡渲染）
const MAX_PREVIEW = 50000
const cap = (s) => {
  return s.length > MAX_PREVIEW
    ? s.slice(0, MAX_PREVIEW) + `\n…(内容过长，仅显示前 ${MAX_PREVIEW} 字符)`
    : s
}

// JSON 尝试格式化，失败回退原文
const formatJson = (raw) => {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

// computed 惰性求值：折叠时详情 v-if 不渲染、不读取，不做无谓格式化
const formattedArgs = computed(() => cap(formatJson(props.toolCall.args)))
const formattedResult = computed(() => cap(formatJson(props.toolCall.result)))
</script>

<style lang="scss" scoped>
$success: #10b981;
$warning: #f59e0b;
$danger: #ef4444;
$primary: #6366f1;
$info: #94a3b8;

.tool-call-item {
  min-width: 0;
}

.tool-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;

  &:hover {
    background: var(--el-fill-color-light);
  }
}

.tool-status-icon {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #fff;
  background: var(--el-fill-color);
  border: 1px solid var(--el-border-color-light);

  &.running {
    background: $primary;
    border-color: $primary;
  }

  &.completed {
    background: $success;
    border-color: $success;
  }

  &.failed {
    background: $danger;
    border-color: $danger;
  }
}

.tool-name {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--el-text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tool-expand-arrow {
  flex-shrink: 0;
  color: var(--el-text-color-secondary);
  transition: transform 0.2s;

  &.expanded {
    transform: rotate(180deg);
  }
}

.spin {
  animation: tool-spin 1s linear infinite;
}

@keyframes tool-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.tool-detail {
  margin: 2px 0 6px 26px;
  padding: 8px 10px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.detail-pre {
  margin: 0;
  padding: 6px 8px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-regular);
  max-height: 240px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.detail-empty {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  padding: 2px 0;
}
</style>
