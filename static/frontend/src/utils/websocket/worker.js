// Worker 内部的连接管理
const connections = new Map();
const connectionTimes = new Map();
const maxConnections = 5;

// 获取用户的连接数量
function getUserConnectionCount(userId) {
  let count = 0;
  for (const conn of connections.values()) {
    if (conn.userId === userId) {
      count++;
    }
  }
  return count;
}

// 移除最早的连接
function removeOldestConnection(userId) {
  let oldestTime = Infinity;
  let oldestChatId = null;

  // 查找该用户最早的连接
  for (const [chatId, conn] of connections.entries()) {
    if (conn.userId === userId) {
      const time = connectionTimes.get(chatId.toString());
      if (time < oldestTime) {
        oldestTime = time;
        oldestChatId = chatId;
      }
    }
  }

  if (oldestChatId) {
    closeConnection(oldestChatId);
  }
}

// 关闭指定的连接
function closeConnection(chatId) {
  const conn = connections.get(chatId.toString());
  if (conn && conn.ws) {
    try {
      conn.ws.close();
    } catch (e) {
      console.error('Error closing WebSocket:', e);
    }
    connections.delete(chatId.toString());
    connectionTimes.delete(chatId.toString());

    // 通知主线程连接已关闭
    self.postMessage({
      type: 'close',
      chatId: chatId
    });
  }
}

// 创建新的 WebSocket 连接
async function createConnection(chatId, userId, token, presetId, wsUrl) {
  // 检查是否超过最大连接数
  if (getUserConnectionCount(userId) >= maxConnections) {
    // 移除最早的连接
    removeOldestConnection(userId);
  }

  try {
    // token 添加前缀
    if (token) {
      token = `Bearer ${token}`;
    }

    let id = chatId;
    // 如果 chatId 前缀是 tmp-，则是新会话，创建连接时 id 参数设为 -1
    if (chatId.toString().startsWith('tmp-')) {
      id = -1;
    }
    if (presetId == undefined) {
      presetId = "";
    }

    // 创建 WebSocket 连接，添加身份验证参数
    const ws = new WebSocket(`${wsUrl}?Authorization=${token}&id=${id}&presetId=${presetId}`);

    // 记录连接时间
    connectionTimes.set(chatId.toString(), Date.now());
    // 存储连接
    connections.set(chatId.toString(), {
      ws,
      userId,
      // 存储可能的部分消息
      partialMessage: {
        content: '',
        reasoningContent: ''
      }
    });

    // 设置基本事件处理
    ws.onopen = () => {
      console.log('WebSocket connection open finished:', chatId);
      self.postMessage({
        type: 'open',
        chatId: chatId
      });
    };

    ws.onerror = () => {
      self.postMessage({
        type: 'error',
        chatId: chatId,
        error: 'WebSocket error'
      });
    };

    ws.onclose = () => {
      self.postMessage({
        type: 'close',
        chatId: chatId
      });
      connections.delete(chatId.toString());
      connectionTimes.delete(chatId.toString());
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        // 处理流式响应
        if (data.conversationId) {
          const conn = connections.get(chatId.toString());
          if (conn) {
            // 累加内容
            conn.partialMessage.content += data.content || '';
            conn.partialMessage.reasoningContent += data.reasoning_content || '';

            // 如果是最后一条消息，则触发完整消息事件
            if (data.end) {
              const completeMessage = {
                type: 'chat',
                data: {
                  content: data.content,
                  reasoningContent: data.reasoningContent,
                  partialContent: conn.partialMessage.content,
                  partialReasoning: conn.partialMessage.reasoningContent,
                  conversationId: data.conversationId
                }
              };
              self.postMessage({
                type: 'message',
                chatId: chatId,
                data: completeMessage
              });

              // 重置部分消息
              conn.partialMessage = {
                content: '',
                reasoningContent: ''
              };
            } else {
              // 非最终消息，触发流式更新事件
              const streamUpdate = {
                type: 'stream',
                data: {
                  content: data.content,
                  reasoningContent: data.reasoning_content,
                  partialContent: conn.partialMessage.content,
                  partialReasoning: conn.partialMessage.reasoningContent,
                  conversationId: data.conversationId
                }
              };
              self.postMessage({
                type: 'message',
                chatId: chatId,
                data: streamUpdate
              });
            }
          }
        } else {
          // 非流式消息直接触发
          self.postMessage({
            type: 'message',
            chatId: chatId,
            data: data
          });
        }
      } catch {
        // 忽略具体错误详情，只通知出错
        self.postMessage({
          type: 'error',
          chatId: chatId,
          error: 'Error parsing message'
        });
      }
    };

    // 等待连接建立
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('WebSocket connection timeout'));
        self.postMessage({
          type: 'error',
          chatId: chatId,
          error: 'Connection timeout'
        });
      }, 5000);

      ws.onopen = () => {
        clearTimeout(timeout);
        resolve();
        self.postMessage({
          type: 'open',
          chatId: chatId
        });
      };
    });
  } catch (error) {
    self.postMessage({
      type: 'error',
      chatId: chatId,
      error: 'Failed to create connection'
    });
    throw error;
  }
}

// 发送消息
function sendMessage(chatId, data) {
  const conn = connections.get(chatId.toString());
  if (conn && conn.ws && conn.ws.readyState === WebSocket.OPEN) {
    conn.ws.send(JSON.stringify(data));
  } else {
    // 记录连接状态以帮助调试
    let connectionKeys = Array.from(connections.keys());
    let errorMessage = '';

    if (conn) {
      errorMessage = `Cannot send message: connection ${chatId} is not in OPEN state (state: ${conn.ws ? conn.ws.readyState : 'undefined'})`;
    } else {
      errorMessage = `Cannot send message: connection ${chatId} is not available. Available connections: ${connectionKeys.join(', ')}`;
    }

    self.postMessage({
      type: 'error',
      chatId: chatId,
      error: errorMessage
    });
  }
}

// 更新连接ID
function updateConnectionId(oldId, newId) {
  const conn = connections.get(oldId.toString());
  if (conn) {
    // 保存连接到新ID
    connections.set(newId.toString(), conn);

    // 移除旧连接记录
    connections.delete(oldId.toString());

    // 更新连接时间
    const time = connectionTimes.get(oldId.toString());
    if (time) {
      connectionTimes.set(newId.toString(), time);
      connectionTimes.delete(oldId.toString());
    }

    self.postMessage({
      type: 'idUpdated',
      oldId: oldId,
      newId: newId
    });
  } else {
    self.postMessage({
      type: 'error',
      chatId: oldId,
      error: `Cannot update connection ID: connection ${oldId} not found`
    });
  }
}

// 检查连接是否存在
function checkConnection(chatId) {
  const conn = connections.get(chatId.toString());
  const exists = !!(conn && conn.ws && conn.ws.readyState === WebSocket.OPEN);

  self.postMessage({
    type: 'connectionStatus',
    chatId: chatId,
    exists: exists
  });

  return exists;
}

// 处理来自主线程的消息
self.onmessage = function(e) {
  const { action, chatId, userId, data, token, presetId, oldId, newId, wsUrl } = e.data;

  switch (action) {
    case 'createConnection':
      createConnection(chatId, userId, token, presetId, wsUrl);
      break;

    case 'sendMessage':
      sendMessage(chatId, data);
      break;

    case 'closeConnection':
      closeConnection(chatId);
      break;

    case 'updateConnectionId':
      updateConnectionId(oldId, newId);
      break;

    case 'checkConnection':
      checkConnection(chatId);
      break;

    default:
      self.postMessage({
        type: 'error',
        error: `Unknown action: ${action}`
      });
  }
};
