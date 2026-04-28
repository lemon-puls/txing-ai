<template>
  <div class="about-container">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="avatar-wrapper">
          <div class="main-avatar">
            <span class="avatar-text">T</span>
          </div>
          <div class="status-badge">
            <span class="status-dot"></span>
            Ready for New Challenges
          </div>
        </div>
        <h1 class="hero-title">
          你好，我是 <span class="highlight">Txing</span>
        </h1>
        <p class="hero-subtitle">
          全栈开发工程师 / AI 架构爱好者 / 产品极客
        </p>
        <div class="hero-actions">
          <el-button type="primary" size="large" round @click="scrollToProjects" class="action-btn">
            查看我的作品
          </el-button>
          <el-button size="large" round @click="scrollToContact" class="action-btn secondary">
            联系我
          </el-button>
        </div>
      </div>
      <div class="hero-bg-blobs">
        <div class="blob blob-1"></div>
        <div class="blob blob-2"></div>
      </div>
    </section>

    <!-- Skills Section -->
    <section class="section skills-section">
      <div class="section-header">
        <h2 class="section-title">核心能力</h2>
        <div class="title-underline"></div>
      </div>
      <div class="skills-grid">
        <div v-for="skill in skillSets" :key="skill.category" class="skill-card">
          <div class="skill-icon">
            <el-icon><component :is="skill.icon" /></el-icon>
          </div>
          <h3 class="skill-title">{{ skill.category }}</h3>
          <div class="skill-tags">
            <span v-for="tag in skill.tags" :key="tag" class="skill-tag">{{ tag }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Projects Section -->
    <section id="projects" class="section projects-section">
      <div class="section-header">
        <h2 class="section-title">精选作品</h2>
        <div class="title-underline"></div>
      </div>
      <div class="projects-grid">
        <div v-for="(project, index) in projects" :key="project.name" class="project-card">
          <div class="project-image" :class="`project-gradient-${index + 1}`">
            <div class="project-icon">
              <el-icon><component :is="project.icon" /></el-icon>
            </div>
            <div class="project-overlay">
              <a :href="project.link" target="_blank" class="overlay-link">
                <span>访问项目</span>
                <el-icon><ArrowRight /></el-icon>
              </a>
            </div>
          </div>
          <div class="project-info">
            <div class="project-tags">
              <span v-for="tag in project.tags" :key="tag" class="project-tag">{{ tag }}</span>
            </div>
            <h3 class="project-name">{{ project.name }}</h3>
            <p class="project-desc">{{ project.desc }}</p>
            <div class="project-highlights">
              <div v-for="hl in project.highlights" :key="hl" class="highlight-item">
                <el-icon><CircleCheck /></el-icon>
                <span>{{ hl }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Timeline Section -->
    <section class="section timeline-section">
      <div class="section-header">
        <h2 class="section-title">成长轨迹</h2>
        <div class="title-underline"></div>
      </div>
      <div class="timeline">
        <div v-for="(item, index) in timeline" :key="index" class="timeline-item">
          <div class="timeline-dot">
            <div class="dot-inner"></div>
          </div>
          <div class="timeline-content">
            <div class="timeline-time">{{ item.time }}</div>
            <h3 class="timeline-title">{{ item.title }}</h3>
            <p class="timeline-desc">{{ item.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Contact Section -->
    <section id="contact" class="section contact-section">
      <div class="contact-card">
        <div class="contact-bg-gradient"></div>
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
  </div>
</template>

<script setup name="AboutPage">
import { ref } from 'vue'
import {
  ArrowRight,
  CircleCheck,
  Message,
  Cpu,
  Monitor,
  Promotion,
  Platform,
  Document
} from '@element-plus/icons-vue'

const Github = {
  template: `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg>`
}

const skillSets = [
  {
    category: '后端开发',
    icon: Cpu,
    tags: ['Go / Gin', 'MySQL', 'Redis', 'GORM', 'RESTful API', 'Microservices']
  },
  {
    category: '前端开发',
    icon: Monitor,
    tags: ['Vue 3', 'Element Plus', 'TypeScript', 'Vite', 'Pinia', 'SCSS']
  },
  {
    category: 'AI & 创新',
    icon: Promotion,
    tags: ['LLM Prompting', 'RAG', 'Agent Orchestration', 'MCP Protocol', 'SSE']
  }
]

const projects = [
  {
    name: 'Txing AI',
    desc: '一个现代化智能 AI 平台，集成多模型能力，提供个性化助手市场与智能工具集。',
    icon: Platform,
    tags: ['Go', 'Vue 3', 'AI Agent'],
    link: 'https://github.com/lemon-puls/txing-ai',
    highlights: [
      '实现多模型动态路由与统一 API 网关',
      '构建可扩展的 AI 助手预设系统',
      '优化 SSE 流式响应，提升交互实时感'
    ]
  },
  {
    name: 'AI 简历优化系统',
    desc: '基于 Agent 架构的简历分析与优化工具，帮助用户针对目标岗位提升竞争力。',
    icon: Document,
    tags: ['Agent', 'PDF Processing', 'RAG'],
    link: '#',
    highlights: [
      '设计多步工作流：分析 → 优化 → 生成',
      '支持 PDF/Word 文件的智能解析与重构',
      '提供详细的 AI 推理过程可视化'
    ]
  }
]

const timeline = [
  {
    time: '2024 - 至今',
    title: '全栈研发 / 独立开发者',
    desc: '主导 Txing AI 项目从 0 到 1 的开发，负责从数据库设计到前端 UI 的全链路实现。'
  },
  {
    time: '2023 - 2024',
    title: '技术钻研期',
    desc: '深耕 Go 语言后端生态，探索 Vue 3 组合式 API，并开始研究大语言模型应用开发。'
  }
]

const scrollToProjects = () => {
  document.getElementById('projects')?.scrollIntoView({ behavior: 'smooth' })
}

const scrollToContact = () => {
  document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<style scoped lang="scss">
.about-container {
  width: 100%;
  background-color: var(--el-bg-color-page, #f8fafc);
  color: var(--el-text-color-primary);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.section {
  padding: 80px 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.section-header {
  text-align: center;
  margin-bottom: 60px;

  .section-title {
    font-size: 32px;
    font-weight: 800;
    margin-bottom: 12px;
    color: var(--el-text-color-primary, #1e293b);
  }

  .title-underline {
    width: 60px;
    height: 4px;
    background: linear-gradient(90deg, var(--el-color-primary), #6366f1);
    margin: 0 auto;
    border-radius: 2px;
  }
}

/* Hero Section */
.hero-section {
  position: relative;
  min-height: 80vh;
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
    margin-bottom: 10px;

    .main-avatar {
      width: 120px;
      height: 120px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--el-color-primary), #6366f1);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 10px 30px rgba(43, 94, 255, 0.3);
      transition: transform 0.3s ease;
      border: 4px solid white;

      &:hover {
        transform: scale(1.05);
      }

      .avatar-text {
        font-size: 48px;
        font-weight: 800;
        color: white;
      }
    }

    .status-badge {
      position: absolute;
      bottom: -10px;
      left: 50%;
      transform: translateX(-50%);
      background: linear-gradient(45deg, #10b981, #34d399);
      color: white;
      padding: 6px 14px;
      border-radius: 20px;
      font-size: 12px;
      font-weight: 600;
      white-space: nowrap;
      box-shadow: 0 4px 10px rgba(16, 185, 129, 0.3);
      display: flex;
      align-items: center;
      gap: 6px;

      .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: white;
        animation: pulse-dot 2s infinite;
      }
    }
  }

  .hero-title {
    font-size: 56px;
    font-weight: 900;
    margin: 0;
    color: var(--el-text-color-primary, #0f172a);
    letter-spacing: -1px;

    .highlight {
      background: linear-gradient(45deg, var(--el-color-primary), #6366f1);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
  }

  .hero-subtitle {
    font-size: 20px;
    color: var(--el-text-color-secondary, #64748b);
    max-width: 600px;
    line-height: 1.6;
  }

  .hero-actions {
    display: flex;
    gap: 16px;
    margin-top: 10px;

    .action-btn {
      padding: 12px 32px;
      font-weight: 600;
      font-size: 16px;
      transition: all 0.3s ease;

      &.secondary {
        background-color: var(--el-bg-color);
        border: 1px solid var(--el-border-color);
        color: var(--el-text-color-regular);

        &:hover {
          background-color: var(--el-fill-color-light);
          border-color: var(--el-border-color-darker);
        }
      }

      &:hover {
        transform: translateY(-3px);
        box-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
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
      filter: blur(80px);
      opacity: 0.4;
      border-radius: 50%;
      animation: float-blob 10s infinite alternate ease-in-out;
    }

    .blob-1 {
      width: 400px;
      height: 400px;
      background: var(--el-color-primary-light-5);
      top: 10%;
      left: 15%;
    }

    .blob-2 {
      width: 300px;
      height: 300px;
      background: #c7d2fe;
      bottom: 15%;
      right: 15%;
      animation-delay: -5s;
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
    transition: all 0.3s ease;
    border: 1px solid var(--el-border-color-lighter);

    &:hover {
      transform: translateY(-5px);
      box-shadow: 0 12px 30px rgba(43, 94, 255, 0.1);
      border-color: var(--el-color-primary-light-5);
    }

    .skill-icon {
      width: 56px;
      height: 56px;
      border-radius: 16px;
      background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-color-primary-light-8));
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 20px;

      .el-icon {
        font-size: 28px;
        color: var(--el-color-primary);
      }
    }

    .skill-title {
      font-size: 20px;
      font-weight: 700;
      margin: 0 0 16px 0;
      color: var(--el-text-color-primary);
    }

    .skill-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;

      .skill-tag {
        background-color: var(--el-fill-color-light);
        color: var(--el-text-color-regular);
        padding: 6px 12px;
        border-radius: 8px;
        font-size: 13px;
        font-weight: 500;
        transition: all 0.2s;

        &:hover {
          background-color: var(--el-color-primary-light-9);
          color: var(--el-color-primary);
        }
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
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 30px;

  .project-card {
    background: var(--el-bg-color);
    border-radius: 24px;
    overflow: hidden;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
    display: flex;
    flex-direction: column;

    &:hover {
      transform: translateY(-10px);
      box-shadow: 0 20px 40px rgba(43, 94, 255, 0.15);

      .project-overlay {
        opacity: 1;
      }

      .project-icon {
        transform: scale(1.1);
      }
    }

    .project-image {
      height: 200px;
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      overflow: hidden;

      &.project-gradient-1 {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }

      &.project-gradient-2 {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }

      .project-icon {
        width: 80px;
        height: 80px;
        background: rgba(255, 255, 255, 0.2);
        backdrop-filter: blur(10px);
        border-radius: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: transform 0.3s ease;

        .el-icon {
          font-size: 40px;
          color: white;
        }
      }

      .project-overlay {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(15, 23, 42, 0.7);
        backdrop-filter: blur(4px);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transition: opacity 0.3s ease;

        .overlay-link {
          display: flex;
          align-items: center;
          gap: 8px;
          background: white;
          color: var(--el-color-primary);
          padding: 12px 24px;
          border-radius: 12px;
          font-weight: 600;
          text-decoration: none;
          transition: transform 0.2s ease;

          &:hover {
            transform: scale(1.05);
          }
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
          padding: 4px 10px;
          border-radius: 6px;
          font-weight: 600;
        }
      }

      .project-name {
        font-size: 22px;
        font-weight: 800;
        margin: 0;
        color: var(--el-text-color-primary);
      }

      .project-desc {
        color: var(--el-text-color-secondary);
        line-height: 1.6;
        margin: 0;
        font-size: 14px;
      }

      .project-highlights {
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin-top: 8px;

        .highlight-item {
          display: flex;
          align-items: flex-start;
          gap: 8px;
          font-size: 13px;
          color: var(--el-text-color-regular);

          .el-icon {
            color: #10b981;
            font-size: 16px;
            margin-top: 2px;
            flex-shrink: 0;
          }
        }
      }
    }
  }
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
    width: 2px;
    background: linear-gradient(180deg, var(--el-color-primary), var(--el-color-primary-light-3));
    border-radius: 2px;
  }

  .timeline-item {
    position: relative;
    padding-left: 60px;
    margin-bottom: 32px;

    &:last-child {
      margin-bottom: 0;
    }

    .timeline-dot {
      position: absolute;
      left: 12px;
      top: 0;
      width: 18px;
      height: 18px;
      background: var(--el-bg-color);
      border: 3px solid var(--el-color-primary);
      border-radius: 50%;
      z-index: 2;
      display: flex;
      align-items: center;
      justify-content: center;

      .dot-inner {
        width: 6px;
        height: 6px;
        background: var(--el-color-primary);
        border-radius: 50%;
      }
    }

    .timeline-content {
      background: var(--el-bg-color);
      padding: 24px;
      border-radius: 16px;
      box-shadow: 0 4px 15px rgba(0, 0, 0, 0.03);
      transition: all 0.3s ease;
      border: 1px solid var(--el-border-color-lighter);

      &:hover {
        transform: translateX(10px);
        box-shadow: 0 8px 20px rgba(43, 94, 255, 0.08);
        border-color: var(--el-color-primary-light-5);
      }

      .timeline-time {
        font-size: 13px;
        font-weight: 600;
        color: var(--el-color-primary);
        margin-bottom: 8px;
      }

      .timeline-title {
        font-size: 18px;
        font-weight: 700;
        margin: 0 0 8px 0;
        color: var(--el-text-color-primary);
      }

      .timeline-desc {
        font-size: 14px;
        color: var(--el-text-color-secondary);
        line-height: 1.6;
        margin: 0;
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
  padding: 60px 40px;
  border-radius: 32px;
  width: 100%;
  max-width: 800px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.3);
  overflow: hidden;

  .contact-bg-gradient {
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(circle, rgba(43, 94, 255, 0.1) 0%, transparent 70%);
    animation: rotate-gradient 15s linear infinite;
    pointer-events: none;
  }

  h2 {
    position: relative;
    font-size: 32px;
    font-weight: 800;
    margin-bottom: 16px;
    z-index: 1;
  }

  p {
    position: relative;
    font-size: 18px;
    color: #94a3b8;
    margin-bottom: 40px;
    z-index: 1;
  }

  .contact-links {
    position: relative;
    display: flex;
    justify-content: center;
    gap: 20px;
    flex-wrap: wrap;
    z-index: 1;

    .contact-link {
      display: flex;
      align-items: center;
      gap: 8px;
      background: rgba(255, 255, 255, 0.1);
      color: white;
      padding: 14px 28px;
      border-radius: 12px;
      text-decoration: none;
      font-weight: 600;
      transition: all 0.3s ease;
      border: 1px solid rgba(255, 255, 255, 0.1);

      &:hover {
        background: rgba(255, 255, 255, 0.2);
        transform: translateY(-3px);
        border-color: rgba(255, 255, 255, 0.3);
      }
    }
  }
}

/* Animations */
@keyframes float-blob {
  0% { transform: translate(0, 0) scale(1); }
  100% { transform: translate(50px, 50px) scale(1.1); }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes rotate-gradient {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
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

  .section {
    padding: 60px 16px;
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
}
</style>
