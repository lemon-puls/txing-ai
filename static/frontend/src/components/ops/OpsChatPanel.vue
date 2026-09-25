<template>
  <div class="ops-chat-panel">
    <!-- 会话工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-title" :title="sessionTitle">{{ sessionTitle || '新对话' }}</div>
      <div class="toolbar-actions">
        <el-popover
          v-model:visible="historyVisible"
          placement="bottom-end"
          :width="300"
          trigger="click"
          popper-class="ops-history-popover"
        >
          <template #reference>
            <el-button text size="small" class="toolbar-btn">
              <el-icon :size="14"><Clock /></el-icon>
              历史会话
            </el-button>
          </template>
          <div class="history-list">
            <div v-if="sessions.length === 0" class="history-empty">暂无历史会话</div>
            <div
              v-for="s in sessions"
              :key="s.id"
              class="history-item"
              :class="{ active: s.id === currentSessionId }"
              role="button"
              @click="switchSession(s)"
            >
              <div class="history-info">
                <div class="history-title">{{ s.title || '未命名会话' }}</div>
                <div class="history-time">{{ getRelativeTime(s.updateTime) }}</div>
              </div>
              <el-icon :size="14" class="history-delete" @click.stop="handleDeleteSession(s)">
                <Delete />
              </el-icon>
            </div>
          </div>
        </el-popover>
        <el-button text size="small" class="toolbar-btn" @click="startNewSession">
          <el-icon :size="14"><Plus /></el-icon>
          新对话
        </el-button>
      </div>
    </div>

    <!-- 消息列表 -->
    <div ref="listRef" class="message-list">
      <!-- 空状态引导 -->
      <div v-if="messages.length === 0" class="empty-guide">
        <div class="hero">
          <div class="hero-halo"></div>
          <div class="empty-icon">
            <el-icon :size="32"><MagicStick /></el-icon>
          </div>
        </div>
        <div class="empty-title">运营助手</div>
        <div class="empty-desc">
          提供一个网站或 GitHub 仓库地址，我会自动抓取信息并生成录入提案，确认后才会写入数据库。
        </div>

        <!-- 三步流程示意 -->
        <div class="flow-steps">
          <div class="flow-step">
            <span class="step-icon"><el-icon :size="12"><Link /></el-icon></span>
            粘贴网址
          </div>
          <el-icon :size="12" class="step-arrow"><ArrowRight /></el-icon>
          <div class="flow-step">
            <span class="step-icon accent"><el-icon :size="12"><Search /></el-icon></span>
            AI 抓取解析
          </div>
          <el-icon :size="12" class="step-arrow"><ArrowRight /></el-icon>
          <div class="flow-step">
            <span class="step-icon success"><el-icon :size="12"><CircleCheck /></el-icon></span>
            确认入库
          </div>
        </div>

        <!-- 示例引导 -->
        <div class="empty-examples">
          <div
            v-for="example in examples"
            :key="example"
            class="example-chip"
            role="button"
            @click="sendMessage(example)"
          >
            <span class="chip-icon"><el-icon :size="13"><Link /></el-icon></span>
            <span class="chip-text">{{ example }}</span>
            <el-icon :size="12" class="chip-arrow"><TopRight /></el-icon>
          </div>
        </div>
      </div>

      <!-- 对话消息 -->
      <template v-for="(msg, msgIndex) in messages" :key="msgIndex">
        <!-- 用户消息 -->
        <div v-if="msg.role === 'user'" class="message-row user">
          <div class="bubble user-bubble">{{ msg.content }}</div>
        </div>

        <!-- 助手消息 -->
        <div v-else class="message-row assistant">
          <div class="assistant-avatar">
            <el-icon :size="15"><MagicStick /></el-icon>
          </div>
          <div class="bubble assistant-bubble">
            <!-- 思考过程（可折叠） -->
            <div
              v-if="msg.reasoning"
              class="reasoning-block"
              :class="{ thinking: msg.streaming && !msg.content }"
            >
              <div
                class="reasoning-toggle"
                role="button"
                @click="msg.reasoningExpanded = !msg.reasoningExpanded"
              >
                <el-icon :size="12"><Opportunity /></el-icon>
                <span>{{ msg.streaming && !msg.content ? '深度思考中…' : '思考过程' }}</span>
                <el-icon :size="12" class="toggle-arrow" :class="{ expanded: msg.reasoningExpanded }">
                  <component :is="ArrowDown" />
                </el-icon>
              </div>
              <div v-show="msg.reasoningExpanded" class="reasoning-content">{{ msg.reasoning }}</div>
            </div>

            <!-- 正文（流式追加） -->
            <div v-if="msg.content" class="msg-content">{{ msg.content }}<span v-if="msg.streaming" class="cursor"></span></div>

            <!-- 工具调用过程 -->
            <div v-if="msg.toolCalls.length" class="tool-calls">
              <ToolCallItem v-for="tc in msg.toolCalls" :key="tc.id" :tool-call="tc" />
            </div>

            <!-- 网站录入提案卡片 -->
            <WebsiteProposalCard
              v-if="msg.proposal"
              :proposal="msg.proposal"
              :proposal-status="msg.proposalStatus"
              :proposal-message="msg.proposalMessage"
              :confirmed="msg.proposalStatus === 'confirmed'"
              @confirmed="handleProposalConfirmed"
            />

            <!-- 错误提示 -->
            <div v-if="msg.error" class="msg-error">
              <el-icon :size="14"><WarningFilled /></el-icon>
              <span>{{ msg.error }}</span>
            </div>

            <!-- 中断标记 -->
            <div v-if="msg.interrupted && !msg.streaming" class="msg-interrupted">
              <el-icon :size="12"><Minus /></el-icon>
              已中断，仅保留已生成内容
            </div>
          </div>
        </div>
      </template>

      <!-- 等待首响应 -->
      <div v-if="waitingFirst" class="message-row assistant">
        <div class="assistant-avatar pulse">
          <el-icon :size="15"><MagicStick /></el-icon>
        </div>
        <div class="bubble assistant-bubble waiting">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="waiting-text">正在思考…</span>
        </div>
      </div>
    </div>

    <!-- 输入区 -->
    <div class="input-area">
      <div class="composer" :class="{ sending }">
        <el-input
          v-model="input"
          type="textarea"
          :rows="2"
          resize="none"
          :disabled="sending"
          placeholder="输入网站或 GitHub 地址，例如：https://github.com/cloudwego/eino"
          @keydown.enter.exact.prevent="handleSend"
        />
        <div class="composer-footer">
          <span class="input-hint">
            <el-icon :size="12"><CircleCheck /></el-icon>
            AI 不会直接写入数据，提案需人工确认
          </span>
          <el-button
            v-if="sending"
            type="danger"
            plain
            round
            size="small"
            @click="stopStreaming"
          >
            <el-icon class="btn-icon"><VideoPause /></el-icon>
            停止
          </el-button>
          <el-button
            v-else
            type="primary"
            round
            size="small"
            class="send-btn"
            :disabled="!input.trim()"
            @click="handleSend"
          >
            <el-icon class="btn-icon"><Promotion /></el-icon>
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, watch, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, ArrowDown, WarningFilled, Clock, Plus, Delete, Minus } from '@element-plus/icons-vue'
import { fetchSSEWithAuth } from '@/api/sseRequest'
import { defaultApi } from '@/api'
import { getRelativeTime } from '@/utils/timeUtils'
import ToolCallItem from '@/components/chat/ToolCallItem.vue'
import WebsiteProposalCard from './WebsiteProposalCard.vue'

const props = defineProps({
  // 页面上下文，注入后端系统提示词，如 { page: 'websites', draft: { url } }
  context: { type: Object, default: null }
})

const emit = defineEmits(['inserted'])

// 示例引导语
const examples = [
  '收录 https://github.com/cloudwego/eino',
  '收录 https://gorm.io/docs/'
]

// 对话消息：{ role, content, reasoning, reasoningExpanded, toolCalls[], proposal, proposalStatus, proposalMessage, error, interrupted, streaming }
const messages = ref([])
const input = ref('')
const sending = ref(false)
const waitingFirst = ref(false)
const listRef = ref(null)

// 会话持久化状态
const currentSessionId = ref(0)
const sessionTitle = ref('')
const sessions = ref([])
const historyVisible = ref(false)

// 多轮历史由服务端持久化并构建，前端只传本次输入
let controller = null

const scrollToBottom = () => {
  nextTick(() => {
    const el = listRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

// 消息长度变化时滚动到底部（字段需判空：user 消息没有 toolCalls，读取 undefined 会中断调度队列导致渲染冻结）
watch(() => messages.value.map(m => (m.content || '').length + (m.reasoning || '').length + (m.toolCalls || []).length), scrollToBottom)

// ===== 会话历史 =====

// 统一消息形状：后端持久化消息 → 前端渲染所需字段（补齐缺省，toolCalls.params 映射为 args）
const normalizeMessage = (m) => ({
  role: m.role,
  content: m.content || '',
  reasoning: m.reasoning || '',
  reasoningExpanded: false,
  toolCalls: (m.toolCalls || []).map(tc => ({
    id: tc.id,
    name: tc.name,
    args: tc.params || '',
    result: tc.result || '',
    status: tc.status || 'completed'
  })),
  proposal: m.proposal || null,
  proposalStatus: m.proposalStatus || '',
  proposalMessage: m.proposalMessage || '',
  error: m.error || '',
  interrupted: !!m.interrupted,
  streaming: false
})

const loadSessions = async () => {
  try {
    const response = await defaultApi.apiAdminOpsChatSessionsListPost({ pageSize: 30 })
    if (response.code === 0) {
      // 游标分页 VO 的记录在 data.data 字段
      sessions.value = response.data?.data || []
    }
  } catch (error) {
    console.warn('加载运营助手会话列表失败:', error)
  }
}

// 加载会话详情并回放消息
const loadSession = async (id) => {
  try {
    const response = await defaultApi.apiAdminOpsChatSessionsIdGet(id)
    if (response.code !== 0) {
      ElMessage.error(response.message || '加载会话失败')
      return false
    }
    const detail = response.data || {}
    currentSessionId.value = detail.id
    sessionTitle.value = detail.title || ''
    messages.value = (detail.messages || []).map(normalizeMessage)
    scrollToBottom()
    return true
  } catch (error) {
    console.error('加载运营助手会话失败:', error)
    ElMessage.error('加载会话失败')
    return false
  }
}

// 挂载时自动续接最近会话
const restoreLatestSession = async () => {
  try {
    const response = await defaultApi.apiAdminOpsChatSessionsListPost({ pageSize: 1 })
    if (response.code !== 0) return
    const latest = response.data?.data?.[0]
    if (latest) await loadSession(latest.id)
  } catch (error) {
    console.warn('恢复运营助手会话失败:', error)
  }
}

onMounted(() => {
  restoreLatestSession()
})

// 打开历史弹层时懒更新列表（标题/排序可能已变化）
watch(historyVisible, (visible) => {
  if (visible) loadSessions()
})

const switchSession = async (session) => {
  if (session.id === currentSessionId.value) {
    historyVisible.value = false
    return
  }
  stopStreaming()
  if (await loadSession(session.id)) {
    historyVisible.value = false
  }
}

const startNewSession = () => {
  stopStreaming()
  currentSessionId.value = 0
  sessionTitle.value = ''
  messages.value = []
  historyVisible.value = false
}

const handleDeleteSession = async (session) => {
  try {
    await ElMessageBox.confirm(`确定要删除会话「${session.title || '未命名会话'}」吗？`, '删除会话', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const response = await defaultApi.apiAdminOpsChatSessionsIdDelete(session.id)
    if (response.code !== 0) {
      ElMessage.error(response.message || '删除失败')
      return
    }
    if (session.id === currentSessionId.value) {
      startNewSession()
    }
    await loadSessions()
    ElMessage.success('会话已删除')
  } catch (error) {
    console.error('删除运营助手会话失败:', error)
    ElMessage.error('删除失败')
  }
}

// SSE 帧解析：每帧形如 "data: {...}"
const parseFrame = (raw) => {
  for (const line of raw.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed.startsWith('data:')) continue
    const payload = trimmed.slice(5).trim()
    if (!payload || payload === '[DONE]') continue
    try {
      handleFrame(JSON.parse(payload))
    } catch (e) {
      console.warn('运营助手：无法解析 SSE 帧', payload, e)
    }
  }
}

// 取当前流式中的助手消息（没有则创建）
const currentAssistant = () => {
  let msg = messages.value[messages.value.length - 1]
  if (!msg || msg.role !== 'assistant' || !msg.streaming) {
    msg = {
      role: 'assistant',
      content: '',
      reasoning: '',
      reasoningExpanded: false,
      toolCalls: [],
      proposal: null,
      proposalStatus: '',
      proposalMessage: '',
      error: '',
      interrupted: false,
      streaming: true
    }
    messages.value.push(msg)
  }
  return msg
}

const handleFrame = (frame) => {
  switch (frame.type) {
    case 'session': {
      // 服务端回传会话ID，绑定后续请求
      if (frame.sessionId) {
        currentSessionId.value = frame.sessionId
        if (!sessionTitle.value) {
          const lastUser = [...messages.value].reverse().find(m => m.role === 'user')
          sessionTitle.value = lastUser ? lastUser.content.slice(0, 35) : ''
        }
      }
      break
    }
    case 'content': {
      const msg = currentAssistant()
      if (frame.content) msg.content += frame.content
      if (frame.reasoningContent) {
        msg.reasoning += frame.reasoningContent
        // 思考阶段自动展开，正文出现后由用户手动控制
        if (!msg.content) msg.reasoningExpanded = true
      }
      waitingFirst.value = false
      break
    }
    case 'show_msg': {
      // 过程提示暂以内容 delta 同级展示（一般出现在工具调用间隙）
      if (frame.showMsg) {
        const msg = currentAssistant()
        if (!msg.content && !msg.reasoning) msg.content = frame.showMsg
      }
      break
    }
    case 'tool_call': {
      const msg = currentAssistant()
      let tc = msg.toolCalls.find(t => t.id === frame.toolCallId)
      if (!tc) {
        tc = { id: frame.toolCallId, name: frame.toolName, args: '', result: '', status: 'running' }
        msg.toolCalls.push(tc)
      }
      tc.name = frame.toolName || tc.name
      tc.args = frame.toolParams || tc.args
      tc.status = frame.status || 'running'
      break
    }
    case 'tool_result': {
      const msg = currentAssistant()
      let tc = msg.toolCalls.find(t => t.id === frame.toolCallId)
      if (!tc) {
        tc = { id: frame.toolCallId, name: frame.toolName, args: '', result: '', status: frame.status }
        msg.toolCalls.push(tc)
      }
      tc.result = frame.toolResult || tc.result
      tc.status = frame.status || tc.status
      break
    }
    case 'proposal': {
      const msg = currentAssistant()
      msg.proposal = frame.proposal
      msg.proposalStatus = frame.status || 'ok'
      msg.proposalMessage = frame.message || ''
      break
    }
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

const finalizeAssistant = () => {
  const msg = messages.value[messages.value.length - 1]
  if (msg && msg.role === 'assistant') {
    msg.streaming = false
    // 空消息（无内容无工具无提案）直接移除
    if (!msg.content && !msg.reasoning && !msg.toolCalls.length && !msg.proposal && !msg.error && !msg.interrupted) {
      messages.value.pop()
    }
  }
  sending.value = false
  waitingFirst.value = false
  controller = null
}

const handleSend = () => {
  const text = input.value.trim()
  if (text && !sending.value) {
    input.value = ''
    sendMessage(text)
  }
}

// 统一消息形状：补齐所有字段，避免模板/watch 读取缺失字段出错
const createUserMessage = (text) => ({
  role: 'user',
  content: text,
  reasoning: '',
  reasoningExpanded: false,
  toolCalls: [],
  proposal: null,
  proposalStatus: '',
  proposalMessage: '',
  error: '',
  interrupted: false,
  streaming: false
})

const sendMessage = async (text) => {
  if (sending.value) return
  messages.value.push(createUserMessage(text))
  sending.value = true
  waitingFirst.value = true
  scrollToBottom()

  // 多轮历史由服务端从持久化会话构建，这里只传本次输入
  // fetchSSEWithAuth 是 async 函数，须 await 拿到 AbortController（否则存的是 Promise，停止按钮失效）
  controller = await fetchSSEWithAuth(
    '/api/admin/ops/chat/stream',
    { sessionId: currentSessionId.value, content: text, context: props.context || undefined },
    parseFrame,
    (error) => {
      // AbortController 主动中断不作为错误展示，标记中断状态
      if (error && error.name === 'AbortError') {
        const msg = messages.value[messages.value.length - 1]
        if (msg && msg.role === 'assistant') msg.interrupted = true
        finalizeAssistant()
        return
      }
      const msg = currentAssistant()
      msg.error = error?.message || '连接失败，请稍后重试'
      finalizeAssistant()
      console.error('运营助手请求失败:', error)
    },
    () => finalizeAssistant()
  )
}

const stopStreaming = () => {
  if (controller) {
    controller.abort()
    controller = null
  }
  finalizeAssistant()
}

const handleProposalConfirmed = () => {
  ElMessage.success('网站已录入')
  emit('inserted')
  // 本地置为已确认，防止回放后二次确认
  const msg = messages.value[messages.value.length - 1]
  if (msg && msg.role === 'assistant' && msg.proposal) {
    msg.proposalStatus = 'confirmed'
  }
  // 追加一条本地提示，明确提案已完成
  messages.value.push({ role: 'assistant', content: '✅ 提案已确认，网站录入完成。', streaming: false, toolCalls: [], reasoning: '', reasoningExpanded: false, proposal: null, proposalStatus: '', proposalMessage: '', error: '', interrupted: false })
  scrollToBottom()

  // 持久化确认状态（失败不影响本地 UI，数据库侧有 URL 查重兜底）
  if (currentSessionId.value) {
    defaultApi
      .apiAdminOpsChatSessionsIdAppendPost(currentSessionId.value, {
        role: 'assistant',
        content: '✅ 提案已确认，网站录入完成。',
        markProposalConfirmed: true
      })
      .catch((error) => console.warn('持久化提案确认状态失败:', error))
  }
}

onBeforeUnmount(stopStreaming)
</script>

<style lang="scss" scoped>
.ops-chat-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.message-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 20px 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 18px;

  // 细滚动条
  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color-lighter);
    border-radius: 3px;

    &:hover {
      background: var(--el-border-color-light);
    }
  }
}

// ===== 空状态引导 =====
.empty-guide {
  margin: auto;
  text-align: center;
  max-width: 380px;
  padding: 24px 0;
}

.hero {
  position: relative;
  width: 64px;
  margin: 0 auto 18px;
}

.empty-icon {
  position: relative;
  z-index: 1;
  width: 64px;
  height: 64px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  box-shadow: 0 8px 24px var(--el-color-primary-light-8);
}

// 图标底部的呼吸光环
.hero-halo {
  position: absolute;
  inset: -9px;
  border-radius: 26px;
  background: var(--el-color-primary-light-9);
  animation: halo-breathe 3s ease-in-out infinite;
}

@keyframes halo-breathe {
  0%, 100% { transform: scale(1); opacity: 0.7; }
  50% { transform: scale(1.08); opacity: 1; }
}

.empty-title {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
}

.empty-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.7;
  margin: 0 auto 18px;
  max-width: 320px;
}

// 三步流程示意
.flow-steps {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: 22px;

  .flow-step {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 5px 10px;
    border-radius: 999px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    white-space: nowrap;
  }

  .step-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border-radius: 6px;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);

    &.accent {
      color: var(--el-color-warning);
      background: var(--el-color-warning-light-9);
    }

    &.success {
      color: var(--el-color-success);
      background: var(--el-color-success-light-9);
    }
  }

  .step-arrow {
    color: var(--el-text-color-placeholder);
    flex-shrink: 0;
  }
}

.empty-examples {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.example-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  cursor: pointer;
  background: var(--el-bg-color);
  transition: all 0.2s ease;
  text-align: left;

  .chip-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 8px;
    flex-shrink: 0;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }

  .chip-text {
    flex: 1;
    min-width: 0;
    word-break: break-all;
  }

  .chip-arrow {
    flex-shrink: 0;
    color: var(--el-text-color-placeholder);
    transition: transform 0.2s ease, color 0.2s ease;
  }

  &:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 4px 12px var(--el-color-primary-light-9);
    transform: translateY(-1px);

    .chip-text {
      color: var(--el-color-primary);
    }

    .chip-arrow {
      color: var(--el-color-primary);
      transform: translate(1px, -1px);
    }
  }
}

// ===== 消息行 =====
.message-row {
  display: flex;
  align-items: flex-start;
  // 入场动画
  animation: msg-in 0.25s ease both;

  &.user {
    justify-content: flex-end;
  }

  &.assistant {
    justify-content: flex-start;
  }
}

@keyframes msg-in {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .message-row,
  .hero-halo {
    animation: none;
  }
}

// 助手头像
.assistant-avatar {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  flex-shrink: 0;
  margin-right: 10px;
  margin-top: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  box-shadow: 0 2px 8px var(--el-color-primary-light-8);

  // 等待响应时头像呼吸
  &.pulse {
    animation: avatar-pulse 1.6s ease-in-out infinite;
  }
}

@keyframes avatar-pulse {
  0%, 100% { box-shadow: 0 2px 8px var(--el-color-primary-light-8); }
  50% { box-shadow: 0 2px 8px var(--el-color-primary-light-5); }
}

.bubble {
  max-width: min(92%, 640px);
  font-size: 14px;
  line-height: 1.6;
}

.user-bubble {
  padding: 10px 14px;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  color: #fff;
  white-space: pre-wrap;
  word-break: break-word;
  border-radius: 16px 16px 4px 16px;
  box-shadow: 0 2px 10px var(--el-color-primary-light-8);
}

.assistant-bubble {
  padding: 12px 14px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
  border-radius: 4px 16px 16px 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);

  &.waiting {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 14px 16px;

    .waiting-text {
      margin-left: 6px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }
}

// 等待打字点
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--el-color-primary);
  opacity: 0.5;
  animation: dot-bounce 1.2s ease-in-out infinite;

  &:nth-child(2) { animation-delay: 0.15s; }
  &:nth-child(3) { animation-delay: 0.3s; }
}

@keyframes dot-bounce {
  0%, 80%, 100% {
    transform: scale(0.6);
    opacity: 0.4;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.msg-content {
  white-space: pre-wrap;
}

// 流式光标
.cursor {
  display: inline-block;
  width: 2px;
  height: 14px;
  margin-left: 2px;
  vertical-align: -2px;
  border-radius: 1px;
  background: var(--el-color-primary);
  animation: blink 1s step-start infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

// ===== 思考过程 =====
.reasoning-block {
  margin-bottom: 10px;
  padding: 8px 10px;
  background: var(--el-fill-color-lighter);
  border-radius: 10px;

  .reasoning-toggle {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    cursor: pointer;
    user-select: none;

    .toggle-arrow {
      transition: transform 0.2s ease;

      &.expanded {
        transform: rotate(180deg);
      }
    }
  }

  // 思考中呼吸提示
  &.thinking .reasoning-toggle {
    color: var(--el-color-primary);
    animation: reason-breathe 1.8s ease-in-out infinite;
  }

  @keyframes reason-breathe {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.55; }
  }

  .reasoning-content {
    margin-top: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
    white-space: pre-wrap;
    max-height: 200px;
    overflow-y: auto;
  }
}

.tool-calls {
  margin: 8px 0 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.msg-error {
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 10px;
  display: flex;
  align-items: flex-start;
  gap: 6px;
  color: var(--el-color-danger);
  font-size: 13px;
  background: var(--el-color-danger-light-9);

  .el-icon {
    margin-top: 2px;
  }
}

// 中断标记
.msg-interrupted {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

// ===== 会话工具栏 =====
.panel-toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  height: 40px;
  padding: 0 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  .toolbar-title {
    min-width: 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .toolbar-actions {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 2px;
  }

  .toolbar-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 5px 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }
}

// ===== 输入区 =====
.input-area {
  flex-shrink: 0;
  padding: 12px 16px 14px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.composer {
  border: 1px solid var(--el-border-color-light);
  border-radius: 16px;
  background: var(--el-bg-color);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;

  &:focus-within {
    border-color: var(--el-color-primary);
    box-shadow: 0 0 0 3px var(--el-color-primary-light-8);
  }

  &.sending {
    background: var(--el-fill-color-lighter);
  }

  :deep(.el-textarea__inner) {
    border: none;
    box-shadow: none;
    background: transparent;
    padding: 12px 14px 2px;
    font-size: 14px;
    line-height: 1.6;
  }
}

.composer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px 8px 14px;

  .input-hint {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

.btn-icon {
  margin-right: 4px;
}

// 发送按钮渐变
.send-btn {
  border: none;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  box-shadow: 0 2px 8px var(--el-color-primary-light-8);

  &:not(:disabled):hover {
    opacity: 0.88;
  }
}

// 暗色模式补充：项目暗色主题未覆写 --el-border-color-lighter / --el-fill-color-lighter / primary-light-N，这里手动补齐
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.ops-chat-panel {
  html.dark & {
    .assistant-bubble {
      border-color: #363637;
      box-shadow: none;
    }

    .reasoning-block {
      background: #262727;
    }

    .flow-step {
      border-color: #363637;
    }

    .hero-halo {
      background: rgba(255, 255, 255, 0.07);
      animation: none;
    }

    .assistant-avatar,
    .empty-icon {
      box-shadow: none;
    }

    .user-bubble {
      background: var(--el-color-primary);
      box-shadow: none;
    }

    .input-area {
      border-top-color: #363637;
    }

    .composer {
      box-shadow: none;

      &.sending {
        background: #1c1c1c;
      }
    }

    .send-btn {
      box-shadow: none;
    }

    .msg-error {
      background: rgba(245, 108, 108, 0.12);
    }

    .panel-toolbar {
      border-bottom-color: #363637;
    }
  }
}
</style>

<style lang="scss">
// 历史会话弹层（el-popover teleport 到 body，scoped 样式不生效，需全局；颜色全部用主题变量以适配暗色）
.ops-history-popover {
  .history-list {
    max-height: 320px;
    overflow-y: auto;
    margin: -6px -12px;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-thumb {
      background: var(--el-border-color-lighter);
      border-radius: 3px;
    }
  }

  .history-empty {
    margin: -6px -12px;
    padding: 20px 0;
    text-align: center;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .history-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 8px 12px;
    cursor: pointer;
    transition: background 0.15s ease;

    &:hover {
      background: var(--el-fill-color-light);

      .history-delete {
        opacity: 1;
      }
    }

    &.active {
      background: var(--el-color-primary-light-9);

      .history-title {
        color: var(--el-color-primary);
      }
    }

    & + .history-item {
      border-top: 1px solid var(--el-border-color-extra-light);
    }

    .history-info {
      flex: 1;
      min-width: 0;
    }

    .history-title {
      font-size: 13px;
      color: var(--el-text-color-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .history-time {
      margin-top: 2px;
      font-size: 11px;
      color: var(--el-text-color-secondary);
    }

    .history-delete {
      flex-shrink: 0;
      opacity: 0;
      color: var(--el-text-color-secondary);
      transition: all 0.15s ease;

      &:hover {
        color: var(--el-color-danger);
      }
    }
  }
}
</style>
