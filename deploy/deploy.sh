#!/bin/bash

# 股票策略分析系统部署脚本
# 适用于腾讯云部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "命令 $1 未安装，请先安装"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help           显示帮助信息"
    echo "  -d, --deploy         部署整个系统"
    echo "  -u, --update         更新系统"
    echo "  -s, --stop           停止系统"
    echo "  -r, --restart        重启系统"
    echo "  -l, --logs           查看日志"
    echo "  -b, --backup         备份数据"
    echo "  -c, --clean          清理系统"
}

# 检查环境
check_environment() {
    log_info "检查部署环境..."

    # 检查Docker
    check_command docker

    # 检查Docker Compose
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    elif docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    else
        log_error "未安装 docker-compose，请先安装"
        exit 1
    fi

    # 检查是否在 deploy 目录或项目根目录
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

    if [ -f "$SCRIPT_DIR/docker-compose.yml" ]; then
        COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
        log_info "使用部署目录的 docker-compose.yml"
    elif [ -f "$PROJECT_ROOT/docker-compose.yml" ]; then
        COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"
        log_info "使用项目根的 docker-compose.yml"
    else
        log_error "找不到 docker-compose.yml"
        exit 1
    fi

    log_info "环境检查通过"
}

# 创建配置文件
create_configs() {
    log_info "创建配置文件..."

    # 切换到项目根目录（deploy 的上一级）
    cd "$(dirname "$COMPOSE_FILE")/.."

    # 创建后端配置文件
    if [ ! -f "backend/config.yaml" ]; then
        cat > backend/config.yaml << EOF
server:
  port: "8080"
  mode: "production"
  read_timeout: 30
  write_timeout: 30

database:
  host: "postgres"
  port: "5432"
  user: "postgres"
  password: "${DB_PASSWORD:-password}"
  dbname: "stock_strategy"
  sslmode: "disable"

redis:
  host: "redis"
  port: "6379"
  password: ""
  db: 0

strategy:
  daily_update_time: "17:45"
  strategy_run_time: "18:00"
  weekend_skip: true
EOF
    fi

    # 创建前端环境变量
    if [ ! -f "frontend/.env.production" ]; then
        cat > frontend/.env.production << EOF
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_TITLE=股票策略分析系统
EOF
    fi

    log_info "配置文件创建完成"
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."

    # 切换到 compose 文件所在目录
    cd "$(dirname "$COMPOSE_FILE")"

    # 使用 docker-compose 构建（复用 compose 文件中的 build 配置）
    $COMPOSE_CMD -f "$COMPOSE_FILE" build

    log_info "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."

    # 切换到 compose 文件所在目录
    cd "$(dirname "$COMPOSE_FILE")"

    # 启动数据库和缓存
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d postgres redis

    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 30

    # 初始化数据库
    log_info "初始化数据库..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" exec postgres psql -U postgres -d stock_strategy -f /docker-entrypoint-initdb.d/init.sql

    # 启动其他服务
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d

    log_info "服务启动完成"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" down
    log_info "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" restart
    log_info "服务已重启"
}

# 查看日志
show_logs() {
    log_info "显示服务日志..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f
}

# 备份数据
backup_data() {
    log_info "备份数据..."

    # 创建备份目录
    BACKUP_DIR="backup/$(date +%Y%m%d_%H%M%S)"
    mkdir -p $BACKUP_DIR

    # 切换到 compose 文件所在目录
    cd "$(dirname "$COMPOSE_FILE")"

    # 备份数据库
    log_info "备份数据库..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" exec -T postgres pg_dump -U postgres stock_strategy > $BACKUP_DIR/database.sql

    # 备份Redis数据
    log_info "备份Redis数据..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" exec -T redis redis-cli save
    docker cp stock-strategy-redis:/data/dump.rdb $BACKUP_DIR/redis.rdb

    # 备份配置文件
    cp -r ../backend/config.yaml $BACKUP_DIR/ 2>/dev/null || true
    cp -r ../frontend/.env.production $BACKUP_DIR/ 2>/dev/null || true

    # 压缩备份文件
    tar -czf $BACKUP_DIR.tar.gz $BACKUP_DIR
    rm -rf $BACKUP_DIR

    log_info "数据备份完成: $BACKUP_DIR.tar.gz"
}

# 清理系统
clean_system() {
    log_warn "此操作将删除所有容器、镜像和数据，确定要继续吗？(y/N)"
    read -r response

    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "清理系统..."

        # 切换到 compose 文件所在目录
        cd "$(dirname "$COMPOSE_FILE")"

        # 停止并删除容器
        $COMPOSE_CMD -f "$COMPOSE_FILE" down -v

        # 删除镜像
        docker rmi stock-strategy-backend:latest stock-strategy-frontend:latest stock-data-collector:latest || true

        # 清理Docker缓存
        docker system prune -f

        log_info "系统清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."

    cd "$(dirname "$COMPOSE_FILE")"

    # 检查服务状态
    if $COMPOSE_CMD -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        log_info "所有服务运行正常"
    else
        log_error "部分服务运行异常"
        $COMPOSE_CMD -f "$COMPOSE_FILE" ps
        exit 1
    fi

    # 检查API接口
    if curl -s http://localhost:8080/health > /dev/null; then
        log_info "后端API服务正常"
    else
        log_error "后端API服务异常"
        exit 1
    fi

    log_info "健康检查通过"
}

# 主函数
main() {
    case "$1" in
        -h|--help)
            show_help
            ;;
        -d|--deploy)
            check_environment
            create_configs
            build_images
            start_services
            health_check
            log_info "部署完成！"
            log_info "前端地址: http://localhost:3000"
            log_info "后端地址: http://localhost:8080"
            ;;
        -u|--update)
            check_environment
            build_images
            restart_services
            health_check
            log_info "系统更新完成"
            ;;
        -s|--stop)
            stop_services
            ;;
        -r|--restart)
            restart_services
            health_check
            ;;
        -l|--logs)
            show_logs
            ;;
        -b|--backup)
            backup_data
            ;;
        -c|--clean)
            clean_system
            ;;
        *)
            log_error "无效选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"