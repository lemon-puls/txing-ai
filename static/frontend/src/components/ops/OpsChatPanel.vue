<template>
  <div class="ops-chat-panel">
    <!-- 消息列表 -->
    <div ref="listRef" class="message-list">
      <!-- 空状态引导 -->
      <div v-if="messages.length === 0" class="empty-guide">
        <div class="empty-icon">
          <el-icon :size="36"><MagicStick /></el-icon>
        </div>
        <div class="empty-title">运营助手</div>
        <div class="empty-desc">
          提供一个网站或 GitHub 仓库地址，我会抓取信息并生成录入提案，确认后才会入库。
        </div>
        <div class="empty-examples">
          <div
            v-for="example in examples"
            :key="example"
            class="example-chip"
            role="button"
            @click="sendMessage(example)"
          >
            {{ example }}
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
          <div class="bubble assistant-bubble">
            <!-- 思考过程（可折叠） -->
            <div v-if="msg.reasoning" class="reasoning-block">
              <div
                class="reasoning-toggle"
                role="button"
                @click="msg.reasoningExpanded = !msg.reasoningExpanded"
              >
                <el-icon :size="12"><component :is="msg.reasoningExpanded ? ArrowDown : ArrowRight" /></el-icon>
                <span>{{ msg.streaming && !msg.content ? '思考中…' : '思考过程' }}</span>
              </div>
              <div v-show="msg.reasoningExpanded" class="reasoning-content">{{ msg.reasoning }}</div>
            </div>

            <!-- 正文（流式追加） -->
            <div v-if="msg.content" class="msg-content">{{ msg.content }}<span v-if="msg.streaming" class="cursor">▍</span></div>

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
              @confirmed="handleProposalConfirmed"
            />

            <!-- 错误提示 -->
            <div v-if="msg.error" class="msg-error">
              <el-icon :size="14"><WarningFilled /></el-icon>
              <span>{{ msg.error }}</span>
            </div>
          </div>
        </div>
      </template>

      <!-- 等待首响应 -->
      <div v-if="waitingFirst" class="message-row assistant">
        <div class="bubble assistant-bubble waiting">
          <el-icon :size="14" class="spin"><Loading /></el-icon>
          <span>正在思考…</span>
        </div>
      </div>
    </div>

    <!-- 输入区 -->
    <div class="input-area">
      <el-input
        v-model="input"
        type="textarea"
        :rows="2"
        resize="none"
        :disabled="sending"
        placeholder="输入网站或 GitHub 地址，例如：https://github.com/cloudwego/eino"
        @keydown.enter.exact.prevent="handleSend"
      />
      <div class="input-actions">
        <span class="input-hint">AI 不会直接写入数据，提案需人工确认</span>
        <el-button
          v-if="sending"
          type="danger"
          plain
          round
          size="small"
          @click="stopStreaming"
        >
          停止
        </el-button>
        <el-button
          v-else
          type="primary"
          round
          size="small"
          :disabled="!input.trim()"
          @click="handleSend"
        >
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, watch, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { MagicStick, ArrowDown, ArrowRight, Loading, WarningFilled } from '@element-plus/icons-vue'
import { fetchSSEWithAuth } from '@/api/sseRequest'
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

// 对话消息：{ role, content, reasoning, reasoningExpanded, toolCalls[], proposal, proposalStatus, proposalMessage, error, streaming }
const messages = ref([])
const input = ref('')
const sending = ref(false)
const waitingFirst = ref(false)
const listRef = ref(null)

// 多轮历史（仅 role + content，proposal 确认后的提示由下一轮 user 消息承载）
let controller = null

const scrollToBottom = () => {
  nextTick(() => {
    const el = listRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

// 消息长度变化时滚动到底部（字段需判空：user 消息没有 toolCalls，读取 undefined 会中断调度队列导致渲染冻结）
watch(() => messages.value.map(m => (m.content || '').length + (m.reasoning || '').length + (m.toolCalls || []).length), scrollToBottom)

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
      streaming: true
    }
    messages.value.push(msg)
  }
  return msg
}

const handleFrame = (frame) => {
  switch (frame.type) {
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
    if (!msg.content && !msg.reasoning && !msg.toolCalls.length && !msg.proposal && !msg.error) {
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
  streaming: false
})

const sendMessage = (text) => {
  if (sending.value) return
  messages.value.push(createUserMessage(text))
  sending.value = true
  waitingFirst.value = true
  scrollToBottom()

  // 组装多轮历史（最后一条为本次输入）
  const payloadMessages = messages.value
    .filter(m => !m.error)
    .map(m => ({ role: m.role, content: m.content || text }))
  // 上一条助手消息若只有提案没有正文，避免空 content 请求失败
  for (const m of payloadMessages) {
    if (!m.content) m.content = '（生成了录入提案）'
  }

  controller = fetchSSEWithAuth(
    '/api/admin/ops/chat/stream',
    { messages: payloadMessages, context: props.context || undefined },
    parseFrame,
    (error) => {
      // AbortController 主动中断不作为错误展示
      if (error && error.name === 'AbortError') {
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
  // 追加一条本地提示，明确提案已完成
  messages.value.push({ role: 'assistant', content: '✅ 提案已确认，网站录入完成。', streaming: false, toolCalls: [], reasoning: '', reasoningExpanded: false, proposal: null, proposalStatus: '', proposalMessage: '', error: '' })
  scrollToBottom()
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
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

// 空状态
.empty-guide {
  margin: auto;
  text-align: center;
  max-width: 360px;

  .empty-icon {
    width: 72px;
    height: 72px;
    margin: 0 auto 16px;
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }

  .empty-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 8px;
  }

  .empty-desc {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
    margin-bottom: 20px;
  }

  .empty-examples {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .example-chip {
    padding: 8px 14px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    cursor: pointer;
    transition: all 0.2s;
    word-break: break-all;

    &:hover {
      border-color: var(--el-color-primary);
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }
}

.message-row {
  display: flex;

  &.user {
    justify-content: flex-end;
  }

  &.assistant {
    justify-content: flex-start;
  }
}

.bubble {
  max-width: 92%;
  border-radius: 14px;
  font-size: 14px;
  line-height: 1.6;
}

.user-bubble {
  padding: 10px 14px;
  background: var(--el-color-primary);
  color: #fff;
  white-space: pre-wrap;
  word-break: break-word;
  border-bottom-right-radius: 4px;
}

.assistant-bubble {
  padding: 10px 14px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
  border-bottom-left-radius: 4px;

  &.waiting {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--el-text-color-secondary);
  }
}

.msg-content {
  white-space: pre-wrap;
}

.cursor {
  display: inline-block;
  animation: blink 1s step-start infinite;
  color: var(--el-color-primary);
}

@keyframes blink {
  50% { opacity: 0; }
}

// 思考过程
.reasoning-block {
  margin-bottom: 8px;
  border-left: 2px solid var(--el-border-color);
  padding-left: 8px;

  .reasoning-toggle {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    cursor: pointer;
    user-select: none;
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
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-color-danger);
  font-size: 13px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// 输入区
.input-area {
  flex-shrink: 0;
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);

  :deep(.el-textarea__inner) {
    border-radius: 12px;
  }
}

.input-actions {
  margin-top: 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;

  .input-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}
</style>
