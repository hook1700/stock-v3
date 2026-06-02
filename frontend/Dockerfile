# ---------- Builder ----------
FROM node:20-alpine AS builder

# ✅ 1️⃣ 国内 apk 源（解决 gcc 卡死）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache \
        python3 \
        make \
        g++ \
        git \
        bash

WORKDIR /app

# 2️⃣ 复制依赖声明
COPY package*.json ./

# 3️⃣ 使用国内 npm 镜像
RUN npm config set registry https://registry.npmmirror.com \
    && npm config set fetch-retries 5 \
    && npm install --omit=optional --silent --no-audit --no-fund

# ✅ 4️⃣ 不 rebuild vite（Alpine 必炸点）
# ❌ RUN npm rebuild vite --force

# 5️⃣ 复制源码
COPY . .

# 6️⃣ 构建（用 npx 保底）
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL:-http://localhost:8080/api}

RUN npx vite build