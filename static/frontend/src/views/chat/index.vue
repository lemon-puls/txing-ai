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
              <div class="message-content" :class="{ 'has-workflow': getMessageWorkflow(message) }">
                <!-- 应用标签 -->
                <div v-if="message.appName" class="message-app-tag">
                  <el-icon><Share /></el-icon>
                  <span>{{ message.appName }}</span>
                </div>
                <!-- 多模态图片显示 -->
                <div v-if="message.images && message.images.length > 0" class="message-images">
                  <div v-for="(imgUrl, idx) in message.images" :key="idx" class="image-item" @click="previewMessageImage(imgUrl)">
                    <img :src="imgUrl" :alt="`图片 ${idx + 1}`" />
                    <div class="image-overlay">
                      <el-icon><ZoomIn /></el-icon>
                    </div>
                  </div>
                </div>
                <!-- 多模态附件显示 -->
                <div v-if="message.attachments && message.attachments.length > 0" class="message-attachments">
                  <div v-for="(att, idx) in message.attachments" :key="idx" class="attachment-item" @click="downloadAttachment(att)">
                    <div class="attachment-icon" :class="getAttachmentClass(att.fileType)">
                      <el-icon><Document /></el-icon>
                    </div>
                    <div class="attachment-info">
                      <div class="attachment-name">{{ att.fileName }}</div>
                      <div class="attachment-size">{{ formatFileSize(att.fileSize) }}</div>
                    </div>
                    <el-icon class="download-icon"><Download /></el-icon>
                  </div>
                </div>
                <!-- 文件附件 -->
                <div v-if="message.files && message.files.length > 0" class="message-files">
                  <div v-for="(fileName, idx) in message.files" :key="idx" class="file-chip">
                    <el-icon><Document /></el-icon>
                    <span>{{ fileName }}</span>
                  </div>
                </div>
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
                <WorkflowMessage
                  v-if="getMessageWorkflow(message)"
                  :app-name="message.appName || ''"
                  :workflow="getMessageWorkflow(message)"
                  :artifacts="parseJsonField(message.artifacts)"
                  :node-logs="parseJsonField(message.executionLogs)"
                />
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
                <div class="action-btn" @click="showPresetMarket = true" v-permission:login>
                  <SvgIcon icon="ai" size="30" hover click/>
                </div>
              </el-tooltip>
              <div class="feature-toggles" v-if="currentModel?.tag?.includes('联网搜索')">
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
          <div class="input-wrapper" style="position: relative;">
            <!-- @ 应用选择浮层 -->
            <AppMentionPopup
              ref="appMentionPopup"
              :visible="showAppMention"
              @select="handleAppSelect"
              @close="showAppMention = false"
            />
            <!-- 应用文件上传区域 -->
            <div v-if="selectedApp && hasFileField" class="app-file-upload">
              <div v-for="field in fileFields" :key="field.name" class="file-field">
                <div v-if="appUploadedFiles[field.name]" class="uploaded-file">
                  <el-icon class="file-icon"><Document /></el-icon>
                  <span class="file-name">{{ appUploadedFiles[field.name].name }}</span>
                  <span class="file-size">{{ formatFileSize(appUploadedFiles[field.name].size) }}</span>
                  <el-icon class="remove-file" @click="removeAppFile(field.name)"><Close /></el-icon>
                </div>
                <div v-else class="upload-trigger" @click="triggerAppFileInput(field.name)">
                  <el-icon><Upload /></el-icon>
                  <span>{{ field.label || '上传文件' }}<span v-if="field.required" class="required-mark">*</span></span>
                  <span class="upload-hint" v-if="field.accept">{{ field.accept }}</span>
                </div>
              </div>
            </div>
            <!-- 多模态文件预览 -->
            <ChatFileUpload
              :files="chatFiles"
              @remove="removeChatFile"
            />
            <!-- 拖拽上传区域 -->
            <div
              ref="textareaWrapperRef"
              class="textarea-wrapper"
              :class="{ 'is-dragging': isDragging }"
              @drop="handleDrop"
              @dragover.prevent
              @dragenter="handleDragEnter"
              @dragleave="handleDragLeave"
            >
              <!-- 文件上传按钮 -->
              <div class="multimodal-actions">
                <el-tooltip v-if="isMultimodalModel" content="上传图片" placement="top">
                  <div class="upload-btn" @click="triggerFileInput('image')">
                    <el-icon><Picture /></el-icon>
                  </div>
                </el-tooltip>
                <el-tooltip content="上传文件" placement="top">
                  <div class="upload-btn" @click="triggerFileInput('file')">
                    <el-icon><Paperclip /></el-icon>
                  </div>
                </el-tooltip>
              </div>
              <div class="input-area" ref="inputAreaRef">
                <div
                  ref="editorRef"
                  class="editor-div"
                  contenteditable="true"
                  :data-placeholder="getInputPlaceholder()"
                  @input="handleEditorInput"
                  @keydown="handleInputKeydown"
                  @paste="handlePaste"
                  spellcheck="false"
                ></div>
              </div>
              <!-- 拖拽提示 -->
              <div v-if="isDragging" class="drag-overlay">
                <el-icon class="drag-icon"><Upload /></el-icon>
                <span>拖拽文件到此处上传</span>
              </div>
            </div>
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

    <!-- @ 应用提及（已通过 input-wrapper 内联渲染） -->

    <!-- Theme Drawer -->
    <ThemeDrawer
      v-model="showThemeDrawer"
    />
  </div>
</template>

<script setup name="ChatView">
import {ref, computed, onMounted, nextTick, onUnmounted, watch} from 'vue'
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
  Delete,
  Document,
  Download,
  HomeFilled,
  Paperclip,
  Picture,
  Plus,
  Position,
  RefreshRight,
  Setting,
  More,
  Close,
  Share,
  Upload,
  ZoomIn,
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
import AppMentionPopup from '@/components/chat/AppMentionPopup.vue'
import WorkflowMessage from '@/components/chat/WorkflowMessage.vue'
import ChatFileUpload from '@/components/chat/ChatFileUpload.vue'
import 'github-markdown-css/github-markdown-light.css'
import 'github-markdown-css/github-markdown-dark.css'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ThemeDrawer from '@/components/common/ThemeDrawer.vue'
import SvgIcon from "@/components/common/SvgIcon.vue";
import wsManager from '@/utils/websocket/manager'
import {createChatMessage, createStopMessage} from '@/utils/websocket/types'
import {defaultApi} from '@/api'
import {getAuthHeaders} from '@/api/auth'
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

// @ 应用提及
const showAppMention = ref(false)
const selectedApp = ref(null)
const appMentionPopup = ref(null)
const appInputSchema = ref([]) // 选中应用的 inputSchema
const appUploadedFiles = ref({}) // 选中应用上传的文件 { fieldName: File }
const textareaWrapperRef = ref(null)
const editorRef = ref(null)
const inputAreaRef = ref(null)
const appFileInputRef = ref(null)

// 多模态文件上传相关
const chatFiles = ref([]) // 当前待上传的文件列表
const isDragging = ref(false) // 拖拽状态
const chatFileInputRef = ref(null) // 文件输入引用

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

// 当前模型是否支持多模态
const isMultimodalModel = computed(() => {
  return currentModel.value?.multimodal === true
})

// 支持的文件类型
const supportedImageTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
const supportedFileTypes = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'text/plain',
  'text/markdown',
  'text/html',
  'text/csv'
]
const maxImageSize = 10 * 1024 * 1024 // 10MB
const maxFileSize = 20 * 1024 * 1024 // 20MB
const maxFiles = 10

// 检查文件类型是否支持
const isSupportedFileType = (file) => {
  return supportedImageTypes.includes(file.type) || supportedFileTypes.includes(file.type)
}

// 检查文件大小是否符合限制
const isValidFileSize = (file) => {
  if (supportedImageTypes.includes(file.type)) {
    return file.size <= maxImageSize
  }
  return file.size <= maxFileSize
}

// 获取文件类型分类
const getFileCategory = (type) => {
  if (supportedImageTypes.includes(type)) return 'image'
  return 'file'
}

// 处理文件选择
const handleFileSelect = (event) => {
  const files = Array.from(event.target.files)
  processFiles(files)
  // 清空 input 以允许重复选择同一文件
  event.target.value = ''
}

// 处理拖拽文件
const handleDrop = (event) => {
  event.preventDefault()
  isDragging.value = false
  const files = Array.from(event.dataTransfer.files)
  processFiles(files)
}

// 处理拖拽进入
const handleDragEnter = (event) => {
  event.preventDefault()
  isDragging.value = true
}

// 处理拖拽离开
const handleDragLeave = (event) => {
  event.preventDefault()
  isDragging.value = false
}

// 处理粘贴事件
const handlePaste = (event) => {
  const items = Array.from(event.clipboardData.items)
  const files = items
    .filter(item => item.kind === 'file')
    .map(item => item.getAsFile())
    .filter(Boolean)

  if (files.length > 0) {
    event.preventDefault()
    processFiles(files)
  }
}

// 处理文件列表
const processFiles = (files) => {
  // 非多模态模型过滤掉图片文件，只保留文档
  if (!isMultimodalModel.value) {
    files = files.filter(file => {
      if (supportedImageTypes.includes(file.type)) {
        return false
      }
      // 通过扩展名判断是否为图片
      const ext = file.name.split('.').pop().toLowerCase()
      const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp']
      if (imageExts.includes(ext)) {
        return false
      }
      return true
    })
    if (files.length === 0) {
      ElMessage.warning('当前模型不支持上传图片，请切换到支持多模态的模型')
      return
    }
  }

  if (chatFiles.value.length + files.length > maxFiles) {
    ElMessage.warning(`最多只能上传 ${maxFiles} 个文件`)
    return
  }

  for (const file of files) {
    console.log('处理文件:', file.name, 'MIME type:', file.type, '扩展名:', file.name.split('.').pop())

    if (!isSupportedFileType(file)) {
      // 尝试通过扩展名判断
      const ext = file.name.split('.').pop().toLowerCase()
      const extToMime = {
        'jpg': 'image/jpeg', 'jpeg': 'image/jpeg', 'png': 'image/png',
        'gif': 'image/gif', 'webp': 'image/webp',
        'pdf': 'application/pdf', 'doc': 'application/msword',
        'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'xls': 'application/vnd.ms-excel',
        'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'txt': 'text/plain', 'md': 'text/markdown', 'html': 'text/html', 'csv': 'text/csv'
      }
      const inferredType = extToMime[ext]
      if (inferredType) {
        console.log('通过扩展名推断类型:', inferredType)
        file = new File([file], file.name, { type: inferredType })
      } else {
        ElMessage.warning(`不支持的文件类型: ${file.name} (${file.type})`)
        continue
      }
    }

    if (!isValidFileSize(file)) {
      const maxSize = supportedImageTypes.includes(file.type) ? '10MB' : '20MB'
      ElMessage.warning(`文件 ${file.name} 大小超过限制 (${maxSize})`)
      continue
    }

    const category = getFileCategory(file.type)
    console.log('文件分类:', file.name, '->', category)

    const fileObj = {
      file,
      name: file.name,
      size: file.size,
      type: file.type,
      category: category,
      preview: null,
      uploading: false,
      progress: 0,
      uploaded: false,
      url: null
    }

    // 为图片生成预览
    if (fileObj.category === 'image') {
      const reader = new FileReader()
      reader.onload = (e) => {
        fileObj.preview = e.target.result
      }
      reader.readAsDataURL(file)
    }

    chatFiles.value.push(fileObj)
  }
}

// 移除文件
const removeChatFile = (index) => {
  chatFiles.value.splice(index, 1)
}

// 触发文件选择
const triggerFileInput = (type) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true

  if (type === 'image') {
    input.accept = supportedImageTypes.join(',')
  } else {
    input.accept = [...supportedImageTypes, ...supportedFileTypes].join(',')
  }

  input.onchange = handleFileSelect
  input.click()
}

// 上传文件到 COS
const uploadFilesToCOS = async () => {
  const uploadedImages = []
  const uploadedAttachments = []

  for (const fileObj of chatFiles.value) {
    if (fileObj.uploaded && fileObj.url) {
      // 已上传的文件直接使用
      if (fileObj.category === 'image') {
        uploadedImages.push(fileObj.url)
      } else {
        uploadedAttachments.push({
          fileName: fileObj.name,
          fileUrl: fileObj.url,
          fileType: fileObj.type,
          fileSize: fileObj.size
        })
      }
      continue
    }

    try {
      fileObj.uploading = true
      fileObj.progress = 0

      // 生成随机文件名
      const ext = fileObj.name.split('.').pop() || 'bin'
      const fileName = `chat/${Date.now()}-${Math.floor(Math.random() * 1000)}.${ext}`

      // 获取预签名 URL
      const res = await defaultApi.apiCosPresignedUrlPost({
        type: 'upload',
        key: fileName
      })

      if (res.code !== 0) {
        throw new Error(res.msg || '获取上传地址失败')
      }

      // 上传文件
      const response = await fetch(res.data.url, {
        method: 'PUT',
        body: fileObj.file,
        headers: {
          'Content-Type': fileObj.type
        }
      })

      if (!response.ok) {
        throw new Error('上传失败')
      }

      // 获取下载 URL
      const downloadRes = await defaultApi.apiCosPresignedUrlPost({
        key: fileName,
        type: 'download'
      })

      if (downloadRes.code !== 0) {
        throw new Error(downloadRes.msg || '获取访问地址失败')
      }

      fileObj.uploaded = true
      fileObj.url = downloadRes.data.url
      fileObj.progress = 100

      if (fileObj.category === 'image') {
        uploadedImages.push(downloadRes.data.url)
      } else {
        uploadedAttachments.push({
          fileName: fileObj.name,
          fileUrl: downloadRes.data.url,
          fileType: fileObj.type,
          fileSize: fileObj.size
        })
      }
    } catch (error) {
      console.error('文件上传失败:', error)
      ElMessage.error(`文件 ${fileObj.name} 上传失败`)
      throw error
    } finally {
      fileObj.uploading = false
    }
  }

  return { images: uploadedImages, attachments: uploadedAttachments }
}

// 发送消息
const sendMessage = async () => {
  // 应用列表弹出中，Enter 用于选中应用而非发送
  if (showAppMention.value) return
  // 同步编辑器内容（processInput 用了 rAF 节流，快速输入时可能还没同步）
  const el = editorRef.value
  if (el) {
    messageInput.value = el.innerText || ''
  }
  if ((!messageInput.value.trim() && chatFiles.value.length === 0) || !currentChat.value) return

  // 校验 @ 应用的必填输入项
  if (selectedApp.value && appInputSchema.value.length > 0) {
    // 检查必填的文件字段
    const requiredFileFields = appInputSchema.value.filter(f => f.type === 'file' && f.required)
    for (const field of requiredFileFields) {
      if (!appUploadedFiles.value[field.name]) {
        ElMessage.warning(`请上传 ${field.label || field.name}`)
        return
      }
    }

    // 检查必填的文本字段（排除 @应用名 前缀后的内容）
    const requiredTextFields = appInputSchema.value.filter(f => f.type === 'text' && f.required)
    if (requiredTextFields.length > 0) {
      // 获取实际输入内容（去掉 @应用名 前缀）
      let actualContent = messageInput.value
      if (selectedApp.value) {
        const mentionPrefix = '@' + selectedApp.value.name + ' '
        actualContent = actualContent.replace(mentionPrefix, '').trim()
      }
      if (!actualContent) {
        ElMessage.warning('请输入内容')
        return
      }
    }
  }

  // 获取用户ID (如果登录的话)
  const userId = userStore.userId || '0'
  await NewChatConnectionIfNeed(currentChat.value, userId, currentChat.value.presetId);

  // 如果是会话的第一条消息，就把该消息设置为会话的名称（剥离 @应用名 前缀）
  let displayMessage = messageInput.value
  if (selectedApp.value) {
    const mentionPrefix = '@' + selectedApp.value.name + ' '
    displayMessage = displayMessage.replace(mentionPrefix, '')
  }
  if (displayMessage.trim()) {
    conversationStore.updateCurrentChatName(displayMessage)
  }

  // 上传多模态文件（如果有）
  let uploadedImages = []
  let uploadedAttachments = []
  if (chatFiles.value.length > 0) {
    try {
      const result = await uploadFilesToCOS()
      uploadedImages = result.images
      uploadedAttachments = result.attachments
    } catch (e) {
      console.error('多模态文件上传失败:', e)
      ElMessage.warning('文件上传失败，请重试')
      return
    }
  }

  // 上传应用文件（如果有）
  let uploadedFileRefs = []
  if (selectedApp.value && Object.keys(appUploadedFiles.value).length > 0) {
    try {
      const authHeaders = getAuthHeaders()
      for (const [fieldName, file] of Object.entries(appUploadedFiles.value)) {
        const formData = new FormData()
        formData.append('file', file)
        const response = await fetch('/api/file/upload', {
          method: 'POST',
          headers: { 'Authorization': authHeaders.Authorization || '' },
          body: formData
        })
        const result = await response.json()
        if (result.code === 0 && result.data) {
          uploadedFileRefs.push({
            fieldName,
            fileUrl: result.data.fileUrl,
            fileName: result.data.fileName || file.name
          })
        } else {
          ElMessage.error(result.msg || '文件上传失败')
          return
        }
      }
    } catch (e) {
      console.error('文件上传失败:', e)
      ElMessage.warning('文件上传失败，请重试')
      return
    }
  }

  // 添加用户消息
  const userMsg = {
    id: Date.now(),
    role: 'user',
    content: messageInput.value
  }
  // 如果有上传的多模态文件，在消息中记录
  if (uploadedImages.length > 0) {
    userMsg.images = uploadedImages
  }
  if (uploadedAttachments.length > 0) {
    userMsg.attachments = uploadedAttachments
  }
  // 如果 @ 了应用，在消息中记录
  if (selectedApp.value) {
    userMsg.appName = selectedApp.value.name
    userMsg.workflowId = selectedApp.value.id
    if (uploadedFileRefs.length > 0) {
      userMsg.files = uploadedFileRefs.map(f => f.fileName)
    }
  }
  currentChat.value.messages.push(userMsg)

  // 发送到后端的消息剥离 @应用名 前缀（后端通过 workflowId 路由）
  let message = messageInput.value
  if (selectedApp.value) {
    const mentionPrefix = '@' + selectedApp.value.name + ' '
    message = message.replace(mentionPrefix, '')
  }
  messageInput.value = ''
  // 同步清空 contenteditable DOM
  const editorEl = editorRef.value
  if (editorEl) {
    editorEl.innerText = ''
    editorEl.style.height = 'auto'
  }
  chatFiles.value = [] // 清空文件列表
  showAppMention.value = false
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

    // 如果有上传的多模态文件，添加到选项中
    if (uploadedImages.length > 0) {
      options.images = uploadedImages
    }
    if (uploadedAttachments.length > 0) {
      options.attachments = uploadedAttachments
    }

    // 如果 @ 了应用，添加 workflowId 和 files
    if (selectedApp.value) {
      options.workflowId = selectedApp.value.id
      if (uploadedFileRefs.length > 0) {
        options.files = uploadedFileRefs
      }
    }

    // 确定使用哪个连接ID发送消息
    // 如果是新会话(没有真实ID)，则使用"-1"
    const connectionId = currentChat.value.id.toString()

    // 通过 WebSocket 发送消息
    wsManager.sendMessage(
      connectionId,
      createChatMessage(message, options)
    )

    // 发送后清除应用状态
    removeSelectedApp()

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
      // 更新工作流最终状态和产物
      if (data.data.workflow) {
        currentStreamingMessage.workflow = data.data.workflow
      }
      if (data.data.artifacts) {
        currentStreamingMessage.artifacts = data.data.artifacts
      }
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
        showThought: true,
        workflow: data.data.workflow || null,
        artifacts: data.data.artifacts || null
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
        showThought: true,
        workflow: null,
        artifacts: null
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

    // 更新工作流状态
    if (data.data.workflow) {
      currentStreamingMessage.workflow = data.data.workflow
    }
    if (data.data.artifacts) {
      currentStreamingMessage.artifacts = data.data.artifacts
    }

    // 同步更新 lastMessageMap 中的消息（仅限登录用户）
    const userStore = useUserStore()
    if (userStore.isLoggedIn) {
      const lastMessage = conversationStore.lastMessageMap[chatId]
      if (lastMessage) {
        lastMessage.content = data.data.partialContent
        lastMessage.reasoningContent = data.data.partialReasoning
        if (data.data.workflow) lastMessage.workflow = data.data.workflow
        if (data.data.artifacts) lastMessage.artifacts = data.data.artifacts
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

  let connCreated = await wsManager.hasConnection(newChat.id);

  if (connCreated) {
    console.log("Connection already created", newChat.id);
    return;
  }

  // 创建 WebSocket 连接
  try {
    const connected = await wsManager.createConnection(newChat.id, userId, presetId);
    if (connected) {
      console.log('WebSocket connection established successfully');
    } else {
      console.error('Failed to establish WebSocket connection');
      ElMessage.error('与服务器建立连接失败')
      return;
    }
  } catch (error) {
    console.error('Failed to establish WebSocket connection:', error);
    ElMessage.error('与服务器建立连接失败')
    return;
  }

  // 添加消息处理器
  wsManager.on(newChat.id, 'message', (data) => {

    // 如果收到的消息包含会话ID，更新当前会话ID
    if (data.conversationId) {
      const actualChatId = data.conversationId.toString()

      // 如果这是第一条消息，且ID与当前不同，需要更新会话ID
      if (newChat.realId === false && newChat.id.toString() !== actualChatId) {
        // 更新会话对象的ID
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

// 通用 JSON 字段解析（兼容字符串和已解析的数组格式）
const parseJsonField = (value) => {
  if (!value) return []
  if (Array.isArray(value)) return value
  try { return JSON.parse(value) } catch { return [] }
}

// 获取输入框占位文本
const getInputPlaceholder = () => {
  if (selectedApp.value) {
    return `向 ${selectedApp.value.name} 提问...`
  }
  return '输入消息，支持拖拽/粘贴文件，Enter 发送，Shift + Enter 换行  输入 @ 可引用应用'
}

// 获取消息的工作流状态对象（兼容原始 API 格式和已处理格式）
const getMessageWorkflow = (message) => {
  if (message.workflow) return message.workflow
  if (message.workflowStatus) return { status: message.workflowStatus }
  return null
}

// 预览消息中的图片
const previewMessageImage = (url) => {
  window.open(url, '_blank')
}

// 下载附件
const downloadAttachment = (attachment) => {
  window.open(attachment.fileUrl, '_blank')
}

// 获取附件类型样式类
const getAttachmentClass = (fileType) => {
  if (!fileType) return 'file-default'
  if (fileType.includes('pdf')) return 'file-pdf'
  if (fileType.includes('word') || fileType.includes('document')) return 'file-word'
  if (fileType.includes('excel') || fileType.includes('spreadsheet')) return 'file-excel'
  if (fileType.includes('text') || fileType.includes('markdown')) return 'file-text'
  if (fileType.includes('html')) return 'file-html'
  return 'file-default'
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

    // 根据会话的 model 字段查找并选中对应的模型
    const model = availableModels.value.find(m => m.name === chat.model)
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

  // 如果当前没有选中会话，则直接返回即可
  if (!currentChat.value) {
    return
  }

  currentChat.value.model = model?.name;
  // currentChat.value.name = model.name;

  // 更新当前会话在列表中的头像
  const chatInList = chatList.value.find(chat => chat.id === currentChat.value.id);
  if (chatInList && !currentChat.value?.preset?.id) {
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

    // 加载可用模型列表
    await loadModels()

    // 检查是否需要创建新对话
    if (route.query.newChat === 'true') {
      const presetId = route.query.presetId
      await createNewChat(presetId)
      // 清除 URL 参数，否则用户只要刷新页面就会再次创建新对话
      router.replace({ query: {} })
    } else if (chatList.value.length > 0) {
      await conversationStore.loadConversationDetail(chatList.value[0].id)

      // await NewChatConnectionIfNeed(chatList.value[0], userStore.userId, "");
    }

    const savedPattern = localStorage.getItem('chatBgPattern')
    if (savedPattern) {
      currentBgPattern.value = savedPattern
    }

    // 初始化主题
    themeStore.initTheme()

    // 设置当前选中模型
    if (currentChat.value) {
      currentModel.value = availableModels.value.find(m => m.name === currentChat.value.model) || availableModels.value[0]
      selectModel(currentModel.value)
    }
  } catch (error) {
    console.error('Failed to initialize chat:', error)
    ElMessage.error('初始化聊天失败')
  } finally {
    // 最后初始化编辑器内容
    nextTick(() => {
      initEditor()
    })
  }
})

// 添加输入框高度相关的状态和方法
const textareaRows = ref(3)
let startY = 0
let startHeight = 0
const minRows = 3
const maxRows = 15
const lineHeight = 22.4 // 14px * 1.6

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

// @ 应用相关
const handleInputKeydown = (e) => {
  if (showAppMention.value) {
    appMentionPopup.value?.handleKeydown(e)
    return
  }
  // Enter 发送，Shift+Enter 换行
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

const handleAppSelect = async (app) => {
  selectedApp.value = app
  showAppMention.value = false
  // 将 @ 替换为 @应用名
  const atIndex = messageInput.value.lastIndexOf('@')
  if (atIndex >= 0) {
    messageInput.value = messageInput.value.slice(0, atIndex) + '@' + app.name + ' '
  }
  // 同步更新 contenteditable DOM，避免 processInput 从 innerText 读到旧内容
  const el = editorRef.value
  if (el) {
    el.innerText = messageInput.value
  }
  // 加载应用的 inputSchema
  await loadAppInputSchema(app.id)
  // 聚焦到编辑器末尾
  nextTick(() => {
    if (el) {
      el.focus()
      // 将光标移到末尾
      const range = document.createRange()
      const sel = window.getSelection()
      range.selectNodeContents(el)
      range.collapse(false)
      sel.removeAllRanges()
      sel.addRange(range)
    }
  })
}

// 处理编辑器输入
let inputRafId = 0
const processInput = () => {
  inputRafId = 0
  const el = editorRef.value
  if (!el) return
  // 去掉 HTML 标签，只保留纯文本（浏览器原生渲染，不做任何 DOM 替换）
  const text = el.innerText || ''
  messageInput.value = text
  autoResizeEditor()
  // @ 检测逻辑
  const atIndex = text.lastIndexOf('@')
  // 防止 autoResizeEditor 等副作用触发的递归 input 事件在显示弹窗时错误关闭它
  const typingMention = atIndex >= 0 && !selectedApp.value
  if (selectedApp.value) {
    const mentionText = '@' + selectedApp.value.name + ' '
    if (!text.includes(mentionText)) {
      selectedApp.value = null
      appInputSchema.value = []
      appUploadedFiles.value = []
      appUploadedFiles.value = {}
    }
  } else if (atIndex >= 0) {
    showAppMention.value = true
  } else if (showAppMention.value && !typingMention) {
    // 延迟关闭：autoResizeEditor 等副作用可能在下一次 input 之前临时让 innerText 变空，
    // 推迟到下一个 microtask 看是否真要把弹窗关掉。
    Promise.resolve().then(() => {
      const cur = (editorRef.value && editorRef.value.innerText) || ''
      if (cur.lastIndexOf('@') < 0 && !selectedApp.value) {
        showAppMention.value = false
      }
    })
  }
}

const handleEditorInput = () => {
  if (inputRafId) return
  inputRafId = requestAnimationFrame(processInput)
}

// 初始化编辑器内容（仅在挂载/切换会话时调用一次）
let editorInitialized = false
const initEditor = () => {
  const el = editorRef.value
  if (!el) return
  if (messageInput.value) {
    el.innerText = messageInput.value
  }
  editorInitialized = true
}

// 自动调整编辑器高度
const autoResizeEditor = () => {
  const el = editorRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, maxRows * lineHeight) + 'px'
}

// 加载应用的 inputSchema
const loadAppInputSchema = async (appId) => {
  try {
    const res = await defaultApi.apiWorkflowPublicIdGet(appId)
    if (res.code === 0 && res.data && res.data.topology) {
      const topo = JSON.parse(res.data.topology)
      if (topo.config && Array.isArray(topo.config.inputSchema)) {
        appInputSchema.value = topo.config.inputSchema
      } else {
        appInputSchema.value = []
      }
    }
  } catch (e) {
    console.error('加载应用配置失败:', e)
    appInputSchema.value = []
  }
}

// 移除选中的应用
const removeSelectedApp = () => {
  if (selectedApp.value) {
    const mentionText = '@' + selectedApp.value.name + ' '
    messageInput.value = messageInput.value.replace(mentionText, '')
  }
  selectedApp.value = null
  appInputSchema.value = []
  appUploadedFiles.value = {}
}

// 应用文件上传
const hasFileField = computed(() => {
  return appInputSchema.value.some(f => f.type === 'file')
})

const fileFields = computed(() => {
  return appInputSchema.value.filter(f => f.type === 'file')
})

const triggerAppFileInput = (fieldName) => {
  // 动态创建 input 元素
  const input = document.createElement('input')
  input.type = 'file'
  const field = appInputSchema.value.find(f => f.name === fieldName)
  if (field && field.accept) {
    input.accept = field.accept
  }
  input.onchange = (e) => {
    const file = e.target.files[0]
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        ElMessage.warning('文件大小不能超过 10MB')
        return
      }
      appUploadedFiles.value[fieldName] = file
    }
  }
  input.click()
}

const removeAppFile = (fieldName) => {
  delete appUploadedFiles.value[fieldName]
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
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
    left: 10%;
    right: 10%;
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.15) 20%,
      rgba(var(--divider-rgb), 0.25) 50%,
      rgba(var(--divider-rgb), 0.15) 80%,
      rgba(var(--divider-rgb), 0) 100%
    );
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

  // 包含 WorkflowMessage 时撑满最大宽度
  &.has-workflow {
    width: 100%;
  }

  // WorkflowMessage 撑满宽度
  .workflow-message {
    margin-left: -16px;
    margin-right: -16px;
    padding-left: 16px;
    padding-right: 16px;
    width: calc(100% + 32px);

    .workflow-card {
      border-radius: 0;
      border-left: none;
      border-right: none;
    }
  }

  .message-app-tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    margin-bottom: 8px;
    background: rgba(25, 118, 210, 0.06);
    border-radius: 12px;
    font-size: 11px;
    color: #1976d2;

    .el-icon {
      font-size: 12px;
    }
  }

  .message-images {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;

    .image-item {
      position: relative;
      width: 120px;
      height: 120px;
      border-radius: 8px;
      overflow: hidden;
      cursor: pointer;
      border: 1px solid var(--el-border-color-lighter);
      transition: all 0.2s;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);

        .image-overlay {
          opacity: 1;
        }
      }

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .image-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.3);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transition: opacity 0.2s;

        .el-icon {
          color: #fff;
          font-size: 24px;
        }
      }
    }
  }

  .message-attachments {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;

    .attachment-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 14px;
      background: var(--el-fill-color-light);
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.2s;
      min-width: 180px;
      max-width: 260px;

      &:hover {
        background: var(--el-fill-color);
        border-color: var(--el-color-primary-light-7);
        transform: translateY(-1px);
      }

      .attachment-icon {
        width: 36px;
        height: 36px;
        border-radius: 6px;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;

        .el-icon {
          font-size: 18px;
          color: #fff;
        }

        &.file-pdf {
          background: linear-gradient(135deg, #ff4757, #ff6b81);
        }

        &.file-word {
          background: linear-gradient(135deg, #2b7de9, #5f9ee9);
        }

        &.file-excel {
          background: linear-gradient(135deg, #2ed573, #7bed9f);
        }

        &.file-text {
          background: linear-gradient(135deg, #a4b0be, #ced6e0);
        }

        &.file-html {
          background: linear-gradient(135deg, #ff6348, #ff7979);
        }

        &.file-default {
          background: linear-gradient(135deg, #747d8c, #a4b0be);
        }
      }

      .attachment-info {
        flex: 1;
        min-width: 0;

        .attachment-name {
          font-size: 13px;
          font-weight: 500;
          color: var(--el-text-color-primary);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .attachment-size {
          font-size: 11px;
          color: var(--el-text-color-secondary);
          margin-top: 2px;
        }
      }

      .download-icon {
        flex-shrink: 0;
        font-size: 16px;
        color: var(--el-text-color-secondary);
        transition: color 0.2s;
      }

      &:hover .download-icon {
        color: var(--el-color-primary);
      }
    }
  }

  .message-files {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 8px;

    .file-chip {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      padding: 3px 8px;
      background: rgba(103, 194, 58, 0.06);
      border: 1px solid rgba(103, 194, 58, 0.15);
      border-radius: 12px;
      font-size: 11px;
      color: #67c23a;

      .el-icon { font-size: 12px; }
    }
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

      table {

        tr {
          background-color: var(--var-bg-color);
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
  padding: 12px 24px 16px;
  background: var(--el-bg-color);
  position: relative;

  &::before {
    content: '';
    position: absolute;
    left: 10%;
    right: 10%;
    top: 0;
    height: 1px;
    background: linear-gradient(90deg,
      rgba(var(--divider-rgb), 0) 0%,
      rgba(var(--divider-rgb), 0.15) 20%,
      rgba(var(--divider-rgb), 0.25) 50%,
      rgba(var(--divider-rgb), 0.15) 80%,
      rgba(var(--divider-rgb), 0) 100%
    );
    z-index: 1;
  }

  .resize-handle {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    cursor: row-resize;
    z-index: 2;
    background: transparent;
    transition: background 0.2s ease;

    &:hover {
      background: linear-gradient(180deg,
        rgba(var(--divider-rgb), 0.2) 0%,
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
        rgba(var(--divider-rgb), 0.3) 20%,
        rgba(var(--divider-rgb), 0.5) 50%,
        rgba(var(--divider-rgb), 0.3) 80%,
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

  .app-file-upload {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 8px;

    .file-field {
      .uploaded-file {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 10px;
        background: rgba(103, 194, 58, 0.06);
        border: 1px solid rgba(103, 194, 58, 0.2);
        border-radius: 8px;
        font-size: 12px;
        color: #67c23a;

        .file-icon { font-size: 14px; }
        .file-name { max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .file-size { color: var(--el-text-color-secondary); }
        .remove-file {
          cursor: pointer;
          border-radius: 50%;
          transition: all 0.15s;
          &:hover { background: rgba(103, 194, 58, 0.15); }
        }
      }

      .upload-trigger {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 10px;
        background: rgba(25, 118, 210, 0.04);
        border: 1px dashed rgba(25, 118, 210, 0.3);
        border-radius: 8px;
        font-size: 12px;
        color: #1976d2;
        cursor: pointer;
        transition: all 0.15s;

        &:hover {
          background: rgba(25, 118, 210, 0.08);
          border-color: #1976d2;
        }

        .upload-hint { color: var(--el-text-color-secondary); }
        .required-mark { color: var(--el-color-danger); margin-left: 2px; }
      }
    }
  }

  .textarea-wrapper {
    position: relative;

    &.is-dragging {
      .editor-div {
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
      }
    }

    .multimodal-actions {
      display: flex;
      gap: 4px;
      margin-bottom: 8px;

      .upload-btn {
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
          transform: translateY(-1px);
        }

        .el-icon {
          font-size: 16px;
        }
      }
    }

    .drag-overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(var(--el-color-primary-rgb), 0.1);
      border: 2px dashed var(--el-color-primary);
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 8px;
      z-index: 10;
      pointer-events: none;

      .drag-icon {
        font-size: 32px;
        color: var(--el-color-primary);
      }

      span {
        font-size: 14px;
        color: var(--el-color-primary);
        font-weight: 500;
      }
    }
  }

  .input-area {
    position: relative;
  }

  .editor-div {
    width: 100%;
    min-height: 60px;
    max-height: 360px;
    overflow-y: auto;
    padding: 5px 90px 5px 11px;
    line-height: 1.6;
    font-size: 14px;
    font-family: inherit;
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    outline: none;
    box-sizing: border-box;
    word-wrap: break-word;
    white-space: pre-wrap;
    overflow-wrap: break-word;
    scrollbar-width: thin;
    scrollbar-color: var(--el-scrollbar-bar-color) transparent;

    &:focus {
      border-color: var(--el-color-primary);
      box-shadow: 0 0 0 2px var(--el-color-primary-light-8);
    }

    &:empty::before {
      content: attr(data-placeholder);
      color: var(--el-text-color-placeholder);
      pointer-events: none;
    }

    :deep(.mention-highlight) {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
      border-radius: 3px;
      padding: 0 2px;
      font-weight: 500;
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
