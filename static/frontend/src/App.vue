<template>
  <router-view />
</template>

<style>
#app {
  width: 100%;
  /* auto: 内容不溢出时不强制渲染滚动条（scroll 会在加载期画出诡异的空轨道）
     auto: no forced scrollbars when content fits (scroll paints phantom tracks while loading) */
  overflow: auto;
}
</style>
