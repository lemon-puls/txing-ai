import authService from './auth'
import apiClient from './config'

/**
 * 使用 fetch API 发送带有认证和拦截功能的 SSE 请求
 * @param {string} url - 请求的 URL
 * @param {FormData|Object} data - 请求的数据，可以是 FormData 或普通对象
 * @param {Function} onMessage - 处理 SSE 消息的回调函数
 * @param {Function} onError - 处理错误的回调函数
 * @param {Function} onComplete - 请求完成时的回调函数
 * @returns {AbortController} - 用于中止请求的控制器
 */
export async function fetchSSEWithAuth(url, data, onMessage, onError, onComplete) {
  const controller = new AbortController()
  const { signal } = controller

  try {
    // 获取认证头
    const authHeaders = authService.getAuthHeaders()

    // 准备请求头
    const headers = {
      'Accept': 'text/event-stream',
      ...authHeaders
    }

    // 准备请求体
    let body
    if (data instanceof FormData) {
      body = data
    } else {
      headers['Content-Type'] = 'application/json'
      body = JSON.stringify(data)
    }

    // 发送请求
    const response = await fetch(url, {
      method: 'POST',
      headers,
      body,
      signal,
    })
    
    // 处理 401 错误（未授权）
    if (response.status === 401) {
      try {
        // 使用公共模块处理401错误
        const retryFetch = async (newToken) => {
          headers['Authorization'] = `Bearer ${newToken}`
          const retryResponse = await fetch(url, {
            method: 'POST',
            headers,
            body,
            signal,
          })

          if (!retryResponse.ok) {
            throw new Error(`HTTP error! status: ${retryResponse.status}`)
          }

          handleSSEResponse(retryResponse, onMessage, onError, onComplete)
          return controller
        }

        // 调用公共模块处理401错误
        return await authService.handle401Error(
          { status: 401 },
          url,
          apiClient.callApi.bind(apiClient, '/api/user/refresh', 'POST'),
          retryFetch,
          onError
        )
      } catch (error) {
        if (onError) onError(error)
        return controller
      }
    }

    // 处理其他 HTTP 错误
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    // 处理 SSE 响应
    handleSSEResponse(response, onMessage, onError, onComplete)

  } catch (error) {
    // 处理网络错误或其他异常
    if (onError) onError(error)
  }

  return controller
}

/**
 * 处理 SSE 响应
 * @param {Response} response - fetch API 的响应对象
 * @param {Function} onMessage - 处理消息的回调
 * @param {Function} onError - 处理错误的回调
 * @param {Function} onComplete - 处理完成的回调
 */
function handleSSEResponse(response, onMessage, onError, onComplete) {
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  function processChunk({ done, value }) {
    if (done) {
      // 处理缓冲区中剩余的数据
      if (buffer.trim() && onMessage) {
        onMessage(buffer.trim())
      }

      if (onComplete) onComplete()
      return
    }

    // 解码并添加到缓冲区
    buffer += decoder.decode(value, { stream: true })

    // 处理完整的 SSE 消息
    const lines = buffer.split('\n\n')
    buffer = lines.pop() // 保留最后一个可能不完整的消息

    for (const line of lines) {
      if (line.trim() && onMessage) {
        onMessage(line.trim())
      }
    }

    // 继续读取下一个数据块
    reader.read().then(processChunk).catch(error => {
      if (onError) onError(error)
    })
  }

  reader.read().then(processChunk).catch(error => {
    if (onError) onError(error)
  })
}

export default fetchSSEWithAuth
