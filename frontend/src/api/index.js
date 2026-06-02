import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const apiClient = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('API请求错误:', error)
    ElMessage.error(error.response?.data?.message || '请求失败，请稍后重试')
    return Promise.reject(error)
  }
)

// ==================== 股票相关API ====================

// 获取股票列表
export const getStockList = async (params) => {
  const response = await apiClient.post('/stocks/list', params)
  return response
}

// 获取股票详情
export const getStockDetail = async (code) => {
  const response = await apiClient.get(`/stocks/${code}`)
  return response
}

// 获取股票历史数据
export const getStockHistory = async (code, params) => {
  const response = await apiClient.get(`/stocks/${code}/history`, { params })
  return response
}

// 获取行业列表
export const getIndustries = async () => {
  const response = await apiClient.get('/stocks/industries')
  return response
}

// 搜索股票
export const searchStocks = async (keyword) => {
  const response = await apiClient.get('/stocks/search', {
    params: { keyword }
  })
  return response
}

// ==================== 策略相关API ====================

// 获取策略列表
export const getStrategies = async (params) => {
  const response = await apiClient.get('/strategies', { params })
  return response
}

// 获取策略结果
export const getStrategyResults = async (strategyId, params) => {
  const response = await apiClient.get(`/strategies/${strategyId}/results`, { params })
  return response
}

// 执行策略
export const runStrategy = async (strategyId) => {
  const response = await apiClient.post(`/strategies/${strategyId}/run`)
  return response
}

// 获取策略历史
export const getStrategyHistory = async (strategyId, params) => {
  const response = await apiClient.get(`/strategies/${strategyId}/history`, { params })
  return response
}

// 按类型获取策略
export const getStrategiesByType = async (type) => {
  const response = await apiClient.get(`/strategies/types/${type}`)
  return response
}

// 更新策略状态
export const updateStrategyStatus = async (strategyId, status) => {
  const response = await apiClient.put(`/strategies/${strategyId}/status`, { status })
  return response
}

// ==================== 资金流相关API ====================

// 获取板块资金流
export const getSectorFundFlow = async (params) => {
  const response = await apiClient.get('/fund-flow/sectors', { params })
  return response
}

// 获取板块历史资金流
export const getSectorHistory = async (sector, params) => {
  const response = await apiClient.get(`/fund-flow/sectors/${sector}/history`, { params })
  return response
}

// 获取资金流摘要
export const getFundFlowSummary = async () => {
  const response = await apiClient.get('/fund-flow/summary')
  return response
}

// 获取资金流趋势
export const getFundFlowTrend = async (params) => {
  const response = await apiClient.get('/fund-flow/trend', { params })
  return response
}

// ==================== 系统相关API ====================

// 获取系统状态
export const getSystemStatus = async () => {
  const response = await apiClient.get('/system/status')
  return response
}

// 立即执行任务
export const runTaskImmediately = async (taskId) => {
  const response = await apiClient.post('/system/tasks/run', { task_id: taskId })
  return response
}

// 获取任务状态
export const getTaskStatus = async () => {
  const response = await apiClient.get('/system/tasks/status')
  return response
}

// 健康检查
export const healthCheck = async () => {
  const response = await apiClient.get('/health')
  return response
}

export default {
  getStockList,
  getStockDetail,
  getStockHistory,
  getIndustries,
  searchStocks,
  getStrategies,
  getStrategyResults,
  runStrategy,
  getStrategyHistory,
  getStrategiesByType,
  updateStrategyStatus,
  getSectorFundFlow,
  getSectorHistory,
  getFundFlowSummary,
  getFundFlowTrend,
  getSystemStatus,
  runTaskImmediately,
  getTaskStatus,
  healthCheck
}
