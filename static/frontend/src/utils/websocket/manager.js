/**
 * WebSocket 连接管理器
 */
class WebSocketManager {
  constructor() {
    // 存储所有的 WebSocket 连接
    this.connections = new Map()
    // 每个用户的最大连接数
    this.maxConnections = 5
    // 连接创建时间记录
    this.connectionTimes = new Map()
  }

  /**
   * 创建新的 WebSocket 连接
   * @param {string} chatId - 会话ID
   * @param {string} userId - 用户ID
   * @returns {Promise<WebSocket>}
   */
  async createConnection(chatId, userId) {
    // 检查是否超过最大连接数
    if (this.getUserConnectionCount(userId) >= this.maxConnections) {
      // 移除最早的连接
      this.removeOldestConnection(userId)
    }

    try {
      // 获取 token
      const token = localStorage.getItem('accessToken') || ''

      // 创建 WebSocket 连接，添加身份验证参数
      const ws = new WebSocket(`ws://localhost:8080/api/chat/ws?Authorization=${token}&id=${chatId}`)

      // 记录连接时间
      this.connectionTimes.set(chatId, Date.now())


      console.log(`WebSocket connection created for chat ${chatId} with user ${userId}`)
      // 存储连接
      this.connections.set(chatId, {
        ws,
        userId,
        handlers: {
          message: new Set(),
          error: new Set(),
          close: new Set()
        },
        // 存储可能的部分消息
        partialMessage: {
          content: '',
          reasoningContent: ''
        }
      })

      // 设置基本事件处理
      ws.onopen = () => {
        console.log(`WebSocket connection established for chat ${chatId}`)
      }

      ws.onerror = (error) => {
        console.error(`WebSocket error for chat ${chatId}:`, error)
        this.triggerHandlers(chatId, 'error', error)
      }

      ws.onclose = () => {
        console.log(`WebSocket connection closed for chat ${chatId}`)
        this.triggerHandlers(chatId, 'close')
        this.removeConnection(chatId)
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)

          // 处理流式响应
          if (data.conversationId) {
            const conn = this.connections.get(chatId)
            if (conn) {
              // 累加内容
              conn.partialMessage.content += data.content || ''
              conn.partialMessage.reasoningContent += data.reasoning_content || ''

              // 如果是最后一条消息，则触发完整消息事件
              if (data.end) {
                const completeMessage = {
                  type: 'chat',
                  data: {
                    content: conn.partialMessage.content,
                    thought_process: conn.partialMessage.reasoningContent
                  }
                }
                this.triggerHandlers(chatId, 'message', completeMessage)

                // 重置部分消息
                conn.partialMessage = {
                  content: '',
                  reasoningContent: ''
                }
              } else {
                // 非最终消息，触发流式更新事件
                const streamUpdate = {
                  type: 'stream',
                  data: {
                    content: data.content,
                    reasoning_content: data.reasoning_content,
                    partial_content: conn.partialMessage.content,
                    partial_reasoning: conn.partialMessage.reasoningContent
                  }
                }
                this.triggerHandlers(chatId, 'message', streamUpdate)
              }
            }
          } else {
            // 非流式消息直接触发
            this.triggerHandlers(chatId, 'message', data)
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error, event.data)
        }
      }

      // 等待连接建立
      await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('WebSocket connection timeout'))
        }, 5000)

        ws.onopen = () => {
          clearTimeout(timeout)
          resolve()
        }
      })

      return ws
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      throw error
    }
  }

  /**
   * 获取用户的连接数量
   * @param {string} userId - 用户ID
   * @returns {number}
   */
  getUserConnectionCount(userId) {
    let count = 0
    for (const conn of this.connections.values()) {
      if (conn.userId === userId) {
        count++
      }
    }
    return count
  }

  /**
   * 移除最早的连接
   * @param {string} userId - 用户ID
   */
  removeOldestConnection(userId) {
    let oldestTime = Infinity
    let oldestChatId = null

    // 查找该用户最早的连接
    for (const [chatId, conn] of this.connections.entries()) {
      if (conn.userId === userId) {
        const time = this.connectionTimes.get(chatId)
        if (time < oldestTime) {
          oldestTime = time
          oldestChatId = chatId
        }
      }
    }

    if (oldestChatId) {
      this.closeConnection(oldestChatId)
    }
  }

  /**
   * 关闭指定的连接
   * @param {string} chatId - 会话ID
   */
  closeConnection(chatId) {
    const conn = this.connections.get(chatId)
    if (conn) {
      conn.ws.close()
      this.removeConnection(chatId)
    }
  }

  /**
   * 移除连接记录
   * @param {string} chatId - 会话ID
   */
  removeConnection(chatId) {
    this.connections.delete(chatId)
    this.connectionTimes.delete(chatId)
  }

  /**
   * 发送消息
   * @param {string} chatId - 会话ID
   * @param {object} data - 要发送的数据
   */
  sendMessage(chatId, data) {
    const conn = this.connections.get(Number(chatId))
    if (conn && conn.ws.readyState === WebSocket.OPEN) {
      conn.ws.send(JSON.stringify(data))
    } else {
      // 如果指定的连接不可用，检查是否有其他可能的连接
      if (chatId !== "-1" && this.connections.has("-1")) {
        console.log(`Connection ${chatId} not available, trying with temporary ID "-1"`)
        const tempConn = this.connections.get("-1")
        if (tempConn && tempConn.ws.readyState === WebSocket.OPEN) {
          tempConn.ws.send(JSON.stringify(data))
          return
        }
      }

      // 记录连接状态以帮助调试
      if (conn) {
        console.error(`Cannot send message: connection ${chatId} is not in OPEN state (state: ${conn.ws.readyState})`)
      } else {
        console.error(`Cannot send message: connection ${chatId} is not available. Available connections: ${Array.from(this.connections.keys()).join(', ')}`)
      }
    }
  }

  /**
   * 添加事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型 ('message' | 'error' | 'close')
   * @param {Function} handler - 处理函数
   */
  on(chatId, event, handler) {
    const conn = this.connections.get(chatId)
    if (conn && conn.handlers[event]) {
      conn.handlers[event].add(handler)
    }
  }

  /**
   * 移除事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型
   * @param {Function} handler - 处理函数
   */
  off(chatId, event, handler) {
    const conn = this.connections.get(chatId)
    if (conn && conn.handlers[event]) {
      conn.handlers[event].delete(handler)
    }
  }

  /**
   * 触发事件处理器
   * @param {string} chatId - 会话ID
   * @param {string} event - 事件类型
   * @param {*} data - 事件数据
   */
  triggerHandlers(chatId, event, data) {
    const conn = this.connections.get(chatId)
    if (conn && conn.handlers[event]) {
      conn.handlers[event].forEach(handler => handler(data))
    }
  }

  /**
   * 更新连接ID
   * @param {string} oldId - 旧ID
   * @param {string} newId - 新ID
   */
  updateConnectionId(oldId, newId) {
    const conn = this.connections.get(oldId)
    if (conn) {
      // 保存连接到新ID
      this.connections.set(newId, conn)

      // 记录事件处理器数量
      for (const event of Object.keys(conn.handlers)) {
        const count = conn.handlers[event].size
        if (count > 0) {
          console.log(`Transferring ${count} ${event} handlers from ${oldId} to ${newId}`)
        }
      }

      // 移除旧连接
      this.connections.delete(oldId)

      // 更新连接时间
      const time = this.connectionTimes.get(oldId)
      if (time) {
        this.connectionTimes.set(newId, time)
        this.connectionTimes.delete(oldId)
      }

      console.log(`WebSocket connection ID updated from ${oldId} to ${newId}`)
    } else {
      console.error(`Cannot update connection ID: connection ${oldId} not found`)
    }
  }
}

// 创建单例实例
const wsManager = new WebSocketManager()
export default wsManager
