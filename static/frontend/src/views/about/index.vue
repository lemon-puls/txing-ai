<template>
  <div class="about-container">
    <!-- 3D 星云星空背景（Three.js 粒子云） -->
    <!-- 3D nebula starfield background (Three.js particle cloud) -->
    <div ref="starfieldRef" class="starfield-bg"></div>

    <!-- Hero Section -->
    <section class="hero-section">
      <!-- 云海天幕：Hero 顶部天空，向下渐隐过渡到星空（滚动即"穿云入星空"） -->
      <!-- Cloud-sky dome: hero sky fading down into the starfield ("through the clouds into space" on scroll) -->
      <CloudSky class="hero-sky" />
      <div class="hero-content animate-on-scroll">
        <div class="avatar-wrapper">
          <div class="main-avatar">
            <span class="avatar-text">{{ heroData.avatarText }}</span>
            <div class="avatar-ring"></div>
            <div class="avatar-ring ring-2"></div>
          </div>
          <div class="status-badge">
            <span class="status-dot"></span>
            {{ heroData.statusText }}
          </div>
        </div>
        <h1 class="hero-title">
          你好，我是
          <!-- FlickerText 名字特效：入场霓虹灯管点亮，悬停再闪三下 -->
          <!-- FlickerText name effect: neon-tube warm-up on enter, three blinks on hover -->
          <FlickerText
            :text="heroData.name"
            tag="span"
            class="flicker-name"
            color-mode="gradient"
            gradient-start="var(--el-color-primary)"
            gradient-end="#a855f7"
            :gradient-angle="45"
            :flicker="nameEnterFlicker"
            :flicker-hover="nameHoverFlicker"
          />
        </h1>
        <p class="hero-subtitle animate-on-scroll delay-1">
          {{ displayedSubtitle }}
        </p>
        <div class="hero-actions animate-on-scroll delay-2">
          <el-button type="primary" size="large" round @click="scrollToProjects" class="action-btn">
            <span class="btn-content">
              查看我的作品
              <el-icon class="btn-arrow"><ArrowDown /></el-icon>
            </span>
          </el-button>
          <el-button size="large" round @click="scrollToContact" class="action-btn secondary">
            联系我
          </el-button>
        </div>
        <div class="scroll-indicator animate-on-scroll delay-3">
          <div class="mouse">
            <div class="wheel"></div>
          </div>
          <span>向下滚动探索</span>
        </div>
      </div>
      <div class="hero-bg-blobs">
        <div class="blob blob-1"></div>
        <div class="blob blob-2"></div>
        <div class="blob blob-3"></div>
      </div>
      <!-- Floating tech icons -->
      <div class="floating-icons">
        <div v-for="(icon, index) in floatingIcons" :key="icon.name"
             class="floating-icon"
             :class="`icon-${index}`"
             :style="{ animationDelay: `${index * 0.5}s` }">
          <span>{{ icon.symbol }}</span>
        </div>
      </div>
    </section>

    <!-- Why Choose Me Section -->
    <section class="section why-me-section">
      <div class="section-header animate-on-scroll">
        <h2 class="section-title">为什么选择我？</h2>
        <div class="title-underline"></div>
        <p class="section-subtitle">专注 · 热情 · 追求极致</p>
      </div>
      <div class="why-me-grid">
        <div v-for="(reason, index) in whyChooseMe" :key="reason.title"
             class="why-me-card animate-on-scroll"
             :style="{ animationDelay: `${index * 0.15}s` }">
          <div class="card-glow"></div>
          <div class="card-icon-wrapper">
            <div class="icon-bg"></div>
            <span class="card-emoji">{{ reason.emoji }}</span>
          </div>
          <h3 class="card-title">{{ reason.title }}</h3>
          <p class="card-desc">{{ reason.desc }}</p>
          <div class="card-stats" v-if="reason.stats">
            <div v-for="stat in reason.stats" :key="stat.label" class="stat-item">
              <span class="stat-number">{{ stat.value }}</span>
              <span class="stat-label">{{ stat.label }}</span>
            </div>
          </div>
          <div class="card-tags">
            <span v-for="tag in reason.tags" :key="tag" class="tag">{{ tag }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Skills Section -->
    <section class="section skills-section">
      <div class="section-header animate-on-scroll">
        <h2 class="section-title">核心能力</h2>
        <div class="title-underline"></div>
      </div>
      <div class="skills-grid">
        <div v-for="(skill, index) in skillSets" :key="skill.category"
             class="skill-card animate-on-scroll"
             :style="{ animationDelay: `${index * 0.1}s` }">
          <div class="skill-icon">
            <el-icon><component :is="skill.icon" /></el-icon>
          </div>
          <h3 class="skill-title">{{ skill.category }}</h3>
          <div class="skill-tags">
            <span v-for="tag in skill.tags" :key="tag" class="skill-tag">{{ tag }}</span>
          </div>
          <div class="skill-progress" v-if="skill.level">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: skill.level + '%' }"></div>
            </div>
            <span class="progress-text">{{ skill.level }}%</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Projects Section -->
    <section id="projects" class="section projects-section">
      <div class="section-header animate-on-scroll">
        <h2 class="section-title">精选作品</h2>
        <div class="title-underline"></div>
        <p class="section-subtitle">点击卡片展开查看详情</p>
      </div>

      <!-- 项目类别切换器：全部 / 公司项目 / 个人项目 -->
      <!-- Project category switcher: all / company / personal -->
      <div class="project-filter animate-on-scroll" role="tablist" aria-label="项目类别切换">
        <div class="filter-track">
          <div class="filter-thumb" :style="{ transform: `translateX(${filterIndex * 100}%)` }"></div>
          <button
            v-for="opt in filterOptions"
            :key="opt.value"
            class="filter-btn"
            :class="{ active: projectFilter === opt.value }"
            role="tab"
            :aria-selected="projectFilter === opt.value"
            @click="switchFilter(opt.value)"
          >
            <span class="filter-count">{{ countByCategory(opt.value) }}</span>
            {{ opt.label }}
          </button>
        </div>
      </div>

      <div class="projects-grid">

        <div v-for="project in filteredProjects" :key="project.id"
             class="project-card"
             :class="{ 'expanded': expandedProject === project.id }">
          <div class="project-image" :class="project.gradient" @click="toggleProject(project.id)">
            <!-- 封面轮播：图片/视频混合，未配置时回退渐变+图标 -->
            <!-- Cover carousel: mixed image/video, falls back to gradient + icon when empty -->
            <el-carousel
              v-if="project.coverMedia && project.coverMedia.length"
              class="cover-carousel"
              :interval="5000"
              arrow="hover"
              :indicator-position="'none'"
            >
              <el-carousel-item v-for="(m, mIdx) in project.coverMedia" :key="mIdx">
                <!-- contain 完整展示 + 模糊垫底：任何比例都不裁切，留边由图片自身的模糊放大填充 -->
                <!-- Contain shows the whole media; a blurred oversized copy fills the letterbox bars -->
                <div class="cover-media-wrap">
                  <template v-if="m.type === 'image'">
                    <img :src="m.url" class="cover-media-blur" alt="" aria-hidden="true" />
                    <img :src="m.url" class="cover-media" alt="" />
                  </template>
                  <video
                    v-else
                    :src="m.url"
                    class="cover-media"
                    muted
                    loop
                    preload="metadata"
                    playsinline
                    @click.stop
                    @mouseenter="$event.target.play()"
                    @mouseleave="$event.target.pause()"
                  />
                </div>
              </el-carousel-item>
            </el-carousel>
            <div v-else class="project-icon">
              <el-icon><component :is="project.icon" /></el-icon>
            </div>
            <div class="project-badge">{{ project.badge }}</div>
            <div class="expand-hint">
              <el-icon class="expand-icon" :class="{ 'rotated': expandedProject === project.id }">
                <ArrowDown />
              </el-icon>
            </div>
          </div>
          <div class="project-info">
            <div class="project-tags">
              <span v-for="tag in project.tags" :key="tag" class="project-tag">{{ tag }}</span>
            </div>
            <h3 class="project-name" @click="toggleProject(project.id)">{{ project.name }}</h3>
            <p class="project-desc">{{ project.desc }}</p>
            <div class="project-highlights">
              <div v-for="hl in project.highlights" :key="hl" class="highlight-item">
                <el-icon><CircleCheck /></el-icon>
                <span>{{ hl }}</span>
              </div>
            </div>
            <a v-if="project.link" :href="project.link" target="_blank" class="project-link">
              <span>访问项目</span>
              <el-icon><ArrowRight /></el-icon>
            </a>
          </div>

          <!-- Expanded Detail Section -->
          <transition name="slide-fade">
            <div v-if="expandedProject === project.id" class="project-detail">
              <div class="detail-header">
                <h4>项目详情</h4>
                <el-button text @click="expandedProject = null">
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>

              <!-- Media Gallery -->
              <div class="media-gallery" v-if="project.media && project.media.length">
                <div class="gallery-scroll">
                  <div v-for="(media, mIndex) in project.media" :key="mIndex" class="media-item">
                    <img v-if="media.type === 'image'" :src="media.url" :alt="media.caption" @click="openMediaPreview(media)" />
                    <video v-else-if="media.type === 'video'" :src="media.url" muted loop @mouseenter="$event.target.play()" @mouseleave="$event.target.pause()" />
                    <div class="media-caption">{{ media.caption }}</div>
                  </div>
                </div>
              </div>

              <!-- Tech Stack -->
              <div class="tech-stack" v-if="project.techStack">
                <h4>技术栈</h4>
                <div class="tech-items">
                  <div v-for="tech in project.techStack" :key="tech.name" class="tech-item">
                    <span class="tech-icon">{{ tech.icon }}</span>
                    <span class="tech-name">{{ tech.name }}</span>
                  </div>
                </div>
              </div>

              <!-- 工作亮点（features 字段暂承载亮点内容，后续如需独立的核心功能/工作职责栏再说） -->
              <!-- Work highlights (the features field carries highlight content for now) -->
              <!-- 点击展开详细设计：detail 按换行拆段，0fr→1fr 网格动画 -->
              <!-- Click to expand design details: detail split by newlines, 0fr→1fr grid animation -->
              <div class="key-features" v-if="project.features">
                <h4>工作亮点</h4>
                <div class="features-grid">
                  <div
                    v-for="feature in project.features"
                    :key="feature.title"
                    class="feature-item"
                    :class="{ 'has-detail': !!feature.detail, 'detail-open': openFeature === feature.title }"
                    @click="feature.detail && toggleFeature(feature.title)"
                  >
                    <span class="feature-icon">{{ feature.icon }}</span>
                    <div class="feature-body">
                      <div class="feature-head">
                        <h5>{{ feature.title }}</h5>
                        <el-icon v-if="feature.detail" class="feature-chevron"><ArrowDown /></el-icon>
                      </div>
                      <p>{{ feature.desc }}</p>
                      <div v-if="feature.detail" class="feature-detail-wrap">
                        <div class="feature-detail-clip">
                          <div class="feature-detail">
                            <p v-for="(para, pi) in splitDetail(feature.detail)" :key="pi">{{ para }}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </section>

    <!-- Timeline Section -->
    <section class="section timeline-section">
      <div class="section-header animate-on-scroll">
        <h2 class="section-title">成长轨迹</h2>
        <div class="title-underline"></div>
      </div>
      <div class="timeline">
        <div v-for="(item, index) in timeline" :key="index"
             class="timeline-item animate-on-scroll"
             :style="{ animationDelay: `${index * 0.2}s` }">
          <div class="timeline-dot">
            <div class="dot-inner"></div>
            <div class="dot-pulse"></div>
          </div>
          <div class="timeline-content">
            <div class="timeline-time">{{ item.time }}</div>
            <h3 class="timeline-title">{{ item.title }}</h3>
            <p class="timeline-desc">{{ item.desc }}</p>
            <div class="timeline-tags" v-if="item.tags">
              <span v-for="tag in item.tags" :key="tag" class="timeline-tag">{{ tag }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- AI 问答 Section（LLM Wiki 知识库） -->
    <section id="ask" class="section ask-section">
      <div class="section-header animate-on-scroll">
        <h2 class="section-title">问问我吧</h2>
        <div class="title-underline"></div>
        <p class="section-subtitle">AI 读过我的全部资料，替我随时在线</p>
      </div>
      <AskPanel />
    </section>

    <!-- Contact Section -->
    <section id="contact" class="section contact-section">
      <div class="contact-card animate-on-scroll">
        <div class="contact-bg-gradient"></div>
        <div class="contact-particles">
          <div v-for="i in 20" :key="i" class="contact-particle" :style="getContactParticleStyle(i)"></div>
        </div>
        <h2>{{ contactView.title }}</h2>
        <p>{{ contactView.desc }}</p>
        <div class="contact-links">
          <a
            v-for="(link, idx) in contactView.links"
            :key="idx"
            :href="link.url"
            target="_blank"
            class="contact-link"
          >
            <el-icon><component :is="link.icon" /></el-icon>
            <span>{{ link.label }}</span>
          </a>
        </div>
      </div>
    </section>

    <!-- Media Preview Dialog -->
    <el-dialog v-model="mediaPreviewVisible" :title="currentMedia?.caption" width="80%" center>
      <img v-if="currentMedia?.type === 'image'" :src="currentMedia.url" style="width: 100%; border-radius: 12px;" />
      <video v-else-if="currentMedia?.type === 'video'" :src="currentMedia.url" controls autoplay style="width: 100%; border-radius: 12px;" />
    </el-dialog>
  </div>
</template>

<script setup name="AboutPage">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import {
  ArrowRight,
  ArrowDown,
  CircleCheck,
  Close,
  Message,
  Cpu,
  Monitor,
  Promotion,
  Platform,
  Document,
  ChatDotRound,
  DataLine
} from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import { resolveIcon } from '@/utils/iconResolver.js'
import { BufferAttribute, BufferGeometry, CanvasTexture, Group, PerspectiveCamera, Points, PointsMaterial, SRGBColorSpace, Scene, WebGLRenderer } from 'three'
import CloudSky from '@/components/CloudSky.vue'
import FlickerText from '@/components/FlickerText.vue'
import AskPanel from '@/components/wiki/AskPanel.vue'

const Github = {
  template: `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg>`
}

// 后台配置数据
// Backend-configured data
const loading = ref(true)
const heroData = ref({
  avatarText: 'T',
  statusText: 'Ready for New Challenges',
  name: 'Txing',
  subtitle: '全栈开发工程师 / AI 架构爱好者 / 产品极客'
})

// Typing effect（仅副标题；主标题名字改由 FlickerText 呈现）
// Typing effect (subtitle only; the hero name is now rendered by FlickerText)
let fullSubtitle = ''
const displayedSubtitle = ref('')
const subtitleIndex = ref(0)

const subtitleInterval = ref(null)
let subtitleDelayTimer = null

const startTyping = () => {
  if (!fullSubtitle) return
  // 重置
  displayedSubtitle.value = ''
  subtitleIndex.value = 0

  if (subtitleInterval.value) clearInterval(subtitleInterval.value)
  if (subtitleDelayTimer) clearTimeout(subtitleDelayTimer)

  // 保留原节奏：延迟 300ms 后开始打副标题
  // Keep the original rhythm: subtitle typing starts after a 300ms delay
  subtitleDelayTimer = setTimeout(() => {
    subtitleInterval.value = setInterval(() => {
      if (subtitleIndex.value < fullSubtitle.length) {
        displayedSubtitle.value += fullSubtitle[subtitleIndex.value]
        subtitleIndex.value++
      } else {
        clearInterval(subtitleInterval.value)
      }
    }, 50)
  }, 300)
}

// FlickerText 名字闪烁配置
// FlickerText name flicker configs
// 入场：霓虹灯管点亮式——描边与实填交替亮起。注意：逐字母透明度闪烁对
// background-clip:text 渐变文字不可见（背景画在父层），故只用整词相位闪烁
// Enter: neon-tube warm-up alternating outline/filled. Letter-opacity flicker is
// invisible on background-clip:text gradients (background paints on the parent),
// so only whole-word phase flicker is used here
const nameEnterFlicker = {
  replay: 'yes',
  ease: { duration: 2.2, ease: 'easeInOut' },
  flickerCount: 9,
  showStroke: true,
  strokePosition: 'start',
  strokeCount: 2,
  strokeColor: 'var(--el-color-primary)',
  strokeWidth: 1.5,
  wordFlickerEnabled: true,
  letterFlickerEnabled: false
}
// 悬停：整词快速闪 3 下
// Hover: three quick whole-word blinks
const nameHoverFlicker = {
  ease: { duration: 1.5, ease: 'easeInOut' },
  flickerCount: 3,
  showStroke: false,
  wordFlickerEnabled: true,
  letterFlickerEnabled: false
}

onMounted(async () => {
  // 初始化 3D 星空背景
  initStarfield()
  // 初始化元素按压动效
  initField()
  // 先加载后台配置
  await loadAboutSnapshot()
  // 同步 typing 字符串（名字由 FlickerText 直接响应 heroData，无需同步）
  fullSubtitle = heroData.value.subtitle || ''
  // 启动打字机
  startTyping()

  // Setup Intersection Observer
  setupScrollAnimations()
})

onUnmounted(() => {
  if (subtitleDelayTimer) clearTimeout(subtitleDelayTimer)
  if (subtitleInterval.value) clearInterval(subtitleInterval.value)
  if (observer.value) observer.value.disconnect()
  disposeStarfield()
  disposeField()
})

const observer = ref(null)

const setupScrollAnimations = () => {
  observer.value = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('animated')
        observer.value.unobserve(entry.target)
      }
    })
  }, {
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px'
  })

  // Observe all elements with animate-on-scroll class
  setTimeout(() => {
    document.querySelectorAll('.animate-on-scroll').forEach(el => {
      observer.value.observe(el)
    })
  }, 100)
}

// 浮动小图标（来自后台）
// Floating icons (from backend)
const floatingIcons = ref([])

// 为什么选择我（来自后台）
// Why choose me cards (from backend)
const whyChooseMe = ref([])

// 核心能力（来自后台）
// Skill sets (from backend)
// 通过 iconKey 动态解析为组件
const skillSets = computed(() =>
  skills.value.map((s) => ({
    category: s.category,
    icon: resolveIcon(s.iconKey) || Monitor,
    tags: s.tags || [],
    level: s.level,
    sort: s.sort
  }))
)

// 精选作品（来自后台）
// Projects (from backend)
const projects = computed(() =>
  projectsRaw.value.map((p) => ({
    ...p,
    icon: resolveIcon(p.iconKey) || Platform,
    gradient: `project-gradient-${p.gradient || 1}`,
    category: p.category || 'company',
    // 封面轮播（图片/视频混合）；url 为后端读取时签名的临时地址
    // Cover carousel (mixed image/video); url is a signed temp URL generated on read
    coverMedia: (p.coverMedia || [])
      .map((m) => ({ type: m.type || 'image', url: m.url }))
      .filter((m) => m.url)
  }))
)

// 项目类别切换：all / company / personal（滑动指示的胶囊分段控件）
// Project category filter: all / company / personal (pill segmented control with sliding thumb)
const filterOptions = [
  { value: 'all', label: '全部' },
  { value: 'company', label: '公司项目' },
  { value: 'personal', label: '个人项目' }
]
const projectFilter = ref('all')
const filterIndex = computed(() => filterOptions.findIndex((o) => o.value === projectFilter.value))
const filteredProjects = computed(() =>
  projectFilter.value === 'all'
    ? projects.value
    : projects.value.filter((p) => p.category === projectFilter.value)
)
const countByCategory = (value) =>
  value === 'all'
    ? projects.value.length
    : projects.value.filter((p) => p.category === value).length
const switchFilter = (value) => {
  if (projectFilter.value === value) return
  projectFilter.value = value
  // 切换类别时收起已展开的卡片与亮点详情
  // Collapse expanded card & highlight detail on category switch
  expandedProject.value = null
  openFeature.value = null
}

// 成长轨迹（来自后台）
// Timeline (from backend)
const timeline = computed(() =>
  timelineRaw.value.map((t) => ({
    ...t,
    tags: t.tags || []
  }))
)

// 联系区（来自后台）
// Contact section (from backend)
const contactData = ref({
  title: '准备好一起创造价值了吗？',
  desc: '我一直在寻找具有挑战性的机会，期待与优秀的团队共同成长。',
  links: [
    {
      iconKey: 'Message',
      label: '邮件联系',
      url: 'mailto:contact@txing.ai'
    },
    {
      iconKey: 'Github',
      label: 'GitHub 仓库',
      url: 'https://github.com/lemon-puls/txing-ai'
    }
  ]
})

// 联系区视图对象（按 iconKey 解析图标组件）
const contactView = computed(() => ({
  title: contactData.value.title,
  desc: contactData.value.desc,
  links: (contactData.value.links || []).map((l) => ({
    ...l,
    icon: resolveIcon(l.iconKey) || Message
  }))
}))

// 内部使用的原始数据 ref（不被组件直接消费，由上面的 computed 加工）
const skills = ref([])
const projectsRaw = ref([])
const timelineRaw = ref([])

// 加载关于我页面聚合数据
// Load aggregated about-me snapshot from backend
const loadAboutSnapshot = async () => {
  loading.value = true
  try {
    const res = await defaultApi.apiAboutGet()
    if (res?.code === 0 && res.data) {
      const data = res.data
      console.log("data", data)
      // Hero
      if (data.hero) {
        heroData.value = {
          avatarText: data.hero.avatarText || 'T',
          statusText: data.hero.statusText || '',
          name: data.hero.name || 'Txing',
          subtitle: data.hero.subtitle || ''
        }
      }
      floatingIcons.value = data.floatingIcons || []
      whyChooseMe.value = (data.reasons || []).map((r) => ({
        emoji: r.emoji,
        title: r.title,
        desc: r.desc,
        tags: r.tags || [],
        stats: r.stats || []
      }))
      skills.value = data.skills || []
      projectsRaw.value = data.projects || []
      console.log(projectsRaw.value)
      timelineRaw.value = data.timeline || []
      if (data.contact) {
        contactData.value = {
          title: data.contact.title || contactData.value.title,
          desc: data.contact.desc || '',
          links: data.contact.links || contactData.value.links
        }
      }
    }
  } catch (err) {
    console.error('[About] 加载关于我数据失败：', err)
  } finally {
    loading.value = false
  }
}

// Project expand state
const expandedProject = ref(null)
// 工作亮点详情展开（单开，按标题记录当前展开项；切换项目时收起）
// Expanded highlight detail (single-open, keyed by title; collapsed on project switch)
const openFeature = ref(null)
// 展开状态按项目 id 记录（类别过滤会改变列表，index 定位会错位）
// Expansion keyed by project id (filtering reshuffles the list, index would mismatch)
const toggleProject = (id) => {
  expandedProject.value = expandedProject.value === id ? null : id
  openFeature.value = null
}
const toggleFeature = (title) => {
  openFeature.value = openFeature.value === title ? null : title
}
// 详情按换行拆段渲染
// Split the detail text into paragraphs by newlines
const splitDetail = (detail) =>
  String(detail || '')
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)

// Media preview
const mediaPreviewVisible = ref(false)
const currentMedia = ref(null)
const openMediaPreview = (media) => {
  currentMedia.value = media
  mediaPreviewVisible.value = true
}

// ===== 3D 星云星空背景（Three.js） =====
// ===== 3D nebula starfield background (Three.js) =====
const CUBE_SIZE = 1000 // 立方体空间边长 / cube space edge length
const STAR_TWINKLE = 0.35 // 闪烁速度系数 / twinkle speed factor
// 三层星等：共 6000 颗，大小/亮度/色温分层营造纵深（贴图为圆形柔光，避免方块颗粒）
// Three magnitude layers: 6000 stars total, sized/tinted for depth (round soft sprite, no square dots)
const STAR_LAYERS = [
  { count: 4000, size: 1.2, color: 0xaac8ff, opacity: 0.55 },
  { count: 1600, size: 2.2, color: 0x6ea8ff, opacity: 0.7 },
  { count: 400, size: 3.6, color: 0x4e8cff, opacity: 0.85 }
]

const starfieldRef = ref(null)
let starAnimationId = null
let starRenderer = null
let starScene = null
let starCamera = null
let starGroup = null
let starLayers = []
let starTexture = null

// 生成圆形柔光星点贴图：中心亮核 + 柔和光晕
// Generate the round soft-glow star sprite: bright core + gentle halo
const createStarTexture = () => {
  const size = 64
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  const gradient = ctx.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2)
  gradient.addColorStop(0, 'rgba(255, 255, 255, 1)')
  gradient.addColorStop(0.3, 'rgba(232, 240, 255, 0.85)')
  gradient.addColorStop(0.6, 'rgba(160, 200, 255, 0.22)')
  gradient.addColorStop(1, 'rgba(160, 200, 255, 0)')
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, size, size)
  const texture = new CanvasTexture(canvas)
  texture.colorSpace = SRGBColorSpace
  return texture
}

// 初始化星空：在立方体空间内随机分布半透明蓝色粒子，相机置于 z=220
// Init starfield: scatter translucent blue particles in a cube space, camera at z=220
const initStarfield = () => {
  const container = starfieldRef.value
  if (!container || starRenderer) return

  starScene = new Scene()
  starCamera = new PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 2000)
  starCamera.position.z = 220

  starTexture = createStarTexture()
  starGroup = new Group()
  starLayers = STAR_LAYERS.map((layer) => {
    const geometry = new BufferGeometry()
    const positions = new Float32Array(layer.count * 3)
    for (let i = 0; i < positions.length; i++) {
      positions[i] = (Math.random() - 0.5) * CUBE_SIZE
    }
    geometry.setAttribute('position', new BufferAttribute(positions, 3))
    const material = new PointsMaterial({
      color: layer.color,
      size: layer.size,
      map: starTexture,
      transparent: true,
      opacity: layer.opacity,
      sizeAttenuation: true,
      depthWrite: false
    })
    const points = new Points(geometry, material)
    starGroup.add(points)
    return { points, material, baseOpacity: layer.opacity, phase: Math.random() * Math.PI * 2 }
  })
  starScene.add(starGroup)

  starRenderer = new WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' })
  starRenderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  starRenderer.setSize(window.innerWidth, window.innerHeight)
  // 透明清屏色，露出容器的 CSS 渐变深空底色
  // Transparent clear color to reveal the container's CSS deep-space gradient
  starRenderer.setClearColor(0x000000, 0)
  container.appendChild(starRenderer.domElement)
  window.addEventListener('resize', handleStarfieldResize)

  animateStarfield()
}

// 每帧绕 Y 轴与 X 轴极缓慢旋转整团粒子云，各层相位错开轻微闪烁
// Rotate the whole cloud very slowly around Y/X axes; layers twinkle out of phase
const animateStarfield = () => {
  starAnimationId = requestAnimationFrame(animateStarfield)
  starGroup.rotation.y += 0.0012
  starGroup.rotation.x += 0.0004
  const t = performance.now() * 0.001 * STAR_TWINKLE
  for (const layer of starLayers) {
    layer.material.opacity = layer.baseOpacity * (0.82 + 0.18 * Math.sin(t + layer.phase))
  }
  starRenderer.render(starScene, starCamera)
}

// 窗口 resize 自适应，画布保持全屏
// Adapt to window resize, keep the canvas fullscreen
const handleStarfieldResize = () => {
  if (!starRenderer || !starCamera) return
  starCamera.aspect = window.innerWidth / window.innerHeight
  starCamera.updateProjectionMatrix()
  starRenderer.setSize(window.innerWidth, window.innerHeight)
}

// 离开页面时释放 WebGL 资源，避免上下文泄漏
// Dispose WebGL resources on leave to avoid context leaks
const disposeStarfield = () => {
  cancelAnimationFrame(starAnimationId)
  window.removeEventListener('resize', handleStarfieldResize)
  for (const layer of starLayers) {
    layer.points.geometry.dispose()
    layer.material.dispose()
  }
  starLayers = []
  if (starTexture) {
    starTexture.dispose()
    starTexture = null
  }
  if (starRenderer) {
    starRenderer.dispose()
    starRenderer.forceContextLoss?.()
    starRenderer.domElement.remove()
    starRenderer = null
  }
  starScene = null
  starCamera = null
  starGroup = null
}

// ===== 鼠标交互元素动效（磁性按压 + 弹性回弹） =====
// ===== Mouse-interactive element motion (magnetic press + elastic rebound) =====
const FIELD_RADIUS = 240 // 交互影响半径 px / influence radius
const FIELD_STRENGTH = 12 // 最大推离位移 px / max push displacement
const FIELD_STIFFNESS = 0.1 // 弹簧刚度 / spring stiffness
const FIELD_DAMPING = 0.75 // 弹簧阻尼（带轻微过冲的弹性感）/ spring damping (subtle overshoot)
const FIELD_SCAN_INTERVAL = 90 // 目标元素重扫间隔（帧）/ target rescan interval (frames)
// 参与按压动效的元素选择器 / selectors of elements joining the press effect
const FIELD_SELECTORS = '.why-me-card, .skill-card, .project-card, .timeline-content, .contact-card, .hero-actions .action-btn, .avatar-wrapper'

const fieldStateMap = new Map()
let fieldAnimationId = null
let fieldFrame = 0
const fieldMouse = { x: 0, y: 0, active: false }

// 收集目标元素：数据异步加载后卡片才会渲染，故周期性重扫并增量维护状态
// Collect target elements: cards render after async data, so rescan periodically
const collectFieldTargets = () => {
  const els = document.querySelectorAll(FIELD_SELECTORS)
  const seen = new Set()
  els.forEach((el) => {
    seen.add(el)
    if (!fieldStateMap.has(el)) {
      fieldStateMap.set(el, { el, x: 0, y: 0, vx: 0, vy: 0 })
    }
  })
  for (const key of [...fieldStateMap.keys()]) {
    if (!seen.has(key)) fieldStateMap.delete(key)
  }
}

// 物理步进：推离目标随距离二次衰减，弹簧阻尼积分产生按压与回弹
// Physics step: push target falls off quadratically, spring-damper integrates press & rebound
const updateField = () => {
  const states = [...fieldStateMap.values()]
  // 先集中读取几何信息，再统一写样式，避免逐元素读写触发布局抖动
  // Read all rects first, then write styles, to avoid per-element layout thrash
  const rects = states.map((s) => s.el.getBoundingClientRect())
  for (let i = 0; i < states.length; i++) {
    const s = states[i]
    const rect = rects[i]
    let tx = 0
    let ty = 0
    if (rect.width > 0 && fieldMouse.active) {
      // 减去当前位移还原真实中心，避免位移反馈干扰距离计算
      // Subtract the applied offset to recover the true center (no feedback)
      const cx = rect.left + rect.width / 2 - s.x
      const cy = rect.top + rect.height / 2 - s.y
      const dx = cx - fieldMouse.x
      const dy = cy - fieldMouse.y
      const dist = Math.sqrt(dx * dx + dy * dy)
      if (dist < FIELD_RADIUS && dist > 0.001) {
        const f = 1 - dist / FIELD_RADIUS
        const push = f * f * FIELD_STRENGTH
        tx = (dx / dist) * push
        ty = (dy / dist) * push
      }
    }
    s.vx = (s.vx + (tx - s.x) * FIELD_STIFFNESS) * FIELD_DAMPING
    s.vy = (s.vy + (ty - s.y) * FIELD_STIFFNESS) * FIELD_DAMPING
    s.x += s.vx
    s.y += s.vy
  }
  for (let i = 0; i < states.length; i++) {
    const s = states[i]
    if (!fieldMouse.active && Math.abs(s.x) < 0.02 && Math.abs(s.y) < 0.02 && Math.abs(s.vx) < 0.02 && Math.abs(s.vy) < 0.02) {
      s.x = 0
      s.y = 0
      s.vx = 0
      s.vy = 0
      s.el.style.translate = ''
      s.el.style.rotate = ''
      continue
    }
    // 用独立 transform 属性 translate/rotate，不覆盖元素自身的 CSS hover transform
    // Use individual transform props so the elements' own CSS hover transforms still compose
    s.el.style.translate = `${s.x.toFixed(2)}px ${s.y.toFixed(2)}px`
    s.el.style.rotate = `${(s.x * 0.06).toFixed(3)}deg`
  }
}

const animateField = () => {
  fieldAnimationId = requestAnimationFrame(animateField)
  if (fieldFrame % FIELD_SCAN_INTERVAL === 0) collectFieldTargets()
  fieldFrame++
  updateField()
}

const handleFieldPointerMove = (e) => {
  fieldMouse.x = e.clientX
  fieldMouse.y = e.clientY
  fieldMouse.active = true
}

// 鼠标移出窗口/页面失焦：释放按压，元素弹性回弹
// Pointer leaves the window / page blurs: release the press, elements rebound
const handleFieldPointerLeave = () => {
  fieldMouse.active = false
}

const initField = () => {
  if (fieldAnimationId) return
  // 尊重系统"减少动态效果"偏好 / respect the system reduced-motion preference
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  collectFieldTargets()
  window.addEventListener('pointermove', handleFieldPointerMove, { passive: true })
  document.documentElement.addEventListener('mouseleave', handleFieldPointerLeave)
  window.addEventListener('blur', handleFieldPointerLeave)
  animateField()
}

const disposeField = () => {
  cancelAnimationFrame(fieldAnimationId)
  window.removeEventListener('pointermove', handleFieldPointerMove)
  document.documentElement.removeEventListener('mouseleave', handleFieldPointerLeave)
  window.removeEventListener('blur', handleFieldPointerLeave)
  fieldAnimationId = null
  fieldStateMap.forEach((s) => {
    s.el.style.translate = ''
    s.el.style.rotate = ''
  })
  fieldStateMap.clear()
  fieldMouse.active = false
}

const getContactParticleStyle = (index) => {
  const size = Math.random() * 4 + 2
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    animationDelay: `${Math.random() * 5}s`
  }
}

const scrollToProjects = () => {
  document.getElementById('projects')?.scrollIntoView({ behavior: 'smooth' })
}

const scrollToContact = () => {
  document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<style scoped lang="scss">
.about-container {
  position: relative;
  width: 100%;
  color: var(--el-text-color-primary);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  overflow-x: hidden;
}

// 3D 星空背景容器：CSS 渐变作深空底色，WebGL 画布透明叠加
// Starfield container: CSS gradient as the deep-space base, WebGL canvas overlaid transparently
.starfield-bg {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  background: linear-gradient(180deg, #eaf1ff 0%, #f6f9ff 45%, #edf2fc 100%);

  html.dark & {
    background: linear-gradient(180deg, #04070f 0%, #0a1024 45%, #101a38 100%);
  }

  canvas {
    display: block;
  }
}

@keyframes float-particle {
  0%, 100% {
    transform: translateY(0) translateX(0) scale(1);
    opacity: 0.3;
  }
  25% {
    transform: translateY(-100px) translateX(50px) scale(1.2);
    opacity: 0.5;
  }
  50% {
    transform: translateY(-200px) translateX(-30px) scale(0.8);
    opacity: 0.2;
  }
  75% {
    transform: translateY(-150px) translateX(80px) scale(1.1);
    opacity: 0.4;
  }
}

.section {
  position: relative;
  z-index: 1;
  padding: 80px 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.section-header {
  text-align: center;
  margin-bottom: 60px;

  .section-title {
    font-size: 36px;
    font-weight: 800;
    margin-bottom: 12px;
    color: var(--el-text-color-primary, #1e293b);
    background: linear-gradient(135deg, var(--el-text-color-primary), var(--el-color-primary));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .title-underline {
    width: 80px;
    height: 4px;
    background: linear-gradient(90deg, var(--el-color-primary), #6366f1);
    margin: 0 auto;
    border-radius: 2px;
    position: relative;

    &::after {
      content: '';
      position: absolute;
      width: 12px;
      height: 12px;
      background: var(--el-color-primary);
      border-radius: 50%;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      box-shadow: 0 0 10px var(--el-color-primary);
    }
  }

  .section-subtitle {
    margin-top: 16px;
    font-size: 16px;
    color: var(--el-text-color-secondary);
  }
}

/* Scroll Animation Base */
.animate-on-scroll {
  opacity: 0;
  transform: translateY(40px);
  transition: opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1), transform 0.8s cubic-bezier(0.4, 0, 0.2, 1);

  &.animated {
    opacity: 1;
    transform: translateY(0);
  }

  &.delay-1 { transition-delay: 0.2s; }
  &.delay-2 { transition-delay: 0.4s; }
  &.delay-3 { transition-delay: 0.6s; }
}

/* Hero Section */
.hero-section {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 80px 24px;
  overflow: hidden;

  // 云海天幕：垫在光斑与内容之下，底部渐隐露出星空，避免与星空生硬切换
  // Cloud-sky dome: beneath blobs and content, fading down into the starfield for a seamless handoff
  .hero-sky {
    z-index: 0;
    -webkit-mask-image: linear-gradient(to bottom, #000 0%, #000 58%, transparent 97%);
    mask-image: linear-gradient(to bottom, #000 0%, #000 58%, transparent 97%);
  }

  .hero-content {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
  }

  .avatar-wrapper {
    position: relative;
    margin-bottom: 20px;

    .main-avatar {
      width: 140px;
      height: 140px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--el-color-primary), #6366f1);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 10px 40px rgba(43, 94, 255, 0.4);
      position: relative;

      .avatar-text {
        font-size: 56px;
        font-weight: 900;
        color: white;
        position: relative;
        z-index: 2;
      }

      .avatar-ring {
        position: absolute;
        width: 100%;
        height: 100%;
        border: 2px solid var(--el-color-primary-light-3);
        border-radius: 50%;
        animation: pulse-ring 3s infinite;

        &.ring-2 {
          animation-delay: 1.5s;
        }
      }
    }

    .status-badge {
      position: absolute;
      bottom: -15px;
      left: 50%;
      transform: translateX(-50%);
      // 毛玻璃胶囊：呼应云海/星空底幕的通透质感；品牌绿只留一线描边与呼吸光点，不再整块铺色
      // Glass capsule: echoes the sky/space backdrop; emerald shrinks to a hairline ring and a breathing dot
      padding: 7px 16px 7px 14px;
      border-radius: 999px;
      background: linear-gradient(135deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.55));
      backdrop-filter: blur(12px) saturate(1.5);
      -webkit-backdrop-filter: blur(12px) saturate(1.5);
      border: 1px solid rgba(16, 185, 129, 0.35);
      box-shadow:
        0 8px 24px rgba(16, 185, 129, 0.22),
        0 2px 8px rgba(15, 23, 42, 0.06),
        inset 0 1px 0 rgba(255, 255, 255, 0.65);
      color: #047857;
      font-size: 12.5px;
      font-weight: 600;
      letter-spacing: 0.02em;
      white-space: nowrap;
      display: flex;
      align-items: center;
      gap: 8px;
      overflow: hidden;
      animation: badge-bounce 3s ease-in-out infinite;

      // 掠过胶囊的绿调扫光（复用全局 shimmer 关键帧）
      // A green-tinted sheen sweeping across the capsule (reuses the global shimmer keyframes)
      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 50%;
        height: 100%;
        background: linear-gradient(105deg, transparent, rgba(52, 211, 153, 0.18), transparent);
        animation: shimmer 3.5s ease-in-out infinite;
        pointer-events: none;
      }

      // 暗色下沉入夜空：深色玻璃底 + 更亮的文字与描边
      // Dark theme sinks into the night: darker glass, brighter text and ring
      html.dark & {
        background: linear-gradient(135deg, rgba(15, 23, 42, 0.72), rgba(15, 23, 42, 0.5));
        border-color: rgba(52, 211, 153, 0.3);
        box-shadow:
          0 8px 24px rgba(0, 0, 0, 0.35),
          0 0 18px rgba(16, 185, 129, 0.12),
          inset 0 1px 0 rgba(255, 255, 255, 0.08);
        color: #6ee7b7;
      }

      .status-dot {
        position: relative;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        flex-shrink: 0;
        background: radial-gradient(circle at 30% 30%, #34d399, #059669);
        box-shadow: 0 0 6px rgba(16, 185, 129, 0.8);
        animation: pulse-dot 2s infinite;

        // 外圈信号波：与主头像 pulse-ring 同语言的“在线心跳”
        // Outer signal ring: an "online heartbeat" echoing the avatar's pulse-ring
        &::after {
          content: '';
          position: absolute;
          inset: 0;
          border-radius: 50%;
          border: 1px solid rgba(16, 185, 129, 0.6);
          animation: dot-ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
        }
      }
    }
  }

  .hero-title {
    font-size: 60px;
    font-weight: 900;
    margin: 0;
    color: var(--el-text-color-primary, #0f172a);
    letter-spacing: -1px;

    // FlickerText 名字特效：在组件内联渐变之上补回原 .highlight 的流动渐变
    // FlickerText name effect: layers the original flowing gradient over the component's inline gradient
    .flicker-name {
      background-size: 200% 200% !important;
      animation: gradient-shift 3s ease infinite;
    }
  }

  .hero-subtitle {
    font-size: 20px;
    color: var(--el-text-color-secondary, #64748b);
    max-width: 600px;
    line-height: 1.6;
    min-height: 32px;
  }

  .hero-actions {
    display: flex;
    gap: 16px;
    margin-top: 10px;

    .action-btn {
      padding: 14px 36px;
      font-weight: 600;
      font-size: 16px;
      // 显式列举过渡属性：排除 translate/rotate（由按压动效逐帧驱动）
      // Explicit transition props: exclude translate/rotate (driven per-frame by the press effect)
      transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1);
      position: relative;
      overflow: hidden;

      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: -100%;
        width: 100%;
        height: 100%;
        background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
        transition: left 0.5s;
      }

      &:hover::before {
        left: 100%;
      }

      .btn-content {
        display: flex;
        align-items: center;
        gap: 8px;
      }

      .btn-arrow {
        animation: bounce-down 1.5s ease-in-out infinite;
      }

      &.secondary {
        background-color: var(--el-bg-color);
        border: 2px solid var(--el-border-color);
        color: var(--el-text-color-regular);

        &:hover {
          background-color: var(--el-fill-color-light);
          border-color: var(--el-color-primary);
          color: var(--el-color-primary);
        }
      }

      &:hover {
        transform: translateY(-4px);
        box-shadow: 0 12px 30px rgba(43, 94, 255, 0.3);
      }
    }
  }

  .scroll-indicator {
    margin-top: 40px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    color: var(--el-text-color-secondary);
    font-size: 14px;

    .mouse {
      width: 28px;
      height: 44px;
      border: 2px solid var(--el-text-color-secondary);
      border-radius: 14px;
      display: flex;
      justify-content: center;
      padding-top: 8px;

      .wheel {
        width: 4px;
        height: 10px;
        background: var(--el-color-primary);
        border-radius: 2px;
        animation: scroll-wheel 2s ease-in-out infinite;
      }
    }
  }

  .hero-bg-blobs {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1;
    pointer-events: none;

    .blob {
      position: absolute;
      filter: blur(100px);
      opacity: 0.5;
      border-radius: 50%;
      animation: float-blob 15s infinite alternate ease-in-out;

      // 暗色下云海已撑起天幕，彩斑调暗以免与夜空云层打架
      // Dark theme: the cloud dome already fills the sky; dim the blobs to avoid clutter
      html.dark & {
        opacity: 0.22;
      }
    }

    .blob-1 {
      width: 500px;
      height: 500px;
      background: var(--el-color-primary-light-5);
      top: 5%;
      left: 10%;
    }

    .blob-2 {
      width: 400px;
      height: 400px;
      background: #c7d2fe;
      bottom: 10%;
      right: 10%;
      animation-delay: -7s;
    }

    .blob-3 {
      width: 300px;
      height: 300px;
      background: #e0e7ff;
      top: 50%;
      left: 50%;
      animation-delay: -3s;
    }
  }

  .floating-icons {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1;
    pointer-events: none;
    overflow: hidden;
  }

  .floating-icon {
    position: absolute;
    font-size: 32px;
    opacity: 0.15;
    animation: float-icon 20s linear infinite;

    &.icon-0 { top: 10%; left: 5%; }
    &.icon-1 { top: 20%; right: 10%; }
    &.icon-2 { top: 60%; left: 8%; }
    &.icon-3 { top: 40%; right: 15%; }
    &.icon-4 { bottom: 20%; left: 15%; }
    &.icon-5 { bottom: 30%; right: 5%; }
  }
}

/* Why Choose Me Section */
.why-me-section {
  background: transparent;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background:
      radial-gradient(ellipse at 20% 50%, rgba(99, 102, 241, 0.08) 0%, transparent 50%),
      radial-gradient(ellipse at 80% 50%, rgba(139, 92, 246, 0.06) 0%, transparent 50%);
    pointer-events: none;
  }
}

.why-me-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 24px;
  position: relative;
  z-index: 1;
}

.why-me-card {
  position: relative;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(20px);
  padding: 32px;
  border-radius: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  // 显式列举过渡属性：排除 translate/rotate（由按压动效逐帧驱动）
  // Explicit transition props: exclude translate/rotate (driven per-frame by the press effect)
  transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.4s cubic-bezier(0.4, 0, 0.2, 1), background 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(255, 255, 255, 0.8);
  overflow: hidden;

  &:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 40px rgba(43, 94, 255, 0.15);
    border-color: rgba(99, 102, 241, 0.3);
    background: rgba(255, 255, 255, 0.8);

    .card-glow {
      opacity: 1;
    }

    .card-icon-wrapper .icon-bg {
      transform: scale(1.2);
    }
  }

  .card-glow {
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(circle, var(--el-color-primary-light-9) 0%, transparent 70%);
    opacity: 0;
    transition: opacity 0.5s;
    pointer-events: none;
  }

  .card-icon-wrapper {
    position: relative;
    width: 72px;
    height: 72px;
    margin-bottom: 24px;

    .icon-bg {
      position: absolute;
      width: 100%;
      height: 100%;
      background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.15));
      border-radius: 20px;
      transition: transform 0.4s;
    }

    .card-emoji {
      position: relative;
      z-index: 2;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 100%;
      font-size: 36px;
    }
  }

  .card-title {
    font-size: 20px;
    font-weight: 700;
    margin: 0 0 12px 0;
    color: var(--el-text-color-primary);
  }

  .card-desc {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    line-height: 1.7;
    margin: 0 0 20px 0;
  }

  .card-stats {
    display: flex;
    gap: 24px;
    margin-bottom: 20px;
    padding: 16px;
    background: rgba(99, 102, 241, 0.06);
    border-radius: 16px;

    .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 4px;

      .stat-number {
        font-size: 24px;
        font-weight: 800;
        background: linear-gradient(135deg, var(--el-color-primary), #6366f1);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
      }

      .stat-label {
        font-size: 12px;
        color: var(--el-text-color-secondary);
      }
    }
  }

  .card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;

    .tag {
      background: rgba(99, 102, 241, 0.08);
      color: var(--el-color-primary);
      padding: 6px 14px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 500;
      transition: all 0.2s;

      &:hover {
        background: rgba(99, 102, 241, 0.15);
        transform: scale(1.05);
      }
    }
  }
}

/* Skills Section */
.skills-section {
  background: transparent;
}

.skills-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;

  .skill-card {
    background: var(--el-bg-color);
    padding: 30px;
    border-radius: 24px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
    // 显式列举过渡属性：排除 translate/rotate（由按压动效逐帧驱动）
    // Explicit transition props: exclude translate/rotate (driven per-frame by the press effect)
    transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.4s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--el-border-color-lighter);

    &:hover {
      transform: translateY(-8px);
      box-shadow: 0 16px 40px rgba(43, 94, 255, 0.12);
      border-color: var(--el-color-primary-light-5);
    }

    .skill-icon {
      width: 60px;
      height: 60px;
      border-radius: 18px;
      background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-color-primary-light-7));
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 20px;

      .el-icon {
        font-size: 30px;
        color: var(--el-color-primary);
      }
    }

    .skill-title {
      font-size: 22px;
      font-weight: 700;
      margin: 0 0 16px 0;
      color: var(--el-text-color-primary);
    }

    .skill-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      margin-bottom: 20px;

      .skill-tag {
        background-color: var(--el-fill-color-light);
        color: var(--el-text-color-regular);
        padding: 8px 14px;
        border-radius: 10px;
        font-size: 13px;
        font-weight: 500;
        transition: all 0.3s;

        &:hover {
          background-color: var(--el-color-primary-light-9);
          color: var(--el-color-primary);
          transform: translateY(-2px);
        }
      }
    }

    .skill-progress {
      display: flex;
      align-items: center;
      gap: 12px;

      .progress-bar {
        flex: 1;
        height: 8px;
        background: var(--el-fill-color-light);
        border-radius: 4px;
        overflow: hidden;

        .progress-fill {
          height: 100%;
          background: linear-gradient(90deg, var(--el-color-primary), #6366f1);
          border-radius: 4px;
          transition: width 1.5s cubic-bezier(0.4, 0, 0.2, 1);
          position: relative;

          &::after {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
            animation: shimmer 2s infinite;
          }
        }
      }

      .progress-text {
        font-size: 14px;
        font-weight: 600;
        color: var(--el-color-primary);
        min-width: 40px;
      }
    }
  }
}

/* Projects Section */
.projects-section {
  background: transparent;
}

.projects-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 36px;
  max-width: 1080px;
  margin: 0 auto;
}

/* 项目类别切换器：滑动指示的胶囊分段控件 */
/* Project category switcher: pill segmented control with sliding thumb */
.project-filter {
  display: flex;
  justify-content: center;
  margin: 0 auto 44px;
  max-width: 1080px;
}

.filter-track {
  position: relative;
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  padding: 5px;
  border-radius: 999px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
}

.filter-thumb {
  position: absolute;
  top: 5px;
  bottom: 5px;
  left: 5px;
  width: calc((100% - 10px) / 3);
  border-radius: 999px;
  background: linear-gradient(135deg, var(--el-color-primary), var(--el-color-primary-light-3));
  box-shadow: 0 4px 14px rgba(43, 94, 255, 0.35);
  transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.filter-btn {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 22px;
  border: none;
  background: transparent;
  border-radius: 999px;
  font-size: 14px;
  font-family: inherit;
  color: var(--el-text-color-regular);
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.3s;
}

.filter-btn:hover {
  color: var(--el-text-color-primary);
}

.filter-btn.active {
  color: #fff;
}

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  font-size: 12px;
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
  transition: background 0.3s, color 0.3s;
}

.filter-btn.active .filter-count {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}

.project-card {
  background: var(--el-bg-color);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1), transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  display: grid;
  grid-template-columns: 1fr;

  @media (min-width: 768px) {
    grid-template-columns: 42% 1fr;
  }

  &:hover {
    transform: translateY(-6px);
    box-shadow: 0 24px 50px rgba(43, 94, 255, 0.15);
  }

  &.expanded {
    box-shadow: 0 24px 50px rgba(43, 94, 255, 0.2);
  }

  .project-image {
    height: 280px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    cursor: pointer;

    @media (min-width: 768px) {
      height: auto;
      min-height: 320px;
    }

    // 封面轮播：铺满整个图区，压在渐变背景之上、角标/展开按钮之下
    // Cover carousel fills the media area, above gradient, below badge/expand hint
    .cover-carousel {
      position: absolute;
      inset: 0;
      z-index: 0;

      :deep(.el-carousel__container) {
        height: 100%;
      }

      :deep(.el-carousel__arrow) {
        background: rgba(0, 0, 0, 0.35);
      }

      .cover-media-wrap {
        position: absolute;
        inset: 0;
        overflow: hidden;
      }

      // 模糊垫底：放大超出容器，避免 blur 边缘露白
      // Blurred oversized copy as backdrop; oversize hides blur edge artifacts
      .cover-media-blur {
        position: absolute;
        inset: -24px;
        width: calc(100% + 48px);
        height: calc(100% + 48px);
        object-fit: cover;
        filter: blur(28px) brightness(0.85) saturate(1.1);
      }

      // 主体媒体：contain 完整展示，不裁切
      // Main media: contain — fully visible, no cropping
      .cover-media {
        position: absolute;
        inset: 0;
        display: block;
        width: 100%;
        height: 100%;
        object-fit: contain;
      }
    }

    &.project-gradient-1 {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }

    &.project-gradient-2 {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    }

    &.project-gradient-3 {
      background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
    }

    &.project-gradient-4 {
      background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
    }

    .project-icon {
      width: 90px;
      height: 90px;
      background: rgba(255, 255, 255, 0.2);
      backdrop-filter: blur(10px);
      border-radius: 24px;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.4s ease;

      .el-icon {
        font-size: 44px;
        color: white;
      }
    }

    .project-badge {
      position: absolute;
      top: 16px;
      right: 16px;
      background: rgba(255, 255, 255, 0.25);
      backdrop-filter: blur(10px);
      color: white;
      padding: 8px 16px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 600;
    }

    .expand-hint {
      position: absolute;
      bottom: 16px;
      left: 50%;
      transform: translateX(-50%);
      background: rgba(255, 255, 255, 0.2);
      backdrop-filter: blur(10px);
      width: 40px;
      height: 40px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.3s;

      .expand-icon {
        color: white;
        font-size: 20px;
        transition: transform 0.3s;

        &.rotated {
          transform: rotate(180deg);
        }
      }
    }

    &:hover {
      .project-icon {
        transform: scale(1.1) rotate(5deg);
      }

      .expand-hint {
        background: rgba(255, 255, 255, 0.3);
      }
    }
  }

  .project-info {
    padding: 32px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    min-width: 0;

    @media (min-width: 768px) {
      padding: 36px 40px;
    }

    .project-tags {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;

      .project-tag {
        font-size: 12px;
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
        padding: 6px 12px;
        border-radius: 8px;
        font-weight: 600;
        transition: all 0.2s;

        &:hover {
          background: var(--el-color-primary-light-7);
          transform: scale(1.05);
        }
      }
    }

    .project-name {
      font-size: 24px;
      font-weight: 800;
      margin: 0;
      color: var(--el-text-color-primary);
      cursor: pointer;
      transition: color 0.2s;

      &:hover {
        color: var(--el-color-primary);
      }
    }

    .project-desc {
      color: var(--el-text-color-secondary);
      line-height: 1.7;
      margin: 0;
      font-size: 14px;
    }

    .project-highlights {
      display: flex;
      flex-direction: column;
      gap: 10px;
      margin-top: 12px;

      .highlight-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        font-size: 13px;
        color: var(--el-text-color-regular);

        .el-icon {
          color: #10b981;
          font-size: 18px;
          margin-top: 2px;
          flex-shrink: 0;
        }
      }
    }

    .project-link {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      color: var(--el-color-primary);
      font-weight: 600;
      font-size: 14px;
      text-decoration: none;
      margin-top: 16px;
      padding: 10px 20px;
      background: var(--el-color-primary-light-9);
      border-radius: 12px;
      width: fit-content;
      transition: all 0.3s;

      &:hover {
        background: var(--el-color-primary);
        color: white;
        transform: translateX(4px);
      }
    }
  }

  /* Expanded Detail */
  .project-detail {
    grid-column: 1 / -1;
    border-top: 1px solid var(--el-border-color-lighter);
    padding: 32px 40px;
    background: var(--el-fill-color-lighter);

    @media (max-width: 767px) {
      padding: 28px;
    }

    .detail-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 24px;

      h4 {
        font-size: 18px;
        font-weight: 700;
        margin: 0;
        color: var(--el-text-color-primary);
      }
    }

    .media-gallery {
      margin-bottom: 24px;

      .gallery-scroll {
        display: flex;
        gap: 16px;
        overflow-x: auto;
        padding-bottom: 12px;
        scrollbar-width: thin;

        &::-webkit-scrollbar {
          height: 6px;
        }

        &::-webkit-scrollbar-thumb {
          background: var(--el-color-primary-light-5);
          border-radius: 3px;
        }
      }

      .media-item {
        flex-shrink: 0;
        width: 280px;
        border-radius: 16px;
        overflow: hidden;
        background: var(--el-bg-color);
        box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
        transition: all 0.3s;
        cursor: pointer;

        &:hover {
          transform: scale(1.03);
          box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
        }

        img, video {
          width: 100%;
          height: 160px;
          object-fit: cover;
        }

        .media-caption {
          padding: 12px;
          font-size: 13px;
          color: var(--el-text-color-secondary);
          text-align: center;
        }
      }
    }

    .tech-stack, .key-features {
      margin-bottom: 24px;

      h4 {
        font-size: 16px;
        font-weight: 700;
        margin: 0 0 16px 0;
        color: var(--el-text-color-primary);
      }
    }

    .tech-items {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;

      .tech-item {
        display: flex;
        align-items: center;
        gap: 8px;
        background: var(--el-bg-color);
        padding: 10px 18px;
        border-radius: 12px;
        font-size: 14px;
        font-weight: 500;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
        transition: all 0.2s;

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .tech-icon {
          font-size: 20px;
        }
      }
    }

    .features-grid {
      // 单列布局：展开的详细设计需要整行宽度才可读
      // Single column: expanded design details need the full row width
      display: grid;
      grid-template-columns: 1fr;
      gap: 12px;

      .feature-item {
        display: flex;
        gap: 14px;
        background: var(--el-bg-color);
        padding: 18px;
        border-radius: 16px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
        border: 1px solid transparent;
        transition: box-shadow 0.2s, border-color 0.2s;

        // 有详情的项可点击展开
        // Items with details are clickable
        &.has-detail {
          cursor: pointer;
        }

        &.detail-open {
          border-color: var(--el-color-primary-light-7);
        }

        &:hover {
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .feature-icon {
          font-size: 28px;
          flex-shrink: 0;
        }

        .feature-body {
          flex: 1;
          min-width: 0;
        }

        .feature-head {
          display: flex;
          align-items: center;
          justify-content: space-between;
          gap: 8px;
          margin-bottom: 6px;

          h5 {
            font-size: 15px;
            font-weight: 700;
            margin: 0;
            color: var(--el-text-color-primary);
          }

          .feature-chevron {
            color: var(--el-color-primary);
            transition: transform 0.3s;
          }
        }

        &.detail-open .feature-chevron {
          transform: rotate(180deg);
        }

        p {
          font-size: 13px;
          color: var(--el-text-color-secondary);
          margin: 0;
          line-height: 1.5;
        }

        // 0fr→1fr 网格展开动画（内容始终渲染，比 max-height 更稳）
        // 0fr→1fr grid expand animation (content stays rendered; more robust than max-height)
        .feature-detail-wrap {
          display: grid;
          grid-template-rows: 0fr;
          transition: grid-template-rows 0.35s cubic-bezier(0.4, 0, 0.2, 1);

          .feature-detail-clip {
            overflow: hidden;
          }

          .feature-detail {
            margin-top: 12px;
            padding: 12px 14px;
            border-left: 2px solid var(--el-color-primary-light-5);
            background: var(--el-fill-color-light);
            border-radius: 0 10px 10px 0;

            p {
              font-size: 13px;
              color: var(--el-text-color-regular);
              line-height: 1.7;
              margin: 0;

              & + p {
                margin-top: 8px;
              }
            }
          }
        }

        &.detail-open .feature-detail-wrap {
          grid-template-rows: 1fr;
        }
      }
    }
  }
}

/* Slide Fade Transition */
.slide-fade-enter-active {
  transition: opacity 0.4s cubic-bezier(0.4, 0, 0.2, 1), transform 0.4s cubic-bezier(0.4, 0, 0.2, 1), max-height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.slide-fade-leave-active {
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.3s cubic-bezier(0.4, 0, 0.2, 1), max-height 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.slide-fade-enter-from {
  transform: translateY(-20px);
  opacity: 0;
  max-height: 0;
}

.slide-fade-leave-to {
  transform: translateY(-10px);
  opacity: 0;
  max-height: 0;
}

.slide-fade-enter-to,
.slide-fade-leave-from {
  max-height: 1000px;
}

/* Timeline Section */
.timeline-section {
  background: transparent;
}

.timeline {
  position: relative;
  max-width: 700px;
  margin: 0 auto;

  &::before {
    content: '';
    position: absolute;
    left: 20px;
    top: 0;
    bottom: 0;
    width: 3px;
    background: linear-gradient(180deg, var(--el-color-primary), var(--el-color-primary-light-3), transparent);
    border-radius: 2px;
  }

  .timeline-item {
    position: relative;
    padding-left: 60px;
    margin-bottom: 40px;

    &:last-child {
      margin-bottom: 0;
    }

    .timeline-dot {
      position: absolute;
      left: 10px;
      top: 0;
      width: 24px;
      height: 24px;
      background: var(--el-bg-color);
      border: 3px solid var(--el-color-primary);
      border-radius: 50%;
      z-index: 2;
      display: flex;
      align-items: center;
      justify-content: center;

      .dot-inner {
        width: 8px;
        height: 8px;
        background: var(--el-color-primary);
        border-radius: 50%;
      }

      .dot-pulse {
        position: absolute;
        width: 100%;
        height: 100%;
        border-radius: 50%;
        border: 2px solid var(--el-color-primary);
        animation: pulse-ring 2s infinite;
      }
    }

    .timeline-content {
      background: var(--el-bg-color);
      padding: 28px;
      border-radius: 20px;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
      // 显式列举过渡属性：排除 translate/rotate（由按压动效逐帧驱动）
      // Explicit transition props: exclude translate/rotate (driven per-frame by the press effect)
      transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.4s cubic-bezier(0.4, 0, 0.2, 1);
      border: 1px solid var(--el-border-color-lighter);

      &:hover {
        transform: translateX(12px);
        box-shadow: 0 12px 30px rgba(43, 94, 255, 0.1);
        border-color: var(--el-color-primary-light-5);
      }

      .timeline-time {
        font-size: 14px;
        font-weight: 600;
        color: var(--el-color-primary);
        margin-bottom: 10px;
        display: inline-block;
        background: var(--el-color-primary-light-9);
        padding: 4px 12px;
        border-radius: 8px;
      }

      .timeline-title {
        font-size: 20px;
        font-weight: 700;
        margin: 0 0 10px 0;
        color: var(--el-text-color-primary);
      }

      .timeline-desc {
        font-size: 14px;
        color: var(--el-text-color-secondary);
        line-height: 1.7;
        margin: 0 0 16px 0;
      }

      .timeline-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;

        .timeline-tag {
          font-size: 12px;
          color: var(--el-text-color-regular);
          background: var(--el-fill-color-light);
          padding: 6px 12px;
          border-radius: 8px;
          font-weight: 500;
        }
      }
    }
  }
}

/* Contact Section */
.contact-section {
  display: flex;
  justify-content: center;
  text-align: center;
  padding-bottom: 100px;
}

.contact-card {
  position: relative;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  color: white;
  padding: 70px 40px;
  border-radius: 36px;
  width: 100%;
  max-width: 850px;
  box-shadow: 0 25px 60px rgba(15, 23, 42, 0.4);
  overflow: hidden;

  .contact-bg-gradient {
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(circle, rgba(43, 94, 255, 0.15) 0%, transparent 70%);
    animation: rotate-gradient 20s linear infinite;
    pointer-events: none;
  }

  .contact-particles {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;

    .contact-particle {
      position: absolute;
      background: rgba(255, 255, 255, 0.1);
      border-radius: 50%;
      animation: float-particle 15s linear infinite;
    }
  }

  h2 {
    position: relative;
    font-size: 36px;
    font-weight: 800;
    margin-bottom: 20px;
    z-index: 1;
  }

  p {
    position: relative;
    font-size: 18px;
    color: #94a3b8;
    margin-bottom: 48px;
    z-index: 1;
  }

  .contact-links {
    position: relative;
    display: flex;
    justify-content: center;
    gap: 24px;
    flex-wrap: wrap;
    z-index: 1;

    .contact-link {
      display: flex;
      align-items: center;
      gap: 10px;
      background: rgba(255, 255, 255, 0.1);
      color: white;
      padding: 16px 32px;
      border-radius: 16px;
      text-decoration: none;
      font-weight: 600;
      font-size: 16px;
      transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
      border: 1px solid rgba(255, 255, 255, 0.1);
      backdrop-filter: blur(10px);

      &:hover {
        background: rgba(255, 255, 255, 0.2);
        transform: translateY(-4px);
        border-color: rgba(255, 255, 255, 0.3);
        box-shadow: 0 12px 30px rgba(0, 0, 0, 0.3);
      }
    }
  }
}

/* Animations */
@keyframes float-blob {
  0% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -50px) scale(1.1); }
  66% { transform: translate(-20px, 30px) scale(0.9); }
  100% { transform: translate(50px, 50px) scale(1.05); }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

@keyframes pulse-ring {
  0% { transform: scale(1); opacity: 1; }
  100% { transform: scale(1.5); opacity: 0; }
}

@keyframes badge-bounce {
  0%, 100% { transform: translateX(-50%) translateY(0); }
  50% { transform: translateX(-50%) translateY(-4px); }
}

@keyframes dot-ping {
  0% { transform: scale(1); opacity: 0.9; }
  80%, 100% { transform: scale(2.4); opacity: 0; }
}

@keyframes rotate-gradient {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes gradient-shift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

@keyframes bounce-down {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(4px); }
}

@keyframes scroll-wheel {
  0% { transform: translateY(0); opacity: 1; }
  100% { transform: translateY(12px); opacity: 0; }
}

@keyframes float-icon {
  0% { transform: translateY(0) rotate(0deg); }
  25% { transform: translateY(-30px) rotate(90deg); }
  50% { transform: translateY(-10px) rotate(180deg); }
  75% { transform: translateY(-40px) rotate(270deg); }
  100% { transform: translateY(0) rotate(360deg); }
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* Responsive */
@media screen and (max-width: 768px) {
  .hero-title {
    font-size: 36px !important;
  }

  .hero-subtitle {
    font-size: 16px !important;
  }

  .hero-actions {
    flex-direction: column;
    width: 100%;
    max-width: 300px;
  }

  .projects-grid {
    max-width: 100%;
  }

  .skills-grid {
    grid-template-columns: 1fr;
  }

  .why-me-grid {
    grid-template-columns: 1fr;
  }

  .section {
    padding: 60px 16px;
  }

  .section-header {
    .section-title {
      font-size: 28px;
    }
  }

  .contact-card {
    padding: 40px 24px;
    border-radius: 24px;

    h2 {
      font-size: 24px;
    }

    p {
      font-size: 16px;
    }
  }

  .project-card {
    grid-template-columns: 1fr;

    .project-image {
      height: 220px;
      min-height: 0;
    }

    .project-info {
      padding: 28px;
    }

    .project-detail {
      .features-grid {
        grid-template-columns: 1fr;
      }
    }
  }
}
</style>
