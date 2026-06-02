# 前端API集成修改指南

## 问题诊断

当前前端页面都在使用**模拟数据（mock data）**，所有真实的API调用都被注释或未实现，导致F12中看不到任何接口请求。

## 已完成的工作

✅ **已创建API配置文件**: `frontend/src/api/index.js`
- 配置了axios实例
- 实现了所有后端API的调用函数

✅ **已修改的页面**:
1. `frontend/src/views/Dashboard.vue` - ✅ 已修改
2. `frontend/src/views/stocks/StockList.vue` - ✅ 已修改
3. `frontend/src/views/strategies/ShortTermStrategy.vue` - ✅ 已修改
4. `frontend/src/views/stocks/Industries.vue` - ✅ 已修改

## 待修改的页面

以下页面仍需要使用模拟数据，需要修改为调用真实API：

### 1. MediumTermStrategy.vue (中期策略)
**文件路径**: `frontend/src/views/strategies/MediumTermStrategy.vue`

**修改步骤**:
1. 替换script部分的导入语句：
```javascript
// 删除
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 替换为
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getStrategies, runStrategy as runStrategyApi, getStrategyResults } from '../../api/index.js'
```

2. 删除模拟数据，改为从API加载：
```javascript
// 删除整个 strategies 的模拟数据定义
const strategies = ref([
  {
    id: 'medium_term_1', 
    // ... 模拟数据
  }
])

// 替换为
const strategies = ref([])
const loading = ref(false)

// 添加加载策略的方法
const loadStrategies = async () => {
  try {
    loading.value = true
    const response = await getStrategies({ type: 'medium_term' })
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

// 在 onMounted 中调用
onMounted(() => {
  loadStrategies()
})
```

3. 修改 `runStrategy` 方法：
```javascript
const runStrategy = async (strategyId) => {
  loading.value = true
  try {
    ElMessage.info(`开始执行策略: ${strategyId}`)
    
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
```

4. 修改 `runAllStrategies` 方法：
```javascript
const runAllStrategies = async () => {
  try {
    await ElMessageBox.confirm('确定要执行所有中线策略吗？', '提示', {
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
```

5. 修改计算属性，添加 toFixed：
```javascript
const avgScore = computed(() => {
  if (strategyResults.value.length === 0) return 0
  const sum = strategyResults.value.reduce((acc, r) => acc + r.score, 0)
  return (sum / strategyResults.value.length).toFixed(2)
})
```

---

### 2. LongTermStrategy.vue (长期策略)
**文件路径**: `frontend/src/views/strategies/LongTermStrategy.vue`

**修改方法**: 与 MediumTermStrategy.vue 相同，只需要将 `type: 'medium_term'` 改为 `type: 'long_term'`

---

### 3. SectorFundFlow.vue (板块资金流)
**文件路径**: `frontend/src/views/fund-flow/SectorFundFlow.vue`

**修改步骤**:
1. 导入API：
```javascript
import { getSectorFundFlow, getSectorHistory } from '../../api/index.js'
```

2. 删除模拟数据，改为从API加载：
```javascript
const loadSectorFundFlow = async () => {
  try {
    loading.value = true
    const response = await getSectorFundFlow()
    if (response && response.data) {
      allSectors.value = response.data.map(item => ({
        name: item.sector_name || item.name,
        netInflow: item.net_inflow || 0
      }))
      
      // 计算统计信息
      inflowSectors.value = allSectors.value.filter(s => s.netInflow > 0).length
      outflowSectors.value = allSectors.value.filter(s => s.netInflow < 0).length
      totalInflow.value = (allSectors.value.reduce((sum, s) => sum + s.netInflow, 0) / 100000000).toFixed(1)
    }
  } catch (error) {
    console.error('加载板块资金流失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}
```

---

### 4. FundFlowSummary.vue (资金流摘要)
**文件路径**: `frontend/src/views/fund-flow/FundFlowSummary.vue`

**修改步骤**:
1. 导入API：
```javascript
import { getFundFlowSummary } from '../../api/index.js'
```

2. 从API加载数据：
```javascript
const loadSummary = async () => {
  try {
    const response = await getFundFlowSummary()
    if (response && response.data) {
      // 更新数据
      Object.assign(summaryData.value, response.data)
    }
  } catch (error) {
    console.error('加载摘要数据失败:', error)
  }
}
```

---

### 5. FundFlowTrend.vue (资金流趋势)
**文件路径**: `frontend/src/views/fund-flow/FundFlowTrend.vue`

**修改步骤**:
1. 导入API：
```javascript
import { getFundFlowTrend } from '../../api/index.js'
```

2. 从API加载趋势数据：
```javascript
const loadTrend = async () => {
  try {
    const response = await getFundFlowTrend({
      sector: selectedSector.value,
      period: selectedPeriod.value
    })
    if (response && response.data) {
      trendData.value = response.data
      initChart()
    }
  } catch (error) {
    console.error('加载趋势数据失败:', error)
  }
}
```

---

### 6. StrategyHistory.vue (策略历史)
**文件路径**: `frontend/src/views/strategies/StrategyHistory.vue`

**修改步骤**:
1. 导入API：
```javascript
import { getStrategyHistory } from '../../api/index.js'
```

2. 从API加载历史数据：
```javascript
const loadHistory = async () => {
  try {
    const response = await getStrategyHistory(selectedStrategy.value)
    if (response && response.data) {
      historyData.value = response.data
    }
  } catch (error) {
    console.error('加载历史数据失败:', error)
  }
}
```

---

### 7. Settings.vue (系统设置)
**文件路径**: `frontend/src/views/Settings.vue`

**修改步骤**:
1. 导入API：
```javascript
import { getSystemStatus, runTaskImmediately } from '../../api/index.js'
```

2. 从API加载系统状态和执行任务

---

## 通用修改模式

每个页面都需要进行以下修改：

### 1. 导入API配置
```javascript
import { 
  getStockList, 
  getStrategies, 
  runStrategy,
  getSectorFundFlow,
  // ... 其他需要的API
} from '../../api/index.js'
```

### 2. 删除模拟数据
```javascript
// ❌ 删除
const mockData = [...]

// ✅ 改用 ref 定义空数组
const data = ref([])
const loading = ref(false)
```

### 3. 添加数据加载方法
```javascript
const loadData = async () => {
  try {
    loading.value = true
    const response = await apiFunction(params)
    
    if (response && response.data) {
      data.value = response.data
    }
  } catch (error) {
    console.error('加载失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
```

### 4. 修改事件处理方法
```javascript
const handleAction = async () => {
  try {
    loading.value = true
    const response = await apiFunction(params)
    ElMessage.success('操作成功')
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    loading.value = false
  }
}
```

---

## 测试步骤

完成所有修改后，按以下步骤测试：

1. **启动前端开发服务器**:
```bash
cd frontend
npm run dev
```

2. **打开浏览器，按F12打开开发者工具**

3. **切换到Network标签页**

4. **访问各个页面，观察是否有API请求**:
   - 仪表盘: `/api/system/status`
   - 股票列表: `/api/stocks/list`
   - 策略页面: `/api/strategies`
   - 资金流页面: `/api/fund-flow/sectors`

5. **检查请求和响应**:
   - 请求URL是否正确
   - 请求参数是否正确
   - 响应数据是否正确

---

## 常见问题

### Q1: 修改后页面报错
**A**: 检查浏览器控制台，查看具体错误信息。通常是：
- API导入路径错误
- API函数名错误
- 响应数据结构不匹配

### Q2: API请求失败
**A**: 检查：
- 后端服务是否正常运行
- API URL是否正确（查看浏览器Network）
- 后端CORS配置是否正确

### Q3: 数据没有显示
**A**: 检查：
- API响应数据结构是否与前端期望的一致
- 是否需要 .data 或 .data.list 来获取数据
- 控制台是否有错误信息

---

## 完成标准

✅ 所有页面都能看到API请求（F12 Network）  
✅ 所有API请求都返回200状态码  
✅ 页面数据能正确显示  
✅ 没有控制台错误  

---

## 需要帮助？

如果在修改过程中遇到问题，可以：
1. 查看浏览器控制台错误信息
2. 查看后端日志
3. 使用Postman测试API接口
4. 向我询问具体的修改方法

祝修改顺利！🚀
