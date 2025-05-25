<template>
  <div class="channel-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="渠道名称">
          <el-input v-model="searchForm.name" placeholder="请输入渠道名称" clearable />
        </el-form-item>
        <el-form-item label="渠道类型">
          <el-select v-model="searchForm.type" placeholder="请选择渠道类型" clearable>
            <el-option v-for="type in channelTypes" :key="type" :value="type" :label="type" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增渠道</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 渠道列表 -->
    <div class="channel-list">
      <el-card v-for="channel in channels" :key="channel.id" class="channel-card">
        <div class="channel-header" @click="toggleExpand(channel)">
          <div class="channel-basic-info">
            <div class="channel-title">
              <h3>{{ channel.name }}</h3>
              <el-tooltip
                :content="!channel.status ? '已停用' : '已启用'"
                placement="top"
              >
                <el-icon
                  class="status-icon"
                  :class="{ 'is-active': channel.status }"
                >
                  <CircleCheck v-if="channel.status" />
                  <CircleClose v-else />
                </el-icon>
              </el-tooltip>
              <el-tag type="warning" size="small">{{ channel.type }}</el-tag>
            </div>
            <div class="channel-models">
              <el-tag
                v-for="model in channel.models"
                :key="model"
                type="info"
                effect="plain"
                size="small"
              >
                {{ model }}
              </el-tag>
            </div>
          </div>
          <div class="channel-actions">
            <el-button text type="primary">
              {{ channel.isExpanded ? '收起' : '展开' }}
              <el-icon>
                <ArrowDown v-if="!channel.isExpanded" />
                <ArrowUp v-else />
              </el-icon>
            </el-button>
          </div>
        </div>

        <!-- 展开的详细信息 -->
        <el-collapse-transition>
          <div v-show="channel.isExpanded" class="channel-detail">
            <el-form
              :model="channel"
              label-width="100px"
              class="channel-form"
            >
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="渠道名称">
                    <el-input v-model="channel.name" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="渠道类型">
                    <el-select v-model="channel.type" class="w-full">
                      <el-option
                        v-for="type in channelTypes"
                        :key="type"
                        :value="type"
                        :label="type"
                      />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="优先级">
                    <el-input-number v-model="channel.priority" :min="0" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="权重">
                    <el-input-number v-model="channel.weight" :min="0" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="重试次数">
                    <el-input-number v-model="channel.retry" :min="0" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="启用状态">
                    <el-switch
                      v-model="channel.status"
                      :active-value="true"
                      :inactive-value="false"
                      active-text="启用"
                      inactive-text="停用"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="24">
                  <el-form-item label="服务地址">
                    <el-input v-model="channel.endpoint" placeholder="请输入服务地址" />
                  </el-form-item>
                </el-col>
                <el-col :span="24">
                  <el-form-item label="支持模型">
                    <el-select
                      v-model="channel.models"
                      multiple
                      filterable
                      class="w-full"
                      placeholder="请选择支持的模型"
                    >
                      <el-option
                        v-for="model in availableModels"
                        :key="model"
                        :value="model"
                        :label="model"
                      />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="24">
                  <el-form-item label="密钥">
                    <el-input
                      v-model="channel.secret"
                      type="textarea"
                      :rows="4"
                      placeholder="请输入密钥，多个密钥请换行输入"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="24">
                  <el-form-item label="模型映射" class="mapping-form-item">
                    <div class="mapping-container">
                      <div v-for="(mapping, mappingIndex) in channel.mappings" :key="mappingIndex" class="mapping-item">
                        <div class="mapping-header">
                          <div class="mapping-source">
                            <el-select v-model="mapping.sourceModel" placeholder="选择源模型" class="mapping-select">
                              <el-option
                                v-for="model in availableModels"
                                :key="model"
                                :value="model"
                                :label="model"
                              />
                            </el-select>
                            <el-icon class="mapping-arrow"><ArrowRight /></el-icon>
                          </div>
                          <el-button type="danger" @click="removeMapping(channel, mappingIndex)" circle>
                            <el-icon><Delete /></el-icon>
                          </el-button>
                        </div>

                        <div class="conditions-container">
                          <div v-for="(condition, conditionIndex) in mapping.conditions" :key="conditionIndex" class="condition-item">
                            <div class="condition-header">
                              <div class="condition-target">
                                <el-input v-model="condition.targetModel" placeholder="输入目标模型" class="mapping-input" />
                                <div class="condition-options">
                                  <el-button
                                    type="primary"
                                    plain
                                    size="small"
                                    @click="addConditionConfig(condition)"
                                  >
                                    <el-icon><Plus /></el-icon>
                                    添加条件
                                  </el-button>
                                </div>
                              </div>
                              <el-button type="danger" @click="removeCondition(channel, mappingIndex, conditionIndex)" circle>
                                <el-icon><Delete /></el-icon>
                              </el-button>
                            </div>

                            <div v-if="condition.conditionConfigs && condition.conditionConfigs.length > 0" class="condition-configs">
                              <div v-for="(config, configIndex) in condition.conditionConfigs" :key="configIndex" class="condition-config-item">
                                <el-select v-model="config.key" placeholder="选择条件" class="condition-key-select">
                                  <el-option label="网页搜索" value="enableWeb" />
                                </el-select>
                                <el-switch
                                  v-if="config.key === 'enableWeb'"
                                  v-model="config.value"
                                  active-text="启用"
                                  inactive-text="禁用"
                                />
                                <el-button type="danger" @click="removeConditionConfig(condition, configIndex)" circle>
                                  <el-icon><Delete /></el-icon>
                                </el-button>
                              </div>
                            </div>
                          </div>

                          <el-button type="primary" @click="addCondition(channel, mappingIndex)" class="add-condition-btn" plain>
                            <el-icon><Plus /></el-icon>
                            添加目标模型
                          </el-button>
                        </div>
                      </div>

                      <el-button type="primary" @click="addMapping(channel)" class="add-mapping-btn" plain>
                        <el-icon><Plus /></el-icon>
                        添加映射规则
                      </el-button>
                    </div>
                  </el-form-item>
                </el-col>
              </el-row>
              <div class="form-actions">
                <el-button type="primary" @click="handleSave(channel)">保存</el-button>
                <el-button type="danger" @click="handleDelete(channel)">删除</el-button>
              </div>
            </el-form>
          </div>
        </el-collapse-transition>
      </el-card>
    </div>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 30, 50]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增渠道' : '编辑渠道'"
      width="700px"
    >
      <el-form
        ref="channelFormRef"
        :model="channelForm"
        :rules="channelRules"
        label-width="100px"
      >
        <el-form-item label="渠道名称" prop="name">
          <el-input v-model="channelForm.name" placeholder="请输入渠道名称" />
        </el-form-item>
        <el-form-item label="渠道类型" prop="type">
          <el-select v-model="channelForm.type" class="w-full" placeholder="请选择渠道类型">
            <el-option
              v-for="type in channelTypes"
              :key="type"
              :value="type"
              :label="type"
            />
          </el-select>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="优先级" prop="priority">
              <el-input-number
                v-model="channelForm.priority"
                :min="0"
                :max="999"
                controls-position="right"
                class="w-full"
                :precision="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="权重" prop="weight">
              <el-input-number
                v-model="channelForm.weight"
                :min="0"
                :max="999"
                controls-position="right"
                class="w-full"
                :precision="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="重试次数" prop="retry">
              <el-input-number
                v-model="channelForm.retry"
                :min="0"
                :max="10"
                controls-position="right"
                class="w-full"
                :precision="0"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="服务地址" prop="endpoint">
          <el-input v-model="channelForm.endpoint" placeholder="请输入服务地址" />
        </el-form-item>
        <el-form-item label="支持模型" prop="models">
          <el-select
            v-model="channelForm.models"
            multiple
            filterable
            class="w-full"
            placeholder="请选择支持的模型"
          >
            <el-option
              v-for="model in availableModels"
              :key="model"
              :value="model"
              :label="model"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="密钥" prop="secret">
          <el-input
            v-model="channelForm.secret"
            type="textarea"
            :rows="4"
            placeholder="请输入密钥，多个密钥请换行输入"
          />
        </el-form-item>
        <el-form-item label="模型映射" prop="mappings">
          <div class="mapping-container">
            <div v-for="(mapping, mappingIndex) in channelForm.mappings" :key="mappingIndex" class="mapping-item">
              <div class="mapping-header">
                <div class="mapping-source">
                  <el-select v-model="mapping.sourceModel" placeholder="选择源模型" class="mapping-select">
                    <el-option
                      v-for="model in availableModels"
                      :key="model"
                      :value="model"
                      :label="model"
                    />
                  </el-select>
                  <el-icon class="mapping-arrow"><ArrowRight /></el-icon>
                </div>
                <el-button type="danger" @click="removeMapping(channelForm, mappingIndex)" circle>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>

              <div class="conditions-container">
                <div v-for="(condition, conditionIndex) in mapping.conditions" :key="conditionIndex" class="condition-item">
                  <div class="condition-header">
                    <div class="condition-target">
                      <el-input v-model="condition.targetModel" placeholder="输入目标模型" class="mapping-input" />
                      <div class="condition-options">
                        <el-button
                          type="primary"
                          plain
                          size="small"
                          @click="addConditionConfig(condition)"
                        >
                          <el-icon><Plus /></el-icon>
                          添加条件
                        </el-button>
                      </div>
                    </div>
                    <el-button type="danger" @click="removeCondition(channelForm, mappingIndex, conditionIndex)" circle>
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>

                  <div v-if="condition.conditionConfigs && condition.conditionConfigs.length > 0" class="condition-configs">
                    <div v-for="(config, configIndex) in condition.conditionConfigs" :key="configIndex" class="condition-config-item">
                      <el-select v-model="config.key" placeholder="选择条件" class="condition-key-select">
                        <el-option label="网页搜索" value="enableWeb" />
                      </el-select>
                      <el-switch
                        v-if="config.key === 'enableWeb'"
                        v-model="config.value"
                        active-text="启用"
                        inactive-text="禁用"
                      />
                      <el-button type="danger" @click="removeConditionConfig(condition, configIndex)" circle>
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </div>

                <el-button type="primary" @click="addCondition(channelForm, mappingIndex)" class="add-condition-btn" plain>
                  <el-icon><Plus /></el-icon>
                  添加目标模型
                </el-button>
              </div>
            </div>

            <el-button type="primary" @click="addMapping(channelForm)" class="add-mapping-btn" plain>
              <el-icon><Plus /></el-icon>
              添加映射规则
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="启用状态" prop="status">
          <el-switch
            v-model="channelForm.status"
            :active-value="true"
            :inactive-value="false"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, ArrowUp, CircleCheck, CircleClose, Delete, InfoFilled, Plus, ArrowRight } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

// 搜索表单
const searchForm = ref({
  name: '',
  type: ''
})

// 分页数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 渠道数据
const channels = ref([])
const channelTypes = ref(['火星引擎', 'polo'])
const availableModels = ref([])
const loadingModels = ref(false)

// 对话框数据
const dialogVisible = ref(false)
const dialogType = ref('add')
const channelFormRef = ref(null)
const channelForm = ref({
  name: '',
  type: '',
  priority: 0,
  weight: 0,
  retry: 3,
  models: [],
  secret: '',
  endpoint: '',
  status: false,
  mappings: []
})

// 表单验证规则
const channelRules = {
  name: [
    { required: true, message: '请输入渠道名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择渠道类型', trigger: 'change' }
  ],
  endpoint: [
    { required: true, message: '请输入服务地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL地址', trigger: 'blur' }
  ],
  models: [
    { required: true, message: '请选择支持的模型', trigger: 'change' },
    { type: 'array', min: 1, message: '请至少选择一个模型', trigger: 'change' }
  ],
  mappings: [
    {
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback()
          return
        }

        for (const mapping of value) {
          if (!mapping.sourceModel) {
            callback(new Error('请选择源模型'))
            return
          }

          if (!mapping.conditions || mapping.conditions.length === 0) {
            callback(new Error('请至少添加一个条件'))
            return
          }

          for (const condition of mapping.conditions) {
            if (!condition.targetModel) {
              callback(new Error('请选择目标模型'))
              return
            }
          }
        }

        callback()
      },
      trigger: 'change'
    }
  ]
}

// 加载渠道列表
const loadChannels = async () => {
  try {
    const response = await defaultApi.apiAdminChannelListGet(
      currentPage.value,
      pageSize.value,
      {
        orderBy: 'id',
        order: 'desc',
        type: searchForm.value.type || undefined,
        name: searchForm.value.name || undefined
      }
    )
    if (response.code === 0 && response.data) {
      channels.value = response.data.records.map(channel => ({
        ...channel,
        isExpanded: false,
        models: channel.models ? channel.models : [],
        mappings: channel.mappings ? channel.mappings.map(mapping => ({
          sourceModel: mapping.sourceModel,
          conditions: mapping.conditions.map(condition => {
            const result = {
              targetModel: condition.targetModel,
              conditionConfigs: []
            }

            // 如果有条件配置，转换为 conditionConfigs 格式
            if (condition.conditions) {
              Object.entries(condition.conditions).forEach(([key, value]) => {
                result.conditionConfigs.push({
                  key,
                  value
                })
              })
            }

            return result
          })
        })) : []
      }))
      total.value = response.data.total || 0
      currentPage.value = response.data.page || 1
      pageSize.value = response.data.limit || 10
    } else {
      ElMessage.error(response.msg || '获取渠道列表失败')
    }
  } catch (error) {
    console.error('Load channels error:', error)
    ElMessage.error(error.body?.msg || '获取渠道列表失败')
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadChannels()
}

// 重置搜索
const handleReset = () => {
  searchForm.value = {
    name: '',
    type: ''
  }
  currentPage.value = 1
  loadChannels()
}

// 切换展开/收起
const toggleExpand = (channel) => {
  channel.isExpanded = !channel.isExpanded
}

// 新增渠道
const handleAdd = () => {
  dialogType.value = 'add'
  channelForm.value = {
    name: '',
    type: '',
    priority: 0,
    weight: 0,
    retry: 3,
    models: [],
    secret: '',
    endpoint: '',
    status: false,
    mappings: []
  }
  dialogVisible.value = true
}

// 保存渠道
const handleSave = async (channel) => {
  try {
    const formData = {
      name: channel.name,
      type: channel.type,
      priority: channel.priority,
      weight: channel.weight,
      retry: channel.retry,
      models: channel.models,
      secret: channel.secret,
      endpoint: channel.endpoint,
      status: channel.status,
      mappings: channel.mappings.map(mapping => ({
        sourceModel: mapping.sourceModel,
        conditions: mapping.conditions.map(condition => {
          const result = {
            targetModel: condition.targetModel,
            conditions: {}
          }

          if (condition.conditionConfigs && condition.conditionConfigs.length > 0) {
            condition.conditionConfigs.forEach(config => {
              result.conditions[config.key] = config.value
            })
          }

          return result
        })
      }))
    }

    const response = await defaultApi.apiAdminChannelIdPut(channel.id, formData)
    if (response.code === 0) {
      ElMessage.success('保存成功')
      await loadChannels()
    } else {
      ElMessage.error(response.msg || '保存失败')
    }
  } catch (error) {
    console.error('Save channel error:', error)
    ElMessage.error(error.body?.msg || '保存失败')
  }
}

// 删除渠道
const handleDelete = async (channel) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除渠道【${channel.name}】吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await defaultApi.apiAdminChannelIdDelete(channel.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      await loadChannels()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete channel error:', error)
      ElMessage.error(error.body?.msg || '删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!channelFormRef.value) return
  await channelFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const formData = {
          name: channelForm.value.name,
          type: channelForm.value.type,
          priority: channelForm.value.priority,
          weight: channelForm.value.weight,
          retry: channelForm.value.retry,
          models: channelForm.value.models,
          secret: channelForm.value.secret,
          endpoint: channelForm.value.endpoint,
          status: channelForm.value.status,
          mappings: channelForm.value.mappings.map(mapping => ({
            sourceModel: mapping.sourceModel,
            conditions: mapping.conditions.map(condition => {
              const result = {
                targetModel: condition.targetModel,
                conditions: {}
              }

              if (condition.conditionConfigs && condition.conditionConfigs.length > 0) {
                condition.conditionConfigs.forEach(config => {
                  result.conditions[config.key] = config.value
                })
              }

              return result
            })
          }))
        }

        let response
        if (dialogType.value === 'add') {
          response = await defaultApi.apiAdminChannelPost(formData)
        } else {
          response = await defaultApi.apiAdminChannelIdPut(channelForm.value.id, formData)
        }

        if (response.code === 0) {
          ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
          dialogVisible.value = false
          await loadChannels()
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('Submit error:', error)
        ElMessage.error(error.body?.msg || '操作失败')
      }
    }
  })
}

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadChannels()
}

// 页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  loadChannels()
}

// 添加映射规则
const addMapping = (channel) => {
  if (!channel.mappings) {
    channel.mappings = []
  }
  channel.mappings.push({
    sourceModel: '',
    conditions: []
  })
}

// 移除映射规则
const removeMapping = (channel, mappingIndex) => {
  channel.mappings.splice(mappingIndex, 1)
}

// 添加条件
const addCondition = (channel, mappingIndex) => {
  channel.mappings[mappingIndex].conditions.push({
    targetModel: '',
    conditionConfigs: []
  })
}

// 添加条件配置
const addConditionConfig = (condition) => {
  if (!condition.conditionConfigs) {
    condition.conditionConfigs = []
  }
  condition.conditionConfigs.push({
    key: 'enableWeb',
    value: false
  })
}

// 移除条件配置
const removeConditionConfig = (condition, configIndex) => {
  condition.conditionConfigs.splice(configIndex, 1)
}

// 移除条件
const removeCondition = (channel, mappingIndex, conditionIndex) => {
  channel.mappings[mappingIndex].conditions.splice(conditionIndex, 1)
}

// 加载模型列表
const loadModels = async () => {
  try {
    loadingModels.value = true
    const response = await defaultApi.apiModelListGet(1, 999, {
      orderBy: 'id',
      order: 'desc'
    })

    if (response.code === 0 && response.data) {
      availableModels.value = response.data.records.map(model => model.name)
    } else {
      ElMessage.error(response.msg || '获取模型列表失败')
    }
  } catch (error) {
    console.error('Load models error:', error)
    ElMessage.error(error.body?.msg || '获取模型列表失败')
  } finally {
    loadingModels.value = false
  }
}

// 在组件挂载时加载数据
onMounted(async () => {
  await Promise.all([
    loadChannels(),
    loadModels()
  ])
})
</script>

<style lang="scss" scoped>
.channel-container {
  padding: 24px;

  :deep(.el-card) {
    border-radius: 16px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    overflow: hidden;

    &:hover {
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
    }
  }
}

.search-form {
  margin-bottom: 24px;

  :deep(.el-form-item) {
    margin-bottom: 0;
  }

  :deep(.el-input__wrapper) {
    border-radius: 12px;
  }

  :deep(.el-select) {
    width: 180px;
  }

  :deep(.el-button) {
    border-radius: 12px;
    padding: 12px 24px;
    transition: all 0.3s;

    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    &.el-button--primary {
      &:hover {
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.3);
      }
    }
  }
}

.channel-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
}

.channel-card {
  .channel-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 20px;
    cursor: pointer;
  }

  .channel-basic-info {
    flex: 1;
  }

  .channel-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;

    h3 {
      margin: 0;
      font-size: 18px;
      line-height: 1.4;
    }

    .status-icon {
      font-size: 16px;
      color: var(--el-text-color-secondary);
      transition: all 0.3s;

      &.is-active {
        color: var(--el-color-success);
      }

      &:not(.is-active) {
        color: var(--el-color-danger);
      }
    }
  }

  .channel-models {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .channel-detail {
    padding: 0 20px 20px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .channel-form {
    padding-top: 20px;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
  }
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;

  :deep(.el-pagination) {
    padding: 12px 24px;
    border-radius: 12px;
    background: var(--el-bg-color);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }
}

.w-full {
  width: 100%;

  :deep(.el-input-number) {
    width: 100%;

    .el-input__wrapper {
      padding-left: 8px;
      padding-right: 35px;
    }

    .el-input__inner {
      text-align: center;
      padding: 0 30px;
    }

    .el-input-number__decrease,
    .el-input-number__increase {
      width: 32px;
      height: 100%;
      top: 1px;
      background: var(--el-fill-color-light);

      &:hover {
        color: var(--el-color-primary);
        background: var(--el-fill-color);
      }
    }

    .el-input-number__decrease {
      left: 1px;
      border-radius: 4px 0 0 4px;
    }

    .el-input-number__increase {
      right: 1px;
      border-radius: 0 4px 4px 0;
    }
  }
}

.mapping-form-item {
  :deep(.el-form-item__content) {
    width: 100%;
  }
}

.mapping-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 32px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 16px;
  background-color: var(--el-fill-color-blank);
  width: 100%;
}

.mapping-item {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 32px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  background-color: var(--el-fill-color-blank);
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  }
}

.mapping-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.mapping-source {
  display: flex;
  align-items: center;
  gap: 24px;
}

.mapping-arrow {
  color: var(--el-text-color-secondary);
  font-size: 24px;
}

.mapping-select {
  width: 320px;
}

.conditions-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-left: 48px;
  padding-left: 32px;
  border-left: 2px dashed var(--el-border-color-lighter);
}

.condition-item {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background-color: var(--el-fill-color-light);
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }
}

.condition-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.condition-target {
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex: 1;
}

.condition-options {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-top: 8px;
}

.condition-type-group {
  :deep(.el-radio-button__inner) {
    padding: 8px 16px;
  }
}

.web-search-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;

  .checkbox-label {
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .info-icon {
    color: var(--el-text-color-secondary);
    font-size: 16px;
    cursor: help;
  }
}

.add-condition-btn, .add-mapping-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: fit-content;
  margin-top: 8px;
  border-radius: 8px;
  padding: 12px 24px;

  .el-icon {
    font-size: 16px;
  }
}

.add-mapping-btn {
  margin-top: 16px;
  padding: 12px 32px;
}

.mapping-input {
  width: 320px;
}

.condition-configs {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
  padding-left: 24px;
}

.condition-config-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background-color: var(--el-fill-color-blank);
}

.condition-key-select {
  width: 160px;
}

.model-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;

  .model-name {
    font-weight: 500;
  }

  .model-desc {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}

:deep(.el-select-dropdown__item) {
  padding: 8px 12px;
}

:deep(.el-select-dropdown__item.selected) {
  .model-option {
    .model-name {
      color: var(--el-color-primary);
    }
  }
}
</style>
