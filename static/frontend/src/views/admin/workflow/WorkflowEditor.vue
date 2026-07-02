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
        <el-button type="info" plain @click="openVersionPanel">
          <el-icon><Clock /></el-icon>
          版本管理
          <el-tag v-if="workflowData?.publishedVersion" size="small" type="success" effect="dark" style="margin-left:4px;">v{{ workflowData.publishedVersion }}</el-tag>
        </el-button>
        <el-button type="success" @click="openTestDialog">
          <el-icon><VideoPlay /></el-icon>
          运行测试
        </el-button>
        <el-button type="primary" @click="saveWorkflow" :loading="saving">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
        <el-dropdown @command="handleMoreAction">
          <el-button type="default" plain>
            更多
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="saveAsTemplate">
                <el-icon><FolderAdd /></el-icon>
                保存为模板
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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

        <!-- 版本管理面板 -->
        <VersionPanel
          :visible="versionPanelVisible"
          :versions="versionList"
          :current-version="workflowData?.currentVersion || 0"
          :loading="versionLoading"
          @close="versionPanelVisible = false"
          @create="handleCreateVersion"
          @publish="handlePublishVersion"
          @rollback="handleRollbackVersion"
          @preview="handlePreviewVersion"
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

    <!-- 保存为模板对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      title="保存为模板"
      width="480px"
      :close-on-click-modal="false"
      class="template-dialog"
    >
      <el-form label-position="top" :model="templateForm" :rules="templateRules" ref="templateFormRef">
        <el-form-item label="模板名称" prop="name">
          <el-input
            v-model="templateForm.name"
            placeholder="请输入模板名称"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="模板描述">
          <el-input
            v-model="templateForm.description"
            type="textarea"
            :rows="2"
            placeholder="简要描述模板的用途（可选）"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="模板分类">
          <el-select v-model="templateForm.category" placeholder="选择分类（可选）" clearable style="width: 100%;">
            <el-option label="通用" value="general" />
            <el-option label="问答" value="qa" />
            <el-option label="写作" value="writing" />
            <el-option label="数据分析" value="data" />
            <el-option label="工具调用" value="tool" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveAsTemplate" :loading="templateSaving">保存模板</el-button>
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
import { Check, VideoPlay, Loading, WarningFilled, Delete, ArrowUp, ArrowDown, CircleCheck, MagicStick, Document, Clock, FolderAdd } from '@element-plus/icons-vue'
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
import AgentNode from '@/components/workflow/AgentNode.vue'
import ParallelNode from '@/components/workflow/ParallelNode.vue'
import JoinNode from '@/components/workflow/JoinNode.vue'
import NodeSidebar from '@/components/workflow/NodeSidebar.vue'
import PropertyPanel from '@/components/workflow/PropertyPanel.vue'
import ExecutionLogPanel from '@/components/workflow/ExecutionLogPanel.vue'
import VersionPanel from '@/components/workflow/VersionPanel.vue'

import { defaultApi } from '@/api'
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
  maxRunSteps: 30,
  inputSchema: []
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

// 版本管理相关
const versionPanelVisible = ref(false)
const versionList = ref([])
const versionLoading = ref(false)
const workflowData = ref(null)

// 模板相关
const templateDialogVisible = ref(false)
const templateSaving = ref(false)
const templateFormRef = ref(null)
const templateForm = ref({ name: '', description: '', category: '' })
const templateRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }]
}

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
  agent: markRaw(AgentNode),
  condition: markRaw(ConditionNode),
  code: markRaw(CodeNode),
  http: markRaw(HTTPNode),
  parallel: markRaw(ParallelNode),
  join: markRaw(JoinNode)
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
    const res = await defaultApi.apiWorkflowIdGet(workflowId)
    if (res.code === 0 && res.data) {
      workflowData.value = res.data
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
              maxRunSteps: flowData.config.maxRunSteps || 30,
              inputSchema: flowData.config.inputSchema || []
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
      defaultApi.apiWorkflowModelsGet(),
      defaultApi.apiWorkflowToolsGet()
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
          contextEnabled: true,
          tools: [],
          maxToolRounds: 5
        }
      } : {}),
      ...(type === 'tool' ? {
        toolConfig: {
          toolName: '',
          params: {},
          tools: []
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
      } : {}),
      ...(type === 'agent' ? {
        agentConfig: {
          systemPrompt: '',
          tools: [],
          maxRunSteps: 30
        },
        modelConfig: {
          model: '',
          systemPrompt: '',
          temperature: 0.7,
          maxTokens: 4096,
          contextEnabled: true,
          tools: [],
          maxToolRounds: 5
        }
      } : {}),
      ...(type === 'parallel' ? {
        parallelConfig: {
          maxConcurrency: 0,
          waitStrategy: 'all',
          timeout: 0
        }
      } : {}),
      ...(type === 'join' ? {
        joinConfig: {
          strategy: 'all',
          timeout: 0
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
    agent: 'Agent',
    condition: '条件分支',
    code: '代码',
    http: 'HTTP',
    parallel: '并行组',
    join: '汇聚'
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
    const res = await defaultApi.apiWorkflowIdPut(workflowId, {
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

// ==================== 版本管理 ====================

// 打开版本面板
const openVersionPanel = async () => {
  versionPanelVisible.value = true
  await loadVersionList()
}

// 加载版本列表
const loadVersionList = async () => {
  versionLoading.value = true
  try {
    const res = await defaultApi.apiWorkflowIdVersionsGet(workflowId, 1, 50)
    const result = res.data || res
    versionList.value = result.records || []
  } catch (error) {
    console.error('加载版本列表失败:', error)
    ElMessage.error('加载版本列表失败')
  } finally {
    versionLoading.value = false
  }
}

// 创建版本
const handleCreateVersion = async (formData) => {
  try {
    await defaultApi.apiWorkflowIdVersionsPost(workflowId, formData)
    ElMessage.success('版本创建成功')
    await loadVersionList()
    // 重新加载工作流数据以更新 currentVersion
    await loadWorkflow()
  } catch (error) {
    console.error('创建版本失败:', error)
    ElMessage.error('创建版本失败')
  }
}

// 发布版本
const handlePublishVersion = async (ver) => {
  try {
    await defaultApi.apiWorkflowIdVersionsPublishPost(workflowId, { version: ver.version })
    ElMessage.success(`版本 v${ver.version} 发布成功`)
    await loadVersionList()
    await loadWorkflow()
  } catch (error) {
    console.error('发布版本失败:', error)
    ElMessage.error('发布版本失败')
  }
}

// 回滚版本
const handleRollbackVersion = async (ver) => {
  try {
    await defaultApi.apiWorkflowIdVersionsVersionRollbackPost(workflowId, ver.version)
    ElMessage.success(`已回滚到版本 v${ver.version}`)
    // 重新加载工作流数据
    await loadWorkflow()
    await loadVersionList()
  } catch (error) {
    console.error('回滚版本失败:', error)
    ElMessage.error('回滚版本失败')
  }
}

// 预览版本（加载版本的拓扑到画布）
const handlePreviewVersion = async (ver) => {
  try {
    const res = await defaultApi.apiWorkflowIdVersionsVersionGet(workflowId, ver.version)
    const versionData = res.data || res
    if (versionData.topology) {
      const flowData = JSON.parse(versionData.topology)
      nodes.value = flowData.nodes || []
      edges.value = (flowData.edges || []).map(edge => ({
        ...edge,
        sourceHandle: migrateHandleId(edge.sourceHandle, 'source'),
        targetHandle: migrateHandleId(edge.targetHandle, 'target')
      }))
      ElMessage.info(`已加载版本 v${ver.version} 的拓扑（只读预览，保存后生效）`)
    }
  } catch (error) {
    console.error('加载版本详情失败:', error)
    ElMessage.error('加载版本详情失败')
  }
}

// ==================== 模板管理 ====================

// 更多操作下拉菜单
const handleMoreAction = (command) => {
  if (command === 'saveAsTemplate') {
    openTemplateDialog()
  }
}

// 打开保存为模板对话框
const openTemplateDialog = () => {
  templateForm.value = {
    name: workflowName.value ? workflowName.value + ' 模板' : '',
    description: '',
    category: ''
  }
  templateDialogVisible.value = true
}

// 保存为模板
const handleSaveAsTemplate = async () => {
  if (!templateFormRef.value) return
  await templateFormRef.value.validate(async (valid) => {
    if (!valid) return
    templateSaving.value = true
    try {
      const res = await defaultApi.apiWorkflowTemplatesPost({
        flowId: parseInt(workflowId),
        name: templateForm.value.name,
        description: templateForm.value.description,
        category: templateForm.value.category
      })
      if (res.code === 0) {
        ElMessage.success('模板保存成功')
        templateDialogVisible.value = false
      } else {
        ElMessage.error(res.msg || '保存模板失败')
      }
    } catch (error) {
      console.error('保存模板失败:', error)
      ElMessage.error('保存模板失败')
    } finally {
      templateSaving.value = false
    }
  })
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
    const res = await defaultApi.apiWorkflowValidatePost({
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
    const res = await defaultApi.apiWorkflowIdValidatePost(workflowId, { useLLM: true })
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
// 设计变量 / Design Variables (浅色主题)
$primary-color: #3b82f6;
$primary-light: #60a5fa;
$primary-dark: #2563eb;
$success-color: #10b981;
$warning-color: #f59e0b;
$danger-color: #ef4444;
$bg-white: #ffffff;
$bg-light: #f8fafc;
$bg-card: #f1f5f9;
$border-color: #e2e8f0;
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.workflow-editor-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
  background: linear-gradient(180deg, $bg-light 0%, $bg-card 100%);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 
    0 4px 20px rgba(0, 0, 0, 0.06),
    0 1px 3px rgba(0, 0, 0, 0.04);

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 24px;
    background: $bg-white;
    border-bottom: 1px solid $border-color;
    position: relative;

    // 顶部装饰线
    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 3px;
      background: linear-gradient(90deg, 
        $primary-color 0%, 
        $primary-light 50%, 
        $primary-color 100%
      );
    }

    .title {
      display: flex;
      align-items: center;
      gap: 20px;

      .workflow-name-input {
        width: 280px;
        font-size: 16px;
        font-weight: 600;

        :deep(.el-input__wrapper) {
          background: $bg-card;
          border-radius: 10px;
          box-shadow: none;
          border: 1px solid $border-color;
          transition: all 0.3s ease;

          &:hover, &:focus-within {
            background: $bg-white;
            border-color: $primary-color;
            box-shadow: 0 0 0 3px rgba($primary-color, 0.1);
          }

          .el-input__inner {
            color: $text-primary;
            font-weight: 500;
            
            &::placeholder {
              color: $text-muted;
            }
          }
        }
      }
    }

    .actions {
      display: flex;
      align-items: center;
      gap: 10px;

      .el-button {
        border-radius: 10px;
        padding: 10px 20px;
        font-weight: 500;
        transition: all 0.2s ease;
        border: 1px solid $border-color;
        background: $bg-white;
        color: $text-secondary;

        &:hover {
          transform: translateY(-1px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
        }

        // 危险按钮
        &.is-danger {
          background: rgba($danger-color, 0.08);
          border-color: rgba($danger-color, 0.2);
          color: $danger-color;
          
          &:hover {
            background: rgba($danger-color, 0.12);
            border-color: rgba($danger-color, 0.3);
          }
        }

        // 警告按钮
        &.is-warning {
          background: rgba($warning-color, 0.08);
          border-color: rgba($warning-color, 0.2);
          color: darken($warning-color, 10%);
          
          &:hover {
            background: rgba($warning-color, 0.12);
            border-color: rgba($warning-color, 0.3);
          }
        }

        // 信息按钮
        &.is-info {
          background: $bg-card;
          border-color: $border-color;
          color: $text-secondary;
          
          &:hover {
            background: darken($bg-card, 2%);
            border-color: darken($border-color, 5%);
          }
        }

        // 成功按钮
        &.is-success {
          background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
          border-color: $primary-color;
          color: white;
          box-shadow: 
            0 2px 8px rgba($primary-color, 0.25),
            inset 0 1px 0 rgba(255, 255, 255, 0.15);
          
          &:hover {
            background: linear-gradient(135deg, $primary-light 0%, $primary-color 100%);
            box-shadow: 
              0 4px 16px rgba($primary-color, 0.35),
              inset 0 1px 0 rgba(255, 255, 255, 0.2);
          }
        }

        // 主要按钮
        &.el-button--primary {
          background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
          border-color: $primary-color;
          color: white;
          box-shadow: 
            0 2px 8px rgba($primary-color, 0.25),
            inset 0 1px 0 rgba(255, 255, 255, 0.15);
          
          &:hover {
            background: linear-gradient(135deg, $primary-light 0%, $primary-color 100%);
            box-shadow: 
              0 4px 16px rgba($primary-color, 0.35),
              inset 0 1px 0 rgba(255, 255, 255, 0.2);
          }
        }

        .el-tag {
          border-radius: 6px;
          font-weight: 600;
        }
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
      background: 
        radial-gradient(ellipse at center, rgba($primary-color, 0.02) 0%, transparent 70%),
        $bg-light;
      outline: none;

      // 网格纹理
      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-image: 
          linear-gradient(rgba($primary-color, 0.04) 1px, transparent 1px),
          linear-gradient(90deg, rgba($primary-color, 0.04) 1px, transparent 1px);
        background-size: 20px 20px;
        pointer-events: none;
      }

      &:focus {
        outline: none;
      }
    }
  }
}

// Vue Flow 全局样式覆盖
:deep(.vue-flow) {
  background: transparent;
  
  .vue-flow__edge-path {
    stroke-width: 2;
    filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.1));
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
    transition: all 0.2s ease;

    &:hover {
      filter: brightness(0.98);
      transform: translateY(-1px);
    }

    // 运行中 - 呼吸脉冲
    &.node-running {
      border: 2px solid #3b82f6 !important;
      box-shadow: 
        0 0 0 3px rgba(59, 130, 246, 0.15),
        0 4px 12px rgba(59, 130, 246, 0.2);
      animation: nodePulse 1.5s ease-in-out infinite;
      z-index: 100;
    }

    // 已完成 - 绿色边框
    &.node-completed {
      border: 2px solid $success-color !important;
      box-shadow: 0 0 0 3px rgba($success-color, 0.1);
    }

    // 失败 - 红色边框
    &.node-failed {
      border: 2px solid $danger-color !important;
      box-shadow: 0 0 0 3px rgba($danger-color, 0.1);
      animation: nodeShake 0.5s ease-in-out;
    }
  }

  .vue-flow__node.selected {
    outline: 2px solid $primary-color;
    outline-offset: 2px;
  }

  .vue-flow__edge.selected .vue-flow__edge-path {
    stroke: $primary-color;
    stroke-width: 3;
  }

  .vue-flow__edge.selected {
    z-index: 10;
  }

  // 活跃边
  .vue-flow__edge.edge-active .vue-flow__edge-path {
    stroke: #3b82f6;
    stroke-width: 3;
    stroke-dasharray: 10 5;
    animation: edgeFlow 1s linear infinite;
  }

  .vue-flow__handle {
    width: 12px;
    height: 12px;
    transition: all 0.2s ease;

    &:hover {
      transform: scale(1.3);
    }
  }

  .vue-flow__connection-path {
    stroke: $primary-color;
    stroke-width: 2;
  }

  .vue-flow__controls {
    border-radius: 12px;
    background: $bg-white;
    border: 1px solid $border-color;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);

    button {
      width: 30px;
      height: 30px;
      background: transparent;
      border: none;
      color: $text-secondary;
      transition: all 0.2s ease;

      &:hover {
        background: rgba($primary-color, 0.1);
        color: $primary-color;
      }

      svg {
        fill: currentColor;
      }
    }
  }

  .vue-flow__minimap {
    background: $bg-white;
    border-radius: 12px;
    border: 1px solid $border-color;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
}

@keyframes nodePulse {
  0%, 100% {
    box-shadow: 
      0 0 0 3px rgba(59, 130, 246, 0.15),
      0 4px 12px rgba(59, 130, 246, 0.2);
  }
  50% {
    box-shadow: 
      0 0 0 6px rgba(59, 130, 246, 0.1),
      0 4px 20px rgba(59, 130, 246, 0.3);
  }
}

@keyframes nodeShake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-3px); }
  40% { transform: translateX(3px); }
  60% { transform: translateX(-2px); }
  80% { transform: translateX(2px); }
}

@keyframes edgeFlow {
  to {
    stroke-dashoffset: -15;
  }
}

// 测试对话框
:deep(.test-dialog) {
  .el-dialog {
    border-radius: 16px;
    background: $bg-white;
    border: 1px solid $border-color;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
    
    .el-dialog__header {
      border-bottom: 1px solid $border-color;
      padding: 20px 24px;
      
      .el-dialog__title {
        color: $text-primary;
        font-weight: 600;
      }
    }
    
    .el-dialog__body {
      padding: 24px;
      
      .el-textarea__inner {
        background: $bg-card;
        border-color: $border-color;
        color: $text-primary;
        border-radius: 10px;
        
        &:focus {
          border-color: $primary-color;
          box-shadow: 0 0 0 3px rgba($primary-color, 0.1);
        }
      }
    }
    
    .el-dialog__footer {
      border-top: 1px solid $border-color;
      padding: 16px 24px;
    }
  }
}

// 运行结果底部面板
.test-result-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: $bg-white;
  border: 1px solid $border-color;
  border-top: 3px solid $primary-color;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
  z-index: 50;
  max-height: 280px;
  display: flex;
  flex-direction: column;
  border-radius: 16px 16px 0 0;

  .result-panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 20px;
    border-bottom: 1px solid $border-color;
    flex-shrink: 0;

    .result-panel-title {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 14px;
      font-weight: 500;
      color: $text-primary;

      .status-dot {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        background: $text-muted;

        &.running {
          background: $warning-color;
          animation: dotPulse 1s ease-in-out infinite;
        }

        &.success {
          background: $success-color;
        }

        &.error {
          background: $danger-color;
        }
      }
    }

    .result-panel-actions {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .el-button {
        border-radius: 8px;
        background: $bg-card;
        border-color: $border-color;
        color: $text-secondary;
        
        &:hover {
          background: rgba($primary-color, 0.1);
          border-color: rgba($primary-color, 0.2);
          color: $primary-color;
        }
      }
    }
  }

  .result-panel-body {
    padding: 16px 20px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;

    .running-indicator {
      display: flex;
      align-items: center;
      gap: 12px;
      color: $warning-color;
      font-size: 13px;

      .loading-icon {
        animation: rotate 1s linear infinite;
        font-size: 16px;
      }
    }

    .error-message {
      display: flex;
      align-items: center;
      gap: 10px;
      color: $danger-color;
      font-size: 13px;
      padding: 12px 16px;
      background: rgba($danger-color, 0.05);
      border-radius: 10px;
      border: 1px solid rgba($danger-color, 0.1);
    }

    .result-text {
      font-size: 13px;
      line-height: 1.7;
      white-space: pre-wrap;
      color: $text-secondary;
      padding: 12px 16px;
      background: $bg-card;
      border-radius: 10px;
      border: 1px solid $border-color;
    }
  }
}

// 底部面板滑入/滑出动画
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
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

@keyframes dotPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

// 校验结果对话框
:deep(.validation-dialog) {
  .el-dialog {
    border-radius: 16px;
    background: $bg-white;
    border: 1px solid $border-color;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
    
    .el-dialog__header {
      border-bottom: 1px solid $border-color;
      padding: 20px 24px;
      
      .el-dialog__title {
        color: $text-primary;
        font-weight: 600;
      }
    }
    
    .el-dialog__body {
      padding: 24px;
    }
    
    .el-dialog__footer {
      border-top: 1px solid $border-color;
      padding: 16px 24px;
    }
  }
}

:deep(.template-dialog) {
  .el-dialog {
    border-radius: 16px;
    background: $bg-white;
    border: 1px solid $border-color;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
    
    .el-dialog__header {
      border-bottom: 1px solid $border-color;
      padding: 20px 24px;
      
      .el-dialog__title {
        color: $text-primary;
        font-weight: 600;
      }
    }
    
    .el-dialog__body {
      padding: 24px;
      
      .el-input__wrapper,
      .el-textarea__inner {
        background: $bg-card;
        border-color: $border-color;
        color: $text-primary;
        border-radius: 10px;
        
        &:focus {
          border-color: $primary-color;
          box-shadow: 0 0 0 3px rgba($primary-color, 0.1);
        }
      }
      
      .el-select {
        .el-input__wrapper {
          border-radius: 10px;
        }
      }
    }
    
    .el-dialog__footer {
      border-top: 1px solid $border-color;
      padding: 16px 24px;
    }
  }
}

.validation-content {
  .validation-status {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 20px;
    border-radius: 12px;
    margin-bottom: 20px;
    font-weight: 600;
    font-size: 15px;

    &.valid {
      background: rgba($success-color, 0.08);
      color: darken($success-color, 5%);
      border: 1px solid rgba($success-color, 0.2);
    }

    &.invalid {
      background: rgba($danger-color, 0.08);
      color: $danger-color;
      border: 1px solid rgba($danger-color, 0.2);
    }
  }

  .validation-section {
    margin-bottom: 20px;

    .section-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      font-weight: 600;
      margin-bottom: 12px;

      &.error-title {
        color: $danger-color;
      }

      &.warning-title {
        color: darken($warning-color, 5%);
      }
    }

    .issue-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .issue-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 16px;
      border-radius: 10px;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.2s ease;

      &:hover {
        transform: translateX(4px);
      }

      &.error-item {
        background: rgba($danger-color, 0.05);
        border: 1px solid rgba($danger-color, 0.15);
      }

      &.warning-item {
        background: rgba($warning-color, 0.05);
        border: 1px solid rgba($warning-color, 0.15);
      }

      .issue-badge {
        flex-shrink: 0;
        padding: 4px 10px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 600;

        &.error {
          background: rgba($danger-color, 0.1);
          color: $danger-color;
        }

        &.warning {
          background: rgba($warning-color, 0.1);
          color: darken($warning-color, 5%);
        }
      }

      .issue-message {
        flex: 1;
        color: $text-primary;
        line-height: 1.5;
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
    gap: 16px;
    padding: 32px;
    color: darken($success-color, 5%);

    p {
      font-size: 15px;
      font-weight: 500;
      margin: 0;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  
  .el-button {
    border-radius: 10px;
    padding: 10px 20px;
    
    &.is-primary {
      background: linear-gradient(135deg, $primary-color, $primary-dark);
      border-color: $primary-color;
      
      &:hover {
        background: linear-gradient(135deg, $primary-light, $primary-color);
      }
    }
  }
}
</style>
