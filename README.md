# 股票策略分析系统 v3

基于股票数据的自动化策略分析系统，支持多种投资策略的自动化执行和结果记录。

## 🎯 项目概述

这是一个企业级的股票策略分析系统，旨在通过自动化策略执行帮助用户筛选符合特定投资逻辑的股票。系统支持多种投资策略（短线、中线、长线），并提供完整的股票数据获取、策略执行、结果记录功能。

## 🚀 技术栈

- **数据获取**: Python (BaoStock + AKShare)
- **后端服务**: Golang + Gin + PostgreSQL
- **前端界面**: Vue.js + Element Plus + ECharts
- **缓存服务**: Redis
- **部署环境**: Docker + Docker Compose
- **云平台**: 腾讯云（支持）

## 📋 核心功能

### ✅ 已完成功能

1. **项目架构搭建**
   - 完整的微服务架构设计
   - 前后端分离开发模式
   - Docker容器化部署方案

2. **数据采集模块**
   - 股票基本信息采集
   - 日线数据获取
   - 行业分类数据
   - 资金流数据统计
   - 定时任务调度

3. **后端服务**
   - RESTful API设计
   - 数据库模型定义
   - 策略引擎框架
   - 定时任务管理
   - 用户操作记录

4. **前端界面**
   - Vue.js单页应用
   - Element Plus UI组件
   - 响应式布局设计
   - 仪表盘和数据可视化

5. **部署配置**
   - Docker容器化配置
   - 腾讯云部署指南
   - 自动化部署脚本
   - 运维管理工具

### 📝 待完成功能

1. **数据采集模块完善**
   - 历史数据批量采集
   - 实时数据流处理
   - 数据质量校验

2. **策略算法实现**
   - 短线策略完整实现
   - 中线策略完整实现
   - 长线策略完整实现
   - 策略回测功能

3. **前端功能完善**
   - 股票详情页面
   - 策略分析页面
   - 资金流可视化
   - 用户设置页面

4. **系统优化**
   - 性能优化和缓存策略
   - 错误处理和日志系统
   - 安全加固和权限管理

## 📁 项目结构

```
stock-v3/
├── data/                 # 数据采集模块
│   ├── python/           # Python数据获取脚本
│   ├── config/           # 数据源配置
│   └── requirements.txt  # Python依赖
├── backend/              # Golang后端服务
│   ├── api/              # API接口处理器
│   ├── config/           # 配置管理
│   ├── model/            # 数据模型
│   ├── strategy/         # 策略引擎
│   ├── task/             # 定时任务
│   ├── go.mod            # Go模块配置
│   ├── main.go           # 主程序入口
│   └── config.yaml       # 应用配置
├── frontend/             # Vue前端界面
│   ├── src/
│   │   ├── views/        # 页面组件
│   │   ├── router/       # 路由配置
│   │   └── main.js       # 应用入口
│   ├── package.json      # 前端依赖
│   ├── vite.config.js    # Vite配置
│   └── index.html        # HTML模板
├── database/             # 数据库脚本
│   └── schema.sql        # 数据库表结构
├── deploy/               # 部署配置
│   ├── docker-compose.yml # 容器编排
│   ├── deploy.sh         # 部署脚本
│   └── nginx/            # Nginx配置
├── docs/                 # 项目文档
│   ├── CLAUDE.md         # 项目详细说明
│   └── DEPLOYMENT.md     # 部署指南
└── README.md             # 项目说明
```

## 🛠️ 快速开始

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 4GB以上内存
- 50GB以上磁盘空间

### 一键部署

```bash
# 克隆项目
git clone <repository-url>
cd stock-v3

# 进入部署目录
cd deploy

# 赋予执行权限
chmod +x deploy.sh

# 执行部署
./deploy.sh --deploy
```

### 访问系统

部署完成后，访问以下地址：

- **前端界面**: http://localhost:3000
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

## 📊 投资策略体系

### 短线策略（持仓1-10天）

1. **均线回踩低吸** - 趋势热点股、板块处上升期
2. **突破缩量回踩** - 横盘震荡后选择方向的票
3. **强势股10日线反抽** - 强于大盘的板块龙头，短期回调

### 中线策略（持仓1-3个月）

1. **行业成长 + 均线多头** - 行业景气股，偏中长线主做
2. **困境反转 / 业绩拐点** - 业绩由差转好、订单/政策催化
3. **高股息红利慢牛** - 震荡市/偏弱市，求稳健

### 长线策略（持仓6月-数年）

1. **优质白马龙头** - 连续3年ROE≥15%，净利↑，现金流好
2. **红利再投收息策略** - 不想频繁操作，重视现金流
3. **真成长PEG低吸** - 能接受波动、愿研究行业和增速

## 🔧 开发指南

### 本地开发环境

```bash
# 启动后端服务
cd backend
go run main.go

# 启动前端开发服务器
cd frontend
npm install
npm run dev

# 启动数据采集服务
cd data/python
pip install -r requirements.txt
python main.py
```

### API接口文档

系统提供完整的RESTful API接口：

- `GET /api/stocks/list` - 获取股票列表
- `GET /api/stocks/{code}` - 获取股票详情
- `GET /api/strategies` - 获取策略列表
- `POST /api/strategies/{id}/run` - 执行策略
- `GET /api/fund-flow/sectors` - 获取板块资金流

### 数据库设计

系统使用PostgreSQL数据库，包含以下核心表：

- `stocks` - 股票基本信息
- `stock_daily_data` - 股票日线数据
- `strategies` - 策略配置
- `strategy_results` - 策略执行结果
- `sector_fund_flow` - 板块资金流数据

## 📈 开发进度

### ✅ 已完成

- [x] 项目架构设计和文档整理
- [x] 基础目录结构搭建
- [x] 数据采集模块框架
- [x] 后端服务基础框架
- [x] 前端Vue应用基础
- [x] Docker容器化配置
- [x] 自动化部署脚本
- [x] 腾讯云部署指南

### 🔄 进行中

- [ ] 数据采集模块完整实现
- [ ] 策略算法详细实现
- [ ] 前端页面功能完善
- [ ] 系统集成测试

### 📋 待开始

- [ ] 性能优化和缓存策略
- [ ] 安全加固和权限管理
- [ ] 监控告警系统集成
- [ ] 生产环境部署验证

## 🤝 贡献指南

欢迎贡献代码！请遵循以下流程：

1. Fork本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## ⚠️ 免责声明

本系统仅供学习和研究使用，不构成任何投资建议。投资有风险，决策需谨慎。

## 📞 联系方式

- 项目维护者: 
- 项目文档: [CLAUDE.md](docs/CLAUDE.md)
- 部署指南: [DEPLOYMENT.md](docs/DEPLOYMENT.md)

---

**最后更新**: 2026年6月1日
**版本**: v1.0.0-alpha