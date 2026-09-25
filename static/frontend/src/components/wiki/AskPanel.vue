<template>
  <div class="ask-panel">
    <div class="ask-card">
      <!-- 标题区 -->
      <div class="ask-header">
        <div class="ask-badge">
          <el-icon><ChatDotRound /></el-icon>
          <span>AI 问答</span>
        </div>
        <h3 class="ask-title">关于我，想问点什么？</h3>
        <p class="ask-subtitle">由我的个人知识库驱动，AI 会现场查阅资料后作答</p>
        <!-- 留存声明（D10：问答全量留存，须明示） -->
        <div class="ask-notice">
          <el-icon><InfoFilled /></el-icon>
          <span>对话将被记录用于改进知识库质量，请勿输入敏感信息</span>
        </div>
      </div>

      <!-- 消息区 -->
      <div class="ask-messages" ref="messagesEl">
        <!-- 欢迎态：推荐问题 -->
        <div v-if="messages.length === 0" class="ask-welcome">
          <p class="welcome-hint">试试这些问题：</p>
          <div class="suggestion-list">
            <button
              v-for="q in suggestions"
              :key="q"
              class="suggestion-chip"
              :disabled="sending"
              @click="sendMessage(q)"
            >{{ q }}</button>
          </div>
        </div>

        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="ask-msg"
          :class="`ask-msg--${msg.role}`"
        >
          <div class="msg-avatar" v-if="msg.role === 'assistant'">
            <el-icon><Cpu /></el-icon>
          </div>
          <div class="msg-body">
            <!-- 工具调用进度 -->
            <div v-if="msg.toolCalls.length" class="msg-tools">
              <div
                v-for="tc in msg.toolCalls"
                :key="tc.id"
                class="tool-chip"
                :class="{ 'is-running': tc.status === 'running' }"
              >
                <el-icon v-if="tc.status === 'running'" class="is-loading"><Loading /></el-icon>
                <el-icon v-else><CircleCheck /></el-icon>
                <span>{{ toolLabel(tc) }}</span>
              </div>
            </div>
            <!-- 正文（markdown 渲染；不挂 markdown-body 类，chat 页全局引入的
                 github-markdown-dark 会将其背景染成 #0d1117） -->
            <div
              v-if="msg.content"
              class="msg-content"
              v-html="renderMarkdown(msg.content)"
            ></div>
            <!-- 状态行 -->
            <div v-if="msg.streaming && !msg.content && !msg.toolCalls.length" class="msg-typing">
              <span></span><span></span><span></span>
            </div>
            <div v-if="msg.error" class="msg-error">
              <el-icon><WarningFilled /></el-icon>
              <span>{{ msg.error }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="ask-input">
        <el-input
          v-model="input"
          placeholder="输入你的问题…"
          :disabled="sending"
          maxlength="4000"
          @keydown.enter.exact.prevent="handleSend"
          clearable
        />
        <el-button
          v-if="!sending"
          class="send-btn"
          type="primary"
          :disabled="!input.trim()"
          @click="handleSend"
        >
          <el-icon><Promotion /></el-icon>
        </el-button>
        <el-button v-else class="send-btn stop-btn" @click="stop">
          <el-icon><VideoPause /></el-icon>
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup name="AskPanel">
import { ref, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ChatDotRound, InfoFilled, Cpu, Loading, CircleCheck,
  WarningFilled, Promotion, VideoPause
} from '@element-plus/icons-vue'
import { fetchSSEWithAuth } from '@/api/sseRequest'
import { renderWikiMarkdown } from '@/utils/wikiMarkdown'

const HISTORY_WINDOW = 10 // 多轮上下文窗口（与服务端约束一致）

const suggestions = [
  '他最熟悉的技术栈是什么？',
  '介绍下他做过的最有挑战的项目',
  '他在 AI 应用方面有什么实践？'
]

const TOOL_LABELS = {
  wiki_read_page: '查阅',
  wiki_search_pages: '搜索'
}

const input = ref('')
const messages = ref([])
const sending = ref(false)
const messagesEl = ref(null)
let controller = null
// 会话标识：前端生成，服务端用于问答留存聚合（D10）
const sessionId = crypto.randomUUID()

const renderMarkdown = (text) => {
  try {
    return renderWikiMarkdown(text)
  } catch {
    return text
  }
}

const toolLabel = (tc) => {
  const label = TOOL_LABELS[tc.name] || '检索'
  let detail = ''
  try {
    const args = typeof tc.args === 'string' ? JSON.parse(tc.args || '{}') : (tc.args || {})
    detail = args.slug || args.keyword || ''
  } catch { /* 参数展示失败不影响主流程 */ }
  return detail ? `${label}：${detail}` : `${label}资料中`
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

const currentAssistant = () => {
  let msg = messages.value[messages.value.length - 1]
  if (!msg || msg.role !== 'assistant' || !msg.streaming) {
    msg = { role: 'assistant', content: '', toolCalls: [], error: '', streaming: true }
    messages.value.push(msg)
  }
  return msg
}

const finalizeAssistant = () => {
  const msg = messages.value[messages.value.length - 1]
  if (msg && msg.role === 'assistant') {
    msg.streaming = false
    if (!msg.content && !msg.toolCalls.length && !msg.error) {
      msg.content = '（未收到回答，请稍后重试）'
    }
  }
  sending.value = false
  controller = null
}

const parseFrame = (raw) => {
  // SSE 帧形如 "data: {...}"；禁用/限流等场景返回纯 JSON，也在此兜底解析
  let handled = false
  for (const line of raw.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed.startsWith('data:')) continue
    handled = true
    const payload = trimmed.slice(5).trim()
    if (!payload || payload === '[DONE]') continue
    try {
      handleFrame(JSON.parse(payload))
    } catch (e) {
      console.warn('知识库问答：无法解析 SSE 帧', payload, e)
    }
  }
  if (!handled) {
    try {
      const body = JSON.parse(raw)
      if (body && body.code !== 0) {
        const msg = currentAssistant()
        msg.error = body.msg || '请求失败'
        msg.streaming = false
        sending.value = false
      }
    } catch { /* 非 JSON 帧，忽略 */ }
  }
}

const handleFrame = (frame) => {
  switch (frame.type) {
    case 'start':
      break
    case 'content': {
      const msg = currentAssistant()
      if (frame.content) msg.content += frame.content
      scrollToBottom()
      break
    }
    case 'tool_call': {
      const msg = currentAssistant()
      let tc = msg.toolCalls.find(t => t.id === frame.toolCallId)
      if (!tc) {
        tc = { id: frame.toolCallId, name: frame.toolName, args: '', status: 'running' }
        msg.toolCalls.push(tc)
      }
      tc.args = frame.toolParams || tc.args
      tc.status = frame.status || 'running'
      scrollToBottom()
      break
    }
    case 'tool_result': {
      const msg = currentAssistant()
      const tc = msg.toolCalls.find(t => t.id === frame.toolCallId)
      if (tc) tc.status = frame.status || 'done'
      break
    }
    case 'show_msg':
      break
    case 'error': {
      const msg = currentAssistant()
      msg.error = frame.error || '执行失败'
      break
    }
    case 'end':
      finalizeAssistant()
      break
    default:
      break
  }
}

const handleSend = () => {
  const text = input.value.trim()
  if (text && !sending.value) {
    input.value = ''
    sendMessage(text)
  }
}

const sendMessage = async (text) => {
  if (sending.value) return
  messages.value.push({ role: 'user', content: text, toolCalls: [], error: '', streaming: false })
  sending.value = true
  scrollToBottom()

  // 近 N 轮历史（仅 user/assistant 正文），服务端无状态
  const history = messages.value
    .filter(m => !m.streaming && (m.role === 'user' || (m.role === 'assistant' && m.content)))
    .slice(0, -1)
    .slice(-HISTORY_WINDOW)
    .map(m => ({ role: m.role, content: m.content }))

  // fetchSSEWithAuth 是 async 函数，须 await 拿到 AbortController
  controller = await fetchSSEWithAuth(
    '/api/wiki/ask',
    { sessionId, question: text, history },
    parseFrame,
    (error) => {
      if (error && error.name === 'AbortError') {
        const msg = messages.value[messages.value.length - 1]
        if (msg && msg.role === 'assistant') {
          msg.streaming = false
          if (!msg.content) msg.content = '（已停止）'
        }
        sending.value = false
        controller = null
        return
      }
      const msg = currentAssistant()
      msg.error = error?.message || '网络异常，请稍后重试'
      finalizeAssistant()
    },
    () => {
      // 流正常关闭但未见 end 帧（如代理截断）也收尾
      if (sending.value) finalizeAssistant()
    }
  )
  scrollToBottom()
}

const stop = () => {
  if (controller) {
    controller.abort()
    controller = null
  }
}

onBeforeUnmount(stop)
</script>

<style scoped lang="scss">
.ask-panel {
  width: 100%;
  max-width: 860px;
  margin: 0 auto;
}

.ask-card {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.ask-header {
  padding: 28px 28px 20px;
  text-align: center;
  background: linear-gradient(135deg, var(--el-color-primary-light-9), transparent);

  .ask-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 14px;
    border-radius: 999px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 14px;
  }

  .ask-title {
    margin: 0 0 6px;
    font-size: 22px;
    font-weight: 700;
    color: var(--el-text-color-primary);
  }

  .ask-subtitle {
    margin: 0;
    font-size: 14px;
    color: var(--el-text-color-secondary);
  }

  .ask-notice {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 14px;
    padding: 6px 14px;
    border-radius: 10px;
    background: var(--el-color-warning-light-9);
    color: var(--el-color-warning-dark-2);
    font-size: 12.5px;

    .el-icon { flex-shrink: 0; }
  }
}

.ask-messages {
  height: 380px;
  overflow-y: auto;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;

  &::-webkit-scrollbar { width: 6px; }
  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color);
    border-radius: 3px;
  }
}

.ask-welcome {
  margin: auto;
  text-align: center;

  .welcome-hint {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .suggestion-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    justify-content: center;
  }

  .suggestion-chip {
    padding: 8px 18px;
    border-radius: 999px;
    border: 1px solid var(--el-border-color);
    background: var(--el-bg-color);
    color: var(--el-text-color-regular);
    font-size: 13.5px;
    cursor: pointer;
    transition: all 0.25s;

    &:hover:not(:disabled) {
      border-color: var(--el-color-primary);
      color: var(--el-color-primary);
      transform: translateY(-2px);
      box-shadow: 0 4px 12px var(--el-color-primary-light-8);
    }

    &:disabled { opacity: 0.5; cursor: not-allowed; }
  }
}

.ask-msg {
  display: flex;
  gap: 10px;

  &--user {
    justify-content: flex-end;

    .msg-body {
      max-width: 78%;
      background: var(--el-color-primary);
      color: #fff;
      border-radius: 16px 16px 4px 16px;
      padding: 10px 16px;
    }
  }

  &--assistant {
    .msg-avatar {
      flex-shrink: 0;
      width: 34px;
      height: 34px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--el-color-primary-light-9);
      color: var(--el-color-primary);
      font-size: 17px;
    }

    .msg-body {
      max-width: 82%;
      background: var(--el-fill-color-light);
      border-radius: 4px 16px 16px 16px;
      padding: 12px 16px;
    }
  }
}

.msg-tools {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;

  &:empty { margin-bottom: 0; }
}

.tool-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &.is-running { opacity: 0.85; }
  .el-icon { flex-shrink: 0; }
}

.msg-content {
  font-size: 14px;
  line-height: 1.75;
  word-break: break-word;

  :deep(p) { margin: 0 0 8px; &:last-child { margin-bottom: 0; } }
  :deep(ul), :deep(ol) { margin: 4px 0 8px; padding-left: 20px; }
  :deep(h1), :deep(h2), :deep(h3), :deep(h4) {
    margin: 14px 0 8px;
    font-size: 15.5px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }
  :deep(blockquote) {
    margin: 8px 0;
    padding: 4px 12px;
    border-left: 3px solid var(--el-color-primary-light-5);
    color: var(--el-text-color-secondary);
  }
  :deep(hr) {
    border: none;
    border-top: 1px solid var(--el-border-color-lighter);
    margin: 12px 0;
  }
  :deep(a) { color: var(--el-color-primary); }
  :deep(code) {
    background: var(--el-fill-color);
    padding: 1px 6px;
    border-radius: 6px;
    font-size: 13px;
  }
  :deep(pre) {
    background: var(--el-fill-color);
    border-radius: 10px;
    padding: 12px;
    overflow-x: auto;
    code { background: transparent; padding: 0; }
  }

  // [[slug|标题]] 引用来源标签（utils/wikiMarkdown 转换）
  :deep(.wiki-cite) {
    display: inline-block;
    padding: 0 8px;
    margin: 0 2px;
    border-radius: 999px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-size: 12px;
    line-height: 20px;
    white-space: nowrap;

    &::before { content: '📄 '; font-size: 11px; }
  }
}

.msg-typing {
  display: inline-flex;
  gap: 5px;
  padding: 4px 0;

  span {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--el-text-color-placeholder);
    animation: ask-bounce 1.2s infinite;

    &:nth-child(2) { animation-delay: 0.2s; }
    &:nth-child(3) { animation-delay: 0.4s; }
  }
}

.msg-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-color-danger);
  font-size: 13px;
}

.ask-input {
  display: flex;
  gap: 10px;
  padding: 16px 20px 20px;
  border-top: 1px solid var(--el-border-color-lighter);

  :deep(.el-input__wrapper) { border-radius: 14px; }

  .send-btn {
    border-radius: 14px;
    padding: 0 18px;
  }

  .stop-btn {
    --el-button-bg-color: var(--el-color-danger);
    --el-button-border-color: var(--el-color-danger);
  }
}

@keyframes ask-bounce {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-6px); }
}
</style>
