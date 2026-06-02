<template>
  <div class="stock-list-page">
    <div class="page-header">
      <h1>股票列表</h1>
      <p class="subtitle">浏览和筛选所有监控的股票</p>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :model="filterForm" label-width="80px">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="股票代码">
              <el-input
                v-model="filterForm.code"
                placeholder="输入股票代码"
                clearable
                @keyup.enter="searchStocks"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="股票名称">
              <el-input
                v-model="filterForm.name"
                placeholder="输入股票名称"
                clearable
                @keyup.enter="searchStocks"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="行业分类">
              <el-select
                v-model="filterForm.industry"
                placeholder="选择行业"
                clearable
                style="width: 100%"
              >
                <el-option
                  v-for="industry in industries"
                  :key="industry"
                  :label="industry"
                  :value="industry"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="市场类型">
              <el-select
                v-model="filterForm.market"
                placeholder="选择市场"
                clearable
                style="width: 100%"
              >
                <el-option label="沪深A股" value="A" />
                <el-option label="上海主板" value="SH" />
                <el-option label="深圳主板" value="SZ" />
                <el-option label="创业板" value="CYB" />
                <el-option label="科创板" value="KCB" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="成交额">
              <el-input-number
                v-model="filterForm.minAmount"
                placeholder="最小成交额"
                :min="0"
                :step="1000000"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="换手率">
              <el-input-number
                v-model="filterForm.minTurnover"
                placeholder="最小换手率"
                :min="0"
                :step="0.1"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="排序方式">
              <el-select
                v-model="filterForm.sortBy"
                style="width: 100%"
              >
                <el-option label="代码排序" value="code" />
                <el-option label="名称排序" value="name" />
                <el-option label="最新价格" value="price" />
                <el-option label="涨跌幅" value="changeRate" />
                <el-option label="成交额" value="amount" />
                <el-option label="换手率" value="turnover" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="排序方向">
              <el-radio-group v-model="filterForm.sortOrder">
                <el-radio label="asc">升序</el-radio>
                <el-radio label="desc">降序</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <div class="filter-actions">
          <el-button type="primary" @click="searchStocks">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetFilter">重置</el-button>
          <el-button type="success" @click="exportData">导出数据</el-button>
        </div>
      </el-form>
    </el-card>

    <!-- 股票列表 -->
    <el-card class="list-card">
      <template #header>
        <div class="list-header">
          <span>股票列表 ({{ totalCount }} 只股票)</span>
          <div class="header-actions">
            <el-button type="text" @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button type="text" @click="toggleAutoRefresh">
              <el-icon><Timer /></el-icon>
              {{ autoRefresh ? '停止自动刷新' : '自动刷新' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="stockList"
        v-loading="loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        stripe
      >
        <el-table-column prop="code" label="代码" width="100" sortable="custom" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="industry" label="行业" width="120" />
        <el-table-column prop="market" label="市场" width="80">
          <template #default="{ row }">
            <el-tag :type="getMarketType(row.market)">{{ row.market }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="closePrice" label="最新价" width="100" sortable="custom">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.changeRate > 0, 'price-down': row.changeRate < 0 }">
              {{ row.closePrice.toFixed(2) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="changeRate" label="涨跌幅" width="100" sortable="custom">
          <template #default="{ row }">
            <span :class="{ 'price-up': row.changeRate > 0, 'price-down': row.changeRate < 0 }">
              {{ (row.changeRate * 100).toFixed(2) }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="成交额" width="120" sortable="custom">
          <template #default="{ row }">
            {{ formatAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column prop="turnoverRate" label="换手率" width="100" sortable="custom">
          <template #default="{ row }">
            {{ (row.turnoverRate * 100).toFixed(2) }}%
          </template>
        </el-table-column>
        <el-table-column prop="peRatio" label="市盈率" width="100" sortable="custom" />
        <el-table-column prop="pbRatio" label="市净率" width="100" sortable="custom" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="viewDetail(row)">详情</el-button>
            <el-button type="success" size="small" @click="runStrategy(row)">策略分析</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="totalCount"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Timer } from '@element-plus/icons-vue'
import { getStockList, getIndustries } from '../../api/index.js'

const router = useRouter()

// 筛选条件
const filterForm = reactive({
  code: '',
  name: '',
  industry: '',
  market: '',
  minAmount: 0,
  minTurnover: 0,
  sortBy: 'code',
  sortOrder: 'asc'
})

// 分页参数
const currentPage = ref(1)
const pageSize = ref(50)
const totalCount = ref(0)

// 数据状态
const stockList = ref([])
const industries = ref([])
const loading = ref(false)
const autoRefresh = ref(false)
let refreshTimer = null

// 方法
const searchStocks = async () => {
  loading.value = true
  try {
    // 调用真实API
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      code: filterForm.code || undefined,
      name: filterForm.name || undefined,
      industry: filterForm.industry || undefined,
      market: filterForm.market || undefined,
      min_amount: filterForm.minAmount > 0 ? filterForm.minAmount : undefined,
      min_turnover: filterForm.minTurnover > 0 ? filterForm.minTurnover : undefined,
      sort_by: filterForm.sortBy,
      sort_order: filterForm.sortOrder
    }

    const response = await getStockList(params)
    
    if (response && response.data) {
      stockList.value = response.data.list || []
      totalCount.value = response.data.total || 0
    } else if (Array.isArray(response)) {
      stockList.value = response
      totalCount.value = response.length
    } else {
      stockList.value = []
      totalCount.value = 0
    }

  } catch (error) {
    console.error('获取股票列表失败:', error)
    ElMessage.error('获取股票列表失败')
  } finally {
    loading.value = false
  }
}

// 初始化行业列表
const initIndustries = async () => {
  try {
    const response = await getIndustries()
    if (response && response.data) {
      industries.value = response.data || []
    } else if (Array.isArray(response)) {
      industries.value = response
    }
  } catch (error) {
    console.error('获取行业列表失败:', error)
    ElMessage.error('获取行业列表失败')
  }
}

const resetFilter = () => {
  Object.keys(filterForm).forEach(key => {
    if (key === 'sortBy' || key === 'sortOrder') return
    filterForm[key] = ''
  })
  filterForm.minAmount = 0
  filterForm.minTurnover = 0
  currentPage.value = 1
  searchStocks()
}

const exportData = () => {
  ElMessage.info('导出功能开发中...')
}

const refreshData = () => {
  searchStocks()
  ElMessage.success('数据已刷新')
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshTimer = setInterval(refreshData, 30000) // 30秒刷新一次
    ElMessage.info('已开启自动刷新')
  } else {
    clearInterval(refreshTimer)
    ElMessage.info('已关闭自动刷新')
  }
}

const handleSortChange = ({ prop, order }) => {
  filterForm.sortBy = prop
  filterForm.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  searchStocks()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  searchStocks()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  searchStocks()
}

const viewDetail = (stock) => {
  router.push(`/stocks/${stock.code}`)
}

const runStrategy = (stock) => {
  ElMessage.info(`对股票 ${stock.code} 执行策略分析...`)
  // 跳转到策略分析页面
}

const getMarketType = (market) => {
  const types = {
    'SH': 'primary',
    'SZ': 'success',
    'CYB': 'warning',
    'KCB': 'danger'
  }
  return types[market] || 'info'
}

const formatAmount = (amount) => {
  if (amount >= 100000000) {
    return `${(amount / 100000000).toFixed(2)}亿`
  } else if (amount >= 10000) {
    return `${(amount / 10000).toFixed(2)}万`
  }
  return amount.toFixed(0)
}

// 初始化行业列表
const initIndustries = () => {
  industries.value = [...new Set(mockStocks.map(stock => stock.industry))]
}

onMounted(() => {
  initIndustries()
  searchStocks()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.stock-list-page {
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

.filter-card {
  margin-bottom: 20px;
}

.filter-actions {
  text-align: center;
  padding-top: 10px;
}

.list-card {
  margin-bottom: 20px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.price-up {
  color: #f56c6c;
}

.price-down {
  color: #67c23a;
}
</style>