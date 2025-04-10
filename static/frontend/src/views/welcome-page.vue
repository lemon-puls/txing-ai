<template>
  <div class="welcome-container">
    <!-- 顶部导航栏 -->
    <div class="nav-header">
      <div class="nav-content">
        <div class="nav-left">
          <div class="logo">
            <span class="logo-text">Txing AI</span>
          </div>
        </div>
        <div class="nav-right">
          <a href="https://github.com/yourusername/txing-ai" target="_blank" class="github-link">
            <el-tooltip content="在 GitHub 上查看" placement="bottom">
              <el-icon class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg></el-icon>
            </el-tooltip>
          </a>
          <div class="user-info">
            <el-dropdown trigger="click">
              <div class="user-avatar">
                <el-avatar :size="32" src="https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png" />
                <el-icon class="el-icon--right"><CaretBottom /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item>
                    <el-icon><User /></el-icon>
                    <span>个人中心</span>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-icon><Setting /></el-icon>
                    <span>设置</span>
                  </el-dropdown-item>
                  <el-dropdown-item divided>
                    <el-icon><SwitchButton /></el-icon>
                    <span>退出登录</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>

    <!-- 动态背景 -->
    <div class="animated-background">
      <div class="light"></div>
    </div>

    <!-- 粒子动画背景 -->
    <div class="particles">
      <div
        v-for="particle in particles"
        :key="particle.id"
        class="particle"
        :style="particle.style"
      ></div>
    </div>

    <!-- 主要内容 -->
    <div class="content">
      <h1 class="main-title">
        <span class="gradient-text">AI Assistant</span>
        <div class="subtitle">
          <span class="typing-text"></span>
          <span class="cursor">|</span>
        </div>
      </h1>

      <div class="cards-container">
        <!-- 聊天入口 -->
        <div
          class="entrance-card chat-card"
          :class="{ 'card-hover': hoveredCard === 'chat' }"
          @click="navigateToChat"
          @mouseenter="handleHover('chat')"
          @mouseleave="handleHover(null)"
        >
          <div class="card-content">
            <el-icon class="card-icon"><ChatRound /></el-icon>
            <h2>开启聊天</h2>
            <p>与 AI 助手展开对话，探索无限可能</p>
            <div class="card-action">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>
          <div class="card-overlay"></div>
        </div>

        <!-- 市场入口 -->
        <div
          class="entrance-card market-card"
          :class="{ 'card-hover': hoveredCard === 'market' }"
          @click="navigateToMarket"
          @mouseenter="handleHover('market')"
          @mouseleave="handleHover(null)"
        >
          <div class="card-content">
            <el-icon class="card-icon"><Shop /></el-icon>
            <h2>AI 助手市场</h2>
            <p>发现更多专业 AI 助手，提升工作效率</p>
            <div class="card-action">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>
          <div class="card-overlay"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup name="WelcomePage">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  ChatRound, 
  Shop, 
  ArrowRight, 
  User, 
  Setting, 
  SwitchButton,
  CaretBottom 
} from '@element-plus/icons-vue'

const router = useRouter()
const hoveredCard = ref(null)
const particles = ref([])

// 打字效果相关变量
const text = '智能助手，助您提升效率'
const typingSpeed = 150 // 打字速度（毫秒）
const eraseSpeed = 100  // 删除速度（毫秒）
const delayBetweenLoops = 2000 // 循环之间的延迟（毫秒）
let currentText = ''
let isTyping = true
let currentIndex = 0

// 生成随机数在指定范围内
const random = (min, max) => Math.random() * (max - min) + min

// 初始化粒子
const initParticles = () => {
  const particlesArray = []
  for (let i = 0; i < 20; i++) {
    particlesArray.push({
      id: i,
      style: {
        width: `${random(1, 4)}px`,
        height: `${random(1, 4)}px`,
        left: `${random(0, 100)}%`,
        top: `${random(0, 100)}%`,
        animationDelay: `${random(-8000, 0)}ms`,
        animationDuration: `${random(3000, 15000)}ms`,
        opacity: random(0.1, 0.5)
      }
    })
  }
  particles.value = particlesArray
}

// 打字效果函数
const typeText = () => {
  const typingElement = document.querySelector('.typing-text')
  if (!typingElement) return

  if (isTyping) {
    if (currentIndex < text.length) {
      currentText += text[currentIndex]
      typingElement.textContent = currentText
      currentIndex++
      setTimeout(typeText, typingSpeed)
    } else {
      isTyping = false
      setTimeout(typeText, delayBetweenLoops)
    }
  } else {
    if (currentText.length > 0) {
      currentText = currentText.slice(0, -1)
      typingElement.textContent = currentText
      setTimeout(typeText, eraseSpeed)
    } else {
      isTyping = true
      currentIndex = 0
      setTimeout(typeText, delayBetweenLoops)
    }
  }
}

onMounted(() => {
  initParticles()
  typeText() // 启动打字效果
})

const handleHover = (card) => {
  hoveredCard.value = card
}

const navigateToChat = () => {
  router.push('/chat')
}

const navigateToMarket = () => {
  router.push('/home')
}
</script>

<style scoped lang="scss">
.welcome-container {
  min-height: 100vh;
  width: 100%;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #0f0f1a;
  overflow: hidden;
}

// 导航栏样式
.nav-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  z-index: 100;
  
  .nav-content {
    max-width: 1200px;
    height: 100%;
    margin: 0 auto;
    padding: 0 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    .nav-left {
      display: flex;
      align-items: center;
      
      .logo {
        display: flex;
        align-items: center;
        gap: 8px;
        
        .logo-text {
          font-size: 24px;
          font-weight: bold;
          background: linear-gradient(45deg, #4158D0, #C850C0);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }
      }
    }

    .nav-right {
      display: flex;
      align-items: center;
      gap: 24px;

      .github-link {
        display: flex;
        align-items: center;
        color: rgba(255, 255, 255, 0.7);
        transition: color 0.3s ease;

        &:hover {
          color: var(--el-color-primary);
        }

        .nav-icon {
          font-size: 24px;
        }
      }

      .user-info {
        .user-avatar {
          display: flex;
          align-items: center;
          gap: 4px;
          cursor: pointer;
          padding: 2px;
          border-radius: 50%;
          transition: background-color 0.3s ease;

          &:hover {
            background: rgba(255, 255, 255, 0.1);
          }

          .el-icon--right {
            font-size: 12px;
            color: rgba(255, 255, 255, 0.7);
          }
        }
      }
    }
  }
}

// 调整内容区域的上边距，为导航栏留出空间
.content {
  padding-top: 64px;
}

// 动态背景
.animated-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;

  .light {
    position: absolute;
    width: 150vmax;
    height: 150vmax;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: radial-gradient(
      circle,
      rgba(65, 88, 208, 0.15) 0%,
      rgba(200, 80, 192, 0.15) 30%,
      rgba(255, 204, 112, 0.15) 70%
    );
    animation: rotate 20s linear infinite;
  }
}

// 粒子动画
.particles {
  position: absolute;
  width: 100%;
  height: 100%;

  .particle {
    position: absolute;
    background: rgba(255, 255, 255, 0.5);
    border-radius: 50%;
    animation: float 8s infinite;
  }
}

.main-title {
  margin-bottom: 60px;

  .gradient-text {
    font-size: 4em;
    font-weight: bold;
    background: linear-gradient(45deg, #4158D0, #C850C0, #FFCC70);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    animation: hue-rotate 6s linear infinite;
  }

  .subtitle {
    margin-top: 20px;
    font-size: 1.5em;
    color: rgba(255, 255, 255, 0.7);
    font-weight: normal;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 2px;

    .cursor {
      display: inline-block;
      animation: blink 1s infinite;
      color: var(--el-color-primary);
      font-weight: 200;
    }
  }
}

.cards-container {
  display: flex;
  justify-content: center;
  gap: 40px;
  flex-wrap: wrap;
  perspective: 1000px;
}

.entrance-card {
  position: relative;
  width: 300px;
  height: 320px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  backdrop-filter: blur(10px);
  cursor: pointer;
  overflow: hidden;
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  transform-style: preserve-3d;

  &.card-hover {
    transform: translateY(-10px) rotateX(5deg) rotateY(5deg);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);

    .card-overlay {
      opacity: 1;
    }

    .card-icon {
      transform: translateY(-10px) scale(1.1);
    }

    .card-action {
      transform: translateX(5px);
      opacity: 1;
    }
  }

  .card-content {
    position: relative;
    z-index: 2;
    height: 100%;
    padding: 30px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: white;
    text-align: center;
  }

  .card-icon {
    font-size: 64px;
    margin-bottom: 20px;
    transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
    color: #4158D0;
  }

  h2 {
    font-size: 24px;
    margin: 20px 0;
    background: linear-gradient(45deg, #4158D0, #C850C0);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  p {
    font-size: 16px;
    color: rgba(255, 255, 255, 0.7);
    line-height: 1.6;
    margin-bottom: 20px;
  }

  .card-action {
    font-size: 24px;
    color: #C850C0;
    opacity: 0;
    transition: all 0.3s ease;
  }

  .card-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      45deg,
      rgba(65, 88, 208, 0.1) 0%,
      rgba(200, 80, 192, 0.1) 50%,
      rgba(255, 204, 112, 0.1) 100%
    );
    opacity: 0;
    transition: opacity 0.5s ease;
  }

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    pointer-events: none;
  }
}

// 动画关键帧
@keyframes rotate {
  from {
    transform: translate(-50%, -50%) rotate(0deg);
  }
  to {
    transform: translate(-50%, -50%) rotate(360deg);
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) translateX(0);
  }
  50% {
    transform: translateY(-20px) translateX(10px);
  }
}

@keyframes hue-rotate {
  from {
    filter: hue-rotate(0deg);
  }
  to {
    filter: hue-rotate(360deg);
  }
}

// 添加光标闪烁动画
@keyframes blink {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0;
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  .main-title {
    margin-bottom: 40px;

    .gradient-text {
      font-size: 2.5em;
    }

    .subtitle {
      font-size: 1.2em;
    }
  }

  .cards-container {
    gap: 20px;
  }

  .entrance-card {
    width: calc(100% - 40px);
    height: 300px;

    .card-icon {
      font-size: 48px;
    }

    h2 {
      font-size: 20px;
      margin: 15px 0;
    }

    p {
      font-size: 14px;
    }
  }
}
</style>
