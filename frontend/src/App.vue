<template>
  <div id="app">
    <el-container class="app-container">
      <!-- 侧边栏导航 -->
      <el-aside width="250px" class="sidebar">
        <div class="logo">
          <h2>📈 股票策略分析</h2>
        </div>
        <el-menu
          :default-active="currentRoute"
          class="sidebar-menu"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <el-menu-item index="/dashboard">
            <el-icon><House /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>

          <el-sub-menu index="stocks">
            <template #title>
              <el-icon><TrendCharts /></el-icon>
              <span>股票管理</span>
            </template>
            <el-menu-item index="/stocks/list">股票列表</el-menu-item>
            <el-menu-item index="/stocks/search">股票搜索</el-menu-item>
            <el-menu-item index="/stocks/industries">行业分类</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="strategies">
            <template #title>
              <el-icon><MagicStick /></el-icon>
              <span>策略分析</span>
            </template>
            <el-menu-item index="/strategies/short-term">短线策略</el-menu-item>
            <el-menu-item index="/strategies/medium-term">中线策略</el-menu-item>
            <el-menu-item index="/strategies/long-term">长线策略</el-menu-item>
            <el-menu-item index="/strategies/history">历史记录</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="fund-flow">
            <template #title>
              <el-icon><Money /></el-icon>
              <span>资金流向</span>
            </template>
            <el-menu-item index="/fund-flow/sectors">板块资金流</el-menu-item>
            <el-menu-item index="/fund-flow/trend">资金流趋势</el-menu-item>
            <el-menu-item index="/fund-flow/summary">资金流摘要</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区域 -->
      <el-container>
        <el-header class="header">
          <div class="header-content">
            <div class="header-left">
              <span class="current-date">{{ currentDate }}</span>
            </div>
            <div class="header-right">
              <el-button type="primary" @click="refreshData">
                <el-icon><Refresh /></el-icon>
                刷新数据
              </el-button>
              <el-button @click="showSystemStatus">
                <el-icon><InfoFilled /></el-icon>
                系统状态
              </el-button>
            </div>
          </div>
        </el-header>

        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  House,
  TrendCharts,
  MagicStick,
  Money,
  Setting,
  Refresh,
  InfoFilled
} from '@element-plus/icons-vue'

const route = useRoute()

// 计算当前路由
const currentRoute = computed(() => route.path)

// 当前日期
const currentDate = ref('')

// 刷新数据
const refreshData = async () => {
  try {
    ElMessage.info('正在刷新数据...')
    // 调用后端API刷新数据
    // await api.refreshData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败')
  }
}

// 显示系统状态
const showSystemStatus = async () => {
  try {
    // 获取系统状态
    // const status = await api.getSystemStatus()
    await ElMessageBox.alert('系统运行正常', '系统状态', {
      confirmButtonText: '确定'
    })
  } catch (error) {
    ElMessage.error('获取系统状态失败')
  }
}

// 初始化
onMounted(() => {
  // 设置当前日期
  const now = new Date()
  currentDate.value = now.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})
</script>

<style scoped>
.app-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  color: white;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
}

.logo h2 {
  margin: 0;
  color: #409EFF;
  font-size: 18px;
}

.sidebar-menu {
  border: none;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.current-date {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.main-content {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>