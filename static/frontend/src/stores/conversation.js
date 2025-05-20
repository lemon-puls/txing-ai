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
      console.log('update conversation:', this.conversations[index])
    },

    async deleteConversation(id) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(conv => conv.id === id)

      if (index !== -1) {

        if (!userStore.isLoggedIn) {
          const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
          const localIndex = localConversations.findIndex(conv => conv.id === id)

          if (localIndex !== -1) {
            localConversations.splice(localIndex, 1)
            localStorage.setItem('conversations', JSON.stringify(localConversations))
          }
        } else {
          // 如果是处于登录状态，需要调用后端接口删除会话
          const response = await defaultApi.apiChatConversationsIdDelete(id)

          if (response.code != 0) {
            ElMessage.error(response.msg)
            return
          }
        }

        this.conversations.splice(index, 1)

        if (this.currentConversation?.id === id) {
          this.currentConversation = this.conversations.length > 0 ? this.conversations[0] : null
        }

        ElMessage.success('会话已删除')
      }
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
          console.log('load conversation detail response:', response)
          if (response.code === 0 && response.data) {
            console.log('load conversation detail:', response.data)
            this.currentConversation = response.data

            // 使用对象的属性访问，不需要检查类型
            const lastMessage = this.lastMessageMap[id]
            if (lastMessage) {
              // 确保 messages 数组已初始化
              if (!this.currentConversation.messages) {
                this.currentConversation.messages = []
              }
              this.currentConversation.messages.push(lastMessage)
            }

            console.log('退出 currentConversation:', this.currentConversation)
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
        maxTokens: 2048,
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
