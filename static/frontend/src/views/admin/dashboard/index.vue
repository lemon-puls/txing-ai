<template>
  <div class="dashboard-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="6" v-for="(item, index) in statistics" :key="index">
        <el-card shadow="hover" class="statistic-card">
          <div class="card-header">
            <span class="title">{{ item.title }}</span>
            <el-icon class="icon" :size="24">
              <component :is="item.icon" />
            </el-icon>
          </div>
          <div class="card-content">
            <div class="value">{{ item.displayValue }}</div>
            <div class="trend" :class="{ 'up': item.trend > 0, 'down': item.trend < 0 }">
              {{ Math.abs(item.trend) }}%
              <el-icon>
                <component :is="item.trend > 0 ? 'ArrowUp' : 'ArrowDown'" />
              </el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图表 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>对话趋势</span>
              <el-radio-group v-model="chartTimeRange" size="small">
                <el-radio-button label="week">本周</el-radio-button>
                <el-radio-button label="month">本月</el-radio-button>
                <el-radio-button label="year">全年</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="chart" :option="chartOption" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>模型使用占比</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="pie-chart" :option="pieChartOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 渠道和助手使用情况 -->
    <el-row :gutter="20" class="mb-4">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>渠道使用占比</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="pie-chart" :option="channelChartOption" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>AI助手使用排行</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="chart" :option="assistantChartOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row>
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
              <el-button text>查看全部</el-button>
            </div>
          </template>
          <el-table :data="activities" style="width: 100%" :max-height="400">
            <el-table-column prop="time" label="时间" width="180" />
            <el-table-column prop="user" label="用户" width="180" />
            <el-table-column prop="action" label="操作" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { use } from 'echarts/core'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

defineOptions({
  name: 'AdminDashboard'
})

// 注册 ECharts 组件
use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

// 顶部统计数据
const statistics = ref([
  {
    title: '总对话数',
    value: '12,345',
    trend: 15.4,
    icon: 'ChatLineRound'
  },
  {
    title: '活跃用户',
    value: '1,234',
    trend: 8.4,
    icon: 'User'
  },
  {
    title: '平均响应时间',
    value: '1.2s',
    trend: -12.5,
    icon: 'Timer'
  },
  {
    title: 'Token 消耗',
    value: '123.4k',
    trend: 5.6,
    icon: 'DataLine'
  }
])

// 图表时间范围
const chartTimeRange = ref('week')

// 趋势图表配置
const chartOption = ref({
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['对话数', '用户数']
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '对话数',
      type: 'line',
      smooth: true,
      data: [120, 132, 101, 134, 90, 230, 210],
      animation: true,
      animationDuration: 2000,
      animationEasing: 'quadraticInOut',
      animationDelay: function (idx) {
        return idx * 100;
      }
    },
    {
      name: '用户数',
      type: 'line',
      smooth: true,
      data: [220, 182, 191, 234, 290, 330, 310],
      animation: true,
      animationDuration: 2000,
      animationEasing: 'quadraticInOut',
      animationDelay: function (idx) {
        return idx * 100 + 1000;
      }
    }
  ]
})

// 饼图配置
const pieChartOption = ref({
  tooltip: {
    trigger: 'item'
  },
  legend: {
    orient: 'vertical',
    left: 'left'
  },
  series: [
    {
      type: 'pie',
      radius: '50%',
      data: [
        { value: 1048, name: 'GPT-4' },
        { value: 735, name: 'GPT-3.5' },
        { value: 580, name: 'Claude' },
        { value: 484, name: 'Gemini' }
      ],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      },
      animation: true,
      animationDuration: 2000,
      animationEasing: 'cubicInOut',
      animationDelay: function (idx) {
        return idx * 200;
      },
      animationDurationUpdate: 1000,
      animationEasingUpdate: 'cubicInOut'
    }
  ]
})

// 渠道使用占比图表配置
const channelChartOption = ref({
  tooltip: {
    trigger: 'item',
    formatter: '{a} <br/>{b}: {c} ({d}%)'
  },
  legend: {
    orient: 'vertical',
    left: 'left'
  },
  series: [
    {
      name: '渠道使用',
      type: 'pie',
      radius: ['50%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 16,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: [
        { value: 1048, name: 'Web端' },
        { value: 735, name: '移动端' },
        { value: 580, name: '微信小程序' },
        { value: 484, name: '企业微信' },
        { value: 300, name: '钉钉' }
      ],
      animation: true,
      animationDuration: 2500,
      animationEasing: 'cubicInOut',
      animationDelay: function (idx) {
        return idx * 200;
      },
      animationDurationUpdate: 1000,
      animationEasingUpdate: 'cubicInOut'
    }
  ]
})

// AI助手使用排行图表配置
const assistantChartOption = ref({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow'
    }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'value'
  },
  yAxis: {
    type: 'category',
    data: ['助手10', '助手9', '助手8', '助手7', '助手6', '助手5', '助手4', '助手3', '助手2', '助手1'],
    axisLabel: {
      interval: 0,
      rotate: 0
    }
  },
  series: [
    {
      name: '使用次数',
      type: 'bar',
      data: [320, 380, 450, 500, 580, 650, 750, 800, 900, 1200],
      itemStyle: {
        color: function(params) {
          const colorList = ['#83bff6', '#188df0', '#188df0', '#188df0', '#188df0', 
                           '#188df0', '#188df0', '#188df0', '#188df0', '#0b5ea8'];
          return colorList[params.dataIndex];
        },
        borderRadius: [0, 4, 4, 0]
      },
      label: {
        show: true,
        position: 'right'
      },
      animation: true,
      animationDuration: 3000,
      animationEasing: 'elasticOut',
      animationDelay: function (idx) {
        return idx * 100;
      },
      animationDurationUpdate: 1000,
      animationEasingUpdate: 'cubicInOut'
    }
  ]
})

// 最近活动数据
const activities = ref([
  {
    time: '2024-01-20 10:00:00',
    user: '张三',
    action: '创建了新的对话',
    status: 'success'
  },
  {
    time: '2024-01-20 09:30:00',
    user: '李四',
    action: '修改了系统设置',
    status: 'success'
  },
  {
    time: '2024-01-20 09:00:00',
    user: '王五',
    action: '导出了对话记录',
    status: 'failed'
  }
])

// 数字动画函数
const animateValue = (start, end, duration, callback) => {
  const startTime = performance.now()
  const update = () => {
    const currentTime = performance.now()
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)
    
    // 使用缓动函数使动画更平滑
    const easeOutQuart = 1 - Math.pow(1 - progress, 4)
    const current = Math.floor(start + (end - start) * easeOutQuart)
    
    callback(current)
    
    if (progress < 1) {
      requestAnimationFrame(update)
    }
  }
  update()
}

// 格式化数字
const formatNumber = (num) => {
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

// 动态更新统计数据
const updateStatistics = () => {
  statistics.value.forEach((stat, index) => {
    const endValue = parseInt(stat.value.replace(/[^0-9]/g, ''))
    let startValue = 0
    
    animateValue(startValue, endValue, 2000, (value) => {
      statistics.value[index] = {
        ...stat,
        displayValue: formatNumber(value)
      }
    })
  })
}

// 监听图表时间范围变化
watch(chartTimeRange, (newValue) => {
  // TODO: 根据时间范围更新图表数据
  console.log('Time range changed:', newValue)
})

onMounted(() => {
  // 添加进入动画延迟
  setTimeout(() => {
    updateStatistics()
  }, 300)

  // 创建图表实例的引用
  const charts = ref([])

  // 获取所有图表实例
  const chartEls = document.querySelectorAll('.chart, .pie-chart')
  chartEls.forEach(el => {
    const chart = echarts.init(el)
    charts.value.push(chart)
  })

  // 清除现有动画并重新设置选项以触发动画
  charts.value.forEach(chart => {
    chart.clear()
    const option = chart.getOption()
    if (option) {
      chart.setOption(option, true)
    }
  })
})
</script>

<style lang="scss" scoped>
.dashboard-container {
  padding: 20px;

  // 添加页面进入动画
  animation: fadeInUp 0.8s ease-out;

  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .mb-4 {
    margin-bottom: 20px;
  }

  // 为所有卡片添加统一的圆角和过渡效果
  :deep(.el-card) {
    border-radius: 16px;
    overflow: hidden;
    transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    animation: cardFadeIn 0.6s ease-out backwards;

    @for $i from 1 through 10 {
      &:nth-child(#{$i}) {
        animation-delay: #{$i * 0.1}s;
      }
    }

    &:hover {
      transform: translateY(-4px) scale(1.02);
      box-shadow: 0 20px 40px rgba(0, 0, 0, 0.12);

      .icon {
        transform: scale(1.2) rotate(10deg);
      }
    }

    .el-card__header {
      border-bottom: 1px solid rgba(0, 0, 0, 0.05);
      padding: 16px 20px;
      background: linear-gradient(135deg, rgba(255,255,255,0.5) 0%, rgba(255,255,255,0) 100%);
    }

    .el-card__body {
      padding: 20px;
      position: relative;
      overflow: hidden;

      &::after {
        content: '';
        position: absolute;
        top: -50%;
        left: -50%;
        width: 200%;
        height: 200%;
        background: radial-gradient(circle, rgba(255,255,255,0.2) 0%, rgba(255,255,255,0) 70%);
        opacity: 0;
        transition: opacity 0.3s;
      }

      &:hover::after {
        opacity: 1;
      }
    }
  }

  @keyframes cardFadeIn {
    from {
      opacity: 0;
      transform: translateY(40px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .statistic-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;

      .title {
        font-size: 16px;
        color: #606266;
        position: relative;
        
        &::after {
          content: '';
          position: absolute;
          bottom: -4px;
          left: 0;
          width: 0;
          height: 2px;
          background: #409EFF;
          transition: width 0.3s ease;
        }
      }

      .icon {
        color: #409EFF;
        background: rgba(64, 158, 255, 0.1);
        padding: 8px;
        border-radius: 12px;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
    }

    &:hover {
      .title::after {
        width: 100%;
      }
    }

    .card-content {
      display: flex;
      justify-content: space-between;
      align-items: flex-end;

      .value {
        font-size: 24px;
        font-weight: bold;
        color: #303133;
        transition: all 0.3s ease;
        
        &:hover {
          transform: scale(1.1);
          color: #409EFF;
        }
      }

      .trend {
        display: flex;
        align-items: center;
        font-size: 14px;
        padding: 4px 8px;
        border-radius: 8px;
        transition: all 0.3s ease;
        
        &.up {
          color: #67C23A;
          background: rgba(103, 194, 58, 0.1);
          
          &:hover {
            background: rgba(103, 194, 58, 0.2);
            transform: translateY(-2px);
          }
        }
        
        &.down {
          color: #F56C6C;
          background: rgba(245, 108, 108, 0.1);
          
          &:hover {
            background: rgba(245, 108, 108, 0.2);
            transform: translateY(-2px);
          }
        }
      }
    }
  }

  .chart-container {
    height: 350px;
    position: relative;
    
    .chart, .pie-chart {
      height: 100%;
      width: 100%;
      transition: all 0.3s ease;
      
      &:hover {
        transform: scale(1.02);
      }
    }
  }

  // 美化表格圆角和动画
  :deep(.el-table) {
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.3s ease;

    th {
      background-color: #f5f7fa;
      transition: background-color 0.3s ease;
    }

    tr {
      transition: all 0.3s ease;
      
      &:hover > td {
        background-color: rgba(64, 158, 255, 0.1) !important;
      }
    }

    .el-tag {
      border-radius: 12px;
      padding: 0 12px;
      transition: all 0.3s ease;
      
      &:hover {
        transform: scale(1.1);
      }
    }
  }

  // 添加按钮动画
  :deep(.el-button) {
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }
  }

  // 添加单选按钮组动画
  :deep(.el-radio-group) {
    .el-radio-button__inner {
      transition: all 0.3s ease;
      
      &:hover {
        transform: translateY(-2px);
      }
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  span {
    font-size: 16px;
    font-weight: 500;
    color: #303133;
  }

  .el-radio-group {
    margin-left: auto; // 确保按钮组靠右
  }
}
</style> 