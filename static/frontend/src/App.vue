<template>
  <router-view />
  <!-- 全局登录弹窗：登录失效时由 api 层（handle401Error）自动唤起，登录成功广播 tick 让页面自刷新 -->
  <!-- Global login dialog: auto-raised by the api layer on session expiry; login success broadcasts a tick so pages can refetch -->
  <AuthDialog v-model="authVisible" @login-success="notifyAuthResolved" />
</template>

<script setup>
import { computed } from 'vue'
import AuthDialog from '@/components/AuthDialog.vue'
import { authDialogVisible, notifyAuthResolved } from '@/utils/authPrompt'

// imported ref 经 computed 中转，保证 v-model 双向绑定可靠
// route the imported ref through a computed for a reliable v-model
const authVisible = computed({
  get: () => authDialogVisible.value,
  set: (v) => { authDialogVisible.value = v }
})
</script>

<style>
#app {
  width: 100%;
  /* auto: 内容不溢出时不强制渲染滚动条（scroll 会在加载期画出诡异的空轨道）
     auto: no forced scrollbars when content fits (scroll paints phantom tracks while loading) */
  overflow: auto;
}
</style>
