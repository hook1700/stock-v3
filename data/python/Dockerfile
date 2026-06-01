# 数据采集服务 Dockerfile
FROM python:3.10-slim

# 使用国内 Debian 镜像源（加速 apt 安装）
RUN sed -i 's|http://deb.debian.org|https://mirrors.tencentyun.com|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's|http://deb.debian.org|https://mirrors.aliyun.com|g' /etc/apt/sources.list 2>/dev/null || true

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .
# 使用清华 PyPI 镜像加速 pip 安装
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# 复制源码
COPY . .

# 设置时区
ENV TZ=Asia/Shanghai

CMD ["python", "main.py"]
