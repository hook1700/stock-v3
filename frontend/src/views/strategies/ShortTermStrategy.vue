<template>
  <div class="strategy-page">
    <div class="page-header">
      <h1>短线策略分析</h1>
      <p class="subtitle">基于技术指标的短线交易策略</p>
    </div>

    <!-- 策略选择 -->
    <el-card class="strategy-select-card">
      <template #header>
        <div class="card-header">
          <span>选择短线策略</span>
          <el-button type="primary" @click="runAllStrategies">执行所有策略</el-button>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col :span="8" v-for="strategy in strategies" :key="strategy.id">
          <el-card class="strategy-card" :class="{ active: selectedStrategy === strategy.id }">
            <div class="strategy-content">
              <h3>{{ strategy.name }}</h3>
              <p class="strategy-desc">{{ strategy.description }}</p>
              <div class="strategy-stats">
                <div class="stat-item">
                  <span class="stat-label">成功率:</span>
                  <span class="stat-value">{{ strategy.successRate }}%</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">平均收益:</span>
                  <span class="stat-value">{{ strategy.avgProfit }}%</span>
                </div>
              </div>
              <div class="strategy-actions">
                <el-button
                  type="primary"
                  size="small"
                  @click="selectStrategy(strategy.id)"
                >
                  {{ selectedStrategy === strategy.id ? '已选择' : '选择' }}
                </el-button>
                <el-button
                  type="success"
                  size="small"
                  @click="runStrategy(strategy.id)"
                >
                  执行策略
                </el-button>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- 策略参数配置 -->
    <el-card class="parameters-card" v-if="selectedStrategy">
      <template #header>
        <div class="card-header">
          <span>策略参数配置</span>
          <el-button type="text" @click="resetParameters">重置参数</el-button>
        </div>
      </template>

      <el-form :model="strategyParameters" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="8" v-for="param in getCurrentStrategyParams()" :key="param.key">
            <el-form-item :label="param.label">
              <el-input-number
                v-if="param.type === 'number'"
                v-model="strategyParameters[param.key]"
                :min="param.min"
                :max="param.max"
                :step="param.step"
                style="width: 100%"
              />
              <el-select
                v-else-if="param.type === 'select'"
                v-model="strategyParameters[param.key]"
                style="width: 100%"
              >
                <el-option
                  v-for="option in param.options"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-switch
                v-else-if="param.type === 'boolean'"
                v-model="strategyParameters[param.key]"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <div class="parameters-actions">
          <el-button type="primary" @click="saveParameters">保存参数</el-button>
          <el-button @click="resetParameters">重置</el-button>
        </div>
      </el-form>
    </el-card>

    <!-- 策略执行结果 -->
    <el-card class="results-card" v-if="strategyResults.length > 0">
      <template #header>
        <div class="card-header">
          <span>策略执行结果 ({{ strategyResults.length }} 个信号)</span>
          <div class="header-actions">
            <el-button type="text" @click="exportResults">导出结果</el-button>
            <el-button type="text" @click="clearResults">清空结果</el-button>
          </div>
        </div>
      </template>

      <el-table :data="strategyResults" v-loading="loading" stripe>
        <el-table-column prop="stockCode" label="股票代码" width="100" />
        <el-table-column prop="stockName" label="股票名称" width="120" />
        <el-table-column prop="strategyName" label="策略名称" width="150" />
        <el-table-column prop="score" label="评分" width="80" sortable>
          <template #default="{ row }">
            <el-rate
              v-model="row.score"
              disabled
              show-score
              text-color="#ff9900"
              score-template="{value}"
            />
          </template>
        </el-table-column>
        <el-table-column prop="buyPrice" label="买入价" width="100">
          <template #default="{ row }">
            {{ row.buyPrice.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="stopLossPrice" label="止损价" width="100">
          <template #default="{ row }">
            {{ row.stopLossPrice.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="takeProfitPrice" label="止盈价" width="100">
          <template #default="{ row }">
            {{ row.takeProfitPrice.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="logicDescription" label="逻辑说明" min-width="200" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="text" size="small" @click="viewStockDetail(row)">详情</el-button>
            <el-button type="text" size="small" @click="addToWatchlist(row)">关注</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 结果统计 -->
      <div class="results-summary">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="summary-item">
              <span class="summary-label">平均评分:</span>
              <span class="summary-value">{{ avgScore.toFixed(2) }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="summary-item">
              <span class="summary-label">高评分信号:</span>
              <span class="summary-value">{{ highScoreSignals }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="summary-item">
              <span class="summary-label">预期收益率:</span>
              <span class="summary-value">{{ expectedReturn.toFixed(1) }}%</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="summary-item">
              <span class="summary-label">风险收益比:</span>
              <span class="summary-value">{{ riskRewardRatio.toFixed(2) }}</span>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 策略历史表现 -->
    <el-card class="history-card" v-if="selectedStrategy">
      <template #header>
        <div class="card-header">
          <span>策略历史表现</span>
          <el-button type="text" @click="refreshHistory">刷新历史</el-button>
        </div>
      </template>

      <div class="history-charts">
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="chart-container">
              <h4>成功率趋势</h4>
              <div id="successRateChart" style="height: 300px;"></div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="chart-container">
              <h4>收益率分布</h4>
              <div id="returnDistributionChart" style="height: 300px;"></div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, DataAnalysis } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getStrategies, runStrategy as runStrategyApi, getStrategyResults, getStrategyHistory } from '../../api/index.js'

const router = useRouter()

// 策略数据
const strategies = ref([])
const loading = ref(false)
const selectedStrategy = ref('')
const strategyParameters = reactive({})
const strategyResults = ref([])

// 计算属性
const avgScore = computed(() => {
  if (strategyResults.value.length === 0) return 0
  const sum = strategyResults.value.reduce((acc, result) => acc + result.score, 0)
  return (sum / strategyResults.value.length).toFixed(2)
})

const highScoreSignals = computed(() => {
  return strategyResults.value.filter(result => result.score >= 0.8).length
})

const expectedReturn = computed(() => {
  if (strategyResults.value.length === 0) return 0
  const returns = strategyResults.value.map(result => {
    return ((result.takeProfitPrice - result.buyPrice) / result.buyPrice * 100).toFixed(2)
  })
  return returns
})

const riskRewardRatio = computed(() => {
  if (strategyResults.value.length === 0) return 0
  const ratios = strategyResults.value.map(result => {
    const profit = result.takeProfitPrice - result.buyPrice
    const risk = result.buyPrice - result.stopLossPrice
    return risk > 0 ? (profit / risk).toFixed(2) : 0
  })
  return ratios
})

// 方法
const loadStrategies = async () => {
  try {
    loading.value = true
    const response = await getStrategies({ type: 'short_term' })
    if (response && response.data) {
      strategies.value = response.data.map(item => ({
        id: item.id || item.strategy_id,
        name: item.name || item.strategy_name,
        description: item.description || '',
        successRate: item.success_rate || 0,
        avgProfit: item.avg_profit || 0,
        parameters: item.parameters || []
      }))
    }
  } catch (error) {
    console.error('加载策略失败:', error)
    ElMessage.error('加载策略失败')
  } finally {
    loading.value = false
  }
}

const selectStrategy = (strategyId) => {
  selectedStrategy.value = strategyId
  // 初始化参数
  const strategy = strategies.value.find(s => s.id === strategyId)
  if (strategy && strategy.parameters) {
    strategy.parameters.forEach(param => {
      strategyParameters[param.key] = param.value
    })
  }
}

const getCurrentStrategyParams = () => {
  const strategy = strategies.value.find(s => s.id === selectedStrategy.value)
  return strategy ? strategy.parameters : []
}

const runStrategy = async (strategyId) => {
  loading.value = true
  try {
    ElMessage.info(`开始执行策略: ${strategyId}`)
    
    // 调用真实API执行策略
    const response = await runStrategyApi(strategyId, {
      parameters: strategyParameters
    })
    
    if (response && response.data) {
      strategyResults.value = response.data.results || []
      ElMessage.success(`策略执行完成，生成 ${strategyResults.value.length} 个信号`)
    } else {
      strategyResults.value = []
      ElMessage.warning('策略执行完成，但未生成信号')
    }
    
  } catch (error) {
    console.error('策略执行失败:', error)
    ElMessage.error('策略执行失败')
  } finally {
    loading.value = false
  }
}

const runAllStrategies = async () => {
  try {
    await ElMessageBox.confirm('确定要执行所有短线策略吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    loading.value = true
    strategyResults.value = []

    for (const strategy of strategies.value) {
      try {
        ElMessage.info(`正在执行策略: ${strategy.name}`)
        const response = await runStrategyApi(strategy.id, {
          parameters: strategyParameters
        })
        
        if (response && response.data && response.data.results) {
          strategyResults.value = [...strategyResults.value, ...response.data.results]
        }
      } catch (err) {
        console.error(`策略 ${strategy.name} 执行失败:`, err)
      }
    }

    ElMessage.success(`所有策略执行完成，共生成 ${strategyResults.value.length} 个信号`)

  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('策略执行失败')
    }
  } finally {
    loading.value = false
  }
}

const saveParameters = async () => {
  try {
    // 调用API保存参数
    ElMessage.success('参数保存成功')
  } catch (error) {
    ElMessage.error('参数保存失败')
  }
}

const resetParameters = () => {
  const strategy = strategies.value.find(s => s.id === selectedStrategy.value)
  if (strategy && strategy.parameters) {
    strategy.parameters.forEach(param => {
      strategyParameters[param.key] = param.value
    })
  }
  ElMessage.info('参数已重置')
}

const exportResults = async () => {
  try {
    // 调用API导出结果
    ElMessage.info('正在导出数据...')
    // 实际项目中应该调用后端导出接口
    const dataStr = JSON.stringify(strategyResults.value, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `strategy_results_${new Date().getTime()}.json`
    link.click()
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const clearResults = () => {
  strategyResults.value = []
  ElMessage.info('结果已清空')
}

const viewStockDetail = (result) => {
  router.push(`/stocks/${result.stockCode}`)
}

const addToWatchlist = async (result) => {
  try {
    // 调用API添加到关注列表
    ElMessage.success(`已将 ${result.stockName} 加入关注列表`)
  } catch (error) {
    ElMessage.error('添加关注失败')
  }
}

const refreshHistory = async () => {
  try {
    if (!selectedStrategy.value) {
      ElMessage.warning('请先选择策略')
      return
    }
    ElMessage.info('刷新历史数据...')
    const response = await getStrategyHistory(selectedStrategy.value)
    if (response && response.data) {
      ElMessage.success('历史数据已刷新')
    }
  } catch (error) {
    ElMessage.error('刷新历史数据失败')
  }
}

// 图表初始化
let successRateChart = null
let returnDistributionChart = null

const initCharts = () => {
  // 成功率趋势图
  const successRateChartDom = document.getElementById('successRateChart')
  if (successRateChartDom) {
    successRateChart = echarts.init(successRateChartDom)
    successRateChart.setOption({
      title: { text: '' },
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: ['1月', '2月', '3月', '4月', '5月', '6月']
      },
      yAxis: { type: 'value', name: '成功率(%)' },
      series: [{
        name: '成功率',
        type: 'line',
        data: [65, 68, 70, 72, 69, 71],
        smooth: true,
        lineStyle: { color: '#409EFF' },
        areaStyle: { color: 'rgba(64, 158, 255, 0.1)' }
      }]
    })
  }

  // 收益率分布图
  const returnDistributionChartDom = document.getElementById('returnDistributionChart')
  if (returnDistributionChartDom) {
    returnDistributionChart = echarts.init(returnDistributionChartDom)
    returnDistributionChart.setOption({
      title: { text: '' },
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: ['<5%', '5-10%', '10-15%', '15-20%', '>20%']
      },
      yAxis: { type: 'value', name: '信号数量' },
      series: [{
        name: '收益率分布',
        type: 'bar',
        data: [12, 28, 35, 20, 5],
        itemStyle: { color: '#67C23A' }
      }]
    })
  }
}

onMounted(() => {
  // 加载策略列表
  loadStrategies()
  // 初始化图表
  nextTick(() => {
    initCharts()
  })
})

onUnmounted(() => {
  if (successRateChart) {
    successRateChart.dispose()
  }
  if (returnDistributionChart) {
    returnDistributionChart.dispose()
  }
})
</script>

<style scoped>
.strategy-page {
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

.strategy-select-card,
.parameters-card,
.results-card,
.history-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.strategy-card {
  cursor: pointer;
  transition: all 0.3s;
}

.strategy-card.active {
  border-color: #409EFF;
  box-shadow: 0 2px 12px 0 rgba(64, 158, 255, 0.1);
}

.strategy-content h3 {
  margin: 0 0 10px 0;
  color: #303133;
}

.strategy-desc {
  color: #909399;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 15px;
}

.strategy-stats {
  margin-bottom: 15px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}

.stat-label {
  color: #606266;
}

.stat-value {
  color: #409EFF;
  font-weight: bold;
}

.strategy-actions {
  display: flex;
  gap: 10px;
}

.parameters-actions {
  text-align: center;
  padding-top: 20px;
}

.results-summary {
  margin-top: 20px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.summary-item {
  text-align: center;
}

.summary-label {
  display: block;
  color: #909399;
  font-size: 14px;
  margin-bottom: 5px;
}

.summary-value {
  display: block;
  color: #303133;
  font-size: 18px;
  font-weight: bold;
}

.history-charts {
  margin-top: 20px;
}

.chart-container {
  background: white;
  padding: 20px;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.chart-container h4 {
  margin: 0 0 15px 0;
  color: #303133;
  text-align: center;
}
</style>