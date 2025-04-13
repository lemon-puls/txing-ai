import {createRouter, createWebHashHistory} from 'vue-router'

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
          component: () => import('@/views/assistant/index.vue'),
          meta: {
            title: '首页',
            icon: 'home'
          }
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
        }
      ]
    },
    {
      path: '/admin',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        {
          path: '/users',
          name: 'users',
          component: () => import('@/views/user/UserList.vue'),
          meta: {
            title: '用户管理',
            icon: 'user'
          }
        },
      ]
    }
  ],
})

// 全局前置守卫
router.beforeEach(async (to, from, next) => {
  // 获取 token
  // const token = localStorage.getItem('token')
  // const refreshToken = localStorage.getItem('refreshToken')

  // 如果是公开路由，直接放行
  // if (to.meta.public) {
  //   console.log('public route, allow access')
  //   next()
  //   return
  // }

  // 如果没有 token，跳转到登录页
  // if (!token) {
  //   console.log('token is null, redirect to login')
  //   next({
  //     path: '/login',
  //     query: {
  //       redirect: to.fullPath // 记录原始路由
  //     }
  //   })
  //   return
  // }

  next()


  // 如果有 token，尝试访问一个需要认证的 API 来验证 token 是否有效
  // try {
  //   // 这里可以调用一个轻量级的 API 来验证 token
  //   await defaultApi.apiUserRefreshTokenPost({
  //     refresh_token: refreshToken
  //   })
  //   next()
  // } catch (error) {
  //   console.log('token is invalid, redirect to login');
  //   // token 无效，清除本地存储并跳转到登录页
  //   localStorage.removeItem('token')
  //   localStorage.removeItem('refreshToken')
  //   localStorage.removeItem('user')
  //   next({
  //     path: '/login',
  //     query: {
  //       redirect: to.fullPath
  //     }
  //   })
  // }
})

export default router
