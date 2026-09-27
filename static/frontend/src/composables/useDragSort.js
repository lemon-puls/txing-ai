import Sortable from 'sortablejs'
import { onMounted, onBeforeUnmount, watch } from 'vue'

/**
 * 为无 id 的列表项（如 media 数组）生成稳定排序标识
 * Stable sort id for list items without an id field (e.g. media arrays)
 */
const uidMap = new WeakMap()
let uidSeq = 0
export const uidOf = (obj) => {
  if (!uidMap.has(obj)) uidMap.set(obj, ++uidSeq)
  return uidMap.get(obj)
}

/**
 * 列表拖拽排序 composable（sortablejs 封装）
 * List drag-sort composable (sortablejs wrapper)
 *
 * 约定：容器直接子元素带 data-sort-id 属性、拖拽手柄带 .drag-handle class。
 * Conventions: direct children carry data-sort-id; drag handles carry .drag-handle.
 *
 * @param {Ref<HTMLElement|null>} containerRef 列表容器
 * @param {Object} opts
 * @param {Ref|ComputedRef} opts.items  列表数组（可写 ref 或 computed get/set，onEnd 通过
 *                                      整体赋值触发 setter/emit，避免双 watcher 死循环）
 * @param {Function}  [opts.getId]      (item) => string|number，默认 item.id；无 id 列表传 uidOf
 * @param {Boolean}   [opts.renumber]   是否回写 sort=index（实体列表 true；表单内调序 false）
 * @param {Function}  [opts.persist]    (next, changed) => Promise；renumber 时 changed 为 sort
 *                                      变化的条目，否则为整个新数组。不传则纯本地重排。
 * @param {Object}    [opts.sortableOptions] 透传给 Sortable.create 的额外配置
 */
export function useDragSort(containerRef, { items, getId = (it) => it.id, renumber = true, persist, sortableOptions = {} }) {
  let sortable = null

  const rebuild = (el) => {
    sortable?.destroy()
    sortable = null
    if (!el) return
    sortable = Sortable.create(el, {
      animation: 180,
      handle: '.drag-handle',
      ghostClass: 'drag-ghost',
      chosenClass: 'drag-chosen',
      // 长列表拖到容器边缘时在滚动容器内自动滚动
      // auto-scroll near container edges (works inside the el-main scroll container)
      scroll: true,
      scrollSensitivity: 90,
      scrollSpeed: 14,
      // 表单控件不作为拖拽起点，避免与输入交互冲突
      // form controls never start a drag
      filter: 'input, textarea, select, .el-slider',
      preventOnFilter: false,
      onEnd(evt) {
        if (evt.oldIndex === evt.newIndex) return
        // 按 DOM 顺序收集 id 重排数组，不依赖 newIndex（规避嵌套/占位元素错位）
        // Reorder by DOM order via data-sort-id instead of trusting newIndex
        const orderedIds = Array.from(evt.from.querySelectorAll('[data-sort-id]')).map(
          (node) => node.dataset.sortId
        )
        const byId = new Map(items.value.map((it) => [String(getId(it)), it]))
        const prevSort = new Map(items.value.map((it, i) => [String(getId(it)), it.sort ?? i]))
        const next = orderedIds.map((id) => byId.get(String(id))).filter(Boolean)
        if (renumber) {
          // 重新编号 0..n-1（与后端 sort 语义一致）
          // renumber 0..n-1 to match backend sort semantics
          next.forEach((it, i) => {
            it.sort = i
          })
        }
        // 整体赋值：走 computed setter/emit 更新父级，保持单向数据流
        // assign wholesale so a computed setter / emit updates the parent
        items.value = next
        if (!persist) return
        const changed = renumber
          ? next.filter((it) => prevSort.get(String(getId(it))) !== it.sort)
          : next
        if (changed.length) {
          Promise.resolve(persist(next, changed)).catch(() => {})
        }
      },
      ...sortableOptions
    })
  }

  onMounted(() => rebuild(containerRef.value))
  // 容器被 v-if 重建时重建实例
  // rebuild when the container element is replaced (e.g. v-if)
  watch(containerRef, (el) => rebuild(el))
  onBeforeUnmount(() => {
    sortable?.destroy()
    sortable = null
  })
}
