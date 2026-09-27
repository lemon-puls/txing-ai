<template>
  <SectionCard
    id="section-skills"
    title="核心能力"
    icon="Cpu"
    description="技能分类卡：图标 + 分类 + 标签 + 熟练度"
    :count="items.length"
  >
    <template #actions>
      <el-button round type="primary" size="small" :icon="Plus" @click="openCreate">新增技能</el-button>
    </template>

    <div ref="listRef" class="skill-grid">
      <div v-for="it in items" :key="it.id" :data-sort-id="it.id" class="skill-card">
        <span class="drag-handle" title="拖拽排序">
          <el-icon><Rank /></el-icon>
        </span>
        <div class="card-top">
          <div class="icon-box">
            <el-icon :size="20">
              <component :is="resolveIcon(it.iconKey) || 'Monitor'" />
            </el-icon>
          </div>
          <div class="skill-name" :title="it.category">{{ it.category }}</div>
        </div>
        <div class="skill-tags">
          <el-tag v-for="t in (it.tags || []).slice(0, 4)" :key="t" size="small" round effect="plain">
            {{ t }}
          </el-tag>
          <span v-if="(it.tags || []).length > 4" class="more-tag">+{{ it.tags.length - 4 }}</span>
        </div>
        <div class="skill-level">
          <el-progress
            :percentage="it.level || 0"
            :stroke-width="8"
            :show-text="false"
            class="level-bar"
          />
          <span class="level-text">{{ it.level || 0 }}%</span>
        </div>
        <div class="card-ops">
          <el-button link type="primary" :icon="EditPen" @click="openEdit(it)">编辑</el-button>
          <el-button link type="danger" :icon="Delete" @click="remove(it)">删除</el-button>
        </div>
      </div>
    </div>
    <el-empty v-if="!items.length" description="还没有技能卡" :image-size="90">
      <el-button round type="primary" :icon="Plus" @click="openCreate">新增技能</el-button>
    </el-empty>
  </SectionCard>

  <EditDrawer
    v-model="drawerVisible"
    :title="form.id ? '编辑技能' : '新增技能'"
    subtitle="前台「核心能力」卡片"
    size="520px"
    :saving="saving"
    @save="save"
  >
    <el-form :model="form" label-width="80px">
      <el-form-item label="分类名称">
        <el-input v-model="form.category" maxlength="20" placeholder="如 后端开发" />
      </el-form-item>
      <el-form-item label="图标">
        <IconSelect v-model="form.iconKey" placeholder="选择技能图标" />
      </el-form-item>
      <el-form-item label="标签">
        <TagsInput v-model="form.tags" add-text="添加技术" />
      </el-form-item>
      <el-form-item label="熟练度">
        <div class="level-editor">
          <el-slider v-model="form.level" :min="0" :max="100" show-input class="level-slider" />
        </div>
      </el-form-item>
    </el-form>
  </EditDrawer>
</template>

<script setup>
// 核心能力区块：技能卡网格 + 拖拽 + 抽屉编辑
// Skills section: skill card grid + drag sort + edit drawer
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, EditPen, Rank } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { resolveIcon } from '@/utils/iconResolver.js'
import { useDragSort } from '@/composables/useDragSort'
import SectionCard from './SectionCard.vue'
import EditDrawer from './EditDrawer.vue'
import IconSelect from './editors/IconSelect.vue'
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
    await Promise.all(changed.map((it) => defaultApi.apiAdminAboutSkillIdPut(it.id, buildPayload(it))))
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
  form.value = { id: 0, category: '', iconKey: '', tags: [], level: 80, sort: props.list.length }
  drawerVisible.value = true
}
const openEdit = (it) => {
  form.value = {
    id: it.id,
    category: it.category || '',
    iconKey: it.iconKey || '',
    tags: [...(it.tags || [])],
    level: it.level ?? 0,
    sort: it.sort ?? 0
  }
  drawerVisible.value = true
}

const buildPayload = (f) => ({
  category: f.category || '',
  iconKey: f.iconKey || '',
  tags: f.tags || [],
  level: f.level ?? 0,
  sort: f.sort ?? 0
})

const save = async () => {
  if (!form.value.category) {
    ElMessage.warning('请填写分类名称')
    return
  }
  saving.value = true
  try {
    const res = form.value.id
      ? await defaultApi.apiAdminAboutSkillIdPut(form.value.id, buildPayload(form.value))
      : await defaultApi.apiAdminAboutSkillPost(buildPayload(form.value))
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
    await ElMessageBox.confirm(`确定删除技能「${it.category}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const res = await defaultApi.apiAdminAboutSkillIdDelete(it.id)
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
.skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
}

.skill-card {
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

    .icon-box {
      width: 42px;
      height: 42px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
      color: var(--el-color-primary);
      background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-color-primary-light-8));
    }

    .skill-name {
      font-size: 14px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .skill-tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;

    .more-tag {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .skill-level {
    display: flex;
    align-items: center;
    gap: 10px;

    .level-bar {
      flex: 1;

      :deep(.el-progress-bar__outer) {
        border-radius: 999px;
      }

      :deep(.el-progress-bar__inner) {
        border-radius: 999px;
        background: linear-gradient(90deg, var(--el-color-primary), #6366f1);
      }
    }

    .level-text {
      font-size: 12px;
      font-weight: 600;
      color: var(--el-color-primary);
      min-width: 34px;
      text-align: right;
    }
  }

  .card-ops {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }
}

.level-editor {
  width: 100%;
}
</style>
