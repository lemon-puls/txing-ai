<template>
  <div class="chat-container" :class="{ 'dark-theme': isDarkTheme }">
    <!-- 左侧会话列表 -->
    <div class="sidebar" :class="{ 'sidebar-collapsed': isSidebarCollapsed }">
      <!-- 新建会话按钮 -->
      <div class="new-chat-btn" @click="createNewChat">
        <el-button type="primary" class="w-full new-chat-button">
          <template #icon><Plus class="icon-bounce" /></template>
          新建会话
        </el-button>
      </div>

      <!-- 会话列表 -->
      <div class="chat-list custom-scrollbar">
        <TransitionGroup name="chat-item">
          <div
            v-for="chat in chatList"
            :key="chat.id"
            class="chat-item"
            :class="{ active: currentChat?.id === chat.id }"
            @click="switchChat(chat)"
          >
            <div class="chat-item-content">
              <div class="chat-icon-wrapper">
                <el-icon class="chat-icon"><ChatRound /></el-icon>
              </div>
              <div class="chat-info">
                <div class="chat-title">{{ chat.title }}</div>
                <div class="chat-preview">{{ chat.lastMessage }}</div>
              </div>
            </div>
            <div class="chat-actions">
              <el-dropdown trigger="hover" @command="handleChatAction($event, chat)">
                <el-icon><More /></el-icon>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="rename">重命名</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除会话</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </TransitionGroup>
      </div>

      <!-- 底部操作区 -->
      <div class="sidebar-footer">
        <el-tooltip content="切换主题" placement="top">
          <div class="theme-toggle" @click="toggleTheme">
            <el-icon><component :is="isDarkTheme ? 'Sunny' : 'Moon'" /></el-icon>
          </div>
        </el-tooltip>
        <div class="sidebar-toggle" @click="toggleSidebar">
          <el-icon><Fold /></el-icon>
        </div>
      </div>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-main">
      <template v-if="currentChat">
        <!-- 聊天头部 -->
        <div class="chat-header">
          <div class="chat-title">
            <span>{{ currentChat.title }}</span>
            <el-tag size="small" effect="plain" class="ml-2 model-tag">{{ currentChat.model }}</el-tag>
          </div>
          <div class="chat-settings">
            <el-tooltip content="模型设置" placement="bottom">
              <el-button circle @click="showSettings = true">
                <el-icon><Setting /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </div>

        <!-- 聊天消息区域 -->
        <div class="chat-messages custom-scrollbar" ref="messagesContainer">
          <TransitionGroup name="message">
            <div
              v-for="message in currentChat.messages"
              :key="message.id"
              class="message-item"
              :class="message.role"
            >
              <div class="message-avatar">
                <el-avatar :size="40" :src="message.role === 'user' ? userAvatar : aiAvatar">
                  {{ message.role === 'user' ? 'U' : 'AI' }}
                </el-avatar>
              </div>
              <div class="message-content">
                <div class="message-text" v-html="formatMessage(message.content)"></div>
                <div class="message-actions">
                  <el-button-group>
                    <el-button text size="small" @click="copyMessage(message)">
                      <template #icon><CopyDocument /></template>
                      复制
                    </el-button>
                    <el-button text size="small" @click="regenerateMessage(message)" v-if="message.role === 'assistant'">
                      <template #icon><RefreshRight /></template>
                      重新生成
                    </el-button>
                  </el-button-group>
                </div>
              </div>
            </div>
          </TransitionGroup>
          
          <!-- 输入提示 -->
          <div v-if="isTyping" class="typing-indicator">
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
          </div>
        </div>

        <!-- 聊天输入区域 -->
        <div class="chat-input">
          <div class="input-wrapper">
            <el-input
              v-model="messageInput"
              type="textarea"
              :rows="3"
              placeholder="输入消息，Enter 发送，Shift + Enter 换行"
              resize="none"
              @keydown.enter.exact.prevent="sendMessage"
              class="custom-input"
            />
            <div class="input-actions">
              <el-button-group>
                <el-tooltip content="上传文件" placement="top">
                  <el-button circle>
                    <el-icon><Upload /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="发送消息" placement="top">
                  <el-button type="primary" circle @click="sendMessage" class="send-button">
                    <el-icon><Position /></el-icon>
                  </el-button>
                </el-tooltip>
              </el-button-group>
            </div>
          </div>
        </div>
      </template>

      <!-- 空状态 -->
      <div v-else class="chat-empty">
        <el-empty description="选择或创建一个会话开始聊天">
          <el-button type="primary" @click="createNewChat" class="create-button">
            <template #icon><Plus /></template>
            新建会话
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 设置对话框 -->
    <el-dialog
      v-model="showSettings"
      title="会话设置"
      width="500px"
      destroy-on-close
      class="settings-dialog"
    >
      <div class="settings-content">
        <el-form label-position="top">
          <el-form-item label="选择模型">
            <el-select v-model="currentChat.model" class="w-full">
              <el-option label="GPT-3.5" value="gpt-3.5-turbo" />
              <el-option label="GPT-4" value="gpt-4" />
              <el-option label="Claude" value="claude" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-switch
              v-model="currentChat.webSearch"
              active-text="启用联网搜索"
            />
          </el-form-item>
          <el-form-item label="温度">
            <el-slider
              v-model="currentChat.temperature"
              :min="0"
              :max="2"
              :step="0.1"
              show-input
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="showSettings = false">取消</el-button>
        <el-button type="primary" @click="saveSettings">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="ChatPage">
import { ref, nextTick, onMounted } from 'vue'
import { 
  Plus, ChatRound, More, Fold, Setting, 
  CopyDocument, RefreshRight, Upload, Position,
  Sunny, Moon
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// 配置 marked
marked.setOptions({
  highlight: function(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  },
  breaks: true
})

// 状态
const isDarkTheme = ref(window.matchMedia('(prefers-color-scheme: dark)').matches)
const isSidebarCollapsed = ref(false)
const showSettings = ref(false)
const messageInput = ref('')
const isTyping = ref(false)
const messagesContainer = ref(null)
const currentChat = ref(null)
const chatList = ref([
  {
    id: 1,
    title: '新对话',
    model: 'gpt-3.5-turbo',
    webSearch: false,
    temperature: 1,
    lastMessage: '你好！我是 AI 助手，有什么我可以帮你的吗？',
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: '你好！我是 AI 助手，有什么我可以帮你的吗？'
      }
    ]
  }
])

// 头像
const userAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'
const aiAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

// 主题切换
const toggleTheme = () => {
  isDarkTheme.value = !isDarkTheme.value
  document.documentElement.classList.toggle('dark', isDarkTheme.value)
}

// 方法
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const createNewChat = () => {
  const newChat = {
    id: Date.now(),
    title: '新对话',
    model: 'gpt-3.5-turbo',
    webSearch: false,
    temperature: 1,
    lastMessage: '你好！我是 AI 助手，有什么我可以帮你的吗？',
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: '你好！我是 AI 助手，有什么我可以帮你的吗？'
      }
    ]
  }
  chatList.value.unshift(newChat)
  currentChat.value = newChat
}

const switchChat = (chat) => {
  currentChat.value = chat
  scrollToBottom()
}

const handleChatAction = (command, chat) => {
  if (command === 'delete') {
    const index = chatList.value.findIndex(c => c.id === chat.id)
    if (index > -1) {
      chatList.value.splice(index, 1)
      if (currentChat.value?.id === chat.id) {
        currentChat.value = chatList.value[0]
      }
    }
  }
}

const formatMessage = (content) => {
  return marked(content)
}

const copyMessage = async (message) => {
  try {
    await navigator.clipboard.writeText(message.content)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const regenerateMessage = () => {
  isTyping.value = true
  // 模拟重新生成
  setTimeout(() => {
    isTyping.value = false
  }, 2000)
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const sendMessage = async () => {
  if (!messageInput.value.trim()) return
  
  // 添加用户消息
  currentChat.value.messages.push({
    id: Date.now(),
    role: 'user',
    content: messageInput.value
  })
  
  currentChat.value.lastMessage = messageInput.value
  messageInput.value = ''
  await scrollToBottom()

  // 模拟 AI 响应
  isTyping.value = true
  setTimeout(() => {
    currentChat.value.messages.push({
      id: Date.now(),
      role: 'assistant',
      content: '这是一个模拟的回复，包含代码示例：\n```javascript\nconsole.log("Hello World!");\n```'
    })
    isTyping.value = false
    scrollToBottom()
  }, 1000)
}

const saveSettings = () => {
  showSettings.value = false
  ElMessage.success('设置已保存')
}

// 监听系统主题变化
onMounted(() => {
  currentChat.value = chatList.value[0]
  
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', e => {
    isDarkTheme.value = e.matches
    document.documentElement.classList.toggle('dark', e.matches)
  })
})
</script>

<style scoped lang="scss">
// 主题变量
:root {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f7fa;
  --text-primary: #333333;
  --text-secondary: #666666;
  --border-color: #e6e6e6;
  --hover-bg: #f5f7fa;
  --active-bg: #ecf5ff;
  --shadow-color: rgba(0, 0, 0, 0.1);
  --message-bg-user: #ecf5ff;
  --message-bg-assistant: #f5f7fa;
  --scrollbar-thumb: #c0c4cc;
  --scrollbar-track: #f5f7fa;
}

.dark-theme {
  --bg-primary: #1a1a1a;
  --bg-secondary: #2d2d2d;
  --text-primary: #ffffff;
  --text-secondary: #a0a0a0;
  --border-color: #3a3a3a;
  --hover-bg: #2d2d2d;
  --active-bg: #363636;
  --shadow-color: rgba(0, 0, 0, 0.3);
  --message-bg-user: #363636;
  --message-bg-assistant: #2d2d2d;
  --scrollbar-thumb: #4a4a4a;
  --scrollbar-track: #2d2d2d;

  :deep(.el-button) {
    --el-button-bg-color: #363636;
    --el-button-border-color: #4a4a4a;
    --el-button-hover-bg-color: #4a4a4a;
    --el-button-hover-border-color: #5a5a5a;
  }

  :deep(.el-input__wrapper) {
    background-color: var(--bg-secondary);
    border-color: var(--border-color);
  }

  :deep(.el-input__inner) {
    color: var(--text-primary);
  }
}

.chat-container {
  display: flex;
  height: 100vh;
  background: var(--bg-secondary);
  color: var(--text-primary);
  transition: all 0.3s ease;
}

// 自定义滚动条
.custom-scrollbar {
  &::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--scrollbar-thumb);
    border-radius: 3px;
  }

  &::-webkit-scrollbar-track {
    background: var(--scrollbar-track);
  }
}

.sidebar {
  width: 300px;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &.sidebar-collapsed {
    width: 0;
    overflow: hidden;
  }
}

.new-chat-btn {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);

  .new-chat-button {
    transition: transform 0.3s ease;

    &:hover {
      transform: translateY(-2px);
    }

    .icon-bounce {
      animation: iconBounce 1s infinite;
    }
  }
}

.chat-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.chat-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-bottom: 4px;
  border: 1px solid transparent;

  &:hover {
    background: var(--hover-bg);
    border-color: var(--border-color);
    transform: translateX(4px);
  }

  &.active {
    background: var(--active-bg);
    border-color: var(--border-color);
  }

  .chat-icon-wrapper {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    background: linear-gradient(135deg, #4158D0, #C850C0);
    
    .chat-icon {
      font-size: 20px;
      color: white;
    }
  }

  .chat-item-content {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .chat-info {
    flex: 1;
    min-width: 0;
  }

  .chat-title {
    font-weight: 500;
    margin-bottom: 4px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-preview {
    font-size: 12px;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-actions {
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  &:hover .chat-actions {
    opacity: 1;
  }
}

.sidebar-footer {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--border-color);

  .theme-toggle, .sidebar-toggle {
    padding: 8px;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.3s ease;

    &:hover {
      background: var(--hover-bg);
      transform: scale(1.1);
    }
  }
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  position: relative;
}

.chat-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-primary);
  z-index: 1;

  .chat-title {
    display: flex;
    align-items: center;
    font-size: 16px;
    font-weight: 500;

    .model-tag {
      background: linear-gradient(135deg, #4158D0, #C850C0);
      color: white;
      border: none;
    }
  }
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  scroll-behavior: smooth;
}

.message-item {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  opacity: 0;
  transform: translateY(20px);
  animation: message-fade-in 0.3s ease forwards;

  &.user {
    flex-direction: row-reverse;

    .message-content {
      background: var(--message-bg-user);
      border-radius: 12px 2px 12px 12px;
    }

    .message-actions {
      justify-content: flex-start;
    }
  }

  &.assistant .message-content {
    background: var(--message-bg-assistant);
    border-radius: 2px 12px 12px 12px;
  }
}

.message-content {
  max-width: 80%;
  padding: 16px;
  box-shadow: 0 2px 8px var(--shadow-color);
  transition: transform 0.3s ease;

  &:hover {
    transform: translateY(-2px);
  }

  .message-text {
    line-height: 1.6;
    white-space: pre-wrap;

    :deep(pre) {
      background: #282c34;
      padding: 16px;
      border-radius: 8px;
      margin: 8px 0;
      overflow-x: auto;
    }

    :deep(code) {
      font-family: 'Fira Code', monospace;
    }

    :deep(p) {
      margin: 8px 0;
    }
  }
}

.message-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.message-item:hover .message-actions {
  opacity: 1;
}

.chat-input {
  padding: 16px 24px;
  background: var(--bg-primary);
  border-top: 1px solid var(--border-color);
}

.input-wrapper {
  position: relative;
  
  .input-actions {
    position: absolute;
    right: 8px;
    bottom: 8px;
  }

  .custom-input {
    transition: all 0.3s ease;

    &:focus {
      transform: translateY(-2px);
    }
  }

  .send-button {
    background: linear-gradient(135deg, #4158D0, #C850C0);
    border: none;
    transition: all 0.3s ease;

    &:hover {
      transform: scale(1.1);
    }
  }
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 16px;
  align-items: center;
}

.typing-dot {
  width: 8px;
  height: 8px;
  background: #409eff;
  border-radius: 50%;
  animation: typing 1s infinite;

  &:nth-child(2) {
    animation-delay: 0.2s;
  }

  &:nth-child(3) {
    animation-delay: 0.4s;
  }
}

.chat-empty {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;

  .create-button {
    background: linear-gradient(135deg, #4158D0, #C850C0);
    border: none;
    transition: transform 0.3s ease;

    &:hover {
      transform: scale(1.1);
    }
  }
}

// 动画
@keyframes message-fade-in {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes typing {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

@keyframes iconBounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

.chat-item-enter-active,
.chat-item-leave-active {
  transition: all 0.3s ease;
}

.chat-item-enter-from,
.chat-item-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.message-enter-active,
.message-leave-active {
  transition: all 0.3s ease;
}

.message-enter-from,
.message-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

// 响应式设计
@media screen and (max-width: 768px) {
  .sidebar {
    position: absolute;
    height: 100%;
    z-index: 2;
  }

  .chat-main {
    width: 100%;
  }

  .message-content {
    max-width: 90%;
  }
}
</style> 