<template>
  <div class="chat-container" :class="[
    { 'dark-theme': isDarkTheme },
    `bg-pattern-${currentBgPattern}`
  ]">
    <!-- 左侧会话列表 -->
    <div class="sidebar" :class="{ 'sidebar-collapsed': isSidebarCollapsed }">
      <!-- 新建会话按钮 -->
      <div class="action-buttons">
        <el-tooltip content="返回首页" placement="right">
          <el-button circle class="home-button" @click="goToHome">
            <el-icon>
              <HomeFilled/>
            </el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="新建会话" placement="right">
          <el-button type="primary" circle class="new-chat-button" @click="createNewChat('')">
            <el-icon>
              <Plus class="icon-bounce"/>
            </el-icon>
          </el-button>
        </el-tooltip>
      </div>

      <!-- 会话列表 -->
      <div class="chat-list custom-scrollbar">
        <TransitionGroup name="chat-item">
          <div
            v-for="chat in chatList"
            :key="chat.id"
            class="chat-item"
            :class="{
              active: currentChat?.id === chat.id,
              selected: selectedChats.includes(chat.id)
            }"
            @click="handleChatClick(chat)"
          >
            <div class="chat-item-content">
              <el-checkbox
                v-if="showCheckboxes"
                v-model="selectedChats"
                :value="chat.id"
                @click.stop
              />
              <div class="chat-icon-wrapper">
                <el-avatar
                  :size="40"
                  :src="chat.avatar"
                >
                  {{ chat.model?.charAt(0) }}
                </el-avatar>
              </div>
              <div class="chat-info">
                <div class="chat-title">{{ chat.name || chat.messages?.[0]?.content }}</div>
              </div>
            </div>
            <div class="chat-actions">
              <el-dropdown trigger="hover" @command="handleChatAction($event, chat)">
                <el-icon>
                  <More/>
                </el-icon>
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

      <!-- 批量操作栏（底部固定，批量模式下显示） -->
      <transition name="batch-bar-fade">
        <div class="batch-actions-bar" v-if="showCheckboxes">
          <el-button type="danger" size="small" class="batch-btn" @click="batchDelete">
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
          <el-button size="small" class="batch-btn cancel-btn" @click="showCheckboxes = false">
            <el-icon><Close /></el-icon>
            取消
          </el-button>
        </div>
      </transition>

      <!-- 底部操作区 -->
      <div class="sidebar-footer">
        <div class="footer-actions">
          <el-tooltip content="切换主题" placement="top">
            <div class="action-btn" @click="toggleTheme">
              <SvgIcon icon="theme" size="20" hover click/>
            </div>
          </el-tooltip>
          <el-tooltip content="切换背景" placement="top">
            <div class="action-btn" @click="showBgPatternSelector">
              <SvgIcon icon="pic" color="#1890ff" size="20" hover click/>
            </div>
          </el-tooltip>
          <el-tooltip content="批量操作" placement="top">
            <div class="action-btn" @click="toggleBatchMode">
              <SvgIcon icon="deletebatch" size="18" hover click/>
            </div>
          </el-tooltip>
        </div>
        <div class="sidebar-toggle" @click="toggleSidebar">
          <el-icon>
            <Fold/>
          </el-icon>
        </div>
      </div>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-main">
      <!-- 添加展开按钮 -->
      <div
        class="sidebar-expand"
        :class="{ show: isSidebarCollapsed }"
        @click="toggleSidebar"
      >
        <el-icon>
          <Fold class="expand-icon"/>
        </el-icon>
      </div>

      <template v-if="currentChat">
        <!-- 聊天头部 -->
        <div class="chat-header">
          <div class="chat-title">
            <span>{{ currentChat.name }}</span>
            <!--            <el-tag size="small" effect="plain" class="ml-2 model-tag">{{ currentChat.model }}</el-tag>-->
          </div>
          <div class="chat-settings">
            <el-tooltip content="模型设置" placement="bottom">
              <el-button circle @click="showSettings = true">
                <el-icon>
                  <Setting/>
                </el-icon>
              </el-button>
            </el-tooltip>
            <UserAvatar/>
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
                <el-avatar
                  :size="40"
                  :src="message.role === 'user' ? userAvatar : (currentChat.preset?.avatar || aiAvatar)"
                >
                  {{ message.role === 'user' ? 'U' : (currentChat.preset?.name?.charAt(0) || 'AI') }}
                </el-avatar>
              </div>
              <div class="message-content">
                <!-- 添加思考过程组件 -->
                <div v-if="message.reasoningContent" class="thought-process">
                  <div class="thought-header" @click="toggleThought(message)">
                    <el-icon :class="{ 'is-fold': !message.showThought }">
                      <ArrowRight/>
                    </el-icon>
                    <span>已深度思考 {{
                        isCurrentStreamingMessage(message) ?
                          `(用时${messageThoughtTimes.get(message.id)?.duration || 0}秒)` :
                          messageThoughtTimes.has(message.id) ?
                            `(用时${messageThoughtTimes.get(message.id).duration}秒)` :
                            ''
                      }}</span>
                  </div>
                  <div v-show="message.showThought" class="thought-content">
                    {{ message.reasoningContent }}
                  </div>
                </div>
                <div class="message-text" v-html="renderMessage(message.content)"></div>
                <div class="message-actions">
                  <el-button-group>
                    <el-button text size="small" @click="copyMessage(message)">
                      <template #icon>
                        <CopyDocument/>
                      </template>
                      复制
                    </el-button>
                    <el-button text size="small" @click="regenerateMessage(message)"
                               v-if="message.role === 'assistant'">
                      <template #icon>
                        <RefreshRight/>
                      </template>
                      重新生成
                    </el-button>
                  </el-button-group>
                </div>
              </div>
            </div>
          </TransitionGroup>

          <!-- 点点加载动画，紧贴下一个 assistant 消息顶部左侧 -->
          <div v-if="currentChat && messageLoadingMap.get(currentChat.id)" class="dot-loading-indicator">
            <div class="dot-typing">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
          </div>

          <!-- 输入提示 -->
          <div v-if="isTyping" class="typing-indicator">
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
            <div class="typing-dot"></div>
          </div>
        </div>

        <!-- 聊天输入区域 -->
        <div class="chat-input">
          <div class="resize-handle" @mousedown="startResize"></div>
          <div class="quick-settings">
            <div class="model-selector">
              <el-popover
                placement="top"
                :width="300"
                trigger="click"
                popper-class="model-popover"
              >
                <template #reference>
                  <div class="current-model">
                    <div class="model-icon">
                      <el-avatar :size="24" :src="currentModel?.avatar">
                        {{ currentModel?.name?.charAt(0) }}
                      </el-avatar>
                    </div>
                    <span class="model-name" v-text="currentModel?.name"></span>
                    <el-icon class="arrow-icon">
                      <ArrowDown/>
                    </el-icon>
                  </div>
                </template>
                <div class="model-list">
                  <el-empty v-if="availableModels.length === 0 && !loadingModels" description="暂无可用模型"/>
                  <el-skeleton v-else-if="loadingModels" :rows="3" animated/>
                  <template v-else>
                    <div
                      v-for="model in availableModels"
                      :key="model.name"
                      class="model-item"
                      :class="{ active: currentChat.model === model.name }"
                      @click="selectModel(model)"
                    >
                      <div class="model-item-icon" :class="model.tag">
                        <el-avatar :size="24" :src="model.avatar">
                          {{ model.name.charAt(0) }}
                        </el-avatar>
                      </div>
                      <div class="model-item-info">
                        <div class="model-item-name">{{ model.name }}</div>
                        <div class="model-item-desc">{{ model.description }}</div>
                      </div>
                      <el-icon v-if="currentChat.model === model.name">
                        <Check/>
                      </el-icon>
                    </div>
                  </template>
                </div>
              </el-popover>
            </div>
            <div class="quick-actions">
              <el-tooltip content="AI 助手市场" placement="top">
                <div class="action-btn" @click="showPresetMarket = true">
                  <SvgIcon icon="ai" size="30" hover click/>
                </div>
              </el-tooltip>
              <div class="feature-toggles">
                <el-tooltip content="联网搜索" placement="top">
                  <div
                    class="feature-toggle"
                    @click="toggleWebSearch"
                  >
                    <SvgIcon v-if="currentChat.webSearch" icon="network-active" size="28" hover click/>
                    <SvgIcon v-else icon="network" size="24" hover click/>
                  </div>
                </el-tooltip>
              </div>
            </div>
          </div>
          <div class="input-wrapper">
            <el-input
              v-model="messageInput"
              type="textarea"
              :rows="textareaRows"
              placeholder="输入消息，Enter 发送，Shift + Enter 换行"
              resize="none"
              @keydown.enter.exact.prevent="sendMessage"
              class="custom-input"
            />
            <div class="input-actions">
              <el-button-group>
                <el-tooltip content="停止生成" placement="top" v-if="isTyping">
                  <el-button circle @click="stopGeneration">
                    <el-icon>
                      <CircleClose/>
                    </el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="发送消息" placement="top">
                  <el-button type="primary" circle @click="sendMessage" class="send-button">
                    <el-icon>
                      <Position/>
                    </el-icon>
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
          <el-button type="primary" @click="createNewChat('')" class="create-button">
            <template #icon>
              <Plus/>
            </template>
            新建会话
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 设置对话框 -->
    <el-dialog
      v-model="showSettings"
      title="高级参数设置"
      width="500px"
      destroy-on-close
      class="settings-dialog"
    >
      <div class="settings-content">
        <el-form label-position="top">
          <el-form-item label="最大Token数">
            <el-input-number
              v-model="currentChat.maxTokens"
              :min="1"
              :max="4096"
              class="w-full"
            />
          </el-form-item>
          <el-form-item label="温度">
            <el-slider
              v-model="currentChat.temperature"
              :min="0"
              :max="2"
              :step="0.1"
              :default-value="0"
              show-input
            />
          </el-form-item>
          <el-form-item label="Top-P采样">
            <el-slider
              v-model="currentChat.topP"
              :min="0"
              :max="1"
              :step="0.05"
              show-input
            />
          </el-form-item>
          <el-form-item label="Top-K采样">
            <el-input-number
              v-model="currentChat.topK"
              :min="1"
              :max="100"
              class="w-full"
            />
          </el-form-item>
          <el-form-item label="存在惩罚">
            <el-slider
              v-model="currentChat.presencePenalty"
              :min="-2"
              :max="2"
              :step="0.1"
              show-input
            />
          </el-form-item>
          <el-form-item label="频率惩罚">
            <el-slider
              v-model="currentChat.frequencyPenalty"
              :min="-2"
              :max="2"
              :step="0.1"
              show-input
            />
          </el-form-item>
          <el-form-item label="重复惩罚">
            <el-slider
              v-model="currentChat.repetitionPenalty"
              :min="1"
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

    <!-- 背景切换对话框 -->
    <el-dialog
      v-model="showBgPatternDialog"
      title="选择背景样式"
      width="360px"
      class="bg-patterns-dialog"
    >
      <div class="patterns-grid">
        <div
          v-for="pattern in bgPatterns"
          :key="pattern.value"
          class="pattern-item"
          :class="[
            `pattern-${pattern.value}`,
            { active: currentBgPattern === pattern.value }
          ]"
          @click="selectBgPattern(pattern.value)"
        >
          <div class="pattern-name">{{ pattern.label }}</div>
        </div>
      </div>
    </el-dialog>

    <!-- AI 助手市场 -->
    <PresetMarket
      v-model:visible="showPresetMarket"
      @select="handlePresetSelect"
    />

    <!-- Theme Drawer -->
    <ThemeDrawer
      v-model="showThemeDrawer"
    />
  </div>
</template>

<script setup name="ChatView">
import {ref, computed, onMounted, nextTick, onUnmounted} from 'vue'
import {useRouter, useRoute} from 'vue-router'
import {ElMessage, ElMessageBox, ElCheckbox} from 'element-plus'
import {useConversationStore} from '@/stores/conversation'
import {useUserStore} from '@/stores/user'
import {useThemeStore} from '@/stores/theme'
import {
  ArrowDown,
  ArrowRight,
  Check,
  CircleClose,
  CopyDocument,
  HomeFilled,
  Picture,
  Plus,
  Position,
  RefreshRight,
  Setting,
  More,
  Delete,
  Close,
} from '@element-plus/icons-vue'
import {marked} from 'marked';
import hljs from 'highlight.js';
import 'highlight.js/styles/atom-one-dark.css';
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import cpp from 'highlight.js/lib/languages/cpp'
import csharp from 'highlight.js/lib/languages/csharp'
import go from 'highlight.js/lib/languages/go'
import rust from 'highlight.js/lib/languages/rust'
import sql from 'highlight.js/lib/languages/sql'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import scss from 'highlight.js/lib/languages/scss'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import markdown from 'highlight.js/lib/languages/markdown'
import bash from 'highlight.js/lib/languages/bash'
import shell from 'highlight.js/lib/languages/shell'
import dockerfile from 'highlight.js/lib/languages/dockerfile'
import PresetMarket from '@/components/chat/PresetMarket.vue'
import 'github-markdown-css/github-markdown-light.css'
import 'github-markdown-css/github-markdown-dark.css'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ThemeDrawer from '@/components/common/ThemeDrawer.vue'
import SvgIcon from "@/components/common/SvgIcon.vue";
import wsManager from '@/utils/websocket/manager'
import {createChatMessage, createStopMessage} from '@/utils/websocket/types'
import {defaultApi} from '@/api'
import aiAvatar from '@/assets/images/ai_avatar.png'

// 注册语言
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('java', java)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('csharp', csharp)
hljs.registerLanguage('go', go)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('scss', scss)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', shell)
hljs.registerLanguage('dockerfile', dockerfile)

// 配置 marked
marked.setOptions({
  highlight: function (code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(lang, code).value;
      } catch (err) {
        console.warn('Language highlight error:', err);
      }
    }
    try {
      return hljs.highlightAuto(code).value;
    } catch (err) {
      console.warn('Auto highlight error:', err);
      return code;
    }
  },
  langPrefix: 'hljs language-',
  breaks: true,
  gfm: true,
  headerIds: false,
  mangle: false
})

// 渲染消息内容
const renderMessage = (content) => {
  try {
    // 自定义代码块渲染
    const renderer = new marked.Renderer();
    renderer.code = ({text, lang}) => {
      // 确保 code 是字符串类型
      const codeStr = String(text || '');
      const validLang = hljs.getLanguage(lang) ? lang : 'plaintext';

      let highlightedCode;
      try {
        highlightedCode = hljs.highlight(codeStr, {language: validLang}).value;
      } catch (err) {
        console.warn('Language highlight error:', err);
        highlightedCode = hljs.highlight(codeStr, {language: 'plaintext'}).value;
      }

      // 生成唯一ID用于复制功能
      const blockId = `code-block-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

      return `
        <pre class="code-block" id="${blockId}">
          <div class="code-header">
            <div class="lang-info">
              <span class="code-lang">${validLang.toUpperCase()}</span>
            </div>
            <button class="copy-button" onclick="(() => {
              const codeBlock = document.getElementById('${blockId}');
              const code = codeBlock.querySelector('code').textContent;
              const button = codeBlock.querySelector('.copy-button');
              navigator.clipboard.writeText(code)
                .then(() => {
                  button.innerHTML = '<span>已复制</span>';
                  setTimeout(() => {
                    button.innerHTML = '<span>复制</span>';
                  }, 2000);
                })
                .catch(() => {
                  button.innerHTML = '<span>复制失败</span>';
                  setTimeout(() => {
                    button.innerHTML = '<span>复制</span>';
                  }, 2000);
                });
            })()">
              <span>复制</span>
            </button>
          </div>
          <code class="hljs language-${validLang}">${highlightedCode}</code>
        </pre>
      `.trim();
    };

    marked.use({renderer});
    const rendered = marked(String(content || ''));
    return `<div class="markdown-body">${rendered}</div>`;
  } catch (err) {
    console.error('Markdown rendering error:', err);
    return String(content || '');
  }
}

// 使用主题 store
const themeStore = useThemeStore()
const userStore = useUserStore()
const conversationStore = useConversationStore()

// 状态
const isDarkTheme = computed(() => themeStore.isDark)
const isSidebarCollapsed = ref(false)
const showSettings = ref(false)
const messageInput = ref('')
const messagesContainer = ref(null)

// 使用 conversationStore 管理会话
const chatList = computed(() => conversationStore.conversations)
const currentChat = computed({
  get: () => conversationStore.currentConversation,
  set: (value) => conversationStore.setCurrentConversation(value)
})


// 头像
const userAvatar = userStore.avatar
// const aiAvatar = '@/assets/images/ai-avatar.png'

// 可用模型列表
const availableModels = ref([])
const loadingModels = ref(false)
// 当前选中模型
const currentModel = ref(null)

// 加载模型列表
const loadModels = async () => {
  try {
    loadingModels.value = true
    const response = await defaultApi.apiModelListGet(1, 999, {
      orderBy: 'id',
      order: 'desc'
    })

    if (response.code === 0 && response.data) {
      availableModels.value = response.data.records.map(model => ({
        ...model,
        tag: model.tag ? model.tag.split(',')[0] : 'default',
      }))

      // 找到默认模型并自动选中
      const defaultModel = availableModels.value.find(model => model.default)
      if (defaultModel) {
        selectModel(defaultModel)
      }
    } else {
      ElMessage.error(response.msg || '获取模型列表失败')
    }
  } catch (error) {
    console.error('Load models error:', error)
    ElMessage.error(error.body?.msg || '获取模型列表失败')
  } finally {
    loadingModels.value = false
  }
}

// 背景相关
const bgPatterns = [
  {label: '渐变青绿', value: '1'},
  {label: '渐变橙粉', value: '2'},
  {label: '渐变紫蓝', value: '3'},
  {label: '渐变粉红', value: '4'},
  {label: '渐变紫粉', value: '5'},
  {label: '无背景', value: 'none'}
]
const currentBgPattern = ref('none')
const showBgPatternDialog = ref(false)

// AI 助手市场
const showPresetMarket = ref(false)

// 移除未使用的变量
// const streamingMessage = ref(null)
const messageThoughtTimes = ref(new Map())
const showThemeDrawer = ref(false)

// 添加计算属性来获取当前会话的流式消息
// const streamingMessage = computed({
//   get: () => {
//     if (!currentChat.value) return null
//     return conversationStore.getStreamingMessage(currentChat.value.id)
//   },
//   set: (value) => {
//     if (!currentChat.value) return
//     conversationStore.setStreamingMessage(currentChat.value.id, value)
//   }
// })

// 添加计算属性来获取当前会话的打字状态
const isTyping = computed(() => {
  if (!currentChat.value) return false
  return conversationStore.getTypingStatus(currentChat.value.id)
})

// 发送消息
const sendMessage = async () => {
  if (!messageInput.value.trim() || !currentChat.value) return

  // 获取用户ID (如果登录的话)
  const userId = userStore.userId || '0'
  await NewChatConnectionIfNeed(currentChat.value, userId, currentChat.value.presetId);

  // 如果是会话的第一条消息，就把该消息设置为会话的名称
  conversationStore.updateCurrentChatName(messageInput.value)
  // 添加用户消息
  currentChat.value.messages.push({
    id: Date.now(),
    role: 'user',
    content: messageInput.value
  })

  const message = messageInput.value
  messageInput.value = ''
  await scrollToBottom()

  // 设置正在输入状态
  conversationStore.setTypingStatus(currentChat.value.id, true)

  // 新增：设置当前会话的 loading 状态为 true
  messageLoadingMap.value.set(currentChat.value.id, true)

  try {
    // 准备消息选项
    const options = {
      model: availableModels.value.find(m => m.value === currentChat.value.model)?.label || currentChat.value.model.toString(),
      enableWeb: currentChat.value.webSearch,
      maxTokens: currentChat.value.maxTokens,
      temperature: currentChat.value.temperature,
      topP: currentChat.value.topP,
      topK: currentChat.value.topK,
      presencePenalty: currentChat.value.presencePenalty,
      frequencyPenalty: currentChat.value.frequencyPenalty,
      repetitionPenalty: currentChat.value.repetitionPenalty
    }

    // 确定使用哪个连接ID发送消息
    // 如果是新会话(没有真实ID)，则使用"-1"
    const connectionId = currentChat.value.id.toString()

    // 通过 WebSocket 发送消息
    wsManager.sendMessage(
      connectionId,
      createChatMessage(message, options)
    )

    // 第一次发送后标记已经不是新会话
    // if (!currentChat.value.realId) {
    //   currentChat.value.realId = false
    // }
  } catch (error) {
    console.error('Failed to send message:', error)
    ElMessage.error('发送消息失败')
    conversationStore.setTypingStatus(currentChat.value.id, false)
  }
}

// 处理 WebSocket 消息
const handleWebSocketMessage = (chatId, data) => {
  // 关闭 loading 动画
  messageLoadingMap.value.set(chatId, false)

  if (data.type === 'chat') {
    // 完整的消息响应
    conversationStore.setTypingStatus(chatId, false)

    // 如果存在流式消息，则更新它而不是创建新消息
    const currentStreamingMessage = conversationStore.getStreamingMessage(chatId)
    if (currentStreamingMessage) {
      currentStreamingMessage.content = data.data.partialContent
      currentStreamingMessage.reasoningContent = data.data.partialReasoning
      // 记录最终的思考时间
      if (data.data.reasoningContent) {
        const endTime = Date.now()
        const startTime = messageThoughtTimes.value.get(currentStreamingMessage.id)?.startTime || endTime
        messageThoughtTimes.value.set(currentStreamingMessage.id, {
          startTime,
          endTime,
          duration: Math.floor((endTime - startTime) / 1000)
        })
      }
      // 消息完成时重置流式消息
      conversationStore.setStreamingMessage(chatId, null)
      // 消息完成时从 lastMessageMap 中删除
      conversationStore.removeLastMessage(chatId)
    } else {
      const message = {
        id: Date.now(),
        role: 'assistant',
        content: data.data.content,
        reasoningContent: data.data.reasoningContent,
        showThought: true
      }
      conversationStore.addMessage(message)
    }

    // 更新最后一条消息预览
    conversationStore.updateLastMessage(chatId, data.data.content)
    scrollToBottom()
  } else if (data.type === 'stream') {
    // 流式响应更新
    let currentStreamingMessage = conversationStore.getStreamingMessage(chatId)
    if (!currentStreamingMessage) {
      // 创建新的流式消息
      currentStreamingMessage = {
        id: Date.now(),
        role: 'assistant',
        content: '',
        reasoningContent: '',
        showThought: true
      }
      // 记录思考开始时间
      if (!messageThoughtTimes.value.has(currentStreamingMessage.id)) {
        messageThoughtTimes.value.set(currentStreamingMessage.id, {
          startTime: Date.now(),
          endTime: null,
          duration: 0
        })
      }

      // 保存到当前会话
      if (currentChat.value && currentChat.value.id === chatId) {
        currentChat.value.messages.push(currentStreamingMessage)
      } else {
        // 如果不是当前会话，则直接添加到对应会话的消息列表中
        const chat = chatList.value.find(c => c.id === chatId)
        if (chat && chat.messages) {
          chat.messages.push(currentStreamingMessage)
        }
      }

      // 设置会话的流式消息
      conversationStore.setStreamingMessage(chatId, currentStreamingMessage)

      // 将正在进行中的消息保存到 lastMessageMap 中（仅限登录用户）
      const userStore = useUserStore()
      if (userStore.isLoggedIn) {
        conversationStore.setLastMessage(chatId, currentStreamingMessage)
      }
    }

    // 更新流式消息内容
    currentStreamingMessage.content = data.data.partialContent
    currentStreamingMessage.reasoningContent = data.data.partialReasoning

    // 同步更新 lastMessageMap 中的消息（仅限登录用户）
    const userStore = useUserStore()
    if (userStore.isLoggedIn) {
      const lastMessage = conversationStore.lastMessageMap[chatId]
      if (lastMessage) {
        lastMessage.content = data.data.partialContent
        lastMessage.reasoningContent = data.data.partialReasoning
      }
    }

    // 更新当前思考时间
    if (data.data.reasoningContent) {
      const currentTime = Date.now()
      const startTime = messageThoughtTimes.value.get(currentStreamingMessage.id)?.startTime || currentTime
      messageThoughtTimes.value.set(currentStreamingMessage.id, {
        startTime,
        endTime: currentTime,
        duration: Math.floor((currentTime - startTime) / 1000)
      })
    }

    // 如果是当前会话则滚动到底部
    if (currentChat.value && currentChat.value.id === chatId) {
      scrollToBottom()
    }
  } else if (data.type === 'error') {
    // 错误消息
    ElMessage.error(data.data?.message || '接收消息出错')
    conversationStore.setTypingStatus(chatId, false)
    conversationStore.setStreamingMessage(chatId, null)

    // 出错时也需要从 lastMessageMap 中删除
    conversationStore.removeLastMessage(chatId)
  }
}

// 停止生成
const stopGeneration = () => {
  if (currentChat.value) {
    wsManager.sendMessage(
      currentChat.value.id.toString(),
      createStopMessage()
    )
    conversationStore.setTypingStatus(currentChat.value.id, false)
  }
}

// 主题切换
const toggleTheme = () => {
  showThemeDrawer.value = true
}

// 方法
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const router = useRouter()
const route = useRoute()

const goToHome = () => {
  router.push('/')
}

// 若没有连接，则创建连接
async function NewChatConnectionIfNeed(newChat, userId, presetId) {

  console.log("New chat connection", newChat.id, userId, presetId);

  let connCreated = await wsManager.hasConnection(newChat.id);

  if (connCreated) {
    console.log("Connection already created", newChat.id);
    return;
  }

  // 创建 WebSocket 连接
  let b = await wsManager.createConnection(newChat.id, userId, presetId);
  if (!b) {
    return;
  }

  // 添加消息处理器
  wsManager.on(newChat.id, 'message', (data) => {
    // console.log("Received message", data);

    // 如果收到的消息包含会话ID，更新当前会话ID
    if (data.conversationId) {
      const actualChatId = data.conversationId.toString()

      // 如果这是第一条消息，且ID与当前不同，需要更新会话ID
      if (newChat.realId === false && newChat.id.toString() !== actualChatId) {
        // 更新会话对象的ID
        console.log(`Updating chat ID from ${newChat.id} to ${actualChatId}`)
        let oldId = newChat.id
        newChat.id = parseInt(actualChatId)
        newChat.realId = true

        // 更新会话ID
        conversationStore.updateConversationId(oldId, newChat.id)

        // 更新连接映射
        wsManager.updateConnectionId(oldId, actualChatId)

        // 保存到本地存储
        conversationStore.saveToLocalStorage()
      }
    }

    handleWebSocketMessage(newChat.id, data)
  })

  wsManager.on(newChat.id, 'error', (error) => {
    console.error('WebSocket error:', error)
    ElMessage.error('连接发生错误')
  })

  wsManager.on(newChat.id, 'close', () => {
    console.log('WebSocket connection closed')
  })
}

const createNewChat = async (assistantId) => {
  let preset = {};
  if (assistantId) {
    // 获取助手详情
    const res = await defaultApi.apiPresetIdGet(Number(assistantId));
    if (res.code === 0) {
      preset = res.data;
    } else {
      ElMessage.error('获取助手详情失败')
    }
  }

  console.log("Assistant", preset)

  // 找到默认模型并自动选中
  const defaultModel = availableModels.value.find(model => model.default)

  const newChat = {
    id: "tmp-" + Date.now(),
    // 标记这是一个还没有真实ID的新会话
    realId: false,
    name: preset.name ? `与 ${preset.name} 对话` : '新对话',
    model: defaultModel?.name || 'gpt-3.5-turbo',
    presetId: preset.id,
    webSearch: false,
    maxTokens: 2048,
    temperature: 1,
    topP: 0.7,
    topK: 50,
    presencePenalty: 0,
    frequencyPenalty: 0,
    repetitionPenalty: 1,
    lastMessage: preset.description || '你好！我是 AI 助手，有什么我可以帮你的吗？',
    avatar: preset.avatar || defaultModel?.avatar || aiAvatar,
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: preset.description
          ? `你好！我是 ${preset.name}，${preset.description}`
          : '你好！我是 AI 助手，有什么我可以帮你的吗？'
      }
    ],
    preset: preset,
  }

  try {

    await conversationStore.addConversation(newChat)

    if (defaultModel) {
      // 选中默认模型
      selectModel(defaultModel)
    }
  } catch (error) {
    console.error('Failed to create chat:', error)
    ElMessage.error('创建会话失败')
  }
}

const switchChat = async (chat) => {
  try {
    // 检查当前会话是否是新会话且没有消息
    if (currentChat.value &&
        currentChat.value.id.toString().startsWith('tmp-') &&
        (!currentChat.value.messages || currentChat.value.messages.filter(m => m.role === 'user').length === 0)) {
      // 删除当前空会话
      conversationStore.batchDeleteConversations([currentChat.value.id])
    }

    // 加载会话详情
    await conversationStore.loadConversationDetail(chat.id)

    console.log(availableModels.value)
    // 根据会话的 model 字段查找并选中对应的模型
    console.log("Switching chat", chat.id, availableModels.value)
    const model = availableModels.value.find(m => m.name === chat.model)
    console.log("Model", model)
    selectModel(model)

    // if (model) {
    //   // 更新会话列表中的显示
    //   const chatInList = chatList.value.find(c => c.id === chat.id)
    //   if (chatInList) {
    //     chatInList.avatar = model.avatar
    //     // chatInList.name = model.name
    //   }
    //   console.log("Switching model", model)
    //   // 更新当前选中模型
    //   currentModel.value = model
    // }
    //
    // console.log("Switching chat", currentChat.value)

    // 建立 WebSocket 连接
    // await NewChatConnectionIfNeed(chat, userStore.userId || '0', "")
    await scrollToBottom()
  } catch (error) {
    console.error('Failed to switch chat:', error)
    ElMessage.error('切换会话失败')
  }
}

const handleChatAction = async (command, chat) => {
  if (command === 'delete') {
    try {
      await conversationStore.batchDeleteConversations([chat.id])
    } catch (error) {
      console.error('Failed to delete chat:', error)
      ElMessage.error('删除会话失败')
    }
  }
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
  if (!currentChat.value) return
  conversationStore.setTypingStatus(currentChat.value.id, true)
  // 模拟重新生成
  setTimeout(() => {
    conversationStore.setTypingStatus(currentChat.value.id, false)
  }, 2000)
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const saveSettings = () => {
  showSettings.value = false
  ElMessage.success('设置已保存')
}

// 选择模型
const selectModel = (model) => {
  console.log("currentChat", currentChat.value, "model:", model)
  currentChat.value.model = model?.name;
  // currentChat.value.name = model.name;

  // 更新当前会话在列表中的头像
  const chatInList = chatList.value.find(chat => chat.id === currentChat.value.id);
  if (chatInList && !currentChat.value?.preset) {
    chatInList.avatar = model?.avatar;
    // chatInList.name = model.name;
    conversationStore.updateConversation(chatInList)

    currentChat.value.avatar = model?.avatar;
  }
  // 设置当前选中模型
  currentModel.value = model;

  // 保存到本地存储
  conversationStore.saveToLocalStorage();
}

// 切换联网搜索
const toggleWebSearch = () => {
  currentChat.value.webSearch = !currentChat.value.webSearch
}

// 切换思考过程的显示/隐藏
const toggleThought = (message) => {
  if (!message.showThought) {
    message.showThought = true;
  } else {
    message.showThought = false;
  }
}

// 判断消息是否为当前会话的流式消息
const isCurrentStreamingMessage = (message) => {
  if (!currentChat.value) return false
  const currentStreamingMessage = conversationStore.getStreamingMessage(currentChat.value.id)
  return currentStreamingMessage && currentStreamingMessage.id === message.id
}

// 监听系统主题变化
onMounted(async () => {
  hljs.highlightAll()

  // 清空进行中消息缓存
  conversationStore.clearLastMessageMap()

  // 重置所有会话的打字状态和流式消息
  conversationStore.resetAllTypingStatus()
  conversationStore.resetAllStreamingMessages()

  try {
    // 加载会话列表
    await conversationStore.loadConversations()

    // 检查是否需要创建新对话
    if (route.query.newChat === 'true') {
      const assistantId = route.query.assistantId
      await createNewChat(assistantId)
    } else if (chatList.value.length > 0) {
      await conversationStore.loadConversationDetail(chatList.value[0].id)
      console.log("Switching chat", chatList.value[0].id)

      // await NewChatConnectionIfNeed(chatList.value[0], userStore.userId, "");
    }

    const savedPattern = localStorage.getItem('chatBgPattern')
    if (savedPattern) {
      currentBgPattern.value = savedPattern
    }

    // 初始化主题
    themeStore.initTheme()
    await loadModels()
    // 设置当前选中模型
    currentModel.value = availableModels.value.find(m => m.name === currentChat.value.model) || availableModels.value[0]
  } catch (error) {
    console.error('Failed to initialize chat:', error)
    ElMessage.error('初始化聊天失败')
  }
})

// 添加输入框高度相关的状态和方法
const textareaRows = ref(3)
let startY = 0
let startHeight = 0
const minRows = 3
const maxRows = 15

// 新增：每个会话的"AI 正在思考中"动画状态 Map
// Map<chatId, boolean>
const messageLoadingMap = ref(new Map())

const startResize = (e) => {
  startY = e.clientY
  startHeight = textareaRows.value

  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
}

const handleResize = (e) => {
  const delta = startY - e.clientY
  const rowHeight = 24 // 每行大约的高度
  const rowDelta = Math.round(delta / rowHeight)

  let newRows = startHeight + rowDelta
  newRows = Math.max(minRows, Math.min(maxRows, newRows))
  textareaRows.value = newRows
}

const stopResize = () => {
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

const showBgPatternSelector = () => {
  showBgPatternDialog.value = true
}

const selectBgPattern = (pattern) => {
  currentBgPattern.value = pattern
  showBgPatternDialog.value = false
  localStorage.setItem('chatBgPattern', pattern)
}

const handlePresetSelect = (preset) => {
  createNewChat(preset.id)
}

// 在组件销毁时关闭所有连接
onUnmounted(() => {
  // 重置所有会话的打字状态和流式消息
  conversationStore.resetAllTypingStatus()
  conversationStore.resetAllStreamingMessages()

  // 关闭当前聊天的连接
  if (currentChat.value) {
    wsManager.closeConnection(currentChat.value.id.toString())
  }

  // 关闭所有聊天的连接
  chatList.value.forEach(chat => {
    if (chat.id !== currentChat.value?.id) {
      wsManager.closeConnection(chat.id.toString())
    }
  })
})

// 批量操作相关
const showCheckboxes = ref(false)
const selectedChats = ref([])

// 切换批量操作模式
const toggleBatchMode = () => {
  showCheckboxes.value = !showCheckboxes.value
  if (!showCheckboxes.value) {
    selectedChats.value = []
  }
}

// 清除选择
const clearSelection = () => {
  selectedChats.value = []
}

// 处理会话点击
const handleChatClick = (chat) => {
  if (showCheckboxes.value) {
    // 批量操作模式下，切换选中状态
    const index = selectedChats.value.indexOf(chat.id)
    if (index === -1) {
      selectedChats.value.push(chat.id)
    } else {
      selectedChats.value.splice(index, 1)
    }
  } else {
    // 普通模式下，切换会话
    switchChat(chat)
  }
}

// 批量删除
const batchDelete = async () => {
  if (selectedChats.value.length === 0) {
    ElMessage.warning('请选择要删除的会话')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedChats.value.length} 个会话吗？`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 调用批量删除方法
    await conversationStore.batchDeleteConversations(selectedChats.value)

    // 清空选择
    selectedChats.value = []
    // 关闭批量操作模式
    showCheckboxes.value = false
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to batch delete:', error)
      ElMessage.error('删除失败')
    }
  }
}
</script>

<style scoped lang="scss">
// 移除本地主题变量定义，使用全局变量
.chat-container {
  display: flex;
  height: 100vh;
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    opacity: 0.1;
    z-index: 0;
    transition: opacity 0.3s ease;
    background: var(--chat-bg-pattern);
    pointer-events: none;
  }

  &.bg-pattern-1 {
    --chat-bg-pattern: linear-gradient(120deg, #84fab0 0%, #8fd3f4 100%);
  }

  &.bg-pattern-2 {
    --chat-bg-pattern: linear-gradient(to right, #fa709a 0%, #fee140 100%);
  }

  &.bg-pattern-3 {
    --chat-bg-pattern: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }

  &.bg-pattern-4 {
    --chat-bg-pattern: linear-gradient(45deg, #ff9a9e 0%, #fad0c4 99%, #fad0c4 100%);
  }

  &.bg-pattern-5 {
    --chat-bg-pattern: linear-gradient(to top, #a18cd1 0%, #fbc2eb 100%);
  }

  &.bg-pattern-none {
    --chat-bg-pattern: none;
  }
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
  background: var(--el-bg-color);
  position: relative;
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-right: 1px solid var(--el-border-color-light);

  &::after {
    content: '';
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: 1px;
    background: linear-gradient(180deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.1) 15%,
      rgba(var(--divider-rgb), 0.2) 30%,
      rgba(var(--divider-rgb), 0.3) 50%,
      rgba(var(--divider-rgb), 0.2) 70%,
      rgba(var(--divider-rgb), 0.1) 85%,
      rgba(var(--divider-rgb), 0) 100%
    );
    box-shadow: 1px 0 2px rgba(0, 0, 0, 0.05);
  }

  &.sidebar-collapsed {
    width: 0;
    overflow: hidden;
  }
}

.action-buttons {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(180deg,
    var(--bg-primary) 0%,
    var(--bg-secondary) 100%
  );
  display: flex;
  gap: 12px;
  align-items: center;

  .home-button {
    transition: all 0.3s ease;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    height: 40px;
    width: 40px;
    color: var(--text-primary);

    &:hover {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
      transform: translateY(-2px);
    }

    .el-icon {
      font-size: 18px;
    }
  }

  .new-chat-button {
    transition: all 0.3s ease;
    background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-primary-light-3) 100%);
    border: none;
    height: 30px;
    width: 30px;
    box-shadow: 0 2px 12px rgba(var(--el-color-primary-rgb), 0.2);

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 16px rgba(var(--el-color-primary-rgb), 0.3);
    }

    .icon-bounce {
      animation: iconBounce 1s infinite;
      font-size: 18px;
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
    background: var(--el-fill-color-light);
    border-color: var(--border-color);
    transform: translateX(4px);
  }

  &.active {
    //background: var(--el-color-primary-light-9);
    border-color: var(--border-color);
  }

  .chat-icon-wrapper {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    overflow: hidden;

    .el-avatar {
      width: 100%;
      height: 100%;
      border-radius: 10px !important;
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
    color: var(--el-text-color-secondary);
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

  .footer-actions {
    display: flex;
    gap: 12px;
  }
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  position: relative;

  // 添加固定的展开按钮
  .sidebar-expand {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    padding: 12px 8px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-left: none;
    border-radius: 0 8px 8px 0;
    cursor: pointer;
    z-index: 10;
    transition: all 0.3s ease;
    opacity: 0;
    pointer-events: none;

    &.show {
      opacity: 1;
      pointer-events: auto;
    }

    &:hover {
      background: var(--hover-bg);
      transform: translateY(-50%) translateX(2px);
    }

    .expand-icon {
      transition: transform 0.3s ease;
      transform: rotate(180deg);
    }
  }
}

.chat-header {
  padding: 16px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-primary);
  z-index: 1;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.1) 15%,
      rgba(var(--divider-rgb), 0.2) 30%,
      rgba(var(--divider-rgb), 0.3) 50%,
      rgba(var(--divider-rgb), 0.2) 70%,
      rgba(var(--divider-rgb), 0.1) 85%,
      rgba(var(--divider-rgb), 0) 100%
    );
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }

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
  position: relative;
  margin: 0 auto;
  width: 100%;
  max-width: 1200px;

  @media screen and (max-width: 768px) {
    padding: 16px;
  }

  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.1) 15%,
      rgba(var(--divider-rgb), 0.2) 30%,
      rgba(var(--divider-rgb), 0.3) 50%,
      rgba(var(--divider-rgb), 0.2) 70%,
      rgba(var(--divider-rgb), 0.1) 85%,
      rgba(var(--divider-rgb), 0) 100%
    );
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }
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
      color: var(--text-primary);
    }

    .message-actions {
      justify-content: flex-start;
    }
  }

  &.assistant {
    .message-content {
      background: var(--message-bg-assistant);
      border-radius: 2px 12px 12px 12px;
    }

    .message-avatar {
      .el-avatar {
        background: transparent;

        img {
          object-fit: contain;
          padding: 4px;
        }
      }
    }
  }

  .message-avatar {
    .el-avatar {
      box-shadow: none;
      transition: transform 0.3s ease;

      &:hover {
        transform: translateY(-2px);
      }
    }
  }
}

.message-content {
  max-width: 85%;
  padding: 16px;
  box-shadow: 0 1px 2px var(--shadow-color);
  transition: transform 0.3s ease;
  font-size: 16px;
  line-height: 1.6;
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;

  @media screen and (min-width: 1200px) {
    max-width: 900px;
  }

  @media screen and (min-width: 1600px) {
    max-width: 1000px;
  }

  .thought-process {
    margin-bottom: 16px;
    border-radius: 6px;
    background: var(--el-bg-color);
    overflow: hidden;

    .thought-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      font-size: 13px;
      color: var(--el-text-color-secondary);
      cursor: pointer;
      transition: all 0.3s ease;
      user-select: none;

      &:hover {
        background: var(--el-fill-color);
      }

      .el-icon {
        font-size: 16px;
        transition: transform 0.3s ease;

        &.is-fold {
          transform: rotate(90deg);
        }
      }
    }

    .thought-content {
      padding: 12px 16px;
      font-size: 14px;
      line-height: 1.6;
      color: var(--el-text-color-regular);
      border-top: 1px solid var(--el-border-color-light);
      background: var(--el-bg-color);
      white-space: pre-wrap;
    }
  }

  .message-text {
    :deep(.markdown-body) {
      background: transparent !important;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      font-size: 16px;
      line-height: 1.6;
      text-rendering: optimizeLegibility;
      color: var(--el-text-color-primary) !important;

      pre.code-block {
        background: #282c34;
        margin: 1em 0;
        padding: 0;
        border-radius: 8px;
        overflow: hidden;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
        display: flex;
        flex-direction: column;

        .code-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 8px 12px;
          background: #21252b;
          border-bottom: 1px solid rgba(255, 255, 255, 0.05);
          height: 40px;

          .lang-info {
            display: flex;
            align-items: center;
            gap: 8px;

            .code-lang {
              color: #abb2bf;
              font-size: 12px;
              font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
              font-weight: 500;
              background: rgba(255, 255, 255, 0.1);
              padding: 2px 8px;
              border-radius: 4px;
              letter-spacing: 0.5px;
            }
          }

          .copy-button {
            display: flex;
            align-items: center;
            gap: 4px;
            background: transparent;
            border: 1px solid rgba(255, 255, 255, 0);
            color: #abb2bf;
            padding: 4px 12px;
            height: 28px;
            font-size: 12px;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.2s ease;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;

            .copy-icon {
              font-size: 14px;
            }

            &:hover {
              background: rgba(255, 255, 255, 0.05);
              border-color: rgba(255, 255, 255, 0.2);
              color: #fff;
            }

            &:active {
              transform: translateY(1px);
            }
          }
        }

        code {
          font-family: 'JetBrains Mono', 'Fira Code', Consolas, Monaco, monospace;
          font-size: 14px;
          background: transparent;
          text-shadow: 0 1px rgba(0, 0, 0, 0.3);
          padding: 16px;
          display: block;
          overflow-x: auto;
          line-height: 1.6;
        }
      }
    }
  }

  .message-actions {
    margin-top: 4px;
    display: flex;
    justify-content: flex-end;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s ease;

    .el-button {
      padding: 2px 6px;
      font-size: 12px;
      height: 24px;
      --el-button-hover-bg-color: var(--el-color-primary-light-8);
      --el-button-hover-text-color: var(--el-color-primary);

      .el-icon {
        margin-right: 2px;
        font-size: 12px;
      }
    }
  }

  &:hover .message-actions {
    opacity: 1;
  }
}

.chat-input {
  padding: 0px 24px 16px;
  background: var(--el-bg-color);
  position: relative;
  border-top: 1px solid var(--el-border-color-light);

  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.5) 15%,
      rgba(var(--divider-rgb), 0.7) 30%,
      rgba(var(--divider-rgb), 0.9) 50%,
      rgba(var(--divider-rgb), 0.7) 70%,
      rgba(var(--divider-rgb), 0.5) 85%,
      rgba(var(--divider-rgb), 0) 100%
    );
    z-index: 1;
  }

  .resize-handle {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    //height: 1px;
    cursor: row-resize;
    z-index: 2;
    background: #f5f7fa;
    transition: background 0.2s ease;

    &:hover {
      background: linear-gradient(180deg,
        rgba(var(--divider-rgb), 0.3) 0%,
        rgba(var(--divider-rgb), 0) 100%
      );
    }

    &::before {
      content: '';
      position: absolute;
      left: 50%;
      top: 50%;
      transform: translate(-50%, -50%);
      width: 48px;
      height: 4px;
      border-radius: 2px;
      background: linear-gradient(90deg,
        rgba(var(--divider-rgb), 0) 0%,
        rgba(var(--divider-rgb), 0.5) 20%,
        rgba(var(--divider-rgb), 0.8) 50%,
        rgba(var(--divider-rgb), 0.5) 80%,
        rgba(var(--divider-rgb), 0) 100%
      );
      opacity: 0;
      transition: opacity 0.2s ease;
    }

    &:hover::before {
      opacity: 1;
    }
  }
}

.quick-settings {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  margin: 0;
  //background: var(--el-fill-color-light);
  border-radius: 6px;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: -8px;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.5) 15%,
      rgba(var(--divider-rgb), 0.7) 30%,
      rgba(var(--divider-rgb), 0.9) 50%,
      rgba(var(--divider-rgb), 0.7) 70%,
      rgba(var(--divider-rgb), 0.5) 85%,
      rgba(var(--divider-rgb), 0) 100%
    );
  }
}

.model-selector {
  .current-model {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s ease;
    border: 1px solid var(--border-color);
    background: var(--bg-primary);

    &:hover {
      border-color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    .model-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      border-radius: 8px;
      //background: linear-gradient(135deg, #4158D0, #C850C0);
      color: white;

      .el-avatar {
        border-radius: 8px !important;
      }
    }

    .model-name {
      font-size: 14px;
      font-weight: 500;
    }

    .arrow-icon {
      font-size: 12px;
      color: var(--text-secondary);
      transition: transform 0.3s ease;
    }

    &:hover .arrow-icon {
      transform: rotate(180deg);
    }
  }
}

.model-list {
  max-height: 400px;
  overflow-y: auto;
  padding: 8px;

  .el-empty {
    padding: 24px;
  }

  .el-skeleton {
    padding: 16px;
  }
}

.model-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.3s ease;

  &:hover {
    background: var(--el-color-primary-light-9);
  }

  &.active {
    background: var(--el-color-primary-light-8);
  }

  .model-item-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 8px;
    overflow: hidden;
    background: var(--el-fill-color-light);
    transition: all 0.3s ease;

    .el-avatar {
      width: 100%;
      height: 100%;
      font-size: 14px;
      background: linear-gradient(135deg, var(--el-color-primary), var(--el-color-primary-light-3));
      color: white;
      font-weight: 500;
      overflow: initial;
      border-radius: 8px;
    }

    &.通用 .el-avatar {
      background: linear-gradient(135deg, #409EFF, #2B5EFF);
    }

    &.对话 .el-avatar {
      background: linear-gradient(135deg, #67C23A, #409EFF);
    }

    &.编程 .el-avatar {
      background: linear-gradient(135deg, #E6A23C, #F56C6C);
    }

    &.创意 .el-avatar {
      background: linear-gradient(135deg, #9C27B0, #E6A23C);
    }

    &.分析 .el-avatar {
      background: linear-gradient(135deg, #F56C6C, #9C27B0);
    }

    &.default .el-avatar {
      background: linear-gradient(135deg, #909399, #606266);
    }
  }

  .model-item-info {
    flex: 1;
    min-width: 0;

    .model-item-name {
      font-size: 14px;
      font-weight: 500;
      margin-bottom: 2px;
    }

    .model-item-desc {
      font-size: 12px;
      color: var(--text-secondary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}

.feature-toggles {
  display: flex;
  gap: 8px;

  .feature-toggle {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.3s ease;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    background: var(--bg-primary);

    &:hover {
      border-color: var(--el-color-primary);
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    //&.active {
    //  background: var(--el-color-primary);
    //  border-color: var(--el-color-primary);
    //  color: white;
    //}
  }
}

:deep(.model-popover) {
  padding: 8px !important;

  .el-popper__arrow::before {
    background: var(--bg-primary) !important;
    border-color: var(--border-color) !important;
  }
}

.input-wrapper {
  position: relative;

  .custom-input {
    transition: all 0.3s ease;

    :deep(.el-textarea__inner) {
      resize: none !important;
      transition: all 0.3s ease;
      padding-right: 90px;
      line-height: 1.6;
      font-size: 14px;

      &:focus {
        box-shadow: 0 0 0 2px var(--el-color-primary-light-8);
      }
    }
  }

  .input-actions {
    position: absolute;
    right: 8px;
    bottom: 8px;

    .el-button {
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-2px);
      }
    }

    .send-button {
      background: var(--el-color-primary);
      border: none;

      &:hover {
        background: var(--el-color-primary-light-3);
      }
    }
  }
}

.quick-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: 12px;
  padding-left: 12px;
  border-left: 1px solid var(--border-color);

  .action-btn {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.3s ease;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    background: var(--bg-primary);

    &:hover {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
      transform: translateY(-2px);
    }

    .el-icon {
      font-size: 16px;
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

.bg-patterns-dialog {
  .patterns-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
    padding: 16px;

    .pattern-item {
      height: 80px;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      border: 2px solid transparent;
      display: flex;
      align-items: flex-end;
      padding: 12px;

      &:hover {
        transform: translateY(-4px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      }

      &.active {
        border-color: var(--el-color-primary);
      }

      .pattern-name {
        color: #fff;
        font-size: 14px;
        text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
        z-index: 1;
      }

      &.pattern-1 {
        background: linear-gradient(120deg, #84fab0 0%, #8fd3f4 100%);
      }

      &.pattern-2 {
        background: linear-gradient(to right, #fa709a 0%, #fee140 100%);
      }

      &.pattern-3 {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }

      &.pattern-4 {
        background: linear-gradient(45deg, #ff9a9e 0%, #fad0c4 99%, #fad0c4 100%);
      }

      &.pattern-5 {
        background: linear-gradient(to top, #a18cd1 0%, #fbc2eb 100%);
      }

      &.pattern-none {
        background: var(--bg-secondary);
        border: 1px dashed var(--border-color);

        .pattern-name {
          color: var(--text-primary);
          text-shadow: none;
        }
      }
    }
  }
}

.chat-settings {
  display: flex;
  align-items: center;
  gap: 12px;

  .user-avatar {
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    padding: 2px;
    border-radius: 50%;
    transition: all 0.3s ease;

    &:hover {
      background: var(--hover-bg);
      transform: translateY(-1px);
    }

    .el-icon--right {
      font-size: 12px;
      color: var(--text-secondary);
      transition: transform 0.3s ease;
    }

    &:hover .el-icon--right {
      transform: rotate(180deg);
    }
  }
}

// 点点加载动画样式，仿微信气泡打字动画，左侧对齐
.dot-loading-indicator {
  display: flex;
  align-items: flex-start;
  margin: 0 0 12px 56px; // 与 assistant 头像左对齐
  min-height: 24px;
}

.dot-typing {
  display: flex;
  align-items: center;
  height: 24px;
}

.dot {
  width: 8px;
  height: 8px;
  margin-right: 4px;
  border-radius: 50%;
  background: var(--el-color-primary);
  opacity: 0.7;
  animation: dot-bounce 1.2s infinite both;
}

.dot:nth-child(2) {
  animation-delay: 0.2s;
}

.dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes dot-bounce {
  0%, 80%, 100% { transform: scale(0.7); opacity: 0.5; }
  40% { transform: scale(1.2); opacity: 1; }
}

.batch-actions {
  padding: 8px;
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
  background: var(--el-bg-color);
  position: sticky;
  top: 0;
  z-index: 1;
}

.chat-item {
  &.selected {
    background: var(--el-color-primary-light-9);
    border-color: var(--el-color-primary);
  }
}

.batch-toggle {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  background: var(--bg-primary);

  &:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }

  &.active {
    background: var(--el-color-primary);
    border-color: var(--el-color-primary);
    color: white;
  }
}

// 批量操作栏美化
.batch-actions-bar {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  background: var(--el-bg-color);
  box-shadow: 0 -2px 16px rgba(0,0,0,0.06);
  border-top: 1.5px solid var(--el-border-color-light);
  border-radius: 16px 16px 0 0;
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 10px 8px 10px 8px;
  z-index: 10;
  gap: 10px;
  animation: batch-bar-in 0.25s;
}

.batch-btn {
  font-size: 14px;
  min-width: 90px;
  height: 32px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: 16px;
  &.el-button--danger {
    background: linear-gradient(90deg, #f56c6c 60%, #f78989 100%);
    color: #fff;
    border: none;
    &:hover {
      background: linear-gradient(90deg, #f56c6c 80%, #f78989 100%);
      box-shadow: 0 4px 16px rgba(245,108,108,0.15);
      transform: translateY(-2px) scale(1.04);
    }
  }
  &.cancel-btn {
    background: #f4f4f5;
    color: #909399;
    border: none;
    &:hover {
      background: #e4e7ed;
      color: #606266;
      transform: translateY(-2px) scale(1.04);
    }
  }
}

@keyframes batch-bar-in {
  from { opacity: 0; transform: translateY(40px); }
  to { opacity: 1; transform: translateY(0); }
}

.batch-bar-fade-enter-active, .batch-bar-fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.batch-bar-fade-enter-from, .batch-bar-fade-leave-to {
  opacity: 0;
  transform: translateY(40px);
}

// 移除顶部批量操作栏样式
.batch-actions {
  display: none;
}
</style>
