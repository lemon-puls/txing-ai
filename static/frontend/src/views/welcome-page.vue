<template>
  <div class="welcome-container">
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
        <div class="subtitle">智能助手，助您提升效率</div>
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
import { ChatRound, Shop, ArrowRight } from '@element-plus/icons-vue'

const router = useRouter()
const hoveredCard = ref(null)
const particles = ref([])

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

onMounted(() => {
  initParticles()
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

.content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1200px;
  padding: 20px;
  text-align: center;
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
  height: 400px;
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