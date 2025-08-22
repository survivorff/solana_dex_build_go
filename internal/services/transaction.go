package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"solana-dex-service/internal/adapters"
	"solana-dex-service/internal/config"
	"solana-dex-service/internal/types"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
)

// TransactionService 交易服务
type TransactionService struct {
	config          *config.Config
	adapterRegistry *adapters.AdapterRegistry
	rpcClient       *rpc.Client
}

// NewTransactionService 创建交易服务
func NewTransactionService(cfg *config.Config) *TransactionService {
	// 创建RPC客户端
	rpcClient := rpc.New(cfg.Solana.RPCURL)

	// 创建适配器注册表
	adapterRegistry := adapters.NewAdapterRegistry()

	// 注册所有DEX适配器
	for _, dexCfg := range cfg.DEXes {
		if !dexCfg.Enabled {
			continue
		}

		switch dexCfg.Name {
		case "raydium":
			if adapter, err := adapters.NewRaydiumAdapter(&dexCfg); err == nil {
				adapterRegistry.Register(dexCfg.Name, adapter)
			}
		case "pumpfun":
			if adapter, err := adapters.NewPumpfunAdapter(&dexCfg); err == nil {
				adapterRegistry.Register(dexCfg.Name, adapter)
			}
		case "pumpswap":
			if adapter, err := adapters.NewPumpSwapAdapter(&dexCfg); err == nil {
				adapterRegistry.Register(dexCfg.Name, adapter)
			}
		}
	}

	return &TransactionService{
		config:          cfg,
		adapterRegistry: adapterRegistry,
		rpcClient:       rpcClient,
	}
}

// EncodeSwapTransaction 编码交换交易
func (ts *TransactionService) EncodeSwapTransaction(req *types.SwapRequest) (*types.TransactionResponse, error) {
	// 生成请求ID
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()

	// 获取对应的DEX适配器
	adapter, err := ts.adapterRegistry.Get(req.DEXType)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("DEX adapter not found: %s", req.DEXType),
		}, nil
	}

	// 验证请求
	if err := adapter.ValidateRequest(req); err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}, nil
	}

	// 构建交换指令
	instructionData, err := adapter.BuildSwapInstruction(req)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to build swap instruction: %v", err),
		}, nil
	}

	// 创建Solana指令
	accounts := make(solana.AccountMetaSlice, len(instructionData.Accounts))
	for i, acc := range instructionData.Accounts {
		accounts[i] = &acc
	}
	instruction := solana.NewInstruction(
		instructionData.ProgramID,
		accounts,
		instructionData.Data,
	)

	// 创建交易
	tx, err := ts.buildTransaction([]solana.Instruction{instruction}, req.UserWallet, req.PriorityFee)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to build transaction: %v", err),
		}, nil
	}

	// 估算费用
	estimatedFee, err := ts.estimateTransactionFee(tx)
	if err != nil {
		// 费用估算失败不影响交易构建，使用默认值
		estimatedFee = 5000 // 默认5000 lamports
	}

	// 序列化交易
	txData, err := tx.MarshalBinary()
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to serialize transaction: %v", err),
		}, nil
	}

	return &types.TransactionResponse{
		Success:      true,
		Transaction:  base64.StdEncoding.EncodeToString(txData),
		EstimatedFee: estimatedFee,
		RequestID:    req.ID,
	}, nil
}

// EncodeLiquidityTransaction 编码流动性交易
func (ts *TransactionService) EncodeLiquidityTransaction(req *types.LiquidityRequest) (*types.TransactionResponse, error) {
	// 生成请求ID
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()

	// 获取对应的DEX适配器
	adapter, err := ts.adapterRegistry.Get(req.DEXType)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("DEX adapter not found: %s", req.DEXType),
		}, nil
	}

	// 验证请求
	if err := adapter.ValidateRequest(req); err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}, nil
	}

	// 构建流动性指令
	instructionData, err := adapter.BuildLiquidityInstruction(req)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to build liquidity instruction: %v", err),
		}, nil
	}

	// 创建Solana指令
	accounts := make(solana.AccountMetaSlice, len(instructionData.Accounts))
	for i, acc := range instructionData.Accounts {
		accounts[i] = &acc
	}
	instruction := solana.NewInstruction(
		instructionData.ProgramID,
		accounts,
		instructionData.Data,
	)

	// 创建交易
	tx, err := ts.buildTransaction([]solana.Instruction{instruction}, req.UserWallet, req.PriorityFee)
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to build transaction: %v", err),
		}, nil
	}

	// 估算费用
	estimatedFee, err := ts.estimateTransactionFee(tx)
	if err != nil {
		estimatedFee = 5000 // 默认5000 lamports
	}

	// 序列化交易
	txData, err := tx.MarshalBinary()
	if err != nil {
		return &types.TransactionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to serialize transaction: %v", err),
		}, nil
	}

	return &types.TransactionResponse{
		Success:      true,
		Transaction:  base64.StdEncoding.EncodeToString(txData),
		EstimatedFee: estimatedFee,
		RequestID:    req.ID,
	}, nil
}

// TestTransaction 测试交易上链
func (ts *TransactionService) TestTransaction(req *types.TransactionTestRequest) (*types.TransactionTestResponse, error) {
	// 解码交易数据
	txData, err := base64.StdEncoding.DecodeString(req.Transaction)
	if err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode transaction: %v", err),
		}, nil
	}

	// 反序列化交易
	var tx solana.Transaction
	if err := bin.NewBorshDecoder(txData).Decode(&tx); err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to deserialize transaction: %v", err),
		}, nil
	}

	// 解析私钥
	privateKey, err := solana.PrivateKeyFromBase58(req.PrivateKey)
	if err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid private key: %v", err),
		}, nil
	}

	// 签名交易
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(privateKey.PublicKey()) {
			return &privateKey
		}
		return nil
	}); err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to sign transaction: %v", err),
		}, nil
	}

	ctx := context.Background()

	if req.SimulateOnly {
		// 仅模拟执行
		return ts.simulateTransaction(ctx, &tx)
	} else {
		// 实际发送交易
		return ts.sendTransaction(ctx, &tx)
	}
}

// SimulateTransaction 模拟交易执行
func (ts *TransactionService) SimulateTransaction(req *types.TransactionTestRequest) (*types.TransactionTestResponse, error) {
	// 强制设置为仅模拟
	req.SimulateOnly = true
	return ts.TestTransaction(req)
}

// buildTransaction 构建交易
func (ts *TransactionService) buildTransaction(instructions []solana.Instruction, payerAddress string, priorityFee uint64) (*solana.Transaction, error) {
	// 解析付款人地址
	payer, err := solana.PublicKeyFromBase58(payerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid payer address: %w", err)
	}

	// 获取最新的区块哈希
	ctx := context.Background()
	recentBlockhash, err := ts.rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// 如果设置了优先费用，添加优先费用指令
	if priorityFee > 0 {
		priorityFeeInstruction := ts.createPriorityFeeInstruction(priorityFee)
		instructions = append([]solana.Instruction{priorityFeeInstruction}, instructions...)
	}

	// 创建交易
	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhash.Value.Blockhash,
		solana.TransactionPayer(payer),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

// createPriorityFeeInstruction 创建优先费用指令
func (ts *TransactionService) createPriorityFeeInstruction(priorityFee uint64) solana.Instruction {
	// 这里应该使用Solana的优先费用程序
	// 简化实现，实际应该使用正确的程序ID和指令格式
	computeBudgetProgramID := solana.MustPublicKeyFromBase58("ComputeBudget111111111111111111111111111111")
	
	// 构建优先费用指令数据
	data := make([]byte, 9)
	data[0] = 3 // SetComputeUnitPrice指令ID
	// 将优先费用写入数据
	for i := 0; i < 8; i++ {
		data[i+1] = byte(priorityFee >> (i * 8))
	}

	return solana.NewInstruction(
		computeBudgetProgramID,
		solana.AccountMetaSlice{},
		data,
	)
}

// estimateTransactionFee 估算交易费用
func (ts *TransactionService) estimateTransactionFee(tx *solana.Transaction) (uint64, error) {
	ctx := context.Background()
	
	// 使用RPC客户端获取费用估算
	feeResponse, err := ts.rpcClient.GetFeeForMessage(
		ctx,
		tx.Message.ToBase64(),
		rpc.CommitmentProcessed,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get fee estimate: %w", err)
	}

	if feeResponse.Value == nil {
		return 0, fmt.Errorf("fee estimation returned null")
	}

	return *feeResponse.Value, nil
}

// simulateTransaction 模拟交易执行
func (ts *TransactionService) simulateTransaction(ctx context.Context, tx *solana.Transaction) (*types.TransactionTestResponse, error) {
	// 使用RPC客户端模拟交易
	simResult, err := ts.rpcClient.SimulateTransaction(ctx, tx)
	if err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Simulation failed: %v", err),
		}, nil
	}

	if simResult.Value.Err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Simulation error: %v", simResult.Value.Err),
			Logs:    simResult.Value.Logs,
		}, nil
	}

	return &types.TransactionTestResponse{
		Success: true,
		Logs:    simResult.Value.Logs,
		GasUsed: *simResult.Value.UnitsConsumed,
	}, nil
}

// sendTransaction 发送交易到链上
func (ts *TransactionService) sendTransaction(ctx context.Context, tx *solana.Transaction) (*types.TransactionTestResponse, error) {
	// 发送交易
	signature, err := ts.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return &types.TransactionTestResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to send transaction: %v", err),
		}, nil
	}

	// 简化处理，不等待确认

	// 获取交易详情
	txDetails, err := ts.rpcClient.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err == nil && txDetails != nil {
		return &types.TransactionTestResponse{
			Success:   true,
			Signature: signature.String(),
			Logs:      txDetails.Meta.LogMessages,
		}, nil
	}

	return &types.TransactionTestResponse{
		Success:   true,
		Signature: signature.String(),
	}, nil
}

// GetSupportedDEXes 获取支持的DEX列表
func (ts *TransactionService) GetSupportedDEXes() []string {
	return ts.adapterRegistry.List()
}

// GetDEXAdapter 获取DEX适配器
func (ts *TransactionService) GetDEXAdapter(name string) (types.DEXAdapter, error) {
	return ts.adapterRegistry.Get(name)
}