<template>
  <SectionCard
    id="section-projects"
    title="精选作品"
    icon="FolderOpened"
    description="项目大卡：封面轮播、公司/个人分类、亮点与技术栈"
    :count="items.length"
  >
    <template #actions>
      <el-button round type="primary" size="small" :icon="Plus" @click="openCreate">新增项目</el-button>
    </template>

    <div ref="listRef" class="project-list">
      <div v-for="it in items" :key="it.id" :data-sort-id="it.id" class="project-card">
        <span class="drag-handle" title="拖拽排序">
          <el-icon><Rank /></el-icon>
        </span>
        <!-- 封面缩略：coverMedia[0] 优先，回退渐变+图标（与前台回退逻辑一致） -->
        <!-- thumbnail: coverMedia[0] first, falls back to gradient + icon (same as front page) -->
        <div class="cover-thumb" :class="`gradient-${normalizedGradient(it.gradient)}`">
          <video
            v-if="coverOf(it)?.type === 'video' && coverOf(it)?.url"
            :src="coverOf(it).url"
            muted
            draggable="false"
          />
          <img
            v-else-if="coverOf(it)?.url"
            :src="coverOf(it).url"
            :alt="it.name"
            draggable="false"
          />
          <el-icon v-else :size="22" class="cover-fallback-icon">
            <component :is="resolveIcon(it.iconKey) || 'Platform'" />
          </el-icon>
          <span v-if="it.badge" class="cover-badge">{{ it.badge }}</span>
        </div>
        <div class="project-info">
          <div class="info-top">
            <span class="project-name" :title="it.name">{{ it.name }}</span>
            <el-tag :type="it.category === 'personal' ? 'success' : 'primary'" size="small" round effect="light">
              {{ it.category === 'personal' ? '个人' : '公司' }}
            </el-tag>
          </div>
          <div class="project-desc" :title="it.desc">{{ it.desc }}</div>
          <div class="project-meta">
            <span class="meta-item">
              <el-icon><Collection /></el-icon>技术栈 {{ (it.techStack || []).length }}
            </span>
            <span class="meta-item">
              <el-icon><Picture /></el-icon>媒体 {{ (it.media || []).length + (it.coverMedia || []).length }}
            </span>
            <span class="meta-item">
              <el-icon><Star /></el-icon>亮点 {{ (it.highlights || []).length }}
            </span>
            <a v-if="it.link" class="meta-link" :href="it.link" target="_blank" rel="noopener noreferrer">
              访问 <el-icon><TopRight /></el-icon>
            </a>
          </div>
        </div>
        <div class="card-ops">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(it)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="remove(it)">删除</el-button>
        </div>
      </div>
    </div>
    <el-empty v-if="!items.length" description="还没有精选项目" :image-size="90">
      <el-button round type="primary" :icon="Plus" @click="openCreate">新增项目</el-button>
    </el-empty>
  </SectionCard>

  <EditDrawer
    v-model="drawerVisible"
    :title="form.id ? '编辑项目' : '新增项目'"
    subtitle="保存后前台「精选作品」立即生效"
    size="92%"
    :saving="saving"
    save-text="保存项目"
    @save="save"
  >
    <ProjectEditPanel :form="form" />
  </EditDrawer>
</template>

<script setup>
// 精选作品区块：项目宽卡 + 拖拽 + 大编辑抽屉（详情接口回填）
// Projects section: wide cards + drag sort + large edit drawer (detail API backfill)
import { ref, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, EditPen, Rank, TopRight, Collection, Picture, Star } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { resolveIcon } from '@/utils/iconResolver.js'
import { useDragSort } from '@/composables/useDragSort'
import { normalizeGradient } from '../constants'
import SectionCard from './SectionCard.vue'
import EditDrawer from './EditDrawer.vue'
import ProjectEditPanel from './ProjectEditPanel.vue'

const props = defineProps({
  list: { type: Array, default: () => [] }
})
const emit = defineEmits(['changed'])

const items = ref([])
watch(
  () => props.list,
  (v) => {
    items.value = (v || []).map((x) => ({ ...x }))
  },
  { immediate: true }
)

const listRef = ref(null)
const persistOrder = async (changed) => {
  try {
    await Promise.all(changed.map((it) => defaultApi.apiAdminAboutProjectIdPut(it.id, buildPayload(it))))
    ElMessage.success('排序已保存')
  } catch (e) {
    ElMessage.error('排序保存失败，已还原')
  } finally {
    emit('changed')
  }
}
useDragSort(listRef, { items, persist: persistOrder })

// ===== 卡片展示辅助 =====
const normalizedGradient = (g) => normalizeGradient(g)
const coverOf = (it) => (it.coverMedia || [])[0] || null

// ===== 编辑抽屉 =====
const drawerVisible = ref(false)
const saving = ref(false)
const detailLoading = ref(false)

const emptyForm = () =>
  reactive({
    id: 0,
    name: '',
    desc: '',
    iconKey: '',
    gradient: '1',
    tags: [],
    link: '',
    badge: '',
    category: 'company',
    highlights: [],
    media: [],
    coverMedia: [],
    techStack: [],
    features: [],
    architecture: '',
    sort: 0
  })

const form = emptyForm()

const openCreate = () => {
  Object.assign(form, emptyForm(), { sort: props.list.length })
  drawerVisible.value = true
}

// 编辑走详情接口：后台列表接口不含 media/coverMedia 全量字段
// editing fetches the detail API: the list endpoint lacks full media fields
const openEdit = async (it) => {
  detailLoading.value = true
  try {
    const res = await defaultApi.apiAdminAboutProjectIdGet(it.id)
    if (res?.code !== 0 || !res.data) {
      ElMessage.error('加载项目详情失败')
      return
    }
    const d = res.data
    Object.assign(form, emptyForm(), {
      id: d.id,
      name: d.name || '',
      desc: d.desc || '',
      iconKey: d.iconKey || '',
      gradient: String(normalizeGradient(d.gradient)),
      tags: [...(d.tags || [])],
      link: d.link || '',
      badge: d.badge || '',
      category: d.category || 'company',
      highlights: [...(d.highlights || [])],
      media: (d.media || []).map((m) => ({ ...m })),
      coverMedia: (d.coverMedia || []).map((m) => ({ ...m })),
      techStack: (d.techStack || []).map((t) => ({ ...t })),
      features: (d.features || []).map((f) => ({ ...f })),
      architecture: d.architecture || '',
      sort: d.sort ?? 0
    })
    drawerVisible.value = true
  } catch (e) {
    ElMessage.error('加载项目详情失败：' + (e?.body?.msg || e.message))
  } finally {
    detailLoading.value = false
  }
}

const buildPayload = (f) => ({
  name: f.name || '',
  desc: f.desc || '',
  iconKey: f.iconKey || '',
  // 只提交 1-4；历史脏值归一化
  // persist normalized 1-4 only
  gradient: String(normalizeGradient(f.gradient)),
  tags: f.tags || [],
  link: f.link || '',
  badge: f.badge || '',
  category: f.category || 'company',
  highlights: f.highlights || [],
  // 只提交 key（URL 由后端读取时签名生成，避免落库过期签名地址）
  // persist keys only; URLs are signed server-side on read
  media: (f.media || []).map(({ type, key, caption }) => ({ type, key, caption })),
  coverMedia: (f.coverMedia || []).map(({ type, key }) => ({ type, key })),
  techStack: f.techStack || [],
  features: f.features || [],
  architecture: f.architecture || '',
  sort: f.sort ?? 0
})

const save = async () => {
  if (!form.name) {
    ElMessage.warning('请填写项目名称')
    return
  }
  saving.value = true
  try {
    const res = form.id
      ? await defaultApi.apiAdminAboutProjectIdPut(form.id, buildPayload(form))
      : await defaultApi.apiAdminAboutProjectPost(buildPayload(form))
    if (res?.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      drawerVisible.value = false
      emit('changed')
    } else {
      ElMessage.error(res?.msg || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  } finally {
    saving.value = false
  }
}

const remove = async (it) => {
  try {
    await ElMessageBox.confirm(`确定删除项目「${it.name}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await defaultApi.apiAdminAboutProjectIdDelete(it.id)
    if (res?.code === 0) {
      ElMessage.success('删除成功')
      emit('changed')
    }
  } catch (e) {
    ElMessage.error('删除失败：' + (e?.body?.msg || e.message))
  }
}
</script>

<style lang="scss" scoped>
.project-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.project-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px 12px 34px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background: var(--el-bg-color);
  transition: box-shadow 0.2s ease, transform 0.2s ease;

  // 拖拽占位态（sortablejs ghost class 挂在卡片本体上）
  // drag placeholder state (ghost class lands on the card itself)
  &.drag-ghost {
    opacity: 0.4;
    outline: 2px dashed var(--el-color-primary);
  }

  &:hover {
    box-shadow: var(--el-box-shadow-light);
    transform: translateY(-2px);

    .card-ops {
      opacity: 1;
    }
  }

  .drag-handle {
    position: absolute;
    top: 50%;
    left: 10px;
    transform: translateY(-50%);
    cursor: grab;
    color: var(--el-text-color-placeholder);

    &:hover {
      color: var(--el-color-primary);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .cover-thumb {
    position: relative;
    width: 128px;
    height: 80px;
    border-radius: 10px;
    overflow: hidden;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;

    // 渐变回退色板：与前台 .project-gradient-1..4 逐字一致
    // fallback gradients identical to the front page
    &.gradient-1 {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }

    &.gradient-2 {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    }

    &.gradient-3 {
      background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
    }

    &.gradient-4 {
      background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
    }

    img,
    video {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .cover-fallback-icon {
      color: rgb(255 255 255 / 92%);
    }

    .cover-badge {
      position: absolute;
      top: 6px;
      right: 6px;
      padding: 2px 8px;
      border-radius: 999px;
      font-size: 11px;
      color: #fff;
      background: rgb(0 0 0 / 45%);
      backdrop-filter: blur(4px);
    }
  }

  .project-info {
    flex: 1;
    min-width: 0;

    .info-top {
      display: flex;
      align-items: center;
      gap: 8px;

      .project-name {
        font-size: 15px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .project-desc {
      margin-top: 4px;
      font-size: 12px;
      line-height: 1.6;
      color: var(--el-text-color-regular);
      display: -webkit-box;
      -webkit-line-clamp: 1;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .project-meta {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 14px;
      margin-top: 8px;

      .meta-item {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--el-text-color-secondary);
      }

      .meta-link {
        display: inline-flex;
        align-items: center;
        gap: 2px;
        font-size: 12px;
        color: var(--el-color-primary);
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }
    }
  }

  .card-ops {
    flex-shrink: 0;
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
}
</style>
