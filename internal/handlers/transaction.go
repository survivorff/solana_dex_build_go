package handlers

import (
	"net/http"
	"strconv"

	"solana-dex-service/internal/services"
	"solana-dex-service/internal/types"

	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
type TransactionHandler struct {
	transactionService *services.TransactionService
}

// NewTransactionHandler 创建交易处理器
func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// EncodeSwap 编码交换交易
// @Summary 编码代币交换交易
// @Description 根据请求参数编码代币交换交易指令
// @Tags 交易编码
// @Accept json
// @Produce json
// @Param request body types.SwapRequest true "交换请求参数"
// @Success 200 {object} types.TransactionResponse "交易编码成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/encode/swap [post]
func (th *TransactionHandler) EncodeSwap(c *gin.Context) {
	var req types.SwapRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 调用服务层编码交易
	resp, err := th.transactionService.EncodeSwapTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to encode swap transaction",
			Details: err.Error(),
		})
		return
	}

	// 返回响应
	if resp.Success {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   resp.Error,
			Details: "Transaction encoding failed",
		})
	}
}

// EncodeLiquidity 编码流动性交易
// @Summary 编码流动性操作交易
// @Description 根据请求参数编码流动性添加/移除交易指令
// @Tags 交易编码
// @Accept json
// @Produce json
// @Param request body types.LiquidityRequest true "流动性请求参数"
// @Success 200 {object} types.TransactionResponse "交易编码成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/encode/liquidity [post]
func (th *TransactionHandler) EncodeLiquidity(c *gin.Context) {
	var req types.LiquidityRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 调用服务层编码交易
	resp, err := th.transactionService.EncodeLiquidityTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to encode liquidity transaction",
			Details: err.Error(),
		})
		return
	}

	// 返回响应
	if resp.Success {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   resp.Error,
			Details: "Transaction encoding failed",
		})
	}
}

// TestTransaction 测试交易上链
// @Summary 测试交易上链
// @Description 签名并发送交易到Solana网络进行测试
// @Tags 交易测试
// @Accept json
// @Produce json
// @Param request body types.TransactionTestRequest true "交易测试请求参数"
// @Success 200 {object} types.TransactionTestResponse "交易测试成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/test/transaction [post]
func (th *TransactionHandler) TestTransaction(c *gin.Context) {
	var req types.TransactionTestRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 调用服务层测试交易
	resp, err := th.transactionService.TestTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to test transaction",
			Details: err.Error(),
		})
		return
	}

	// 返回响应
	if resp.Success {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   resp.Error,
			Details: "Transaction test failed",
		})
	}
}

// SimulateTransaction 模拟交易执行
// @Summary 模拟交易执行
// @Description 模拟交易执行，不实际发送到网络
// @Tags 交易测试
// @Accept json
// @Produce json
// @Param request body types.TransactionTestRequest true "交易模拟请求参数"
// @Success 200 {object} types.TransactionTestResponse "交易模拟成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/test/simulate [post]
func (th *TransactionHandler) SimulateTransaction(c *gin.Context) {
	var req types.TransactionTestRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 强制设置为仅模拟
	req.SimulateOnly = true

	// 调用服务层模拟交易
	resp, err := th.transactionService.SimulateTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to simulate transaction",
			Details: err.Error(),
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, resp)
}

// GetQuote 获取交易报价
// @Summary 获取交易报价
// @Description 获取指定DEX上的代币交换报价
// @Tags 交易查询
// @Accept json
// @Produce json
// @Param dex query string true "DEX名称"
// @Param inputMint query string true "输入代币地址"
// @Param outputMint query string true "输出代币地址"
// @Param amountIn query string true "输入金额"
// @Success 200 {object} types.QuoteResponse "报价获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/quote [get]
func (th *TransactionHandler) GetQuote(c *gin.Context) {
	// 获取查询参数
	dexName := c.Query("dex")
	inputMint := c.Query("inputMint")
	outputMint := c.Query("outputMint")
	amountInStr := c.Query("amountIn")

	// 验证必需参数
	if dexName == "" || inputMint == "" || outputMint == "" || amountInStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Missing required parameters",
			Details: "dex, inputMint, outputMint, and amountIn are required",
		})
		return
	}

	// 解析金额
	amountIn, err := strconv.ParseUint(amountInStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid amountIn parameter",
			Details: err.Error(),
		})
		return
	}

	// 获取DEX适配器
	adapter, err := th.transactionService.GetDEXAdapter(dexName)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX not found",
			Details: err.Error(),
		})
		return
	}

	// 获取报价
	quote, err := adapter.GetQuote(inputMint, outputMint, amountIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to get quote",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, quote)
}

// GetSupportedDEXes 获取支持的DEX列表
// @Summary 获取支持的DEX列表
// @Description 获取当前支持的所有DEX名称列表
// @Tags 交易查询
// @Produce json
// @Success 200 {object} types.SuccessResponse "DEX列表获取成功"
// @Router /api/v1/dexes [get]
func (th *TransactionHandler) GetSupportedDEXes(c *gin.Context) {
	dexes := th.transactionService.GetSupportedDEXes()
	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    dexes,
		Message: "Supported DEXes retrieved successfully",
	})
}

// EstimateFee 估算交易费用
// @Summary 估算交易费用
// @Description 估算给定交易的执行费用
// @Tags 交易查询
// @Accept json
// @Produce json
// @Param transaction body string true "Base64编码的交易数据"
// @Success 200 {object} map[string]interface{} "费用估算成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/estimate-fee [post]
func (th *TransactionHandler) EstimateFee(c *gin.Context) {
	var req struct {
		Transaction string `json:"transaction" binding:"required"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 这里可以添加费用估算逻辑
	// 目前返回一个默认的估算值
	c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"estimated_fee": 5000, // 默认5000 lamports
		"message":       "Fee estimation completed",
	})
}