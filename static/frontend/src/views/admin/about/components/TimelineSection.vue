<template>
  <SectionCard
    id="section-timeline"
    title="成长轨迹"
    icon="Timer"
    description="个人时间线：时间段 + 事件标题 + 描述 + 标签"
    :count="items.length"
  >
    <template #actions>
      <el-button round type="primary" size="small" :icon="Plus" @click="openCreate">新增节点</el-button>
    </template>

    <div ref="listRef" class="timeline-list">
      <div v-for="it in items" :key="it.id" :data-sort-id="it.id" class="timeline-row">
        <span class="drag-handle" title="拖拽排序">
          <el-icon><Rank /></el-icon>
        </span>
        <div class="time-col">
          <span class="time-pill">{{ it.time }}</span>
        </div>
        <div class="content-col">
          <div class="row-title">{{ it.title }}</div>
          <div class="row-desc" :title="it.desc">{{ it.desc }}</div>
          <div v-if="(it.tags || []).length" class="row-tags">
            <el-tag v-for="t in it.tags" :key="t" size="small" round effect="plain" type="info">
              {{ t }}
            </el-tag>
          </div>
        </div>
        <div class="row-ops">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(it)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="remove(it)">删除</el-button>
        </div>
      </div>
    </div>
    <el-empty v-if="!items.length" description="还没有时间线节点" :image-size="90">
      <el-button round type="primary" :icon="Plus" @click="openCreate">新增节点</el-button>
    </el-empty>
  </SectionCard>

  <EditDrawer
    v-model="drawerVisible"
    :title="form.id ? '编辑节点' : '新增节点'"
    subtitle="前台「成长轨迹」时间轴"
    size="520px"
    :saving="saving"
    @save="save"
  >
    <el-form :model="form" label-width="80px">
      <el-form-item label="时间范围">
        <el-input v-model="form.time" maxlength="20" placeholder="如 2024 - 至今" />
      </el-form-item>
      <el-form-item label="标题">
        <el-input v-model="form.title" maxlength="30" placeholder="如 深化 AI 工程方向" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.desc" type="textarea" :rows="3" maxlength="200" show-word-limit placeholder="该阶段的关键事件与成长" />
      </el-form-item>
      <el-form-item label="标签">
        <TagsInput v-model="form.tags" />
      </el-form-item>
    </el-form>
  </EditDrawer>
</template>

<script setup>
// 成长轨迹区块：时间线条目 + 拖拽 + 抽屉编辑
// Timeline section: entries + drag sort + edit drawer
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, EditPen, Rank } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { useDragSort } from '@/composables/useDragSort'
import SectionCard from './SectionCard.vue'
import EditDrawer from './EditDrawer.vue'
import TagsInput from './editors/TagsInput.vue'

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
    await Promise.all(changed.map((it) => defaultApi.apiAdminAboutTimelineIdPut(it.id, buildPayload(it))))
    ElMessage.success('排序已保存')
  } catch (e) {
    ElMessage.error('排序保存失败，已还原')
  } finally {
    emit('changed')
  }
}
useDragSort(listRef, { items, persist: persistOrder })

const drawerVisible = ref(false)
const saving = ref(false)
const form = ref({})

const openCreate = () => {
  form.value = { id: 0, time: '', title: '', desc: '', tags: [], sort: props.list.length }
  drawerVisible.value = true
}
const openEdit = (it) => {
  form.value = {
    id: it.id,
    time: it.time || '',
    title: it.title || '',
    desc: it.desc || '',
    tags: [...(it.tags || [])],
    sort: it.sort ?? 0
  }
  drawerVisible.value = true
}

const buildPayload = (f) => ({
  time: f.time || '',
  title: f.title || '',
  desc: f.desc || '',
  tags: f.tags || [],
  sort: f.sort ?? 0
})

const save = async () => {
  if (!form.value.time || !form.value.title) {
    ElMessage.warning('请填写时间范围和标题')
    return
  }
  saving.value = true
  try {
    const res = form.value.id
      ? await defaultApi.apiAdminAboutTimelineIdPut(form.value.id, buildPayload(form.value))
      : await defaultApi.apiAdminAboutTimelinePost(buildPayload(form.value))
    if (res?.code === 0) {
      ElMessage.success(form.value.id ? '更新成功' : '创建成功')
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
    await ElMessageBox.confirm(`确定删除节点「${it.title}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await defaultApi.apiAdminAboutTimelineIdDelete(it.id)
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
.timeline-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.timeline-row {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background: var(--el-bg-color);
  transition: box-shadow 0.2s ease, transform 0.2s ease;

  // 拖拽占位态（sortablejs ghost class 挂在条目本体上）
  // drag placeholder state (ghost class lands on the row itself)
  &.drag-ghost {
    opacity: 0.4;
    outline: 2px dashed var(--el-color-primary);
  }

  &:hover {
    box-shadow: var(--el-box-shadow-light);
    transform: translateX(2px);

    .row-ops {
      opacity: 1;
    }
  }

  .drag-handle {
    cursor: grab;
    color: var(--el-text-color-placeholder);
    margin-top: 2px;

    &:hover {
      color: var(--el-color-primary);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .time-col {
    flex-shrink: 0;
    width: 120px;

    .time-pill {
      display: inline-block;
      padding: 4px 10px;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 600;
      font-variant-numeric: tabular-nums;
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }

  .content-col {
    flex: 1;
    min-width: 0;

    .row-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--el-text-color-primary);
    }

    .row-desc {
      margin-top: 4px;
      font-size: 12px;
      line-height: 1.6;
      color: var(--el-text-color-regular);
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .row-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      margin-top: 8px;
    }
  }

  .row-ops {
    flex-shrink: 0;
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
}
</style>
