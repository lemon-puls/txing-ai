<template>
  <div class="websites-container">

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
      <!-- 页面标题 -->
      <div class="page-header">
        <h1 class="page-title">
          <div class="subtitle">精选实用网站，提升工作效率</div>
        </h1>
      </div>

      <!-- 搜索和筛选区域 -->
      <div class="search-section">
        <div class="search-container">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索网站名称或描述..."
            class="search-input"
            :prefix-icon="Search"
            clearable
            @input="handleSearch"
          />
        </div>

        <!-- 标签筛选 -->
        <div class="tags-filter">
          <div class="filter-label">标签筛选：</div>
          <div class="tags-container">
            <el-tag
              v-for="tag in allTags"
              :key="tag"
              :type="selectedTags.includes(tag) ? 'primary' : ''"
              :effect="selectedTags.includes(tag) ? 'dark' : 'plain'"
              class="tag-item"
              @click="toggleTag(tag)"
              style="color: rgba(0, 0, 0, 0.7)"
            >
              {{ tag }}
            </el-tag>
          </div>
        </div>
      </div>

      <!-- 网站列表 -->
      <div class="websites-grid" v-loading="loading">
        <div
          v-for="website in filteredWebsites"
          :key="website.id"
          class="website-card"
          @click="openWebsite(website.url)"
        >
          <div class="card-header">
            <div class="website-avatar">
              <img
                v-if="website.avatar"
                :src="website.avatar"
                :alt="website.name"
                @error="handleImageError"
              />
              <el-icon v-else class="default-icon"><Link /></el-icon>
            </div>
            <div class="website-info">
              <h3 class="website-name">{{ website.name }}</h3>
              <p class="website-description">{{ website.description }}</p>
            </div>
          </div>

          <div class="card-footer">
            <div class="website-tags">
              <el-tag
                v-for="tag in website.tags"
                :key="tag"
                size="small"
                type="info"
                effect="plain"
                style="color: rgba(0, 0, 0, 0.7)"
              >
                {{ tag }}
              </el-tag>
            </div>
            <div class="visit-btn">
              <el-icon><Link /></el-icon>
            </div>
          </div>

          <div class="card-overlay"></div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && filteredWebsites.length === 0" class="empty-state">
        <el-empty description="暂无相关网站" />
      </div>
    </div>
  </div>
</template>

<script setup name="WebsitesPage">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowLeft,
  Search,
  Link
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const searchKeyword = ref('')
const selectedTags = ref([])
const websites = ref([])
const particles = ref([])

import { defaultApi } from '@/api'
import {useThemeStore} from "@/stores/theme.js";
import {useUserStore} from "@/stores/user.js";


// 计算所有标签
const allTags = computed(() => {
  const tags = new Set()
  websites.value.forEach(website => {
    // 以 , 分割 website.tags
    website.tags.split(',').forEach(tag => {
      tags.add(tag.trim())
    })
  })
  return Array.from(tags)
})

// 过滤后的网站列表
const filteredWebsites = computed(() => {
  let result = websites.value

  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(website =>
      website.name.toLowerCase().includes(keyword) ||
      website.description.toLowerCase().includes(keyword)
    )
  }

  // 标签筛选
  if (selectedTags.value.length > 0) {
    result = result.filter(website =>
      selectedTags.value.some(tag => website.tags.includes(tag))
    )
  }

  return result
})

// 生成随机数在指定范围内
const random = (min, max) => Math.random() * (max - min) + min

// 初始化粒子
const initParticles = () => {
  const particlesArray = []
  for (let i = 0; i < 15; i++) {
    particlesArray.push({
      id: i,
      style: {
        width: `${random(1, 3)}px`,
        height: `${random(1, 3)}px`,
        left: `${random(0, 100)}%`,
        top: `${random(0, 100)}%`,
        animationDelay: `${random(-8000, 0)}ms`,
        animationDuration: `${random(5000, 20000)}ms`,
        opacity: random(0.1, 0.3)
      }
    })
  }
  particles.value = particlesArray
}

// 加载网站数据
const loadWebsites = async () => {
  loading.value = true
  try {
    const response = await defaultApi.apiWebsitesListGet(
      1, // page
      100, // pageSize - 获取所有数据
      {
        name: searchKeyword.value || undefined,
        tag: selectedTags.value.length > 0 ? selectedTags.value[0] : undefined
      }
    )

    if (response.code === 0) {
      websites.value = response.data.records || []
    } else {
      ElMessage.error(response.message || '加载网站列表失败')
    }
  } catch (error) {
    console.error('加载网站列表失败:', error)
    ElMessage.error('加载网站列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在computed中处理
}

// 标签切换
const toggleTag = (tag) => {
  const index = selectedTags.value.indexOf(tag)
  if (index > -1) {
    selectedTags.value.splice(index, 1)
  } else {
    selectedTags.value.push(tag)
  }
}

// 打开网站
const openWebsite = (url) => {
  window.open(url, '_blank')
}

// 图片加载错误处理
const handleImageError = (event) => {
  event.target.style.display = 'none'
}

// 返回上一页
const goBack = () => {
  router.back()
}

const themeStore = useThemeStore()
// 存储进入页面时的主题状态
const previousThemeState = ref(null)
const previousTheme = ref("'#409EFF'")

onMounted(() => {
  initParticles()
  loadWebsites()

  // 保存当前主题状态
  previousThemeState.value = themeStore.isDark
  previousTheme.value = themeStore.primaryColor
  themeStore.setPrimaryColor('#409EFF')
  // 如果当前是暗色主题，切换为明亮主题
  if (themeStore.isDark) {
    themeStore.toggleTheme()
  }
})
</script>

<style scoped lang="scss">
.websites-container {
  min-height: 100vh;
  width: 100%;
  position: relative;
  background: #f5f7fa;
  overflow-x: hidden;
}

// 导航栏样式
.nav-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  z-index: 100;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.8);

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
          background: linear-gradient(45deg, #2B5EFF, #1E88E5);
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
        color: rgba(0, 0, 0, 0.7);
        transition: color 0.3s ease;

        &:hover {
          color: var(--el-color-primary);
        }

        .nav-icon {
          font-size: 24px;
        }
      }
    }
  }
}

// 动态背景
.animated-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  z-index: 1;

  .light {
    position: absolute;
    width: 150vmax;
    height: 150vmax;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: radial-gradient(
      circle,
      rgba(43, 94, 255, 0.05) 0%,
      rgba(30, 136, 229, 0.05) 30%,
      rgba(3, 169, 244, 0.05) 70%
    );
    animation: rotate 30s linear infinite;
  }
}

// 粒子动画
.particles {
  position: fixed;
  width: 100%;
  height: 100%;
  z-index: 2;

  .particle {
    position: absolute;
    background: linear-gradient(45deg, rgba(43, 94, 255, 0.2), rgba(3, 169, 244, 0.2));
    border-radius: 50%;
    animation: float 10s infinite;
  }
}

// 主要内容
.content {
  position: relative;
  z-index: 10;
  padding: 40px 24px 40px;
  max-width: 1200px;
  margin: 0 auto;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;

  .page-title {
    .subtitle {
      margin-top: 10px;
      font-size: 1.2em;
      color: rgba(0, 0, 0, 0.7);
      font-weight: normal;
    }
  }

  .back-btn {
    width: 48px;
    height: 48px;
    background: rgba(255, 255, 255, 0.8);
    border: 1px solid rgba(0, 0, 0, 0.1);
    backdrop-filter: blur(10px);
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 25px rgba(43, 94, 255, 0.3);
    }
  }
}

// 搜索区域
.search-section {
  margin-bottom: 40px;

  .search-container {
    margin-bottom: 20px;

    .search-input {
      max-width: 500px;

      :deep(.el-input__wrapper) {
        background: rgba(255, 255, 255, 0.9);
        border: 1px solid rgba(0, 0, 0, 0.1);
        backdrop-filter: blur(10px);
        border-radius: 12px;
        transition: all 0.3s ease;

        &:hover {
          border-color: rgba(43, 94, 255, 0.5);
        }

        &.is-focus {
          border-color: var(--el-color-primary);
          box-shadow: 0 0 0 2px rgba(43, 94, 255, 0.2);
        }
      }

      :deep(.el-input__inner) {
        color: #333;

        &::placeholder {
          color: rgba(0, 0, 0, 0.4);
        }
      }
    }
  }

  .tags-filter {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;

    .filter-label {
      color: rgba(0, 0, 0, 0.8);
      font-weight: 500;
      white-space: nowrap;
    }

    .tags-container {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;

      .tag-item {
        cursor: pointer;
        transition: all 0.3s ease;
        border-radius: 16px;
        color: rgba(0, 0, 0, 0.7);

        &:hover {
          transform: translateY(-2px);
        }
      }
    }
  }
}

// 网站网格
.websites-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
  margin-bottom: 40px;
}

// 网站卡片
.website-card {
  position: relative;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 16px;
  padding: 24px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(0, 0, 0, 0.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);

  &:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 40px rgba(43, 94, 255, 0.15);
    border-color: rgba(43, 94, 255, 0.3);

    .card-overlay {
      opacity: 1;
    }

    .visit-btn {
      opacity: 1;
      transform: translateX(0);
    }
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 20px;

    .website-avatar {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      overflow: hidden;
      background: rgba(0, 0, 0, 0.05);
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .default-icon {
        font-size: 24px;
        color: rgba(0, 0, 0, 0.6);
      }
    }

    .website-info {
      flex: 1;
      min-width: 0;

      .website-name {
        font-size: 18px;
        font-weight: 600;
        color: #333;
        margin: 0 0 8px 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .website-description {
        font-size: 14px;
        color: rgba(0, 0, 0, 0.7);
        margin: 0;
        line-height: 1.5;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }
    }
  }

  .card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .website-tags {
      display: flex;
      gap: 6px;
      flex-wrap: wrap;
      flex: 1;
    }

    .visit-btn {
      opacity: 0;
      transform: translateX(10px);
      transition: all 0.3s ease;
      color: var(--el-color-primary);
      font-size: 18px;
    }
  }

  .card-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      45deg,
      rgba(43, 94, 255, 0.03) 0%,
      rgba(30, 136, 229, 0.03) 50%,
      rgba(3, 169, 244, 0.03) 100%
    );
    opacity: 0;
    transition: opacity 0.4s ease;
  }
}

// 空状态
.empty-state {
  text-align: center;
  padding: 60px 20px;

  :deep(.el-empty__description p) {
    color: rgba(0, 0, 0, 0.6);
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
    transform: translateY(-30px) translateX(15px);
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
  .content {
    padding: 80px 16px 40px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;

    .page-title {
      .gradient-text {
        font-size: 2.5em;
      }

      .subtitle {
        font-size: 1em;
      }
    }
  }

  .websites-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .search-section {
    .tags-filter {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
  }
}
</style>
