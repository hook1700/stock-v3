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
          <div id="marketPieChart" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header><div class="card-header"><span>资金类型占比</span></div></template>
          <div id="fundTypeChart" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 板块资金汇总表 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>板块资金汇总</span>
          <el-button type="text" @click="exportData">导出数据</el-button>
        </div>
      </template>
      <el-table :data="sectorSummary" stripe>
        <el-table-column prop="name" label="板块名称" width="120" />
        <el-table-column prop="mainInflow" label="主力流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.mainInflow) }}</template>
        </el-table-column>
        <el-table-column prop="mainOutflow" label="主力流出" width="120">
          <template #default="{ row }">{{ formatAmount(row.mainOutflow) }}</template>
        </el-table-column>
        <el-table-column prop="retailInflow" label="散户流入" width="120">
          <template #default="{ row }">{{ formatAmount(row.retailInflow) }}</template>
        </el-table-column>
        <el-table-column prop="retailOutflow" label="散户流出" width="120">
          <template #default="{ row }">{{ formatAmount(row.retailOutflow) }}</template>
        </el-table-column>
        <el-table-column prop="netInflow" label="净流入" width="120">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.netInflow > 0, 'price-down': row.netInflow < 0 }">{{ formatAmount(row.netInflow) }}</span>
          </template>
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
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'

const summaryStats = ref([
  { icon: '💰', label: '总成交额', value: 12580, display: '1.26万亿' },
  { icon: '📈', label: '主力净流入', value: 524, display: '+524亿' },
  { icon: '📉', label: '散户净流出', value: -320, display: '-320亿' },
  { icon: '🏢', label: '机构流入', value: 680, display: '+680亿' },
  { icon: '🔻', label: '北向资金', value: -85, display: '-85亿' },
  { icon: '⚡', label: '活跃板块', value: 8, display: '8个' }
])

const sectorSummary = ref([
  { name: '新能源', mainInflow: 850000000, mainOutflow: 320000000, retailInflow: 210000000, retailOutflow: 415000000, netInflow: 1250000000, changeRate: 0.032 },
  { name: '半导体', mainInflow: 620000000, mainOutflow: 180000000, retailInflow: 150000000, retailOutflow: 300000000, netInflow: 890000000, changeRate: 0.028 },
  { name: '医药', mainInflow: 480000000, mainOutflow: 210000000, retailInflow: 120000000, retailOutflow: 280000000, netInflow: 670000000, changeRate: -0.012 },
  { name: '金融', mainInflow: 250000000, mainOutflow: 380000000, retailInflow: 80000000, retailOutflow: 100000000, netInflow: -150000000, changeRate: -0.008 },
  { name: '消费', mainInflow: 180000000, mainOutflow: 450000000, retailInflow: 60000000, retailOutflow: 110000000, netInflow: -320000000, changeRate: 0.015 }
])

let marketPieChart = null
let fundTypeChart = null

const initCharts = () => {
  marketPieChart = echarts.init(document.getElementById('marketPieChart'))
  fundTypeChart = echarts.init(document.getElementById('fundTypeChart'))

  marketPieChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie', radius: ['40%', '70%'],
      data: [
        { value: 524, name: '净流入', itemStyle: { color: '#f56c6c' } },
        { value: 320, name: '净流出', itemStyle: { color: '#67c23a' } },
        { value: 14736, name: '平衡', itemStyle: { color: '#909399' } }
      ]
    }]
  })

  fundTypeChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie', radius: '60%',
      data: [
        { value: 680, name: '机构', itemStyle: { color: '#409EFF' } },
        { value: 520, name: '主力', itemStyle: { color: '#e6a23c' } },
        { value: -320, name: '散户', itemStyle: { color: '#f56c6c' } },
        { value: -85, name: '北向', itemStyle: { color: '#67c23a' } }
      ]
    }]
  })
}

const formatAmount = (amount) => {
  if (Math.abs(amount) >= 100000000) return `${(amount / 100000000).toFixed(1)}亿`
  if (Math.abs(amount) >= 10000) return `${(amount / 10000).toFixed(1)}万`
  return amount.toLocaleString()
}

const exportData = () => ElMessage.info('导出功能开发中...')

onMounted(() => { nextTick(() => initCharts()) })
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
