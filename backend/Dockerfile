# 后端 Go 服务 Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 设置 Go 模块代理（加速国内访问）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# 复制依赖文件
COPY go.mod go.sum ./
# 先同步依赖关系，再下载
RUN go mod tidy && go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# 运行阶段
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .
COPY config.yaml ./

EXPOSE 8080

CMD ["./main"]
