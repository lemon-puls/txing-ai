import { defineStore } from 'pinia'
import { useUserStore } from './user'
import {defaultApi} from "@/api/index.js";

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    // 会话列表
    conversations: [],
    // 当前会话
    currentConversation: null,
    // 加载状态
    loading: false,
    // 是否已到最后一页
    isLastPage: false,
    // 游标
    cursor: ''
  }),
  getters: {
    // 判断是否有会话
    hasConversations: (state) => state.conversations.length > 0,
    // 获取当前会话ID
    currentConversationId: (state) => state.currentConversation?.id || -1
  },
  actions: {
    // 清空会话数据
    clearConversations() {
      this.conversations = []
      this.currentConversation = null
      this.cursor = ''
      this.isLastPage = false
    },

    // 添加会话
    addConversation(conversation) {
      const userStore = useUserStore()

      // 如果是游客模式，则存储到localStorage
      if (!userStore.isLoggedIn) {
        // 生成一个本地ID
        conversation.id = Date.now()
        conversation.createTime = new Date()
        conversation.updateTime = new Date()

        // 保存到本地存储
        const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
        localConversations.unshift(conversation)
        localStorage.setItem('conversations', JSON.stringify(localConversations))
      }

      // 无论哪种模式，都更新到内存中
      this.conversations.unshift(conversation)
      this.currentConversation = conversation

      return conversation
    },

    // 更新会话
    updateConversation(conversation) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(conv => conv.id === conversation.id)

      if (index !== -1) {
        this.conversations[index] = { ...this.conversations[index], ...conversation }

        // 如果是当前会话，也要更新
        if (this.currentConversation?.id === conversation.id) {
          this.currentConversation = this.conversations[index]
        }

        // 如果是游客模式，则更新本地存储
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

    // 删除会话
    deleteConversation(id) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(conv => conv.id === id)

      if (index !== -1) {
        this.conversations.splice(index, 1)

        // 如果删除的是当前会话，则重置当前会话
        if (this.currentConversation?.id === id) {
          this.currentConversation = this.conversations.length > 0 ? this.conversations[0] : null
        }

        // 如果是游客模式，则更新本地存储
        if (!userStore.isLoggedIn) {
          const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
          const localIndex = localConversations.findIndex(conv => conv.id === id)

          if (localIndex !== -1) {
            localConversations.splice(localIndex, 1)
            localStorage.setItem('conversations', JSON.stringify(localConversations))
          }
        }
      }
    },

    // 设置当前会话
    setCurrentConversation(conversation) {
      this.currentConversation = conversation
    },

    // 添加消息到当前会话
    addMessage(message) {
      const userStore = useUserStore()

      if (!this.currentConversation) return

      if (!this.currentConversation.messages) {
        this.currentConversation.messages = []
      }

      this.currentConversation.messages.push(message)

      // 更新会话的更新时间
      this.currentConversation.updateTime = new Date()

      // 如果是游客模式，则更新本地存储
      if (!userStore.isLoggedIn) {
        const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
        const localIndex = localConversations.findIndex(conv => conv.id === this.currentConversation.id)

        if (localIndex !== -1) {
          if (!localConversations[localIndex].messages) {
            localConversations[localIndex].messages = []
          }

          localConversations[localIndex].messages.push(message)
          localConversations[localIndex].updateTime = new Date()

          localStorage.setItem('conversations', JSON.stringify(localConversations))
        }
      }
    },

    // 加载会话列表
    async loadConversations(pageSize = 20) {
      const userStore = useUserStore()
      this.loading = true

      try {
        // 如果是已登录状态，从服务器获取
        if (userStore.isLoggedIn) {
          const response = await defaultApi.apiChatConversationListPost({
            cursor: this.cursor,
            pageSize: pageSize
          })

          if (response.code === 0 && response.data) {
            // 添加新加载的会话到列表
            this.conversations = this.cursor === '' ?
              response.data.data : // 如果是第一页，直接替换
              [...this.conversations, ...response.data.data] // 否则追加
            this.cursor = response.data.cursor
            this.isLastPage = response.data.isLast

            // 如果还没有设置当前会话，且有会话数据，则设置第一个为当前会话
            if (!this.currentConversation && this.conversations.length > 0) {
              this.currentConversation = this.conversations[0]
            }
          }
        } else {
          // 如果是游客模式，从localStorage获取
          const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
          this.conversations = localConversations
          this.isLastPage = true // 本地存储只有一页

          // 如果还没有设置当前会话，且有会话数据，则设置第一个为当前会话
          if (!this.currentConversation && this.conversations.length > 0) {
            this.currentConversation = this.conversations[0]
          }
        }
      } catch (error) {
        console.error('加载会话列表失败:', error)
      } finally {
        this.loading = false
      }
    },

    // 加载会话详情
    async loadConversationDetail(id) {
      const userStore = useUserStore()

      try {
        // 如果是已登录状态，从服务器获取
        if (userStore.isLoggedIn) {
          const response = await defaultApi.apiChatConversationsIdGet(id)

          if (response.code === 0 && response.data) {
            // 更新会话详情
            const conversationDetail = response.data

            // 查找并更新当前内存中的会话
            const index = this.conversations.findIndex(conv => conv.id === id)
            if (index !== -1) {
              this.conversations[index] = { ...this.conversations[index], ...conversationDetail }
            } else {
              // 如果在当前列表中找不到，则添加到列表
              this.conversations.unshift(conversationDetail)
            }

            // 设置为当前会话
            this.currentConversation = conversationDetail
            return conversationDetail
          }
        } else {
          // 如果是游客模式，从localStorage获取
          const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')
          const conversation = localConversations.find(conv => conv.id === id)

          if (conversation) {
            // 设置为当前会话
            this.currentConversation = conversation
            return conversation
          }
        }
      } catch (error) {
        console.error('加载会话详情失败:', error)
        return null
      }

      return null
    },

    // 创建新会话
    createNewConversation(modelId = 'gpt-3.5-turbo') {
      // 创建一个新的会话对象
      const newConversation = {
        id: Date.now(), // 临时ID，后端会替换
        // 标记这是一个还没有真实ID的新会话
        realId: false,
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

    // 保存会话到本地存储
    saveConversations() {
      const userStore = useUserStore()

      // 仅在游客模式下保存到本地
      if (!userStore.isLoggedIn) {
        localStorage.setItem('conversations', JSON.stringify(this.conversations))
      }
    }
  },
  persist: {
    key: 'conversation-store',
    storage: sessionStorage, // 使用sessionStorage只在当前会话期间保留
    paths: ['currentConversation', 'cursor'] // 只持久化当前会话和游标
  }
})
