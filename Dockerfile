# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置 Go 环境变量
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build

# 先复制 go.mod 和 go.sum 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制其余代码
COPY . .

# 编译应用
RUN go build -o main .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 设置时区（如果 DSN 或者应用依赖特定时区）
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 从构建阶段复制二进制文件
COPY --from=builder /build/main .

EXPOSE 8080

CMD ["./main"]
