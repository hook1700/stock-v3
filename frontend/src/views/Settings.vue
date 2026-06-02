
<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>系统设置</h1>
      <p class="subtitle">配置股票策略分析系统的各项参数</p>
    </div>

    <el-row :gutter="20">
      <!-- 基本设置 -->
      <el-col :span="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header"><span>基本设置</span></div>
          </template>
          <el-form :model="basicSettings" label-width="140px" v-loading="basicLoading">
            <el-form-item label="系统名称">
              <el-input v-model="basicSettings.systemName" />
            </el-form-item>
            <el-form-item label="数据刷新间隔(秒)">
              <el-input-number v-model="basicSettings.refreshInterval" :min="5" :max="300" :step="5" />
            </el-form-item>
            <el-form-item label="默认股票市场">
              <el-select v-model="basicSettings.defaultMarket" style="width: 100%">
                <el-option label="沪深A股" value="A" />
                <el-option label="上海主板" value="SH" />
                <el-option label="深圳主板" value="SZ" />
                <el-option label="创业板" value="CYB" />
                <el-option label="科创板" value="KCB" />
              </el-select>
            </el-form-item>
            <el-form-item label="显示精度">
              <el-radio-group v-model="basicSettings.precision">
                <el-radio label="2">2位小数</el-radio>
                <el-radio label="3">3位小数</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="主题模式">
              <el-radio-group v-model="basicSettings.theme">
                <el-radio-button label="light">浅色</el-radio-button>
                <el-radio-button label="dark">深色</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveBasicSettings" :loading="basicLoading">保存设置</el-button>
              <el-button @click="resetBasicSettings">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 通知设置 -->
      <el-col :span="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header"><span>通知设置</span></div>
          </template>
          <el-form :model="notificationSettings" label-width="140px" v-loading="notifyLoading">
            <el-form-item label="启用策略通知">
              <el-switch v-model="notificationSettings.strategyNotify" />
            </el-form-item>
            <el-form-item label="信号推送方式">
              <el-checkbox-group v-model="notificationSettings.notifyChannels">
                <el-checkbox label="web">网页弹窗</el-checkbox>
                <el-checkbox label="email">邮件</el-checkbox>
                <el-checkbox label="sms">短信</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="评分阈值">
              <el-slider v-model="notificationSettings.scoreThreshold" :min="0" :max="5" :step="0.1" show-stops />
            </el-form-item>
            <el-form-item label="每日汇总时间">
              <el-time-picker v-model="notificationSettings.dailySummaryTime" format="HH:mm" placeholder="选择时间" />
            </el-form-item>
            <el-form-item label="异常波动提醒">
              <el-switch v-model="notificationSettings.abnormalAlert" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveNotifySettings" :loading="notifyLoading">保存设置</el-button>
              <el-button @click="resetNotifySettings">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <!-- 数据采集设置 -->
    <el-card class="settings-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header"><span>数据采集设置</span></div>
      </template>
      <el-form :model="dataSettings" label-width="140px" v-loading="dataLoading">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="日线数据更新">
              <el-time-picker v-model="dataSettings.dailyUpdateTime" format="HH:mm" placeholder="选择时间" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="实时数据间隔(秒)">
              <el-input-number v-model="dataSettings.realtimeInterval" :min="1" :max="60" :step="1" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="历史数据保留(天)">
              <el-input-number v-model="dataSettings.historyRetention" :min="30" :max="3650" :step="30" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="数据源">
              <el-select v-model="dataSettings.dataSource" style="width: 100%">
                <el-option label="BaoStock" value="baostock" />
                <el-option label="AKShare" value="akshare" />
                <el-option label="Tushare" value="tushare" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="自动补全缺失数据">
              <el-switch v-model="dataSettings.autoFillMissing" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item>
              <el-button type="primary" @click="saveDataSettings" :loading="dataLoading">保存采集设置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- API 配置 -->
    <el-card class="settings-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header"><span>API 配置</span></div>
      </template>
      <el-form :model="apiSettings" label-width="140px" v-loading="apiLoading">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="后端API地址">
              <el-input v-model="apiSettings.baseUrl" placeholder="http://localhost:8080/api" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="API密钥">
              <el-input v-model="apiSettings.apiKey" type="password" show-password placeholder="输入API密钥" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="请求超时(秒)">
              <el-input-number v-model="apiSettings.timeout" :min="5" :max="120" :step="5" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <el-button type="primary" @click="testApiConnection" :loading="testLoading">测试连接</el-button>
              <el-button type="primary" @click="saveApiSettings" :loading="apiLoading">保存配置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 系统信息 -->
    <el-card class="settings-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header"><span>系统信息</span></div>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="系统版本">{{ systemInfo.version }}</el-descriptions-item>
        <el-descriptions-item label="前端框架">{{ systemInfo.frontendFramework }}</el-descriptions-item>
        <el-descriptions-item label="UI组件库">{{ systemInfo.uiLibrary }}</el-descriptions-item>
        <el-descriptions-item label="后端语言">{{ systemInfo.backendLanguage }}</el-descriptions-item>
        <el-descriptions-item label="数据库">{{ systemInfo.database }}</el-descriptions-item>
        <el-descriptions-item label="缓存">{{ systemInfo.cache }}</el-descriptions-item>
        <el-descriptions-item label="最后更新">{{ systemInfo.lastUpdate }}</el-descriptions-item>
        <el-descriptions-item label="维护者">{{ systemInfo.maintainer }}</el-descriptions-item>
        <el-descriptions-item label="License">{{ systemInfo.license }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { settingsApi } from '@/services/api'

const basicLoading = ref(false)
const notifyLoading = ref(false)
const dataLoading = ref(false)
const apiLoading = ref(false)
const testLoading = ref(false)

const basicSettings = reactive({
  systemName: '股票策略分析系统',
  refreshInterval: 30,
  defaultMarket: 'A',
  precision: '2',
  theme: 'light'
})

const notificationSettings = reactive({
  strategyNotify: true,
  notifyChannels: ['web'],
  scoreThreshold: 3.5,
  dailySummaryTime: new Date(2026, 0, 1, 18, 0),
  abnormalAlert: true
})

const dataSettings = reactive({
  dailyUpdateTime: new Date(2026, 0, 1, 18, 30),
  realtimeInterval: 5,
  historyRetention: 1095,
  dataSource: 'akshare',
  autoFillMissing: true
})

const apiSettings = reactive({
  baseUrl: 'http://localhost:8080/api',
  apiKey: '',
  timeout: 30
})

const systemInfo = reactive({
  version: 'v1.0.0',
  frontendFramework: 'Vue 3 + Vite',
  uiLibrary: 'Element Plus',
  backendLanguage: 'Go 1.19+',
  database: 'PostgreSQL 15',
  cache: 'Redis 7',
  lastUpdate: '2026-06-01',
  maintainer: 'v_zwhozhang',
  license: 'MIT'
})

const loadAllSettings = async () => {
  try {
    const response = await settingsApi.getAllSettings()

    if (response.data && response.data.success) {
      const data = response.data.data || {}

      // 更新基本设置
      if (data.basic) {
        Object.assign(basicSettings, data.basic)
      }

      // 更新通知设置
      if (data.notification) {
        Object.assign(notificationSettings, data.notification)
      }

      // 更新数据采集设置
      if (data.data) {
        Object.assign(dataSettings, data.data)
      }

      // 更新API配置
      if (data.api) {
        Object.assign(apiSettings, data.api)
      }

      // 更新系统信息
      if (data.system) {
        Object.assign(systemInfo, data.system)
      }

      ElMessage.success('设置加载成功')
    } else {
      throw new Error(response.data?.message || '加载失败')
    }
  } catch (error) {
    console.error('加载设置失败:', error)
    ElMessage.warning('使用本地默认设置')
    // 使用默认值，不做任何操作
  }
}

const saveBasicSettings = async () => {
  basicLoading.value = true
  try {
    const response = await settingsApi.saveBasicSettings(basicSettings)

    if (response.data && response.data.success) {
      ElMessage.success('基本设置已保存')
    } else {
      throw new Error(response.data?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存基本设置失败:', error)
    ElMessage.error('保存基本设置失败: ' + (error.message || '未知错误'))
  } finally {
    basicLoading.value = false
  }
}

const resetBasicSettings = () => {
  Object.assign(basicSettings, {
    systemName: '股票策略分析系统',
    refreshInterval: 30,
    defaultMarket: 'A',
    precision: '2',
    theme: 'light'
  })
  ElMessage.info('已重置')
}

const saveNotifySettings = async () => {
  notifyLoading.value = true
  try {
    const response = await settingsApi.saveNotificationSettings(notificationSettings)

    if (response.data && response.data.success) {
      ElMessage.success('通知设置已保存')
    } else {
      throw new Error(response.data?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存通知设置失败:', error)
    ElMessage.error('保存通知设置失败: ' + (error.message || '未知错误'))
  } finally {
    notifyLoading.value = false
  }
}

const resetNotifySettings = () => {
  Object.assign(notificationSettings, {
    strategyNotify: true,
    notifyChannels: ['web'],
    scoreThreshold: 3.5,
    dailySummaryTime: new Date(2026, 0, 1, 18, 0),
    abnormalAlert: true
  })
  ElMessage.info('已重置')
}

const saveDataSettings = async () => {
  dataLoading.value = true
  try {
    const response = await settingsApi.saveDataSettings(dataSettings)

    if (response.data && response.data.success) {
      ElMessage.success('采集设置已保存')
    } else {
      throw new Error(response.data?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存采集设置失败:', error)
    ElMessage.error('保存采集设置失败: ' + (error.message || '未知错误'))
  } finally {
    dataLoading.value = false
  }
}

const saveApiSettings = async () => {
  apiLoading.value = true
  try {
    const response = await settingsApi.saveApiSettings(apiSettings)

    if (response.data && response.data.success) {
      ElMessage.success('API配置已保存')
    } else {
      throw new Error(response.data?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存API配置失败:', error)
    ElMessage.error('保存API配置失败: ' + (error.message || '未知错误'))
  } finally {
    apiLoading.value = false
  }
}

const testApiConnection = async () => {
  testLoading.value = true
  try {
    const response = await settingsApi.testConnection(apiSettings)

    if (response.data && response.data.success) {
      ElMessage.success('连接成功')
    } else {
      throw new Error(response.data?.message || '连接失败')
    }
  } catch (error) {
    console.error('API连接测试失败:', error)
    ElMessage.error('连接失败: ' + (error.message || '未知错误'))
  } finally {
    testLoading.value = false
  }
}

onMounted(() => {
  loadAllSettings()
})
</script>

<style scoped>
.settings-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.settings-card { margin-bottom: 0; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
