# 股票策略分析系统部署指南

## 系统架构

本系统采用微服务架构，包含以下组件：

- **前端**: Vue.js + Element Plus (端口: 3000)
- **后端**: Golang + Gin (端口: 8080)
- **数据库**: PostgreSQL (端口: 5432)
- **缓存**: Redis (端口: 6379)
- **数据采集**: Python + BaoStock/AKShare
- **反向代理**: Nginx (端口: 80/443)

## 环境要求

### 硬件要求
- CPU: 4核以上
- 内存: 8GB以上
- 硬盘: 50GB以上可用空间
- 网络: 稳定的互联网连接

### 软件要求
- Docker 20.10+
- Docker Compose 2.0+
- Git
- 腾讯云账号（可选，用于生产环境部署）

## 快速开始

### 1. 克隆项目
```bash
git clone <repository-url>
cd stock-v3
```

### 2. 环境准备
确保已安装Docker和Docker Compose：
```bash
# 检查Docker版本
docker --version

# 检查Docker Compose版本
docker-compose --version
```

### 3. 一键部署
```bash
# 进入部署目录
cd deploy

# 赋予执行权限
chmod +x deploy.sh

# 执行部署
./deploy.sh --deploy
```

### 4. 访问系统
部署完成后，可以通过以下地址访问：
- 前端界面: http://localhost:3000
- 后端API: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 腾讯云部署

### 1. 创建云服务器
- 选择CentOS 7.6或Ubuntu 20.04 LTS
- 推荐配置: 4核8GB内存
- 系统盘: 100GB SSD
- 安全组开放端口: 22, 80, 443, 3000, 8080

### 2. 服务器初始化
```bash
# 更新系统
sudo yum update -y

# 安装Docker
curl -fsSL https://get.docker.com | sh

# 启动Docker
sudo systemctl start docker
sudo systemctl enable docker

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 创建项目目录
mkdir -p /opt/stock-strategy
cd /opt/stock-strategy
```

### 3. 上传项目文件
```bash
# 上传项目文件到服务器
scp -r ./stock-v3 root@your-server-ip:/opt/stock-strategy/
```

### 4. 生产环境配置
```bash
cd /opt/stock-strategy/stock-v3/deploy

# 修改生产环境配置
vim docker-compose.prod.yml

# 设置数据库密码
export DB_PASSWORD=your_secure_password

# 部署
./deploy.sh --deploy
```

## 配置说明

### 数据库配置
```yaml
# backend/config.yaml
database:
  host: "postgres"
  port: "5432"
  user: "postgres"
  password: "${DB_PASSWORD}"
  dbname: "stock_strategy"
  sslmode: "disable"
```

### Redis配置
```yaml
redis:
  host: "redis"
  port: "6379"
  password: ""
  db: 0
```

### 策略执行时间配置
```yaml
strategy:
  daily_update_time: "9:25"    # 每日数据更新时间
  strategy_run_time: "18:00"    # 策略执行时间
  weekend_skip: true            # 周末跳过执行
```

## 数据初始化

### 1. 数据库初始化
```bash
# 执行数据库初始化脚本
docker-compose exec postgres psql -U postgres -d stock_strategy -f /docker-entrypoint-initdb.d/init.sql
```

### 2. 历史数据采集
```bash
# 进入数据采集容器
docker exec -it stock-data-collector bash

# 采集历史数据
python main.py --init 365

# 启动定时任务
python main.py --schedule
```

## 运维管理

### 常用命令
```bash
# 查看服务状态
./deploy.sh --status

# 重启服务
./deploy.sh --restart

# 查看日志
./deploy.sh --logs

# 备份数据
./deploy.sh --backup

# 更新系统
./deploy.sh --update

# 停止服务
./deploy.sh --stop

# 清理系统
./deploy.sh --clean
```

### 监控和日志
- 查看容器日志: `docker-compose logs -f`
- 查看特定服务日志: `docker-compose logs -f backend`
- 监控资源使用: `docker stats`
- 查看容器状态: `docker-compose ps`

### 数据备份
```bash
# 自动备份脚本
./deploy.sh --backup

# 手动备份数据库
docker-compose exec postgres pg_dump -U postgres stock_strategy > backup.sql

# 恢复数据库
docker-compose exec -T postgres psql -U postgres -d stock_strategy < backup.sql
```

## 故障排除

### 常见问题

#### 1. 端口冲突
```bash
# 检查端口占用
netstat -tulpn | grep :8080

# 修改端口配置
vim docker-compose.yml
```

#### 2. 数据库连接失败
```bash
# 检查数据库状态
docker-compose logs postgres

# 重启数据库服务
docker-compose restart postgres
```

#### 3. 内存不足
```bash
# 查看内存使用
docker stats

# 清理无用镜像
docker system prune
```

#### 4. 磁盘空间不足
```bash
# 查看磁盘使用
df -h

# 清理日志文件
sudo find /var/lib/docker/containers -name "*.log" -type f -delete
```

### 性能优化

#### 1. 数据库优化
```sql
-- 创建索引
CREATE INDEX idx_stock_daily_data_code_date ON stock_daily_data(stock_code, trade_date);
CREATE INDEX idx_strategy_results_date_strategy ON strategy_results(trade_date, strategy_id);

-- 定期清理旧数据
DELETE FROM stock_daily_data WHERE trade_date < NOW() - INTERVAL '2 years';
```

#### 2. Redis优化
```bash
# 配置Redis内存限制
maxmemory 1gb
maxmemory-policy allkeys-lru
```

#### 3. 应用优化
```yaml
# 调整容器资源限制
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2'
        reservations:
          memory: 1G
          cpus: '1'
```

## 安全配置

### 1. 防火墙配置
```bash
# 只开放必要端口
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

### 2. SSL证书配置
```bash
# 使用Let's Encrypt获取免费证书
certbot --nginx -d your-domain.com

# 配置Nginx SSL
vim nginx/conf.d/ssl.conf
```

### 3. 数据库安全
```bash
# 修改默认密码
ALTER USER postgres WITH PASSWORD 'new_secure_password';

# 限制连接IP
vim postgresql.conf
```

## 扩展部署

### 1. 集群部署
对于高可用需求，可以考虑：
- PostgreSQL主从复制
- Redis集群
- 负载均衡
- 多节点部署

### 2. 监控告警
集成监控系统：
- Prometheus + Grafana
- ELK日志系统
- 腾讯云监控

### 3. CI/CD流水线
配置自动化部署：
- GitHub Actions
- Jenkins
- GitLab CI

## 联系方式

如有部署问题，请联系：
- 项目维护者: v_zwhozhang
- 邮箱: your-email@example.com
- 文档更新: 请参考项目README.md

## 版本历史

- v1.0.0 (2026-06-01): 初始版本发布
- 支持基础股票数据采集和策略分析
- 提供Web界面和API接口
- 支持腾讯云部署

---

**注意**: 本系统仅供学习和研究使用，投资有风险，决策需谨慎。