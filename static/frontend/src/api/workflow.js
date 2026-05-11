import apiClient from '@/api/config'

export const getWorkflows = async (params) => {
  const response = await apiClient.callApi(
    '/api/workflow', 'GET',
    {}, params, {}, {}, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}

export const getWorkflow = async (id) => {
  const response = await apiClient.callApi(
    '/api/workflow/{id}', 'GET',
    { id }, {}, {}, {}, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}

export const createWorkflow = async (data) => {
  const response = await apiClient.callApi(
    '/api/workflow', 'POST',
    {}, {}, {}, {}, data,
    [], ['application/json'], ['application/json'], Object, null
  )
  return response.body || response.data
}

export const updateWorkflow = async (id, data) => {
  const response = await apiClient.callApi(
    '/api/workflow/{id}', 'PUT',
    { id }, {}, {}, {}, data,
    [], ['application/json'], ['application/json'], Object, null
  )
  return response.body || response.data
}

export const deleteWorkflows = async (id) => {
  const response = await apiClient.callApi(
    '/api/workflow/{id}', 'DELETE',
    { id }, {}, {}, {}, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}

// 获取可用模型列表
export const getWorkflowModels = async () => {
  const response = await apiClient.callApi(
    '/api/workflow/models', 'GET',
    {}, {}, {}, {}, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}

// 获取可用工具列表
export const getWorkflowTools = async () => {
  const response = await apiClient.callApi(
    '/api/workflow/tools', 'GET',
    {}, {}, {}, {}, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}

// 运行工作流
export const runWorkflow = async (id, content) => {
  const response = await apiClient.callApi(
    '/api/workflow/{id}/run', 'POST',
    { id }, {}, {}, { content }, null,
    [], [], ['application/json'], Object, null
  )
  return response.body || response.data
}
