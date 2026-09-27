<template>
  <SectionCard
    id="section-reasons"
    title="为什么选择我"
    icon="Stamp"
    description="亮点卡片：emoji + 标题 + 描述 + 标签 + 统计数据"
    :count="items.length"
  >
    <template #actions>
      <el-button round type="primary" size="small" :icon="Plus" @click="openCreate">新增卡片</el-button>
    </template>

    <div ref="listRef" class="reason-grid">
      <div v-for="it in items" :key="it.id" :data-sort-id="it.id" class="reason-card">
        <span class="drag-handle" title="拖拽排序">
          <el-icon><Rank /></el-icon>
        </span>
        <div class="card-top">
          <div class="emoji-circle">{{ it.emoji || '⭐' }}</div>
          <div class="card-title" :title="it.title">{{ it.title }}</div>
        </div>
        <div class="card-desc" :title="it.desc">{{ it.desc }}</div>
        <div class="card-tags">
          <el-tag v-for="t in (it.tags || []).slice(0, 3)" :key="t" size="small" round effect="light">
            {{ t }}
          </el-tag>
          <span v-if="(it.tags || []).length > 3" class="more-tag">+{{ it.tags.length - 3 }}</span>
        </div>
        <div v-if="(it.stats || []).length" class="card-stats">
          <div v-for="(s, i) in it.stats" :key="i" class="stat-item">
            <span class="stat-value">{{ s.value }}</span>
            <span class="stat-label">{{ s.label }}</span>
          </div>
        </div>
        <div class="card-ops">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(it)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="remove(it)">删除</el-button>
        </div>
      </div>
    </div>
    <el-empty v-if="!items.length" description="还没有亮点卡片" :image-size="90">
      <el-button round type="primary" :icon="Plus" @click="openCreate">新增卡片</el-button>
    </el-empty>
  </SectionCard>

  <EditDrawer
    v-model="drawerVisible"
    :title="form.id ? '编辑卡片' : '新增卡片'"
    subtitle="前台「为什么选择我」三列卡片"
    size="560px"
    :saving="saving"
    @save="save"
  >
    <el-form :model="form" label-width="80px">
      <el-form-item label="Emoji">
        <el-input v-model="form.emoji" maxlength="4" placeholder="如 🚀" class="emoji-input" />
      </el-form-item>
      <el-form-item label="标题">
        <el-input v-model="form.title" maxlength="20" placeholder="如 全栈交付能力" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.desc" type="textarea" :rows="3" maxlength="120" show-word-limit placeholder="卡片正文描述" />
      </el-form-item>
      <el-form-item label="标签">
        <TagsInput v-model="form.tags" />
      </el-form-item>
      <el-form-item label="统计数据">
        <StatsEditor v-model="form.stats" />
      </el-form-item>
    </el-form>
  </EditDrawer>
</template>

<script setup>
// 「为什么选择我」区块：亮点卡片网格 + 拖拽 + 抽屉编辑
// Reasons section: highlight card grid + drag sort + edit drawer
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, EditPen, Rank } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { useDragSort } from '@/composables/useDragSort'
import SectionCard from './SectionCard.vue'
import EditDrawer from './EditDrawer.vue'
import TagsInput from './editors/TagsInput.vue'
import StatsEditor from './editors/StatsEditor.vue'

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
    await Promise.all(changed.map((it) => defaultApi.apiAdminAboutReasonIdPut(it.id, buildPayload(it))))
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
  form.value = { id: 0, emoji: '', title: '', desc: '', tags: [], stats: [], sort: props.list.length }
  drawerVisible.value = true
}
const openEdit = (it) => {
  form.value = {
    id: it.id,
    emoji: it.emoji || '',
    title: it.title || '',
    desc: it.desc || '',
    tags: [...(it.tags || [])],
    stats: (it.stats || []).map((s) => ({ ...s })),
    sort: it.sort ?? 0
  }
  drawerVisible.value = true
}

const buildPayload = (f) => ({
  emoji: f.emoji || '',
  title: f.title || '',
  desc: f.desc || '',
  tags: f.tags || [],
  stats: f.stats || [],
  sort: f.sort ?? 0
})

const save = async () => {
  if (!form.value.title) {
    ElMessage.warning('请填写标题')
    return
  }
  saving.value = true
  try {
    const res = form.value.id
      ? await defaultApi.apiAdminAboutReasonIdPut(form.value.id, buildPayload(form.value))
      : await defaultApi.apiAdminAboutReasonPost(buildPayload(form.value))
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
    await ElMessageBox.confirm(`确定删除卡片「${it.title}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await defaultApi.apiAdminAboutReasonIdDelete(it.id)
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
.reason-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.reason-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px 14px 12px;
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
    top: 8px;
    right: 8px;
    cursor: grab;
    color: var(--el-text-color-placeholder);

    &:hover {
      color: var(--el-color-primary);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .card-top {
    display: flex;
    align-items: center;
    gap: 10px;

    .emoji-circle {
      width: 44px;
      height: 44px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 22px;
      flex-shrink: 0;
      background: var(--el-color-primary-light-9);
    }

    .card-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .card-desc {
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-regular);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 38px;
  }

  .card-tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;

    .more-tag {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .card-stats {
    display: flex;
    gap: 14px;
    padding-top: 8px;
    border-top: 1px dashed var(--el-border-color-lighter);

    .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;

      .stat-value {
        font-size: 16px;
        font-weight: 700;
        background: linear-gradient(135deg, var(--el-color-primary), #6366f1);
        -webkit-background-clip: text;
        background-clip: text;
        -webkit-text-fill-color: transparent;
      }

      .stat-label {
        font-size: 11px;
        color: var(--el-text-color-secondary);
      }
    }
  }

  .card-ops {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
}

.emoji-input {
  width: 120px;
}
</style>
