<template>
  <SectionCard
    id="section-contact"
    title="联系区"
    icon="ChatDotRound"
    :description="`页面底部联系卡：标题、描述与外链按钮（共 ${form.links.length} 条链接）`"
    :count="null"
  >
    <div v-loading="saving" class="contact-form-wrap">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="form.title" maxlength="30" placeholder="如 一起构建有趣的东西" class="contact-input" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.desc" type="textarea" :rows="2" maxlength="120" show-word-limit placeholder="联系区引导文案" class="contact-input" />
        </el-form-item>
        <el-form-item label="链接">
          <div class="links-editor">
            <div ref="linksRef" class="links-list">
              <div v-for="(link, idx) in form.links" :key="uidOf(link)" :data-sort-id="uidOf(link)" class="link-row">
                <span class="drag-handle" title="拖拽排序（保存后生效）">
                  <el-icon><Rank /></el-icon>
                </span>
                <IconSelect v-model="link.iconKey" placeholder="图标" class="link-icon" />
                <el-input v-model="link.label" placeholder="名称，如 GitHub" class="link-label" />
                <el-input v-model="link.url" placeholder="https:// 或 mailto:" class="link-url" />
                <el-button text type="danger" :icon="Delete" @click="form.links.splice(idx, 1)" />
              </div>
            </div>
            <el-button round size="small" :icon="Plus" @click="addLink">添加链接</el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button round type="primary" :loading="saving" @click="save">保存</el-button>
          <span class="form-tip">链接顺序拖拽调整，点击保存后生效并同步前台</span>
        </el-form-item>
      </el-form>
    </div>
  </SectionCard>
</template>

<script setup>
// 联系区单例区块：表单 + links 编辑（组内拖拽为本地重排，保存才生效，与列表区块的「拖完即存」不同）
// Contact singleton section; in-form link drag is local-only and persists on save
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, Rank } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { useDragSort, uidOf } from '@/composables/useDragSort'
import SectionCard from './SectionCard.vue'
import IconSelect from './editors/IconSelect.vue'

const props = defineProps({
  data: { type: Object, default: null }
})
const emit = defineEmits(['changed'])

const form = reactive({ title: '', desc: '', links: [] })
const saving = ref(false)

watch(
  () => props.data,
  (v) => {
    if (!v) return
    form.title = v.title || ''
    form.desc = v.desc || ''
    form.links = (v.links || []).map((l) => ({ ...l }))
  },
  { immediate: true }
)

// links 组内拖拽：仅本地重排（get/set 计算属性直连 form.links）
// in-form drag: local reorder through a get/set computed
const linksRef = ref(null)
const linkItems = computed({
  get: () => form.links,
  set: (v) => {
    form.links = v
  }
})
useDragSort(linksRef, { items: linkItems, getId: uidOf, renumber: false })

const addLink = () => {
  form.links.push({ iconKey: '', label: '', url: '' })
}

const save = async () => {
  saving.value = true
  try {
    const payload = {
      title: form.title,
      desc: form.desc,
      links: form.links.map((l) => ({ iconKey: l.iconKey || '', label: l.label || '', url: l.url || '' }))
    }
    const res = await defaultApi.apiAdminAboutContactPut(payload)
    if (res?.code === 0) {
      ElMessage.success('保存成功')
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
</script>

<style lang="scss" scoped>
.contact-form-wrap {
  max-width: 760px;
}

.contact-input {
  max-width: 560px;
}

.links-editor {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  width: 100%;

  .links-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .link-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 8px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    transition: box-shadow 0.2s ease, opacity 0.2s ease;

    // 拖拽占位态（sortablejs ghost class 挂在行本体上）
    // drag placeholder state (ghost class lands on the row itself)
    &.drag-ghost {
      opacity: 0.4;
      outline: 2px dashed var(--el-color-primary);
    }

    .drag-handle {
      cursor: grab;
      color: var(--el-text-color-placeholder);

      &:hover {
        color: var(--el-color-primary);
      }

      &:active {
        cursor: grabbing;
      }
    }

    .link-icon {
      width: 150px;
      flex-shrink: 0;
    }

    .link-label {
      width: 160px;
      flex-shrink: 0;
    }

    .link-url {
      flex: 1;
    }
  }
}

.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
