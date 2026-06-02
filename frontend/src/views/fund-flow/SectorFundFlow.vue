
<template>
  <div class="fund-flow-page">
    <div class="page-header">
      <h1>板块资金流向</h1>
      <p class="subtitle">实时监控各行业板块的资金流入与流出情况</p>
    </div>

    <!-- 统计概览 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value price-up">{{ inflowSectors }}</div>
            <div class="stat-label">资金流入板块</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value price-down">{{ outflowSectors }}</div>
            <div class="stat-label">资金流出板块</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ (totalInflow / 100000000).toFixed(1) }}亿</div>
            <div class="stat-label">总净流入</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ formatAmount(mainForceFlow) }}</div>
            <div class="stat-label">主力净流入</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 板块资金排行 -->
    <el-row :gutter="20">
      <!-- 净流入排行 -->
      <el-col :span="12">
        <el-card class="rank-card">
          <template #header>
            <div class="card-header">
              <span>板块净流入排行 TOP10</span>
              <el-button type="text" @click="refreshData" :loading="loading">刷新</el-button>
            </div>
          </template>
          <div class="sector-rank-list" v-loading="loading">
            <div v-for="(sector, index) in inflowRank" :key="sector.sector_name" class="rank-item">
              <div class="rank-number" :class="{ 'top3': index < 3 }">{{ index + 1 }}</div>
              <div class="rank-info">
                <div class="sector-name">{{ sector.sector_name }}</div>
                <el-progress :percentage="getFlowPercentage(sector.net_inflow)" :color="'#f56c6c'" :show-text="false" />
              </div>
              <div class="rank-amount price-up">+{{ formatAmount(sector.net_inflow) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 净流出排行 -->
      <el-col :span="12">
        <el-card class="rank-card">
          <template #header>
            <div class="card-header">
              <span>板块净流出排行 TOP10</span>
              <el-button type="text" @click="refreshData" :loading="loading">刷新</el-button>
            </div>
          </template>
          <div class="sector-rank-list" v-loading="loading">
            <div v-for="(sector, index) in outflowRank" :key="sector.sector_name" class="rank-item">
              <div class="rank-number" :class="{ 'top3': index < 3 }">{{ index + 1 }}</div>
              <div class="rank-info">
                <div class="sector-name">{{ sector.sector_name }}</div>
                <el-progress :percentage="getFlowPercentage(Math.abs(sector.net_inflow))" :color="'#67c23a'" :show-text="false" />
              </div>
              <div class="rank-amount price-down">{{ formatAmount(sector.net_inflow) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 板块资金分布图 -->
    <el-card class="chart-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>板块资金分布</span>
          <el-radio-group v-model="chartType" size="small">
            <el-radio-button label="bar">柱状图</el-radio-button>
            <el-radio-button label="pie">饼图</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div id="sectorChart" style="height: 400px;" v-loading="loading"></div>
    </el-card>

    <!-- 个股资金明细 -->
    <el-card class="detail-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>板块内个股资金流向 TOP5</span>
          <el-select v-model="selectedSector" placeholder="选择板块" style="width: 150px" size="small" @change="loadSectorStocks">
            <el-option v-for="s in allSectors" :key="s.sector_name" :label="s.sector_name" :value="s.sector_name" />
          </el-select>
        </div>
      </template>
      <el-table :data="topStocks" v-loading="loading" stripe size="small">
        <el-table-column prop="stock_code" label="代码" width="90" />
        <el-table-column prop="stock_name" label="名称" width="100" />
        <el-table-column prop="net_inflow" label="净流入" width="120">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.net_inflow > 0, 'price-down': row.net_inflow < 0 }">{{ formatAmount(row.net_inflow) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="main_inflow" label="主力流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.main_inflow) }}</template>
        </el-table-column>
        <el-table-column prop="retail_inflow" label="散户流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.retail_inflow) }}</template>
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
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getSectorFundFlow } from '../../api/index.js'

const inflowSectors = ref(0)
const outflowSectors = ref(0)
const totalInflow = ref(0)
const mainForceFlow = ref(0)
const chartType = ref('bar')
const selectedSector = ref('')
const loading = ref(false)

const allSectors = ref([])
const topStocks = ref([])

let sectorChart = null

const inflowRank = computed(() =>
  allSectors.value.filter(s => s.net_inflow > 0).sort((a, b) => b.net_inflow - a.net_inflow).slice(0, 10)
)

const outflowRank = computed(() =>
  allSectors.value.filter(s => s.net_inflow < 0).sort((a, b) => a.net_inflow - b.net_inflow).slice(0, 10)
)

const maxFlow = computed(() =>
  allSectors.value.length > 0 ? Math.max(...allSectors.value.map(s => Math.abs(s.net_inflow))) : 1
)

const getFlowPercentage = (amount) => {
  if (maxFlow.value === 0) return 0
  return Math.min((Math.abs(amount) / maxFlow.value) * 100, 100)
}

const loadSectorData = async () => {
  loading.value = true
  try {
    const response = await getSectorFundFlow()

    if (response && response.data) {
      const data = response.data || []
      allSectors.value = data

      // 计算统计数据
      inflowSectors.value = data.filter(s => s.net_inflow > 0).length
      outflowSectors.value = data.filter(s => s.net_inflow < 0).length
      totalInflow.value = data.reduce((sum, s) => sum + s.net_inflow, 0)
      mainForceFlow.value = data.reduce((sum, s) => sum + (s.main_inflow || 0), 0)

      // 默认选择第一个板块
      if (data.length > 0 && !selectedSector.value) {
        selectedSector.value = data[0].sector_name
        await loadSectorStocks()
      }

      // 更新图表
      nextTick(() => {
        updateChart()
      })

      ElMessage.success('板块资金流向数据加载成功')
    } else {
      // 使用模拟数据作为备选
      loadMockData()
    }
  } catch (error) {
    console.error('加载板块资金流向失败:', error)
    ElMessage.error('加载板块资金流向失败: ' + (error.message || '未知错误'))

    // 使用模拟数据作为备选
    loadMockData()
  } finally {
    loading.value = false
  }
}

const loadSectorStocks = async () => {
  if (!selectedSector.value) return

  loading.value = true
  try {
    const response = await fundFlowApi.getSectorStocks({ sector_name: selectedSector.value })

    if (response.data && response.data.success) {
      topStocks.value = response.data.data || []
    } else {
      throw new Error(response.data?.message || '加载失败')
    }
  } catch (error) {
    console.error('加载板块个股失败:', error)
    ElMessage.error('加载板块个股失败: ' + (error.message || '未知错误'))

    // 使用模拟数据作为备选
    topStocks.value = generateMockStocks()
  } finally {
    loading.value = false
  }
}

const loadMockData = () => {
  allSectors.value = [
    { sector_name: '新能源', net_inflow: 1250000000, main_inflow: 890000000 },
    { sector_name: '半导体', net_inflow: 890000000, main_inflow: 650000000 },
    { sector_name: '医药', net_inflow: 670000000, main_inflow: 480000000 },
    { sector_name: '科技', net_inflow: 520000000, main_inflow: 380000000 },
    { sector_name: '消费', net_inflow: -320000000, main_inflow: -220000000 },
    { sector_name: '金融', net_inflow: -150000000, main_inflow: -110000000 },
    { sector_name: '房地产', net_inflow: -280000000, main_inflow: -200000000 },
    { sector_name: '传媒', net_inflow: -120000000, main_inflow: -85000000 },
    { sector_name: '军工', net_inflow: 380000000, main_inflow: 280000000 },
    { sector_name: '农业', net_inflow: -95000000, main_inflow: -68000000 }
  ]

  inflowSectors.value = 5
  outflowSectors.value = 5
  totalInflow.value = allSectors.value.reduce((sum, s) => sum + s.net_inflow, 0)
  mainForceFlow.value = allSectors.value.reduce((sum, s) => sum + s.main_inflow, 0)

  if (allSectors.value.length > 0) {
    selectedSector.value = allSectors.value[0].sector_name
    topStocks.value = generateMockStocks()
  }

  ElMessage.warning('使用模拟数据展示')
}

const generateMockStocks = () => [
  { stock_code: '000001', stock_name: '平安银行', net_inflow: 125000000, main_inflow: 89000000, retail_inflow: 36000000, change_rate: 0.032 },
  { stock_code: '600519', stock_name: '贵州茅台', net_inflow: 98000000, main_inflow: 75000000, retail_inflow: 23000000, change_rate: 0.028 },
  { stock_code: '000858', stock_name: '五粮液', net_inflow: 76000000, main_inflow: 58000000, retail_inflow: 18000000, change_rate: 0.025 },
  { stock_code: '002594', stock_name: '比亚迪', net_inflow: 65000000, main_inflow: 48000000, retail_inflow: 17000000, change_rate: 0.041 },
  { stock_code: '601318', stock_name: '中国平安', net_inflow: 52000000, main_inflow: 39000000, retail_inflow: 13000000, change_rate: 0.018 }
]

const initChart = () => {
  const chartDom = document.getElementById('sectorChart')
  if (!chartDom) return

  sectorChart = echarts.init(chartDom)
  updateChart()
}

const updateChart = () => {
  if (!sectorChart) return

  const data = [...allSectors.value].sort((a, b) => b.net_inflow - a.net_inflow)

  if (chartType.value === 'bar') {
    sectorChart.setOption({
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params) => {
          const item = params[0]
          return `${item.name}<br/>净流入: ${formatAmount(item.value)}`
        }
      },
      xAxis: {
        type: 'category',
        data: data.map(s => s.sector_name),
        axisLabel: { rotate: 45 }
      },
      yAxis: {
        type: 'value',
        name: '净流入(亿元)',
        axisLabel: { formatter: v => (v / 100000000).toFixed(1) }
      },
      series: [{
        type: 'bar',
        data: data.map(s => ({
          value: s.net_inflow,
          itemStyle: { color: s.net_inflow > 0 ? '#f56c6c' : '#67c23a' }
        }))
      }]
    }, true)
  } else {
    sectorChart.setOption({
      tooltip: {
        trigger: 'item',
        formatter: (params) => {
          return `${params.name}<br/>净流入: ${formatAmount(params.value)}`
        }
      },
      series: [{
        type: 'pie',
        radius: ['30%', '60%'],
        data: data.map(s => ({
          name: s.sector_name,
          value: Math.abs(s.net_inflow),
          itemStyle: { color: s.net_inflow > 0 ? '#f56c6c' : '#67c23a' }
        }))
      }]
    }, true)
  }
}

watch(chartType, updateChart)

const refreshData = () => {
  loadSectorData()
}

const formatAmount = (amount) => {
  if (!amount || amount === 0) return '0'
  const absAmount = Math.abs(amount)
  if (absAmount >= 100000000) return `${amount > 0 ? '+' : ''}${(amount / 100000000).toFixed(1)}亿`
  if (absAmount >= 10000) return `${amount > 0 ? '+' : ''}${(amount / 10000).toFixed(1)}万`
  return `${amount > 0 ? '+' : ''}${amount.toLocaleString()}`
}

onMounted(() => {
  nextTick(() => {
    initChart()
    loadSectorData()
  })
})

// 监听窗口大小变化，重绘图表
window.addEventListener('resize', () => {
  if (sectorChart) {
    sectorChart.resize()
  }
})
</script>

<style scoped>
.fund-flow-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.stats-row { margin-bottom: 20px; }
.stat-card { text-align: center; }
.stat-value { font-size: 24px; font-weight: bold; color: #409EFF; }
.stat-label { font-size: 14px; color: #909399; margin-top: 5px; }
.rank-card { margin-bottom: 0; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.sector-rank-list { max-height: 500px; overflow-y: auto; }
.rank-item { display: flex; align-items: center; padding: 10px 0; border-bottom: 1px solid #f0f0f0; }
.rank-item:last-child { border-bottom: none; }
.rank-number { width: 28px; height: 28px; line-height: 28px; text-align: center; border-radius: 50%; background: #f0f0f0; color: #909399; font-size: 14px; margin-right: 12px; flex-shrink: 0; }
.rank-number.top3 { background: #409EFF; color: white; }
.rank-info { flex: 1; min-width: 0; }
.sector-name { font-size: 14px; margin-bottom: 6px; color: #303133; }
.rank-amount { width: 100px; text-align: right; font-size: 14px; font-weight: bold; flex-shrink: 0; }
.price-up { color: #f56c6c; }
.price-down { color: #67c23a; }
.chart-card, .detail-card { margin-bottom: 20px; }
</style>
