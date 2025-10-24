# ---------- 构建阶段 ----------
FROM golang:1.24 AS builder

WORKDIR /app

# 复制依赖定义并下载（提高缓存利用）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# ---------- 运行阶段 ----------
FROM debian:bookworm-slim

LABEL stage="runtime" \
      base_image="debian:bookworm-slim" \
      app_version="1.0.0" \
      description="Golang server runtime image with log, db, cfg, fdb directories"

WORKDIR /app

# 拷贝可执行文件
COPY --from=builder /app/server .

# 创建运行所需目录
RUN mkdir -p /app/log /app/db /app/cfg /app/fdb

# 拷贝数据库文件与配置文件
COPY db.db /app/db/db.db
COPY server.json /app/cfg/server.json
COPY  fdb/** /app/**

# 设置权限（可选）
RUN chmod 755 /app && chmod -R 755 /app/*

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]