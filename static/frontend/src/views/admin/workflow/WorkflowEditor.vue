<template>
  <div class="workflow-editor-container">
    <!-- 顶部工具栏 -->
    <div class="header">
      <div class="title">
        <el-button icon="ArrowLeft" link @click="goBack">返回</el-button>
        <el-input
          v-model="workflowName"
          class="workflow-name-input"
          size="small"
          placeholder="未命名工作流"
        />
      </div>
      <div class="actions">
        <el-button type="danger" plain :disabled="!selectedNode && !selectedEdgeId" @click="deleteSelected">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
        <el-button type="warning" plain @click="runLLMValidation" :loading="llmValidating">
          <el-icon><MagicStick /></el-icon>
          AI 校验
        </el-button>
        <el-button type="success" @click="openTestDialog">
          <el-icon><VideoPlay /></el-icon>
          运行测试
        </el-button>
        <el-button type="primary" @click="saveWorkflow" :loading="saving">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </div>
    </div>

    <!-- 主体区域 -->
    <div class="editor-main">
      <!-- 左侧节点库 -->
      <NodeSidebar
        :saving="saving"
        :model-list="modelList"
        :workflow-config="workflowConfig"
        @save="saveWorkflow"
        @config-change="handleConfigChange"
      />

      <!-- 中间画布 -->
      <div class="canvas-area" @drop="onDrop" @dragover.prevent @keydown="onKeyDown" tabindex="0" ref="canvasRef">
        <VueFlow
          v-model:nodes="nodes"
          v-model:edges="edges"
          :default-zoom="1"
          :min-zoom="0.2"
          :max-zoom="4"
          :node-types="nodeTypes"
          :connection-line-style="connectionLineStyle"
          :default-edge-options="defaultEdgeOptions"
          :snap-to-grid="true"
          :snap-grid="[15, 15]"
          :fit-view-on-init="true"
          @node-click="onNodeClick"
          @edge-click="onEdgeClick"
          @pane-click="onPaneClick"
          @connect="onConnect"
          @paneReady="onPaneReady"
        >
          <Background pattern-color="#e0e0e0" :gap="20" />
          <Controls />
        </VueFlow>

        <!-- 运行结果底部面板 -->
        <transition name="slide-up">
          <div v-if="testRunning || testResult" class="test-result-panel">
            <div class="result-panel-header">
              <div class="result-panel-title">
                <span class="status-dot" :class="{ running: testRunning, success: !testRunning && !testError, error: testError }"></span>
                <span>{{ testRunning ? '工作流运行中...' : (testError ? '运行失败' : '运行完成') }}</span>
              </div>
              <div class="result-panel-actions">
                <el-button v-if="executionLogs.length > 0" text size="small" @click="executionLogVisible = !executionLogVisible">
                  <el-icon><Document /></el-icon>
                  {{ executionLogVisible ? '隐藏日志' : '查看日志' }}
                </el-button>
                <el-button v-if="!testRunning" text size="small" @click="clearTestResult">
                  <el-icon><Delete /></el-icon>
                  清除
                </el-button>
                <el-button text size="small" @click="testResultPanelCollapsed = !testResultPanelCollapsed">
                  <el-icon>
                    <ArrowDown v-if="testResultPanelCollapsed" />
                    <ArrowUp v-else />
                  </el-icon>
                </el-button>
              </div>
            </div>
            <div v-show="!testResultPanelCollapsed" class="result-panel-body">
              <div v-if="testRunning" class="running-indicator">
                <el-icon class="loading-icon"><Loading /></el-icon>
                <span>正在执行工作流，请观察画布中的节点状态...</span>
              </div>
              <div v-else-if="testError" class="error-message">
                <el-icon><WarningFilled /></el-icon>
                <span>{{ testError }}</span>
              </div>
              <div v-else class="result-text">
                {{ testResult }}
              </div>
            </div>
          </div>
        </transition>

        <!-- 执行日志面板 -->
        <ExecutionLogPanel
          :visible="executionLogVisible"
          :logs="executionLogs"
          @close="executionLogVisible = false"
          @clear="clearExecutionLogs"
        />
      </div>

      <!-- 右侧属性面板 -->
      <PropertyPanel
        :selected-node="selectedNode"
        :model-list="modelList"
        :tool-list="toolList"
        @update="handleNodeUpdate"
        @close="selectedNode = null"
      />
    </div>

    <!-- 运行测试对话框 -->
    <el-dialog
      v-model="testDialogVisible"
      title="运行工作流测试"
      width="420px"
      :close-on-click-modal="false"
      class="test-dialog"
    >
      <el-form label-position="top">
        <el-form-item label="输入内容">
          <el-input
            v-model="testInput"
            type="textarea"
            :rows="3"
            placeholder="请输入要测试的内容，例如：请介绍一下人工智能的发展历史"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="testDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="startTestRun" :disabled="!testInput">
          开始运行
        </el-button>
      </template>
    </el-dialog>

    <!-- 校验结果对话框 -->
    <el-dialog
      v-model="validationDialogVisible"
      :title="validationForSave ? '保存前校验结果' : 'AI 校验结果'"
      width="560px"
      :close-on-click-modal="false"
      class="validation-dialog"
    >
      <div v-if="validationResult" class="validation-content">
        <!-- 校验状态 -->
        <div class="validation-status" :class="{ valid: validationResult.valid, invalid: !validationResult.valid }">
          <el-icon v-if="validationResult.valid" :size="20"><CircleCheck /></el-icon>
          <el-icon v-else :size="20"><WarningFilled /></el-icon>
          <span>{{ validationResult.valid ? '校验通过' : '校验未通过' }}</span>
        </div>

        <!-- 错误列表 -->
        <div v-if="validationResult.errors && validationResult.errors.length > 0" class="validation-section">
          <div class="section-title error-title">
            <el-icon><WarningFilled /></el-icon>
            <span>错误 ({{ validationResult.errors.length }})</span>
          </div>
          <div class="issue-list">
            <div
              v-for="(error, index) in validationResult.errors"
              :key="'error-' + index"
              class="issue-item error-item"
              @click="locateNode(error.nodeId)"
            >
              <span class="issue-badge error">错误</span>
              <span class="issue-message">{{ error.message }}</span>
              <el-button v-if="error.nodeId" text size="small" type="primary" class="locate-btn">
                定位
              </el-button>
            </div>
          </div>
        </div>

        <!-- 警告列表 -->
        <div v-if="validationResult.warnings && validationResult.warnings.length > 0" class="validation-section">
          <div class="section-title warning-title">
            <el-icon><WarningFilled /></el-icon>
            <span>警告 ({{ validationResult.warnings.length }})</span>
          </div>
          <div class="issue-list">
            <div
              v-for="(warning, index) in validationResult.warnings"
              :key="'warning-' + index"
              class="issue-item warning-item"
              @click="locateNode(warning.nodeId)"
            >
              <span class="issue-badge warning">警告</span>
              <span class="issue-message">{{ warning.message }}</span>
              <el-button v-if="warning.nodeId" text size="small" type="primary" class="locate-btn">
                定位
              </el-button>
            </div>
          </div>
        </div>

        <!-- 全部通过 -->
        <div v-if="validationResult.valid && (!validationResult.errors || validationResult.errors.length === 0) && (!validationResult.warnings || validationResult.warnings.length === 0)" class="validation-all-pass">
          <el-icon :size="48" color="#4caf50"><CircleCheck /></el-icon>
          <p>所有校验项目均通过</p>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="validationDialogVisible = false">关闭</el-button>
          <el-button
            v-if="validationForSave && !validationResult.valid"
            type="warning"
            plain
            @click="forceSaveWorkflow"
          >
            强制保存
          </el-button>
          <el-button
            v-if="validationForSave && validationResult.valid"
            type="primary"
            @click="forceSaveWorkflow"
          >
            继续保存
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, computed, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { Check, VideoPlay, Loading, WarningFilled, Delete, ArrowUp, ArrowDown, CircleCheck, MagicStick, Document } from '@element-plus/icons-vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'

// 导入自定义节点组件
import StartNode from '@/components/workflow/StartNode.vue'
import EndNode from '@/components/workflow/EndNode.vue'
import LLMNode from '@/components/workflow/LLMNode.vue'
import ToolNode from '@/components/workflow/ToolNode.vue'
import ConditionNode from '@/components/workflow/ConditionNode.vue'
import CodeNode from '@/components/workflow/CodeNode.vue'
import HTTPNode from '@/components/workflow/HTTPNode.vue'
import NodeSidebar from '@/components/workflow/NodeSidebar.vue'
import PropertyPanel from '@/components/workflow/PropertyPanel.vue'
import ExecutionLogPanel from '@/components/workflow/ExecutionLogPanel.vue'

import { getWorkflow, updateWorkflow, getWorkflowModels, getWorkflowTools, runWorkflow, validateWorkflow, validateWorkflowById } from '@/api/workflow'
import { useUserStore } from '@/stores/user'
import authService from '@/api/auth.js'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const workflowId = route.params.id

const workflowName = ref('')
const saving = ref(false)

// Vue Flow 实例
const { project, fitView, addSelectedEdges, removeSelectedEdges } = useVueFlow()

// 画布 ref
const canvasRef = ref(null)

// 节点和边数据
const nodes = ref([])
const edges = ref([])
const selectedNode = ref(null)
const selectedEdgeId = ref(null)

// 模型和工具列表
const modelList = ref([])
const toolList = ref([])

// 工作流配置
const workflowConfig = ref({
  defaultModel: '',
  maxRunSteps: 30
})

// 工作流运行状态跟踪
const runningNodeId = ref(null)
const completedNodeIds = ref(new Set())
const failedNodeIds = ref(new Set())
const activeEdgeIds = ref(new Set())

// 测试相关
const testDialogVisible = ref(false)
const testInput = ref('')
const testRunning = ref(false)
const testResult = ref('')
const testError = ref('')
const testResultPanelCollapsed = ref(false)

// 执行日志相关
const executionLogs = ref([])
const executionLogVisible = ref(false)

// 校验相关
const validating = ref(false)
const llmValidating = ref(false)
const validationResult = ref(null)
const validationDialogVisible = ref(false)
const validationForSave = ref(false) // 是否是保存前触发的校验

// 注册自定义节点类型
const nodeTypes = {
  start: markRaw(StartNode),
  end: markRaw(EndNode),
  llm: markRaw(LLMNode),
  tool: markRaw(ToolNode),
  condition: markRaw(ConditionNode),
  code: markRaw(CodeNode),
  http: markRaw(HTTPNode)
}

// 连接线样式
const connectionLineStyle = { stroke: '#1976d2', strokeWidth: 2 }

// 默认边配置
const defaultEdgeOptions = {
  type: 'smoothstep',
  animated: true,
  style: { stroke: '#1976d2', strokeWidth: 2 },
  markerEnd: {
    type: 'arrowclosed',
    color: '#1976d2'
  }
}

// 迁移旧的 handle ID 到新的
const migrateHandleId = (handleId, type) => {
  if (!handleId) return type === 'source' ? 'output-right' : 'input-left'
  if (handleId === 'output') return 'output-right'
  if (handleId === 'input') return 'input-left'
  return handleId
}

// 加载工作流数据
const loadWorkflow = async () => {
  if (!workflowId) return
  try {
    const res = await getWorkflow(workflowId)
    if (res.code === 0 && res.data) {
      workflowName.value = res.data.name || ''
      if (res.data.topology) {
        try {
          const flowData = JSON.parse(res.data.topology)
          nodes.value = flowData.nodes || []
          // 迁移旧的 handle ID
          edges.value = (flowData.edges || []).map(edge => ({
            ...edge,
            sourceHandle: migrateHandleId(edge.sourceHandle, 'source'),
            targetHandle: migrateHandleId(edge.targetHandle, 'target')
          }))
          // 加载工作流配置
          if (flowData.config) {
            workflowConfig.value = {
              defaultModel: flowData.config.defaultModel || '',
              maxRunSteps: flowData.config.maxRunSteps || 30
            }
          }
        } catch (e) {
          console.error('解析工作流数据失败:', e)
        }
      }
    } else {
      ElMessage.error(res.msg || '获取工作流信息失败')
    }
  } catch (error) {
    console.error('加载失败:', error)
    ElMessage.error('加载工作流失败')
  }
}

// 加载模型和工具列表
const loadOptions = async () => {
  try {
    // 并行加载模型和工具
    const [modelRes, toolRes] = await Promise.all([
      getWorkflowModels(),
      getWorkflowTools()
    ])

    if (modelRes.code === 0 && modelRes.data) {
      modelList.value = modelRes.data
    }
    if (toolRes.code === 0 && toolRes.data) {
      toolList.value = toolRes.data
    }
  } catch (error) {
    console.error('加载配置列表失败:', error)
  }
}

// 拖拽放置
const onDrop = (event) => {
  const type = event.dataTransfer?.getData('application/vueflow')
  if (!type) return

  const canvasBounds = document.querySelector('.canvas-area').getBoundingClientRect()
  const x = event.clientX - canvasBounds.left
  const y = event.clientY - canvasBounds.top
  const position = project({ x, y })

  const newNode = {
    id: `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    type,
    position,
    data: {
      label: getDefaultLabel(type),
      nodeType: type,
      // 根据类型初始化配置
      ...(type === 'llm' ? {
        modelConfig: {
          model: '',
          systemPrompt: '',
          temperature: 0.7,
          maxTokens: 4096,
          contextEnabled: true
        }
      } : {}),
      ...(type === 'tool' ? {
        toolConfig: {
          tools: [],
          params: {}
        }
      } : {}),
      ...(type === 'condition' ? {
        conditionConfig: {
          type: 'expression',
          expression: '',
          llmPrompt: '',
          toolName: '',
          toolResultKey: '',
          expectedValue: '',
          failureAction: 'default_false',
          failureBranch: 'false'
        }
      } : {}),
      ...(type === 'code' ? {
        codeConfig: {
          language: 'javascript',
          code: '// 在此编写代码\n// 输入变量: input\n// 输出: return 结果\nreturn input;',
          timeout: 30
        }
      } : {}),
      ...(type === 'http' ? {
        httpConfig: {
          method: 'GET',
          url: '',
          headers: {},
          body: '',
          timeout: 30
        }
      } : {})
    }
  }

  nodes.value.push(newNode)
}

const getDefaultLabel = (type) => {
  const map = {
    start: '开始',
    end: '结束',
    llm: '大模型',
    tool: '工具',
    condition: '条件分支',
    code: '代码',
    http: 'HTTP'
  }
  return map[type] || '未知节点'
}

// 节点点击
const onNodeClick = ({ node }) => {
  selectedNode.value = node
  selectedEdgeId.value = null
  removeSelectedEdges()
}

// 边点击 - 选中边
const onEdgeClick = ({ edge }) => {
  selectedNode.value = null
  selectedEdgeId.value = edge.id
  addSelectedEdges([edge])
}

// 画布点击（取消选中）
const onPaneClick = () => {
  selectedNode.value = null
  selectedEdgeId.value = null
  removeSelectedEdges()
}

// 删除选中的节点
const deleteSelectedNode = () => {
  if (!selectedNode.value) return
  const nodeId = selectedNode.value.id
  nodes.value = nodes.value.filter(n => n.id !== nodeId)
  edges.value = edges.value.filter(e => e.source !== nodeId && e.target !== nodeId)
  selectedNode.value = null
}

// 删除选中的边
const deleteSelectedEdge = () => {
  if (!selectedEdgeId.value) return
  edges.value = edges.value.filter(e => e.id !== selectedEdgeId.value)
  selectedEdgeId.value = null
  removeSelectedEdges()
}

// 通用删除（节点优先，其次边）
const deleteSelected = () => {
  if (selectedNode.value) {
    deleteSelectedNode()
  } else if (selectedEdgeId.value) {
    deleteSelectedEdge()
  }
}

// 键盘事件
const onKeyDown = (event) => {
  if (event.key === 'Delete' || event.key === 'Backspace') {
    if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') return
    deleteSelected()
  }
}

// 连线事件
const onConnect = (params) => {
  // 映射旧的 handle ID 到新的
  const mapHandle = (handleId, type) => {
    if (!handleId) return type === 'source' ? 'output-right' : 'input-left'
    if (handleId === 'output') return 'output-right'
    if (handleId === 'input') return 'input-left'
    return handleId
  }

  const newEdge = {
    id: `edge_${Date.now()}`,
    source: params.source,
    target: params.target,
    sourceHandle: mapHandle(params.sourceHandle, 'source'),
    targetHandle: mapHandle(params.targetHandle, 'target'),
    type: 'smoothstep',
    animated: true
  }
  edges.value.push(newEdge)
}

// 节点数据更新
const handleNodeUpdate = ({ id, data }) => {
  const nodeIndex = nodes.value.findIndex(n => n.id === id)
  if (nodeIndex !== -1) {
    nodes.value[nodeIndex] = {
      ...nodes.value[nodeIndex],
      data: { ...data }
    }
    selectedNode.value = null
  }
}

// 处理工作流配置变更
const handleConfigChange = (config) => {
  workflowConfig.value = { ...config }
}

// 保存工作流（先校验再保存）
const saveWorkflow = async () => {
  if (!workflowId) return

  // 先进行结构校验
  const canSave = await validateBeforeSave()
  if (!canSave) return

  await doSaveWorkflow()
}

// 实际保存逻辑
const doSaveWorkflow = async () => {
  if (!workflowId) return
  saving.value = true
  try {
    const flowData = {
      nodes: nodes.value,
      edges: edges.value,
      config: workflowConfig.value
    }
    const res = await updateWorkflow(workflowId, {
      name: workflowName.value || '未命名工作流',
      description: '',
      topology: JSON.stringify(flowData)
    })

    if (res.code === 0) {
      ElMessage.success('保存成功')
    } else {
      ElMessage.error(res.msg || '保存失败')
    }
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存工作流失败')
  } finally {
    saving.value = false
  }
}

// 画布就绪
const onPaneReady = () => {
  fitView({ padding: 0.2 })
}

const goBack = () => {
  router.push('/admin/workflow')
}

// 清除运行状态
const clearExecutionState = () => {
  runningNodeId.value = null
  completedNodeIds.value = new Set()
  failedNodeIds.value = new Set()
  activeEdgeIds.value = new Set()
  executionLogs.value = []
  executionLogVisible.value = false
  // 重置所有节点和边的 class
  nodes.value = nodes.value.map(node => ({ ...node, class: '' }))
  edges.value = edges.value.map(edge => ({ ...edge, class: '' }))
}

// 处理节点状态变化
const handleNodeStatus = (nodeId, nodeStatus) => {
  if (nodeStatus === 'running') {
    runningNodeId.value = nodeId
    completedNodeIds.value.delete(nodeId)
    failedNodeIds.value.delete(nodeId)
    // 激活从该节点出发的边
    const outEdges = edges.value.filter(e => e.source === nodeId)
    outEdges.forEach(e => activeEdgeIds.value.add(e.id))
  } else if (nodeStatus === 'completed') {
    completedNodeIds.value.add(nodeId)
    if (runningNodeId.value === nodeId) {
      runningNodeId.value = null
    }
    // 将从该节点出发的边标记为已完成（从 active 中移除）
    const outEdges = edges.value.filter(e => e.source === nodeId)
    outEdges.forEach(e => activeEdgeIds.value.delete(e.id))
  } else if (nodeStatus === 'failed') {
    failedNodeIds.value.add(nodeId)
    if (runningNodeId.value === nodeId) {
      runningNodeId.value = null
    }
    const outEdges = edges.value.filter(e => e.source === nodeId)
    outEdges.forEach(e => activeEdgeIds.value.delete(e.id))
  }
}

// 更新节点 CSS class
const updateNodeClasses = () => {
  nodes.value = nodes.value.map(node => {
    const classes = []
    if (node.id === runningNodeId.value) classes.push('node-running')
    if (completedNodeIds.value.has(node.id)) classes.push('node-completed')
    if (failedNodeIds.value.has(node.id)) classes.push('node-failed')
    return { ...node, class: classes.join(' ') }
  })
}

// 更新边 CSS class
const updateEdgeClasses = () => {
  edges.value = edges.value.map(edge => {
    const classes = []
    if (activeEdgeIds.value.has(edge.id)) classes.push('edge-active')
    return { ...edge, class: classes.join(' ') }
  })
}

watch([runningNodeId, completedNodeIds, failedNodeIds], updateNodeClasses, { deep: true })
watch(activeEdgeIds, updateEdgeClasses, { deep: true })

// 打开测试对话框
const openTestDialog = () => {
  testInput.value = ''
  testResult.value = ''
  testError.value = ''
  testResultPanelCollapsed.value = false
  testDialogVisible.value = true
}

// 从对话框开始运行
const startTestRun = () => {
  testDialogVisible.value = false
  testResultPanelCollapsed.value = false
  runWorkflowTest()
}

// 清除测试结果
const clearTestResult = () => {
  testResult.value = ''
  testError.value = ''
  clearExecutionState()
}

const clearExecutionLogs = () => {
  executionLogs.value = []
}

// 运行工作流测试
const runWorkflowTest = async () => {
  if (!testInput.value || !workflowId) return

  // 清除上次运行状态
  clearExecutionState()

  testRunning.value = true
  testResult.value = ''
  testError.value = ''

  try {
    // 使用 authService 获取认证头
    const authHeaders = authService.getAuthHeaders()
    if (!authHeaders.Authorization) {
      testError.value = '未登录，请先登录'
      testRunning.value = false
      return
    }

    // 调用工作流运行接口
    const url = `${window.location.origin}/api/workflow/${workflowId}/run`
    console.log('开始运行工作流:', url)

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Authorization': authHeaders.Authorization,
        'Content-Type': 'application/x-www-form-urlencoded',
        'Accept': 'text/event-stream'
      },
      body: `content=${encodeURIComponent(testInput.value)}`
    })

    console.log('响应状态:', response.status)

    if (!response.ok) {
      const errorText = await response.text()
      console.error('请求失败:', errorText)
      throw new Error(`请求失败: ${response.status} - ${errorText}`)
    }

    // 读取 SSE 流
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmedLine = line.trim()
        if (trimmedLine.startsWith('data: ')) {
          const data = trimmedLine.slice(6)
          try {
            const event = JSON.parse(data)
            console.log('收到事件:', event)
            if (event.end) {
              testRunning.value = false
              // 运行结束，清除运行中状态
              runningNodeId.value = null
              activeEdgeIds.value = new Set()
              // 自动显示执行日志
              if (executionLogs.value.length > 0) {
                executionLogVisible.value = true
              }
              ElMessage.success('工作流执行完成')
              return
            }
            if (event.error) {
              testError.value = event.error
              testRunning.value = false
              runningNodeId.value = null
              activeEdgeIds.value = new Set()
              return
            }
            // 处理节点状态事件
            if (event.nodeId && event.nodeStatus) {
              handleNodeStatus(event.nodeId, event.nodeStatus)
            }
            // 收集执行日志
            if (event.execution_log) {
              const logIndex = executionLogs.value.findIndex(
                log => log.nodeId === event.nodeId && log.status === 'running'
              )
              if (logIndex !== -1) {
                // 更新现有的运行中日志
                executionLogs.value[logIndex] = {
                  ...executionLogs.value[logIndex],
                  status: event.nodeStatus || 'completed',
                  endTime: event.execution_log.endTime,
                  duration: event.execution_log.duration,
                  input: event.execution_log.input,
                  output: event.execution_log.output,
                  error: event.execution_log.error,
                  retry: event.execution_log.retry
                }
              } else {
                // 添加新的日志
                executionLogs.value.push({
                  nodeId: event.nodeId,
                  nodeType: event.nodeType,
                  nodeLabel: event.nodeLabel,
                  status: event.nodeStatus || 'completed',
                  startTime: event.execution_log.startTime,
                  endTime: event.execution_log.endTime,
                  duration: event.execution_log.duration,
                  input: event.execution_log.input,
                  output: event.execution_log.output,
                  error: event.execution_log.error,
                  retry: event.execution_log.retry
                })
              }
            } else if (event.nodeId && event.nodeStatus === 'running') {
              // 节点开始运行时添加日志
              const existingLog = executionLogs.value.find(
                log => log.nodeId === event.nodeId && log.status === 'running'
              )
              if (!existingLog) {
                executionLogs.value.push({
                  nodeId: event.nodeId,
                  nodeType: event.nodeType,
                  nodeLabel: event.nodeLabel,
                  status: 'running',
                  startTime: Date.now(),
                  endTime: null,
                  duration: null,
                  input: null,
                  output: null,
                  error: null,
                  retry: 0
                })
              }
            }
            if (event.content) {
              testResult.value += event.content
            }
            if (event.showMsg) {
              ElMessage.info(event.showMsg)
            }
          } catch (e) {
            console.error('Parse SSE error:', e, 'data:', data)
          }
        }
      }
    }

    testRunning.value = false
    ElMessage.success('工作流执行完成')
  } catch (error) {
    console.error('运行工作流失败:', error)
    testError.value = error.message || '运行失败，请检查配置'
    testRunning.value = false
    ElMessage.error(error.message || '运行失败')
  }
}

// 结构校验工作流
const runStructValidation = async () => {
  validating.value = true
  try {
    const flowData = {
      nodes: nodes.value,
      edges: edges.value
    }
    const res = await validateWorkflow({
      topology: JSON.stringify(flowData)
    })
    if (res.code === 0 && res.data) {
      validationResult.value = res.data
      return res.data
    } else {
      ElMessage.error(res.msg || '校验失败')
      return null
    }
  } catch (error) {
    console.error('校验失败:', error)
    ElMessage.error('校验请求失败')
    return null
  } finally {
    validating.value = false
  }
}

// LLM 语义校验（需要先保存）
const runLLMValidation = async () => {
  if (!workflowId) return
  llmValidating.value = true
  try {
    const res = await validateWorkflowById(workflowId, { useLLM: true })
    if (res.code === 0 && res.data) {
      validationResult.value = res.data
      validationDialogVisible.value = true
    } else {
      ElMessage.error(res.msg || 'LLM 校验失败')
    }
  } catch (error) {
    console.error('LLM 校验失败:', error)
    ElMessage.error('LLM 校验请求失败')
  } finally {
    llmValidating.value = false
  }
}

// 保存前校验
const validateBeforeSave = async () => {
  validationForSave.value = true
  const result = await runStructValidation()
  if (!result) return true // 校验请求失败，允许保存

  if (!result.valid) {
    validationDialogVisible.value = true
    return false // 有错误，阻止保存
  }

  if (result.warnings && result.warnings.length > 0) {
    validationDialogVisible.value = true
    // 有警告但没有错误，允许保存
  }

  return true
}

// 强制保存（跳过校验提示）
const forceSaveWorkflow = async () => {
  validationDialogVisible.value = false
  await doSaveWorkflow()
}

// 定位到节点
const locateNode = (nodeId) => {
  if (!nodeId) return
  const node = nodes.value.find(n => n.id === nodeId)
  if (node) {
    selectedNode.value = node
    // 飞入视图到该节点
    fitView({ nodes: [nodeId], padding: 0.5, duration: 500 })
  }
}

// 全局键盘事件监听
const handleGlobalKeyDown = (e) => {
  if (e.key === 'Delete' || e.key === 'Backspace') {
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return
    deleteSelected()
  }
}

onMounted(() => {
  loadWorkflow()
  loadOptions()
  document.addEventListener('keydown', handleGlobalKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleGlobalKeyDown)
})
</script>

<style lang="scss" scoped>
.workflow-editor-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
  background: #f8f9fa;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 24px;
    background: white;
    border-bottom: 1px solid #e8e8e8;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.03);

    .title {
      display: flex;
      align-items: center;
      gap: 20px;

      .workflow-name-input {
        width: 280px;
        font-size: 16px;
        font-weight: 500;

        :deep(.el-input__wrapper) {
          background: #f5f5f5;
          border-radius: 10px;
          box-shadow: none;

          &:hover, &:focus-within {
            background: white;
            box-shadow: 0 0 0 1px #1976d2;
          }
        }
      }
    }

    .actions {
      .el-button {
        border-radius: 10px;
        padding: 10px 24px;
        font-weight: 500;
      }
    }
  }

  .editor-main {
    display: flex;
    flex: 1;
    overflow: hidden;

    .canvas-area {
      flex: 1;
      position: relative;
      background: linear-gradient(135deg, #fafbfc 0%, #f0f2f5 100%);
      outline: none;

      &:focus {
        outline: none;
      }
    }
  }
}

// Vue Flow 全局样式覆盖
:deep(.vue-flow) {
  .vue-flow__edge-path {
    stroke-width: 2;
  }

  .vue-flow__edge.animated path {
    stroke-dasharray: 5;
    animation: flowDash 0.5s linear infinite;
  }

  @keyframes flowDash {
    to {
      stroke-dashoffset: -10;
    }
  }

  .vue-flow__node {
    cursor: pointer;
    transition: box-shadow 0.2s ease, border-color 0.3s ease;

    &:hover {
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    }

    // 运行中 - 呼吸脉冲
    &.node-running {
      border: 2px solid #2196f3 !important;
      box-shadow: 0 0 20px rgba(33, 150, 243, 0.4);
      animation: nodePulse 1.5s ease-in-out infinite;
      z-index: 100;
    }

    // 已完成 - 绿色边框
    &.node-completed {
      border: 2px solid #4caf50 !important;
      box-shadow: 0 0 10px rgba(76, 175, 80, 0.3);
    }

    // 失败 - 红色边框 + 抖动
    &.node-failed {
      border: 2px solid #f44336 !important;
      animation: nodeShake 0.5s ease-in-out;
    }
  }

  .vue-flow__node.selected {
    outline: 2px solid #1976d2;
    outline-offset: 2px;
  }

  .vue-flow__edge.selected .vue-flow__edge-path {
    stroke: #1976d2;
    stroke-width: 3;
  }

  .vue-flow__edge.selected {
    z-index: 10;
  }

  // 活跃边 - 流动动画
  .vue-flow__edge.edge-active .vue-flow__edge-path {
    stroke: #2196f3;
    stroke-width: 3;
    stroke-dasharray: 10 5;
    animation: edgeFlow 1s linear infinite;
  }

  .vue-flow__handle {
    width: 12px;
    height: 12px;
    transition: all 0.15s ease;

    &:hover {
      transform: scale(1.3);
    }
  }

  .vue-flow__connection-path {
    stroke: #1976d2;
    stroke-width: 2;
  }

  .vue-flow__controls {
    border-radius: 10px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    border: 1px solid #e0e0e0;

    button {
      width: 28px;
      height: 28px;

      &:hover {
        background: #f0f0f0;
      }
    }
  }
}

// 测试对话框 - 缩小宽度，运行时自动关闭
:deep(.test-dialog) {
  .el-dialog {
    border-radius: 14px;
  }
}

// 运行结果底部面板
.test-result-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid #e0e0e0;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.08);
  z-index: 50;
  border-radius: 14px 14px 0 0;
  max-height: 260px;
  display: flex;
  flex-direction: column;

  .result-panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 20px;
    border-bottom: 1px solid #f0f0f0;
    flex-shrink: 0;

    .result-panel-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      font-weight: 500;
      color: #333;

      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #bdbdbd;

        &.running {
          background: #ef6c00;
          animation: dotPulse 1s ease-in-out infinite;
        }

        &.success {
          background: #4caf50;
        }

        &.error {
          background: #f44336;
        }
      }
    }

    .result-panel-actions {
      display: flex;
      align-items: center;
      gap: 4px;
    }
  }

  .result-panel-body {
    padding: 14px 20px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;

    .running-indicator {
      display: flex;
      align-items: center;
      gap: 10px;
      color: #ef6c00;
      font-size: 13px;

      .loading-icon {
        animation: rotate 1s linear infinite;
      }
    }

    .error-message {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #c62828;
      font-size: 13px;
    }

    .result-text {
      font-size: 13px;
      line-height: 1.6;
      white-space: pre-wrap;
      color: #424242;
    }
  }
}

// 底部面板滑入/滑出动画
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes nodePulse {
  0%, 100% {
    box-shadow: 0 0 10px rgba(33, 150, 243, 0.3);
  }
  50% {
    box-shadow: 0 0 25px rgba(33, 150, 243, 0.6), 0 0 40px rgba(33, 150, 243, 0.2);
  }
}

@keyframes dotPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

@keyframes nodeShake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-4px); }
  40% { transform: translateX(4px); }
  60% { transform: translateX(-3px); }
  80% { transform: translateX(3px); }
}

@keyframes edgeFlow {
  to {
    stroke-dashoffset: -15;
  }
}

// 校验结果对话框
:deep(.validation-dialog) {
  .el-dialog {
    border-radius: 14px;
  }

  .el-dialog__body {
    padding: 16px 20px;
  }
}

.validation-content {
  .validation-status {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    border-radius: 10px;
    margin-bottom: 16px;
    font-weight: 500;
    font-size: 14px;

    &.valid {
      background: #e8f5e9;
      color: #2e7d32;
    }

    &.invalid {
      background: #ffebee;
      color: #c62828;
    }
  }

  .validation-section {
    margin-bottom: 16px;

    .section-title {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;
      font-weight: 500;
      margin-bottom: 8px;

      &.error-title {
        color: #c62828;
      }

      &.warning-title {
        color: #ef6c00;
      }
    }

    .issue-list {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .issue-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      border-radius: 8px;
      font-size: 13px;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: #f5f5f5;
      }

      &.error-item {
        background: #fff3f3;
      }

      &.warning-item {
        background: #fff8e1;
      }

      .issue-badge {
        flex-shrink: 0;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 500;

        &.error {
          background: #ffcdd2;
          color: #c62828;
        }

        &.warning {
          background: #ffe0b2;
          color: #ef6c00;
        }
      }

      .issue-message {
        flex: 1;
        color: #333;
      }

      .locate-btn {
        flex-shrink: 0;
        padding: 0;
      }
    }
  }

  .validation-all-pass {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 24px;
    color: #2e7d32;

    p {
      font-size: 14px;
      margin: 0;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
