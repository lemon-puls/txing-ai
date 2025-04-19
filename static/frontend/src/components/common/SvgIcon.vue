<template>
  <svg
    :class="[
      'svg-icon',
      className,
      { 'svg-icon--spin': spin },
      { 'svg-icon--hover': hover },
      { 'svg-icon--click': click }
    ]"
    :style="svgStyle"
    aria-hidden="true"
  >
    <use :href="iconName" />
  </svg>
</template>

<script>
import { defineComponent, computed } from 'vue'
import './style.scss'

export default defineComponent({
  name: 'SvgIcon',
  props: {
    // 图标ID
    id: {
      type: [String, Number],
      default: ''
    },
    // 自定义类名
    class: {
      type: String,
      default: ''
    },
    // 图标名称
    icon: {
      type: String,
      required: true
    },
    // 图标尺寸
    size: {
      type: [String, Number],
      default: '1em'
    },
    // 图标颜色
    color: {
      type: String,
      default: ''
    },
    // 旋转角度
    rotate: {
      type: [String, Number],
      default: 0
    },
    // 是否旋转动画
    spin: {
      type: Boolean,
      default: false
    },
    // 是否启用悬停动画
    hover: {
      type: Boolean,
      default: false
    },
    // 是否启用点击动画
    click: {
      type: Boolean,
      default: false
    }
  },
  setup(props) {
    // 计算类名
    const className = computed(() => props.class)

    // 计算图标名称
    const iconName = computed(() => `#icon-${props.icon}`)

    // 计算样式
    const svgStyle = computed(() => {
      const style = {
        width: typeof props.size === 'number' ? `${props.size}px` : props.size,
        height: typeof props.size === 'number' ? `${props.size}px` : props.size,
        color: props.color
      }

      if (props.rotate && !props.spin) {
        style.transform = `rotate(${props.rotate}deg)`
      }

      return style
    })

    return {
      className,
      iconName,
      svgStyle
    }
  }
})
</script>

<style lang="scss" scoped>
.svg-icon {
  display: inline-block;
  overflow: hidden;
  vertical-align: -0.15em;
  //fill: currentColor;
  transition: all 0.3s ease;
  width: 1em;
  height: 1em;

  &:focus {
    outline: none;
  }

  // 旋转动画
  &--spin {
    animation: svg-icon-spin 1s infinite linear;
  }

  // 悬停动画
  &--hover {
    &:hover {
      transform: scale(1.2) !important;
      filter: brightness(1.2) !important;
    }
  }

  // 点击动画
  &--click {
    cursor: pointer;
    &:active {
      transform: scale(0.9);
      opacity: 0.8;
    }
  }
}

@keyframes svg-icon-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
