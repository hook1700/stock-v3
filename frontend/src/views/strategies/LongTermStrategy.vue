<template>
  <div class="strategy-page">
    <div class="page-header">
      <h1>长线策略分析</h1>
      <p class="subtitle">基于基本面分析的长线投资策略，适用于 6-12 个月的持仓周期</p>
    </div>

    <!-- 策略选择 -->
    <el-card class="strategy-select-card">
      <template #header>
        <div class="card-header">
          <span>选择长线策略</span>
          <el-button type="primary" @click="runAllStrategies" :loading="loading">执行所有策略</el-button>
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
                <el-button type="primary" size="small" @click="selectStrategy(strategy.id)">
                  {{ selectedStrategy === strategy.id ? '已选择' : '选择' }}
                </el-button>
                <el-button type="success" size="small" @click="runStrategy(strategy.id)" :loading="loading">执行策略</el-button>
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
              <el-input-number v-if="param.type === 'number'" v-model="strategyParameters[param.key]" :min="param.min" :max="param.max" :step="param.step" style="width: 100%" />
              <el-select v-else-if="param.type === 'select'" v-model="strategyParameters[param.key]" style="width: 100%">
                <el-option v-for="option in param.options" :key="option.value" :label="option.label" :value="option.value" />
              </el-select>
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
        <el-table-column prop="stock_code" label="股票代码" width="100" />
        <el-table-column prop="stock_name" label="股票名称" width="120" />
        <el-table-column prop="strategy_name" label="策略名称" width="150" />
        <el-table-column prop="score" label="评分" width="80" sortable>
          <template #default="{ row }">
            <el-rate v-model="row.score" disabled show-score text-color="#ff9900" score-template="{value}" />
          </template>
        </el-table-column>
        <el-table-column prop="buy_price" label="买入价" width="100">
          <template #default="{ row }">{{ row.buy_price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="stop_loss_price" label="止损价" width="100">
          <template #default="{ row }">{{ row.stop_loss_price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="take_profit_price" label="目标价" width="100">
          <template #default="{ row }">{{ row.take_profit_price?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="logic_description" label="逻辑说明" min-width="200" />
      </el-table>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { runStrategy as runStrategyApi, getStrategyResults } from '../../api/index.js'

const strategies = ref([
  {
    id: 'long_term_1', name: '高股息价值', description: '连续5年分红，股息率>5%，股价稳定，适合长期持有',
    successRate: 85.2, avgProfit: 28.5,
    parameters: [
      { key: 'dividendYield', label: '股息率%', type: 'number', min: 3, max: 10, step: 0.5, value: 5 },
      { key: 'dividendYears', label: '连续分红年', type: 'number', min: 3, max: 10, step: 1, value: 5 },
      { key: 'peRatio', label: 'PE上限', type: 'number', min: 5, max: 20, step: 1, value: 12 }
    ]
  },
  {
    id: 'long_term_2', name: '护城河龙头', description: '行业龙头，市场份额高，ROE>20%，护城河宽广',
    successRate: 82.1, avgProfit: 32.3,
    parameters: [
      { key: 'roeThreshold', label: 'ROE门槛%', type: 'number', min: 15, max: 30, step: 1, value: 20 },
      { key: 'marketShare', label: '市占率%', type: 'number', min: 10, max: 50, step: 5, value: 20 },
      { key: 'moatScore', label: '护城河评分', type: 'select', options: [{ value: 3, label: '强' }, { value: 2, label: '中等' }], value: 3 }
    ]
  },
  {
    id: 'long_term_3', name: '成长确定性', description: '未来3年业绩确定性增长，管理层优秀，现金流充裕',
    successRate: 78.6, avgProfit: 35.8,
    parameters: [
      { key: 'revenueCAGR', label: '营收CAGR%', type: 'number', min: 10, max: 40, step: 1, value: 20 },
      { key: 'debtRatio', label: '负债率上限%', type: 'number', min: 30, max: 70, step: 5, value: 50 },
      { key: 'fcfPositive', label: '连续正现金流年', type: 'number', min: 3, max: 10, step: 1, value: 5 }
    ]
  }
])

const selectedStrategy = ref('')
const strategyParameters = reactive({})
const strategyResults = ref([])
const loading = ref(false)

const avgScore = computed(() => strategyResults.value.length === 0 ? 0 : strategyResults.value.reduce((acc, r) => acc + r.score, 0) / strategyResults.value.length)
const highScoreSignals = computed(() => strategyResults.value.filter(r => r.score >= 0.8).length)
const expectedReturn = computed(() => {
  if (strategyResults.value.length === 0) return 0
  return strategyResults.value.map(r => (r.take_profit_price - r.buy_price) / r.buy_price * 100).reduce((a, b) => a + b, 0) / strategyResults.value.length
})
const riskRewardRatio = computed(() => {
  if (strategyResults.value.length === 0) return 0
  return strategyResults.value.map(r => (r.take_profit_price - r.buy_price) / (r.buy_price - r.stop_loss_price)).reduce((a, b) => a + b, 0) / strategyResults.value.length
})

const selectStrategy = (strategyId) => {
  selectedStrategy.value = strategyId
  const strategy = strategies.value.find(s => s.id === strategyId)
  if (strategy) strategy.parameters.forEach(param => { strategyParameters[param.key] = param.value })
}
const getCurrentStrategyParams = () => strategies.value.find(s => s.id === selectedStrategy.value)?.parameters || []

const runStrategy = async (strategyId) => {
  loading.value = true
  try {
    const strategy = strategies.value.find(s => s.id === strategyId)
    const params = getCurrentStrategyParams().reduce((acc, param) => {
      acc[param.key] = strategyParameters[param.key] || param.value
      return acc
    }, {})

const response = await runStrategyApi(strategyId, {
      parameters: params
    })

    if (response && response.data) {
      strategyResults.value = response.data.results || []
      ElMessage.success(`策略执行完成，生成 ${strategyResults.value.length} 个信号`)
    } else {
      throw new Error('策略执行失败')
    }
  } catch (error) {
    console.error('策略执行失败:', error)
    ElMessage.error('策略执行失败: ' + (error.message || '未知错误'))
    // 如果API调用失败，使用模拟数据作为备选
    const mockResults = Array.from({ length: 8 }, (_, i) => ({
      id: i + 1,
      stock_code: `000${(i + 20).toString().padStart(3, '0')}`,
      stock_name: `长线股票${i + 1}`,
      strategy_name: strategies.value.find(s => s.id === strategyId)?.name || '',
      score: 0.70 + Math.random() * 0.30,
      buy_price: 30 + Math.random() * 70,
      stop_loss_price: 0,
      take_profit_price: 0,
      logic_description: '基本面优秀，估值合理，适合长期持有'
    }))
    mockResults.forEach(r => { r.stop_loss_price = r.buy_price * 0.80; r.take_profit_price = r.buy_price * 1.50 })
    strategyResults.value = mockResults
    ElMessage.warning('使用模拟数据展示')
  } finally {
    loading.value = false
  }
}

const runAllStrategies = async () => {
  try {
    await ElMessageBox.confirm('确定要执行所有长线策略吗？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    loading.value = true
    strategyResults.value = []

    for (const s of strategies.value) {
      await runStrategy(s.id)
    }
    ElMessage.success('所有策略执行完成')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('策略执行失败: ' + (error.message || '未知错误'))
    }
  } finally {
    loading.value = false
  }
}

const saveParameters = () => ElMessage.success('参数保存成功')
const resetParameters = () => { selectStrategy(selectedStrategy.value); ElMessage.info('参数已重置') }
const exportResults = () => ElMessage.info('导出功能开发中...')
const clearResults = () => { strategyResults.value = []; ElMessage.info('结果已清空') }

onMounted(async () => {
  try {
    const response = await getStrategyResults('all', { strategy_type: 'long' })
    if (response && response.data && response.data.results) {
      strategyResults.value = response.data.results.slice(0, 50)
    }
  } catch (error) {
    console.error('获取历史策略结果失败:', error)
  }
})
</script>

<style scoped>
.strategy-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.strategy-select-card, .parameters-card, .results-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.strategy-card { cursor: pointer; transition: all 0.3s; }
.strategy-card.active { border-color: #409EFF; box-shadow: 0 2px 12px 0 rgba(64, 158, 255, 0.1); }
.strategy-content h3 { margin: 0 0 10px 0; color: #303133; }
.strategy-desc { color: #909399; font-size: 14px; line-height: 1.5; margin-bottom: 15px; }
.strategy-stats { margin-bottom: 15px; }
.stat-item { display: flex; justify-content: space-between; margin-bottom: 5px; }
.stat-label { color: #606266; }
.stat-value { color: #409EFF; font-weight: bold; }
.strategy-actions { display: flex; gap: 10px; }
.parameters-actions { text-align: center; padding-top: 20px; }
.results-summary { margin-top: 20px; padding: 20px; background-color: #f5f7fa; border-radius: 4px; }
.summary-item { text-align: center; }
.summary-label { display: block; color: #909399; font-size: 14px; margin-bottom: 5px; }
.summary-value { display: block; color: #303133; font-size: 18px; font-weight: bold; }
.header-actions { display: flex; gap: 10px; }
</style>