# 部署指南

本文档详细介绍了如何在不同环境中部署Solana DEX交易编码服务。

## 目录

- [环境要求](#环境要求)
- [本地开发部署](#本地开发部署)
- [生产环境部署](#生产环境部署)
- [Docker部署](#docker部署)
- [云平台部署](#云平台部署)
- [监控和日志](#监控和日志)
- [故障排除](#故障排除)

## 环境要求

### 系统要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+), macOS, Windows
- **CPU**: 2核心以上
- **内存**: 4GB以上
- **存储**: 10GB可用空间
- **网络**: 稳定的互联网连接

### 软件依赖

- **Go**: 1.21或更高版本
- **Git**: 用于代码管理
- **Docker**: (可选) 用于容器化部署

## 本地开发部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd solana-dex-service
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境

复制配置文件：
```bash
cp config/config.yaml config/config.local.yaml
```

编辑配置文件：
```yaml
# config/config.local.yaml
server:
  port: 8080
  mode: "debug"

solana:
  rpc_url: "https://api.devnet.solana.com"  # 开发环境使用devnet
  network: "devnet"
```

### 4. 运行服务

```bash
# 开发模式运行
go run cmd/main.go

# 或者构建后运行
go build -o solana-dex-service cmd/main.go
./solana-dex-service
```

### 5. 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# 获取DEX列表
curl http://localhost:8080/api/v1/dex/list
```

## 生产环境部署

### 1. 服务器准备

#### Ubuntu/Debian系统

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y curl wget git build-essential

# 安装Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### CentOS/RHEL系统

```bash
# 更新系统
sudo yum update -y

# 安装必要工具
sudo yum install -y curl wget git gcc

# 安装Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. 部署应用

```bash
# 创建应用目录
sudo mkdir -p /opt/solana-dex-service
sudo chown $USER:$USER /opt/solana-dex-service
cd /opt/solana-dex-service

# 克隆代码
git clone <repository-url> .

# 安装依赖
go mod tidy

# 构建应用
go build -o solana-dex-service cmd/main.go
```

### 3. 配置生产环境

创建生产配置文件：
```bash
sudo mkdir -p /etc/solana-dex-service
sudo cp config/config.yaml /etc/solana-dex-service/config.yaml
```

编辑生产配置：
```yaml
# /etc/solana-dex-service/config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "release"
  read_timeout: 30s
  write_timeout: 30s

solana:
  rpc_url: "https://api.mainnet-beta.solana.com"
  network: "mainnet"
  timeout: 30s
  retry_count: 3
  commitment: "confirmed"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/solana-dex-service/app.log"
  max_size: 100
  max_backups: 5
  max_age: 30

security:
  enable_https: true
  cert_file: "/etc/ssl/certs/solana-dex-service.crt"
  key_file: "/etc/ssl/private/solana-dex-service.key"
  rate_limit_rps: 1000
  max_request_size: 1048576
```

### 4. 创建系统服务

创建systemd服务文件：
```bash
sudo tee /etc/systemd/system/solana-dex-service.service > /dev/null <<EOF
[Unit]
Description=Solana DEX Transaction Encoding Service
After=network.target

[Service]
Type=simple
User=solana-dex
Group=solana-dex
WorkingDirectory=/opt/solana-dex-service
ExecStart=/opt/solana-dex-service/solana-dex-service
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/etc/solana-dex-service/config.yaml

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/solana-dex-service

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
```

创建专用用户：
```bash
sudo useradd -r -s /bin/false solana-dex
sudo chown -R solana-dex:solana-dex /opt/solana-dex-service
```

创建日志目录：
```bash
sudo mkdir -p /var/log/solana-dex-service
sudo chown solana-dex:solana-dex /var/log/solana-dex-service
```

### 5. 启动服务

```bash
# 重新加载systemd配置
sudo systemctl daemon-reload

# 启用服务
sudo systemctl enable solana-dex-service

# 启动服务
sudo systemctl start solana-dex-service

# 检查服务状态
sudo systemctl status solana-dex-service

# 查看日志
sudo journalctl -u solana-dex-service -f
```

### 6. 配置反向代理 (Nginx)

安装Nginx：
```bash
sudo apt install -y nginx  # Ubuntu/Debian
# 或
sudo yum install -y nginx  # CentOS/RHEL
```

创建Nginx配置：
```bash
sudo tee /etc/nginx/sites-available/solana-dex-service > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;
    
    # 重定向到HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL配置
    ssl_certificate /etc/ssl/certs/your-domain.crt;
    ssl_certificate_key /etc/ssl/private/your-domain.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    
    # 限流
    limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;
    
    location / {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # 健康检查
    location /health {
        access_log off;
        proxy_pass http://127.0.0.1:8080/health;
    }
}
EOF
```

启用配置：
```bash
sudo ln -s /etc/nginx/sites-available/solana-dex-service /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Docker部署

### 1. 创建Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o solana-dex-service cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN adduser -D -s /bin/sh solana-dex

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/solana-dex-service .
COPY --from=builder /app/config ./config

# 创建日志目录
RUN mkdir -p /var/log/solana-dex-service && \
    chown -R solana-dex:solana-dex /app /var/log/solana-dex-service

# 切换到非root用户
USER solana-dex

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./solana-dex-service"]
```

### 2. 创建docker-compose.yml

```yaml
# docker-compose.yml
version: '3.8'

services:
  solana-dex-service:
    build: .
    container_name: solana-dex-service
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config:ro
      - ./logs:/var/log/solana-dex-service
    environment:
      - CONFIG_PATH=/app/config/config.yaml
      - GIN_MODE=release
    networks:
      - solana-dex-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  nginx:
    image: nginx:alpine
    container_name: solana-dex-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/ssl:ro
    depends_on:
      - solana-dex-service
    networks:
      - solana-dex-network

networks:
  solana-dex-network:
    driver: bridge

volumes:
  logs:
    driver: local
```

### 3. 构建和运行

```bash
# 构建镜像
docker build -t solana-dex-service .

# 使用docker-compose运行
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 4. Docker生产部署

```bash
# 创建生产环境配置
mkdir -p /opt/solana-dex-service/{config,logs,ssl}

# 复制配置文件
cp config/config.yaml /opt/solana-dex-service/config/

# 复制SSL证书
cp your-cert.crt /opt/solana-dex-service/ssl/
cp your-key.key /opt/solana-dex-service/ssl/

# 运行容器
docker run -d \
  --name solana-dex-service \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/solana-dex-service/config:/app/config:ro \
  -v /opt/solana-dex-service/logs:/var/log/solana-dex-service \
  solana-dex-service
```

## 云平台部署

### AWS部署

#### 使用EC2

1. **创建EC2实例**
   - 选择Ubuntu 20.04 LTS AMI
   - 实例类型：t3.medium或更高
   - 配置安全组：开放80, 443, 8080端口

2. **部署应用**
   ```bash
   # 连接到EC2实例
   ssh -i your-key.pem ubuntu@your-ec2-ip
   
   # 按照生产环境部署步骤进行部署
   ```

3. **配置负载均衡器**
   - 创建Application Load Balancer
   - 配置目标组指向EC2实例
   - 配置SSL证书

#### 使用ECS

1. **创建任务定义**
   ```json
   {
     "family": "solana-dex-service",
     "networkMode": "awsvpc",
     "requiresCompatibilities": ["FARGATE"],
     "cpu": "512",
     "memory": "1024",
     "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
     "containerDefinitions": [
       {
         "name": "solana-dex-service",
         "image": "your-account.dkr.ecr.region.amazonaws.com/solana-dex-service:latest",
         "portMappings": [
           {
             "containerPort": 8080,
             "protocol": "tcp"
           }
         ],
         "logConfiguration": {
           "logDriver": "awslogs",
           "options": {
             "awslogs-group": "/ecs/solana-dex-service",
             "awslogs-region": "us-west-2",
             "awslogs-stream-prefix": "ecs"
           }
         }
       }
     ]
   }
   ```

### Google Cloud Platform部署

#### 使用Cloud Run

```bash
# 构建并推送镜像到GCR
gcloud builds submit --tag gcr.io/PROJECT-ID/solana-dex-service

# 部署到Cloud Run
gcloud run deploy solana-dex-service \
  --image gcr.io/PROJECT-ID/solana-dex-service \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 1Gi \
  --cpu 1 \
  --max-instances 10
```

### Azure部署

#### 使用Container Instances

```bash
# 创建资源组
az group create --name solana-dex-rg --location eastus

# 部署容器
az container create \
  --resource-group solana-dex-rg \
  --name solana-dex-service \
  --image your-registry/solana-dex-service:latest \
  --cpu 1 \
  --memory 2 \
  --ports 8080 \
  --dns-name-label solana-dex-service
```

## 监控和日志

### 1. 日志配置

```yaml
# config/config.yaml
logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/solana-dex-service/app.log"
  max_size: 100  # MB
  max_backups: 5
  max_age: 30    # days
```

### 2. 日志轮转

创建logrotate配置：
```bash
sudo tee /etc/logrotate.d/solana-dex-service > /dev/null <<EOF
/var/log/solana-dex-service/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 solana-dex solana-dex
    postrotate
        systemctl reload solana-dex-service
    endscript
}
EOF
```

### 3. 监控指标

添加Prometheus监控端点：
```go
// 在main.go中添加
import "github.com/prometheus/client_golang/prometheus/promhttp"

// 添加metrics路由
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 4. 健康检查

```bash
# 创建健康检查脚本
sudo tee /usr/local/bin/health-check.sh > /dev/null <<EOF
#!/bin/bash
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ $response -eq 200 ]; then
    exit 0
else
    exit 1
fi
EOF

sudo chmod +x /usr/local/bin/health-check.sh
```

## 故障排除

### 常见问题

1. **服务无法启动**
   ```bash
   # 检查服务状态
   sudo systemctl status solana-dex-service
   
   # 查看详细日志
   sudo journalctl -u solana-dex-service -f
   
   # 检查配置文件
   go run cmd/main.go --config-check
   ```

2. **端口被占用**
   ```bash
   # 查找占用端口的进程
   sudo lsof -i :8080
   
   # 或使用netstat
   sudo netstat -tlnp | grep :8080
   ```

3. **内存不足**
   ```bash
   # 检查内存使用
   free -h
   
   # 检查进程内存使用
   ps aux | grep solana-dex-service
   ```

4. **网络连接问题**
   ```bash
   # 测试Solana RPC连接
   curl -X POST https://api.mainnet-beta.solana.com \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"getHealth"}'
   ```

### 性能优化

1. **调整Go运行时参数**
   ```bash
   export GOMAXPROCS=4
   export GOGC=100
   ```

2. **系统参数优化**
   ```bash
   # 增加文件描述符限制
   echo "* soft nofile 65536" >> /etc/security/limits.conf
   echo "* hard nofile 65536" >> /etc/security/limits.conf
   
   # 优化网络参数
   echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
   echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
   sysctl -p
   ```

3. **数据库连接池优化**
   ```yaml
   # config/config.yaml
   database:
     max_open_conns: 25
     max_idle_conns: 5
     conn_max_lifetime: 300s
   ```

### 备份和恢复

1. **配置备份**
   ```bash
   # 创建备份脚本
   #!/bin/bash
   BACKUP_DIR="/backup/solana-dex-service/$(date +%Y%m%d)"
   mkdir -p $BACKUP_DIR
   
   # 备份配置文件
   cp -r /etc/solana-dex-service $BACKUP_DIR/
   
   # 备份日志
   cp -r /var/log/solana-dex-service $BACKUP_DIR/
   
   # 压缩备份
   tar -czf $BACKUP_DIR.tar.gz $BACKUP_DIR
   rm -rf $BACKUP_DIR
   ```

2. **自动备份**
   ```bash
   # 添加到crontab
   0 2 * * * /usr/local/bin/backup-solana-dex.sh
   ```

## 安全最佳实践

1. **防火墙配置**
   ```bash
   # 使用ufw配置防火墙
   sudo ufw default deny incoming
   sudo ufw default allow outgoing
   sudo ufw allow ssh
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. **SSL/TLS配置**
   - 使用Let's Encrypt免费证书
   - 定期更新证书
   - 配置强加密套件

3. **访问控制**
   - 使用API密钥认证
   - 实施IP白名单
   - 配置速率限制

4. **定期更新**
   ```bash
   # 定期更新系统和依赖
   sudo apt update && sudo apt upgrade -y
   go get -u ./...
   ```

通过遵循本部署指南，您可以在各种环境中成功部署和运行Solana DEX交易编码服务。记住要根据您的具体需求调整配置，并定期监控服务的运行状态。