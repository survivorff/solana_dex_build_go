package adapters

import (
	"context"
	"encoding/binary"
	"fmt"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	"github.com/gagliardetto/solana-go"
)

// RaydiumAdapter Raydium DEX适配器
type RaydiumAdapter struct {
	*BaseAdapter
	programID solana.PublicKey
}

// NewRaydiumAdapter 创建Raydium适配器
func NewRaydiumAdapter(cfg *config.DEXConfig) (*RaydiumAdapter, error) {
	programID, err := solana.PublicKeyFromBase58(cfg.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	return &RaydiumAdapter{
		BaseAdapter: NewBaseAdapter(cfg.Name, cfg),
		programID:   programID,
	}, nil
}

// GetQuote 获取交易报价
func (r *RaydiumAdapter) GetQuote(inputMint, outputMint string, amountIn uint64) (*types.QuoteResponse, error) {
	ctx := context.Background()
	
	// 构建请求URL
	quoteURL := r.config.Endpoints["quote"]
	if quoteURL == "" {
		return nil, fmt.Errorf("quote endpoint not configured")
	}

	// 构建请求参数
	reqParams := map[string]interface{}{
		"inputMint":  inputMint,
		"outputMint": outputMint,
		"amount":     amountIn,
		"slippageBps": 50, // 默认0.5%滑点
	}

	// 发起请求
	var quoteResp struct {
		Data struct {
			InputMint    string  `json:"inputMint"`
			OutputMint   string  `json:"outputMint"`
			InAmount     string  `json:"inAmount"`
			OutAmount    string  `json:"outAmount"`
			MinOutAmount string  `json:"minOutAmount"`
			PriceImpact  float64 `json:"priceImpactPct"`
			Fee          string  `json:"fee"`
		} `json:"data"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := r.makeRequest(ctx, "GET", quoteURL, reqParams, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	if !quoteResp.Success {
		return nil, fmt.Errorf("quote request failed: %s", quoteResp.Message)
	}

	// 解析响应
	var amountOut, minAmountOut, fee uint64
	fmt.Sscanf(quoteResp.Data.OutAmount, "%d", &amountOut)
	fmt.Sscanf(quoteResp.Data.MinOutAmount, "%d", &minAmountOut)
	fmt.Sscanf(quoteResp.Data.Fee, "%d", &fee)

	return &types.QuoteResponse{
		InputMint:    inputMint,
		OutputMint:   outputMint,
		AmountIn:     amountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  quoteResp.Data.PriceImpact,
		Fee:          fee,
	}, nil
}

// BuildSwapInstruction 构建交换指令
func (r *RaydiumAdapter) BuildSwapInstruction(req *types.SwapRequest) (*types.InstructionData, error) {
	if err := r.ValidateSwapRequest(req); err != nil {
		return nil, err
	}

	// 解析公钥
	userWallet, err := solana.PublicKeyFromBase58(req.UserWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid user wallet: %w", err)
	}

	inputMint, err := solana.PublicKeyFromBase58(req.InputMint)
	if err != nil {
		return nil, fmt.Errorf("invalid input mint: %w", err)
	}

	outputMint, err := solana.PublicKeyFromBase58(req.OutputMint)
	if err != nil {
		return nil, fmt.Errorf("invalid output mint: %w", err)
	}

	// 查找或创建关联代币账户
	userInputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, inputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find input token account: %w", err)
	}

	userOutputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find output token account: %w", err)
	}

	// 构建交换指令数据
	instructionData := r.buildSwapInstructionData(req)

	// 构建账户列表
	accounts := []solana.AccountMeta{
		{PublicKey: userWallet, IsSigner: true, IsWritable: false},
		{PublicKey: userInputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userOutputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: inputMint, IsSigner: false, IsWritable: false},
		{PublicKey: outputMint, IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return r.createInstruction(r.programID, accounts, instructionData), nil
}

// BuildLiquidityInstruction 构建流动性指令
func (r *RaydiumAdapter) BuildLiquidityInstruction(req *types.LiquidityRequest) (*types.InstructionData, error) {
	if err := r.validateLiquidityRequest(req); err != nil {
		return nil, err
	}

	// 解析公钥
	userWallet, err := solana.PublicKeyFromBase58(req.UserWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid user wallet: %w", err)
	}

	tokenAMint, err := solana.PublicKeyFromBase58(req.TokenAMint)
	if err != nil {
		return nil, fmt.Errorf("invalid token A mint: %w", err)
	}

	tokenBMint, err := solana.PublicKeyFromBase58(req.TokenBMint)
	if err != nil {
		return nil, fmt.Errorf("invalid token B mint: %w", err)
	}

	// 查找关联代币账户
	userTokenAAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, tokenAMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find token A account: %w", err)
	}

	userTokenBAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, tokenBMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find token B account: %w", err)
	}

	// 构建流动性指令数据
	instructionData := r.buildLiquidityInstructionData(req)

	// 构建账户列表
	accounts := []solana.AccountMeta{
		{PublicKey: userWallet, IsSigner: true, IsWritable: false},
		{PublicKey: userTokenAAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userTokenBAccount, IsSigner: false, IsWritable: true},
		{PublicKey: tokenAMint, IsSigner: false, IsWritable: false},
		{PublicKey: tokenBMint, IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return r.createInstruction(r.programID, accounts, instructionData), nil
}

// GetPools 获取流动性池信息
func (r *RaydiumAdapter) GetPools() ([]types.PoolInfo, error) {
	ctx := context.Background()
	
	poolsURL := r.config.Endpoints["pools"]
	if poolsURL == "" {
		return nil, fmt.Errorf("pools endpoint not configured")
	}

	var poolsResp struct {
		Data []struct {
			ID          string  `json:"id"`
			BaseMint    string  `json:"baseMint"`
			QuoteMint   string  `json:"quoteMint"`
			BaseSymbol  string  `json:"baseSymbol"`
			QuoteSymbol string  `json:"quoteSymbol"`
			Liquidity   string  `json:"liquidity"`
			Volume24h   float64 `json:"volume24h"`
			FeeRate     float64 `json:"feeRate"`
			TVL         float64 `json:"tvl"`
			APR         float64 `json:"apr"`
		} `json:"data"`
		Success bool `json:"success"`
	}

	if err := r.makeRequest(ctx, "GET", poolsURL, nil, &poolsResp); err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}

	if !poolsResp.Success {
		return nil, fmt.Errorf("pools request failed")
	}

	var pools []types.PoolInfo
	for _, pool := range poolsResp.Data {
		var liquidity uint64
		fmt.Sscanf(pool.Liquidity, "%d", &liquidity)

		pools = append(pools, types.PoolInfo{
			Address:    pool.ID,
			TokenAMint: pool.BaseMint,
			TokenBMint: pool.QuoteMint,
			TokenAName: pool.BaseSymbol,
			TokenBName: pool.QuoteSymbol,
			Liquidity:  liquidity,
			FeeRate:    pool.FeeRate,
			TVL:        pool.TVL,
			Volume24h:  pool.Volume24h,
			APR:        pool.APR,
		})
	}

	return pools, nil
}

// ValidateRequest 验证请求
func (r *RaydiumAdapter) ValidateRequest(req interface{}) error {
	switch v := req.(type) {
	case *types.SwapRequest:
		return r.ValidateSwapRequest(v)
	case *types.LiquidityRequest:
		return r.validateLiquidityRequest(v)
	default:
		return fmt.Errorf("unsupported request type")
	}
}

// buildSwapInstructionData 构建交换指令数据
func (r *RaydiumAdapter) buildSwapInstructionData(req *types.SwapRequest) []byte {
	// Raydium交换指令格式
	// [指令ID: 1字节] [金额: 8字节] [最小输出金额: 8字节]
	data := make([]byte, 17)
	
	// 指令ID (假设交换指令ID为9)
	data[0] = 9
	
	// 输入金额
	binary.LittleEndian.PutUint64(data[1:9], req.AmountIn)
	
	// 最小输出金额（考虑滑点）
	minAmountOut := r.calculateMinAmountOut(req.AmountIn, req.Slippage)
	binary.LittleEndian.PutUint64(data[9:17], minAmountOut)
	
	return data
}

// buildLiquidityInstructionData 构建流动性指令数据
func (r *RaydiumAdapter) buildLiquidityInstructionData(req *types.LiquidityRequest) []byte {
	// Raydium流动性指令格式
	// [指令ID: 1字节] [操作类型: 1字节] [金额A: 8字节] [金额B: 8字节]
	data := make([]byte, 18)
	
	// 指令ID (假设流动性指令ID为10)
	data[0] = 10
	
	// 操作类型 (0: 添加, 1: 移除)
	if req.Operation == "add" {
		data[1] = 0
	} else {
		data[1] = 1
	}
	
	// 代币A金额
	binary.LittleEndian.PutUint64(data[2:10], req.AmountA)
	
	// 代币B金额
	binary.LittleEndian.PutUint64(data[10:18], req.AmountB)
	
	return data
}