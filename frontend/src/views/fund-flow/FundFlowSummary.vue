
<template>
  <div class="fund-flow-summary-page">
    <div class="page-header">
      <h1>资金流摘要</h1>
      <p class="subtitle">资金流向的整体概览与汇总分析</p>
    </div>

    <!-- 总体统计 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="4" v-for="stat in summaryStats" :key="stat.label">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon">{{ stat.icon }}</div>
            <div class="stat-value" :class="{ 'price-up': stat.value > 0, 'price-down': stat.value < 0 }">{{ stat.display }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 市场资金流向 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header><div class="card-header"><span>市场整体资金流向</span></div></template>
          <div id="marketPieChart" style="height: 300px;" v-loading="loading"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header><div class="card-header"><span>资金类型占比</span></div></template>
          <div id="fundTypeChart" style="height: 300px;" v-loading="loading"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 板块资金汇总表 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>板块资金汇总</span>
          <el-button type="text" @click="exportData" :loading="loading">导出数据</el-button>
        </div>
      </template>
      <el-table :data="sectorSummary" v-loading="loading" stripe>
        <el-table-column prop="sector_name" label="板块名称" width="120" />
        <el-table-column prop="main_inflow" label="主力流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.main_inflow) }}</template>
        </el-table-column>
        <el-table-column prop="main_outflow" label="主力流出" width="120">
          <template #default="{ row }">{{ formatAmount(row.main_outflow) }}</template>
        </el-table-column>
        <el-table-column prop="retail_inflow" label="散户流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.retail_inflow) }}</template>
        </el-table-column>
        <el-table-column prop="retail_outflow" label="散户流出" width="120">
          <template #default="{ row }">{{ formatAmount(row.retail_outflow) }}</template>
        </el-table-column>
        <el-table-column prop="net_inflow" label="净流入" width="120">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.net_inflow > 0, 'price-down': row.net_inflow < 0 }">{{ formatAmount(row.net_inflow) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="change_rate" label="涨跌幅" width="90">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.change_rate > 0, 'price-down': row.change_rate < 0 }">{{ (row.change_rate * 100).toFixed(2) }}%</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { fundFlowApi } from '@/services/api'

const loading = ref(false)

const summaryStats = ref([
  { icon: '💰', label: '总成交额', value: 0, display: '0' },
  { icon: '📈', label: '主力净流入', value: 0, display: '0' },
  { icon: '📉', label: '散户净流出', value: 0, display: '0' },
  { icon: '🏢', label: '机构流入', value: 0, display: '0' },
  { icon: '🔻', label: '北向资金', value: 0, display: '0' },
  { icon: '⚡', label: '活跃板块', value: 0, display: '0个' }
])

const sectorSummary = ref([])

let marketPieChart = null
let fundTypeChart = null

const loadSummaryData = async () => {
  loading.value = true
  try {
    const response = await fundFlowApi.getFundFlowSummary()

    if (response.data && response.data.success) {
      const data = response.data.data || {}

      // 更新统计数据显示
      updateSummaryStats(data)

      // 更新板块汇总数据
      sectorSummary.value = data.sector_summary || []

      // 更新图表
      nextTick(() => {
        updateCharts(data)
      })

      ElMessage.success('资金流摘要数据加载成功')
    } else {
      throw new Error(response.data?.message || '加载失败')
    }
  } catch (error) {
    console.error('加载资金流摘要失败:', error)
    ElMessage.error('加载资金流摘要失败: ' + (error.message || '未知错误'))

    // 使用模拟数据作为备选
    loadMockData()
  } finally {
    loading.value = false
  }
}

const updateSummaryStats = (data) => {
  const totalAmount = data.total_amount || 1258000000000 // 默认1.26万亿
  const mainNetInflow = data.main_net_inflow || 52400000000
  const retailNetOutflow = data.retail_net_outflow || -32000000000
  const institutionInflow = data.institution_inflow || 68000000000
  const northboundFlow = data.northbound_flow || -8500000000
  const activeSectors = data.active_sectors || 8

  summaryStats.value = [
    { icon: '💰', label: '总成交额', value: totalAmount, display: formatLargeAmount(totalAmount) },
    { icon: '📈', label: '主力净流入', value: mainNetInflow, display: formatAmount(mainNetInflow) },
    { icon: '📉', label: '散户净流出', value: retailNetOutflow, display: formatAmount(retailNetOutflow) },
    { icon: '🏢', label: '机构流入', value: institutionInflow, display: formatAmount(institutionInflow) },
    { icon: '🔻', label: '北向资金', value: northboundFlow, display: formatAmount(northboundFlow) },
    { icon: '⚡', label: '活跃板块', value: activeSectors, display: `${activeSectors}个` }
  ]
}

const updateCharts = (data) => {
  if (!marketPieChart || !fundTypeChart) return

  const netInflow = data.main_net_inflow || 524
  const netOutflow = Math.abs(data.retail_net_outflow || 320)
  const balanced = data.total_amount ? (data.total_amount / 100000000) - netInflow - netOutflow : 14736

  // 市场整体资金流向饼图
  marketPieChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}亿 ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      data: [
        { value: netInflow / 100000000, name: '净流入', itemStyle: { color: '#f56c6c' } },
        { value: netOutflow / 100000000, name: '净流出', itemStyle: { color: '#67c23a' } },
        { value: balanced / 100000000, name: '平衡', itemStyle: { color: '#909399' } }
      ]
    }]
  })

  // 资金类型占比饼图
  const institution = (data.institution_inflow || 680) / 100000000
  const main = (data.main_net_inflow || 520) / 100000000
  const retail = Math.abs(data.retail_net_outflow || -320) / 100000000
  const northbound = Math.abs(data.northbound_flow || -85) / 100000000

  fundTypeChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}亿 ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [{
      type: 'pie',
      radius: '60%',
      data: [
        { value: institution, name: '机构', itemStyle: { color: '#409EFF' } },
        { value: main, name: '主力', itemStyle: { color: '#e6a23c' } },
        { value: retail, name: '散户', itemStyle: { color: '#f56c6c' } },
        { value: northbound, name: '北向', itemStyle: { color: '#67c23a' } }
      ]
    }]
  })
}

const initCharts = () => {
  const marketChartDom = document.getElementById('marketPieChart')
  const fundTypeChartDom = document.getElementById('fundTypeChart')

  if (!marketChartDom || !fundTypeChartDom) return

  marketPieChart = echarts.init(marketChartDom)
  fundTypeChart = echarts.init(fundTypeChartDom)
}

const loadMockData = () => {
  summaryStats.value = [
    { icon: '💰', label: '总成交额', value: 1258000000000, display: '1.26万亿' },
    { icon: '📈', label: '主力净流入', value: 52400000000, display: '+524亿' },
    { icon: '📉', label: '散户净流出', value: -32000000000, display: '-320亿' },
    { icon: '🏢', label: '机构流入', value: 68000000000, display: '+680亿' },
    { icon: '🔻', label: '北向资金', value: -85000000000, display: '-85亿' },
    { icon: '⚡', label: '活跃板块', value: 8, display: '8个' }
  ]

  sectorSummary.value = [
    { sector_name: '新能源', main_inflow: 850000000, main_outflow: 320000000, retail_inflow: 210000000, retail_outflow: 415000000, net_inflow: 1250000000, change_rate: 0.032 },
    { sector_name: '半导体', main_inflow: 620000000, main_outflow: 180000000, retail_inflow: 150000000, retail_outflow: 300000000, net_inflow: 890000000, change_rate: 0.028 },
    { sector_name: '医药', main_inflow: 480000000, main_outflow: 210000000, retail_inflow: 120000000, retail_outflow: 280000000, net_inflow: 670000000, change_rate: -0.012 },
    { sector_name: '金融', main_inflow: 250000000, main_outflow: 380000000, retail_inflow: 80000000, retail_outflow: 100000000, net_inflow: -150000000, change_rate: -0.008 },
    { sector_name: '消费', main_inflow: 180000000, main_outflow: 450000000, retail_inflow: 60000000, retail_outflow: 110000000, net_inflow: -320000000, change_rate: 0.015 }
  ]

  nextTick(() => {
    updateCharts({
      total_amount: 1258000000000,
      main_net_inflow: 52400000000,
      retail_net_outflow: -32000000000,
      institution_inflow: 68000000000,
      northbound_flow: -85000000000,
      active_sectors: 8
    })
  })

  ElMessage.warning('使用模拟数据展示')
}

const formatAmount = (amount) => {
  if (!amount || amount === 0) return '0'
  const absAmount = Math.abs(amount)
  if (absAmount >= 100000000) return `${amount > 0 ? '+' : ''}${(amount / 100000000).toFixed(1)}亿`
  if (absAmount >= 10000) return `${amount > 0 ? '+' : ''}${(amount / 10000).toFixed(1)}万`
  return `${amount > 0 ? '+' : ''}${amount.toLocaleString()}`
}

const formatLargeAmount = (amount) => {
  if (!amount || amount === 0) return '0'
  if (amount >= 1000000000000) return `${(amount / 1000000000000).toFixed(2)}万亿`
  if (amount >= 100000000) return `${(amount / 100000000).toFixed(2)}亿`
  return formatAmount(amount)
}

const exportData = () => {
  ElMessage.info('导出功能开发中...')
}

onMounted(() => {
  nextTick(() => {
    initCharts()
    loadSummaryData()
  })
})

// 监听窗口大小变化，重绘图表
window.addEventListener('resize', () => {
  if (marketPieChart) {
    marketPieChart.resize()
  }
  if (fundTypeChart) {
    fundTypeChart.resize()
  }
})
</script>

<style scoped>
.fund-flow-summary-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.stats-row { margin-bottom: 0; }
.stat-card { text-align: center; }
.stat-icon { font-size: 28px; margin-bottom: 8px; }
.stat-value { font-size: 20px; font-weight: bold; color: #409EFF; }
.stat-label { font-size: 13px; color: #909399; margin-top: 5px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.price-up { color: #f56c6c; }
.price-down { color: #67c23a; }
</style>
