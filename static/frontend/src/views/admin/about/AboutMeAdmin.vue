<template>
  <div class="about-admin-container">
    <div class="page-header">
      <h2>关于我页面配置</h2>
      <p class="page-desc">在这里维护 /about 页面的所有内容（Hero、Reasons、Skills、Projects、Timeline、Contact），前台会立即生效。</p>
    </div>

    <el-tabs v-model="activeTab" class="about-tabs" type="border-card">
      <!-- ==================== Hero 区 ==================== -->
      <el-tab-pane label="Hero 顶部" name="hero">
        <el-card>
          <el-form :model="heroForm" label-width="120px" v-loading="heroLoading">
            <el-form-item label="头像文字">
              <el-input v-model="heroForm.avatarText" maxlength="5" show-word-limit placeholder="如 T" />
            </el-form-item>
            <el-form-item label="状态徽标">
              <el-input v-model="heroForm.statusText" placeholder="如 Ready for New Challenges" />
            </el-form-item>
            <el-form-item label="主标题姓名">
              <el-input v-model="heroForm.name" placeholder="如 Txing（用于打字机效果）" />
            </el-form-item>
            <el-form-item label="副标题">
              <el-input v-model="heroForm.subtitle" placeholder="用于打字机效果的副标题" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveHero">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- ==================== 浮动图标 ==================== -->
      <el-tab-pane label="浮动图标" name="floating">
        <div class="pane-toolbar">
          <el-button type="primary" @click="openFloatingDialog()">新增浮动图标</el-button>
        </div>
        <el-table :data="floatingList" border v-loading="floatingLoading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="symbol" label="符号">
            <template #default="{ row }">
              <span style="font-size: 20px">{{ row.symbol }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="sort" label="排序" width="100" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button text type="primary" @click="openFloatingDialog(row)">编辑</el-button>
              <el-button text type="danger" @click="deleteFloating(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 为什么选择我 ==================== -->
      <el-tab-pane label="为什么选择我" name="reasons">
        <div class="pane-toolbar">
          <el-button type="primary" @click="openReasonDialog()">新增卡片</el-button>
        </div>
        <el-table :data="reasonList" border v-loading="reasonLoading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="emoji" label="图标" width="80">
            <template #default="{ row }">
              <span style="font-size: 20px">{{ row.emoji }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="desc" label="描述" show-overflow-tooltip />
          <el-table-column prop="sort" label="排序" width="80" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button text type="primary" @click="openReasonDialog(row)">编辑</el-button>
              <el-button text type="danger" @click="deleteReason(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 核心能力 ==================== -->
      <el-tab-pane label="核心能力" name="skills">
        <div class="pane-toolbar">
          <el-button type="primary" @click="openSkillDialog()">新增技能</el-button>
        </div>
        <el-table :data="skillList" border v-loading="skillLoading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="category" label="分类" />
          <el-table-column prop="iconKey" label="图标Key" width="140" />
          <el-table-column label="标签">
            <template #default="{ row }">
              <el-tag
                v-for="t in row.tags || []"
                :key="t"
                size="small"
                style="margin-right: 4px"
              >{{ t }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="level" label="熟练度" width="100" />
          <el-table-column prop="sort" label="排序" width="80" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button text type="primary" @click="openSkillDialog(row)">编辑</el-button>
              <el-button text type="danger" @click="deleteSkill(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 精选作品 ==================== -->
      <el-tab-pane label="精选作品" name="projects">
        <div class="pane-toolbar">
          <el-button type="primary" @click="openProjectDialog()">新增项目</el-button>
        </div>
        <el-table :data="projectList" border v-loading="projectLoading">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="项目名" width="160" />
          <el-table-column prop="iconKey" label="图标" width="120" />
          <el-table-column prop="gradient" label="渐变" width="80" />
          <el-table-column prop="badge" label="角标" width="100" />
          <el-table-column prop="link" label="链接" show-overflow-tooltip />
          <el-table-column prop="sort" label="排序" width="70" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button text type="primary" @click="openProjectDialog(row)">编辑</el-button>
              <el-button text type="danger" @click="deleteProject(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 成长轨迹 ==================== -->
      <el-tab-pane label="成长轨迹" name="timeline">
        <div class="pane-toolbar">
          <el-button type="primary" @click="openTimelineDialog()">新增时间线</el-button>
        </div>
        <el-table :data="timelineList" border v-loading="timelineLoading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="time" label="时间" width="160" />
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="desc" label="描述" show-overflow-tooltip />
          <el-table-column prop="sort" label="排序" width="80" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button text type="primary" @click="openTimelineDialog(row)">编辑</el-button>
              <el-button text type="danger" @click="deleteTimeline(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 联系区 ==================== -->
      <el-tab-pane label="联系区" name="contact">
        <el-card>
          <el-form :model="contactForm" label-width="120px" v-loading="contactLoading">
            <el-form-item label="标题">
              <el-input v-model="contactForm.title" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="contactForm.desc" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="链接">
              <div class="link-list">
                <div
                  v-for="(link, idx) in contactForm.links"
                  :key="idx"
                  class="link-row"
                >
                  <el-input
                    v-model="link.iconKey"
                    placeholder="图标Key，如 Message/Github"
                    style="width: 180px"
                  />
                  <el-input
                    v-model="link.label"
                    placeholder="链接名称"
                    style="width: 200px"
                  />
                  <el-input
                    v-model="link.url"
                    placeholder="URL（http:// 或 mailto:）"
                    style="flex: 1"
                  />
                  <el-button
                    type="danger"
                    text
                    @click="contactForm.links.splice(idx, 1)"
                  >删除</el-button>
                </div>
                <el-button @click="contactForm.links.push({ iconKey: '', label: '', url: '' })">
                  + 添加链接
                </el-button>
              </div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveContact">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- ==================== 浮动图标对话框 ==================== -->
    <el-dialog
      v-model="floatingDialogVisible"
      :title="floatingForm.id ? '编辑浮动图标' : '新增浮动图标'"
      width="480px"
    >
      <el-form :model="floatingForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="floatingForm.name" />
        </el-form-item>
        <el-form-item label="符号">
          <el-input v-model="floatingForm.symbol" placeholder="emoji 或字符" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="floatingForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="floatingDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveFloating">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Reason 对话框 ==================== -->
    <el-dialog
      v-model="reasonDialogVisible"
      :title="reasonForm.id ? '编辑卡片' : '新增卡片'"
      width="640px"
    >
      <el-form :model="reasonForm" label-width="100px">
        <el-form-item label="Emoji">
          <el-input v-model="reasonForm.emoji" maxlength="5" />
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="reasonForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="reasonForm.desc" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            :model-value="(reasonForm.tags || []).join(',')"
            @update:model-value="(v) => (reasonForm.tags = v ? v.split(',').map(s => s.trim()).filter(Boolean) : [])"
            placeholder="多个标签用英文逗号分隔"
          />
        </el-form-item>
        <el-form-item label="统计数据">
          <div class="stats-list">
            <div
              v-for="(stat, idx) in reasonForm.stats || []"
              :key="idx"
              class="stat-row"
            >
              <el-input v-model="stat.value" placeholder="数值，如 10+" style="width: 120px" />
              <el-input v-model="stat.label" placeholder="标签名" style="flex: 1" />
              <el-button text type="danger" @click="reasonForm.stats.splice(idx, 1)">删除</el-button>
            </div>
            <el-button @click="(reasonForm.stats = reasonForm.stats || []).push({ value: '', label: '' })">
              + 添加统计
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="reasonForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reasonDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveReason">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Skill 对话框 ==================== -->
    <el-dialog
      v-model="skillDialogVisible"
      :title="skillForm.id ? '编辑技能' : '新增技能'"
      width="560px"
    >
      <el-form :model="skillForm" label-width="100px">
        <el-form-item label="分类">
          <el-input v-model="skillForm.category" />
        </el-form-item>
        <el-form-item label="图标Key">
          <el-select v-model="skillForm.iconKey" filterable placeholder="选择图标">
            <el-option
              v-for="icon in AVAILABLE_ICONS"
              :key="icon.key"
              :label="icon.label"
              :value="icon.key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            :model-value="(skillForm.tags || []).join(',')"
            @update:model-value="(v) => (skillForm.tags = v ? v.split(',').map(s => s.trim()).filter(Boolean) : [])"
            placeholder="多个标签用英文逗号分隔"
          />
        </el-form-item>
        <el-form-item label="熟练度">
          <el-slider v-model="skillForm.level" :min="0" :max="100" show-input />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="skillForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="skillDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSkill">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Project 对话框 ==================== -->
    <el-dialog
      v-model="projectDialogVisible"
      :title="projectForm.id ? '编辑项目' : '新增项目'"
      width="780px"
      top="5vh"
    >
      <el-form :model="projectForm" label-width="100px">
        <el-form-item label="项目名">
          <el-input v-model="projectForm.name" />
        </el-form-item>
        <el-form-item label="图标Key">
          <el-select v-model="projectForm.iconKey" filterable placeholder="选择图标">
            <el-option
              v-for="icon in AVAILABLE_ICONS"
              :key="icon.key"
              :label="icon.label"
              :value="icon.key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="渐变编号">
          <el-input-number v-model="projectForm.gradient" :min="1" :max="6" />
          <span class="form-tip">1-6，对应前台不同的渐变色</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="projectForm.desc" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            :model-value="(projectForm.tags || []).join(',')"
            @update:model-value="(v) => (projectForm.tags = v ? v.split(',').map(s => s.trim()).filter(Boolean) : [])"
            placeholder="多个标签用英文逗号分隔"
          />
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="projectForm.link" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="角标">
          <el-input v-model="projectForm.badge" placeholder="如 旗舰项目" />
        </el-form-item>
        <el-form-item label="亮点">
          <el-input
            :model-value="(projectForm.highlights || []).join('\n')"
            @update:model-value="(v) => (projectForm.highlights = v ? v.split('\n').map(s => s.trim()).filter(Boolean) : [])"
            type="textarea"
            :rows="3"
            placeholder="每行一个亮点"
          />
        </el-form-item>
        <el-form-item label="媒体列表">
          <div class="media-list">
            <div
              v-for="(m, idx) in projectForm.media || []"
              :key="idx"
              class="media-row"
            >
              <el-select v-model="m.type" style="width: 90px" @change="onMediaTypeChange(m)">
                <el-option label="图片" value="image" />
                <el-option label="视频" value="video" />
              </el-select>
              <MediaUploader
                v-model="m.url"
                :media-type="m.type"
                :width="240"
                :height="120"
                :placeholder="`点击上传${m.type === 'video' ? '视频' : '图片'}`"
                :max-size="50"
                @change="(info) => onMediaUploaded(idx, info)"
              />
              <el-input v-model="m.caption" placeholder="说明" style="width: 200px" />
              <el-button text type="danger" @click="removeMedia(idx)">删除</el-button>
            </div>
            <el-button @click="addMediaRow">+ 添加媒体</el-button>
            <p class="form-tip">支持上传图片或视频到 OSS，超过 50MB 会被限制</p>
          </div>
        </el-form-item>
        <el-form-item label="技术栈">
          <div class="tech-list">
            <div
              v-for="(t, idx) in projectForm.techStack || []"
              :key="idx"
              class="tech-row"
            >
              <el-input v-model="t.icon" placeholder="emoji" style="width: 60px" />
              <el-input v-model="t.name" placeholder="技术名" style="flex: 1" />
              <el-button text type="danger" @click="projectForm.techStack.splice(idx, 1)">删除</el-button>
            </div>
            <el-button @click="(projectForm.techStack = projectForm.techStack || []).push({ name: '', icon: '' })">
              + 添加技术栈
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="核心功能">
          <div class="feature-list">
            <div
              v-for="(f, idx) in projectForm.features || []"
              :key="idx"
              class="feature-row"
            >
              <el-input v-model="f.icon" placeholder="emoji" style="width: 60px" />
              <el-input v-model="f.title" placeholder="标题" style="width: 200px" />
              <el-input v-model="f.desc" placeholder="描述" style="flex: 1" />
              <el-button text type="danger" @click="projectForm.features.splice(idx, 1)">删除</el-button>
            </div>
            <el-button @click="(projectForm.features = projectForm.features || []).push({ icon: '', title: '', desc: '' })">
              + 添加功能
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="projectForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="projectDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProject">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Timeline 对话框 ==================== -->
    <el-dialog
      v-model="timelineDialogVisible"
      :title="timelineForm.id ? '编辑时间线' : '新增时间线'"
      width="560px"
    >
      <el-form :model="timelineForm" label-width="100px">
        <el-form-item label="时间范围">
          <el-input v-model="timelineForm.time" placeholder="如 2024 - 至今" />
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="timelineForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="timelineForm.desc" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            :model-value="(timelineForm.tags || []).join(',')"
            @update:model-value="(v) => (timelineForm.tags = v ? v.split(',').map(s => s.trim()).filter(Boolean) : [])"
            placeholder="多个标签用英文逗号分隔"
          />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="timelineForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="timelineDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTimeline">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { defaultApi } from '@/api'
import { AVAILABLE_ICONS } from '@/utils/iconResolver.js'
import MediaUploader from '@/components/common/MediaUploader.vue'

const activeTab = ref('hero')

// ==================== Hero ====================
const heroForm = reactive({
  id: 0,
  avatarText: 'T',
  statusText: '',
  name: '',
  subtitle: ''
})
const heroLoading = ref(false)
const loadHero = async () => {
  heroLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutHeroGet()
    if (res?.code === 0 && res.data) {
      Object.assign(heroForm, res.data)
    }
  } finally {
    heroLoading.value = false
  }
}
const saveHero = async () => {
  try {
    const payload = {
      avatarText: heroForm.avatarText,
      statusText: heroForm.statusText,
      name: heroForm.name,
      subtitle: heroForm.subtitle
    }
    const res = await defaultApi.apiAdminAboutHeroPut(payload)
    if (res?.code === 0) {
      ElMessage.success('保存成功')
    } else {
      ElMessage.error(res?.msg || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}

// ==================== Floating Icon ====================
const floatingList = ref([])
const floatingLoading = ref(false)
const floatingDialogVisible = ref(false)
const floatingForm = reactive({ id: 0, name: '', symbol: '', sort: 0 })

const loadFloating = async () => {
  floatingLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutFloatingIconListGet(1, 200, {})
    if (res?.code === 0) floatingList.value = res.data?.records || []
  } finally {
    floatingLoading.value = false
  }
}
const openFloatingDialog = (row) => {
  Object.assign(floatingForm, { id: 0, name: '', symbol: '', sort: 0 })
  if (row) Object.assign(floatingForm, row)
  floatingDialogVisible.value = true
}
const saveFloating = async () => {
  try {
    if (floatingForm.id) {
      const res = await defaultApi.apiAdminAboutFloatingIconIdPut(floatingForm.id, floatingForm)
      if (res?.code === 0) ElMessage.success('更新成功')
    } else {
      const res = await defaultApi.apiAdminAboutFloatingIconPost(floatingForm)
      if (res?.code === 0) ElMessage.success('创建成功')
    }
    floatingDialogVisible.value = false
    loadFloating()
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}
const deleteFloating = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除浮动图标 "${row.name}" ?`, '提示', { type: 'warning' })
    const res = await defaultApi.apiAdminAboutFloatingIconIdDelete(row.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      loadFloating()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// ==================== Reason ====================
const reasonList = ref([])
const reasonLoading = ref(false)
const reasonDialogVisible = ref(false)
const reasonForm = reactive({
  id: 0,
  emoji: '',
  title: '',
  desc: '',
  tags: [],
  stats: [],
  sort: 0
})

const loadReasons = async () => {
  reasonLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutReasonListGet(1, 200, {})
    if (res?.code === 0) reasonList.value = res.data?.records || []
  } finally {
    reasonLoading.value = false
  }
}
const openReasonDialog = (row) => {
  Object.assign(reasonForm, { id: 0, emoji: '', title: '', desc: '', tags: [], stats: [], sort: 0 })
  if (row) {
    Object.assign(reasonForm, row)
    reasonForm.tags = row.tags || []
    reasonForm.stats = row.stats || []
  }
  reasonDialogVisible.value = true
}
const saveReason = async () => {
  try {
    const payload = {
      emoji: reasonForm.emoji,
      title: reasonForm.title,
      desc: reasonForm.desc,
      tags: reasonForm.tags || [],
      stats: reasonForm.stats || [],
      sort: reasonForm.sort
    }
    if (reasonForm.id) {
      const res = await defaultApi.apiAdminAboutReasonIdPut(reasonForm.id, payload)
      if (res?.code === 0) ElMessage.success('更新成功')
    } else {
      const res = await defaultApi.apiAdminAboutReasonPost(payload)
      if (res?.code === 0) ElMessage.success('创建成功')
    }
    reasonDialogVisible.value = false
    loadReasons()
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}
const deleteReason = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除卡片 "${row.title}" ?`, '提示', { type: 'warning' })
    const res = await defaultApi.apiAdminAboutReasonIdDelete(row.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      loadReasons()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// ==================== Skill ====================
const skillList = ref([])
const skillLoading = ref(false)
const skillDialogVisible = ref(false)
const skillForm = reactive({ id: 0, category: '', iconKey: '', tags: [], level: 0, sort: 0 })

const loadSkills = async () => {
  skillLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutSkillListGet(1, 200, {})
    if (res?.code === 0) skillList.value = res.data?.records || []
  } finally {
    skillLoading.value = false
  }
}
const openSkillDialog = (row) => {
  Object.assign(skillForm, { id: 0, category: '', iconKey: '', tags: [], level: 0, sort: 0 })
  if (row) {
    Object.assign(skillForm, row)
    skillForm.tags = row.tags || []
  }
  skillDialogVisible.value = true
}
const saveSkill = async () => {
  try {
    const payload = {
      category: skillForm.category,
      iconKey: skillForm.iconKey,
      tags: skillForm.tags || [],
      level: skillForm.level,
      sort: skillForm.sort
    }
    if (skillForm.id) {
      const res = await defaultApi.apiAdminAboutSkillIdPut(skillForm.id, payload)
      if (res?.code === 0) ElMessage.success('更新成功')
    } else {
      const res = await defaultApi.apiAdminAboutSkillPost(payload)
      if (res?.code === 0) ElMessage.success('创建成功')
    }
    skillDialogVisible.value = false
    loadSkills()
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}
const deleteSkill = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除技能 "${row.category}" ?`, '提示', { type: 'warning' })
    const res = await defaultApi.apiAdminAboutSkillIdDelete(row.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      loadSkills()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// ==================== Project ====================
const projectList = ref([])
const projectLoading = ref(false)
const projectDialogVisible = ref(false)
const projectForm = reactive({
  id: 0,
  name: '',
  desc: '',
  iconKey: '',
  gradient: 1,
  tags: [],
  link: '',
  badge: '',
  highlights: [],
  media: [],
  techStack: [],
  features: [],
  sort: 0
})

const loadProjects = async () => {
  projectLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutProjectListGet(1, 200, {})
    if (res?.code === 0) projectList.value = res.data?.records || []
  } finally {
    projectLoading.value = false
  }
}
const openProjectDialog = async (row) => {
  Object.assign(projectForm, {
    id: 0, name: '', desc: '', iconKey: '', gradient: 1,
    tags: [], link: '', badge: '', highlights: [], media: [], techStack: [], features: [], sort: 0
  })
  if (row && row.id) {
    try {
      const res = await defaultApi.apiAdminAboutProjectIdGet(row.id)
      if (res?.code === 0) {
        Object.assign(projectForm, res.data)
        projectForm.tags = res.data.tags || []
        projectForm.highlights = res.data.highlights || []
        projectForm.media = res.data.media || []
        projectForm.techStack = res.data.techStack || []
        projectForm.features = res.data.features || []
      }
    } catch (e) {
      ElMessage.error('加载项目失败')
      return
    }
  }
  projectDialogVisible.value = true
}
const saveProject = async () => {
  try {
    const payload = {
      name: projectForm.name,
      desc: projectForm.desc,
      iconKey: projectForm.iconKey,
      gradient: String(projectForm.gradient || 1),
      tags: projectForm.tags || [],
      link: projectForm.link,
      badge: projectForm.badge,
      highlights: projectForm.highlights || [],
      media: projectForm.media || [],
      techStack: projectForm.techStack || [],
      features: projectForm.features || [],
      sort: projectForm.sort
    }
    if (projectForm.id) {
      const res = await defaultApi.apiAdminAboutProjectIdPut(projectForm.id, payload)
      if (res?.code === 0) ElMessage.success('更新成功')
    } else {
      const res = await defaultApi.apiAdminAboutProjectPost(payload)
      if (res?.code === 0) ElMessage.success('创建成功')
    }
    projectDialogVisible.value = false
    loadProjects()
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}
const deleteProject = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除项目 "${row.name}" ?`, '提示', { type: 'warning' })
    const res = await defaultApi.apiAdminAboutProjectIdDelete(row.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      loadProjects()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// 媒体行辅助函数：新增/删除/类型切换/上传完成回调
// Media row helpers: add/remove/change-type/uploaded callback
const addMediaRow = () => {
  if (!projectForm.media) projectForm.media = []
  projectForm.media.push({ type: 'image', url: '', caption: '' })
}
const removeMedia = (idx) => {
  projectForm.media.splice(idx, 1)
}
// 切换类型时清空旧 URL，避免出现类型与资源不匹配
const onMediaTypeChange = (m) => {
  m.url = ''
}
// 上传成功后回填 caption（如果用户已在弹窗里填写）
const onMediaUploaded = (idx, info) => {
  const m = projectForm.media[idx]
  if (!m) return
  m.type = info.type
  if (info.caption) m.caption = info.caption
}

// ==================== Timeline ====================
const timelineList = ref([])
const timelineLoading = ref(false)
const timelineDialogVisible = ref(false)
const timelineForm = reactive({ id: 0, time: '', title: '', desc: '', tags: [], sort: 0 })

const loadTimelines = async () => {
  timelineLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutTimelineListGet(1, 200, {})
    if (res?.code === 0) timelineList.value = res.data?.records || []
  } finally {
    timelineLoading.value = false
  }
}
const openTimelineDialog = (row) => {
  Object.assign(timelineForm, { id: 0, time: '', title: '', desc: '', tags: [], sort: 0 })
  if (row) {
    Object.assign(timelineForm, row)
    timelineForm.tags = row.tags || []
  }
  timelineDialogVisible.value = true
}
const saveTimeline = async () => {
  try {
    const payload = {
      time: timelineForm.time,
      title: timelineForm.title,
      desc: timelineForm.desc,
      tags: timelineForm.tags || [],
      sort: timelineForm.sort
    }
    if (timelineForm.id) {
      const res = await defaultApi.apiAdminAboutTimelineIdPut(timelineForm.id, payload)
      if (res?.code === 0) ElMessage.success('更新成功')
    } else {
      const res = await defaultApi.apiAdminAboutTimelinePost(payload)
      if (res?.code === 0) ElMessage.success('创建成功')
    }
    timelineDialogVisible.value = false
    loadTimelines()
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}
const deleteTimeline = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除时间线 "${row.title}" ?`, '提示', { type: 'warning' })
    const res = await defaultApi.apiAdminAboutTimelineIdDelete(row.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      loadTimelines()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// ==================== Contact ====================
const contactForm = reactive({
  title: '',
  desc: '',
  links: []
})
const contactLoading = ref(false)
const loadContact = async () => {
  contactLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutContactGet()
    if (res?.code === 0 && res.data) {
      contactForm.title = res.data.title || ''
      contactForm.desc = res.data.desc || ''
      contactForm.links = res.data.links || []
    }
  } finally {
    contactLoading.value = false
  }
}
const saveContact = async () => {
  try {
    const res = await defaultApi.apiAdminAboutContactPut({
      title: contactForm.title,
      desc: contactForm.desc,
      links: contactForm.links
    })
    if (res?.code === 0) {
      ElMessage.success('保存成功')
    } else {
      ElMessage.error(res?.msg || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  }
}

// ==================== Init ====================
onMounted(() => {
  loadHero()
  loadFloating()
  loadReasons()
  loadSkills()
  loadProjects()
  loadTimelines()
  loadContact()
})
</script>

<style lang="scss" scoped>
.about-admin-container {
  padding: 0;
}

.page-header {
  margin-bottom: 16px;

  h2 {
    margin: 0 0 8px 0;
    font-size: 20px;
    font-weight: 600;
  }

  .page-desc {
    margin: 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
}

.about-tabs {
  background: var(--el-bg-color);
  border-radius: 8px;
}

.pane-toolbar {
  margin-bottom: 16px;
}

.form-tip {
  margin-left: 12px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.link-list,
.stats-list,
.media-list,
.tech-list,
.feature-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.link-row,
.stat-row,
.media-row,
.tech-row,
.feature-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.media-row {
  flex-wrap: wrap;
}

:deep(.el-tabs__content) {
  padding: 16px;
}

:deep(.el-card) {
  border-radius: 8px;
}
</style>