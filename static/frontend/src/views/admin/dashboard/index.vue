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
            <div class="value">{{ item.value }}</div>
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
      data: [120, 132, 101, 134, 90, 230, 210]
    },
    {
      name: '用户数',
      type: 'line',
      smooth: true,
      data: [220, 182, 191, 234, 290, 330, 310]
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
      }
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
      radius: ['50%', '70%'], // 环形图
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
      ]
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
          // 颜色渐变
          const colorList = ['#83bff6', '#188df0', '#188df0', '#188df0', '#188df0', 
                           '#188df0', '#188df0', '#188df0', '#188df0', '#0b5ea8'];
          return colorList[params.dataIndex];
        },
        borderRadius: [0, 4, 4, 0]
      },
      label: {
        show: true,
        position: 'right'
      }
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

// 监听图表时间范围变化
watch(chartTimeRange, (newValue) => {
  // TODO: 根据时间范围更新图表数据
  console.log('Time range changed:', newValue)
})

onMounted(() => {
  // TODO: 从后端获取实际数据
})
</script>

<style lang="scss" scoped>
.dashboard-container {
  padding: 20px;

  .mb-4 {
    margin-bottom: 20px;
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
      }

      .icon {
        color: #409EFF;
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
      }

      .trend {
        display: flex;
        align-items: center;
        font-size: 14px;
        
        &.up {
          color: #67C23A;
        }
        
        &.down {
          color: #F56C6C;
        }
      }
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .chart-container {
    height: 350px;
    
    .chart, .pie-chart {
      height: 100%;
      width: 100%;
    }
  }
}
</style> 