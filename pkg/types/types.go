package types

import (
	"time"

	"github.com/gagliardetto/solana-go"
)

// SwapRequest 代币交换请求
type SwapRequest struct {
	ID          string    `json:"id"`
	DEXType     string    `json:"dex_type" binding:"required,oneof=pumpfun pumpswap raydium"`
	InputMint   string    `json:"input_mint" binding:"required"`
	OutputMint  string    `json:"output_mint" binding:"required"`
	AmountIn    uint64    `json:"amount_in" binding:"required,min=1"`
	Slippage    float64   `json:"slippage" binding:"min=0,max=1"`
	PriorityFee uint64    `json:"priority_fee"`
	UserWallet  string    `json:"user_wallet" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

// LiquidityRequest 流动性操作请求
type LiquidityRequest struct {
	ID           string    `json:"id"`
	DEXType      string    `json:"dex_type" binding:"required,oneof=pumpfun pumpswap raydium"`
	Operation    string    `json:"operation" binding:"required,oneof=add remove"`
	TokenAMint   string    `json:"token_a_mint" binding:"required"`
	TokenBMint   string    `json:"token_b_mint" binding:"required"`
	AmountA      uint64    `json:"amount_a" binding:"required,min=1"`
	AmountB      uint64    `json:"amount_b" binding:"required,min=1"`
	Slippage     float64   `json:"slippage" binding:"min=0,max=1"`
	PriorityFee  uint64    `json:"priority_fee"`
	UserWallet   string    `json:"user_wallet" binding:"required"`
	CreatedAt    time.Time `json:"created_at"`
}

// TransactionTestRequest 交易测试请求
type TransactionTestRequest struct {
	Transaction  string `json:"transaction" binding:"required"`
	SimulateOnly bool   `json:"simulate_only"`
	PrivateKey   string `json:"private_key" binding:"required"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	Success      bool   `json:"success"`
	Transaction  string `json:"transaction,omitempty"`
	EstimatedFee uint64 `json:"estimated_fee,omitempty"`
	Error        string `json:"error,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// TransactionTestResponse 交易测试响应
type TransactionTestResponse struct {
	Success   bool     `json:"success"`
	Signature string   `json:"signature,omitempty"`
	Logs      []string `json:"logs,omitempty"`
	Error     string   `json:"error,omitempty"`
	GasUsed   uint64   `json:"gas_used,omitempty"`
}

// DEXInfo DEX信息
type DEXInfo struct {
	Name          string            `json:"name"`
	ProgramID     string            `json:"program_id"`
	RouterAddress string            `json:"router_address"`
	Endpoints     map[string]string `json:"endpoints"`
	Enabled       bool              `json:"enabled"`
	Status        string            `json:"status"` // online, offline, maintenance
}

// PoolInfo 流动性池信息
type PoolInfo struct {
	Address     string  `json:"address"`
	TokenAMint  string  `json:"token_a_mint"`
	TokenBMint  string  `json:"token_b_mint"`
	TokenAName  string  `json:"token_a_name"`
	TokenBName  string  `json:"token_b_name"`
	ReserveA    uint64  `json:"reserve_a"`
	ReserveB    uint64  `json:"reserve_b"`
	Liquidity   uint64  `json:"liquidity"`
	FeeRate     float64 `json:"fee_rate"`
	TVL         float64 `json:"tvl"` // Total Value Locked in USD
	Volume24h   float64 `json:"volume_24h"`
	APR         float64 `json:"apr"`
}

// QuoteResponse 报价响应
type QuoteResponse struct {
	InputMint    string  `json:"input_mint"`
	OutputMint   string  `json:"output_mint"`
	AmountIn     uint64  `json:"amount_in"`
	AmountOut    uint64  `json:"amount_out"`
	MinAmountOut uint64  `json:"min_amount_out"`
	PriceImpact  float64 `json:"price_impact"`
	Fee          uint64  `json:"fee"`
	Route        []Route `json:"route,omitempty"`
}

// Route 交易路由
type Route struct {
	DEX        string `json:"dex"`
	PoolID     string `json:"pool_id"`
	InputMint  string `json:"input_mint"`
	OutputMint string `json:"output_mint"`
	AmountIn   uint64 `json:"amount_in"`
	AmountOut  uint64 `json:"amount_out"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TransactionResult 交易结果
type TransactionResult struct {
	RequestID       string    `json:"request_id"`
	TransactionData string    `json:"transaction_data"`
	Signature       string    `json:"signature,omitempty"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	GasUsed         uint64    `json:"gas_used,omitempty"`
	ExecutedAt      time.Time `json:"executed_at"`
}

// TokenInfo 代币信息
type TokenInfo struct {
	Mint     string `json:"mint"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logo_uri,omitempty"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	Address   string     `json:"address"`
	Balance   uint64     `json:"balance"`
	Tokens    []TokenBalance `json:"tokens,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TokenBalance 代币余额
type TokenBalance struct {
	Mint    string `json:"mint"`
	Balance uint64 `json:"balance"`
	Symbol  string `json:"symbol,omitempty"`
}

// InstructionData 指令数据
type InstructionData struct {
	ProgramID   solana.PublicKey   `json:"program_id"`
	Accounts    []solana.AccountMeta `json:"accounts"`
	Data        []byte             `json:"data"`
	Description string             `json:"description,omitempty"`
}

// TransactionBuilder 交易构建器接口
type TransactionBuilder interface {
	BuildSwapTransaction(req *SwapRequest) (*solana.Transaction, error)
	BuildLiquidityTransaction(req *LiquidityRequest) (*solana.Transaction, error)
	EstimateFee(tx *solana.Transaction) (uint64, error)
}

// DEXAdapter DEX适配器接口
type DEXAdapter interface {
	GetName() string
	GetQuote(inputMint, outputMint string, amountIn uint64) (*QuoteResponse, error)
	BuildSwapInstruction(req *SwapRequest) (*InstructionData, error)
	BuildLiquidityInstruction(req *LiquidityRequest) (*InstructionData, error)
	GetPools() ([]PoolInfo, error)
	ValidateRequest(req interface{}) error
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	GetConfig() interface{}
	UpdateConfig(config interface{}) error
	GetDEXConfig(name string) (*DEXInfo, error)
	UpdateDEXConfig(name string, config *DEXInfo) error
	ReloadConfig() error
}