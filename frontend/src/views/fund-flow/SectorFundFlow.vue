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
            <div class="stat-value">{{ totalInflow }}亿</div>
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
              <el-button type="text" @click="refreshData">刷新</el-button>
            </div>
          </template>
          <div class="sector-rank-list">
            <div v-for="(sector, index) in inflowRank" :key="sector.name" class="rank-item">
              <div class="rank-number" :class="{ 'top3': index < 3 }">{{ index + 1 }}</div>
              <div class="rank-info">
                <div class="sector-name">{{ sector.name }}</div>
                <el-progress :percentage="getFlowPercentage(sector.netInflow)" :color="'#f56c6c'" :show-text="false" />
              </div>
              <div class="rank-amount price-up">+{{ formatAmount(sector.netInflow) }}</div>
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
              <el-button type="text" @click="refreshData">刷新</el-button>
            </div>
          </template>
          <div class="sector-rank-list">
            <div v-for="(sector, index) in outflowRank" :key="sector.name" class="rank-item">
              <div class="rank-number" :class="{ 'top3': index < 3 }">{{ index + 1 }}</div>
              <div class="rank-info">
                <div class="sector-name">{{ sector.name }}</div>
                <el-progress :percentage="getFlowPercentage(Math.abs(sector.netInflow))" :color="'#67c23a'" :show-text="false" />
              </div>
              <div class="rank-amount price-down">{{ formatAmount(sector.netInflow) }}</div>
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
      <div id="sectorChart" style="height: 400px;"></div>
    </el-card>

    <!-- 个股资金明细 -->
    <el-card class="detail-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>板块内个股资金流向 TOP5</span>
          <el-select v-model="selectedSector" placeholder="选择板块" style="width: 150px" size="small">
            <el-option v-for="s in allSectors" :key="s.name" :label="s.name" :value="s.name" />
          </el-select>
        </div>
      </template>
      <el-table :data="topStocks" stripe size="small">
        <el-table-column prop="code" label="代码" width="90" />
        <el-table-column prop="name" label="名称" width="100" />
        <el-table-column prop="netInflow" label="净流入" width="120">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.netInflow > 0, 'price-down': row.netInflow < 0 }">{{ formatAmount(row.netInflow) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="mainInflow" label="主力流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.mainInflow) }}</template>
        </el-table-column>
        <el-table-column prop="retailInflow" label="散户流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.retailInflow) }}</template>
        </el-table-column>
        <el-table-column prop="changeRate" label="涨跌幅" width="90">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.changeRate > 0, 'price-down': row.changeRate < 0 }">{{ (row.changeRate * 100).toFixed(2) }}%</span>
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

const inflowSectors = ref(22)
const outflowSectors = ref(13)
const totalInflow = ref(892.5)
const mainForceFlow = ref(5243000000)
const chartType = ref('bar')
const selectedSector = ref('新能源')

const allSectors = ref([
  { name: '新能源', netInflow: 1250000000 },
  { name: '半导体', netInflow: 890000000 },
  { name: '医药', netInflow: 670000000 },
  { name: '科技', netInflow: 520000000 },
  { name: '消费', netInflow: -320000000 },
  { name: '金融', netInflow: -150000000 },
  { name: '房地产', netInflow: -280000000 },
  { name: '传媒', netInflow: -120000000 },
  { name: '军工', netInflow: 380000000 },
  { name: '农业', netInflow: -95000000 }
])

const inflowRank = computed(() => allSectors.value.filter(s => s.netInflow > 0).sort((a, b) => b.netInflow - a.netInflow).slice(0, 10))
const outflowRank = computed(() => allSectors.value.filter(s => s.netInflow < 0).sort((a, b) => a.netInflow - b.netInflow).slice(0, 10))

const maxFlow = computed(() => Math.max(...allSectors.value.map(s => Math.abs(s.netInflow))))
const getFlowPercentage = (amount) => Math.min((Math.abs(amount) / maxFlow.value) * 100, 100)

const topStocks = ref([
  { code: '000001', name: '平安银行', netInflow: 125000000, mainInflow: 89000000, retailInflow: 36000000, changeRate: 0.032 },
  { code: '600519', name: '贵州茅台', netInflow: 98000000, mainInflow: 75000000, retailInflow: 23000000, changeRate: 0.028 },
  { code: '000858', name: '五粮液', netInflow: 76000000, mainInflow: 58000000, retailInflow: 18000000, changeRate: 0.025 },
  { code: '002594', name: '比亚迪', netInflow: 65000000, mainInflow: 48000000, retailInflow: 17000000, changeRate: 0.041 },
  { code: '601318', name: '中国平安', netInflow: 52000000, mainInflow: 39000000, retailInflow: 13000000, changeRate: 0.018 }
])

let sectorChart = null

const initChart = () => {
  sectorChart = echarts.init(document.getElementById('sectorChart'))
  updateChart()
}

const updateChart = () => {
  const data = allSectors.value.sort((a, b) => b.netInflow - a.netInflow)
  if (chartType.value === 'bar') {
    sectorChart.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      xAxis: { type: 'category', data: data.map(s => s.name) },
      yAxis: { type: 'value', name: '净流入(亿元)', axisLabel: { formatter: v => (v / 100000000).toFixed(0) } },
      series: [{
        type: 'bar',
        data: data.map(s => ({
          value: s.netInflow,
          itemStyle: { color: s.netInflow > 0 ? '#f56c6c' : '#67c23a' }
        }))
      }]
    }, true)
  } else {
    sectorChart.setOption({
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: ['30%', '60%'],
        data: data.map(s => ({ name: s.name, value: Math.abs(s.netInflow) }))
      }]
    }, true)
  }
}

watch(chartType, updateChart)

const refreshData = () => { ElMessage.success('数据已刷新') }

const formatAmount = (amount) => {
  if (Math.abs(amount) >= 100000000) return `${(amount / 100000000).toFixed(1)}亿`
  if (Math.abs(amount) >= 10000) return `${(amount / 10000).toFixed(1)}万`
  return amount.toLocaleString()
}

onMounted(() => { nextTick(() => initChart()) })
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
