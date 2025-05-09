/**
 * WebSocket 连接管理器 (使用 Web Worker)
 */
class WebSocketManager {
  constructor() {
    // 存储所有的回调处理器
    this.handlers = new Map()
    // 初始化 Worker
    this.initWorker()
  }

  /**
   * 初始化 Web Worker
   */
  initWorker() {
    // 创建 worker
    this.worker = new Worker(new URL('./worker.js', import.meta.url))

    // 设置 worker 消息处理
    this.worker.onmessage = (e) => {
      const { type, chatId, data, error, oldId, newId } = e.data

      switch (type) {
        case 'message':
          this.triggerHandlers(chatId, 'message', data)
          break

        case 'error':
          console.error(`WebSocket error for chat ${chatId}:`, error)
          this.triggerHandlers(chatId, 'error', error)
          break

        case 'close':
          console.log(`WebSocket connection closed for chat ${chatId}`)
          this.triggerHandlers(chatId, 'close')
          break

        case 'open':
          console.log(`WebSocket connection established for chat ${chatId}`)
          break

        case 'idUpdated':
          console.log(`WebSocket connection ID updated from ${oldId} to ${newId}`)
          break

        default:
          console.warn(`Unknown message type from worker: ${type}`)
      }
    }
  }

  /**
   * 检查是否存在指定的连接
   * @param {string} chatId - 会话ID
   * @returns {Promise<boolean>}
   */
  async hasConnection(chatId) {
    return new Promise((resolve) => {
      // 创建一次性消息处理器
      const messageHandler = (e) => {
        const { type, chatId: responseChatId, exists } = e.data;
        if (type === 'connectionStatus' && responseChatId === chatId) {
          this.worker.removeEventListener('message', messageHandler);
          resolve(exists);
        }
      };

      // 添加消息处理器
      this.worker.addEventListener('message', messageHandler);

      // 向 worker 发送查询消息
      this.worker.postMessage({
        action: 'checkConnection',
        chatId
      });
    });
  }

  /**
   * 创建新的 WebSocket 连接
   * @param {string} chatId - 会话ID
   * @param {string} userId - 用户ID
   * @returns {Promise<boolean>}
   */
  async createConnection(chatId, userId, presetId) {
    try {

      // 检查是否已经存在 WebSocket 连接
      const has = await this.hasConnection(chatId)
      if (has) {
         return false
      }

      // 获取 token
      const token = localStorage.getItem('token') || ''

      // 通过 worker 创建连接
      this.worker.postMessage({
        action: 'createConnection',
        chatId,
        userId,
        token,
        presetId
      })

      console.log(`WebSocket connection creation requested for chat ${chatId} with user ${userId}`)

      // 初始化此聊天ID的处理器集合
      if (!this.handlers.has(chatId)) {
        this.handlers.set(chatId, {
          message: new Set(),
          error: new Set(),
          close: new Set()
        })
      }

      return true
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      throw error
    }
  }

  /**
   * 关闭指定的连接
   * @param {string} chatId - 会话ID
   */
  closeConnection(chatId) {
    this.worker.postMessage({
      action: 'closeConnection',
      chatId
    })

    // 清理处理器
    this.handlers.delete(chatId)
  }

  /**
   * 发送消息
   * @param {string} chatId - 会话ID
   * @param {object} data - 要发送的数据
   */
  sendMessage(chatId, data) {
    console.log(`Sending message to chat ${chatId}:`, data)
    this.worker.postMessage({
      action: 'sendMessage',
      chatId,
      data
    })
  }

  /**
   * 添加事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型 ('message' | 'error' | 'close')
   * @param {Function} handler - 处理函数
   */
  on(chatId, event, handler) {
    if (!this.handlers.has(chatId)) {
      this.handlers.set(chatId, {
        message: new Set(),
        error: new Set(),
        close: new Set()
      })
    }

    const chatHandlers = this.handlers.get(chatId)
    if (chatHandlers && chatHandlers[event]) {
      chatHandlers[event].add(handler)
    }
  }

  /**
   * 移除事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型
   * @param {Function} handler - 处理函数
   */
  off(chatId, event, handler) {
    const chatHandlers = this.handlers.get(chatId)
    if (chatHandlers && chatHandlers[event]) {
      chatHandlers[event].delete(handler)
    }
  }

  /**
   * 触发事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型
   * @param {*} data - 事件数据
   */
  triggerHandlers(chatId, event, data) {
    const chatHandlers = this.handlers.get(chatId)
    if (chatHandlers && chatHandlers[event]) {
      chatHandlers[event].forEach(handler => handler(data))
    }
  }

  /**
   * 更新连接ID
   * @param {string} oldId - 旧ID
   * @param {string} newId - 新ID
   */
  updateConnectionId(oldId, newId) {
    // 更新 worker 中的连接 ID
    this.worker.postMessage({
      action: 'updateConnectionId',
      oldId,
      newId
    })

    // 更新处理器集合的 ID
    const handlers = this.handlers.get(oldId)
    if (handlers) {
      this.handlers.set(newId, handlers)
      this.handlers.delete(oldId)

      // 记录事件处理器数量
      for (const event of Object.keys(handlers)) {
        const count = handlers[event].size
        if (count > 0) {
          console.log(`Transferring ${count} ${event} handlers from ${oldId} to ${newId}`)
        }
      }
    } else {
      console.error(`Cannot update handlers for ID: handlers for ${oldId} not found`)
    }
  }
}

// 创建单例实例
const wsManager = new WebSocketManager()
export default wsManager
