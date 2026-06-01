<template>
  <div class="fund-flow-trend-page">
    <div class="page-header">
      <h1>资金流趋势</h1>
      <p class="subtitle">分析资金流向的历史趋势和变化规律</p>
    </div>

    <!-- 时间范围选择 -->
    <el-card class="filter-card">
      <el-form inline>
        <el-form-item label="时间周期">
          <el-radio-group v-model="timePeriod">
            <el-radio-button label="day">日</el-radio-button>
            <el-radio-button label="week">周</el-radio-button>
            <el-radio-button label="month">月</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="板块筛选">
          <el-select v-model="selectedSectors" multiple placeholder="选择板块" style="width: 300px">
            <el-option v-for="s in sectorOptions" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="updateCharts">更新图表</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 主力趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>主力资金流向趋势</span></div>
      </template>
      <div id="mainForceTrendChart" style="height: 350px;"></div>
    </el-card>

    <!-- 散户趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>散户资金流向趋势</span></div>
      </template>
      <div id="retailTrendChart" style="height: 350px;"></div>
    </el-card>

    <!-- 净流入变化趋势 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>板块净流入变化趋势</span></div>
      </template>
      <div id="netInflowChart" style="height: 350px;"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'

const timePeriod = ref('day')
const selectedSectors = ref(['新能源', '半导体', '科技'])
const sectorOptions = ref(['新能源', '半导体', '医药', '科技', '消费', '金融', '房地产', '传媒', '军工', '农业'])

let mainForceChart = null
let retailChart = null
let netInflowChart = null

const initCharts = () => {
  mainForceChart = echarts.init(document.getElementById('mainForceTrendChart'))
  retailChart = echarts.init(document.getElementById('retailTrendChart'))
  netInflowChart = echarts.init(document.getElementById('netInflowChart'))
  updateCharts()
}

const getDates = () => {
  const dates = []
  const base = new Date('2026-05-23')
  for (let i = 0; i < 8; i++) {
    const d = new Date(base)
    d.setDate(d.getDate() + i)
    dates.push(`${d.getMonth() + 1}/${d.getDate()}`)
  }
  return dates
}

const updateCharts = () => {
  const dates = getDates()
  const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399']

  const generateSeries = (prefix) => selectedSectors.value.map((name, i) => ({
    name, type: 'line', smooth: true,
    data: Array.from({ length: 8 }, () => (Math.random() - 0.3) * 2000000000 + prefix),
    lineStyle: { color: colors[i % colors.length] }
  }))

  mainForceChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: selectedSectors.value },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value', name: '金额(亿元)', axisLabel: { formatter: v => (v / 100000000).toFixed(0) } },
    series: generateSeries(500000000)
  })

  retailChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: selectedSectors.value },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value', name: '金额(亿元)', axisLabel: { formatter: v => (v / 100000000).toFixed(0) } },
    series: generateSeries(-200000000)
  })

  netInflowChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: selectedSectors.value },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value', name: '净流入(亿元)', axisLabel: { formatter: v => (v / 100000000).toFixed(0) } },
    series: generateSeries(300000000)
  })
}

watch([timePeriod, selectedSectors], updateCharts)

onMounted(() => { nextTick(() => initCharts()) })
</script>

<style scoped>
.fund-flow-trend-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.filter-card { margin-bottom: 20px; }
.chart-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
