import {createRouter, createWebHashHistory} from 'vue-router'
import {ElMessage} from "element-plus";
import {hasRole} from "@/utils/auth.js";
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/BlankLayout.vue'),
      children: [
        {
          path: '', // 默认路由为欢迎页
          name: 'welcome',
          component: () => import('@/views/welcome-page.vue'),
          meta: {
            title: '欢迎',
            icon: 'home'
          }
        },
        {
          path: 'assistant',
          name: 'assistant',
          component: () => import('@/layouts/HeaderLayout.vue'),
          children: [
            {
              path: '',
              component: () => import('@/views/assistant/index.vue'),
              meta: {
                title: '首页',
                icon: 'home'
              }
            }
          ]
        },
        {
          path: 'chat',
          name: 'chat',
          component: () => import('@/views/chat/index.vue'),
          meta: {
            title: '智能对话',
            icon: 'chat'
          }
        },
        {
          path: 'test',
          name: 'test',
          component: () => import('@/views/test/index.vue'),
          meta: {
            title: '智能对话',
            icon: 'chat'
          }
        },
        {
          path: '/profile',
          name: 'Profile',
          component: () => import('@/views/profile/index.vue'),
          meta: {
            title: '个人中心'
          }
        },
        {
          path: 'websites',
          name: 'WebsitesPage',
          component: () => import('@/layouts/HeaderLayout.vue'),
          children: [
            {
              path: '',
              component: () => import('@/views/websites/index.vue'),
              meta: {
                title: '实用网站',
                icon: 'link'
              }
            }
          ]
        },
        {
          path: 'resume',
          name: 'resume',
          component: () => import('@/layouts/HeaderLayout.vue'),
          children: [
            {
              component: () => import('@/views/resume/index.vue'),
              path: '',
              meta: {
                title: 'AI简历优化',
                icon: 'document'
              }
            }
          ]

        },
        {
          path: 'travel',
          name: 'travel',
          component: () => import('@/layouts/HeaderLayout.vue'),
          children: [
            {
              component: () => import('@/views/travel/index.vue'),
              path: '',
              meta: {
                title: 'AI旅游攻略',
                icon: 'document'
              }
            }
          ]
        }
      ]
    },
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: {
        roles: ['admin']  // 需要 admin 角色
      },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/admin/dashboard/index.vue'),
          meta: {
            title: '控制台',
            icon: 'DataLine',
            roles: ['admin']
          }
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/admin/user/UserList.vue'),
          meta: {
            title: '用户管理',
            icon: 'user',
            roles: ['admin']
          }
        },
        {
          path: 'models',
          name: 'models',
          component: () => import('@/views/admin/model/ModelList.vue'),
          meta: {
            title: '模型管理',
            icon: 'cpu',
            roles: ['admin']
          }
        },
        {
          path: 'channels',
          name: 'channels',
          component: () => import('@/views/admin/channel/ChannelList.vue'),
          meta: {
            title: '渠道管理',
            icon: 'Connection',
            roles: ['admin']
          }
        },
        {
          path: 'preset',
          name: 'PresetList',
          component: () => import('@/views/admin/preset/PresetList.vue'),
          meta: {
            title: 'AI 助手管理',
            icon: 'Robot',
            roles: ['admin']
          }
        },
        {
          path: 'websites',
          name: 'AdminWebsites',
          component: () => import('@/views/admin/websites/WebsiteList.vue'),
          meta: {
            title: '网站管理',
            icon: 'Link',
            roles: ['admin']
          }
        }
      ]
    },
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('@/views/error/403.vue'),
      meta: {
        title: '访问被拒绝'
      }
    }
  ],
})



// 全局路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const isLoggedIn = userStore.isLoggedIn

  // 检查路由是否需要角色权限
  // Check if route requires role permission
  const requiredRoles = to.matched.map(record => record.meta.roles).filter(Boolean)[0]

  if (requiredRoles) {
    // 如果需要角色权限但用户未登录，重定向到登录页
    // If role permission is required but user is not logged in, redirect to login page
    if (!isLoggedIn) {
      ElMessage({
        message: '请先登录',
        type: 'warning',
        duration: 2000
      })
      // Go back to home page
      next(`/`)
      return
    }

    // 检查用户是否具有所需角色
    // Check if user has required roles
    if (!hasRole(userStore.userRoles, requiredRoles)) {
      ElMessage({
        message: '您没有权限访问该页面',
        type: 'warning',
        duration: 2000
      })
      next('/403')
      return
    }
  }

  // 其他情况放行
  // Allow access in other cases
  next()
})

export default router
