<template>
  <div class="about-container">
    <!-- Particle Background -->
    <div class="particle-bg" ref="particleBg">
      <div v-for="i in 50" :key="i" class="particle" :style="getParticleStyle(i)"></div>
    </div>

    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content animate-on-scroll">
        <div class="avatar-wrapper">
          <div class="main-avatar">
            <span class="avatar-text">T</span>
            <div class="avatar-ring"></div>
            <div class="avatar-ring ring-2"></div>
          </div>
          <div class="status-badge">
            <span class="status-dot"></span>
            Ready for New Challenges
          </div>
        </div>
        <h1 class="hero-title">
          你好，我是 <span class="highlight typing-text">{{ displayedName }}</span>
          <span class="cursor">|</span>
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
      <div class="projects-grid">

        <div v-for="(project, index) in projects" :key="project.name" 
             class="project-card"
             :class="{ 'expanded': expandedProject === index }">
          <div class="project-image" :class="`project-gradient-${index + 1}`" @click="toggleProject(index)">
            <div class="project-icon">
              <el-icon><component :is="project.icon" /></el-icon>
            </div>
            <div class="project-badge">{{ project.badge }}</div>
            <div class="expand-hint">
              <el-icon class="expand-icon" :class="{ 'rotated': expandedProject === index }">
                <ArrowDown />
              </el-icon>
            </div>
          </div>
          <div class="project-info">
            <div class="project-tags">
              <span v-for="tag in project.tags" :key="tag" class="project-tag">{{ tag }}</span>
            </div>
            <h3 class="project-name" @click="toggleProject(index)">{{ project.name }}</h3>
            <p class="project-desc">{{ project.desc }}</p>
            <div class="project-highlights">
              <div v-for="hl in project.highlights" :key="hl" class="highlight-item">
                <el-icon><CircleCheck /></el-icon>
                <span>{{ hl }}</span>
              </div>
            </div>
            <a :href="project.link" target="_blank" class="project-link">
              <span>访问项目</span>
              <el-icon><ArrowRight /></el-icon>
            </a>
          </div>
          
          <!-- Expanded Detail Section -->
          <transition name="slide-fade">
            <div v-if="expandedProject === index" class="project-detail">
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

              <!-- Key Features -->
              <div class="key-features" v-if="project.features">
                <h4>核心功能</h4>
                <div class="features-grid">
                  <div v-for="feature in project.features" :key="feature.title" class="feature-item">
                    <span class="feature-icon">{{ feature.icon }}</span>
                    <div>
                      <h5>{{ feature.title }}</h5>
                      <p>{{ feature.desc }}</p>
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

    <!-- Contact Section -->
    <section id="contact" class="section contact-section">
      <div class="contact-card animate-on-scroll">
        <div class="contact-bg-gradient"></div>
        <div class="contact-particles">
          <div v-for="i in 20" :key="i" class="contact-particle" :style="getContactParticleStyle(i)"></div>
        </div>
        <h2>准备好一起创造价值了吗？</h2>
        <p>我一直在寻找具有挑战性的机会，期待与优秀的团队共同成长。</p>
        <div class="contact-links">
          <a href="mailto:contact@txing.ai" class="contact-link">
            <el-icon><Message /></el-icon>
            <span>邮件联系</span>
          </a>
          <a href="https://github.com/lemon-puls/txing-ai" target="_blank" class="contact-link">
            <el-icon><Github /></el-icon>
            <span>GitHub 仓库</span>
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

const Github = {
  template: `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg>`
}

// Typing effect
const fullText = 'Txing'
const displayedName = ref('')
const nameIndex = ref(0)

const fullSubtitle = '全栈开发工程师 / AI 架构爱好者 / 产品极客'
const displayedSubtitle = ref('')
const subtitleIndex = ref(0)

const typingInterval = ref(null)
const subtitleInterval = ref(null)

onMounted(() => {
  // Start typing name
  typingInterval.value = setInterval(() => {
    if (nameIndex.value < fullText.length) {
      displayedName.value += fullText[nameIndex.value]
      nameIndex.value++
    } else {
      clearInterval(typingInterval.value)
      // Start typing subtitle after name is done
      setTimeout(() => {
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
  }, 150)

  // Setup Intersection Observer
  setupScrollAnimations()
})

onUnmounted(() => {
  if (typingInterval.value) clearInterval(typingInterval.value)
  if (subtitleInterval.value) clearInterval(subtitleInterval.value)
  if (observer.value) observer.value.disconnect()
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

// Floating icons data
const floatingIcons = [
  { name: 'Vue', symbol: '⚡' },
  { name: 'Go', symbol: '🔧' },
  { name: 'AI', symbol: '🤖' },
  { name: 'Code', symbol: '💻' },
  { name: 'Database', symbol: '🗄️' },
  { name: 'Cloud', symbol: '☁️' }
]

// Why Choose Me data
const whyChooseMe = [
  {
    emoji: '🔥',
    title: '热爱技术，深度钻研',
    desc: '不仅仅是使用框架，更追求理解底层原理。深入研读 Hermes、Gin 等优秀开源项目源码，从大师代码中汲取设计智慧。',
    tags: ['源码阅读', '架构设计', '持续学习'],
    stats: [
      { value: '10+', label: '源码项目' },
      { value: '500+', label: '技术文章' }
    ]
  },
  {
    emoji: '🚀',
    title: '独立交付，全栈能力',
    desc: '从需求分析到架构设计，从后端 API 到前端 UI，独立完成多个完整系统的开发与上线。具备端到端的项目交付能力。',
    tags: ['全栈开发', '独立交付', '产品思维'],
    stats: [
      { value: '5+', label: '独立项目' },
      { value: '100%', label: '交付率' }
    ]
  },
  {
    emoji: '🤖',
    title: 'AI 原生，创新实践',
    desc: '深度实践 AI Agent 架构，掌握 MCP 协议、RAG、Prompt Engineering 等前沿技术。将 AI 能力融入产品，创造真实价值。',
    tags: ['AI Agent', 'MCP Protocol', 'RAG'],
    stats: [
      { value: '3+', label: 'AI 项目' },
      { value: '20+', label: 'Agent 技能' }
    ]
  },
  {
    emoji: '💡',
    title: '产品思维，用户至上',
    desc: '不只是写代码，更关注用户体验和产品价值。每一个功能都经过精心设计，确保简洁、高效、易用。',
    tags: ['UX 设计', '性能优化', '用户研究']
  }
]

const skillSets = [
  {
    category: '后端开发',
    icon: Cpu,
    tags: ['Go / Gin', 'MySQL', 'Redis', 'GORM', 'RESTful API', 'Microservices'],
    level: 92
  },
  {
    category: '前端开发',
    icon: Monitor,
    tags: ['Vue 3', 'Element Plus', 'TypeScript', 'Vite', 'Pinia', 'SCSS'],
    level: 88
  },
  {
    category: 'AI & 创新',
    icon: Promotion,
    tags: ['LLM Prompting', 'RAG', 'Agent Orchestration', 'MCP Protocol', 'SSE'],
    level: 85
  }
]

const projects = [
  {
    name: 'Txing AI',
    desc: '一个现代化智能 AI 平台，集成多模型能力，提供个性化助手市场与智能工具集。',
    icon: Platform,
    tags: ['Go', 'Vue 3', 'AI Agent'],
    link: 'https://github.com/lemon-puls/txing-ai',
    badge: '旗舰项目',
    highlights: [
      '实现多模型动态路由与统一 API 网关',
      '构建可扩展的 AI 助手预设系统',
      '优化 SSE 流式响应，提升交互实时感'
    ],
    media: [
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=modern+AI+platform+dashboard+dark+theme+purple+gradient&image_size=landscape_16_9', caption: '平台主界面' },
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=AI+chat+interface+with+streaming+response+modern+UI&image_size=landscape_16_9', caption: '智能对话界面' },
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=workflow+editor+node+based+visual+programming&image_size=landscape_16_9', caption: '可视化工作流编辑器' }
    ],
    techStack: [
      { name: 'Go', icon: '🔧' },
      { name: 'Vue 3', icon: '⚡' },
      { name: 'MySQL', icon: '🗄️' },
      { name: 'Redis', icon: '💾' },
      { name: 'SSE', icon: '📡' },
      { name: 'Docker', icon: '🐳' }
    ],
    features: [
      { icon: '🤖', title: '多模型支持', desc: '集成 OpenAI、Claude 等主流 LLM，智能路由分发' },
      { icon: '⚡', title: '实时流式响应', desc: 'SSE 长连接，打字机效果，极致交互体验' },
      { icon: '🔧', title: '可视化工作流', desc: '拖拽式 Agent 编排，支持复杂业务逻辑' },
      { icon: '🏪', title: '助手市场', desc: '丰富的预设模板，一键创建专属 AI 助手' }
    ]
  },
  {
    name: 'AI 简历优化系统',
    desc: '基于 Agent 架构的简历分析与优化工具，帮助用户针对目标岗位提升竞争力。',
    icon: Document,
    tags: ['Agent', 'PDF Processing', 'RAG'],
    link: '#',
    badge: '创新项目',
    highlights: [
      '设计多步工作流：分析 → 优化 → 生成',
      '支持 PDF/Word 文件的智能解析与重构',
      '提供详细的 AI 推理过程可视化'
    ],
    media: [
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=resume+analysis+AI+dashboard+modern+clean+design&image_size=landscape_16_9', caption: '简历分析界面' },
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=AI+workflow+visualization+steps+arrows+modern+UI&image_size=landscape_16_9', caption: '优化工作流' }
    ],
    techStack: [
      { name: 'Go', icon: '🔧' },
      { name: 'Vue 3', icon: '⚡' },
      { name: 'PDF.js', icon: '📄' },
      { name: 'LLM API', icon: '🤖' }
    ],
    features: [
      { icon: '📄', title: '智能解析', desc: '支持 PDF、Word 等多种格式，精准提取简历内容' },
      { icon: '🎯', title: '岗位匹配', desc: '根据目标岗位要求，智能分析简历匹配度' },
      { icon: '✨', title: 'AI 优化', desc: '自动生成优化建议，一键应用改进方案' },
      { icon: '📊', title: '可视化报告', desc: '详细的分析报告，直观展示优化效果' }
    ]
  },
  {
    name: '智能客服助手',
    desc: '基于 RAG 技术的智能客服系统，支持多轮对话、知识库检索和意图识别。',
    icon: ChatDotRound,
    tags: ['RAG', 'NLP', 'WebSocket'],
    link: '#',
    badge: '企业级',
    highlights: [
      '实现基于向量检索的知识库问答',
      '支持多轮对话上下文理解',
      '实时 WebSocket 消息推送'
    ],
    media: [
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=AI+customer+service+chat+interface+modern+dark+theme&image_size=landscape_16_9', caption: '客服对话界面' },
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=knowledge+base+management+dashboard+analytics&image_size=landscape_16_9', caption: '知识库管理' }
    ],
    techStack: [
      { name: 'Go', icon: '🔧' },
      { name: 'Vue 3', icon: '⚡' },
      { name: 'WebSocket', icon: '🔌' },
      { name: 'Milvus', icon: '🧠' }
    ],
    features: [
      { icon: '💬', title: '多轮对话', desc: '支持上下文理解，连续对话不丢失语境' },
      { icon: '🔍', title: '知识检索', desc: '基于向量相似度的精准知识库检索' },
      { icon: '📊', title: '数据分析', desc: '对话数据统计分析，优化服务质量' },
      { icon: '🔌', title: 'API 集成', desc: '提供标准 RESTful API，轻松接入现有系统' }
    ]
  },
  {
    name: '数据可视化平台',
    desc: '企业级数据分析与可视化平台，支持多数据源接入和自定义仪表盘。',
    icon: DataLine,
    tags: ['ECharts', 'Data Analysis', 'Dashboard'],
    link: '#',
    badge: '数据驱动',
    highlights: [
      '支持 10+ 种图表类型',
      '拖拽式仪表盘构建',
      '实时数据流更新'
    ],
    media: [
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=data+visualization+dashboard+charts+graphs+modern+dark+theme&image_size=landscape_16_9', caption: '数据仪表盘' },
      { type: 'image', url: 'https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=drag+drop+dashboard+builder+interface+modern+UI&image_size=landscape_16_9', caption: '仪表盘编辑器' }
    ],
    techStack: [
      { name: 'Vue 3', icon: '⚡' },
      { name: 'ECharts', icon: '📊' },
      { name: 'Pinia', icon: '🍍' },
      { name: 'Tailwind', icon: '🎨' }
    ],
    features: [
      { icon: '📊', title: '丰富图表', desc: '支持折线图、柱状图、饼图等 10+ 种图表' },
      { icon: '🎛️', title: '拖拽构建', desc: '可视化拖拽编辑器，零代码构建仪表盘' },
      { icon: '⚡', title: '实时更新', desc: 'WebSocket 推送，数据变化即时呈现' },
      { icon: '📱', title: '响应式', desc: '完美适配桌面端和移动端' }
    ]
  }
]

const timeline = [
  {
    time: '2024 - 至今',
    title: '全栈研发 / 独立开发者',
    desc: '主导 Txing AI 项目从 0 到 1 的开发，负责从数据库设计到前端 UI 的全链路实现。',
    tags: ['Go', 'Vue 3', 'AI Agent', 'Docker']
  },
  {
    time: '2023 - 2024',
    title: '技术钻研期',
    desc: '深耕 Go 语言后端生态，探索 Vue 3 组合式 API，并开始研究大语言模型应用开发。',
    tags: ['Go', 'Vue 3', 'LLM']
  },
  {
    time: '2022 - 2023',
    title: '全栈成长期',
    desc: '系统学习前后端技术栈，完成多个全栈项目，积累实战经验。',
    tags: ['JavaScript', 'Node.js', 'React']
  }
]

// Project expand state
const expandedProject = ref(null)
const toggleProject = (index) => {
  expandedProject.value = expandedProject.value === index ? null : index
}

// Media preview
const mediaPreviewVisible = ref(false)
const currentMedia = ref(null)
const openMediaPreview = (media) => {
  currentMedia.value = media
  mediaPreviewVisible.value = true
}

// Particle styles
const getParticleStyle = (index) => {
  const size = Math.random() * 6 + 2
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    animationDelay: `${Math.random() * 10}s`,
    animationDuration: `${Math.random() * 20 + 10}s`
  }
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
  background-color: var(--el-bg-color-page, #f8fafc);
  color: var(--el-text-color-primary);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  overflow-x: hidden;
}

// Particle Background
.particle-bg {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.particle {
  position: absolute;
  background: var(--el-color-primary-light-5);
  border-radius: 50%;
  opacity: 0.3;
  animation: float-particle linear infinite;
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
      background: linear-gradient(45deg, #10b981, #34d399);
      color: white;
      padding: 8px 18px;
      border-radius: 24px;
      font-size: 13px;
      font-weight: 600;
      white-space: nowrap;
      box-shadow: 0 4px 15px rgba(16, 185, 129, 0.4);
      display: flex;
      align-items: center;
      gap: 8px;
      animation: badge-bounce 2s ease-in-out infinite;

      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: white;
        animation: pulse-dot 2s infinite;
      }
    }
  }

  .hero-title {
    font-size: 60px;
    font-weight: 900;
    margin: 0;
    color: var(--el-text-color-primary, #0f172a);
    letter-spacing: -1px;

    .highlight {
      background: linear-gradient(45deg, var(--el-color-primary), #6366f1, #a855f7);
      background-size: 200% 200%;
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      animation: gradient-shift 3s ease infinite;
    }

    .cursor {
      display: inline-block;
      -webkit-text-fill-color: var(--el-color-primary);
      animation: blink 1s step-end infinite;
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
      transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
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
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
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
  background: var(--el-bg-color-page);
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
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
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
  background: var(--el-bg-color-page);
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
  gap: 30px;

  @media (max-width: 500px) {
    grid-template-columns: 1fr;
  }
}

.project-card {
  background: var(--el-bg-color);
  border-radius: 24px;
  overflow: visible;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.4s cubic-bezier(0.4, 0, 0.2, 1), transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;

  &:hover {
    transform: translateY(-10px);
    box-shadow: 0 24px 50px rgba(43, 94, 255, 0.15);
  }

  &.expanded {
    box-shadow: 0 24px 50px rgba(43, 94, 255, 0.2);
  }

  .project-image {
    height: 220px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    cursor: pointer;

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
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 12px;

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
    border-top: 1px solid var(--el-border-color-lighter);
    padding: 28px;
    background: var(--el-fill-color-lighter);
    border-radius: 0 0 24px 24px;

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
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 16px;

      .feature-item {
        display: flex;
        gap: 14px;
        background: var(--el-bg-color);
        padding: 18px;
        border-radius: 16px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
        transition: all 0.2s;

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .feature-icon {
          font-size: 28px;
          flex-shrink: 0;
        }

        h5 {
          font-size: 15px;
          font-weight: 700;
          margin: 0 0 6px 0;
          color: var(--el-text-color-primary);
        }

        p {
          font-size: 13px;
          color: var(--el-text-color-secondary);
          margin: 0;
          line-height: 1.5;
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
  background: var(--el-bg-color-page);
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
      transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
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
  50% { transform: translateX(-50%) translateY(-5px); }
}

@keyframes rotate-gradient {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
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
    grid-template-columns: 1fr;
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
    .project-detail {
      .features-grid {
        grid-template-columns: 1fr;
      }
    }
  }
}
</style>
