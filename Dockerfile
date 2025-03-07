# 选择一个轻量级基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制二进制文件到容器
COPY auto-pprof /app/auto-pprof

# 赋予执行权限（如果需要）
RUN chmod +x /app/auto-pprof

# 设置默认的启动命令，允许传递参数
ENTRYPOINT ["/app/auto-pprof"]

