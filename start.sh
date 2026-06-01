#!/bin/bash

# 股票策略分析系统启动脚本
# 快速启动开发环境

set -e

echo "=========================================="
echo "   股票策略分析系统 v3 - 启动脚本"
echo "=========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
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

# 检查环境
check_environment() {
    log_info "检查系统环境..."

    # 检查Python
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
        log_info "Python版本: $PYTHON_VERSION"
    else
        log_error "Python3未安装"
        exit 1
    fi

    # 检查Node.js
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        log_info "Node.js版本: $NODE_VERSION"
    else
        log_error "Node.js未安装"
        exit 1
    fi

    # 检查Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        log_info "Go版本: $GO_VERSION"
    else
        log_error "Go未安装"
        exit 1
    fi

    # 检查Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | sed 's/,//')
        log_info "Docker版本: $DOCKER_VERSION"
    else
        log_warn "Docker未安装，将使用本地环境"
    fi

    # 检查Docker Compose
    if command -v docker-compose &> /dev/null; then
        log_info "Docker Compose已安装"
    else
        log_warn "Docker Compose未安装，将使用本地环境"
    fi

    log_success "环境检查完成"
}

# 安装Python依赖
install_python_deps() {
    log_info "安装Python依赖..."

    cd data/python
    if [ ! -f "requirements.txt" ]; then
        log_error "requirements.txt文件不存在"
        exit 1
    fi

    if [ ! -d "venv" ]; then
        log_info "创建Python虚拟环境..."
        python3 -m venv venv
    fi

    source venv/bin/activate
    pip install --upgrade pip
    pip install -r requirements.txt

    log_success "Python依赖安装完成"
    cd ../..
}

# 安装前端依赖
install_frontend_deps() {
    log_info "安装前端依赖..."

    cd frontend
    if [ ! -f "package.json" ]; then
        log_error "package.json文件不存在"
        exit 1
    fi

    npm install

    log_success "前端依赖安装完成"
    cd ..
}

# 安装后端依赖
install_backend_deps() {
    log_info "安装后端依赖..."

    cd backend
    if [ ! -f "go.mod" ]; then
        log_error "go.mod文件不存在"
        exit 1
    fi

    go mod tidy

    log_success "后端依赖安装完成"
    cd ..
}

# 启动数据库服务
start_database() {
    log_info "启动数据库服务..."

    if command -v docker-compose &> /dev/null; then
        cd deploy
        docker-compose up -d postgres redis
        cd ..
        log_success "数据库服务启动完成"
    else
        log_warn "Docker Compose未安装，请手动启动数据库服务"
        log_info "可以使用: docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15"
        log_info "可以使用: docker run -d --name redis -p 6379:6379 redis:7-alpine"
    fi
}

# 初始化数据库
init_database() {
    log_info "初始化数据库..."

    # 等待数据库启动
    sleep 10

    # 执行数据库初始化脚本
    if command -v psql &> /dev/null; then
        cd database
        psql -h localhost -U postgres -f schema.sql
        cd ..
        log_success "数据库初始化完成"
    else
        log_warn "psql未安装，请手动执行数据库初始化脚本"
        log_info "脚本位置: database/schema.sql"
    fi
}

# 启动数据采集服务
start_data_collector() {
    log_info "启动数据采集服务..."

    cd data/python
    source venv/bin/activate

    # 启动数据采集服务（后台运行）
    nohup python main.py --schedule > collector.log 2>&1 &
    COLLECTOR_PID=$!
    echo $COLLECTOR_PID > collector.pid

    log_success "数据采集服务已启动 (PID: $COLLECTOR_PID)"
    cd ../..
}

# 启动后端服务
start_backend() {
    log_info "启动后端服务..."

    cd backend

    # 编译后端程序
    go build -o stock-strategy-backend

    # 启动后端服务（后台运行）
    nohup ./stock-strategy-backend > backend.log 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > backend.pid

    log_success "后端服务已启动 (PID: $BACKEND_PID)"
    cd ..
}

# 启动前端服务
start_frontend() {
    log_info "启动前端服务..."

    cd frontend

    # 启动前端开发服务器（后台运行）
    nohup npm run dev > frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > frontend.pid

    log_success "前端服务已启动 (PID: $FRONTEND_PID)"
    cd ..
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."

    sleep 5

    # 检查后端服务
    if curl -s http://localhost:8080/health > /dev/null; then
        log_success "后端服务运行正常"
    else
        log_error "后端服务启动失败"
    fi

    # 检查前端服务
    if curl -s http://localhost:3000 > /dev/null; then
        log_success "前端服务运行正常"
    else
        log_error "前端服务启动失败"
    fi

    # 检查数据采集服务
    if [ -f "data/python/collector.pid" ] && ps -p $(cat data/python/collector.pid) > /dev/null; then
        log_success "数据采集服务运行正常"
    else
        log_error "数据采集服务启动失败"
    fi
}

# 显示服务信息
show_service_info() {
    echo ""
    echo "=========================================="
    echo "           服务启动信息"
    echo "=========================================="
    echo ""
    echo "前端界面: http://localhost:3000"
    echo "后端API:  http://localhost:8080"
    echo "健康检查: http://localhost:8080/health"
    echo ""
    echo "数据库:   PostgreSQL (localhost:5432)"
    echo "缓存:     Redis (localhost:6379)"
    echo ""
    echo "日志文件:"
    echo "  - 后端服务: backend/backend.log"
    echo "  - 前端服务: frontend/frontend.log"
    echo "  - 数据采集: data/python/collector.log"
    echo ""
    echo "PID文件:"
    echo "  - 后端服务: backend/backend.pid"
    echo "  - 前端服务: frontend/frontend.pid"
    echo "  - 数据采集: data/python/collector.pid"
    echo ""
    echo "=========================================="
    echo ""
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."

    # 停止数据采集服务
    if [ -f "data/python/collector.pid" ]; then
        COLLECTOR_PID=$(cat data/python/collector.pid)
        if ps -p $COLLECTOR_PID > /dev/null; then
            kill $COLLECTOR_PID
            log_info "数据采集服务已停止 (PID: $COLLECTOR_PID)"
        fi
        rm data/python/collector.pid
    fi

    # 停止后端服务
    if [ -f "backend/backend.pid" ]; then
        BACKEND_PID=$(cat backend/backend.pid)
        if ps -p $BACKEND_PID > /dev/null; then
            kill $BACKEND_PID
            log_info "后端服务已停止 (PID: $BACKEND_PID)"
        fi
        rm backend/backend.pid
    fi

    # 停止前端服务
    if [ -f "frontend/frontend.pid" ]; then
        FRONTEND_PID=$(cat frontend/frontend.pid)
        if ps -p $FRONTEND_PID > /dev/null; then
            kill $FRONTEND_PID
            log_info "前端服务已停止 (PID: $FRONTEND_PID)"
        fi
        rm frontend/frontend.pid
    fi

    # 停止Docker服务
    if command -v docker-compose &> /dev/null; then
        cd deploy
        docker-compose down
        cd ..
        log_info "Docker服务已停止"
    fi

    log_success "所有服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    sleep 2
    start_services
}

# 查看服务状态
status_services() {
    log_info "服务状态检查..."

    echo ""
    echo "服务状态:"
    echo "---------"

    # 检查数据采集服务
    if [ -f "data/python/collector.pid" ] && ps -p $(cat data/python/collector.pid) > /dev/null; then
        echo "数据采集服务: 运行中"
    else
        echo "数据采集服务: 未运行"
    fi

    # 检查后端服务
    if [ -f "backend/backend.pid" ] && ps -p $(cat backend/backend.pid) > /dev/null; then
        echo "后端服务:     运行中"
    else
        echo "后端服务:     未运行"
    fi

    # 检查前端服务
    if [ -f "frontend/frontend.pid" ] && ps -p $(cat frontend/frontend.pid) > /dev/null; then
        echo "前端服务:     运行中"
    else
        echo "前端服务:     未运行"
    fi

    # 检查Docker服务
    if command -v docker-compose &> /dev/null && docker-compose ps | grep -q "Up"; then
        echo "Docker服务:   运行中"
    else
        echo "Docker服务:   未运行"
    fi

    echo ""
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  start     启动所有服务"
    echo "  stop      停止所有服务"
    echo "  restart   重启所有服务"
    echo "  status    查看服务状态"
    echo "  install   安装所有依赖"
    echo "  help      显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start    # 启动所有服务"
    echo "  $0 stop     # 停止所有服务"
    echo "  $0 status   # 查看服务状态"
}

# 主函数
main() {
    case "$1" in
        "start")
            check_environment
            install_python_deps
            install_frontend_deps
            install_backend_deps
            start_database
            init_database
            start_data_collector
            start_backend
            start_frontend
            check_services
            show_service_info
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "status")
            status_services
            ;;
        "install")
            check_environment
            install_python_deps
            install_frontend_deps
            install_backend_deps
            ;;
        "help"|"")
            show_help
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