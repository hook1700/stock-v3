<template>
  <div class="strategy-history-page">
    <div class="page-header">
      <h1>策略历史</h1>
      <p class="subtitle">查看历史策略执行记录和回测结果</p>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :model="filterForm" label-width="80px" inline>
        <el-form-item label="策略类型">
          <el-select v-model="filterForm.strategyType" placeholder="选择策略类型" clearable style="width: 150px">
            <el-option label="短线策略" value="short" />
            <el-option label="中线策略" value="medium" />
            <el-option label="长线策略" value="long" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 250px"
          />
        </el-form-item>
        <el-form-item label="评分筛选">
          <el-select v-model="filterForm.minScore" placeholder="最低评分" clearable style="width: 120px">
            <el-option label="≥ 0.8" :value="0.8" />
            <el-option label="≥ 0.7" :value="0.7" />
            <el-option label="≥ 0.6" :value="0.6" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchHistory" :loading="loading">
            <el-icon><Search /></el-icon> 搜索
          </el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计概览 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalRuns }}</div>
            <div class="stat-label">总执行次数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalSignals }}</div>
            <div class="stat-label">总信号数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.avgSuccessRate }}%</div>
            <div class="stat-label">平均成功率</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalProfit }}%</div>
            <div class="stat-label">总收益率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 历史记录 -->
    <el-card class="history-card">
      <template #header>
        <div class="card-header">
          <span>策略执行历史</span>
          <el-button type="text" @click="exportHistory">导出历史</el-button>
        </div>
      </template>
      <el-table :data="historyList" v-loading="loading" stripe>
        <el-table-column prop="run_time" label="执行时间" width="170" sortable />
        <el-table-column prop="strategy_type" label="策略类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.strategy_type)">{{ getTypeName(row.strategy_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="strategy_name" label="策略名称" width="150" />
        <el-table-column prop="signals_count" label="信号数" width="80" />
        <el-table-column prop="avg_score" label="平均评分" width="90">
          <template #default="{ row }">
            <el-rate v-model="row.avg_score" disabled show-score text-color="#ff9900" score-template="{value}" />
          </template>
        </el-table-column>
        <el-table-column prop="success_rate" label="成功率" width="90">
          <template #default="{ row }">{{ row.success_rate }}%</template>
        </el-table-column>
        <el-table-column prop="avg_return" label="平均收益" width="90">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.avg_return > 0, 'price-down': row.avg_return < 0 }">{{ row.avg_return }}%</span>
          </template>
        </el-table-column>
        <el-table-column prop="best_stock" label="最佳标的" min-width="200">
          <template #default="{ row }">{{ row.best_stock }} <span class="best-return">({{ row.best_return }}%)</span></template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="text" size="small" @click="viewDetail(row)">详情</el-button>
            <el-button type="text" size="small" @click="rerunStrategy(row)">重跑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="totalCount"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 收益率曲线 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header">
          <span>策略收益率曲线</span>
          <el-radio-group v-model="chartPeriod" size="small" @change="handleChartPeriodChange">
            <el-radio-button label="month">近1月</el-radio-button>
            <el-radio-button label="quarter">近3月</el-radio-button>
            <el-radio-button label="year">近1年</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div id="returnChart" style="height: 350px;" v-loading="chartLoading"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getStrategyHistory as getStrategyHistoryApi } from '../../api/index.js'

const loading = ref(false)
const chartLoading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const totalCount = ref(0)
const chartPeriod = ref('month')

const stats = ref({
  totalRuns: 0,
  totalSignals: 0,
  avgSuccessRate: 0,
  totalProfit: 0
})

const filterForm = reactive({
  strategyType: '',
  dateRange: null,
  minScore: null
})

const historyList = ref([])

let returnChart = null

const loadHistoryData = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      strategy_type: filterForm.strategyType || undefined,
      min_score: filterForm.minScore || undefined
    }

    // 添加日期范围参数
    if (filterForm.dateRange && filterForm.dateRange.length === 2) {
      params.start_date = formatDate(filterForm.dateRange[0])
      params.end_date = formatDate(filterForm.dateRange[1])
    }

    const response = await getStrategyHistoryApi('all', params)

    if (response && response.data) {
      const data = response.data || {}

      // 更新历史记录列表
      historyList.value = (data.records || []).map(item => ({
        run_time: item.run_time,
        strategy_type: item.strategy_type,
        strategy_name: item.strategy_name,
        signals_count: item.signals_count,
        avg_score: item.avg_score,
        success_rate: item.success_rate,
        avg_return: item.avg_return,
        best_stock: item.best_stock,
        best_return: item.best_return
      }))

      // 更新分页信息
      totalCount.value = data.total || 0

      ElMessage.success('策略历史数据加载成功')
    } else {
      // 使用模拟数据作为备选
      loadMockHistoryData()
    }
  } catch (error) {
    console.error('加载策略历史失败:', error)
    ElMessage.error('加载策略历史失败: ' + (error.message || '未知错误'))

    // 使用模拟数据作为备选
    loadMockHistoryData()
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    // API 暂未实现，使用模拟数据
    stats.value = {
      totalRuns: 256,
      totalSignals: 3245,
      avgSuccessRate: 68.5,
      totalProfit: 125.3
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
    stats.value = {
      totalRuns: 256,
      totalSignals: 3245,
      avgSuccessRate: 68.5,
      totalProfit: 125.3
    }
  }
}

const loadReturnChart = async () => {
  chartLoading.value = true
  try {
    // API 暂未实现，使用模拟数据
    nextTick(() => {
      updateChart({
        dates: ['5/23', '5/24', '5/25', '5/26', '5/27', '5/28', '5/29', '5/30'],
        short_strategy: [5.5, 4.2, 3.8, 7.2, 3.5, 4.8, 6.8, 8.5],
        medium_strategy: [3.2, 2.8, 14.2, 3.5, 4.1, 12.5, 5.2, 6.8],
        long_strategy: [2.1, 18.3, 2.5, 3.2, 15.8, 3.8, 4.5, 5.2]
      })
    })
  } catch (error) {
    console.error('加载收益率曲线失败:', error)
    loadMockChartData()
  } finally {
    chartLoading.value = false
  }
}

const updateChart = (data) => {
  if (!returnChart) return

  const dates = data.dates || ['5/23', '5/24', '5/25', '5/26', '5/27', '5/28', '5/29', '5/30']
  const shortData = data.short_strategy || [5.5, 4.2, 3.8, 7.2, 3.5, 4.8, 6.8, 8.5]
  const mediumData = data.medium_strategy || [3.2, 2.8, 14.2, 3.5, 4.1, 12.5, 5.2, 6.8]
  const longData = data.long_strategy || [2.1, 18.3, 2.5, 3.2, 15.8, 3.8, 4.5, 5.2]

  returnChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: function(params) {
        let result = `${params[0].axisValue}<br/>`
        params.forEach(item => {
          result += `${item.marker} ${item.seriesName}: ${item.value}%<br/>`
        })
        return result
      }
    },
    legend: { data: ['短线策略', '中线策略', '长线策略'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value', name: '收益率(%)' },
    series: [
      { name: '短线策略', type: 'line', data: shortData, smooth: true },
      { name: '中线策略', type: 'line', data: mediumData, smooth: true },
      { name: '长线策略', type: 'line', data: longData, smooth: true }
    ]
  })
}

const initChart = () => {
  const chartDom = document.getElementById('returnChart')
  if (!chartDom) return

  returnChart = echarts.init(chartDom)
}

const loadMockHistoryData = () => {
  historyList.value = [
    { run_time: '2026-05-30 18:00:00', strategy_type: 'short', strategy_name: '均线回踩低吸', signals_count: 23, avg_score: 4.2, success_rate: 72, avg_return: 8.5, best_stock: '600519 贵州茅台', best_return: 15.2 },
    { run_time: '2026-05-29 18:00:00', strategy_type: 'short', strategy_name: '突破缩量回踩', signals_count: 18, avg_score: 3.8, success_rate: 65, avg_return: 6.8, best_stock: '000858 五粮液', best_return: 12.3 },
    { run_time: '2026-05-28 18:00:00', strategy_type: 'medium', strategy_name: '低估值修复', signals_count: 12, avg_score: 4.0, success_rate: 75, avg_return: 12.5, best_stock: '601398 工商银行', best_return: 18.6 },
    { run_time: '2026-05-27 18:00:00', strategy_type: 'long', strategy_name: '高股息价值', signals_count: 8, avg_score: 4.3, success_rate: 85, avg_return: 15.8, best_stock: '600036 招商银行', best_return: 22.1 },
    { run_time: '2026-05-26 18:00:00', strategy_type: 'short', strategy_name: '强势股10日线反抽', signals_count: 15, avg_score: 3.9, success_rate: 68, avg_return: 7.2, best_stock: '000001 平安银行', best_return: 11.5 },
    { run_time: '2026-05-25 18:00:00', strategy_type: 'medium', strategy_name: '业绩持续增长', signals_count: 14, avg_score: 4.1, success_rate: 70, avg_return: 14.2, best_stock: '002594 比亚迪', best_return: 25.8 },
    { run_time: '2026-05-24 18:00:00', strategy_type: 'long', strategy_name: '护城河龙头', signals_count: 6, avg_score: 4.5, success_rate: 82, avg_return: 18.3, best_stock: '000333 美的集团', best_return: 28.5 },
    { run_time: '2026-05-23 18:00:00', strategy_type: 'short', strategy_name: '均线回踩低吸', signals_count: 20, avg_score: 3.7, success_rate: 62, avg_return: 5.5, best_stock: '600887 伊利股份', best_return: 10.2 }
  ]

  totalCount.value = 100

  ElMessage.warning('使用模拟数据展示')
}

const loadMockChartData = () => {
  nextTick(() => {
    updateChart({
      dates: ['5/23', '5/24', '5/25', '5/26', '5/27', '5/28', '5/29', '5/30'],
      short_strategy: [5.5, 4.2, 3.8, 7.2, 3.5, 4.8, 6.8, 8.5],
      medium_strategy: [3.2, 2.8, 14.2, 3.5, 4.1, 12.5, 5.2, 6.8],
      long_strategy: [2.1, 18.3, 2.5, 3.2, 15.8, 3.8, 4.5, 5.2]
    })
  })
}

const formatDate = (date) => {
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const searchHistory = () => {
  currentPage.value = 1
  loadHistoryData()
}

const resetFilter = () => {
  filterForm.strategyType = ''
  filterForm.dateRange = null
  filterForm.minScore = null
  currentPage.value = 1
  searchHistory()
}

const exportHistory = () => ElMessage.info('导出功能开发中...')

const viewDetail = (row) => {
  ElMessage.info(`查看 ${row.strategy_name} 的详细记录`)
}

const rerunStrategy = (row) => {
  ElMessage.info(`重新执行 ${row.strategy_name}`)
}

const getTypeTag = (type) => ({ short: 'primary', medium: 'success', long: 'warning' })[type] || 'info'

const getTypeName = (type) => ({ short: '短线', medium: '中线', long: '长线' })[type] || type

const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
  loadHistoryData()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadHistoryData()
}

const handleChartPeriodChange = () => {
  loadReturnChart()
}

onMounted(() => {
  nextTick(() => {
    initChart()
    loadStats()
    loadHistoryData()
    loadReturnChart()
  })
})

// 监听窗口大小变化，重绘图表
window.addEventListener('resize', () => {
  if (returnChart) {
    returnChart.resize()
  }
})
</script>

<style scoped>
.strategy-history-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.filter-card { margin-bottom: 20px; }
.stats-row { margin-bottom: 20px; }
.stat-card { text-align: center; }
.stat-value { font-size: 24px; font-weight: bold; color: #409EFF; }
.stat-label { font-size: 14px; color: #909399; margin-top: 5px; }
.history-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 20px; text-align: center; }
.chart-card { margin-bottom: 20px; }
.price-up { color: #f56c6c; }
.price-down { color: #67c23a; }
.best-return { color: #409EFF; font-weight: bold; }
</style>