<template>
  <SectionCard
    id="section-floating"
    title="浮动图标"
    icon="MagicStick"
    description="Hero 区漂浮的技术 emoji 图标（最多展示 6 个）"
    :count="items.length"
  >
    <template #actions>
      <el-button round type="primary" size="small" :icon="Plus" @click="openCreate">新增图标</el-button>
    </template>

    <div ref="listRef" class="floating-grid">
      <div v-for="it in items" :key="it.id" :data-sort-id="it.id" class="floating-card">
        <span class="drag-handle" title="拖拽排序">
          <el-icon><Rank /></el-icon>
        </span>
        <div class="emoji-box">{{ it.symbol || '?' }}</div>
        <div class="icon-name" :title="it.name">{{ it.name }}</div>
        <div class="card-ops">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(it)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="remove(it)">删除</el-button>
        </div>
      </div>
    </div>
    <el-empty v-if="!items.length" description="还没有浮动图标" :image-size="90">
      <el-button round type="primary" :icon="Plus" @click="openCreate">新增图标</el-button>
    </el-empty>
  </SectionCard>

  <EditDrawer
    v-model="drawerVisible"
    :title="form.id ? '编辑浮动图标' : '新增浮动图标'"
    subtitle="emoji 会漂浮在前台 Hero 首屏"
    size="420px"
    :saving="saving"
    @save="save"
  >
    <el-form :model="form" label-width="80px">
      <el-form-item label="名称">
        <el-input v-model="form.name" placeholder="如 Go / Vue / 云原生" maxlength="20" />
      </el-form-item>
      <el-form-item label="符号">
        <div class="symbol-editor">
          <el-input v-model="form.symbol" placeholder="emoji 或字符" maxlength="4" class="symbol-input" />
          <span class="symbol-preview">{{ form.symbol || '?' }}</span>
        </div>
      </el-form-item>
    </el-form>
  </EditDrawer>
</template>

<script setup>
// 浮动图标区块：卡片网格 + 拖拽排序 + 抽屉编辑（列表区块标准模板）
// Floating icons section: card grid + drag sort + edit drawer (template for list sections)
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, EditPen, Rank } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { useDragSort } from '@/composables/useDragSort'
import SectionCard from './SectionCard.vue'
import EditDrawer from './EditDrawer.vue'

const props = defineProps({
  list: { type: Array, default: () => [] }
})
const emit = defineEmits(['changed'])

// 本地副本：父组件 reload 整体替换 = 排序保存失败时自动还原
// local copy; wholesale parent reload on failure restores order automatically
const items = ref([])
watch(
  () => props.list,
  (v) => {
    items.value = (v || []).map((x) => ({ ...x }))
  },
  { immediate: true }
)

// ===== 拖拽排序：只 PUT 变更条目，成败都 emit changed（成功取干净编号，失败还原旧序） =====
const listRef = ref(null)
const persistOrder = async (changed) => {
  try {
    await Promise.all(
      changed.map((it) => defaultApi.apiAdminAboutFloatingIconIdPut(it.id, buildPayload(it)))
    )
    ElMessage.success('排序已保存')
  } catch (e) {
    ElMessage.error('排序保存失败，已还原')
  } finally {
    emit('changed')
  }
}
useDragSort(listRef, { items, persist: persistOrder })

// ===== 新增/编辑抽屉 =====
const drawerVisible = ref(false)
const saving = ref(false)
const form = ref({ id: 0, name: '', symbol: '' })

const openCreate = () => {
  form.value = { id: 0, name: '', symbol: '', sort: props.list.length }
  drawerVisible.value = true
}
const openEdit = (it) => {
  form.value = { id: it.id, name: it.name || '', symbol: it.symbol || '', sort: it.sort ?? 0 }
  drawerVisible.value = true
}

const buildPayload = (f) => ({ name: f.name || '', symbol: f.symbol || '', sort: f.sort ?? 0 })

const save = async () => {
  if (!form.value.name || !form.value.symbol) {
    ElMessage.warning('请填写名称和符号')
    return
  }
  saving.value = true
  try {
    const res = form.value.id
      ? await defaultApi.apiAdminAboutFloatingIconIdPut(form.value.id, buildPayload(form.value))
      : await defaultApi.apiAdminAboutFloatingIconPost(buildPayload(form.value))
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
    await ElMessageBox.confirm(`确定删除浮动图标「${it.name}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await defaultApi.apiAdminAboutFloatingIconIdDelete(it.id)
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
.floating-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
}

.floating-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 18px 12px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background: var(--el-bg-color);
  transition: box-shadow 0.2s ease, transform 0.2s ease;

  // 拖拽占位态（sortablejs ghost class 挂在卡片本体上）
  // drag placeholder state (sortablejs ghost class lands on the card itself)
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
    top: 6px;
    left: 6px;
    cursor: grab;
    color: var(--el-text-color-placeholder);

    &:hover {
      color: var(--el-color-primary);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .emoji-box {
    width: 56px;
    height: 56px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    background: var(--el-fill-color-light);
  }

  .icon-name {
    max-width: 100%;
    font-size: 13px;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .card-ops {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
}

.symbol-editor {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;

  .symbol-input {
    width: 140px;
  }

  .symbol-preview {
    width: 52px;
    height: 52px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 26px;
    background: var(--el-fill-color-light);
  }
}
</style>
