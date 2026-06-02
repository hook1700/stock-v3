# ---------- Builder ----------
FROM node:20-alpine AS builder

# 1️⃣ 安装必要工具（解决 Alpine + vite 权限 & native 依赖问题）
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    git \
    bash

WORKDIR /app

# 2️⃣ 仅复制依赖声明（缓存层）
COPY package*.json ./

# 3️⃣ 安装依赖（禁止 audit / progress，CI 更快）
RUN npm config set fetch-retries 5 \
 && npm install --omit=optional --silent --no-audit --no-fund

# 4️⃣ 强制重建 vite bin（彻底根治 126）
RUN npm rebuild vite --force

# 5️⃣ 复制源码（最后一步，变动最频繁）
COPY . .

# 6️⃣ 构建（明确失败即退出）
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL:-http://localhost:8080/api}

RUN npm run build

# ---------- Runtime ----------
FROM nginx:alpine

# 1️⃣ 拷贝构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 2️⃣ 使用自定义 nginx 配置（如有）
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# 3️⃣ 暴露端口（仅文档作用）
EXPOSE 3000

# 4️⃣ 非 root 运行（CI / K8s 推荐）
USER nginx

CMD ["nginx", "-g", "daemon off;"]