<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>股票策略分析仪表盘</h1>
      <p class="subtitle">实时监控股票市场动态和策略执行情况</p>
    </div>

    <!-- 关键指标卡片 -->
    <div class="metrics-grid">
      <el-card class="metric-card">
        <div class="metric-content">
          <div class="metric-icon">📊</div>
          <div class="metric-info">
            <div class="metric-value">{{ metrics.totalStocks }}</div>
            <div class="metric-label">监控股票数量</div>
          </div>
        </div>
      </el-card>

      <el-card class="metric-card">
        <div class="metric-content">
          <div class="metric-icon">⚡</div>
          <div class="metric-info">
            <div class="metric-value">{{ metrics.todaySignals }}</div>
            <div class="metric-label">今日信号数量</div>
          </div>
        </div>
      </el-card>

      <el-card class="metric-card">
        <div class="metric-content">
          <div class="metric-icon">📈</div>
          <div class="metric-info">
            <div class="metric-value">{{ metrics.successRate }}%</div>
            <div class="metric-label">策略成功率</div>
          </div>
        </div>
      </el-card>

      <el-card class="metric-card">
        <div class="metric-content">
          <div class="metric-icon">💰</div>
          <div class="metric-info">
            <div class="metric-value">{{ metrics.avgProfit }}%</div>
            <div class="metric-label">平均收益率</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 策略执行情况 -->
    <div class="strategy-section">
      <el-card class="strategy-card">
        <template #header>
          <div class="card-header">
            <span>策略执行情况</span>
            <el-button type="primary" size="small" @click="runAllStrategies">
              执行所有策略
            </el-button>
          </div>
        </template>

        <el-table :data="strategyResults" style="width: 100%">
          <el-table-column prop="strategyName" label="策略名称" width="200" />
          <el-table-column prop="signalsCount" label="信号数量" width="100" />
          <el-table-column prop="avgScore" label="平均评分" width="100">
            <template #default="{ row }">
              <el-rate
                v-model="row.avgScore"
                disabled
                show-score
                text-color="#ff9900"
                score-template="{value}"
              />
            </template>
          </el-table-column>
          <el-table-column prop="lastRunTime" label="最后执行时间" width="180" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="runStrategy(row.strategyId)">
                单独执行
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 板块资金流 -->
    <div class="fund-flow-section">
      <el-card class="fund-flow-card">
        <template #header>
          <div class="card-header">
            <span>板块资金流向</span>
            <el-button type="text" @click="refreshFundFlow">刷新</el-button>
          </div>
        </template>

        <div class="fund-flow-chart">
          <div v-for="sector in fundFlowData" :key="sector.name" class="sector-item">
            <div class="sector-name">{{ sector.name }}</div>
            <div class="flow-bar">
              <div
                class="flow-value"
                :class="{ 'positive': sector.netInflow > 0, 'negative': sector.netInflow < 0 }"
                :style="{ width: getFlowWidth(sector.netInflow) }"
              ></div>
            </div>
            <div class="flow-amount">{{ formatAmount(sector.netInflow) }}</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 最近信号 -->
    <div class="recent-signals">
      <el-card class="signals-card">
        <template #header>
          <div class="card-header">
            <span>最近信号</span>
            <el-button type="text" @click="viewAllSignals">查看全部</el-button>
          </div>
        </template>

        <div class="signals-list">
          <div v-for="signal in recentSignals" :key="signal.id" class="signal-item">
            <div class="signal-stock">
              <span class="stock-code">{{ signal.stockCode }}</span>
              <span class="stock-name">{{ signal.stockName }}</span>
            </div>
            <div class="signal-info">
              <span class="strategy-name">{{ signal.strategyName }}</span>
              <span class="signal-time">{{ signal.time }}</span>
            </div>
            <div class="signal-action">
              <el-tag :type="getSignalType(signal.action)">{{ signal.action }}</el-tag>
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 模拟数据
const metrics = ref({
  totalStocks: 4521,
  todaySignals: 127,
  successRate: 68.5,
  avgProfit: 12.3
})

const strategyResults = ref([
  {
    strategyId: 'short_term_1',
    strategyName: '均线回踩低吸',
    signalsCount: 23,
    avgScore: 4.2,
    lastRunTime: '2026-06-01 18:00:00'
  },
  {
    strategyId: 'short_term_2',
    strategyName: '突破缩量回踩',
    signalsCount: 18,
    avgScore: 3.8,
    lastRunTime: '2026-06-01 18:00:00'
  },
  {
    strategyId: 'short_term_3',
    strategyName: '强势股10日线反抽',
    signalsCount: 15,
    avgScore: 4.0,
    lastRunTime: '2026-06-01 18:00:00'
  }
])

const fundFlowData = ref([
  { name: '新能源', netInflow: 1250000000 },
  { name: '半导体', netInflow: 890000000 },
  { name: '医药', netInflow: 670000000 },
  { name: '消费', netInflow: -320000000 },
  { name: '金融', netInflow: -150000000 }
])

const recentSignals = ref([
  {
    id: 1,
    stockCode: '000001',
    stockName: '平安银行',
    strategyName: '均线回踩低吸',
    action: '买入',
    time: '18:05:23'
  },
  {
    id: 2,
    stockCode: '600036',
    stockName: '招商银行',
    strategyName: '高股息红利',
    action: '持有',
    time: '18:04:15'
  },
  {
    id: 3,
    stockCode: '000858',
    stockName: '五粮液',
    strategyName: '突破缩量回踩',
    action: '买入',
    time: '18:03:42'
  }
])

// 方法
const runAllStrategies = async () => {
  try {
    await ElMessageBox.confirm('确定要执行所有策略吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    ElMessage.info('开始执行所有策略...')
    // 调用API执行策略
    // await api.runAllStrategies()
    ElMessage.success('策略执行完成')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('策略执行失败')
    }
  }
}

const runStrategy = async (strategyId) => {
  try {
    ElMessage.info(`开始执行策略: ${strategyId}`)
    // 调用API执行单个策略
    // await api.runStrategy(strategyId)
    ElMessage.success('策略执行完成')
  } catch (error) {
    ElMessage.error('策略执行失败')
  }
}

const refreshFundFlow = () => {
  ElMessage.info('刷新资金流数据...')
  // 调用API刷新数据
}

const viewAllSignals = () => {
  // 跳转到信号页面
}

const getFlowWidth = (amount) => {
  const maxAmount = Math.max(...fundFlowData.value.map(s => Math.abs(s.netInflow)))
  const percentage = (Math.abs(amount) / maxAmount) * 100
  return `${Math.min(percentage, 100)}%`
}

const formatAmount = (amount) => {
  if (amount >= 100000000) {
    return `${(amount / 100000000).toFixed(1)}亿`
  } else if (amount >= 10000) {
    return `${(amount / 10000).toFixed(1)}万`
  }
  return amount.toLocaleString()
}

const getSignalType = (action) => {
  switch (action) {
    case '买入': return 'success'
    case '卖出': return 'danger'
    case '持有': return 'warning'
    default: return 'info'
  }
}

onMounted(() => {
  // 初始化数据
  // loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.dashboard-header {
  margin-bottom: 30px;
}

.dashboard-header h1 {
  margin: 0;
  color: #303133;
}

.subtitle {
  margin: 10px 0 0;
  color: #909399;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.metric-card {
  border-radius: 8px;
}

.metric-content {
  display: flex;
  align-items: center;
  padding: 20px;
}

.metric-icon {
  font-size: 40px;
  margin-right: 15px;
}

.metric-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.metric-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.strategy-section,
.fund-flow-section,
.recent-signals {
  margin-bottom: 30px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fund-flow-chart {
  padding: 10px 0;
}

.sector-item {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.sector-name {
  width: 100px;
  font-size: 14px;
}

.flow-bar {
  flex: 1;
  height: 20px;
  background-color: #f0f0f0;
  border-radius: 10px;
  margin: 0 15px;
  overflow: hidden;
}

.flow-value {
  height: 100%;
  transition: width 0.3s;
}

.flow-value.positive {
  background-color: #f56c6c;
}

.flow-value.negative {
  background-color: #67c23a;
}

.flow-amount {
  width: 100px;
  text-align: right;
  font-size: 14px;
}

.signals-list {
  max-height: 300px;
  overflow-y: auto;
}

.signal-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.signal-item:last-child {
  border-bottom: none;
}

.signal-stock {
  display: flex;
  flex-direction: column;
}

.stock-code {
  font-weight: bold;
  color: #303133;
}

.stock-name {
  font-size: 12px;
  color: #909399;
}

.signal-info {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.strategy-name {
  font-size: 14px;
}

.signal-time {
  font-size: 12px;
  color: #909399;
}

.signal-action {
  min-width: 60px;
}
</style>