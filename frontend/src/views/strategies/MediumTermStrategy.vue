<template>
  <div class="strategy-page">
    <div class="page-header">
      <h1>中线策略分析</h1>
      <p class="subtitle">基于中线逻辑的交易策略，适用于 1-3 个月的持仓周期</p>
    </div>

    <!-- 策略选择 -->
    <el-card class="strategy-select-card">
      <template #header>
        <div class="card-header">
          <span>选择中线策略</span>
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
                <el-button type="primary" size="small" @click="selectStrategy(strategy.id)">
                  {{ selectedStrategy === strategy.id ? '已选择' : '选择' }}
                </el-button>
                <el-button type="success" size="small" @click="runStrategy(strategy.id)">执行策略</el-button>
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
        <el-table-column prop="stockCode" label="股票代码" width="100" />
        <el-table-column prop="stockName" label="股票名称" width="120" />
        <el-table-column prop="strategyName" label="策略名称" width="150" />
        <el-table-column prop="score" label="评分" width="80" sortable>
          <template #default="{ row }">
            <el-rate v-model="row.score" disabled show-score text-color="#ff9900" score-template="{value}" />
          </template>
        </el-table-column>
        <el-table-column prop="buyPrice" label="买入价" width="100">
          <template #default="{ row }">{{ row.buyPrice.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="stopLossPrice" label="止损价" width="100">
          <template #default="{ row }">{{ row.stopLossPrice.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="takeProfitPrice" label="止盈价" width="100">
          <template #default="{ row }">{{ row.takeProfitPrice.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="logicDescription" label="逻辑说明" min-width="200" />
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
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const strategies = ref([
  {
    id: 'medium_term_1', name: '低估值修复', description: 'PE/PB低于行业均值20%，近期营收改善，机构开始关注',
    successRate: 72.3, avgProfit: 15.6,
    parameters: [
      { key: 'peRatio', label: '目标PE上限', type: 'number', min: 5, max: 30, step: 1, value: 15 },
      { key: 'pbRatio', label: '目标PB上限', type: 'number', min: 0.5, max: 3, step: 0.1, value: 1.5 },
      { key: 'revenueGrowth', label: '营收增长%', type: 'number', min: 0, max: 30, step: 1, value: 10 }
    ]
  },
  {
    id: 'medium_term_2', name: '业绩持续增长', description: '连续2-3个季度业绩增长，ROE>15%，机构持仓增加',
    successRate: 70.1, avgProfit: 18.2,
    parameters: [
      { key: 'roeThreshold', label: 'ROE门槛%', type: 'number', min: 10, max: 25, step: 1, value: 15 },
      { key: 'quarterCount', label: '连续季度', type: 'number', min: 2, max: 6, step: 1, value: 3 },
      { key: 'institutionGrowth', label: '机构增仓%', type: 'number', min: 0, max: 20, step: 1, value: 5 }
    ]
  },
  {
    id: 'medium_term_3', name: '行业景气回升', description: '行业处于复苏周期，产品价格企稳回升，龙头企业受益',
    successRate: 68.5, avgProfit: 16.8,
    parameters: [
      { key: 'industryPosition', label: '行业地位', type: 'select', options: [{ value: 'leader', label: '龙头' }, { value: 'top3', label: '前三' }], value: 'leader' },
      { key: 'priceRecovery', label: '产品回升%', type: 'number', min: 0, max: 30, step: 1, value: 10 },
      { key: 'capacityUtilization', label: '产能利用率%', type: 'number', min: 50, max: 100, step: 5, value: 70 }
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
  return strategyResults.value.map(r => (r.takeProfitPrice - r.buyPrice) / r.buyPrice * 100).reduce((a, b) => a + b, 0) / strategyResults.value.length
})
const riskRewardRatio = computed(() => {
  if (strategyResults.value.length === 0) return 0
  return strategyResults.value.map(r => (r.takeProfitPrice - r.buyPrice) / (r.buyPrice - r.stopLossPrice)).reduce((a, b) => a + b, 0) / strategyResults.value.length
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
    await new Promise(resolve => setTimeout(resolve, 2000))
    const mockResults = Array.from({ length: 12 }, (_, i) => ({
      id: i + 1, stockCode: `600${(i + 1).toString().padStart(3, '0')}`, stockName: `中线股票${i + 1}`,
      strategyName: strategies.value.find(s => s.id === strategyId)?.name || '',
      score: 0.65 + Math.random() * 0.35, buyPrice: 20 + Math.random() * 80,
      stopLossPrice: 0, takeProfitPrice: 0,
      logicDescription: '中线趋势确立，基本面改善，符合策略买入条件'
    }))
    mockResults.forEach(r => { r.stopLossPrice = r.buyPrice * 0.88; r.takeProfitPrice = r.buyPrice * 1.20 })
    strategyResults.value = mockResults
    ElMessage.success(`策略执行完成，生成 ${mockResults.length} 个信号`)
  } catch { ElMessage.error('策略执行失败') }
  finally { loading.value = false }
}
const runAllStrategies = async () => {
  try {
    await ElMessageBox.confirm('确定要执行所有中线策略吗？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    loading.value = true; strategyResults.value = []
    for (const s of strategies.value) await runStrategy(s.id)
    ElMessage.success('所有策略执行完成')
  } catch (error) { if (error !== 'cancel') ElMessage.error('策略执行失败') }
  finally { loading.value = false }
}
const saveParameters = () => ElMessage.success('参数保存成功')
const resetParameters = () => { selectStrategy(selectedStrategy.value); ElMessage.info('参数已重置') }
const exportResults = () => ElMessage.info('导出功能开发中...')
const clearResults = () => { strategyResults.value = []; ElMessage.info('结果已清空') }
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
