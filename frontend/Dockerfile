# 前端 Node.js 服务 Dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY package*.json ./
RUN npm install

# 确保 vite 等有执行权限（修复 Windows 下构建权限问题）
RUN chmod +x node_modules/.bin/*

# 复制源码
COPY . .

# 构建生产版本
ARG VITE_API_BASE_URL=http://localhost:8080/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN npm run build

# 运行阶段 - 使用 nginx 托管静态文件
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
