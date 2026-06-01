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
          <el-date-picker v-model="filterForm.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 250px" />
        </el-form-item>
        <el-form-item label="评分筛选">
          <el-select v-model="filterForm.minScore" placeholder="最低评分" clearable style="width: 120px">
            <el-option label="≥ 0.8" :value="0.8" />
            <el-option label="≥ 0.7" :value="0.7" />
            <el-option label="≥ 0.6" :value="0.6" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchHistory">
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
        <el-table-column prop="runTime" label="执行时间" width="170" sortable />
        <el-table-column prop="strategyType" label="策略类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.strategyType)">{{ getTypeName(row.strategyType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="strategyName" label="策略名称" width="150" />
        <el-table-column prop="signalsCount" label="信号数" width="80" />
        <el-table-column prop="avgScore" label="平均评分" width="90">
          <template #default="{ row }">
            <el-rate v-model="row.avgScore" disabled show-score text-color="#ff9900" score-template="{value}" />
          </template>
        </el-table-column>
        <el-table-column prop="successRate" label="成功率" width="90">
          <template #default="{ row }">{{ row.successRate }}%</template>
        </el-table-column>
        <el-table-column prop="avgReturn" label="平均收益" width="90">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.avgReturn > 0, 'price-down': row.avgReturn < 0 }">{{ row.avgReturn }}%</span>
          </template>
        </el-table-column>
        <el-table-column prop="bestStock" label="最佳标的" min-width="200">
          <template #default="{ row }">{{ row.bestStock }} <span class="best-return">({{ row.bestReturn }}%)</span></template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="text" size="small" @click="viewDetail(row)">详情</el-button>
            <el-button type="text" size="small" @click="rerunStrategy(row)">重跑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50]" :total="totalCount" layout="total, sizes, prev, pager, next, jumper" />
      </div>
    </el-card>

    <!-- 收益率曲线 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header">
          <span>策略收益率曲线</span>
          <el-radio-group v-model="chartPeriod" size="small">
            <el-radio-button label="month">近1月</el-radio-button>
            <el-radio-button label="quarter">近3月</el-radio-button>
            <el-radio-button label="year">近1年</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div id="returnChart" style="height: 350px;"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const totalCount = ref(100)
const chartPeriod = ref('month')

const stats = ref({ totalRuns: 256, totalSignals: 3245, avgSuccessRate: 68.5, totalProfit: 125.3 })

const filterForm = reactive({ strategyType: '', dateRange: null, minScore: null })

const historyList = ref([
  { runTime: '2026-05-30 18:00:00', strategyType: 'short', strategyName: '均线回踩低吸', signalsCount: 23, avgScore: 4.2, successRate: 72, avgReturn: 8.5, bestStock: '600519 贵州茅台', bestReturn: 15.2 },
  { runTime: '2026-05-29 18:00:00', strategyType: 'short', strategyName: '突破缩量回踩', signalsCount: 18, avgScore: 3.8, successRate: 65, avgReturn: 6.8, bestStock: '000858 五粮液', bestReturn: 12.3 },
  { runTime: '2026-05-28 18:00:00', strategyType: 'medium', strategyName: '低估值修复', signalsCount: 12, avgScore: 4.0, successRate: 75, avgReturn: 12.5, bestStock: '601398 工商银行', bestReturn: 18.6 },
  { runTime: '2026-05-27 18:00:00', strategyType: 'long', strategyName: '高股息价值', signalsCount: 8, avgScore: 4.3, successRate: 85, avgReturn: 15.8, bestStock: '600036 招商银行', bestReturn: 22.1 },
  { runTime: '2026-05-26 18:00:00', strategyType: 'short', strategyName: '强势股10日线反抽', signalsCount: 15, avgScore: 3.9, successRate: 68, avgReturn: 7.2, bestStock: '000001 平安银行', bestReturn: 11.5 },
  { runTime: '2026-05-25 18:00:00', strategyType: 'medium', strategyName: '业绩持续增长', signalsCount: 14, avgScore: 4.1, successRate: 70, avgReturn: 14.2, bestStock: '002594 比亚迪', bestReturn: 25.8 },
  { runTime: '2026-05-24 18:00:00', strategyType: 'long', strategyName: '护城河龙头', signalsCount: 6, avgScore: 4.5, successRate: 82, avgReturn: 18.3, bestStock: '000333 美的集团', bestReturn: 28.5 },
  { runTime: '2026-05-23 18:00:00', strategyType: 'short', strategyName: '均线回踩低吸', signalsCount: 20, avgScore: 3.7, successRate: 62, avgReturn: 5.5, bestStock: '600887 伊利股份', bestReturn: 10.2 }
])

let returnChart = null

const initChart = () => {
  returnChart = echarts.init(document.getElementById('returnChart'))
  returnChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['短线策略', '中线策略', '长线策略'] },
    xAxis: { type: 'category', data: ['5/23', '5/24', '5/25', '5/26', '5/27', '5/28', '5/29', '5/30'] },
    yAxis: { type: 'value', name: '收益率(%)' },
    series: [
      { name: '短线策略', type: 'line', data: [5.5, 4.2, 3.8, 7.2, 3.5, 4.8, 6.8, 8.5], smooth: true },
      { name: '中线策略', type: 'line', data: [3.2, 2.8, 14.2, 3.5, 4.1, 12.5, 5.2, 6.8], smooth: true },
      { name: '长线策略', type: 'line', data: [2.1, 18.3, 2.5, 3.2, 15.8, 3.8, 4.5, 5.2], smooth: true }
    ]
  })
}

const searchHistory = () => { loading.value = true; setTimeout(() => { loading.value = false; ElMessage.success('搜索完成') }, 500) }
const resetFilter = () => { filterForm.strategyType = ''; filterForm.dateRange = null; filterForm.minScore = null; searchHistory() }
const exportHistory = () => ElMessage.info('导出功能开发中...')
const viewDetail = (row) => ElMessage.info(`查看 ${row.strategyName} 的详细记录`)
const rerunStrategy = (row) => ElMessage.info(`重新执行 ${row.strategyName}`)
const getTypeTag = (type) => ({ short: 'primary', medium: 'success', long: 'warning' })[type] || 'info'
const getTypeName = (type) => ({ short: '短线', medium: '中线', long: '长线' })[type] || type

onMounted(() => { nextTick(() => initChart()) })
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
