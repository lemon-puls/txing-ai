import { defineStore } from 'pinia'
import { useUserStore } from './user'
import { defaultApi } from '@/api'
import {ElMessage} from "element-plus";

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    conversations: [],
    currentConversation: null,
    loading: false,
    isLastPage: false,
    cursor: '',
    lastMessageMap: {},
    typingMap: {},
    streamingMessageMap: {},
  }),
  getters: {
    hasConversations: (state) => state.conversations.length > 0,
    currentConversationId: (state) => state.currentConversation?.id || -1
  },
  actions: {
    clearConversations() {
      this.conversations = []
      this.currentConversation = null
      this.cursor = ''
      this.isLastPage = false
      this.lastMessageMap = {}
      localStorage.removeItem('conversations')
    },

    addConversation(conversation) {
      const userStore = useUserStore()

      if (!userStore.isLoggedIn) {
        conversation.createTime = new Date()
        conversation.updateTime = new Date()

        const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
        localConversations.unshift(conversation)
        localStorage.setItem('conversations', JSON.stringify(localConversations))
      }

      this.conversations.unshift(conversation)
      this.currentConversation = conversation

      return conversation
    },

    updateConversation(conversation) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(conv => conv.id === conversation.id)

      if (index !== -1) {
        this.conversations[index] = { ...this.conversations[index], ...conversation }

        if (!userStore.isLoggedIn) {
          const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
          const localIndex = localConversations.findIndex(conv => conv.id === conversation.id)

          if (localIndex !== -1) {
            localConversations[localIndex] = { ...localConversations[localIndex], ...conversation }
            localStorage.setItem('conversations', JSON.stringify(localConversations))
          }
        }
      }
    },

    // 批量删除会话
    async batchDeleteConversations(ids) {
      if (!ids || ids.length === 0) {
        ElMessage.warning('请选择要删除的会话')
        return
      }

      const userStore = useUserStore()

      if (!userStore.isLoggedIn) {
        // 游客模式下，直接从本地存储和状态中删除
        const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
        const newLocalConversations = localConversations.filter(conv => !ids.includes(conv.id))
        localStorage.setItem('conversations', JSON.stringify(newLocalConversations))


      } else {
        // 登录状态下，调用后端批量删除接口

        // 过滤掉 id 不为数字字符串的会话
        // id 不为数字字符串的会话是新建会话，并且没有发送过消息（此时 id 包含 tmp 前缀）, 会话暂未保存到服务器，不用请求后端删除接口
        const filteredIds = ids.filter(id => /^\d+$/.test(id))

        if (filteredIds.length > 0) {
          const response = await defaultApi.apiChatConversationsDeletebatchPost({
            ids: filteredIds
          })

          if (response.code != 0) {
            ElMessage.error(response.msg || '删除失败')
            return
          }
        }
      }

      // 从状态中删除
      this.conversations = this.conversations.filter(conv => !ids.includes(conv.id))

      // 如果当前会话被删除，切换到第一个会话
      if (this.currentConversation && ids.includes(this.currentConversation.id)) {
        if (this.conversations.length > 0) {
          const next = this.conversations[0]
          this.currentConversation = next
          // 登录用户的列表项（ConversationSimpleVO）不含消息，
          // 需加载会话详情，否则切换后聊天记录为空
          try {
            await this.loadConversationDetail(next.id)
          } catch (error) {
            // 详情加载失败不阻断删除流程，仅记录日志
            console.error('Failed to load next conversation detail:', error)
          }
        } else {
          this.currentConversation = null
        }
      }

      ElMessage.success('删除成功')
    },

    setCurrentConversation(conversation) {
      this.currentConversation = conversation
    },

    addMessage(message) {
      const userStore = useUserStore()

      if (this.currentConversation) {
        this.currentConversation.messages.push(message)

        // 如果是游客模式，保存到本地存储
        if (!userStore.isLoggedIn) {
          this.saveToLocalStorage()
        }
      }
    },

    updateLastMessage(chatId, content) {
      const userStore = useUserStore()
      const chat = this.conversations.find(c => c.id === chatId)
      if (chat) {
        chat.lastMessage = content.substring(0, 50) + (content.length > 50 ? '...' : '')
        // 更新会话时间并重新排序
        this.updateConversationTime(chatId)

        // 如果是游客模式，保存到本地存储
        if (!userStore.isLoggedIn) {
          this.saveToLocalStorage()
        }
      }
    },

    // 更新会话时间并重新排序会话列表
    updateConversationTime(chatId) {
      const chat = this.conversations.find(c => c.id === chatId)
      if (chat) {
        // 更新时间
        chat.updateTime = new Date()

        // 从列表中移除该会话
        const index = this.conversations.findIndex(c => c.id === chatId)
        if (index !== -1) {
          this.conversations.splice(index, 1)
          // 将会话添加到列表开头
          this.conversations.unshift(chat)
        }
      }
    },

    updateConversationId(oldId, newId) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(c => c.id === oldId)
      if (index !== -1) {
        this.conversations[index].id = newId
        this.conversations[index].realId = true

        if (this.currentConversation?.id === oldId) {
          this.currentConversation.id = newId
          this.currentConversation.realId = true
        }

        // 如果是游客模式，保存到本地存储
        if (!userStore.isLoggedIn) {
          this.saveToLocalStorage()
        }
      }
    },

    async loadConversations() {
      const userStore = useUserStore()

      if (userStore.isLoggedIn) {
        try {
          const response = await defaultApi.apiChatConversationListPost({
            pageSize: 999
          })
          if (response.code === 0 && response.data) {
            this.conversations = response.data.data
          }
        } catch (error) {
          console.error('Load conversations error:', error)
          throw error
        }
      } else {
        const localConversations = localStorage.getItem('conversations')
        this.conversations = localConversations ? JSON.parse(localConversations) : []
      }
    },

    async loadConversationDetail(id) {
      const userStore = useUserStore()

      if (userStore.isLoggedIn) {
        try {
          const response = await defaultApi.apiChatConversationsIdGet(id)
          if (response.code === 0 && response.data) {
            this.currentConversation = response.data

            // 确保 messages 数组已初始化
            if (!this.currentConversation.messages) {
              this.currentConversation.messages = []
            }

            // 恢复进行中的流式消息：切换会话期间后端详情不含未完成的流式消息，
            // 切回时把仍在执行的消息对象挂回消息列表，流式更新会继续落到该对象。
            // streamingMessageMap 对所有用户都维护（lastMessageMap 仅登录用户），
            // 二者可能指向同一对象，按 id 去重
            const existingIds = new Set(this.currentConversation.messages.map(m => m.id))
            const streaming = this.streamingMessageMap[id]
            if (streaming && !existingIds.has(streaming.id)) {
              this.currentConversation.messages.push(streaming)
              existingIds.add(streaming.id)
            }
            const lastMessage = this.lastMessageMap[id]
            if (lastMessage && !existingIds.has(lastMessage.id)) {
              this.currentConversation.messages.push(lastMessage)
            }

            return response.data
          }
        } catch (error) {
          console.error('Load conversation detail error:', error)
          throw error
        }
      } else {
        const conversation = this.conversations.find(c => c.id === id)
        if (conversation) {
          this.currentConversation = conversation
          return conversation
        }
      }
      return null
    },

    createNewConversation(modelId = 'gpt-3.5-turbo') {
      const newConversation = {
        id: Date.now(),
        name: '新的会话',
        model: modelId,
        messages: [],
        createTime: new Date(),
        updateTime: new Date(),
        enableWeb: false,
        context: 10,
        maxTokens: 8192,
        temperature: 1.0,
        topP: 0.7,
        topK: 50,
        presencePenalty: 0.0,
        frequencyPenalty: 0.0,
        repetitionPenalty: 1.0
      }

      return this.addConversation(newConversation)
    },

    async saveConversation(conversation) {
      const userStore = useUserStore()

      if (!userStore.isLoggedIn) {
        const index = this.conversations.findIndex(c => c.id === conversation.id)
        if (index > -1) {
          this.conversations[index] = conversation
        } else {
          this.conversations.unshift(conversation)
        }
        localStorage.setItem('conversations', JSON.stringify(this.conversations))
      }
    },

    updateCurrentChatName(name) {
      // 如果当前会话存在且是第一条用户消息，更新会话名称
      if (this.currentConversation && this.currentConversation.messages) {
        const userMessages = this.currentConversation.messages.filter(m => m.role === 'user')
        if (userMessages.length === 0) {  // 只有一条用户消息时才更新名称
          // 最多保留 20 个字符
          name = name.substring(0, 35)
          this.currentConversation.name = name
          // 使用 updateConversation 方法来确保响应式更新
          this.updateConversation({
            id: this.currentConversation.id,
            name: name
          })
        }
      }
    },

    saveToLocalStorage() {
      localStorage.setItem('conversations', JSON.stringify(this.conversations))
    },

    setLastMessage(conversationId, message) {
      // 直接设置属性，不需要使用 Map 方法
      this.lastMessageMap[conversationId] = message
    },

    removeLastMessage(conversationId) {
      // 删除属性
      if (this.lastMessageMap[conversationId]) {
        delete this.lastMessageMap[conversationId]
      }
    },
    // 清空 lastMessageMap
    clearLastMessageMap() {
      this.lastMessageMap = {}
    },

    // 设置会话的打字状态
    setTypingStatus(conversationId, isTyping) {
      this.typingMap[conversationId] = isTyping
    },

    // 获取会话的打字状态
    getTypingStatus(conversationId) {
      return this.typingMap[conversationId] || false
    },

    // 重置所有会话的打字状态
    resetAllTypingStatus() {
      this.typingMap = {}
    },

    // 设置会话的流式消息
    setStreamingMessage(conversationId, message) {
      this.streamingMessageMap[conversationId] = message
    },

    // 获取会话的流式消息
    getStreamingMessage(conversationId) {
      return this.streamingMessageMap[conversationId] || null
    },

    // 清空所有会话的流式消息
    resetAllStreamingMessages() {
      this.streamingMessageMap = {}
    },
  },
  persist: {
    key: 'conversation-store',
    storage: sessionStorage,
    paths: ['currentConversation', 'cursor', 'lastMessageMap', 'typingMap', 'streamingMessageMap']
  }
})
