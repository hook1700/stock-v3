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
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getIndustries } from '../../api/index.js'

const loading = ref(false)
const selectedIndustry = ref(null)

const industryStats = ref({
  totalIndustries: 0,
  avgChangeRate: 0,
  risingIndustries: 0,
  fallingIndustries: 0
})

const industryList = ref([])

let distributionChart = null
let marketCapChart = null

// 加载行业列表
const loadIndustries = async () => {
  try {
    loading.value = true
    const response = await getIndustries()
    
    if (response && response.data) {
      const industries = response.data || []
      
      // 转换为前端需要的格式
      industryList.value = industries.map(item => ({
        name: typeof item === 'string' ? item : (item.name || item.industry_name || ''),
        stockCount: item.stock_count || item.stockCount || 0,
        changeRate: item.change_rate || item.changeRate || 0,
        totalAmount: item.total_amount || item.totalAmount || 0
      }))
      
      industryStats.value.totalIndustries = industryList.value.length
      
      // 计算上涨和下跌行业数
      industryStats.value.risingIndustries = industryList.value.filter(i => i.changeRate > 0).length
      industryStats.value.fallingIndustries = industryList.value.filter(i => i.changeRate < 0).length
      
      // 计算平均涨跌幅
      if (industryList.value.length > 0) {
        const totalChange = industryList.value.reduce((sum, i) => sum + i.changeRate, 0)
        industryStats.value.avgChangeRate = (totalChange / industryList.value.length * 100).toFixed(2)
      }
      
      // 自动选择第一个行业
      if (industryList.value.length > 0) {
        selectIndustry(industryList.value[0])
      }
    }
  } catch (error) {
    console.error('加载行业列表失败:', error)
    ElMessage.error('加载行业列表失败')
  } finally {
    loading.value = false
  }
}

const selectIndustry = (row) => {
  selectedIndustry.value = {
    ...row,
    topStocks: [] // 后面可以通过API加载
  }
  nextTick(() => {
    initCharts()
  })
}

const initCharts = () => {
  if (!selectedIndustry.value) return

  // 涨跌幅分布图
  const distributionChartDom = document.getElementById('industryDistributionChart')
  if (distributionChartDom) {
    if (distributionChart) {
      distributionChart.dispose()
    }
    distributionChart = echarts.init(distributionChartDom)
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
  }

  // 市值分布图
  const marketCapChartDom = document.getElementById('industryMarketCapChart')
  if (marketCapChartDom) {
    if (marketCapChart) {
      marketCapChart.dispose()
    }
    marketCapChart = echarts.init(marketCapChartDom)
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
}

const refreshData = async () => {
  loading.value = true
  try {
    await loadIndustries()
    ElMessage.success('数据已刷新')
  } catch (error) {
    ElMessage.error('刷新数据失败')
  } finally {
    loading.value = false
  }
}

const formatAmount = (amount) => {
  if (!amount) return '0'
  if (amount >= 100000000) {
    return `${(amount / 100000000).toFixed(1)}亿`
  } else if (amount >= 10000) {
    return `${(amount / 10000).toFixed(1)}万`
  }
  return amount.toLocaleString()
}

onMounted(() => {
  loadIndustries()
})

onUnmounted(() => {
  if (distributionChart) {
    distributionChart.dispose()
  }
  if (marketCapChart) {
    marketCapChart.dispose()
  }
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
