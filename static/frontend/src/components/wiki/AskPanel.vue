<template>
  <div class="ask-panel">
    <div class="ask-card">
      <!-- 极光氛围层：缓慢漂移的淡彩辉光，纯装饰 -->
      <div class="ask-aurora aurora-a"></div>
      <div class="ask-aurora aurora-b"></div>

      <div class="ask-grid">
        <!-- 左栏：品牌 + 常驻推荐问题（不随对话消失） -->
        <aside class="ask-rail">
          <div class="ask-badge">
            <el-icon><ChatDotRound /></el-icon>
            <span>AI 问答</span>
          </div>
          <h3 class="ask-title">关于我，<br />想问点什么？</h3>
          <p class="ask-subtitle">由我的个人知识库驱动，AI 会现场查阅资料后作答</p>

          <div class="ask-suggest">
            <p class="suggest-hint">试试这些问题</p>
            <button
              v-for="q in suggestions"
              :key="q"
              class="suggest-chip"
              :disabled="sending"
              @click="sendMessage(q)"
            >
              <span class="chip-spark">✦</span>
              <span class="chip-text">{{ q }}</span>
              <el-icon class="chip-arrow"><ArrowRight /></el-icon>
            </button>
          </div>

          <!-- 留存声明（D10：问答全量留存，须明示） -->
          <div class="ask-notice">
            <el-icon><InfoFilled /></el-icon>
            <span>对话将被记录用于改进知识库质量，请勿输入敏感信息</span>
          </div>
        </aside>

        <!-- 右栏：聊天窗 -->
        <div class="ask-chat">
          <div class="ask-messages" ref="messagesEl">
            <!-- 欢迎态 -->
            <div v-if="messages.length === 0" class="ask-empty">
              <div class="empty-orb">
                <el-icon><Cpu /></el-icon>
              </div>
              <p class="empty-text">我是这份主页的 AI 分身<br />知识库实时检索，欢迎考我</p>
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
    </div>
  </div>
</template>

<script setup name="AskPanel">
import { ref, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ChatDotRound, InfoFilled, Cpu, Loading, CircleCheck,
  WarningFilled, Promotion, VideoPause, ArrowRight
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
/* ========== 舞台：白玻璃拟态，与页面卡片同语言 ========== */
.ask-panel {
  width: 100%;
}

.ask-card {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.06);
}

/* 极光氛围：两团缓慢漂移的淡彩辉光（浅色系） */
.ask-aurora {
  position: absolute;
  border-radius: 50%;
  filter: blur(70px);
  pointer-events: none;
  opacity: 0.75;

  &.aurora-a {
    width: 480px;
    height: 480px;
    left: -140px;
    top: -180px;
    background: radial-gradient(circle, rgba(99, 102, 241, 0.16), transparent 65%);
    animation: aurora-drift 16s ease-in-out infinite alternate;
  }

  &.aurora-b {
    width: 420px;
    height: 420px;
    right: -120px;
    bottom: -160px;
    background: radial-gradient(circle, rgba(34, 211, 238, 0.13), transparent 65%);
    animation: aurora-drift 20s ease-in-out infinite alternate-reverse;
  }
}

@keyframes aurora-drift {
  from { transform: translate(0, 0) scale(1); }
  to { transform: translate(60px, 40px) scale(1.15); }
}

.ask-grid {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 320px 1fr;
}

/* ========== 左栏 ========== */
.ask-rail {
  display: flex;
  flex-direction: column;
  padding: 36px 30px 28px;
  border-right: 1px solid rgba(99, 102, 241, 0.12);
}

.ask-badge {
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 14px;
  border-radius: 999px;
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  color: var(--el-color-primary);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.ask-title {
  margin: 18px 0 10px;
  font-size: 27px;
  line-height: 1.35;
  font-weight: 800;
  background: linear-gradient(120deg, var(--el-text-color-primary, #1e293b), var(--el-color-primary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.ask-subtitle {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.7;
  color: var(--el-text-color-secondary);
}

.ask-suggest {
  margin-top: auto;
  padding-top: 26px;

  .suggest-hint {
    margin: 0 0 12px;
    font-size: 12px;
    letter-spacing: 2px;
    text-transform: uppercase;
    color: var(--el-text-color-placeholder);
  }

  .suggest-chip {
    display: flex;
    align-items: center;
    gap: 9px;
    width: 100%;
    margin-bottom: 10px;
    padding: 11px 14px;
    border-radius: 14px;
    border: 1px solid var(--el-border-color-lighter);
    background: rgba(255, 255, 255, 0.75);
    color: var(--el-text-color-regular);
    font-size: 13.5px;
    text-align: left;
    cursor: pointer;
    transition: all 0.25s;

    .chip-spark {
      color: var(--el-color-primary);
      font-size: 12px;
      transition: transform 0.25s;
    }

    .chip-text {
      flex: 1;
      min-width: 0;
    }

    .chip-arrow {
      flex-shrink: 0;
      font-size: 13px;
      color: var(--el-color-primary);
      opacity: 0;
      transform: translateX(-6px);
      transition: all 0.25s;
    }

    &:hover:not(:disabled) {
      border-color: var(--el-color-primary-light-5);
      background: #fff;
      color: var(--el-text-color-primary);
      transform: translateX(4px);
      box-shadow: 0 6px 18px var(--el-color-primary-light-8);

      .chip-arrow { opacity: 1; transform: translateX(0); }
      .chip-spark { transform: rotate(90deg) scale(1.2); }
    }

    &:disabled { opacity: 0.5; cursor: not-allowed; }
  }
}

.ask-notice {
  display: flex;
  align-items: flex-start;
  gap: 7px;
  margin-top: 18px;
  padding: 10px 13px;
  border-radius: 12px;
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning-dark-2);
  font-size: 12px;
  line-height: 1.6;

  .el-icon { flex-shrink: 0; margin-top: 2px; }
}

/* ========== 右栏：聊天窗 ========== */
.ask-chat {
  display: flex;
  flex-direction: column;
  min-width: 0;
  /* 固定总高：消息再多也只在内层滚动，面板不被撑高 */
  height: 640px;
}

.ask-messages {
  flex: 1 1 0;
  min-height: 0; /* 允许被压缩，滚动收进内部 */
  overflow-y: auto;
  padding: 26px 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.04), rgba(34, 211, 238, 0.03));

  &::-webkit-scrollbar { width: 6px; }
  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color);
    border-radius: 3px;
  }
}

/* 欢迎态：AI 核心 + 引导语 */
.ask-empty {
  margin: auto;
  text-align: center;

  .empty-orb {
    width: 76px;
    height: 76px;
    margin: 0 auto 18px;
    border-radius: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 34px;
    color: #fff;
    background: linear-gradient(135deg, #2563eb, #4f46e5);
    box-shadow: 0 12px 32px rgba(79, 70, 229, 0.35);
    animation: orb-breathe 3s ease-in-out infinite;
  }

  .empty-text {
    margin: 0;
    font-size: 14.5px;
    line-height: 1.9;
    color: var(--el-text-color-secondary);
  }
}

@keyframes orb-breathe {
  0%, 100% { box-shadow: 0 12px 28px rgba(79, 70, 229, 0.3); transform: scale(1); }
  50% { box-shadow: 0 12px 44px rgba(79, 70, 229, 0.48); transform: scale(1.04); }
}

/* 消息气泡 */
.ask-msg {
  display: flex;
  gap: 10px;

  &--user {
    justify-content: flex-end;

    .msg-body {
      max-width: 78%;
      /* 深一档的蓝→靛：白字对比 5.2:1+，浅色页面上也清晰可读 */
      background: linear-gradient(135deg, #2563eb, #4f46e5);
      color: #fff;
      font-weight: 500;
      border-radius: 16px 16px 4px 16px;
      padding: 10px 16px;
      box-shadow: 0 6px 20px rgba(79, 70, 229, 0.3);
    }

    /* 正文层自带深灰文字色，须在用户气泡内翻白，否则深底深字不可读 */
    .msg-content {
      color: #fff;

      :deep(a) { color: #bfd7ff; }
      :deep(code) {
        background: rgba(255, 255, 255, 0.18);
        color: #fff;
      }
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
      background: linear-gradient(135deg, #2563eb, #4f46e5);
      color: #fff;
      font-size: 17px;
      box-shadow: 0 6px 16px rgba(79, 70, 229, 0.3);
    }

    .msg-body {
      max-width: 82%;
      background: rgba(255, 255, 255, 0.92);
      border: 1px solid var(--el-border-color-lighter);
      color: var(--el-text-color-regular);
      border-radius: 4px 16px 16px 16px;
      padding: 12px 16px;
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03);
    }
  }
}

/* 工具调用进度 */
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
  border: 1px solid var(--el-color-primary-light-7);
  color: var(--el-color-primary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &.is-running { opacity: 0.85; }
  .el-icon { flex-shrink: 0; }
}

/* 正文（浅色底配色） */
.msg-content {
  font-size: 14px;
  line-height: 1.75;
  word-break: break-word;
  color: var(--el-text-color-regular);

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
    background: var(--el-color-primary);
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

/* 输入区 */
.ask-input {
  display: flex;
  gap: 10px;
  padding: 16px 24px 22px;
  border-top: 1px solid rgba(99, 102, 241, 0.12);

  :deep(.el-input__wrapper) {
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.9);
    transition: box-shadow 0.25s;
  }

  .send-btn {
    border-radius: 14px;
    padding: 0 18px;
    border: none;
    background: linear-gradient(135deg, #2563eb, #4f46e5);
    transition: all 0.25s;

    &:not(:disabled):hover {
      box-shadow: 0 8px 22px rgba(79, 70, 229, 0.45);
      transform: translateY(-1px);
    }

    &.is-disabled {
      background: var(--el-fill-color);
      color: var(--el-text-color-placeholder);
    }
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

/* ========== 响应式：窄屏收成单列 ========== */
@media (max-width: 960px) {
  .ask-grid {
    grid-template-columns: 1fr;
    min-height: 0;
  }

  .ask-rail {
    border-right: none;
    border-bottom: 1px solid rgba(99, 102, 241, 0.12);
    padding: 28px 24px 22px;
  }

  .ask-title { margin-top: 14px; font-size: 22px; }

  .ask-suggest {
    margin-top: 18px;
    padding-top: 0;

    .suggest-chip {
      width: auto;
      margin-bottom: 0;

      .chip-arrow { display: none; }
    }
  }

  .ask-chat { height: auto; }
  .ask-messages { flex: none; height: 360px; padding: 20px 18px; }
  .ask-input { padding: 14px 18px 18px; }
}
</style>
