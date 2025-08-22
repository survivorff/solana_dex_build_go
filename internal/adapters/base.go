package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	"github.com/gagliardetto/solana-go"
)

// BaseAdapter DEX适配器基础实现
type BaseAdapter struct {
	name   string
	config *config.DEXConfig
	client *http.Client
}

// NewBaseAdapter 创建基础适配器
func NewBaseAdapter(name string, cfg *config.DEXConfig) *BaseAdapter {
	return &BaseAdapter{
		name:   name,
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GetName 获取DEX名称
func (b *BaseAdapter) GetName() string {
	return b.name
}

// GetConfig 获取DEX配置
func (b *BaseAdapter) GetConfig() *config.DEXConfig {
	return b.config
}

// makeRequest 发起HTTP请求的通用方法
func (b *BaseAdapter) makeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var err error
	var reqBodyBytes []byte

	if body != nil {
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(reqBodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "solana-dex-service/1.0")

	// 重试机制
	var resp *http.Response
	for i := 0; i < b.config.RetryCount; i++ {
		resp, err = b.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < b.config.RetryCount-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("request failed after %d retries: %w", b.config.RetryCount, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// ValidateSwapRequest 验证交换请求
func (b *BaseAdapter) ValidateSwapRequest(req *types.SwapRequest) error {
	// Basic validation
	if req.InputMint == "" {
		return errors.New("input mint is required")
	}
	if req.OutputMint == "" {
		return errors.New("output mint is required")
	}
	if req.AmountIn <= 0 {
		return errors.New("amount must be positive")
	}
	if req.UserWallet == "" {
		return errors.New("user wallet is required")
	}

	return nil
}

// validateLiquidityRequest 验证流动性请求
func (b *BaseAdapter) validateLiquidityRequest(req *types.LiquidityRequest) error {
	if req == nil {
		return fmt.Errorf("liquidity request is nil")
	}

	if req.Operation != "add" && req.Operation != "remove" {
		return fmt.Errorf("operation must be 'add' or 'remove'")
	}

	if req.TokenAMint == "" {
		return fmt.Errorf("token A mint is required")
	}

	if req.TokenBMint == "" {
		return fmt.Errorf("token B mint is required")
	}

	if req.AmountA == 0 {
		return fmt.Errorf("amount A must be greater than 0")
	}

	if req.AmountB == 0 {
		return fmt.Errorf("amount B must be greater than 0")
	}

	if req.UserWallet == "" {
		return fmt.Errorf("user wallet is required")
	}

	if req.Slippage < 0 || req.Slippage > 1 {
		return fmt.Errorf("slippage must be between 0 and 1")
	}

	// 验证公钥格式
	if _, err := solana.PublicKeyFromBase58(req.TokenAMint); err != nil {
		return fmt.Errorf("invalid token A mint address: %w", err)
	}

	if _, err := solana.PublicKeyFromBase58(req.TokenBMint); err != nil {
		return fmt.Errorf("invalid token B mint address: %w", err)
	}

	if _, err := solana.PublicKeyFromBase58(req.UserWallet); err != nil {
		return fmt.Errorf("invalid user wallet address: %w", err)
	}

	return nil
}

// createInstruction 创建Solana指令的辅助方法
func (b *BaseAdapter) createInstruction(programID solana.PublicKey, accounts []solana.AccountMeta, data []byte) *types.InstructionData {
	return &types.InstructionData{
		ProgramID: programID,
		Accounts:  accounts,
		Data:      data,
	}
}

// calculateMinAmountOut 计算最小输出金额（考虑滑点）
func (b *BaseAdapter) calculateMinAmountOut(amountOut uint64, slippage float64) uint64 {
	if slippage <= 0 {
		return amountOut
	}
	return uint64(float64(amountOut) * (1.0 - slippage))
}

// AdapterRegistry DEX适配器注册表
type AdapterRegistry struct {
	adapters map[string]types.DEXAdapter
}

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]types.DEXAdapter),
	}
}

// Register 注册DEX适配器
func (r *AdapterRegistry) Register(name string, adapter types.DEXAdapter) {
	r.adapters[name] = adapter
}

// Get 获取DEX适配器
func (r *AdapterRegistry) Get(name string) (types.DEXAdapter, error) {
	adapter, exists := r.adapters[name]
	if !exists {
		return nil, fmt.Errorf("adapter not found: %s", name)
	}
	return adapter, nil
}

// List 列出所有注册的适配器
func (r *AdapterRegistry) List() []string {
	var names []string
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// GetAll 获取所有适配器
func (r *AdapterRegistry) GetAll() map[string]types.DEXAdapter {
	return r.adapters
}