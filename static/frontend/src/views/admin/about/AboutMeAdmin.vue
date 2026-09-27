<template>
  <div ref="pageRootRef" class="about-admin">
    <!-- ===== 页头（复刻 WebsiteList page-header 范式） ===== -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-icon">
          <el-icon :size="18"><UserFilled /></el-icon>
        </div>
        <div class="header-text">
          <div class="header-title">关于我页面</div>
          <div class="header-subtitle">维护前台 /about 个人主页全部内容，保存后立即生效</div>
        </div>
      </div>
      <div class="header-actions">
        <el-tooltip content="重新加载所有分区数据" placement="top">
          <el-button circle :loading="refreshing" @click="refreshAll">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </el-tooltip>
        <el-button round @click="openFrontPage">
          <el-icon class="btn-icon"><TopRight /></el-icon>
          打开前台
        </el-button>
        <el-button round class="preview-btn" @click="previewVisible = true">
          <el-icon class="btn-icon"><View /></el-icon>
          预览前台
        </el-button>
      </div>
    </div>

    <!-- ===== 概览统计 chips + 空区块警示 ===== -->
    <el-card class="stats-card" shadow="never">
      <div class="stats-row">
        <button
          v-for="s in stats"
          :key="s.key"
          class="chip"
          :class="{ active: activeId === sectionIdOf(s.key), empty: s.count === 0 }"
          @click="scrollToSection(sectionIdOf(s.key))"
        >
          <el-icon v-if="s.count === 0" class="chip-warn-icon"><WarningFilled /></el-icon>
          {{ s.label }} · {{ s.count }}
        </button>
        <span class="flex-spacer" />
        <span class="stats-hint">点击跳转到对应分区</span>
      </div>
      <el-alert
        v-if="emptySections.length"
        class="empty-alert"
        type="warning"
        :closable="false"
        show-icon
        :title="`以下区块暂无内容，前台将不展示：${emptySections.map((s) => s.label).join('、')}`"
      />
    </el-card>

    <!-- ===== 主体：左侧锚点导航 + 右侧分区 ===== -->
    <div class="page-main">
      <!-- 窄屏时的顶部横向锚点 -->
      <div v-if="isNarrow" class="narrow-chips">
        <button
          v-for="sec in SECTIONS"
          :key="sec.id"
          class="chip"
          :class="{ active: activeId === sec.id }"
          @click="scrollToSection(sec.id)"
        >
          {{ sec.label }}
        </button>
      </div>

      <aside v-show="!isNarrow" class="anchor-nav">
        <div class="nav-title">内容分区</div>
        <button
          v-for="sec in SECTIONS"
          :key="sec.id"
          class="nav-item"
          :class="{ active: activeId === sec.id, 'nav-empty': countOf(sec.key) === 0 }"
          @click="scrollToSection(sec.id)"
        >
          <el-icon class="nav-icon"><component :is="resolveIcon(sec.icon) || 'Document'" /></el-icon>
          <span class="nav-label">{{ sec.label }}</span>
          <span v-if="countOf(sec.key) > 0" class="nav-count">{{ countOf(sec.key) }}</span>
          <el-tooltip v-else content="暂无内容" placement="right">
            <span class="nav-dot" />
          </el-tooltip>
        </button>
      </aside>

      <div class="sections">
        <HeroSection :data="hero" @changed="() => onSectionChanged('hero')" />
        <FloatingIconSection :list="lists.floating" @changed="() => onSectionChanged('floating')" />
        <ReasonSection :list="lists.reasons" @changed="() => onSectionChanged('reasons')" />
        <SkillSection :list="lists.skills" @changed="() => onSectionChanged('skills')" />
        <ProjectSection :list="lists.projects" @changed="() => onSectionChanged('projects')" />
        <TimelineSection :list="lists.timeline" @changed="() => onSectionChanged('timeline')" />
        <ContactSection :data="contact" @changed="() => onSectionChanged('contact')" />
      </div>
    </div>

    <PreviewDrawer v-model="previewVisible" :reload-token="previewToken" />
  </div>
</template>

<script setup name="AboutMeAdmin">
// 关于我页面配置：分区单页 + 锚点导航 + scrollspy + 实时预览
// About page admin: sectioned page + anchor nav + scrollspy + live preview
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, TopRight, View, UserFilled, WarningFilled } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { resolveIcon } from '@/utils/iconResolver.js'
import { SECTIONS, SECTION_IDS, previewUrl } from './constants'
import PreviewDrawer from './components/PreviewDrawer.vue'
import HeroSection from './components/HeroSection.vue'
import FloatingIconSection from './components/FloatingIconSection.vue'
import ReasonSection from './components/ReasonSection.vue'
import SkillSection from './components/SkillSection.vue'
import ProjectSection from './components/ProjectSection.vue'
import TimelineSection from './components/TimelineSection.vue'
import ContactSection from './components/ContactSection.vue'

// ==================== 数据：父组件统一加载并持有，子组件写入后 emit changed 分区 reload ====================
const lists = reactive({
  floating: [],
  reasons: [],
  skills: [],
  projects: [],
  timeline: []
})
const hero = ref(null)
const contact = ref(null)

const loaders = {
  hero: async () => {
    const res = await defaultApi.apiAdminAboutHeroGet()
    if (res?.code === 0) hero.value = res.data || null
  },
  floating: async () => {
    const res = await defaultApi.apiAdminAboutFloatingIconListGet(1, 200, {})
    if (res?.code === 0) lists.floating = res.data?.records || []
  },
  reasons: async () => {
    const res = await defaultApi.apiAdminAboutReasonListGet(1, 200, {})
    if (res?.code === 0) lists.reasons = res.data?.records || []
  },
  skills: async () => {
    const res = await defaultApi.apiAdminAboutSkillListGet(1, 200, {})
    if (res?.code === 0) lists.skills = res.data?.records || []
  },
  projects: async () => {
    const res = await defaultApi.apiAdminAboutProjectListGet(1, 200, {})
    if (res?.code === 0) lists.projects = res.data?.records || []
  },
  timeline: async () => {
    const res = await defaultApi.apiAdminAboutTimelineListGet(1, 200, {})
    if (res?.code === 0) lists.timeline = res.data?.records || []
  },
  contact: async () => {
    const res = await defaultApi.apiAdminAboutContactGet()
    if (res?.code === 0) contact.value = res.data || null
  }
}

const refreshing = ref(false)
const loadAll = async () => {
  refreshing.value = true
  await Promise.all(Object.values(loaders).map((fn) => fn().catch(() => {})))
  refreshing.value = false
}
const refreshAll = () => {
  loadAll()
  ElMessage.success('已刷新')
}

// 分区数据变更：分区 reload + 预览 iframe 刷新
// section mutated: reload that section and bump the preview token
const previewToken = ref(0)
const onSectionChanged = (key) => {
  loaders[key]?.()?.catch(() => {})
  previewToken.value++
}

// ==================== 概览统计 ====================
const stats = computed(() => [
  { key: 'floating', label: '浮动图标', count: lists.floating.length },
  { key: 'reasons', label: '选择理由', count: lists.reasons.length },
  { key: 'skills', label: '核心能力', count: lists.skills.length },
  { key: 'projects', label: '精选作品', count: lists.projects.length },
  { key: 'timeline', label: '成长轨迹', count: lists.timeline.length },
  { key: 'contact', label: '联系链接', count: (contact.value?.links || []).length }
])
const emptySections = computed(() => stats.value.filter((s) => s.count === 0))

const sectionIdOf = (key) => SECTIONS.find((s) => s.key === key)?.id || ''
// hero 是单例区块（不在 stats chips 中）：有数据即为 1，避免导航上误显示「暂无内容」警示点
// hero is a singleton (not in stats chips): treat presence as 1 so the nav dot doesn't misfire
const countOf = (key) =>
  key === 'hero' ? (hero.value ? 1 : 0) : (stats.value.find((s) => s.key === key)?.count ?? null)

// ==================== 锚点导航 + scrollspy ====================
// 注意：滚动容器是 el-main（AdminLayout .main），不是 window —— IntersectionObserver 必须指定 root
// the scroll container is el-main, not the window — pass it as the observer root
const pageRootRef = ref(null)
const activeId = ref(SECTION_IDS[0])
const navLockUntil = ref(0)
const visibleMap = new Map()
let observer = null
let resizeObserver = null
let scrollRoot = null

const isNarrow = ref(false)

onMounted(() => {
  scrollRoot = pageRootRef.value?.closest('.main') || null
  const root = scrollRoot
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((e) => visibleMap.set(e.target.id, e.isIntersecting))
      // 点击导航后的锁定窗口内不更新高亮，避免与平滑滚动竞争
      // suppress highlight updates right after a nav click (smooth-scroll race)
      if (Date.now() < navLockUntil.value) return
      // 滚动容器触底时最后一个分区（联系区）可能进不了观察窗口，强制高亮末项
      // when scrolled to the bottom the last section can't enter the observation window — force it
      if (scrollRoot && scrollRoot.scrollHeight - scrollRoot.scrollTop - scrollRoot.clientHeight < 4) {
        activeId.value = SECTION_IDS[SECTION_IDS.length - 1]
        return
      }
      const first = SECTION_IDS.find((id) => visibleMap.get(id))
      if (first) activeId.value = first
    },
    { root, rootMargin: '-88px 0px -55% 0px', threshold: [0, 0.05] }
  )
  SECTION_IDS.forEach((id) => {
    const el = document.getElementById(id)
    if (el) observer.observe(el)
  })

  // 窄屏判断监听页面根宽（侧边栏折叠不触发 window resize，media query 感知不到）
  // watch the page root width: sidebar collapse doesn't fire window resize
  if (pageRootRef.value) {
    resizeObserver = new ResizeObserver(([entry]) => {
      isNarrow.value = entry.contentRect.width < 1200
    })
    resizeObserver.observe(pageRootRef.value)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
  resizeObserver?.disconnect()
})

const scrollToSection = (id) => {
  if (!id) return
  activeId.value = id
  navLockUntil.value = Date.now() + 700
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// ==================== 预览 ====================
const previewVisible = ref(false)
const openFrontPage = () => window.open(previewUrl, '_blank')

onMounted(loadAll)
</script>

<style lang="scss" scoped>
.about-admin {
  padding: 0;
}

// ===== 页头 =====
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-icon {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
    box-shadow: 0 4px 12px var(--el-color-primary-light-8);
    flex-shrink: 0;
  }

  .header-title {
    font-size: 17px;
    font-weight: 700;
    line-height: 1.3;
    // 显式主题变量：admin 全局给 main 继承的硬编码深色在暗色下不可读
    // explicit token: the hardcoded inherited color is unreadable in dark mode
    color: var(--el-text-color-primary);
  }

  .header-subtitle {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 2px;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 4px;

    :deep(.el-button) {
      padding: 10px 20px;

      .btn-icon {
        margin-right: 4px;
      }
    }
  }
}

// 预览前台：渐变 CTA（与 WebsiteList ai-btn 同款）
// gradient CTA for the preview action
.preview-btn {
  border: none !important;
  color: #fff !important;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary)) !important;
  box-shadow: 0 2px 10px var(--el-color-primary-light-7);

  &:hover {
    opacity: 0.88;
  }
}

// ===== 概览统计卡 =====
.stats-card {
  margin-bottom: 16px;
  border-radius: 14px;

  :deep(.el-card__body) {
    padding: 12px 16px;
  }

  .stats-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;

    .flex-spacer {
      flex: 1;
    }

    .stats-hint {
      font-size: 12px;
      color: var(--el-text-color-placeholder);
    }
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 14px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border: 1px solid transparent;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s ease;

    .chip-warn-icon {
      color: var(--el-color-warning);
    }

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    &.active {
      color: #fff;
      background: var(--el-color-primary);
      border-color: var(--el-color-primary);

      .chip-warn-icon {
        color: rgb(255 255 255 / 85%);
      }
    }

    // 空区块警示态
    // warning state for empty sections
    &.empty {
      color: var(--el-color-warning);
      background: var(--el-color-warning-light-9);

      &:hover {
        background: var(--el-color-warning-light-8);
      }

      &.active {
        color: #fff;
        background: var(--el-color-warning);
        border-color: var(--el-color-warning);
      }
    }
  }

  .empty-alert {
    margin-top: 10px;
    border-radius: 8px;
  }
}

// ===== 主体布局 =====
// 注意：sticky 祖先链上禁止任何 overflow，否则左侧导航粘性失效
// never set overflow on sticky ancestors or the anchor nav loses stickiness
.page-main {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.anchor-nav {
  position: sticky;
  top: 16px;
  width: 176px;
  flex-shrink: 0;
  padding: 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  background: var(--el-bg-color);
  box-shadow: var(--el-box-shadow-light);

  .nav-title {
    padding: 6px 10px 10px;
    font-size: 12px;
    font-weight: 600;
    color: var(--el-text-color-secondary);
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 9px 10px;
    margin-bottom: 2px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    background: transparent;
    border: none;
    border-radius: 10px;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s ease;

    .nav-icon {
      color: var(--el-text-color-secondary);
      flex-shrink: 0;
    }

    .nav-label {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .nav-count {
      min-width: 20px;
      padding: 0 6px;
      border-radius: 999px;
      font-size: 11px;
      line-height: 18px;
      text-align: center;
      color: var(--el-text-color-secondary);
      background: var(--el-fill-color);
    }

    .nav-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: var(--el-color-warning);
    }

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);

      .nav-icon {
        color: var(--el-color-primary);
      }
    }

    &.active {
      color: var(--el-color-primary);
      font-weight: 600;
      background: var(--el-color-primary-light-9);
      box-shadow: inset 3px 0 0 var(--el-color-primary);

      .nav-icon {
        color: var(--el-color-primary);
      }

      .nav-count {
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-8);
      }
    }
  }
}

// 窄屏顶部横向锚点
.narrow-chips {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  margin-bottom: 4px;
  padding-bottom: 10px;
  overflow-x: auto;
  scrollbar-width: thin;

  .chip {
    flex-shrink: 0;
    padding: 4px 14px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border: 1px solid transparent;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    &.active {
      color: #fff;
      background: var(--el-color-primary);
      border-color: var(--el-color-primary);
    }
  }
}

.sections {
  flex: 1;
  min-width: 0;
  // 分区间距：各区块组件的兄弟位置夹着抽屉的 el-overlay 占位节点，
  // 相邻选择器(.about-section + .about-section)永不匹配 → 用 flex gap 保证间隔；
  // 隐藏(display:none)与 fixed 定位的 overlay 不产生 flex 盒子，gap 不受抽屉开关影响
  // flex gap: drawer overlays interleave between section siblings and break the
  // adjacent-sibling margin; hidden/fixed overlays generate no flex boxes
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
