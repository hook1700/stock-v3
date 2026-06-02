FROM node:20-alpine AS builder

# 1️⃣ 国内 apk 源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache python3 make g++ git bash

WORKDIR /app

# 2️⃣ 依赖
COPY package*.json ./
RUN npm config set registry https://registry.npmmirror.com \
    && npm install --omit=optional --no-audit --no-fund

# 3️⃣ 源码（必须在 chmod 之前）
COPY . .

# ✅ 4️⃣ 关键：在 COPY 之后再给权限
RUN chmod +x node_modules/.bin/vite

# 5️⃣ 构建
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL:-http://localhost:8080/api}

RUN npx vite build

# 6️⃣ 运行阶段 - 使用 nginx 托管静态文件
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]