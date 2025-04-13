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
            <el-icon><HomeFilled /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="新建会话" placement="right">
          <el-button type="primary" circle class="new-chat-button" @click="createNewChat">
            <el-icon><Plus class="icon-bounce" /></el-icon>
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
        <div class="footer-actions">
          <el-tooltip content="切换主题" placement="top">
            <div class="theme-toggle" @click="toggleTheme">
              <el-icon><component :is="isDarkTheme ? 'Sunny' : 'Moon'" /></el-icon>
            </div>
          </el-tooltip>
          <el-tooltip content="切换背景" placement="top">
            <div class="bg-toggle" @click="showBgPatternSelector">
              <el-icon><Picture /></el-icon>
            </div>
          </el-tooltip>
        </div>
        <div class="sidebar-toggle" @click="toggleSidebar">
          <el-icon><Fold /></el-icon>
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
        <el-icon><Fold class="expand-icon" /></el-icon>
      </div>

      <template v-if="currentChat">
        <!-- 聊天头部 -->
        <div class="chat-header">
          <div class="chat-title">
            <span>{{ currentChat.title }}</span>
<!--            <el-tag size="small" effect="plain" class="ml-2 model-tag">{{ currentChat.model }}</el-tag>-->
          </div>
          <div class="chat-settings">
            <el-tooltip content="模型设置" placement="bottom">
              <el-button circle @click="showSettings = true">
                <el-icon><Setting /></el-icon>
              </el-button>
            </el-tooltip>
            <el-dropdown trigger="click">
              <div class="user-avatar">
                <el-avatar :size="32" :src="userAvatar">U</el-avatar>
                <el-icon class="el-icon--right"><CaretBottom /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item>
                    <el-icon><User /></el-icon>
                    <span>个人中心</span>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-icon><Setting /></el-icon>
                    <span>设置</span>
                  </el-dropdown-item>
                  <el-dropdown-item divided>
                    <el-icon><SwitchButton /></el-icon>
                    <span>退出登录</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
                  :src="message.role === 'user' ? userAvatar : (currentChat.assistant?.avatar || aiAvatar)"
                >
                  {{ message.role === 'user' ? 'U' : (currentChat.assistant?.name?.charAt(0) || 'AI') }}
                </el-avatar>
              </div>
              <div class="message-content">
                <div class="message-text" v-html="renderMessage(message.content)"></div>
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
                      <el-icon><ChatRound /></el-icon>
                    </div>
                    <span class="model-name">{{ currentChat.model }}</span>
                    <el-icon class="arrow-icon"><ArrowDown /></el-icon>
                  </div>
                </template>
                <div class="model-list">
                  <div
                    v-for="model in availableModels"
                    :key="model.value"
                    class="model-item"
                    :class="{ active: currentChat.model === model.value }"
                    @click="selectModel(model)"
                  >
                    <div class="model-item-icon" :class="model.type">
                      <el-icon><component :is="model.icon" /></el-icon>
                    </div>
                    <div class="model-item-info">
                      <div class="model-item-name">{{ model.label }}</div>
                      <div class="model-item-desc">{{ model.description }}</div>
                    </div>
                    <el-icon v-if="currentChat.model === model.value"><Check /></el-icon>
                  </div>
                </div>
              </el-popover>
            </div>
            <div class="quick-actions">
              <el-tooltip content="AI 助手市场" placement="top">
                <div class="action-btn" @click="showPresetMarket = true">
                  <el-icon><Shop /></el-icon>
                </div>
              </el-tooltip>
              <div class="feature-toggles">
                <el-tooltip content="联网搜索" placement="top">
                  <div
                    class="feature-toggle"
                    :class="{ active: currentChat.webSearch }"
                    @click="toggleWebSearch"
                  >
                    <el-icon><Connection /></el-icon>
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
  </div>
</template>

<script setup name="ChatPage">
import { ref, nextTick, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Plus, ChatRound, More, Fold, Setting,
  CopyDocument, RefreshRight, Upload, Position,
  Connection, ArrowDown, Check, Picture, HomeFilled,
  Shop, User, CaretBottom, SwitchButton, Document
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { marked } from 'marked';
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
  highlight: function(code, lang) {
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
    renderer.code = ({raw, lang, text}) => {
      console.log(text, lang)
      var code = text;
      // 确保 code 是字符串类型
      const codeStr = String(code || '');
      const validLang = hljs.getLanguage(lang) ? lang : 'plaintext';
      const escapedCode = codeStr
        .replace(/`/g, '\\`')
        .replace(/\$/g, '\\$')
        .replace(/\\/g, '\\\\');

      let highlightedCode;
      try {
        highlightedCode = hljs.highlight(codeStr, { language: validLang }).value;
      } catch (err) {
        console.warn('Language highlight error:', err);
        highlightedCode = hljs.highlight(codeStr, { language: 'plaintext' }).value;
      }

      return `
        <pre class="code-block">
          <div class="code-header">
            <div class="lang-info">
              <span class="code-lang">${validLang.toUpperCase()}</span>
            </div>
            <button class="copy-button">
                <span> 复制 </span>
            </button>
          </div>
          <code class="hljs language-${validLang}">${highlightedCode}</code>
        </pre>
      `.trim();
    };

    marked.use({ renderer });
    const rendered = marked(String(content || ''));
    return `<div class="markdown-body">${rendered}</div>`;
  } catch (err) {
    console.error('Markdown rendering error:', err);
    return String(content || '');
  }
}

// 状态
const isDarkTheme = ref(window.matchMedia('(prefers-color-scheme: dark)').matches)
const isSidebarCollapsed = ref(false)
const showSettings = ref(false)
const messageInput = ref('')
const isTyping = ref(false)
const messagesContainer = ref(null)
const currentChat = ref({
  id: 1,
  title: '冒泡排序实现',
  model: 'gpt-3.5-turbo',
  webSearch: false,
  maxTokens: 2048,
  temperature: 1,
  topP: 0.7,
  topK: 50,
  presencePenalty: 0,
  frequencyPenalty: 0,
  repetitionPenalty: 1,
  messages: [
    {
      id: 1,
      role: 'assistant',
      content: `# Java实现冒泡排序

冒泡排序是一种简单的排序算法，它重复地遍历要排序的列表，比较相邻的元素并交换它们的位置，直到列表排序完成。

以下是Java实现冒泡排序的代码：

\`\`\`java
public class BubbleSort {

    public static void bubbleSort(int[] arr) {
        int n = arr.length;
        // 外层循环控制排序轮数
        for (int i = 0; i < n - 1; i++) {
            // 内层循环控制每轮比较次数
            for (int j = 0; j < n - i - 1; j++) {
                // 如果前一个元素比后一个元素大，则交换它们
                if (arr[j] > arr[j + 1]) {
                    // 交换arr[j]和arr[j+1]
                    int temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                }
            }
        }
    }

    // 优化版的冒泡排序（如果某一轮没有发生交换，说明已经有序）
    public static void optimizedBubbleSort(int[] arr) {
        int n = arr.length;
        boolean swapped;
        for (int i = 0; i < n - 1; i++) {
            swapped = false;
            for (int j = 0; j < n - i - 1; j++) {
                if (arr[j] > arr[j + 1]) {
                    // 交换arr[j]和arr[j+1]
                    int temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                    swapped = true;
                }
            }
            // 如果没有发生交换，提前结束排序
            if (!swapped) {
                break;
            }
        }
    }

    public static void main(String[] args) {
        int[] arr = {64, 34, 25, 12, 22, 11, 90};

        System.out.println("排序前的数组:");
        printArray(arr);

        bubbleSort(arr);
        // 或者使用优化版: optimizedBubbleSort(arr);

        System.out.println("排序后的数组:");
        printArray(arr);
    }

    // 辅助方法：打印数组
    public static void printArray(int[] arr) {
        for (int value : arr) {
            System.out.print(value + " ");
        }
        System.out.println();
    }
}
\`\`\`

## 代码说明

1. **基本冒泡排序**：
   - 外层循环控制排序轮数（n-1轮）
   - 内层循环比较相邻元素，如果顺序不对就交换
   - 每轮结束后，最大的元素会"冒泡"到数组末尾

2. **优化版冒泡排序**：
   - 添加了\`swapped\`标志位
   - 如果某一轮没有发生交换，说明数组已经有序，可以提前结束排序
   - 对于基本有序的数组，能显著提高效率

3. **时间复杂度**：
   - 最坏情况：O(n²)（完全逆序）
   - 最好情况：O(n)（已经有序，使用优化版）
   - 平均情况：O(n²)

4. **空间复杂度**：O(1)，是原地排序算法

5. **稳定性**：冒泡排序是稳定的排序算法，因为相等的元素不会交换位置

你可以根据需要选择基本版或优化版的实现。对于小型数组或基本有序的数组，冒泡排序是一个不错的选择。`
    }
  ]
})
const chatList = ref([
  {
    id: 1,
    title: '冒泡排序实现',
    model: 'gpt-3.5-turbo',
    webSearch: false,
    temperature: 1,
    lastMessage: 'Java实现冒泡排序的详细说明和代码示例',
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: currentChat.value.messages[0].content
      }
    ]
  }
])

// 头像
const userAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'
const aiAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

// 可用模型列表
const availableModels = [
  {
    label: 'Deepseek-R1',
    value: 'deepseek-r1',
    type: 'deepseek',
    icon: 'ChatRound',
    description: '强大的代码理解和生成能力'
  },
  {
    label: 'GPT-4 Turbo',
    value: 'gpt-4-turbo',
    type: 'gpt',
    icon: 'ChatRound',
    description: '最新的 GPT-4 模型，支持更长上下文'
  },
  {
    label: 'GPT-4',
    value: 'gpt-4',
    type: 'gpt',
    icon: 'ChatRound',
    description: '强大的推理和分析能力'
  },
  {
    label: 'GPT-3.5 Turbo',
    value: 'gpt-3.5-turbo',
    type: 'gpt',
    icon: 'ChatRound',
    description: '快速响应，性价比高'
  },
  {
    label: 'Claude 2',
    value: 'claude-2',
    type: 'claude',
    icon: 'ChatRound',
    description: '优秀的写作和分析能力'
  }
]

// 背景相关
const bgPatterns = [
  { label: '渐变青绿', value: '1' },
  { label: '渐变橙粉', value: '2' },
  { label: '渐变紫蓝', value: '3' },
  { label: '渐变粉红', value: '4' },
  { label: '渐变紫粉', value: '5' },
  { label: '无背景', value: 'none' }
]
const currentBgPattern = ref('none')
const showBgPatternDialog = ref(false)

// AI 助手市场
const showPresetMarket = ref(false)

// 主题切换
const toggleTheme = () => {
  isDarkTheme.value = !isDarkTheme.value
  document.documentElement.classList.toggle('dark', isDarkTheme.value)
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

const createNewChat = () => {
  const newChat = {
    id: Date.now(),
    title: route.query.assistantName ? `与 ${route.query.assistantName} 对话` : '新对话',
    model: 'gpt-3.5-turbo',
    webSearch: false,
    temperature: 1,
    lastMessage: route.query.assistantDescription || '你好！我是 AI 助手，有什么我可以帮你的吗？',
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: route.query.assistantDescription
          ? `你好！我是 ${route.query.assistantName}，${route.query.assistantDescription}`
          : '你好！我是 AI 助手，有什么我可以帮你的吗？'
      }
    ],
    assistant: route.query.assistantId ? {
      id: route.query.assistantId,
      name: route.query.assistantName,
      avatar: route.query.assistantAvatar,
      description: route.query.assistantDescription,
      type: route.query.assistantType
    } : null
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

// 选择模型
const selectModel = (model) => {
  currentChat.value.model = model.value
}

// 切换联网搜索
const toggleWebSearch = () => {
  currentChat.value.webSearch = !currentChat.value.webSearch
}

// 监听系统主题变化
onMounted(() => {
  hljs.highlightAll()

  // 检查是否需要创建新对话
  if (route.query.newChat === 'true') {
    createNewChat()
  } else {
    currentChat.value = chatList.value[0]
  }

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', e => {
    isDarkTheme.value = e.matches
    document.documentElement.classList.toggle('dark', e.matches)
  })

  const savedPattern = localStorage.getItem('chatBgPattern')
  if (savedPattern) {
    currentBgPattern.value = savedPattern
  }
})

// 添加输入框高度相关的状态和方法
const textareaRows = ref(3)
let startY = 0
let startHeight = 0
const minRows = 3
const maxRows = 15

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
  const newChat = {
    id: Date.now(),
    title: `与 ${preset.name} 对话`,
    model: currentChat.value?.model || 'gpt-3.5-turbo',
    webSearch: false,
    temperature: 1,
    lastMessage: preset.description,
    messages: [
      {
        id: 1,
        role: 'assistant',
        content: preset.context || `你好！我是 ${preset.name}，${preset.description}`
      }
    ]
  }
  chatList.value.unshift(newChat)
  currentChat.value = newChat
}
</script>

<style scoped lang="scss">
// 主题变量
:root {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f7fa;
  --text-primary: #333333;
  --text-secondary: #666666;
  --border-color: rgba(0, 0, 0, 0.1);
  --hover-bg: #f5f7fa;
  --active-bg: #ecf5ff;
  --shadow-color: rgba(0, 0, 0, 0.05);
  --message-bg-user: #ecf5ff;
  --message-bg-assistant: #f5f7fa;
  --scrollbar-thumb: #c0c4cc;
  --scrollbar-track: #f5f7fa;
  --divider-rgb: 0, 0, 0;
  --el-color-primary-rgb: 64, 158, 255;
}

.dark-theme {
  --bg-primary: #1a1a1a;
  --bg-secondary: #2d2d2d;
  --text-primary: #e6edf3;
  --text-secondary: #8b949e;
  --border-color: rgba(255, 255, 255, 0.1);
  --hover-bg: #2d2d2d;
  --active-bg: #363636;
  --shadow-color: rgba(0, 0, 0, 0.2);
  --message-bg-user: #363636;
  --message-bg-assistant: #2d2d2d;
  --scrollbar-thumb: #4a4a4a;
  --scrollbar-track: #2d2d2d;
  --divider-rgb: 255, 255, 255;
  --el-color-primary-rgb: 64, 158, 255;

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

  :deep(pre) {
    background: #1e1e1e !important;

    &::before {
      background: #2d2d2d !important;
      color: #a0a0a0 !important;
    }
  }

  :deep(code:not(pre code)) {
    background-color: rgba(var(--el-color-primary-rgb), 0.15);
  }

}

.chat-container {
  display: flex;
  height: 100vh;
  background: var(--bg-secondary);
  color: var(--text-primary);
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
  background: var(--bg-primary);
  position: relative;
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

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
        background: linear-gradient(135deg, #4158D0, #C850C0);
      }
    }
  }

  .message-avatar {
    .el-avatar {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
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

  &:hover {
    transform: translateY(-1px);
  }

  .message-text {
    :deep(.markdown-body) {
      background: transparent;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      font-size: 16px;
      line-height: 1.6;
      text-rendering: optimizeLegibility;
      color: var(--text-primary);

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

      code:not(pre code) {
        font-family: 'JetBrains Mono', 'Fira Code', monospace;
        font-size: 13px;
        padding: 0.2em 0.4em;
        border-radius: 4px;
        background-color: rgba(var(--el-color-primary-rgb), 0.1);
        color: var(--el-color-primary);
        font-weight: 500;
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
  background: var(--bg-primary);
  position: relative;
  border-top: 1px solid var(--border-color);

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
    height: 1px;
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
  background: var(--bg-secondary);
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
    border-radius: 6px;
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
      border-radius: 6px;
      background: linear-gradient(135deg, #4158D0, #C850C0);
      color: white;
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
      width: 36px;
      height: 36px;
      border-radius: 8px;
      color: white;

      &.gpt {
        background: linear-gradient(135deg, #4158D0, #C850C0);
      }

      &.claude {
        background: linear-gradient(135deg, #FF4B2B, #FF416C);
      }

      &.deepseek {
        background: linear-gradient(135deg, #11998e, #38ef7d);
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

    &.active {
      background: var(--el-color-primary);
      border-color: var(--el-color-primary);
      color: white;
    }
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
</style>
