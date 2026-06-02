
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
          <el-radio-group v-model="timePeriod" @change="handleTimePeriodChange">
            <el-radio-button label="day">日</el-radio-button>
            <el-radio-button label="week">周</el-radio-button>
            <el-radio-button label="month">月</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="板块筛选">
          <el-select
            v-model="selectedSectors"
            multiple
            placeholder="选择板块"
            style="width: 300px"
            @change="handleSectorChange"
          >
            <el-option v-for="s in sectorOptions" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadTrendData" :loading="loading">更新图表</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 主力趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>主力资金流向趋势</span></div>
      </template>
      <div id="mainForceTrendChart" style="height: 350px;" v-loading="loading"></div>
    </el-card>

    <!-- 散户趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>散户资金流向趋势</span></div>
      </template>
      <div id="retailTrendChart" style="height: 350px;" v-loading="loading"></div>
    </el-card>

    <!-- 净流入变化趋势 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header"><span>板块净流入变化趋势</span></div>
      </template>
      <div id="netInflowChart" style="height: 350px;" v-loading="loading"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { fundFlowApi } from '@/services/api'

const loading = ref(false)
const timePeriod = ref('day')
const selectedSectors = ref(['新能源', '半导体', '科技'])
const sectorOptions = ref(['新能源', '半导体', '医药', '科技', '消费', '金融', '房地产', '传媒', '军工', '农业'])

let mainForceChart = null
let retailChart = null
let netInflowChart = null

const loadTrendData = async () => {
  loading.value = true
  try {
    // 调用API获取资金流趋势数据
    const params = {
      period: timePeriod.value,
      sectors: selectedSectors.value.join(',')
    }

    const response = await fundFlowApi.getFundFlowTrend(params)

    if (response.data && response.data.success) {
      const data = response.data.data || {}

      // 更新图表数据
      nextTick(() => {
        updateChartsWithData(data)
      })

      ElMessage.success('资金流趋势数据加载成功')
    } else {
      throw new Error(response.data?.message || '加载失败')
    }
  } catch (error) {
    console.error('加载资金流趋势失败:', error)
    ElMessage.error('加载资金流趋势失败: ' + (error.message || '未知错误'))

    // 使用模拟数据作为备选
    loadMockTrendData()
  } finally {
    loading.value = false
  }
}

const updateChartsWithData = (data) => {
  if (!mainForceChart || !retailChart || !netInflowChart) return

  const dates = data.dates || generateMockDates()
  const sectorData = data.sector_data || {}

  const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399']

  // 主力资金流向趋势图
  const mainForceSeries = selectedSectors.value.map((sectorName, i) => {
    const sectorDataItem = sectorData[sectorName] || {}
    return {
      name: sectorName,
      type: 'line',
      smooth: true,
      data: sectorDataItem.main_force || generateMockData(500000000),
      lineStyle: { color: colors[i % colors.length] },
      itemStyle: { color: colors[i % colors.length] }
    }
  })

  mainForceChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: function(params) {
        let result = `${params[0].axisValue}<br/>`
        params.forEach(item => {
          const value = (item.value / 100000000).toFixed(2)
          result += `${item.marker} ${item.seriesName}: ${value}亿<br/>`
        })
        return result
      }
    },
    legend: { data: selectedSectors.value },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: dates },
    yAxis: {
      type: 'value',
      name: '金额(亿元)',
      axisLabel: { formatter: v => (v / 100000000).toFixed(0) }
    },
    series: mainForceSeries
  })

  // 散户资金流向趋势图
  const retailSeries = selectedSectors.value.map((sectorName, i) => {
    const sectorDataItem = sectorData[sectorName] || {}
    return {
      name: sectorName,
      type: 'line',
      smooth: true,
      data: sectorDataItem.retail || generateMockData(-200000000),
      lineStyle: { color: colors[i % colors.length] },
      itemStyle: { color: colors[i % colors.length] }
    }
  })

  retailChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: function(params) {
        let result = `${params[0].axisValue}<br/>`
        params.forEach(item => {
          const value = (item.value / 100000000).toFixed(2)
          result += `${item.marker} ${item.seriesName}: ${value}亿<br/>`
        })
        return result
      }
    },
    legend: { data: selectedSectors.value },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: dates },
    yAxis: {
      type: 'value',
      name: '金额(亿元)',
      axisLabel: { formatter: v => (v / 100000000).toFixed(0) }
    },
    series: retailSeries
  })

  // 板块净流入变化趋势图
  const netInflowSeries = selectedSectors.value.map((sectorName, i) => {
    const sectorDataItem = sectorData[sectorName] || {}
    return {
      name: sectorName,
      type: 'line',
      smooth: true,
      data: sectorDataItem.net_inflow || generateMockData(300000000),
      lineStyle: { color: colors[i % colors.length] },
      itemStyle: { color: colors[i % colors.length] }
    }
  })

  netInflowChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: function(params) {
        let result = `${params[0].axisValue}<br/>`
        params.forEach(item => {
          const value = (item.value / 100000000).toFixed(2)
          result += `${item.marker} ${item.seriesName}: ${value}亿<br/>`
        })
        return result
      }
    },
    legend: { data: selectedSectors.value },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: dates },
    yAxis: {
      type: 'value',
      name: '净流入(亿元)',
      axisLabel: { formatter: v => (v / 100000000).toFixed(0) }
    },
    series: netInflowSeries
  })
}

const initCharts = () => {
  const mainForceDom = document.getElementById('mainForceTrendChart')
  const retailDom = document.getElementById('retailTrendChart')
  const netInflowDom = document.getElementById('netInflowChart')

  if (!mainForceDom || !retailDom || !netInflowDom) return

  mainForceChart = echarts.init(mainForceDom)
  retailChart = echarts.init(retailDom)
  netInflowChart = echarts.init(netInflowDom)
}

const loadMockTrendData = () => {
  nextTick(() => {
    updateChartsWithData({
      dates: generateMockDates(),
      sector_data: {}
    })
  })

  ElMessage.warning('使用模拟数据展示')
}

const generateMockDates = () => {
  const dates = []
  const base = new Date('2026-05-23')
  for (let i = 0; i < 8; i++) {
    const d = new Date(base)
    d.setDate(d.getDate() + i)
    dates.push(`${d.getMonth() + 1}/${d.getDate()}`)
  }
  return dates
}

const generateMockData = (prefix) => {
  return Array.from({ length: 8 }, () => (Math.random() - 0.3) * 2000000000 + prefix)
}

const handleTimePeriodChange = () => {
  loadTrendData()
}

const handleSectorChange = () => {
  loadTrendData()
}

onMounted(() => {
  nextTick(() => {
    initCharts()
    loadTrendData()
  })
})

// 监听窗口大小变化，重绘图表
window.addEventListener('resize', () => {
  if (mainForceChart) {
    mainForceChart.resize()
  }
  if (retailChart) {
    retailChart.resize()
  }
  if (netInflowChart) {
    netInflowChart.resize()
  }
})
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
