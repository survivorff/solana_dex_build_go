package tests

import (
	"testing"
	"time"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/services"
	"solana-dex-service/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionService 测试交易服务
func TestTransactionService(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建交易服务
	transactionService := services.NewTransactionService(cfg)
	require.NotNil(t, transactionService)

	// 测试获取支持的DEX列表
	supportedDEXes := transactionService.GetSupportedDEXes()
	assert.True(t, len(supportedDEXes) > 0)
	assert.Contains(t, supportedDEXes, "raydium")

	// 测试获取DEX适配器
	adapter, err := transactionService.GetDEXAdapter("raydium")
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.Equal(t, "raydium", adapter.GetName())

	// 测试获取不存在的DEX适配器
	adapter, err = transactionService.GetDEXAdapter("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, adapter)
}

// TestTransactionServiceEncodeSwap 测试交换交易编码
func TestTransactionServiceEncodeSwap(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建交易服务
	transactionService := services.NewTransactionService(cfg)

	// 创建有效的交换请求
	swapReq := &types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112", // SOL
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountIn:   1000000000, // 1 SOL
		Slippage:   0.005,      // 0.5%
		UserWallet: "11111111111111111111111111111112",
	}

	// 测试编码交换交易
	resp, err := transactionService.EncodeSwapTransaction(swapReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 验证响应
	if resp.Success {
		assert.NotEmpty(t, resp.Transaction)
		assert.NotEmpty(t, resp.RequestID)
		assert.True(t, resp.EstimatedFee > 0)
	} else {
		// 如果失败，应该有错误信息
		assert.NotEmpty(t, resp.Error)
	}

	// 测试无效的DEX类型
	invalidSwapReq := &types.SwapRequest{
		DEXType:    "invalid-dex",
		InputMint:  "So11111111111111111111111111111111111111112",
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountIn:   1000000000,
		UserWallet: "11111111111111111111111111111112",
	}

	resp, err = transactionService.EncodeSwapTransaction(invalidSwapReq)
	assert.NoError(t, err) // 服务层不应该返回错误，而是在响应中标记失败
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "DEX adapter not found")
}

// TestTransactionServiceEncodeLiquidity 测试流动性交易编码
func TestTransactionServiceEncodeLiquidity(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建交易服务
	transactionService := services.NewTransactionService(cfg)

	// 创建有效的流动性请求
	liquidityReq := &types.LiquidityRequest{
		DEXType:     "raydium",
		Operation:   "add",
		TokenAMint:  "So11111111111111111111111111111111111111112", // SOL
		TokenBMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountA:     1000000000, // 1 SOL
		AmountB:     1000000,    // 1 USDC
		Slippage:    0.005,      // 0.5%
		UserWallet:  "11111111111111111111111111111112",
	}

	// 测试编码流动性交易
	resp, err := transactionService.EncodeLiquidityTransaction(liquidityReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 验证响应
	if resp.Success {
		assert.NotEmpty(t, resp.Transaction)
		assert.NotEmpty(t, resp.RequestID)
		assert.True(t, resp.EstimatedFee > 0)
	} else {
		// 如果失败，应该有错误信息
		assert.NotEmpty(t, resp.Error)
	}

	// 测试Pumpfun的流动性请求（应该失败）
	pumpfunLiquidityReq := &types.LiquidityRequest{
		DEXType:    "pumpfun",
		Operation:  "add",
		UserWallet: "11111111111111111111111111111112",
	}

	resp, err = transactionService.EncodeLiquidityTransaction(pumpfunLiquidityReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "does not support")
}

// TestDEXService 测试DEX服务
func TestDEXService(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建DEX服务
	dexService := services.NewDEXService(cfg)
	require.NotNil(t, dexService)

	// 创建交易服务并设置到DEX服务
	transactionService := services.NewTransactionService(cfg)
	dexService.SetTransactionService(transactionService)

	// 测试列出所有DEX
	dexes, err := dexService.ListDEXes()
	assert.NoError(t, err)
	assert.True(t, len(dexes) > 0)

	// 验证DEX信息
	found := false
	for _, dex := range dexes {
		if dex.Name == "raydium" {
			found = true
			assert.Equal(t, "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8", dex.ProgramID)
			assert.True(t, dex.Enabled)
			assert.Equal(t, "online", dex.Status)
			break
		}
	}
	assert.True(t, found, "Raydium DEX should be found in the list")

	// 测试获取指定DEX
	dex, err := dexService.GetDEX("raydium")
	assert.NoError(t, err)
	assert.NotNil(t, dex)
	assert.Equal(t, "raydium", dex.Name)

	// 测试获取不存在的DEX
	dex, err = dexService.GetDEX("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, dex)

	// 测试获取启用的DEX
	enabledDEXes, err := dexService.GetEnabledDEXes()
	assert.NoError(t, err)
	assert.True(t, len(enabledDEXes) > 0)

	// 测试检查DEX状态
	status, err := dexService.CheckDEXStatus("raydium")
	assert.NoError(t, err)
	assert.Equal(t, "online", status)

	// 测试验证DEX请求
	swapReq := &types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112",
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountIn:   1000000000,
		Slippage:   0.005,
		UserWallet: "11111111111111111111111111111112",
	}

	err = dexService.ValidateDEXRequest("raydium", swapReq)
	assert.NoError(t, err)
}

// TestConfigServiceMethods 测试配置服务方法
func TestConfigServiceMethods(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建配置服务
	configService := services.NewConfigService(cfg)
	require.NotNil(t, configService)

	// 测试获取配置
	gotConfig := configService.GetConfig()
	assert.NotNil(t, gotConfig)
	assert.Equal(t, cfg.Server.Port, gotConfig.Server.Port)

	// 测试获取DEX配置
	dexConfigs := configService.GetDEXConfig()
	assert.True(t, len(dexConfigs) > 0)

	// 测试获取服务器配置
	serverConfig := configService.GetServerConfig()
	assert.NotNil(t, serverConfig)
	assert.Equal(t, 8080, serverConfig.Port)

	// 测试获取Solana配置
	solanaConfig := configService.GetSolanaConfig()
	assert.NotNil(t, solanaConfig)
	assert.Equal(t, "mainnet", solanaConfig.Network)

	// 测试验证配置
	err := configService.ValidateConfig()
	assert.NoError(t, err)

	// 测试获取配置摘要
	summary := configService.GetConfigSummary()
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "server_port")
	assert.Contains(t, summary, "solana_network")
	assert.Contains(t, summary, "total_dexes")
	assert.Contains(t, summary, "enabled_dexes")
}

// TestTransactionTestRequest 测试交易测试请求验证
func TestTransactionTestRequest(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建交易服务
	transactionService := services.NewTransactionService(cfg)

	// 测试无效的交易数据
	invalidTestReq := &types.TransactionTestRequest{
		Transaction:  "invalid-base64-data",
		SimulateOnly: true,
		PrivateKey:   "invalid-private-key",
	}

	resp, err := transactionService.TestTransaction(invalidTestReq)
	assert.NoError(t, err) // 服务层不应该返回错误
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Error)

	// 测试模拟交易
	resp, err = transactionService.SimulateTransaction(invalidTestReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
}

// TestServiceIntegration 测试服务集成
func TestServiceIntegration(t *testing.T) {
	// 创建测试配置
	cfg := createTestConfig()

	// 创建所有服务
	transactionService := services.NewTransactionService(cfg)
	dexService := services.NewDEXService(cfg)
	configService := services.NewConfigService(cfg)

	// 设置服务依赖
	dexService.SetTransactionService(transactionService)

	// 测试完整的交易流程
	swapReq := &types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112",
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountIn:   1000000000,
		Slippage:   0.005,
		UserWallet: "11111111111111111111111111111112",
	}

	// 1. 验证DEX请求
	err := dexService.ValidateDEXRequest("raydium", swapReq)
	assert.NoError(t, err)

	// 2. 编码交易
	txResp, err := transactionService.EncodeSwapTransaction(swapReq)
	assert.NoError(t, err)
	assert.NotNil(t, txResp)

	// 3. 验证配置
	err = configService.ValidateConfig()
	assert.NoError(t, err)

	// 4. 检查DEX状态
	status, err := dexService.CheckDEXStatus("raydium")
	assert.NoError(t, err)
	assert.Equal(t, "online", status)
}

// createTestConfig 创建测试配置
func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			Mode:         "debug",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Solana: config.SolanaConfig{
			RPCURL:     "https://api.mainnet-beta.solana.com",
			WSURL:      "wss://api.mainnet-beta.solana.com",
			Network:    "mainnet",
			Timeout:    30 * time.Second,
			RetryCount: 3,
			Commitment: "confirmed",
		},
		DEXes: []config.DEXConfig{
			{
				Name:      "raydium",
				ProgramID: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
				RouterAddress: "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS",
				Endpoints: map[string]string{
					"swap":  "https://api.raydium.io/v2/sdk/swap",
					"pools": "https://api.raydium.io/v2/sdk/liquidity/mainnet.json",
					"quote": "https://quote-api.jup.ag/v6/quote",
				},
				Enabled:    true,
				Timeout:    30 * time.Second,
				RetryCount: 3,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				Name:      "pumpfun",
				ProgramID: "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
				RouterAddress: "39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg",
				Endpoints: map[string]string{
					"api":   "https://pumpportal.fun/api",
					"trade": "https://pumpportal.fun/api/trade",
					"quote": "https://pumpportal.fun/api/quote",
				},
				Enabled:    true,
				Timeout:    30 * time.Second,
				RetryCount: 3,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				Name:      "pumpswap",
				ProgramID: "PSwapMdSai8tjrEXcxFeQth87xC4rRsa4VA5mhGhXkP",
				RouterAddress: "PSwapRouterV1111111111111111111111111111111",
				Endpoints: map[string]string{
					"api":  "https://api.pumpswap.com/v1",
					"swap": "https://api.pumpswap.com/v1/swap",
					"quote": "https://api.pumpswap.com/v1/quote",
				},
				Enabled:    true,
				Timeout:    30 * time.Second,
				RetryCount: 3,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		Logging: config.LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			FilePath:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		},
		Security: config.SecurityConfig{
			EnableHTTPS:    false,
			RateLimitRPS:   100,
			MaxRequestSize: 1024 * 1024,
		},
	}
}