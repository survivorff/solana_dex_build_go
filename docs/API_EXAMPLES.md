# API使用示例

本文档提供了Solana DEX交易编码服务的详细API使用示例。

## 目录

- [基础配置](#基础配置)
- [交易编码](#交易编码)
- [交易测试](#交易测试)
- [DEX管理](#dex管理)
- [配置管理](#配置管理)
- [错误处理](#错误处理)
- [完整示例](#完整示例)

## 基础配置

### 服务地址

```
Base URL: http://localhost:8080/api/v1
Content-Type: application/json
```

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "timestamp": 1703123456
}
```

## 交易编码

### 1. 编码Raydium交换交易

```bash
curl -X POST http://localhost:8080/api/v1/encode/swap \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "raydium",
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_in": 1000000000,
    "slippage": 0.005,
    "priority_fee": 5000,
    "user_wallet": "你的钱包地址"
  }'
```

响应：
```json
{
  "success": true,
  "transaction": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAEDArczbMia1tLmq2poP39/+Hhqfz...",
  "estimated_fee": 5000,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. 编码Pumpfun交换交易

```bash
curl -X POST http://localhost:8080/api/v1/encode/swap \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "pumpfun",
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "代币地址",
    "amount_in": 100000000,
    "slippage": 0.01,
    "user_wallet": "你的钱包地址"
  }'
```

### 3. 编码流动性添加交易

```bash
curl -X POST http://localhost:8080/api/v1/encode/liquidity \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "raydium",
    "operation": "add",
    "token_a_mint": "So11111111111111111111111111111111111111112",
    "token_b_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_a": 1000000000,
    "amount_b": 1000000,
    "slippage": 0.005,
    "user_wallet": "你的钱包地址"
  }'
```

### 4. 编码流动性移除交易

```bash
curl -X POST http://localhost:8080/api/v1/encode/liquidity \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "pumpswap",
    "operation": "remove",
    "token_a_mint": "So11111111111111111111111111111111111111112",
    "token_b_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_a": 500000000,
    "amount_b": 500000,
    "slippage": 0.01,
    "user_wallet": "你的钱包地址"
  }'
```

## 交易测试

### 1. 模拟交易执行

```bash
curl -X POST http://localhost:8080/api/v1/test/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": "base64编码的交易数据",
    "private_key": "你的私钥(Base58编码)"
  }'
```

响应：
```json
{
  "success": true,
  "logs": [
    "Program 11111111111111111111111111111112 invoke [1]",
    "Program 11111111111111111111111111111112 success"
  ],
  "gas_used": 5000
}
```

### 2. 实际交易上链

```bash
curl -X POST http://localhost:8080/api/v1/test/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": "base64编码的交易数据",
    "simulate_only": false,
    "private_key": "你的私钥(Base58编码)"
  }'
```

响应：
```json
{
  "success": true,
  "signature": "5VERv8NMvQX9TuWicJG5tRkakgBtAHpf6Ki8b4tHoADRhGVgU3xmNrpF2VuGHBEjwqtxJVwqzQXzjQGhFXxSMA7VRUVv",
  "logs": [
    "Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 invoke [1]",
    "Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 consumed 12345 of 200000 compute units",
    "Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success"
  ]
}
```

## DEX管理

### 1. 获取所有DEX列表

```bash
curl http://localhost:8080/api/v1/dex/list
```

响应：
```json
{
  "success": true,
  "data": [
    {
      "name": "raydium",
      "program_id": "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
      "router_address": "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS",
      "endpoints": {
        "swap": "https://api.raydium.io/v2/sdk/swap",
        "pools": "https://api.raydium.io/v2/sdk/liquidity/mainnet.json"
      },
      "enabled": true,
      "status": "online"
    },
    {
      "name": "pumpfun",
      "program_id": "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
      "router_address": "39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg",
      "endpoints": {
        "api": "https://pumpportal.fun/api"
      },
      "enabled": true,
      "status": "online"
    }
  ],
  "message": "DEX list retrieved successfully"
}
```

### 2. 获取指定DEX信息

```bash
curl http://localhost:8080/api/v1/dex/raydium
```

### 3. 获取DEX流动性池信息

```bash
curl "http://localhost:8080/api/v1/dex/raydium/pools?limit=10&offset=0"
```

响应：
```json
{
  "success": true,
  "data": {
    "pools": [
      {
        "address": "58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2",
        "token_a_mint": "So11111111111111111111111111111111111111112",
        "token_b_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        "token_a_name": "SOL",
        "token_b_name": "USDC",
        "reserve_a": 1000000000000,
        "reserve_b": 50000000000,
        "liquidity": 223606797749,
        "fee_rate": 0.0025,
        "tvl": 100000000,
        "volume_24h": 5000000,
        "apr": 15.5
      }
    ],
    "total": 150,
    "limit": 10,
    "offset": 0
  }
}
```

### 4. 获取交易报价

```bash
curl "http://localhost:8080/api/v1/dex/raydium/quote?inputMint=So11111111111111111111111111111111111111112&outputMint=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v&amountIn=1000000000"
```

响应：
```json
{
  "success": true,
  "data": {
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_in": 1000000000,
    "amount_out": 50000000,
    "min_amount_out": 49750000,
    "price_impact": 0.001,
    "fee": 25000,
    "route": [
      {
        "dex": "raydium",
        "pool_id": "58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2",
        "input_mint": "So11111111111111111111111111111111111111112",
        "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        "amount_in": 1000000000,
        "amount_out": 50000000
      }
    ]
  }
}
```

### 5. 验证交换请求

```bash
curl -X POST http://localhost:8080/api/v1/dex/raydium/validate/swap \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "raydium",
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_in": 1000000000,
    "slippage": 0.005,
    "user_wallet": "你的钱包地址"
  }'
```

### 6. 检查DEX状态

```bash
curl http://localhost:8080/api/v1/dex/raydium/status
```

响应：
```json
{
  "success": true,
  "data": {
    "dex": "raydium",
    "status": "online"
  }
}
```

## 配置管理

### 1. 获取系统配置

```bash
curl http://localhost:8080/api/v1/config/
```

### 2. 获取配置摘要

```bash
curl http://localhost:8080/api/v1/config/summary
```

响应：
```json
{
  "success": true,
  "data": {
    "server_port": 8080,
    "server_mode": "debug",
    "solana_network": "mainnet",
    "solana_rpc": "https://api.mainnet-beta.solana.com",
    "total_dexes": 3,
    "enabled_dexes": 3,
    "log_level": "info"
  }
}
```

### 3. 获取DEX配置

```bash
curl http://localhost:8080/api/v1/config/dex
```

### 4. 验证配置

```bash
curl -X POST http://localhost:8080/api/v1/config/validate
```

### 5. 启用/禁用DEX

```bash
# 启用DEX
curl -X POST http://localhost:8080/api/v1/config/dex/pumpfun/enable

# 禁用DEX
curl -X POST http://localhost:8080/api/v1/config/dex/pumpfun/disable
```

## 错误处理

### 常见错误响应格式

```json
{
  "error": "错误描述",
  "code": 400,
  "details": "详细错误信息"
}
```

### 错误代码说明

| 状态码 | 说明 | 示例 |
|--------|------|------|
| 400 | 请求参数错误 | 无效的代币地址 |
| 404 | 资源不存在 | DEX不存在 |
| 500 | 服务器内部错误 | 网络连接失败 |

### 错误处理示例

```bash
# 无效的DEX类型
curl -X POST http://localhost:8080/api/v1/encode/swap \
  -H "Content-Type: application/json" \
  -d '{
    "dex_type": "invalid-dex",
    "input_mint": "So11111111111111111111111111111111111111112",
    "output_mint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    "amount_in": 1000000000,
    "user_wallet": "你的钱包地址"
  }'
```

响应：
```json
{
  "success": false,
  "error": "DEX adapter not found: invalid-dex"
}
```

## 完整示例

### Go语言客户端

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const BaseURL = "http://localhost:8080/api/v1"

type Client struct {
    httpClient *http.Client
    baseURL    string
}

type SwapRequest struct {
    DEXType    string  `json:"dex_type"`
    InputMint  string  `json:"input_mint"`
    OutputMint string  `json:"output_mint"`
    AmountIn   uint64  `json:"amount_in"`
    Slippage   float64 `json:"slippage"`
    UserWallet string  `json:"user_wallet"`
}

type TransactionResponse struct {
    Success      bool   `json:"success"`
    Transaction  string `json:"transaction,omitempty"`
    EstimatedFee uint64 `json:"estimated_fee,omitempty"`
    RequestID    string `json:"request_id,omitempty"`
    Error        string `json:"error,omitempty"`
}

func NewClient() *Client {
    return &Client{
        httpClient: &http.Client{},
        baseURL:    BaseURL,
    }
}

func (c *Client) EncodeSwap(req SwapRequest) (*TransactionResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Post(
        c.baseURL+"/encode/swap",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result TransactionResponse
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    client := NewClient()

    // 创建交换请求
    swapReq := SwapRequest{
        DEXType:    "raydium",
        InputMint:  "So11111111111111111111111111111111111111112", // SOL
        OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
        AmountIn:   1000000000, // 1 SOL
        Slippage:   0.005,      // 0.5%
        UserWallet: "你的钱包地址",
    }

    // 编码交易
    resp, err := client.EncodeSwap(swapReq)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    if resp.Success {
        fmt.Printf("交易编码成功!\n")
        fmt.Printf("请求ID: %s\n", resp.RequestID)
        fmt.Printf("预估费用: %d lamports\n", resp.EstimatedFee)
        fmt.Printf("交易数据: %s\n", resp.Transaction[:50]+"...")
    } else {
        fmt.Printf("交易编码失败: %s\n", resp.Error)
    }
}
```

### Python客户端

```python
import requests
import json

class SolanaDEXClient:
    def __init__(self, base_url="http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})
    
    def encode_swap(self, dex_type, input_mint, output_mint, amount_in, 
                   slippage=0.005, user_wallet=None, priority_fee=0):
        """编码交换交易"""
        data = {
            "dex_type": dex_type,
            "input_mint": input_mint,
            "output_mint": output_mint,
            "amount_in": amount_in,
            "slippage": slippage,
            "user_wallet": user_wallet,
            "priority_fee": priority_fee
        }
        
        response = self.session.post(f"{self.base_url}/encode/swap", json=data)
        return response.json()
    
    def get_dex_list(self):
        """获取DEX列表"""
        response = self.session.get(f"{self.base_url}/dex/list")
        return response.json()
    
    def get_quote(self, dex_name, input_mint, output_mint, amount_in):
        """获取报价"""
        params = {
            "inputMint": input_mint,
            "outputMint": output_mint,
            "amountIn": amount_in
        }
        response = self.session.get(f"{self.base_url}/dex/{dex_name}/quote", params=params)
        return response.json()

# 使用示例
if __name__ == "__main__":
    client = SolanaDEXClient()
    
    # 获取DEX列表
    dexes = client.get_dex_list()
    print("支持的DEX:", [dex['name'] for dex in dexes['data']])
    
    # 编码交换交易
    result = client.encode_swap(
        dex_type="raydium",
        input_mint="So11111111111111111111111111111111111111112",  # SOL
        output_mint="EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",  # USDC
        amount_in=1000000000,  # 1 SOL
        user_wallet="你的钱包地址"
    )
    
    if result['success']:
        print(f"交易编码成功! 请求ID: {result['request_id']}")
        print(f"预估费用: {result['estimated_fee']} lamports")
    else:
        print(f"交易编码失败: {result['error']}")
```

### JavaScript/Node.js客户端

```javascript
const axios = require('axios');

class SolanaDEXClient {
    constructor(baseURL = 'http://localhost:8080/api/v1') {
        this.baseURL = baseURL;
        this.client = axios.create({
            baseURL: this.baseURL,
            headers: {
                'Content-Type': 'application/json'
            }
        });
    }

    async encodeSwap({
        dexType,
        inputMint,
        outputMint,
        amountIn,
        slippage = 0.005,
        userWallet,
        priorityFee = 0
    }) {
        try {
            const response = await this.client.post('/encode/swap', {
                dex_type: dexType,
                input_mint: inputMint,
                output_mint: outputMint,
                amount_in: amountIn,
                slippage: slippage,
                user_wallet: userWallet,
                priority_fee: priorityFee
            });
            return response.data;
        } catch (error) {
            throw new Error(`编码交易失败: ${error.response?.data?.error || error.message}`);
        }
    }

    async getDEXList() {
        try {
            const response = await this.client.get('/dex/list');
            return response.data;
        } catch (error) {
            throw new Error(`获取DEX列表失败: ${error.response?.data?.error || error.message}`);
        }
    }

    async getQuote(dexName, inputMint, outputMint, amountIn) {
        try {
            const response = await this.client.get(`/dex/${dexName}/quote`, {
                params: {
                    inputMint,
                    outputMint,
                    amountIn
                }
            });
            return response.data;
        } catch (error) {
            throw new Error(`获取报价失败: ${error.response?.data?.error || error.message}`);
        }
    }
}

// 使用示例
async function main() {
    const client = new SolanaDEXClient();

    try {
        // 获取DEX列表
        const dexes = await client.getDEXList();
        console.log('支持的DEX:', dexes.data.map(dex => dex.name));

        // 编码交换交易
        const result = await client.encodeSwap({
            dexType: 'raydium',
            inputMint: 'So11111111111111111111111111111111111111112',  // SOL
            outputMint: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v', // USDC
            amountIn: 1000000000, // 1 SOL
            userWallet: '你的钱包地址'
        });

        if (result.success) {
            console.log(`交易编码成功! 请求ID: ${result.request_id}`);
            console.log(`预估费用: ${result.estimated_fee} lamports`);
        } else {
            console.log(`交易编码失败: ${result.error}`);
        }
    } catch (error) {
        console.error('错误:', error.message);
    }
}

main();
```

## 注意事项

1. **安全性**: 永远不要在客户端代码中硬编码私钥
2. **网络**: 确保使用正确的Solana网络（mainnet/devnet/testnet）
3. **费用**: 交易费用会根据网络拥堵情况变化
4. **滑点**: 设置合理的滑点容忍度以避免交易失败
5. **测试**: 在主网使用前，先在devnet上进行充分测试
6. **错误处理**: 始终检查API响应中的success字段
7. **限流**: 注意API调用频率限制

## 更多资源

- [Solana官方文档](https://docs.solana.com/)
- [Raydium文档](https://docs.raydium.io/)
- [项目GitHub仓库](https://github.com/your-repo/solana-dex-service)