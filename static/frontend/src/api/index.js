import apiClient from './config'
import DefaultApi from './generated/src/api/DefaultApi.js'
import {AgentApi} from "@/api/generated/src/index.js";

// 创建API实例
export const defaultApi = new DefaultApi(apiClient)

export const agentApi = new AgentApi(apiClient)

