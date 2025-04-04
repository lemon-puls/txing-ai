import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

// 这块当前仅是示例代码，实际上还没有用到
export const useCounterStore = defineStore('counter', () => {
  const count = ref(0)
  const doubleCount = computed(() => count.value * 2)
  function increment() {
    count.value++
  }

  return { count, doubleCount, increment }
})
