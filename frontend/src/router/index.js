import { createRouter, createWebHistory } from 'vue-router'

// 导入页面组件
import Dashboard from '../views/Dashboard.vue'
import StockList from '../views/stocks/StockList.vue'
import StockSearch from '../views/stocks/StockSearch.vue'
import Industries from '../views/stocks/Industries.vue'
import ShortTermStrategy from '../views/strategies/ShortTermStrategy.vue'
import MediumTermStrategy from '../views/strategies/MediumTermStrategy.vue'
import LongTermStrategy from '../views/strategies/LongTermStrategy.vue'
import StrategyHistory from '../views/strategies/StrategyHistory.vue'
import SectorFundFlow from '../views/fund-flow/SectorFundFlow.vue'
import FundFlowTrend from '../views/fund-flow/FundFlowTrend.vue'
import FundFlowSummary from '../views/fund-flow/FundFlowSummary.vue'
import Settings from '../views/Settings.vue'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: {
      title: '仪表盘',
      icon: 'House'
    }
  },
  // 股票管理路由
  {
    path: '/stocks',
    redirect: '/stocks/list'
  },
  {
    path: '/stocks/list',
    name: 'StockList',
    component: StockList,
    meta: {
      title: '股票列表',
      icon: 'List'
    }
  },
  {
    path: '/stocks/search',
    name: 'StockSearch',
    component: StockSearch,
    meta: {
      title: '股票搜索',
      icon: 'Search'
    }
  },
  {
    path: '/stocks/industries',
    name: 'Industries',
    component: Industries,
    meta: {
      title: '行业分类',
      icon: 'Grid'
    }
  },
  // 策略分析路由
  {
    path: '/strategies',
    redirect: '/strategies/short-term'
  },
  {
    path: '/strategies/short-term',
    name: 'ShortTermStrategy',
    component: ShortTermStrategy,
    meta: {
      title: '短线策略',
      icon: 'Lightning'
    }
  },
  {
    path: '/strategies/medium-term',
    name: 'MediumTermStrategy',
    component: MediumTermStrategy,
    meta: {
      title: '中线策略',
      icon: 'TrendCharts'
    }
  },
  {
    path: '/strategies/long-term',
    name: 'LongTermStrategy',
    component: LongTermStrategy,
    meta: {
      title: '长线策略',
      icon: 'Clock'
    }
  },
  {
    path: '/strategies/history',
    name: 'StrategyHistory',
    component: StrategyHistory,
    meta: {
      title: '策略历史',
      icon: 'Document'
    }
  },
  // 资金流向路由
  {
    path: '/fund-flow',
    redirect: '/fund-flow/sectors'
  },
  {
    path: '/fund-flow/sectors',
    name: 'SectorFundFlow',
    component: SectorFundFlow,
    meta: {
      title: '板块资金流',
      icon: 'Money'
    }
  },
  {
    path: '/fund-flow/trend',
    name: 'FundFlowTrend',
    component: FundFlowTrend,
    meta: {
      title: '资金流趋势',
      icon: 'TrendCharts'
    }
  },
  {
    path: '/fund-flow/summary',
    name: 'FundFlowSummary',
    component: FundFlowSummary,
    meta: {
      title: '资金流摘要',
      icon: 'Summary'
    }
  },
  // 系统设置
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: {
      title: '系统设置',
      icon: 'Setting'
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 股票策略分析系统`
  }
  next()
})

export default router