import apiClient from './config'
import DefaultApi from './generated/src/api/DefaultApi.js'

// 创建API实例
export const defaultApi = new DefaultApi(apiClient)

