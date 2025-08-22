package services

import (
	"fmt"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"
)

// DEXService DEX服务
type DEXService struct {
	config            *config.Config
	transactionService *TransactionService
}

// NewDEXService 创建DEX服务
func NewDEXService(cfg *config.Config) *DEXService {
	return &DEXService{
		config: cfg,
	}
}

// SetTransactionService 设置交易服务（避免循环依赖）
func (ds *DEXService) SetTransactionService(ts *TransactionService) {
	ds.transactionService = ts
}

// ListDEXes 获取所有DEX列表
func (ds *DEXService) ListDEXes() ([]types.DEXInfo, error) {
	var dexes []types.DEXInfo

	for _, dexCfg := range ds.config.DEXes {
		status := "offline"
		if dexCfg.Enabled {
			status = "online"
		}

		dexes = append(dexes, types.DEXInfo{
			Name:          dexCfg.Name,
			ProgramID:     dexCfg.ProgramID,
			RouterAddress: dexCfg.RouterAddress,
			Endpoints:     dexCfg.Endpoints,
			Enabled:       dexCfg.Enabled,
			Status:        status,
		})
	}

	return dexes, nil
}

// GetDEX 获取指定DEX信息
func (ds *DEXService) GetDEX(name string) (*types.DEXInfo, error) {
	dexCfg, err := ds.config.GetDEXConfig(name)
	if err != nil {
		return nil, err
	}

	status := "offline"
	if dexCfg.Enabled {
		status = "online"
	}

	return &types.DEXInfo{
		Name:          dexCfg.Name,
		ProgramID:     dexCfg.ProgramID,
		RouterAddress: dexCfg.RouterAddress,
		Endpoints:     dexCfg.Endpoints,
		Enabled:       dexCfg.Enabled,
		Status:        status,
	}, nil
}

// GetPools 获取指定DEX的流动性池信息
func (ds *DEXService) GetPools(dexName string) ([]types.PoolInfo, error) {
	if ds.transactionService == nil {
		return nil, fmt.Errorf("transaction service not initialized")
	}

	// 获取DEX适配器
	adapter, err := ds.transactionService.GetDEXAdapter(dexName)
	if err != nil {
		return nil, fmt.Errorf("failed to get DEX adapter: %w", err)
	}

	// 获取流动性池信息
	pools, err := adapter.GetPools()
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}

	return pools, nil
}

// GetQuote 获取交易报价
func (ds *DEXService) GetQuote(dexName, inputMint, outputMint string, amountIn uint64) (*types.QuoteResponse, error) {
	if ds.transactionService == nil {
		return nil, fmt.Errorf("transaction service not initialized")
	}

	// 获取DEX适配器
	adapter, err := ds.transactionService.GetDEXAdapter(dexName)
	if err != nil {
		return nil, fmt.Errorf("failed to get DEX adapter: %w", err)
	}

	// 获取报价
	quote, err := adapter.GetQuote(inputMint, outputMint, amountIn)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	return quote, nil
}

// ValidateDEXRequest 验证DEX请求
func (ds *DEXService) ValidateDEXRequest(dexName string, req interface{}) error {
	if ds.transactionService == nil {
		return fmt.Errorf("transaction service not initialized")
	}

	// 获取DEX适配器
	adapter, err := ds.transactionService.GetDEXAdapter(dexName)
	if err != nil {
		return fmt.Errorf("failed to get DEX adapter: %w", err)
	}

	// 验证请求
	return adapter.ValidateRequest(req)
}

// GetEnabledDEXes 获取所有启用的DEX
func (ds *DEXService) GetEnabledDEXes() ([]types.DEXInfo, error) {
	allDEXes, err := ds.ListDEXes()
	if err != nil {
		return nil, err
	}

	var enabledDEXes []types.DEXInfo
	for _, dex := range allDEXes {
		if dex.Enabled {
			enabledDEXes = append(enabledDEXes, dex)
		}
	}

	return enabledDEXes, nil
}

// CheckDEXStatus 检查DEX状态
func (ds *DEXService) CheckDEXStatus(dexName string) (string, error) {
	dexInfo, err := ds.GetDEX(dexName)
	if err != nil {
		return "unknown", err
	}

	if !dexInfo.Enabled {
		return "disabled", nil
	}

	// 这里可以添加更复杂的健康检查逻辑
	// 比如检查DEX的API端点是否可访问
	return "online", nil
}