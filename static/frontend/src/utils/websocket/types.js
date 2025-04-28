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
 * @param {object} options - 可选参数
 * @returns {object} 符合后端要求的消息对象
 */
export function createChatMessage(content, options = {}) {
  const {
    model,
    enableWeb = false,
    context = 1,
    maxTokens,
    temperature,
    topP,
    topK,
    presencePenalty,
    frequencyPenalty,
    repetitionPenalty
  } = options;

  const message = {
    type: "chat",
    content: content,
    model: model || "gpt-3.5-turbo", // 默认模型
    context: context,
    enableWeb: enableWeb
  };

  // 添加可选参数
  if (maxTokens !== undefined) message.max_tokens = maxTokens;
  if (temperature !== undefined) message.temperature = temperature;
  if (topP !== undefined) message.top_p = topP;
  if (topK !== undefined) message.top_k = topK;
  if (presencePenalty !== undefined) message.presence_penalty = presencePenalty;
  if (frequencyPenalty !== undefined) message.frequency_penalty = frequencyPenalty;
  if (repetitionPenalty !== undefined) message.repetition_penalty = repetitionPenalty;

  return message;
}

/**
 * 创建停止生成消息
 * @returns {object} 停止生成消息对象
 */
export function createStopMessage() {
  return {
    type: "stop"
  };
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