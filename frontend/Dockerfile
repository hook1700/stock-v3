# ---------- Builder ----------
FROM node:20-alpine AS builder

# ✅ 1️⃣ 换国内 Alpine 源（关键）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache \
        python3 \
        make \
        g++ \
        git \
        bash

WORKDIR /app

# 2️⃣ 仅复制依赖声明（缓存层）
COPY package*.json ./

# 3️⃣ 安装依赖（国内 npm 镜像）
RUN npm config set registry https://registry.npmmirror.com \
    && npm config set fetch-retries 5 \
    && npm install --omit=optional --silent --no-audit --no-fund

# 4️⃣ 强制重建 vite bin
RUN npm rebuild vite --force

# 5️⃣ 复制源码
COPY . .

# 6️⃣ 构建
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL:-http://localhost:8080/api}
RUN npm run build

# ---------- Runtime ----------
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 3000

USER nginx

CMD ["nginx", "-g", "daemon off;"]