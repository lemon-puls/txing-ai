import { defineStore } from 'pinia'
import { useUserStore } from './user'
import { defaultApi } from '@/api'

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    conversations: [],
    currentConversation: null,
    loading: false,
    isLastPage: false,
    cursor: ''
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

    deleteConversation(id) {
      const userStore = useUserStore()
      const index = this.conversations.findIndex(conv => conv.id === id)

      if (index !== -1) {
        this.conversations.splice(index, 1)

        if (this.currentConversation?.id === id) {
          this.currentConversation = this.conversations.length > 0 ? this.conversations[0] : null
        }

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

        // 如果是游客模式，保存到本地存储
        if (!userStore.isLoggedIn) {
          this.saveToLocalStorage()
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

    saveToLocalStorage() {
      localStorage.setItem('conversations', JSON.stringify(this.conversations))
    },
  },
  persist: {
    key: 'conversation-store',
    storage: sessionStorage,
    paths: ['currentConversation', 'cursor']
  }
})
