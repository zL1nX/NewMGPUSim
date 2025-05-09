# 使用 Ubuntu 22.04 作为基础镜像
FROM ubuntu:22.04

# 设置环境变量
ENV GO_VERSION=1.22.8

# 更新软件包列表并安装必要的软件包
RUN apt-get update && \
    apt-get install -y wget git vim build-essential make telnet xinetd curl iputils-ping lsof && \
    rm -rf /var/lib/apt/lists/*

# 下载并安装 Go
RUN wget https://golang.google.cn/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# 设置 Go 环境变量
ENV PATH="/usr/local/go/bin:${PATH}"

# 验证 Go 安装
RUN go version && \
    go env -w GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /home/mgpu/

# 复制当前宿主机目录内容到 Docker 镜像工作目录
# COPY . .

# # 暴露应用程序端口（根据需要修改）
EXPOSE 5173