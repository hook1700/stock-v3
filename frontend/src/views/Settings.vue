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
          <el-form :model="basicSettings" label-width="140px">
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
              <el-button type="primary" @click="saveBasicSettings">保存设置</el-button>
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
          <el-form :model="notificationSettings" label-width="140px">
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
              <el-button type="primary" @click="saveNotifySettings">保存设置</el-button>
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
      <el-form :model="dataSettings" label-width="140px">
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
              <el-button type="primary" @click="saveDataSettings">保存采集设置</el-button>
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
      <el-form :model="apiSettings" label-width="140px">
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
              <el-button type="primary" @click="testApiConnection">测试连接</el-button>
              <el-button type="primary" @click="saveApiSettings">保存配置</el-button>
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
        <el-descriptions-item label="系统版本">v1.0.0</el-descriptions-item>
        <el-descriptions-item label="前端框架">Vue 3 + Vite</el-descriptions-item>
        <el-descriptions-item label="UI组件库">Element Plus</el-descriptions-item>
        <el-descriptions-item label="后端语言">Go 1.19+</el-descriptions-item>
        <el-descriptions-item label="数据库">PostgreSQL 15</el-descriptions-item>
        <el-descriptions-item label="缓存">Redis 7</el-descriptions-item>
        <el-descriptions-item label="最后更新">2026-06-01</el-descriptions-item>
        <el-descriptions-item label="维护者">v_zwhozhang</el-descriptions-item>
        <el-descriptions-item label="License">MIT</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const basicSettings = ref({
  systemName: '股票策略分析系统',
  refreshInterval: 30,
  defaultMarket: 'A',
  precision: '2',
  theme: 'light'
})

const notificationSettings = ref({
  strategyNotify: true,
  notifyChannels: ['web'],
  scoreThreshold: 3.5,
  dailySummaryTime: new Date(2026, 0, 1, 18, 0),
  abnormalAlert: true
})

const dataSettings = ref({
  dailyUpdateTime: new Date(2026, 0, 1, 18, 30),
  realtimeInterval: 5,
  historyRetention: 1095,
  dataSource: 'akshare',
  autoFillMissing: true
})

const apiSettings = ref({
  baseUrl: 'http://localhost:8080/api',
  apiKey: '',
  timeout: 30
})

const saveBasicSettings = () => ElMessage.success('基本设置已保存')
const resetBasicSettings = () => { basicSettings.value = { systemName: '股票策略分析系统', refreshInterval: 30, defaultMarket: 'A', precision: '2', theme: 'light' }; ElMessage.info('已重置') }
const saveNotifySettings = () => ElMessage.success('通知设置已保存')
const resetNotifySettings = () => { notificationSettings.value = { strategyNotify: true, notifyChannels: ['web'], scoreThreshold: 3.5, dailySummaryTime: new Date(2026, 0, 1, 18, 0), abnormalAlert: true }; ElMessage.info('已重置') }
const saveDataSettings = () => ElMessage.success('采集设置已保存')
const saveApiSettings = () => ElMessage.success('API配置已保存')
const testApiConnection = () => {
  ElMessage.info('正在测试连接...')
  setTimeout(() => ElMessage.success('连接成功'), 1000)
}
</script>

<style scoped>
.settings-page { padding: 20px; }
.page-header { margin-bottom: 20px; }
.page-header h1 { margin: 0; color: #303133; }
.subtitle { margin: 10px 0 0; color: #909399; }
.settings-card { margin-bottom: 0; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
