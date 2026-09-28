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
          <div class="history-search">
            <el-input
              v-model="historyKeyword"
              size="small"
              clearable
              placeholder="搜索会话标题"
              :prefix-icon="Search"
            />
          </div>
          <div class="history-list">
            <div v-if="filteredSessions.length === 0" class="history-empty">
              {{ historyKeyword.trim() ? '没有匹配的会话' : '暂无历史会话' }}
            </div>
            <div
              v-for="s in filteredSessions"
              :key="s.id"
              class="history-item"
              :class="{ active: s.id === currentSessionId }"
              role="button"
              @click="switchSession(s)"
            >
              <div class="history-info">
                <div class="history-title">
                  <span class="history-title-text">{{ s.title || '未命名会话' }}</span>
                  <span v-if="PAGE_LABELS[s.page]" class="history-page-tag">{{ PAGE_LABELS[s.page] }}</span>
                </div>
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

    <!-- 消息列表（list-wrap 承载「回到底部」悬浮按钮的定位） -->
    <div class="list-wrap">
    <div ref="listRef" class="message-list" @scroll="handleListScroll">
      <!-- 空状态引导 -->
      <div v-if="messages.length === 0" class="empty-guide">
        <div class="hero">
          <div class="hero-halo"></div>
          <div class="empty-icon">
            <el-icon :size="32"><MagicStick /></el-icon>
          </div>
        </div>
        <div class="empty-title">运营助手</div>
        <div class="empty-desc">{{ pageMeta.desc }}</div>

        <!-- 三步流程示意 -->
        <div class="flow-steps">
          <div class="flow-step">
            <span class="step-icon"><el-icon :size="12"><component :is="pageMeta.steps[0].icon" /></el-icon></span>
            {{ pageMeta.steps[0].label }}
          </div>
          <el-icon :size="12" class="step-arrow"><ArrowRight /></el-icon>
          <div class="flow-step">
            <span class="step-icon accent"><el-icon :size="12"><component :is="pageMeta.steps[1].icon" /></el-icon></span>
            {{ pageMeta.steps[1].label }}
          </div>
          <el-icon :size="12" class="step-arrow"><ArrowRight /></el-icon>
          <div class="flow-step">
            <span class="step-icon success"><el-icon :size="12"><CircleCheck /></el-icon></span>
            {{ pageMeta.steps[2].label }}
          </div>
        </div>

        <!-- 示例引导 -->
        <div class="empty-examples">
          <div
            v-for="example in pageMeta.examples"
            :key="example"
            class="example-chip"
            role="button"
            @click="sendMessage(example)"
          >
            <span class="chip-icon"><el-icon :size="13"><MagicStick /></el-icon></span>
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
          <button class="copy-btn" title="复制" @click="copyMessage(msg)">
            <el-icon :size="13"><CopyDocument /></el-icon>
          </button>
        </div>

        <!-- 助手消息 -->
        <div v-else class="message-row assistant">
          <div class="assistant-avatar">
            <el-icon :size="15"><MagicStick /></el-icon>
          </div>
          <div class="assistant-main">
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
                @click="toggleReasoning(msg)"
              >
                <el-icon :size="12"><Opportunity /></el-icon>
                <span>{{ msg.streaming && !msg.content ? '深度思考中…' : '思考过程' }}</span>
                <el-icon :size="12" class="toggle-arrow" :class="{ expanded: msg.reasoningExpanded }">
                  <component :is="ArrowDown" />
                </el-icon>
              </div>
              <div v-show="msg.reasoningExpanded" class="reasoning-content">{{ msg.reasoning }}</div>
            </div>

            <!-- 正文（流式追加，markdown 渲染） -->
            <div v-if="msg.content" class="msg-content" v-html="renderMarkdown(msg.content)"></div>

            <!-- 工具调用过程 -->
            <div v-if="msg.toolCalls.length" class="tool-calls">
              <ToolCallItem v-for="tc in msg.toolCalls" :key="tc.id" :tool-call="tc" />
            </div>

            <!-- 结构化提案卡片（按 proposal.type 分发；一条消息可携带多张提案，批量录入时逐张渲染） -->
            <template v-for="(item, pIndex) in msg.proposals" :key="pIndex">
              <WebsiteProposalCard
                v-if="item.proposal?.type === 'website'"
                :proposal="item.proposal"
                :proposal-status="item.proposalStatus"
                :proposal-message="item.proposalMessage"
                :confirmed="item.proposalStatus === 'confirmed'"
                @confirmed="handleProposalConfirmed(msg, item)"
              />
              <ModelProposalCard
                v-else-if="item.proposal?.type === 'model'"
                :proposal="item.proposal"
                :proposal-status="item.proposalStatus"
                :proposal-message="item.proposalMessage"
                :confirmed="item.proposalStatus === 'confirmed'"
                :original="props.context?.draft || null"
                @confirmed="handleProposalConfirmed(msg, item)"
              />
              <ChannelProposalCard
                v-else-if="item.proposal?.type === 'channel'"
                :proposal="item.proposal"
                :proposal-status="item.proposalStatus"
                :proposal-message="item.proposalMessage"
                :confirmed="item.proposalStatus === 'confirmed'"
                :original="props.context?.draft || null"
                @confirmed="handleProposalConfirmed(msg, item)"
              />
              <PresetProposalCard
                v-else-if="item.proposal?.type === 'preset'"
                :proposal="item.proposal"
                :proposal-status="item.proposalStatus"
                :proposal-message="item.proposalMessage"
                :confirmed="item.proposalStatus === 'confirmed'"
                :original="props.context?.draft || null"
                @confirmed="handleProposalConfirmed(msg, item)"
              />
            </template>

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

          <!-- 流式光标（正文下方） -->
          <div v-if="msg.streaming && msg.content" class="streaming-cursor"><span class="cursor"></span></div>

          <!-- 消息操作栏（完成后 hover 显示） -->
          <div v-if="!msg.streaming && (msg.content || msg.error)" class="msg-actions">
            <button v-if="msg.content" class="action-btn" title="复制" @click="copyMessage(msg)">
              <el-icon :size="13"><CopyDocument /></el-icon>
              复制
            </button>
            <button
              v-if="canRegenerate(msg, msgIndex)"
              class="action-btn"
              :title="msg.error ? '重试' : '重新生成'"
              @click="regenerate(msg, msgIndex)"
            >
              <el-icon :size="13"><RefreshRight /></el-icon>
              {{ msg.error ? '重试' : '重新生成' }}
            </button>
            <span v-if="msg.durationMs" class="action-meta">耗时 {{ formatDuration(msg.durationMs) }}</span>
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

    <!-- 回到底部悬浮按钮（用户上翻后出现） -->
    <transition name="fade">
      <button v-show="showScrollBtn" class="scroll-bottom-btn" title="回到底部" @click="scrollToBottom(true)">
        <el-icon :size="14"><ArrowDown /></el-icon>
      </button>
    </transition>
    </div>

    <!-- 输入区 -->
    <div class="input-area">
      <div class="composer" :class="{ sending }">
        <el-input
          ref="inputRef"
          v-model="input"
          type="textarea"
          :autosize="{ minRows: 2, maxRows: 6 }"
          resize="none"
          :disabled="sending"
          :placeholder="pageMeta.placeholder"
          @keydown="handleKeydown"
        />
        <div class="composer-footer">
          <span class="input-hint">
            <el-icon :size="12"><CircleCheck /></el-icon>
            提案需人工确认 · Enter 发送 / Shift+Enter 换行
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
import { ref, computed, nextTick, watch, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, ArrowDown, WarningFilled, Clock, Plus, Delete, Minus, Link, Search, EditPen, CircleCheck, CopyDocument, RefreshRight } from '@element-plus/icons-vue'
import { fetchSSEWithAuth } from '@/api/sseRequest'
import { defaultApi } from '@/api'
import { getRelativeTime } from '@/utils/timeUtils'
import { renderMarkdown, copyText } from '@/utils/markdown'
import ToolCallItem from '@/components/chat/ToolCallItem.vue'
import WebsiteProposalCard from './WebsiteProposalCard.vue'
import ModelProposalCard from './ModelProposalCard.vue'
import ChannelProposalCard from './ChannelProposalCard.vue'
import PresetProposalCard from './PresetProposalCard.vue'

const props = defineProps({
  // 页面上下文，注入后端系统提示词，如 { page: 'websites', draft: { url } }
  context: { type: Object, default: null }
})

const emit = defineEmits(['inserted'])

// ===== 页面差异化文案（空状态引导 / 输入框占位）=====
// page 与后端 internal/tool/ops 的 OpsPage 常量对应
const PAGE_META = {
  websites: {
    desc: '提供一个网站或 GitHub 仓库地址，我会自动抓取信息并生成录入提案，确认后才会写入数据库。',
    steps: [
      { icon: Link, label: '粘贴网址' },
      { icon: Search, label: 'AI 抓取解析' },
      { icon: CircleCheck, label: '确认入库' }
    ],
    examples: [
      '收录 https://github.com/cloudwego/eino',
      '收录 https://gorm.io/docs/'
    ],
    placeholder: '输入网站或 GitHub 地址，例如：https://github.com/cloudwego/eino'
  },
  models: {
    desc: '描述模型名称或特点，我会补全模型资料并生成录入提案；也可以优化已有模型的描述与标签，确认后才会写入数据库。',
    steps: [
      { icon: EditPen, label: '描述模型' },
      { icon: Search, label: 'AI 补全校验' },
      { icon: CircleCheck, label: '确认入库' }
    ],
    examples: [
      '新增一个支持联网搜索的 GLM-5 模型',
      '优化「DeepSeek」模型的介绍'
    ],
    placeholder: '描述要录入或优化的模型，例如：新增一个高上下文的 Kimi K2'
  },
  channels: {
    desc: '描述渠道信息，我会搭建渠道配置与模型映射提案（密钥需要你在确认卡片中手动填写）；也可以优化已有渠道配置，确认后才会写入数据库。',
    steps: [
      { icon: EditPen, label: '描述渠道' },
      { icon: Search, label: 'AI 搭建校验' },
      { icon: CircleCheck, label: '确认入库' }
    ],
    examples: [
      '新增一个 polo 渠道，支持 gpt-4o',
      '帮我优化名称含「polo」的渠道'
    ],
    placeholder: '描述要录入或优化的渠道，例如：新增一个火星引擎渠道'
  },
  presets: {
    desc: '描述助手定位，我会生成名称、简介、标签与系统提示词提案；也可以优化已有助手的上下文设定，确认后才会写入数据库。',
    steps: [
      { icon: EditPen, label: '描述助手' },
      { icon: Search, label: 'AI 生成校验' },
      { icon: CircleCheck, label: '确认入库' }
    ],
    examples: [
      '新增一个英文写作助手',
      '优化「编程助手」的上下文设定'
    ],
    placeholder: '描述要录入或优化的助手，例如：新增一个前端开发助手'
  }
}

const pageMeta = computed(() => PAGE_META[props.context?.page] || PAGE_META.websites)

// 会话列表的页面徽标文案（与后端 OpsPage 常量对应）
const PAGE_LABELS = {
  websites: '网站',
  models: '模型',
  channels: '渠道',
  presets: '助手'
}

// 提案确认提示文案（按提案类型）
const CONFIRM_TOASTS = {
  website: '网站已录入',
  model: '模型已保存',
  channel: '渠道已保存',
  preset: '助手已保存'
}
const CONFIRM_NOTICES = {
  website: '✅ 提案已确认，网站录入完成。',
  model: '✅ 提案已确认，模型保存完成。',
  channel: '✅ 提案已确认，渠道保存完成。',
  preset: '✅ 提案已确认，助手保存完成。'
}

// 对话消息：{ role, content, reasoning, reasoningExpanded, reasoningAuto, toolCalls[], proposals[], error, interrupted, streaming, durationMs }
// proposals 为本条消息携带的提案卡片数组 [{ proposal, proposalStatus, proposalMessage }]（批量录入一条消息多张卡片）
const messages = ref([])
const input = ref('')
const sending = ref(false)
const waitingFirst = ref(false)
const listRef = ref(null)
const inputRef = ref(null)

// 智能滚动：用户贴近底部时自动跟随；上翻阅读时不拽动，仅显示「回到底部」按钮
const autoFollow = ref(true)
const showScrollBtn = ref(false)

// 会话持久化状态
const currentSessionId = ref(0)
const sessionTitle = ref('')
const sessions = ref([])
const historyVisible = ref(false)
const historyKeyword = ref('')

// 历史会话按标题过滤
const filteredSessions = computed(() => {
  const kw = historyKeyword.value.trim().toLowerCase()
  if (!kw) return sessions.value
  return sessions.value.filter(s => (s.title || '').toLowerCase().includes(kw))
})

// 多轮历史由服务端持久化并构建，前端只传本次输入
let controller = null

const scrollToBottom = (smooth = false) => {
  nextTick(() => {
    const el = listRef.value
    if (!el) return
    if (smooth) {
      el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
    } else {
      el.scrollTop = el.scrollHeight
    }
  })
}

const handleListScroll = () => {
  const el = listRef.value
  if (!el) return
  autoFollow.value = el.scrollHeight - el.scrollTop - el.clientHeight < 80
  showScrollBtn.value = !autoFollow.value
}

// 消息长度变化时按需滚动（字段需判空：user 消息没有 toolCalls，读取 undefined 会中断调度队列导致渲染冻结）
watch(
  () => messages.value.map(m => (m.content || '').length + (m.reasoning || '').length + (m.toolCalls || []).length + (m.proposals || []).length),
  () => {
    if (autoFollow.value) {
      showScrollBtn.value = false
      scrollToBottom()
    } else if (sending.value) {
      // 流式期间用户上翻阅读：不打断，仅提示有新内容
      showScrollBtn.value = true
    }
  }
)

// ===== 会话历史 =====

// 统一消息形状：后端持久化消息 → 前端渲染所需字段（补齐缺省，toolCalls.params 映射为 args）
const normalizeMessage = (m) => ({
  role: m.role,
  content: m.content || '',
  reasoning: m.reasoning || '',
  reasoningExpanded: false,
  reasoningAuto: false,
  toolCalls: (m.toolCalls || []).map(tc => ({
    id: tc.id,
    name: tc.name,
    args: tc.params || '',
    result: tc.result || '',
    status: tc.status || 'completed'
  })),
  proposals: (m.proposals || []).map(p => ({
    proposal: p.proposal || null,
    proposalStatus: p.proposalStatus || '',
    proposalMessage: p.proposalMessage || ''
  })),
  error: m.error || '',
  interrupted: !!m.interrupted,
  confirmed: !!m.confirmed,
  durationMs: m.durationMs || 0,
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
    // 会话切换后重置滚动跟随（沿用旧状态会导致流式回复不自动滚动）
    autoFollow.value = true
    showScrollBtn.value = false
    scrollToBottom()
    return true
  } catch (error) {
    console.error('加载运营助手会话失败:', error)
    ElMessage.error('加载会话失败')
    return false
  }
}

// 挂载时自动续接最近会话：优先与当前页面同页的会话，避免跨页串台（如模型页恢复了网站页的对话）
const restoreLatestSession = async () => {
  try {
    const response = await defaultApi.apiAdminOpsChatSessionsListPost({ pageSize: 30 })
    if (response.code !== 0) return
    const list = response.data?.data || []
    const page = props.context?.page || ''
    const matched = page ? list.find(s => s.page === page) : list[0]
    if (matched) await loadSession(matched.id)
  } catch (error) {
    console.warn('恢复运营助手会话失败:', error)
  }
}

onMounted(() => {
  restoreLatestSession()
  // 抽屉打开（destroy-on-close 每次重建）自动聚焦输入框
  nextTick(() => inputRef.value?.focus?.())
})

// 打开历史弹层时懒更新列表并清空过滤（标题/排序可能已变化）
watch(historyVisible, (visible) => {
  if (visible) {
    historyKeyword.value = ''
    loadSessions()
  }
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
  autoFollow.value = true
  showScrollBtn.value = false
  historyVisible.value = false
  nextTick(() => inputRef.value?.focus?.())
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
      reasoningAuto: false,
      toolCalls: [],
      proposals: [],
      error: '',
      interrupted: false,
      confirmed: false,
      durationMs: 0,
      streaming: true
    }
    messages.value.push(msg)
  }
  return msg
}

// 思考过程折叠开关（用户手动操作后取消「自动收起」标记）
const toggleReasoning = (msg) => {
  msg.reasoningExpanded = !msg.reasoningExpanded
  msg.reasoningAuto = false
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
        // 思考阶段自动展开，正文出现后自动收起（用户手动操作过则不再干预）
        if (!msg.content) {
          msg.reasoningExpanded = true
          msg.reasoningAuto = true
        }
      }
      if (frame.content && msg.reasoningAuto) {
        msg.reasoningExpanded = false
        msg.reasoningAuto = false
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
      // 批量录入：提案卡片追加进当前流式消息的 proposals 数组（一条消息渲染多张卡片，不切断消息流）
      const msg = currentAssistant()
      msg.proposals.push({
        proposal: frame.proposal,
        proposalStatus: frame.status || 'ok',
        proposalMessage: frame.message || ''
      })
      waitingFirst.value = false
      break
    }
    case 'error': {
      const msg = currentAssistant()
      msg.error = frame.error || '执行失败'
      break
    }
    case 'end': {
      // 服务端回传总耗时，挂到当前助手消息上展示
      const msg = messages.value[messages.value.length - 1]
      if (msg && msg.role === 'assistant' && frame.durationMs) {
        msg.durationMs = frame.durationMs
      }
      finalizeAssistant()
      break
    }
    default:
      break
  }
}

const finalizeAssistant = () => {
  const msg = messages.value[messages.value.length - 1]
  if (msg && msg.role === 'assistant') {
    msg.streaming = false
    // 空消息（无内容无工具无提案）直接移除
    if (!msg.content && !msg.reasoning && !msg.toolCalls.length && !(msg.proposals && msg.proposals.length) && !msg.error && !msg.interrupted) {
      messages.value.pop()
    }
  }
  sending.value = false
  waitingFirst.value = false
  controller = null
  // 回归输入框，方便连续追问（主流对话产品的默认行为）
  nextTick(() => inputRef.value?.focus?.())
}

// 复制消息正文
const copyMessage = async (msg) => {
  const ok = await copyText(msg.content || '')
  if (ok) {
    ElMessage.success('已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

// 耗时格式化：60s 内显示秒，超过显示分秒
const formatDuration = (ms) => {
  if (!ms || ms <= 0) return ''
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  const m = Math.floor(ms / 60000)
  const s = Math.round((ms % 60000) / 1000)
  return `${m}m${s}s`
}

// 回退中标记：防止回退请求期间重复点击导致多轮回退
const rewinding = ref(false)

// 是否可重新生成：仅最后一条助手消息、非确认提示、无已确认提案、本轮存在用户提问
const canRegenerate = (msg, index) =>
  !sending.value &&
  !rewinding.value &&
  index === messages.value.length - 1 &&
  !msg.confirmed &&
  !(msg.proposals || []).some(p => p.proposalStatus === 'confirmed') &&
  messages.value.slice(0, index).some(m => m.role === 'user')

// 重新生成/失败重试：服务端回退最后一轮问答（避免历史与界面出现重复轮次），再重发原问题
const regenerate = async (msg, index) => {
  if (sending.value || rewinding.value) return
  // 定位本轮的用户提问（该助手消息前最近的用户消息）
  let userIdx = -1
  for (let i = index - 1; i >= 0; i--) {
    if (messages.value[i].role === 'user') {
      userIdx = i
      break
    }
  }
  if (userIdx < 0 || !messages.value[userIdx].content) return
  const question = messages.value[userIdx].content

  rewinding.value = true
  try {
    if (currentSessionId.value) {
      const response = await defaultApi.apiAdminOpsChatSessionsIdAppendPost(currentSessionId.value, {
        role: 'assistant',
        content: '',
        rewindLastRound: true
      })
      if (response.code !== 0) {
        ElMessage.error(response.message || '操作失败')
        return
      }
    }
    // 本地同步移除该轮（从用户提问起全部移除），随后按普通发送流程重发
    messages.value = messages.value.slice(0, userIdx)
    sendMessage(question)
  } catch (error) {
    console.error('回退运营助手会话失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  } finally {
    rewinding.value = false
  }
}

// 键盘发送：Enter 发送，Shift+Enter 换行；
// 输入法组词中的回车（isComposing/keyCode 229）是确认候选词，不能触发发送
const handleKeydown = (e) => {
  if (e.isComposing || e.keyCode === 229) return
  if (e.key === 'Enter' && !e.shiftKey && !e.ctrlKey && !e.metaKey && !e.altKey) {
    e.preventDefault()
    handleSend()
  }
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
  reasoningAuto: false,
  toolCalls: [],
  proposals: [],
  error: '',
  interrupted: false,
  confirmed: false,
  durationMs: 0,
  streaming: false
})

const sendMessage = async (text) => {
  if (sending.value) return
  messages.value.push(createUserMessage(text))
  sending.value = true
  waitingFirst.value = true
  // 用户主动发送后强制跟随滚动，保证提问与回答可见
  autoFollow.value = true
  showScrollBtn.value = false
  scrollToBottom(true)

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

// msg / item 由模板传入（confirm 事件所在的消息与提案项），批量提案时精确标记被确认的那张
const handleProposalConfirmed = (msg, item) => {
  const proposalType = item?.proposal?.type || 'website'
  const toast = CONFIRM_TOASTS[proposalType] || CONFIRM_TOASTS.website
  const notice = CONFIRM_NOTICES[proposalType] || CONFIRM_NOTICES.website

  ElMessage.success(toast)
  emit('inserted')
  // 本地置为已确认，防止回放后二次确认
  if (item) {
    item.proposalStatus = 'confirmed'
  }
  // 追加一条本地提示，明确提案已完成
  messages.value.push({ role: 'assistant', content: notice, streaming: false, toolCalls: [], proposals: [], reasoning: '', reasoningExpanded: false, reasoningAuto: false, error: '', interrupted: false, confirmed: true, durationMs: 0 })
  scrollToBottom()

  // 持久化确认状态（失败不影响本地 UI，数据库侧有重名/URL 查重兜底）；
  // proposalName 让后端在多条提案中精确匹配被确认的那张
  if (currentSessionId.value) {
    defaultApi
      .apiAdminOpsChatSessionsIdAppendPost(currentSessionId.value, {
        role: 'assistant',
        content: notice,
        markProposalConfirmed: true,
        proposalName: item?.proposal?.name || ''
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

// 列表外层：承载「回到底部」按钮的定位
.list-wrap {
  flex: 1;
  min-height: 0;
  position: relative;
  display: flex;
  flex-direction: column;
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

// 回到底部悬浮按钮
.scroll-bottom-btn {
  position: absolute;
  right: 18px;
  bottom: 14px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  color: var(--el-text-color-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 5;
  transition: color 0.15s ease, border-color 0.15s ease;

  &:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
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

// 用户消息 hover 复制按钮
.message-row.user {
  .copy-btn {
    margin-left: 6px;
    align-self: center;
    width: 26px;
    height: 26px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--el-text-color-secondary);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.15s ease, color 0.15s ease, background 0.15s ease;

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }

  &:hover .copy-btn {
    opacity: 1;
  }
}

// 助手消息主体（气泡 + 底部操作栏）
.assistant-main {
  min-width: 0;
  max-width: min(92%, 640px);
  display: flex;
  flex-direction: column;

  // 宽度上限移到 main 上统一控制，气泡随内容撑满 main
  .assistant-bubble {
    max-width: 100%;
  }
}

// 消息操作栏（hover 行显示）
.msg-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  opacity: 0;
  transition: opacity 0.15s ease;

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    border: none;
    background: transparent;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    padding: 3px 7px;
    border-radius: 6px;
    cursor: pointer;
    transition: color 0.15s ease, background 0.15s ease;

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }

  .action-meta {
    margin-left: 4px;
    font-size: 11px;
    color: var(--el-text-color-placeholder);
  }
}

.message-row.assistant:hover .msg-actions {
  opacity: 1;
}

// 流式光标行（正文下方）
.streaming-cursor {
  padding: 3px 4px 0;
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

// 正文（markdown 渲染）：空白由 markdown 排版接管
// v-html 注入的 DOM 无 scope 属性，需用 :deep 匹配
.msg-content {
  white-space: normal;

  :deep(.markdown-body) {
    font-size: 14px;
    line-height: 1.65;
    color: var(--el-text-color-primary);
    word-break: break-word;

    > *:first-child { margin-top: 0; }
    > *:last-child { margin-bottom: 0; }

    p {
      margin: 0 0 8px;
    }

    h1, h2, h3, h4, h5, h6 {
      margin: 14px 0 8px;
      font-weight: 600;
      line-height: 1.4;

      &:first-child { margin-top: 0; }
    }

    h1 { font-size: 17px; }
    h2 { font-size: 16px; }
    h3 { font-size: 15px; }
    h4, h5, h6 { font-size: 14px; }

    ul, ol {
      margin: 4px 0 10px;
      padding-left: 22px;

      li {
        margin: 3px 0;

        &::marker {
          color: var(--el-text-color-secondary);
        }
      }

      p {
        margin: 0 0 4px;
      }
    }

    // 行内代码
    code:not(pre code) {
      padding: 1px 6px;
      margin: 0 1px;
      border-radius: 5px;
      font-size: 13px;
      font-family: 'JetBrains Mono', 'Fira Code', Consolas, Monaco, monospace;
      background: var(--el-fill-color);
      color: var(--el-color-primary);
    }

    // 代码块（结构与 chat 页一致：语言标签 + 复制按钮）
    pre.code-block {
      background: #282c34;
      margin: 10px 0;
      padding: 0;
      border-radius: 8px;
      overflow: hidden;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
      display: flex;
      flex-direction: column;

      .code-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0 8px 0 12px;
        background: #21252b;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        height: 34px;

        .code-lang {
          color: #abb2bf;
          font-size: 11px;
          font-weight: 500;
          background: rgba(255, 255, 255, 0.1);
          padding: 1px 7px;
          border-radius: 4px;
          letter-spacing: 0.5px;
        }

        .copy-button {
          background: transparent;
          border: none;
          color: #abb2bf;
          padding: 4px 10px;
          font-size: 12px;
          border-radius: 4px;
          cursor: pointer;
          transition: all 0.2s ease;

          &:hover {
            background: rgba(255, 255, 255, 0.08);
            color: #fff;
          }
        }
      }

      code {
        display: block;
        padding: 12px 14px;
        overflow-x: auto;
        font-family: 'JetBrains Mono', 'Fira Code', Consolas, Monaco, monospace;
        font-size: 13px;
        line-height: 1.6;
        background: transparent;
      }
    }

    blockquote {
      margin: 8px 0;
      padding: 2px 12px;
      border-left: 3px solid var(--el-border-color);
      color: var(--el-text-color-secondary);

      p {
        margin: 4px 0;
      }
    }

    table {
      margin: 8px 0;
      border-collapse: collapse;
      font-size: 13px;
      display: block;
      max-width: 100%;
      overflow-x: auto;

      th, td {
        border: 1px solid var(--el-border-color-lighter);
        padding: 5px 10px;
        text-align: left;
      }

      th {
        background: var(--el-fill-color-light);
        font-weight: 600;
        white-space: nowrap;
      }
    }

    a {
      color: var(--el-color-primary);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }

    hr {
      margin: 12px 0;
      border: none;
      border-top: 1px solid var(--el-border-color-lighter);
    }

    img {
      max-width: 100%;
      border-radius: 8px;
    }
  }
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

    .scroll-bottom-btn {
      border-color: #363637;
      box-shadow: none;
    }

    .message-row.user .copy-btn,
    .msg-actions .action-btn {
      &:hover {
        background: rgba(255, 255, 255, 0.08);
      }
    }
  }
}
</style>

<style lang="scss">
// 历史会话弹层（el-popover teleport 到 body，scoped 样式不生效，需全局；颜色全部用主题变量以适配暗色）
.ops-history-popover {
  .history-search {
    margin: -6px -12px 0;
    padding: 8px 12px;
    border-bottom: 1px solid var(--el-border-color-extra-light);
  }

  .history-list {
    max-height: 320px;
    overflow-y: auto;
    margin: 0 -12px -6px;

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
      display: flex;
      align-items: center;
      gap: 6px;
      min-width: 0;
      font-size: 13px;
      color: var(--el-text-color-primary);

      .history-title-text {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .history-page-tag {
        flex-shrink: 0;
        font-size: 10px;
        line-height: 1;
        padding: 3px 6px;
        border-radius: 999px;
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
      }
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
