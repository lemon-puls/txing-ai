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
      <NodeSidebar :saving="saving" @save="saveWorkflow" />

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
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form label-position="top">
        <el-form-item label="输入内容">
          <el-input
            v-model="testInput"
            type="textarea"
            :rows="4"
            placeholder="请输入要测试的内容，例如：请介绍一下人工智能的发展历史"
          />
        </el-form-item>
      </el-form>

      <!-- 运行结果显示区域 -->
      <div class="test-result-area" v-if="testRunning || testResult">
        <div class="result-header">
          <span class="status-badge" :class="{ running: testRunning, success: !testRunning && !testError }">
            {{ testRunning ? '运行中...' : (testError ? '失败' : '完成') }}
          </span>
        </div>
        <div class="result-content">
          <div v-if="testRunning" class="running-indicator">
            <el-icon class="loading-icon"><Loading /></el-icon>
            <span>正在执行工作流...</span>
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

      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="runWorkflowTest" :loading="testRunning" :disabled="!testInput">
          开始运行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { Check, VideoPlay, Loading, WarningFilled, Delete } from '@element-plus/icons-vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'

// 导入自定义节点组件
import StartNode from '@/components/workflow/StartNode.vue'
import EndNode from '@/components/workflow/EndNode.vue'
import LLMNode from '@/components/workflow/LLMNode.vue'
import ToolNode from '@/components/workflow/ToolNode.vue'
import ConditionNode from '@/components/workflow/ConditionNode.vue'
import NodeSidebar from '@/components/workflow/NodeSidebar.vue'
import PropertyPanel from '@/components/workflow/PropertyPanel.vue'

import { getWorkflow, updateWorkflow, getWorkflowModels, getWorkflowTools, runWorkflow } from '@/api/workflow'
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

// 测试相关
const testDialogVisible = ref(false)
const testInput = ref('')
const testRunning = ref(false)
const testResult = ref('')
const testError = ref('')

// 注册自定义节点类型
const nodeTypes = {
  start: markRaw(StartNode),
  end: markRaw(EndNode),
  llm: markRaw(LLMNode),
  tool: markRaw(ToolNode),
  condition: markRaw(ConditionNode)
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
    condition: '条件分支'
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

// 保存工作流
const saveWorkflow = async () => {
  if (!workflowId) return
  saving.value = true
  try {
    const flowData = {
      nodes: nodes.value,
      edges: edges.value
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

// 打开测试对话框
const openTestDialog = () => {
  testInput.value = ''
  testResult.value = ''
  testError.value = ''
  testDialogVisible.value = true
}

// 运行工作流测试
const runWorkflowTest = async () => {
  if (!testInput.value || !workflowId) return

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
              ElMessage.success('工作流执行完成')
              return
            }
            if (event.error) {
              testError.value = event.error
              testRunning.value = false
              return
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
    transition: box-shadow 0.2s ease;

    &:hover {
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
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

// 测试对话框样式
.test-result-area {
  margin-top: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 12px;
  background: #fafafa;

  .result-header {
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;

    .status-badge {
      display: inline-flex;
      align-items: center;
      padding: 4px 12px;
      border-radius: 8px;
      font-size: 13px;
      font-weight: 500;

      &.running {
        background: #fff3e0;
        color: #ef6c00;
      }

      &.success {
        background: #e8f5e9;
        color: #2e7d32;
      }
    }
  }

  .result-content {
    padding: 16px;
    min-height: 100px;
    max-height: 300px;
    overflow-y: auto;

    .running-indicator {
      display: flex;
      align-items: center;
      gap: 10px;
      color: #ef6c00;

      .loading-icon {
        animation: rotate 1s linear infinite;
      }
    }

    .error-message {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #c62828;
      font-size: 14px;
    }

    .result-text {
      font-size: 14px;
      line-height: 1.6;
      white-space: pre-wrap;
      color: #424242;
    }
  }
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
