FROM node:20-alpine AS builder

# 1️⃣ 国内 apk 源（解决 gcc 卡死）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache python3 make g++ git bash

WORKDIR /app

# 2️⃣ 复制依赖
COPY package*.json ./

# 3️⃣ 国内 npm 镜像
RUN npm config set registry https://registry.npmmirror.com \
    && npm install --omit=optional --no-audit --no-fund

# ✅ 4️⃣ 关键：强制赋予 vite 执行权限（Alpine 必加）
RUN chmod +x node_modules/.bin/vite

# 5️⃣ 复制源码
COPY . .

# 6️⃣ 构建
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL:-http://localhost:8080/api}

RUN npx vite build