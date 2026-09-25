<template>
  <div class="websites-page">
    <!-- 环境光斑背景 -->
    <div class="ambient" aria-hidden="true">
      <div class="blob blob-a"></div>
      <div class="blob blob-b"></div>
    </div>

    <div class="page-body">
      <!-- 头部 -->
      <header class="hero">
        <h1 class="hero-title">网站导航</h1>
        <div class="hero-accent" aria-hidden="true"></div>
        <p class="hero-subtitle">精选优质站点，发现实用工具、AI 与开源资源</p>
      </header>

      <!-- 搜索与分类筛选 -->
      <div class="toolbar">
        <div class="search-row">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索站点名称或描述…"
            class="search-input"
            :prefix-icon="Search"
            clearable
          />
        </div>

        <!-- 内定分类 -->
        <div class="cat-row">
          <button class="pill" :class="{ active: selectedTags.length === 0 }" @click="selectedTags = []">
            全部
          </button>
          <button
            v-for="tag in PRESET_WEBSITE_TAGS"
            :key="tag"
            class="pill"
            :class="{ active: selectedTags[0] === tag }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
        </div>

        <!-- 其他标签（次级） -->
        <div v-if="customTags.length" class="extra-row">
          <button
            v-for="tag in customTags"
            :key="tag"
            class="mini-chip"
            :class="{ active: selectedTags[0] === tag }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>

      <!-- 列表标题 -->
      <div class="section-head">
        <span class="section-title">{{ filterActive ? '筛选结果' : '全部站点' }}</span>
        <span class="count-badge">共 {{ websites.length }} 个</span>
      </div>

      <!-- 站点网格 -->
      <div class="websites-grid" v-loading="loading">
        <a
          v-for="website in websites"
          :key="website.id"
          class="site-card"
          :href="website.url"
          target="_blank"
          rel="noopener noreferrer"
        >
          <div class="card-top">
            <div class="site-avatar">
              <span class="avatar-letter">{{ (website.name || '?').slice(0, 1).toUpperCase() }}</span>
              <img
                v-if="website.avatar"
                :src="website.avatar"
                :alt="website.name"
                @error="handleImageError"
              />
            </div>
            <div class="site-head-info">
              <div class="site-name" :title="website.name">{{ website.name }}</div>
              <div class="site-host">
                {{ hostOf(website.url) }}
                <el-icon :size="11" class="host-icon"><TopRight /></el-icon>
              </div>
            </div>
          </div>

          <p class="site-desc" :title="website.description">{{ website.description }}</p>

          <div class="site-tags">
            <el-tag
              v-for="tag in orderTagsPresetFirst(splitTags(website.tags)).slice(0, 4)"
              :key="tag"
              size="small"
              round
              :type="isPresetTag(tag) ? 'primary' : 'info'"
              :effect="isPresetTag(tag) ? 'light' : 'plain'"
            >
              {{ tag }}
            </el-tag>
            <span v-if="splitTags(website.tags).length > 4" class="more-count">
              +{{ splitTags(website.tags).length - 4 }}
            </span>
          </div>
        </a>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && websites.length === 0" class="empty-state">
        <el-empty :description="filterActive ? '没有符合筛选条件的站点' : '暂无收录站点'">
          <el-button v-if="filterActive" round @click="resetFilter">清空筛选</el-button>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup name="WebsitesPage">
import { ref, computed, watchEffect } from 'vue'
import { Search, TopRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import { PRESET_WEBSITE_TAGS, isPresetTag, splitTags, orderTagsPresetFirst } from '@/constants/websiteTags'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()
// 与 chat 页一致：直接访问本页时也应用用户主题（挂载 dark.scss 与暗色变量）
themeStore.initTheme()

const loading = ref(false)
const searchKeyword = ref('')
const selectedTags = ref([])
const websites = ref([])

// 自定义标签：站点实际使用、内定分类之外的补充标签
const customTags = computed(() => {
  const tags = new Set()
  websites.value.forEach((website) => {
    splitTags(website.tags).forEach((tag) => {
      if (!isPresetTag(tag)) {
        tags.add(tag)
      }
    })
  })
  return Array.from(tags)
})

// 是否处于筛选状态（标题与空态文案用）
const filterActive = computed(() => !!(searchKeyword.value || selectedTags.value.length))

// 展示用主机名（解析失败时原样返回）
const hostOf = (url) => {
  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

// 加载网站数据（搜索关键词与选中标签变化时经 watchEffect 自动重载）
const loadWebsites = async () => {
  loading.value = true
  try {
    const response = await defaultApi.apiWebsitesListGet({
      page: 1,
      limit: 100, // 一次取全量
      name: searchKeyword.value || undefined,
      tag: selectedTags.value[0] || undefined
    })

    if (response.code === 0) {
      websites.value = response.data.records || []
    } else {
      ElMessage.error(response.message || '加载站点列表失败')
    }
  } catch (error) {
    console.error('加载站点列表失败:', error)
    ElMessage.error('加载站点列表失败')
  } finally {
    loading.value = false
  }
}

watchEffect(() => {
  loadWebsites()
})

// 分类切换（列表接口仅支持单标签筛选：单选语义，再点一次取消）
const toggleTag = (tag) => {
  selectedTags.value = selectedTags.value[0] === tag ? [] : [tag]
}

// 清空全部筛选
const resetFilter = () => {
  searchKeyword.value = ''
  selectedTags.value = []
}

// 头像加载失败时露出字母兜底
const handleImageError = (event) => {
  event.target.style.display = 'none'
}
</script>

<style scoped lang="scss">
.websites-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background: #f5f7fa;
}

// ===== 环境光斑 =====
.ambient {
  position: absolute;
  inset: 0;
  pointer-events: none;

  .blob {
    position: absolute;
    border-radius: 50%;
    filter: blur(90px);
  }

  .blob-a {
    width: 520px;
    height: 520px;
    top: -180px;
    right: -80px;
    background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  }

  .blob-b {
    width: 420px;
    height: 420px;
    top: 320px;
    left: -160px;
    background: color-mix(in srgb, var(--el-color-primary) 7%, transparent);
  }
}

.page-body {
  position: relative;
  z-index: 1;
  max-width: 1080px;
  margin: 0 auto;
  padding: 44px 24px 60px;
}

// ===== 头部 =====
.hero {
  text-align: center;

  .hero-title {
    margin: 0;
    font-size: 32px;
    font-weight: 800;
    letter-spacing: 1px;
    background: linear-gradient(120deg, var(--el-text-color-primary) 30%, var(--el-color-primary));
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  // 标题下的渐变短横线点缀
  .hero-accent {
    width: 36px;
    height: 4px;
    margin: 14px auto 0;
    border-radius: 999px;
    background: linear-gradient(90deg, var(--el-color-primary-light-3), var(--el-color-primary));
  }

  .hero-subtitle {
    margin: 12px 0 0;
    font-size: 14px;
    color: var(--el-text-color-secondary);
  }
}

// ===== 搜索与分类 =====
.toolbar {
  margin-top: 30px;
  padding: 18px 22px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 18px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);

  .search-row {
    display: flex;
    justify-content: center;

    .search-input {
      width: min(520px, 100%);

      :deep(.el-input__wrapper) {
        border-radius: 999px;
      }
    }
  }

  .cat-row {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 10px;
    margin-top: 16px;

    .pill {
      padding: 6px 18px;
      font-size: 13px;
      line-height: 1.6;
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
        box-shadow: 0 2px 10px var(--el-color-primary-light-7);
      }
    }
  }

  .extra-row {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 6px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px dashed var(--el-border-color-lighter);

    .mini-chip {
      padding: 2px 12px;
      font-size: 12px;
      line-height: 1.6;
      color: var(--el-text-color-secondary);
      background: transparent;
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 999px;
      cursor: pointer;
      transition: all 0.2s ease;

      &:hover {
        color: var(--el-color-primary);
        border-color: var(--el-color-primary-light-5);
      }

      &.active {
        color: var(--el-color-primary);
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
      }
    }
  }
}

// ===== 列表标题 =====
.section-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 28px 2px 16px;

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .count-badge {
    padding: 2px 10px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border-radius: 999px;
  }
}

// ===== 站点卡片 =====
.websites-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 18px;
}

.site-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px 20px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  text-decoration: none;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    transform: translateY(-4px);
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 12px 28px color-mix(in srgb, var(--el-color-primary) 12%, transparent);

    .host-icon {
      color: var(--el-color-primary);
    }
  }

  .card-top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .site-avatar {
    position: relative;
    width: 46px;
    height: 46px;
    border-radius: 12px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    background: linear-gradient(135deg, var(--el-color-primary-light-8), var(--el-color-primary-light-6));
    box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.04);

    .avatar-letter {
      font-size: 17px;
      font-weight: 700;
      color: var(--el-color-primary);
    }

    img {
      position: absolute;
      inset: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .site-head-info {
    flex: 1;
    min-width: 0;

    .site-name {
      font-size: 15px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .site-host {
      display: flex;
      align-items: center;
      gap: 3px;
      margin-top: 2px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;

      .host-icon {
        flex-shrink: 0;
        color: var(--el-text-color-placeholder);
        transition: color 0.2s ease;
      }
    }
  }

  .site-desc {
    margin: 0;
    font-size: 13px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .site-tags {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    margin-top: auto;
    min-height: 24px;

    .more-count {
      font-size: 12px;
      color: var(--el-text-color-placeholder);
    }
  }
}

// ===== 空状态 =====
.empty-state {
  padding: 60px 0;
}

// ===== 响应式 =====
@media screen and (max-width: 768px) {
  .page-body {
    padding: 32px 16px 44px;
  }

  .hero .hero-title {
    font-size: 24px;
  }

  .toolbar {
    margin-top: 22px;
    padding: 14px 16px;

    .cat-row {
      gap: 8px;

      .pill {
        padding: 5px 14px;
        font-size: 12px;
      }
    }
  }

  .websites-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}

// ===== 暗色模式：项目暗色变量未覆盖部分，这里手动补齐 =====
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.websites-page {
  html.dark & {
    background: #141414;
  }
}

.ambient .blob {
  html.dark & {
    filter: blur(110px);

    &.blob-a {
      background: color-mix(in srgb, var(--el-color-primary) 9%, transparent);
    }

    &.blob-b {
      background: color-mix(in srgb, var(--el-color-primary) 6%, transparent);
    }
  }
}

.toolbar {
  html.dark & {
    background: #1c1c1c;
    border-color: #363637;
    box-shadow: none;

    .cat-row .pill {
      background: #262627;

      &:hover {
        background: var(--el-color-primary-light-9);
      }

      &.active {
        color: #fff;
        background: var(--el-color-primary);
      }
    }

    .extra-row {
      border-top-color: #363637;

      .mini-chip {
        border-color: #363637;

        &:hover {
          border-color: var(--el-color-primary-light-5);
        }

        &.active {
          background: var(--el-color-primary-light-9);
        }
      }
    }
  }
}

.section-head .count-badge {
  html.dark & {
    background: #262627;
  }
}

.site-card {
  html.dark & {
    background: #1c1c1c;
    border-color: #363637;

    .site-avatar {
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06);
    }
  }
}
</style>
