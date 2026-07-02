<template>
  <div class="node-sidebar">
    <div class="sidebar-header">
      <h3>节点组件</h3>
    </div>

    <!-- 工作流配置 -->
    <div class="workflow-config">
      <div class="config-title">工作流配置</div>
      <div class="config-item">
        <label>默认模型</label>
        <el-select
          v-model="defaultModel"
          size="small"
          placeholder="选择默认模型"
          clearable
          @change="onModelChange"
        >
          <el-option
            v-for="model in modelList"
            :key="model.name"
            :label="model.displayName || model.name"
            :value="model.name"
          />
        </el-select>
      </div>
      <div class="config-item">
        <label>最大步数 <span class="hint">（可选，默认30）</span></label>
        <el-input-number
          v-model="maxRunSteps"
          size="small"
          :min="1"
          :max="1000"
          placeholder="留空使用默认值"
          controls-position="right"
          @change="onMaxRunStepsChange"
        />
      </div>
      <div class="config-item">
        <label>输入字段配置</label>
        <el-button size="small" @click="inputSchemaDialogVisible = true" style="width: 100%;">
          <el-icon><Setting /></el-icon>
          配置输入字段 ({{ inputSchema.length }})
        </el-button>
      </div>
    </div>

    <!-- 输入字段配置弹窗 -->
    <el-dialog
      v-model="inputSchemaDialogVisible"
      title="配置输入字段"
      width="600px"
      :close-on-click-modal="false"
      append-to-body
    >
      <div class="input-schema-editor">
        <div v-if="inputSchema.length === 0" class="empty-schema">
          <p>暂未配置输入字段</p>
          <p class="hint">配置后，执行页面将显示对应的输入表单</p>
        </div>
        <div v-for="(field, index) in inputSchema" :key="index" class="schema-field-item">
          <div class="field-header">
            <span class="field-index">#{{ index + 1 }}</span>
            <el-button text type="danger" size="small" @click="removeSchemaField(index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <el-form label-position="top" size="small">
            <el-form-item label="字段标识">
              <el-input v-model="field.name" placeholder="英文标识，如：resume、destination" />
            </el-form-item>
            <el-form-item label="显示标签">
              <el-input v-model="field.label" placeholder="如：上传简历、目的地" />
            </el-form-item>
            <el-form-item label="字段类型">
              <el-select v-model="field.type" style="width: 100%;">
                <el-option label="文件上传" value="file" />
                <el-option label="多行文本" value="textarea" />
                <el-option label="单行文本" value="text" />
              </el-select>
            </el-form-item>
            <el-form-item label="占位提示">
              <el-input v-model="field.placeholder" placeholder="输入框中的提示文字" />
            </el-form-item>
            <el-form-item label="是否必填">
              <el-switch v-model="field.required" />
            </el-form-item>
          </el-form>
          <el-divider v-if="index < inputSchema.length - 1" />
        </div>
      </div>
      <template #footer>
        <el-button @click="addSchemaField">
          <el-icon><Plus /></el-icon>
          添加字段
        </el-button>
        <el-button type="primary" @click="saveInputSchema">确定</el-button>
      </template>
    </el-dialog>

    <div class="node-categories">
      <!-- 基础节点 -->
      <div class="category">
        <div class="category-title">基础</div>
        <div class="category-items">
          <div
            class="node-item start"
            draggable="true"
            @dragstart="onDragStart($event, 'start')"
          >
            <div class="item-icon">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <span>开始节点</span>
          </div>
          <div
            class="node-item end"
            draggable="true"
            @dragstart="onDragStart($event, 'end')"
          >
            <div class="item-icon">
              <el-icon><CircleClose /></el-icon>
            </div>
            <span>结束节点</span>
          </div>
        </div>
      </div>

      <!-- AI节点 -->
      <div class="category">
        <div class="category-title">AI能力</div>
        <div class="category-items">
          <div
            class="node-item llm"
            draggable="true"
            @dragstart="onDragStart($event, 'llm')"
          >
            <div class="item-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <span>大模型节点</span>
          </div>
          <div
            class="node-item tool"
            draggable="true"
            @dragstart="onDragStart($event, 'tool')"
          >
            <div class="item-icon">
              <el-icon><Tools /></el-icon>
            </div>
            <span>工具节点</span>
          </div>
          <div
            class="node-item agent"
            draggable="true"
            @dragstart="onDragStart($event, 'agent')"
          >
            <div class="item-icon">
              <el-icon><Avatar /></el-icon>
            </div>
            <span>Agent 节点</span>
          </div>
        </div>
      </div>

      <!-- 逻辑节点 -->
      <div class="category">
        <div class="category-title">逻辑控制</div>
        <div class="category-items">
          <div
            class="node-item condition"
            draggable="true"
            @dragstart="onDragStart($event, 'condition')"
          >
            <div class="item-icon">
              <el-icon><Share /></el-icon>
            </div>
            <span>条件分支</span>
          </div>
          <div
            class="node-item parallel"
            draggable="true"
            @dragstart="onDragStart($event, 'parallel')"
          >
            <div class="item-icon">
              <el-icon><Grid /></el-icon>
            </div>
            <span>并行组</span>
          </div>
          <div
            class="node-item join"
            draggable="true"
            @dragstart="onDragStart($event, 'join')"
          >
            <div class="item-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <span>汇聚</span>
          </div>
        </div>
      </div>

      <!-- 集成节点 -->
      <div class="category">
        <div class="category-title">集成</div>
        <div class="category-items">
          <div
            class="node-item code"
            draggable="true"
            @dragstart="onDragStart($event, 'code')"
          >
            <div class="item-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <span>代码节点</span>
          </div>
          <div
            class="node-item http"
            draggable="true"
            @dragstart="onDragStart($event, 'http')"
          >
            <div class="item-icon">
              <el-icon><Link /></el-icon>
            </div>
            <span>HTTP 节点</span>
          </div>
        </div>
      </div>
    </div>

    <div class="sidebar-footer">
      <el-button type="primary" size="small" @click="$emit('save')" :loading="saving">
        <el-icon><Check /></el-icon>
        保存流程
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { VideoPlay, CircleClose, ChatDotRound, Tools, Share, Check, Monitor, Link, Avatar, Setting, Delete, Plus, Grid, Connection } from '@element-plus/icons-vue'

const props = defineProps({
  saving: Boolean,
  modelList: {
    type: Array,
    default: () => []
  },
  workflowConfig: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['save', 'config-change'])

const defaultModel = ref(props.workflowConfig?.defaultModel || '')
const maxRunSteps = ref(props.workflowConfig?.maxRunSteps || null)
const inputSchema = ref(props.workflowConfig?.inputSchema ? JSON.parse(JSON.stringify(props.workflowConfig.inputSchema)) : [])
const inputSchemaDialogVisible = ref(false)

watch(() => props.workflowConfig, (newConfig) => {
  if (newConfig) {
    defaultModel.value = newConfig.defaultModel || ''
    maxRunSteps.value = newConfig.maxRunSteps || null
    inputSchema.value = newConfig.inputSchema ? JSON.parse(JSON.stringify(newConfig.inputSchema)) : []
  }
}, { deep: true })

const onModelChange = (value) => {
  emit('config-change', {
    ...props.workflowConfig,
    defaultModel: value || ''
  })
}

const onMaxRunStepsChange = (value) => {
  emit('config-change', {
    ...props.workflowConfig,
    maxRunSteps: value || null
  })
}

const addSchemaField = () => {
  inputSchema.value.push({
    name: '',
    type: 'text',
    label: '',
    placeholder: '',
    required: false,
    accept: '',
    description: ''
  })
}

const removeSchemaField = (index) => {
  inputSchema.value.splice(index, 1)
}

const saveInputSchema = () => {
  const validFields = inputSchema.value.filter(f => f.name && f.label)
  emit('config-change', {
    ...props.workflowConfig,
    inputSchema: validFields
  })
  inputSchemaDialogVisible.value = false
}

const onDragStart = (event, nodeType) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/vueflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }
}
</script>

<style lang="scss" scoped>
// 设计变量 / Design Variables (浅色主题)
$primary-color: #3b82f6;
$primary-light: #60a5fa;
$primary-dark: #2563eb;
$success-color: #10b981;
$warning-color: #f59e0b;
$bg-white: #ffffff;
$bg-light: #f8fafc;
$bg-card: #f1f5f9;
$border-color: #e2e8f0;
$text-primary: #1e293b;
$text-secondary: #64748b;
$text-muted: #94a3b8;

.node-sidebar {
  width: 260px;
  height: 100%;
  background: $bg-white;
  border-right: 1px solid $border-color;
  display: flex;
  flex-direction: column;

  .sidebar-header {
    padding: 20px 20px;
    border-bottom: 1px solid $border-color;

    h3 {
      margin: 0;
      font-size: 15px;
      font-weight: 600;
      color: $text-primary;
      display: flex;
      align-items: center;
      gap: 10px;
      
      &::before {
        content: '';
        width: 4px;
        height: 16px;
        background: linear-gradient(180deg, $primary-color, $primary-light);
        border-radius: 2px;
      }
    }
  }

  .workflow-config {
    padding: 16px;
    border-bottom: 1px solid $border-color;
    background: rgba($primary-color, 0.03);

    .config-title {
      font-size: 11px;
      font-weight: 600;
      color: $primary-color;
      margin-bottom: 12px;
      text-transform: uppercase;
      letter-spacing: 1px;
    }

    .config-item {
      margin-bottom: 12px;

      &:last-child {
        margin-bottom: 0;
      }

      label {
        display: block;
        font-size: 12px;
        color: $text-secondary;
        margin-bottom: 6px;
        font-weight: 500;

        .hint {
          color: $text-muted;
          font-size: 11px;
          font-weight: 400;
        }
      }

      .el-select {
        width: 100%;
        
        :deep(.el-input__wrapper) {
          background: $bg-white;
          border: 1px solid $border-color;
          border-radius: 8px;
          box-shadow: none;
          
          &:hover, &.is-focus {
            border-color: $primary-color;
          }
        }
      }

      .el-input-number {
        width: 100%;
        
        :deep(.el-input__wrapper) {
          background: $bg-white;
          border: 1px solid $border-color;
          border-radius: 8px;
        }
      }
      
      .el-button {
        width: 100%;
        justify-content: flex-start;
        background: $bg-white;
        border: 1px solid $border-color;
        color: $text-secondary;
        border-radius: 8px;
        
        &:hover {
          background: rgba($primary-color, 0.05);
          border-color: rgba($primary-color, 0.3);
          color: $primary-color;
        }
      }
    }
  }

  .node-categories {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    
    &::-webkit-scrollbar {
      width: 6px;
    }
    
    &::-webkit-scrollbar-track {
      background: transparent;
    }
    
    &::-webkit-scrollbar-thumb {
      background: $border-color;
      border-radius: 3px;
      
      &:hover {
        background: darken($border-color, 10%);
      }
    }

    .category {
      margin-bottom: 20px;

      .category-title {
        font-size: 11px;
        font-weight: 600;
        color: $text-muted;
        margin-bottom: 10px;
        padding-left: 8px;
        text-transform: uppercase;
        letter-spacing: 1px;
      }

      .category-items {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .node-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 14px;
        background: $bg-white;
        border: 1px solid $border-color;
        border-radius: 12px;
        cursor: grab;
        transition: all 0.2s ease;

        &:hover {
          border-color: $primary-color;
          box-shadow: 0 4px 12px rgba($primary-color, 0.1);
          transform: translateY(-1px);
        }

        &:active {
          cursor: grabbing;
          transform: scale(0.98);
        }

        .item-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 36px;
          height: 36px;
          border-radius: 10px;
          font-size: 18px;
          transition: all 0.2s ease;
        }

        span {
          font-size: 13px;
          font-weight: 500;
          color: $text-primary;
        }

        // 不同类型节点的颜色
        &.start {
          border-left: 3px solid $success-color;
          .item-icon { background: rgba($success-color, 0.1); color: $success-color; }
          &:hover { border-color: rgba($success-color, 0.5); }
        }
        &.end {
          border-left: 3px solid #ef4444;
          .item-icon { background: rgba(#ef4444, 0.1); color: #ef4444; }
          &:hover { border-color: rgba(#ef4444, 0.5); }
        }
        &.llm {
          border-left: 3px solid #3b82f6;
          .item-icon { background: rgba(#3b82f6, 0.1); color: #3b82f6; }
          &:hover { border-color: rgba(#3b82f6, 0.5); }
        }
        &.tool {
          border-left: 3px solid $warning-color;
          .item-icon { background: rgba($warning-color, 0.1); color: $warning-color; }
          &:hover { border-color: rgba($warning-color, 0.5); }
        }
        &.condition {
          border-left: 3px solid #7c3aed;
          .item-icon { background: rgba(#7c3aed, 0.1); color: #7c3aed; }
          &:hover { border-color: rgba(#7c3aed, 0.5); }
        }
        &.code {
          border-left: 3px solid #64748b;
          .item-icon { background: rgba(#64748b, 0.1); color: #64748b; }
          &:hover { border-color: rgba(#64748b, 0.5); }
        }
        &.http {
          border-left: 3px solid #06b6d4;
          .item-icon { background: rgba(#06b6d4, 0.1); color: #06b6d4; }
          &:hover { border-color: rgba(#06b6d4, 0.5); }
        }
        &.agent {
          border-left: 3px solid #ec4899;
          .item-icon { background: rgba(#ec4899, 0.1); color: #ec4899; }
          &:hover { border-color: rgba(#ec4899, 0.5); }
        }
        &.parallel {
          border-left: 3px solid #8b5cf6;
          .item-icon { background: rgba(#8b5cf6, 0.1); color: #8b5cf6; }
          &:hover { border-color: rgba(#8b5cf6, 0.5); }
        }
        &.join {
          border-left: 3px solid #14b8a6;
          .item-icon { background: rgba(#14b8a6, 0.1); color: #14b8a6; }
          &:hover { border-color: rgba(#14b8a6, 0.5); }
        }
      }
    }
  }

  .sidebar-footer {
    padding: 16px;
    border-top: 1px solid $border-color;

    .el-button {
      width: 100%;
      border-radius: 12px;
      justify-content: center;
      font-weight: 600;
      background: linear-gradient(135deg, $primary-color, $primary-dark);
      border: none;
      color: white;
      
      &:hover {
        background: linear-gradient(135deg, $primary-light, $primary-color);
      }
    }
  }
}

.input-schema-editor {
  max-height: 500px;
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 6px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: $border-color;
    border-radius: 3px;
  }

  .empty-schema {
    text-align: center;
    padding: 32px;
    color: $text-muted;

    p { 
      margin: 0 0 10px;
      font-size: 14px;
    }
    .hint { 
      font-size: 12px; 
      color: $text-muted;
    }
  }

  .schema-field-item {
    background: $bg-card;
    border: 1px solid $border-color;
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 12px;

    &:last-child {
      margin-bottom: 0;
    }

    .field-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 14px;

      .field-index {
        font-size: 14px;
        font-weight: 600;
        color: $primary-color;
        background: rgba($primary-color, 0.1);
        padding: 4px 10px;
        border-radius: 6px;
      }
      
      .el-button {
        color: #ef4444;
      }
    }
    
    :deep(.el-form-item) {
      margin-bottom: 12px;
      
      &:last-child {
        margin-bottom: 0;
      }
      
      .el-form-item__label {
        color: $text-secondary;
        font-weight: 500;
        font-size: 12px;
      }
      
      .el-input__wrapper,
      .el-textarea__inner {
        background: $bg-white;
        border: 1px solid $border-color;
        border-radius: 8px;
        
        &:hover, &:focus {
          border-color: $primary-color;
        }
      }
    }
  }
}
</style>
