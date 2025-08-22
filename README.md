# Solana DEX交易编码服务

一个基于Go语言开发的Solana区块链DEX交易编码服务，专门用于在多个去中心化交易所（DEX）上进行交易指令的编码、打包和上链操作。

## 🚀 特性

- **多DEX支持**: 支持Raydium、Pumpfun、PumpSwap等主流DEX
- **统一接口**: 为不同DEX提供统一的交易编码接口
- **高度可扩展**: 采用适配器模式，易于添加新的DEX支持
- **完整测试**: 提供交易模拟和实际上链测试功能
- **RESTful API**: 提供完整的HTTP API接口
- **类型安全**: 使用TypeScript风格的Go结构体定义
- **配置管理**: 灵活的配置管理系统

## 📋 支持的DEX

| DEX | 状态 | 交换 | 流动性 | 说明 |
|-----|------|------|--------|------|
| Raydium | ✅ | ✅ | ✅ | 完整支持 |
| Pumpfun | ✅ | ✅ | ❌ | 仅支持bonding curve交易 |
| PumpSwap | ✅ | ✅ | ✅ | 完整支持 |

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **Solana SDK**: gagliardetto/solana-go
- **配置**: YAML
- **测试**: testify
- **HTTP客户端**: 标准库net/http

## 📦 安装

### 前置要求

- Go 1.21 或更高版本
- Git

### 克隆项目

```bash
git clone <repository-url>
cd solana-dex-service
```

### 安装依赖

```bash
go mod tidy
```

## ⚙️ 配置

### 配置文件

复制示例配置文件并根据需要修改：

```bash
cp config/config.yaml config/config.local.yaml
```

主要配置项：

```yaml
# HTTP服务器配置
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"  # debug, release, test

# Solana网络配置
solana:
  rpc_url: "https://api.mainnet-beta.solana.com"
  network: "mainnet"  # mainnet, devnet, testnet
  commitment: "confirmed"

# DEX配置
dexes:
  - name: "raydium"
    program_id: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
    enabled: true
```

### 环境变量

可以通过环境变量覆盖配置：

```bash
export SOLANA_RPC_URL="https://api.devnet.solana.com"
export SERVER_PORT=3000
```

## 🚀 运行

### 开发模式

```bash
go run cmd/main.go
```

### 生产模式

```bash
# 构建
go build -o solana-dex-service cmd/main.go

# 运行
./solana-dex-service
```

### 使用Docker

```bash
# 构建镜像
docker build -t solana-dex-service .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/config:/app/config solana-dex-service
```

## 📚 API文档

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`

### 核心接口

#### 1. 编码交换交易

```http
POST /api/v1/encode/swap
```

**请求体**:
```json
{
  "dex_type": "raydium",
  "input_mint": "So11111111111111111111111111111111111111112",
  "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
  "amount_in": 1000000000,
  "slippage": 0.005,
  "priority_fee": 5000,
  "user_wallet": "你的钱包地址"
}
```

**响应**:
```json
{
  "success": true,
  "transaction": "base64编码的交易数据",
  "estimated_fee": 5000,
  "request_id": "uuid"
}
```

#### 2. 测试交易上链

```http
POST /api/v1/test/transaction
```

**请求体**:
```json
{
  "transaction": "base64编码的交易数据",
  "simulate_only": true,
  "private_key": "你的私钥(Base58编码)"
}
```

#### 3. 获取DEX列表

```http
GET /api/v1/dex/list
```

#### 4. 获取交易报价

```http
GET /api/v1/dex/{dex_name}/quote?inputMint=xxx&outputMint=yyy&amountIn=1000000000
```

### 完整API文档

启动服务后访问 `http://localhost:8080/docs` 查看完整的API文档。

## 🧪 测试

### 运行所有测试

```bash
go test ./... -v
```

### 运行特定测试

```bash
# 配置测试
go test ./tests -run TestConfig -v

# 适配器测试
go test ./tests -run TestAdapter -v

# 集成测试
go test ./tests -run TestIntegration -v
```

### 测试覆盖率

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📝 使用示例

### Go客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type SwapRequest struct {
    DEXType    string  `json:"dex_type"`
    InputMint  string  `json:"input_mint"`
    OutputMint string  `json:"output_mint"`
    AmountIn   uint64  `json:"amount_in"`
    Slippage   float64 `json:"slippage"`
    UserWallet string  `json:"user_wallet"`
}

func main() {
    req := SwapRequest{
        DEXType:    "raydium",
        InputMint:  "So11111111111111111111111111111111111111112", // SOL
        OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
        AmountIn:   1000000000, // 1 SOL
        Slippage:   0.005,      // 0.5%
        UserWallet: "你的钱包地址",
    }

    jsonData, _ := json.Marshal(req)
    resp, err := http.Post(
        "http://localhost:8080/api/v1/encode/swap",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Printf("交易编码结果: %+v\n", result)
}
```

### cURL示例

```bash
# 编码交换交易
curl -X POST http://localhost:8080/api/v1/encode/swap \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "raydium",
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_in": 1000000000,
    "slippage": 0.005,
    "user_wallet": "你的钱包地址"
  }'

# 获取DEX列表
curl http://localhost:8080/api/v1/dex/list

# 健康检查
curl http://localhost:8080/health
```

## 🔧 开发

### 项目结构

```
.
├── cmd/                    # 应用程序入口
│   └── main.go
├── internal/               # 内部包
│   ├── adapters/          # DEX适配器
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP处理器
│   ├── services/          # 业务逻辑
│   └── models/            # 数据模型
├── pkg/                   # 公共包
│   ├── types/             # 类型定义
│   └── utils/             # 工具函数
├── config/                # 配置文件
├── tests/                 # 测试文件
├── docs/                  # 文档
└── README.md
```

### 添加新的DEX支持

1. 在 `internal/adapters/` 目录下创建新的适配器文件
2. 实现 `types.DEXAdapter` 接口
3. 在 `services/transaction.go` 中注册新适配器
4. 在配置文件中添加DEX配置
5. 编写测试用例

示例：

```go
// internal/adapters/newdex.go
package adapters

import (
    "solana-dex-service/internal/config"
    "solana-dex-service/pkg/types"
)

type NewDEXAdapter struct {
    *BaseAdapter
}

func NewNewDEXAdapter(cfg *config.DEXConfig) (*NewDEXAdapter, error) {
    return &NewDEXAdapter{
        BaseAdapter: NewBaseAdapter(cfg.Name, cfg),
    }, nil
}

func (n *NewDEXAdapter) GetQuote(inputMint, outputMint string, amountIn uint64) (*types.QuoteResponse, error) {
    // 实现报价逻辑
    return nil, nil
}

// 实现其他接口方法...
```

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 为公共函数添加注释
- 编写单元测试
- 使用有意义的变量和函数名

## 🐛 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 查找占用端口的进程
   lsof -i :8080
   # 或修改配置文件中的端口
   ```

2. **Solana RPC连接失败**
   - 检查网络连接
   - 验证RPC URL是否正确
   - 尝试使用其他RPC节点

3. **交易编码失败**
   - 验证代币地址格式
   - 检查钱包地址是否有效
   - 确认DEX是否支持该交易对

4. **测试失败**
   - 确保测试环境配置正确
   - 检查网络连接
   - 验证测试数据的有效性

### 日志

查看详细日志：

```bash
# 设置日志级别
export LOG_LEVEL=debug

# 运行服务
go run cmd/main.go
```

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 贡献指南

- 确保代码通过所有测试
- 添加适当的测试用例
- 更新相关文档
- 遵循现有的代码风格

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [gagliardetto/solana-go](https://github.com/gagliardetto/solana-go) - Solana Go SDK
- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Web框架
- [stretchr/testify](https://github.com/stretchr/testify) - 测试框架

## 📞 支持

如果您遇到问题或有疑问，请：

1. 查看 [FAQ](docs/FAQ.md)
2. 搜索现有的 [Issues](../../issues)
3. 创建新的 Issue
4. 联系维护者

---

**注意**: 本项目仅用于教育和开发目的。在生产环境中使用前，请进行充分的测试和安全审计。交易涉及风险，请谨慎操作。