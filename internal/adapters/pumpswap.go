package adapters

import (
	"context"
	"encoding/binary"
	"fmt"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	"github.com/gagliardetto/solana-go"
)

// PumpSwapAdapter PumpSwap DEX适配器
type PumpSwapAdapter struct {
	*BaseAdapter
	programID solana.PublicKey
	routerAddress solana.PublicKey
}

// NewPumpSwapAdapter 创建PumpSwap适配器
func NewPumpSwapAdapter(cfg *config.DEXConfig) (*PumpSwapAdapter, error) {
	programID, err := solana.PublicKeyFromBase58(cfg.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}

	routerAddress, err := solana.PublicKeyFromBase58(cfg.RouterAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid router address: %w", err)
	}

	return &PumpSwapAdapter{
		BaseAdapter:   NewBaseAdapter(cfg.Name, cfg),
		programID:     programID,
		routerAddress: routerAddress,
	}, nil
}

// GetQuote 获取交易报价
func (ps *PumpSwapAdapter) GetQuote(inputMint, outputMint string, amountIn uint64) (*types.QuoteResponse, error) {
	ctx := context.Background()
	
	// 构建请求URL
	quoteURL := ps.config.Endpoints["quote"]
	if quoteURL == "" {
		quoteURL = fmt.Sprintf("%s/quote", ps.config.Endpoints["api"])
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
			Route         []struct {
				PoolID     string `json:"poolId"`
				InputMint  string `json:"inputMint"`
				OutputMint string `json:"outputMint"`
				AmountIn   uint64 `json:"amountIn"`
				AmountOut  uint64 `json:"amountOut"`
			} `json:"route"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := ps.makeRequest(ctx, "POST", quoteURL, reqParams, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	if !quoteResp.Success {
		return nil, fmt.Errorf("quote request failed: %s", quoteResp.Error)
	}

	// 转换路由信息
	var routes []types.Route
	for _, r := range quoteResp.Data.Route {
		routes = append(routes, types.Route{
			DEX:        "pumpswap",
			PoolID:     r.PoolID,
			InputMint:  r.InputMint,
			OutputMint: r.OutputMint,
			AmountIn:   r.AmountIn,
			AmountOut:  r.AmountOut,
		})
	}

	return &types.QuoteResponse{
		InputMint:    inputMint,
		OutputMint:   outputMint,
		AmountIn:     amountIn,
		AmountOut:    quoteResp.Data.AmountOut,
		MinAmountOut: quoteResp.Data.MinAmountOut,
		PriceImpact:  quoteResp.Data.PriceImpact,
		Fee:          quoteResp.Data.Fee,
		Route:        routes,
	}, nil
}

// BuildSwapInstruction 构建交换指令
func (ps *PumpSwapAdapter) BuildSwapInstruction(req *types.SwapRequest) (*types.InstructionData, error) {
	if err := ps.ValidateSwapRequest(req); err != nil {
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

	// 查找交易对池地址
	poolAddress, err := ps.findPoolAddress(inputMint, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find pool: %w", err)
	}

	// 查找池的代币账户
	poolInputTokenAccount, _, err := solana.FindAssociatedTokenAddress(poolAddress, inputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find pool input token account: %w", err)
	}

	poolOutputTokenAccount, _, err := solana.FindAssociatedTokenAddress(poolAddress, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find pool output token account: %w", err)
	}

	// 查找用户关联代币账户
	userInputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, inputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find user input token account: %w", err)
	}

	userOutputTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, outputMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find user output token account: %w", err)
	}

	// 构建交换指令数据
	instructionData := ps.buildSwapInstructionData(req)

	// 构建账户列表
	accounts := []solana.AccountMeta{
		{PublicKey: userWallet, IsSigner: true, IsWritable: false},
		{PublicKey: userInputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userOutputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: poolAddress, IsSigner: false, IsWritable: true},
		{PublicKey: poolInputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: poolOutputTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: ps.routerAddress, IsSigner: false, IsWritable: false},
		{PublicKey: inputMint, IsSigner: false, IsWritable: false},
		{PublicKey: outputMint, IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return ps.createInstruction(ps.programID, accounts, instructionData), nil
}

// BuildLiquidityInstruction 构建流动性指令
func (ps *PumpSwapAdapter) BuildLiquidityInstruction(req *types.LiquidityRequest) (*types.InstructionData, error) {
	if err := ps.validateLiquidityRequest(req); err != nil {
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

	// 查找或创建流动性池
	poolAddress, err := ps.findOrCreatePoolAddress(tokenAMint, tokenBMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find/create pool: %w", err)
	}

	// 查找LP代币mint
	lpMint, err := ps.deriveLPMintAddress(poolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to derive LP mint: %w", err)
	}

	// 查找用户关联代币账户
	userTokenAAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, tokenAMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find user token A account: %w", err)
	}

	userTokenBAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, tokenBMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find user token B account: %w", err)
	}

	userLPTokenAccount, _, err := solana.FindAssociatedTokenAddress(userWallet, lpMint)
	if err != nil {
		return nil, fmt.Errorf("failed to find user LP token account: %w", err)
	}

	// 构建流动性指令数据
	instructionData := ps.buildLiquidityInstructionData(req)

	// 构建账户列表
	accounts := []solana.AccountMeta{
		{PublicKey: userWallet, IsSigner: true, IsWritable: false},
		{PublicKey: userTokenAAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userTokenBAccount, IsSigner: false, IsWritable: true},
		{PublicKey: userLPTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: poolAddress, IsSigner: false, IsWritable: true},
		{PublicKey: lpMint, IsSigner: false, IsWritable: true},
		{PublicKey: tokenAMint, IsSigner: false, IsWritable: false},
		{PublicKey: tokenBMint, IsSigner: false, IsWritable: false},
		{PublicKey: solana.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return ps.createInstruction(ps.programID, accounts, instructionData), nil
}

// GetPools 获取流动性池信息
func (ps *PumpSwapAdapter) GetPools() ([]types.PoolInfo, error) {
	ctx := context.Background()
	
	poolsURL := fmt.Sprintf("%s/pools", ps.config.Endpoints["api"])

	var poolsResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Address     string  `json:"address"`
			TokenAMint  string  `json:"tokenAMint"`
			TokenBMint  string  `json:"tokenBMint"`
			TokenAName  string  `json:"tokenAName"`
			TokenBName  string  `json:"tokenBName"`
			ReserveA    uint64  `json:"reserveA"`
			ReserveB    uint64  `json:"reserveB"`
			Liquidity   uint64  `json:"liquidity"`
			FeeRate     float64 `json:"feeRate"`
			TVL         float64 `json:"tvl"`
			Volume24h   float64 `json:"volume24h"`
			APR         float64 `json:"apr"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := ps.makeRequest(ctx, "GET", poolsURL, nil, &poolsResp); err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}

	if !poolsResp.Success {
		return nil, fmt.Errorf("pools request failed: %s", poolsResp.Error)
	}

	var pools []types.PoolInfo
	for _, pool := range poolsResp.Data {
		pools = append(pools, types.PoolInfo{
			Address:    pool.Address,
			TokenAMint: pool.TokenAMint,
			TokenBMint: pool.TokenBMint,
			TokenAName: pool.TokenAName,
			TokenBName: pool.TokenBName,
			ReserveA:   pool.ReserveA,
			ReserveB:   pool.ReserveB,
			Liquidity:  pool.Liquidity,
			FeeRate:    pool.FeeRate,
			TVL:        pool.TVL,
			Volume24h:  pool.Volume24h,
			APR:        pool.APR,
		})
	}

	return pools, nil
}

// ValidateRequest 验证请求
func (ps *PumpSwapAdapter) ValidateRequest(req interface{}) error {
	switch v := req.(type) {
	case *types.SwapRequest:
		return ps.ValidateSwapRequest(v)
	case *types.LiquidityRequest:
		return ps.validateLiquidityRequest(v)
	default:
		return fmt.Errorf("unsupported request type")
	}
}

// buildSwapInstructionData 构建交换指令数据
func (ps *PumpSwapAdapter) buildSwapInstructionData(req *types.SwapRequest) []byte {
	// PumpSwap交换指令格式
	// [指令ID: 1字节] [金额: 8字节] [最小输出金额: 8字节]
	data := make([]byte, 17)
	
	// 指令ID (假设交换指令ID为1)
	data[0] = 1
	
	// 输入金额
	binary.LittleEndian.PutUint64(data[1:9], req.AmountIn)
	
	// 最小输出金额（考虑滑点）
	minAmountOut := ps.calculateMinAmountOut(req.AmountIn, req.Slippage)
	binary.LittleEndian.PutUint64(data[9:17], minAmountOut)
	
	return data
}

// buildLiquidityInstructionData 构建流动性指令数据
func (ps *PumpSwapAdapter) buildLiquidityInstructionData(req *types.LiquidityRequest) []byte {
	// PumpSwap流动性指令格式
	// [指令ID: 1字节] [操作类型: 1字节] [金额A: 8字节] [金额B: 8字节]
	data := make([]byte, 18)
	
	// 指令ID (假设添加流动性指令ID为2，移除流动性指令ID为3)
	if req.Operation == "add" {
		data[0] = 2
	} else {
		data[0] = 3
	}
	
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

// findPoolAddress 查找交易对池地址
func (ps *PumpSwapAdapter) findPoolAddress(tokenA, tokenB solana.PublicKey) (solana.PublicKey, error) {
	// PumpSwap的池地址推导逻辑
	// 通常使用两个代币的mint地址来推导池地址
	var seeds [][]byte
	if tokenA.String() < tokenB.String() {
		seeds = [][]byte{
			[]byte("pool"),
			tokenA.Bytes(),
			tokenB.Bytes(),
		}
	} else {
		seeds = [][]byte{
			[]byte("pool"),
			tokenB.Bytes(),
			tokenA.Bytes(),
		}
	}
	
	address, _, err := solana.FindProgramAddress(seeds, ps.programID)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to derive pool address: %w", err)
	}
	
	return address, nil
}

// findOrCreatePoolAddress 查找或创建池地址
func (ps *PumpSwapAdapter) findOrCreatePoolAddress(tokenA, tokenB solana.PublicKey) (solana.PublicKey, error) {
	// 首先尝试查找现有池
	poolAddress, err := ps.findPoolAddress(tokenA, tokenB)
	if err != nil {
		return solana.PublicKey{}, err
	}
	
	// 这里可以添加检查池是否存在的逻辑
	// 如果不存在，可能需要先创建池
	
	return poolAddress, nil
}

// deriveLPMintAddress 推导LP代币mint地址
func (ps *PumpSwapAdapter) deriveLPMintAddress(poolAddress solana.PublicKey) (solana.PublicKey, error) {
	seeds := [][]byte{
		[]byte("lp-mint"),
		poolAddress.Bytes(),
	}
	
	address, _, err := solana.FindProgramAddress(seeds, ps.programID)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to derive LP mint address: %w", err)
	}
	
	return address, nil
}