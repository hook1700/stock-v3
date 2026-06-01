# 开发指南

## 项目概述

股票策略分析系统是一个基于微服务架构的企业级应用，包含数据采集、策略计算、Web界面等多个模块。

## 开发环境搭建

### 1. 环境要求

- **操作系统**: Linux/macOS/Windows (推荐Linux)
- **Python**: 3.8+ (数据采集模块)
- **Node.js**: 16+ (前端界面)
- **Go**: 1.19+ (后端服务)
- **Docker**: 20.10+ (可选，用于容器化部署)
- **PostgreSQL**: 15+ (数据库)
- **Redis**: 7+ (缓存)

### 2. 快速开始

```bash
# 克隆项目
git clone <repository-url>
cd stock-v3

# 一键启动开发环境
chmod +x start.sh
./start.sh install  # 安装依赖
./start.sh start    # 启动所有服务
```

### 3. 手动启动（可选）

```bash
# 1. 启动数据库服务
cd deploy
docker-compose up -d postgres redis

# 2. 启动数据采集服务
cd data/python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python main.py --schedule

# 3. 启动后端服务
cd backend
go mod tidy
go run main.go

# 4. 启动前端服务
cd frontend
npm install
npm run dev
```

## 项目架构

### 1. 数据流架构

```
数据源 (BaoStock/AKShare)
    ↓
数据采集器 (Python)
    ↓
数据库 (PostgreSQL)
    ↓
策略引擎 (Golang)
    ↓
API服务 (Gin)
    ↓
前端界面 (Vue.js)
```

### 2. 目录结构

```
stock-v3/
├── data/                 # 数据采集模块
│   ├── python/           # Python脚本
│   │   ├── main.py       # 主程序
│   │   ├── collector.py  # 数据采集器
│   │   ├── writer.py     # 数据库写入器
│   │   └── requirements.txt
│   └── config/           # 配置管理
├── backend/              # Golang后端
│   ├── api/              # API接口
│   ├── strategy/         # 策略引擎
│   ├── model/            # 数据模型
│   ├── task/             # 定时任务
│   ├── config/           # 配置管理
│   ├── go.mod           # Go模块
│   └── main.go          # 主程序
├── frontend/             # Vue前端
│   ├── src/
│   │   ├── views/        # 页面组件
│   │   ├── router/      # 路由配置
│   │   ├── components/   # 公共组件
│   │   └── main.js       # 应用入口
│   ├── package.json
│   └── vite.config.js
├── database/             # 数据库脚本
├── deploy/               # 部署配置
└── docs/                 # 项目文档
```

## 开发规范

### 1. 代码规范

#### Python代码规范
- 遵循PEP8规范
- 使用类型注解
- 函数和类要有文档字符串
- 使用Black进行代码格式化

```python
def get_stock_data(stock_code: str, start_date: str, end_date: str) -> pd.DataFrame:
    """
    获取股票历史数据
    
    Args:
        stock_code: 股票代码
        start_date: 开始日期 (YYYY-MM-DD)
        end_date: 结束日期 (YYYY-MM-DD)
    
    Returns:
        DataFrame: 股票历史数据
    """
    # 实现代码
    pass
```

#### Go代码规范
- 遵循Effective Go规范
- 使用gofmt进行代码格式化
- 错误处理要明确
- 包名使用小写字母

```go
// GetStockList 获取股票列表
func GetStockList(c *gin.Context) {
    stocks, err := stockRepo.GetAllStocks()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "获取股票列表失败"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"stocks": stocks})
}
```

#### JavaScript/Vue代码规范
- 使用ESLint + Prettier
- 遵循Vue官方风格指南
- 使用Composition API
- 组件名使用PascalCase

```vue
<template>
  <div class="stock-list">
    <StockTable :stocks="stocks" @select="handleSelect" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import StockTable from '@/components/StockTable.vue'

const stocks = ref([])

const handleSelect = (stock) => {
  // 处理选择逻辑
}

onMounted(async () => {
  stocks.value = await fetchStocks()
})
</script>
```

### 2. 数据库设计规范

#### 表命名规范
- 使用小写字母和下划线
- 表名使用复数形式
- 关联表使用下划线连接

```sql
-- 股票基本信息表
CREATE TABLE stocks (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    industry VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 股票日线数据表
CREATE TABLE stock_daily_data (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,
    trade_date DATE NOT NULL,
    close_price DECIMAL(10,3),
    volume BIGINT,
    UNIQUE(stock_code, trade_date)
);
```

### 3. API设计规范

#### RESTful API设计
- 使用HTTP状态码
- 资源使用复数形式
- 版本控制使用URL路径

```
GET    /api/v1/stocks          # 获取股票列表
GET    /api/v1/stocks/{code}    # 获取股票详情
POST   /api/v1/strategies/run  # 执行策略
GET    /api/v1/fund-flow/sectors # 获取板块资金流
```

#### 响应格式
```json
{
  "success": true,
  "data": {...},
  "message": "操作成功",
  "timestamp": "2026-06-01T00:00:00Z"
}
```

## 模块开发指南

### 1. 数据采集模块开发

#### 添加新的数据源

1. 在`data/python/collector.py`中添加新的数据源类
2. 实现数据获取方法
3. 在`data/python/main.py`中集成新数据源

```python
class NewDataSource:
    def get_stock_data(self, stock_code, start_date, end_date):
        # 实现数据获取逻辑
        pass
    
    def get_industry_data(self):
        # 实现行业数据获取
        pass
```

#### 添加新的技术指标

1. 在`data/python/indicators.py`中添加指标计算函数
2. 在数据采集流程中调用指标计算

```python
def calculate_rsi(prices, period=14):
    """计算RSI指标"""
    # RSI计算逻辑
    pass
```

### 2. 策略模块开发

#### 添加新的策略

1. 在`backend/strategy/`目录下创建策略文件
2. 实现策略接口
3. 在策略管理器中注册新策略

```go
// 创建新策略文件
// backend/strategy/new_strategy.go

type NewStrategy struct {
    name string
    parameters map[string]interface{}
}

func (s *NewStrategy) Execute(stock *model.Stock, data []model.StockDailyData) *model.StrategyResult {
    // 策略逻辑实现
    return &model.StrategyResult{
        StockCode: stock.Code,
        Score:     0.8,
        BuyPrice:  100.0,
        LogicDescription: "新策略逻辑说明",
    }
}
```

#### 策略参数配置

```go
// 在策略管理器中注册新策略
func NewStrategyManager() *StrategyManager {
    return &StrategyManager{
        strategies: map[string]Strategy{
            "new_strategy": &NewStrategy{
                name: "新策略",
                parameters: map[string]interface{}{
                    "param1": 10,
                    "param2": 0.5,
                },
            },
        },
    }
}
```

### 3. 前端模块开发

#### 添加新的页面

1. 在`frontend/src/views/`目录下创建新页面组件
2. 在路由配置中添加新路由
3. 在侧边栏菜单中添加导航项

```vue
<!-- frontend/src/views/NewPage.vue -->
<template>
  <div class="new-page">
    <h1>新页面</h1>
    <!-- 页面内容 -->
  </div>
</template>

<script setup>
// 页面逻辑
</script>
```

#### 添加新的图表组件

1. 在`frontend/src/components/charts/`目录下创建图表组件
2. 使用ECharts进行数据可视化

```vue
<template>
  <div ref="chartRef" style="width: 100%; height: 400px;"></div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'

const chartRef = ref()

onMounted(() => {
  const chart = echarts.init(chartRef.value)
  chart.setOption({
    // 图表配置
  })
})
</script>
```

## 测试指南

### 1. 单元测试

#### Python单元测试
```python
# tests/test_collector.py
import unittest
from data.python.collector import StockDataCollector

class TestStockCollector(unittest.TestCase):
    def test_get_stock_data(self):
        collector = StockDataCollector()
        data = collector.get_stock_data("000001", "2024-01-01", "2024-01-31")
        self.assertIsNotNone(data)
        self.assertGreater(len(data), 0)
```

#### Go单元测试
```go
// backend/strategy/strategy_test.go
package strategy

import (
    "testing"
    "stock-strategy-backend/model"
)

func TestShortTermStrategy(t *testing.T) {
    strategy := NewShortTermStrategy("short_term_1")
    stock := &model.Stock{Code: "000001", Name: "测试股票"}
    
    result := strategy.Execute(stock, []model.StockDailyData{})
    if result == nil {
        t.Error("策略执行失败")
    }
}
```

### 2. 集成测试

```bash
# 运行所有测试
cd backend && go test ./...
cd ../data/python && python -m pytest
```

## 部署指南

### 1. 开发环境部署

```bash
# 使用Docker Compose部署
cd deploy
docker-compose up -d

# 验证服务状态
docker-compose ps
curl http://localhost:8080/health
```

### 2. 生产环境部署

```bash
# 使用部署脚本
cd deploy
./deploy.sh --deploy

# 配置Nginx反向代理
sudo cp nginx/stock-strategy.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/stock-strategy.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 性能优化

### 1. 数据库优化

```sql
-- 添加索引
CREATE INDEX idx_stock_daily_data_code_date ON stock_daily_data(stock_code, trade_date);
CREATE INDEX idx_strategy_results_date_strategy ON strategy_results(trade_date, strategy_id);

-- 分区表
CREATE TABLE stock_daily_data_2024 PARTITION OF stock_daily_data 
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
```

### 2. 缓存优化

```go
// 使用Redis缓存热点数据
func GetStockWithCache(stockCode string) (*model.Stock, error) {
    cacheKey := fmt.Sprintf("stock:%s", stockCode)
    
    // 从缓存获取
    if cached, err := redis.Get(cacheKey); err == nil {
        return cached, nil
    }
    
    // 从数据库获取
    stock, err := stockRepo.GetStockByCode(stockCode)
    if err != nil {
        return nil, err
    }
    
    // 设置缓存
    redis.Set(cacheKey, stock, 30*time.Minute)
    
    return stock, nil
}
```

### 3. 前端性能优化

```javascript
// 使用虚拟滚动优化大数据列表
import { ElTableV2 } from 'element-plus'

const columns = [
  { key: 'code', title: '代码', width: 100 },
  { key: 'name', title: '名称', width: 120 },
  // ...更多列
]

<ElTableV2 :columns="columns" :data="stocks" :height="400" />
```

## 故障排除

### 常见问题

#### 1. 数据库连接失败
```bash
# 检查数据库服务状态
sudo systemctl status postgresql

# 检查连接配置
cat backend/config.yaml | grep database
```

#### 2. 数据采集失败
```bash
# 检查数据源API状态
curl https://api.baostock.com/health

# 查看采集日志
tail -f data/python/collector.log
```

#### 3. 前端构建失败
```bash
# 清理缓存
rm -rf frontend/node_modules frontend/package-lock.json
npm install
npm run build
```

## 贡献指南

### 1. 代码提交规范

- 使用约定式提交格式
- 提交信息使用英文
- 一个提交对应一个功能或修复

```bash
git commit -m "feat: 添加新的短线策略"
git commit -m "fix: 修复数据采集内存泄漏"
git commit -m "docs: 更新API文档"
```

### 2. Pull Request流程

1. Fork项目仓库
2. 创建功能分支
3. 提交代码变更
4. 创建Pull Request
5. 等待代码审查
6. 合并到主分支

## 联系方式

- 项目维护者: v_zwhozhang
- 问题反馈: GitHub Issues
- 文档更新: 提交Pull Request

---

**最后更新**: 2026年6月1日  
**版本**: v1.0.0