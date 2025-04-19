/**
 * WebSocket 消息类型
 */
export const MessageType = {
  CHAT: 'chat',    // 聊天消息
  STOP: 'stop',    // 停止生成
  ERROR: 'error'   // 错误消息
}

/**
 * 创建聊天消息
 * @param {string} content - 消息内容
 * @returns {object} 消息对象
 */
export function createChatMessage(content) {
  return {
    type: MessageType.CHAT,
    data: {
      content
    }
  }
}

/**
 * 创建停止生成消息
 * @returns {object} 消息对象
 */
export function createStopMessage() {
  return {
    type: MessageType.STOP
  }
}

/**
 * 创建错误消息
 * @param {string} error - 错误信息
 * @returns {object} 消息对象
 */
export function createErrorMessage(error) {
  return {
    type: MessageType.ERROR,
    data: {
      error
    }
  }
} 