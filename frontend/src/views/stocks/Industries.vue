<template>
  <div class="industries-page">
    <div class="page-header">
      <h1>行业分类</h1>
      <p class="subtitle">浏览各行业板块及其包含的股票信息</p>
    </div>

    <!-- 行业概览统计 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ industryStats.totalIndustries }}</div>
            <div class="stat-label">行业总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ industryStats.avgChangeRate }}%</div>
            <div class="stat-label">平均涨跌幅</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ industryStats.risingIndustries }}</div>
            <div class="stat-label">上涨行业数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ industryStats.fallingIndustries }}</div>
            <div class="stat-label">下跌行业数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 行业列表与详情 -->
    <el-card class="industries-card">
      <template #header>
        <div class="card-header">
          <span>行业板块列表</span>
          <el-button type="text" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- 行业列表 -->
        <el-col :span="8">
          <el-table
            :data="industryList"
            v-loading="loading"
            highlight-current-row
            @row-click="selectIndustry"
            stripe
            size="small"
          >
            <el-table-column prop="name" label="行业名称" min-width="120" />
            <el-table-column prop="stockCount" label="股票数" width="80" />
            <el-table-column prop="changeRate" label="涨跌幅" width="90">
              <template #default="{ row }">
                <span :class="{ 'price-up': row.changeRate > 0, 'price-down': row.changeRate < 0 }">
                  {{ (row.changeRate * 100).toFixed(2) }}%
                </span>
              </template>
            </el-table-column>
          </el-table>
        </el-col>

        <!-- 行业详情 -->
        <el-col :span="16">
          <div v-if="selectedIndustry" class="industry-detail">
            <div class="detail-header">
              <h3>{{ selectedIndustry.name }}</h3>
              <div class="detail-stats">
                <span>股票数: {{ selectedIndustry.stockCount }}</span>
                <span :class="{ 'price-up': selectedIndustry.changeRate > 0, 'price-down': selectedIndustry.changeRate < 0 }">
                  涨跌幅: {{ (selectedIndustry.changeRate * 100).toFixed(2) }}%
                </span>
                <span>成交额: {{ formatAmount(selectedIndustry.totalAmount) }}</span>
              </div>
            </div>

            <div class="detail-charts">
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="chart-container">
                    <div class="chart-title">涨跌幅分布</div>
                    <div id="industryDistributionChart" style="height: 250px;"></div>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="chart-container">
                    <div class="chart-title">市值分布</div>
                    <div id="industryMarketCapChart" style="height: 250px;"></div>
                  </div>
                </el-col>
              </el-row>
            </div>

            <h4 class="subsection-title">行业头部股票</h4>
            <el-table :data="selectedIndustry.topStocks" size="small" stripe>
              <el-table-column prop="code" label="代码" width="90" />
              <el-table-column prop="name" label="名称" width="100" />
              <el-table-column prop="closePrice" label="最新价" width="90">
                <template #default="{ row }">
                  {{ row.closePrice.toFixed(2) }}
                </template>
              </el-table-column>
              <el-table-column prop="changeRate" label="涨跌幅" width="90">
                <template #default="{ row }">
                  <span :class="{ 'price-up': row.changeRate > 0, 'price-down': row.changeRate < 0 }">
                    {{ (row.changeRate * 100).toFixed(2) }}%
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="marketCap" label="市值" width="120">
                <template #default="{ row }">
                  {{ formatAmount(row.marketCap) }}
                </template>
              </el-table-column>
            </el-table>
          </div>
          <el-empty v-else description="请选择左侧行业查看详情" />
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const loading = ref(false)
const selectedIndustry = ref(null)

const industryStats = ref({
  totalIndustries: 35,
  avgChangeRate: 1.23,
  risingIndustries: 22,
  fallingIndustries: 13
})

const industryList = ref([
  { name: '新能源', stockCount: 156, changeRate: 0.032, totalAmount: 12500000000 },
  { name: '半导体', stockCount: 132, changeRate: 0.028, totalAmount: 8900000000 },
  { name: '医药', stockCount: 198, changeRate: -0.012, totalAmount: 6700000000 },
  { name: '消费', stockCount: 145, changeRate: 0.015, totalAmount: 5400000000 },
  { name: '金融', stockCount: 87, changeRate: -0.008, totalAmount: 3200000000 },
  { name: '科技', stockCount: 112, changeRate: 0.041, totalAmount: 7800000000 },
  { name: '房地产', stockCount: 67, changeRate: -0.025, totalAmount: 2100000000 },
  { name: '传媒', stockCount: 54, changeRate: 0.019, totalAmount: 1800000000 }
])

const generateTopStocks = (industryName, count) => {
  return Array.from({ length: count }, (_, i) => ({
    code: `000${(i + 1).toString().padStart(3, '0')}`,
    name: `${industryName}股票${i + 1}`,
    closePrice: 10 + Math.random() * 90,
    changeRate: (Math.random() - 0.5) * 0.1,
    marketCap: Math.random() * 5000000000 + 500000000
  }))
}

let distributionChart = null
let marketCapChart = null

const selectIndustry = (row) => {
  selectedIndustry.value = {
    ...row,
    topStocks: generateTopStocks(row.name, 10)
  }
  nextTick(() => {
    initCharts()
  })
}

const initCharts = () => {
  if (!selectedIndustry.value) return

  // 涨跌幅分布图
  distributionChart = echarts.init(document.getElementById('industryDistributionChart'))
  distributionChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['>5%', '3-5%', '1-3%', '0-1%', '-1-0%', '-3至-1%', '-5至-3%', '<-5%'] },
    yAxis: { type: 'value', name: '股票数量' },
    series: [{
      type: 'bar',
      data: [5, 12, 28, 35, 20, 15, 8, 3],
      itemStyle: { color: '#409EFF' }
    }]
  })

  // 市值分布图
  marketCapChart = echarts.init(document.getElementById('industryMarketCapChart'))
  marketCapChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      data: [
        { value: 35, name: '大盘股 (>500亿)' },
        { value: 45, name: '中盘股 (100-500亿)' },
        { value: 76, name: '小盘股 (<100亿)' }
      ]
    }]
  })
}

const refreshData = () => {
  loading.value = true
  setTimeout(() => {
    loading.value = false
    ElMessage.success('数据已刷新')
  }, 500)
}

const formatAmount = (amount) => {
  if (amount >= 100000000) {
    return `${(amount / 100000000).toFixed(1)}亿`
  } else if (amount >= 10000) {
    return `${(amount / 10000).toFixed(1)}万`
  }
  return amount.toLocaleString()
}

onMounted(() => {
  selectIndustry(industryList.value[0])
})
</script>

<style scoped>
.industries-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #303133;
}

.subtitle {
  margin: 10px 0 0;
  color: #909399;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409EFF;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.industries-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.industry-detail {
  padding: 10px;
}

.detail-header {
  margin-bottom: 20px;
}

.detail-header h3 {
  margin: 0 0 10px 0;
  color: #303133;
}

.detail-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #606266;
}

.detail-charts {
  margin-bottom: 20px;
}

.chart-container {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
}

.chart-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 10px;
  text-align: center;
}

.subsection-title {
  margin: 20px 0 10px 0;
  color: #303133;
  font-size: 16px;
}

.price-up {
  color: #f56c6c;
}

.price-down {
  color: #67c23a;
}
</style>
