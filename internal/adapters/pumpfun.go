package adapters

import (
	"context"
	"encoding/binary"
	"fmt"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	"github.com/gagliardetto/solana-go"
)

// PumpfunAdapter Pumpfun DEX适配器
type PumpfunAdapter struct {
	*BaseAdapter
	programID solana.PublicKey
}

// NewPumpfunAdapter 创建Pumpfun适配器
func NewPumpfunAdapter(cfg *config.DEXConfig) (*PumpfunAdapter, error) {
	programID, err := solana.PublicKeyFromBase58(cfg.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	return &PumpfunAdapter{
		BaseAdapter: NewBaseAdapter(cfg.Name, cfg),
		programID:   programID,
	}, nil
}

// GetQuote 获取交易报价
func (p *PumpfunAdapter) GetQuote(inputMint, outputMint string, amountIn uint64) (*types.QuoteResponse, error) {
	ctx := context.Background()
	
	// 构建请求URL
	quoteURL := fmt.Sprintf("%s/quote", p.config.Endpoints["api"])
	if p.config.Endpoints["quote"] != "" {
		quoteURL = p.config.Endpoints["quote"]
	}

	// 构建请求参数
	reqParams := map[string]interface{}{
		"inputMint":  inputMint,
		"outputMint": outputMint,
		"amount":     amountIn,
		"slippage":   0.005, // 默认0.5%滑点
	}

	// 发起请求
	var quoteResp struct {
		Success bool `json:"success"`
		Data    struct {
			InputMint     string  `json:"inputMint"`
			OutputMint    string  `json:"outputMint"`
			AmountIn      uint64  `json:"amountIn"`
			AmountOut     uint64  `json:"amountOut"`
			MinAmountOut  uint64  `json:"minAmountOut"`
			PriceImpact   float64 `json:"priceImpact"`
			Fee           uint64  `json:"fee"`
			MarketCap     float64 `json:"marketCap"`
			Liquidity     float64 `json:"liquidity"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := p.makeRequest(ctx, "POST", quoteURL, reqParams, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	if !quoteResp.Success {
		return nil, fmt.Errorf("quote request failed: %s", quoteResp.Error)
	}

	return &types.QuoteResponse{
		InputMint:    inputMint,
		OutputMint:   outputMint,
		AmountIn:     amountIn,
		AmountOut:    quoteResp.Data.AmountOut,
		MinAmountOut: quoteResp.Data.MinAmountOut,
		PriceImpact:  quoteResp.Data.PriceImpact,
		Fee:          quoteResp.Data.Fee,
	}, nil
}

// BuildSwapInstruction 构建交换指令
func (p *PumpfunAdapter) BuildSwapInstruction(req *types.SwapRequest) (*types.InstructionData, error) {
	if err := p.ValidateSwapRequest(req); err != nil {
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

	// Pumpfun特有的账户推导
	bondingCurve, err := p.deriveBondingCurveAddress(outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve: %w", err)
	}

	bondingCurveTokenAccount, err := p.deriveBondingCurveTokenAccount(bondingCurve, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve token account: %w", err)
	}

	// 查找用户关联代币账户
	userInputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, inputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find input token account: %w", err)
	}

	userOutputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find output token account: %w", err)
	}

	// 构建交换指令数据
	instructionData := p.buildSwapInstructionData(req)

	// 构建账户列表（Pumpfun特有的账户结构）
	accounts := []solana.AccountMeta{
		{PublicKey: userWallet, IsSigner: true, IsWritable: false},
		{PublicKey: userInputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userOutputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: bondingCurve, IsSigner: false, IsWritable: true},
		{PublicKey: bondingCurveTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: inputMint, IsSigner: false, IsWritable: false},
		{PublicKey: outputMint, IsSigner: false, IsWritable: true},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
	}

	return p.createInstruction(p.programID, accounts, instructionData), nil
}

// BuildLiquidityInstruction 构建流动性指令
func (p *PumpfunAdapter) BuildLiquidityInstruction(req *types.LiquidityRequest) (*types.InstructionData, error) {
	// Pumpfun主要是bonding curve模式，不支持传统的流动性池
	return nil, fmt.Errorf("pumpfun does not support traditional liquidity operations")
}

// GetPools 获取流动性池信息
func (p *PumpfunAdapter) GetPools() ([]types.PoolInfo, error) {
	ctx := context.Background()
	
	poolsURL := fmt.Sprintf("%s/tokens", p.config.Endpoints["api"])

	var tokensResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Mint        string  `json:"mint"`
			Name        string  `json:"name"`
			Symbol      string  `json:"symbol"`
			MarketCap   float64 `json:"market_cap"`
			Liquidity   float64 `json:"liquidity"`
			Volume24h   float64 `json:"volume_24h"`
			PriceChange float64 `json:"price_change_24h"`
			CreatedAt   string  `json:"created_at"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := p.makeRequest(ctx, "GET", poolsURL, nil, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	if !tokensResp.Success {
		return nil, fmt.Errorf("tokens request failed: %s", tokensResp.Error)
	}

	var pools []types.PoolInfo
	for _, token := range tokensResp.Data {
		// Pumpfun的"池"实际上是bonding curve
		pools = append(pools, types.PoolInfo{
			Address:    token.Mint, // 使用mint地址作为池地址
			TokenAMint: "So11111111111111111111111111111111111111112", // SOL
			TokenBMint: token.Mint,
			TokenAName: "SOL",
			TokenBName: token.Symbol,
			Liquidity:  uint64(token.Liquidity),
			FeeRate:    0.01, // Pumpfun通常收取1%费用
			TVL:        token.MarketCap,
			Volume24h:  token.Volume24h,
			APR:        0, // Bonding curve没有APR概念
		})
	}

	return pools, nil
}

// ValidateRequest 验证请求
func (p *PumpfunAdapter) ValidateRequest(req interface{}) error {
	switch v := req.(type) {
	case *types.SwapRequest:
		return p.ValidateSwapRequest(v)
	case *types.LiquidityRequest:
		return fmt.Errorf("pumpfun does not support liquidity operations")
	default:
		return fmt.Errorf("unsupported request type")
	}
}

// buildSwapInstructionData 构建交换指令数据
func (p *PumpfunAdapter) buildSwapInstructionData(req *types.SwapRequest) []byte {
	// Pumpfun交换指令格式
	// [指令ID: 1字节] [金额: 8字节] [最小输出金额: 8字节] [方向: 1字节]
	data := make([]byte, 18)
	
	// 指令ID (假设买入指令ID为6，卖出指令ID为7)
	if req.InputMint == "So11111111111111111111111111111111111111112" {
		// SOL -> Token (买入)
		data[0] = 6
	} else {
		// Token -> SOL (卖出)
		data[0] = 7
	}
	
	// 输入金额
	binary.LittleEndian.PutUint64(data[1:9], req.AmountIn)
	
	// 最小输出金额（考虑滑点）
	minAmountOut := p.calculateMinAmountOut(req.AmountIn, req.Slippage)
	binary.LittleEndian.PutUint64(data[9:17], minAmountOut)
	
	// 交易方向 (0: 买入, 1: 卖出)
	if req.InputMint == "So11111111111111111111111111111111111111112" {
		data[17] = 0
	} else {
		data[17] = 1
	}
	
	return data
}

// deriveBondingCurveAddress 推导bonding curve地址
func (p *PumpfunAdapter) deriveBondingCurveAddress(mint solana.PublicKey) (solana.PublicKey, error) {
	// Pumpfun的bonding curve地址推导逻辑
	// 通常使用PDA (Program Derived Address)
	seeds := [][]byte{
		[]byte("bonding-curve"),
		mint.Bytes(),
	}
	
	address, _, err := solana.FindProgramAddress(seeds, p.programID)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to derive bonding curve address: %w", err)
	}
	
	return address, nil
}

// deriveBondingCurveTokenAccount 推导bonding curve代币账户地址
func (p *PumpfunAdapter) deriveBondingCurveTokenAccount(bondingCurve, mint solana.PublicKey) (solana.PublicKey, error) {
	// 查找bonding curve的关联代币账户
	address, _, err := solana.FindAssociatedTokenAddress(bondingCurve, mint)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to derive bonding curve token account: %w", err)
	}
	
	return address, nil
}