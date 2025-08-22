package tests

import (
	"testing"
	"time"

	"solana-dex-service/internal/adapters"
	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapterRegistry 测试适配器注册表
func TestAdapterRegistry(t *testing.T) {
	registry := adapters.NewAdapterRegistry()

	// 创建测试适配器配置
	raydiumConfig := &config.DEXConfig{
		Name:      "raydium",
		ProgramID: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
		Enabled:   true,
		Timeout:   30 * time.Second,
		RetryCount: 3,
	}

	// 创建Raydium适配器
	raydiumAdapter, err := adapters.NewRaydiumAdapter(raydiumConfig)
	require.NoError(t, err)
	require.NotNil(t, raydiumAdapter)

	// 注册适配器
	registry.Register("raydium", raydiumAdapter)

	// 测试获取适配器
	adapter, err := registry.Get("raydium")
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.Equal(t, "raydium", adapter.GetName())

	// 测试获取不存在的适配器
	adapter, err = registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, adapter)

	// 测试列出所有适配器
	adapterNames := registry.List()
	assert.Contains(t, adapterNames, "raydium")

	// 测试获取所有适配器
	allAdapters := registry.GetAll()
	assert.Contains(t, allAdapters, "raydium")
}

// TestRaydiumAdapter 测试Raydium适配器
func TestRaydiumAdapter(t *testing.T) {
	config := &config.DEXConfig{
		Name:      "raydium",
		ProgramID: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
		Endpoints: map[string]string{
			"quote": "https://quote-api.jup.ag/v6/quote",
			"pools": "https://api.raydium.io/v2/sdk/liquidity/mainnet.json",
		},
		Enabled:    true,
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}

	adapter, err := adapters.NewRaydiumAdapter(config)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// 测试获取名称
	assert.Equal(t, "raydium", adapter.GetName())

	// 测试验证交换请求
	validSwapReq := &types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112", // SOL
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountIn:   1000000000, // 1 SOL
		Slippage:   0.005,      // 0.5%
		UserWallet: "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(validSwapReq)
	assert.NoError(t, err)

	// 测试无效的交换请求
	invalidSwapReq := &types.SwapRequest{
		DEXType:   "raydium",
		InputMint: "", // 空的输入代币
	}

	err = adapter.ValidateRequest(invalidSwapReq)
	assert.Error(t, err)

	// 测试验证流动性请求
	validLiquidityReq := &types.LiquidityRequest{
		DEXType:     "raydium",
		Operation:   "add",
		TokenAMint:  "So11111111111111111111111111111111111111112",
		TokenBMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountA:     1000000000,
		AmountB:     1000000,
		Slippage:    0.005,
		UserWallet:  "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(validLiquidityReq)
	assert.NoError(t, err)

	// 测试构建交换指令
	swapInstruction, err := adapter.BuildSwapInstruction(validSwapReq)
	if err == nil {
		assert.NotNil(t, swapInstruction)
		assert.NotEmpty(t, swapInstruction.Data)
		assert.True(t, len(swapInstruction.Accounts) > 0)
	}

	// 测试构建流动性指令
	liquidityInstruction, err := adapter.BuildLiquidityInstruction(validLiquidityReq)
	if err == nil {
		assert.NotNil(t, liquidityInstruction)
		assert.NotEmpty(t, liquidityInstruction.Data)
		assert.True(t, len(liquidityInstruction.Accounts) > 0)
	}
}

// TestPumpfunAdapter 测试Pumpfun适配器
func TestPumpfunAdapter(t *testing.T) {
	config := &config.DEXConfig{
		Name:      "pumpfun",
		ProgramID: "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
		Endpoints: map[string]string{
			"api": "https://pumpportal.fun/api",
		},
		Enabled:    true,
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}

	adapter, err := adapters.NewPumpfunAdapter(config)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// 测试获取名称
	assert.Equal(t, "pumpfun", adapter.GetName())

	// 测试验证交换请求
	validSwapReq := &types.SwapRequest{
		DEXType:    "pumpfun",
		InputMint:  "So11111111111111111111111111111111111111112", // SOL
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // Token
		AmountIn:   1000000000, // 1 SOL
		Slippage:   0.005,      // 0.5%
		UserWallet: "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(validSwapReq)
	assert.NoError(t, err)

	// 测试流动性请求（Pumpfun不支持）
	liquidityReq := &types.LiquidityRequest{
		DEXType:    "pumpfun",
		Operation:  "add",
		UserWallet: "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(liquidityReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support liquidity operations")

	// 测试构建交换指令
	swapInstruction, err := adapter.BuildSwapInstruction(validSwapReq)
	if err == nil {
		assert.NotNil(t, swapInstruction)
		assert.NotEmpty(t, swapInstruction.Data)
		assert.True(t, len(swapInstruction.Accounts) > 0)
	}

	// 测试构建流动性指令（应该返回错误）
	liquidityInstruction, err := adapter.BuildLiquidityInstruction(liquidityReq)
	assert.Error(t, err)
	assert.Nil(t, liquidityInstruction)
}

// TestPumpSwapAdapter 测试PumpSwap适配器
func TestPumpSwapAdapter(t *testing.T) {
	config := &config.DEXConfig{
		Name:          "pumpswap",
		ProgramID:     "PSwapMdSai8tjrEXcxFeQth87xC4rRsa4VA5mhGhXkP",
		RouterAddress: "PSwapRouterV1111111111111111111111111111111",
		Endpoints: map[string]string{
			"api": "https://api.pumpswap.com/v1",
		},
		Enabled:    true,
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}

	adapter, err := adapters.NewPumpSwapAdapter(config)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// 测试获取名称
	assert.Equal(t, "pumpswap", adapter.GetName())

	// 测试验证交换请求
	validSwapReq := &types.SwapRequest{
		DEXType:    "pumpswap",
		InputMint:  "So11111111111111111111111111111111111111112", // SOL
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountIn:   1000000000, // 1 SOL
		Slippage:   0.005,      // 0.5%
		UserWallet: "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(validSwapReq)
	assert.NoError(t, err)

	// 测试验证流动性请求
	validLiquidityReq := &types.LiquidityRequest{
		DEXType:     "pumpswap",
		Operation:   "add",
		TokenAMint:  "So11111111111111111111111111111111111111112",
		TokenBMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountA:     1000000000,
		AmountB:     1000000,
		Slippage:    0.005,
		UserWallet:  "11111111111111111111111111111112",
	}

	err = adapter.ValidateRequest(validLiquidityReq)
	assert.NoError(t, err)

	// 测试构建交换指令
	swapInstruction, err := adapter.BuildSwapInstruction(validSwapReq)
	if err == nil {
		assert.NotNil(t, swapInstruction)
		assert.NotEmpty(t, swapInstruction.Data)
		assert.True(t, len(swapInstruction.Accounts) > 0)
	}

	// 测试构建流动性指令
	liquidityInstruction, err := adapter.BuildLiquidityInstruction(validLiquidityReq)
	if err == nil {
		assert.NotNil(t, liquidityInstruction)
		assert.NotEmpty(t, liquidityInstruction.Data)
		assert.True(t, len(liquidityInstruction.Accounts) > 0)
	}
}

// TestBaseAdapterValidation 测试基础适配器验证功能
func TestBaseAdapterValidation(t *testing.T) {
	config := &config.DEXConfig{
		Name:       "test",
		ProgramID:  "11111111111111111111111111111112",
		Enabled:    true,
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}

	// 测试有效的交换请求验证
	validSwapReq := &types.SwapRequest{
		InputMint:  "So11111111111111111111111111111111111111112",
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountIn:   1000000000,
		Slippage:   0.005,
		UserWallet: "11111111111111111111111111111112",
	}

	// 由于validateSwapRequest是私有方法，我们通过创建具体适配器来测试
	raydiumConfig := &config.DEXConfig{
		Name:      "raydium",
		ProgramID: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
		Enabled:   true,
		Timeout:   30 * time.Second,
		RetryCount: 3,
	}

	raydiumAdapter, err := adapters.NewRaydiumAdapter(raydiumConfig)
	require.NoError(t, err)

	// 测试各种无效请求
	tests := []struct {
		name    string
		request *types.SwapRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			request: validSwapReq,
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
		{
			name: "empty input mint",
			request: &types.SwapRequest{
				InputMint:  "",
				OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				AmountIn:   1000000000,
				UserWallet: "11111111111111111111111111111112",
			},
			wantErr: true,
		},
		{
			name: "empty output mint",
			request: &types.SwapRequest{
				InputMint:  "So11111111111111111111111111111111111111112",
				OutputMint: "",
				AmountIn:   1000000000,
				UserWallet: "11111111111111111111111111111112",
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			request: &types.SwapRequest{
				InputMint:  "So11111111111111111111111111111111111111112",
				OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				AmountIn:   0,
				UserWallet: "11111111111111111111111111111112",
			},
			wantErr: true,
		},
		{
			name: "empty user wallet",
			request: &types.SwapRequest{
				InputMint:  "So11111111111111111111111111111111111111112",
				OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				AmountIn:   1000000000,
				UserWallet: "",
			},
			wantErr: true,
		},
		{
			name: "invalid slippage",
			request: &types.SwapRequest{
				InputMint:  "So11111111111111111111111111111111111111112",
				OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				AmountIn:   1000000000,
				Slippage:   1.5, // 超过100%
				UserWallet: "11111111111111111111111111111112",
			},
			wantErr: true,
		},
		{
			name: "invalid input mint format",
			request: &types.SwapRequest{
				InputMint:  "invalid-mint-address",
				OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				AmountIn:   1000000000,
				UserWallet: "11111111111111111111111111111112",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := raydiumAdapter.ValidateRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestInvalidAdapterConfigs 测试无效的适配器配置
func TestInvalidAdapterConfigs(t *testing.T) {
	// 测试无效的程序ID
	invalidConfig := &config.DEXConfig{
		Name:      "invalid",
		ProgramID: "invalid-program-id",
		Enabled:   true,
	}

	_, err := adapters.NewRaydiumAdapter(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid program ID")

	_, err = adapters.NewPumpfunAdapter(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid program ID")

	_, err = adapters.NewPumpSwapAdapter(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid program ID")
}