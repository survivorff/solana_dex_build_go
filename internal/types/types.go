package types

import (
	"time"

	"github.com/gagliardetto/solana-go"
	"solana-dex-service/internal/config"
)

// SwapRequest 交换请求结构
type SwapRequest struct {
	InputMint    string    `json:"input_mint"`    // 输入代币地址
	OutputMint   string    `json:"output_mint"`   // 输出代币地址
	AmountIn     uint64    `json:"amount_in"`     // 输入金额
	UserWallet   string    `json:"user_wallet"`   // 用户钱包地址
	Slippage     float64   `json:"slippage"`      // 滑点容忍度
	PriorityFee  uint64    `json:"priority_fee"`  // 优先费用
	DEXType      string    `json:"dex_type"`      // DEX类型
	ID           string    `json:"id"`            // 请求ID
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
}

// LiquidityRequest 流动性请求结构
type LiquidityRequest struct {
	TokenAMint   string    `json:"token_a_mint"`   // 代币A地址
	TokenBMint   string    `json:"token_b_mint"`   // 代币B地址
	AmountA      uint64    `json:"amount_a"`       // 代币A数量
	AmountB      uint64    `json:"amount_b"`       // 代币B数量
	UserWallet   string    `json:"user_wallet"`    // 用户钱包地址
	Slippage     float64   `json:"slippage"`       // 滑点容忍度
	PriorityFee  uint64    `json:"priority_fee"`   // 优先费用
	Operation    string    `json:"operation"`      // 操作类型: "add" 或 "remove"
	DEXType      string    `json:"dex_type"`       // DEX类型
	ID           string    `json:"id"`             // 请求ID
	CreatedAt    time.Time `json:"created_at"`     // 创建时间
}

// TransactionRequest 交易请求结构
type TransactionRequest struct {
	Transaction string `json:"transaction"` // Base64编码的交易数据
	SendTx      bool   `json:"send_tx"`    // 是否发送交易到链上
}

// QuoteRequest 报价请求结构
type QuoteRequest struct {
	InputMint  string `json:"input_mint"`  // 输入代币地址
	OutputMint string `json:"output_mint"` // 输出代币地址
	Amount     uint64 `json:"amount"`      // 输入金额
}

// PoolInfo 池子信息结构
type PoolInfo struct {
	Address     string  `json:"address"`       // 池子地址
	TokenAMint  string  `json:"token_a_mint"`  // 代币A地址
	TokenBMint  string  `json:"token_b_mint"`  // 代币B地址
	TokenAName  string  `json:"token_a_name"`  // 代币A名称
	TokenBName  string  `json:"token_b_name"`  // 代币B名称
	ReserveA    uint64  `json:"reserve_a"`     // 代币A储备量
	ReserveB    uint64  `json:"reserve_b"`     // 代币B储备量
	Liquidity   uint64  `json:"liquidity"`     // 流动性
	FeeRate     float64 `json:"fee_rate"`      // 手续费率
	TVL         float64 `json:"tvl"`           // 总锁定价值
	Volume24h   float64 `json:"volume_24h"`    // 24小时交易量
	APR         float64 `json:"apr"`           // 年化收益率
}

// DEXInfo DEX信息结构
type DEXInfo struct {
	Name          string            `json:"name"`           // DEX名称
	ProgramID     string            `json:"program_id"`     // 程序ID
	RouterAddress string            `json:"router_address"` // 路由地址
	Endpoints     map[string]string `json:"endpoints"`      // 端点配置
	Enabled       bool              `json:"enabled"`        // 是否启用
	Status        string            `json:"status"`         // 状态
	Description   string            `json:"description"`    // 描述
}

// TransactionResult 交易结果结构
type TransactionResult struct {
	Transaction string `json:"transaction"` // Base64编码的交易
	Signature   string `json:"signature"`  // 交易签名(如果发送)
	Success     bool   `json:"success"`    // 是否成功
	Error       string `json:"error"`      // 错误信息
}

// QuoteResult 报价结果结构
type QuoteResult struct {
	InputAmount  uint64  `json:"input_amount"`  // 输入金额
	OutputAmount uint64  `json:"output_amount"` // 输出金额
	PriceImpact  float64 `json:"price_impact"`  // 价格影响
	Fee          uint64  `json:"fee"`           // 手续费
	Route        []Route `json:"route"`        // 路由路径
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`   // 错误信息
	Code    int    `json:"code"`    // 错误代码
	Details string `json:"details"` // 详细信息
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Success bool        `json:"success"` // 是否成功
	Data    interface{} `json:"data"`    // 响应数据
	Message string      `json:"message"` // 消息
}

// InstructionData Solana指令数据结构
type InstructionData struct {
	ProgramID solana.PublicKey     `json:"program_id"` // 程序ID
	Accounts  []solana.AccountMeta `json:"accounts"`   // 账户列表
	Data      []byte               `json:"data"`       // 指令数据
}

// TransactionResponse 交易响应结构
type TransactionResponse struct {
	Success      bool   `json:"success"`       // 是否成功
	Transaction  string `json:"transaction"`   // Base64编码的交易
	EstimatedFee uint64 `json:"estimated_fee"` // 估算费用
	RequestID    string `json:"request_id"`    // 请求ID
	Error        string `json:"error"`         // 错误信息
}

// TransactionTestRequest 交易测试请求结构
type TransactionTestRequest struct {
	Transaction   string `json:"transaction"`    // Base64编码的交易数据
	PrivateKey    string `json:"private_key"`    // 私钥
	SimulateOnly  bool   `json:"simulate_only"`  // 是否仅模拟
}

// TransactionTestResponse 交易测试响应结构
type TransactionTestResponse struct {
	Success   bool     `json:"success"`   // 是否成功
	Signature string   `json:"signature"` // 交易签名
	Logs      []string `json:"logs"`      // 日志
	GasUsed   uint64   `json:"gas_used"`  // 消耗的Gas
	Error     string   `json:"error"`     // 错误信息
}

// QuoteResponse 报价响应结构
type QuoteResponse struct {
	InputMint    string  `json:"input_mint"`    // 输入代币地址
	OutputMint   string  `json:"output_mint"`   // 输出代币地址
	AmountIn     uint64  `json:"amount_in"`     // 输入金额
	AmountOut    uint64  `json:"amount_out"`    // 输出金额
	MinAmountOut uint64  `json:"min_amount_out"` // 最小输出金额
	PriceImpact  float64 `json:"price_impact"`  // 价格影响
	Fee          uint64  `json:"fee"`           // 手续费
	Route        []Route `json:"route"`        // 路由路径
}

// Route 路由信息结构
type Route struct {
	InputMint  string  `json:"input_mint"`  // 输入代币
	OutputMint string  `json:"output_mint"` // 输出代币
	PoolID     string  `json:"pool_id"`     // 池子ID
	FeeRate    float64 `json:"fee_rate"`    // 手续费率
	DEX        string  `json:"dex"`         // DEX名称
	AmountIn   uint64  `json:"amount_in"`   // 输入金额
	AmountOut  uint64  `json:"amount_out"`  // 输出金额
}

// DEXAdapter DEX适配器接口
type DEXAdapter interface {
	GetName() string
	GetConfig() *config.DEXConfig
	ValidateRequest(interface{}) error
	BuildSwapInstruction(*SwapRequest) (*InstructionData, error)
	BuildLiquidityInstruction(*LiquidityRequest) (*InstructionData, error)
	GetQuote(inputMint, outputMint string, amountIn uint64) (*QuoteResponse, error)
	GetPools() ([]PoolInfo, error)
}