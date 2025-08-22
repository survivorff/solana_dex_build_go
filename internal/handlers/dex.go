package handlers

import (
	"net/http"

	"solana-dex-service/internal/services"
	"solana-dex-service/internal/types"

	"github.com/gin-gonic/gin"
)

// DEXHandler DEX处理器
type DEXHandler struct {
	dexService *services.DEXService
}

// NewDEXHandler 创建DEX处理器
func NewDEXHandler(dexService *services.DEXService) *DEXHandler {
	return &DEXHandler{
		dexService: dexService,
	}
}

// ListDEXes 获取所有DEX列表
// @Summary 获取所有DEX列表
// @Description 获取系统中配置的所有DEX信息
// @Tags DEX管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "DEX列表获取成功"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/list [get]
func (dh *DEXHandler) ListDEXes(c *gin.Context) {
	dexes, err := dh.dexService.ListDEXes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to get DEX list",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    dexes,
		Message: "DEX list retrieved successfully",
	})
}

// GetDEX 获取指定DEX信息
// @Summary 获取指定DEX信息
// @Description 根据DEX名称获取详细信息
// @Tags DEX管理
// @Produce json
// @Param name path string true "DEX名称"
// @Success 200 {object} types.SuccessResponse "DEX信息获取成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name} [get]
func (dh *DEXHandler) GetDEX(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	dex, err := dh.dexService.GetDEX(dexName)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    dex,
		Message: "DEX information retrieved successfully",
	})
}

// GetPools 获取指定DEX的流动性池信息
// @Summary 获取DEX流动性池信息
// @Description 获取指定DEX的所有流动性池信息
// @Tags DEX管理
// @Produce json
// @Param name path string true "DEX名称"
// @Param limit query int false "返回数量限制" default(50)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} types.SuccessResponse "流动性池信息获取成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name}/pools [get]
func (dh *DEXHandler) GetPools(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	// 获取查询参数
	limit := 50 // 默认限制
	offset := 0 // 默认偏移

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseIntParam(limitStr, "limit"); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := parseIntParam(offsetStr, "offset"); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取流动性池信息
	pools, err := dh.dexService.GetPools(dexName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to get pools",
			Details: err.Error(),
		})
		return
	}

	// 应用分页
	total := len(pools)
	start := offset
	end := offset + limit

	if start >= total {
		pools = []types.PoolInfo{}
	} else {
		if end > total {
			end = total
		}
		pools = pools[start:end]
	}

	response := map[string]interface{}{
		"pools":  pools,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Pools retrieved successfully",
	})
}

// GetEnabledDEXes 获取所有启用的DEX
// @Summary 获取启用的DEX列表
// @Description 获取所有当前启用的DEX信息
// @Tags DEX管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "启用DEX列表获取成功"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/enabled [get]
func (dh *DEXHandler) GetEnabledDEXes(c *gin.Context) {
	dexes, err := dh.dexService.GetEnabledDEXes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to get enabled DEXes",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    dexes,
		Message: "Enabled DEXes retrieved successfully",
	})
}

// CheckDEXStatus 检查DEX状态
// @Summary 检查DEX状态
// @Description 检查指定DEX的当前运行状态
// @Tags DEX管理
// @Produce json
// @Param name path string true "DEX名称"
// @Success 200 {object} types.SuccessResponse "DEX状态检查成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name}/status [get]
func (dh *DEXHandler) CheckDEXStatus(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	status, err := dh.dexService.CheckDEXStatus(dexName)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Failed to check DEX status",
			Details: err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"dex":    dexName,
		"status": status,
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "DEX status checked successfully",
	})
}

// GetQuote 获取交易报价
// @Summary 获取DEX交易报价
// @Description 获取指定DEX上的代币交换报价
// @Tags DEX交易
// @Produce json
// @Param name path string true "DEX名称"
// @Param inputMint query string true "输入代币地址"
// @Param outputMint query string true "输出代币地址"
// @Param amountIn query string true "输入金额"
// @Success 200 {object} types.SuccessResponse "报价获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name}/quote [get]
func (dh *DEXHandler) GetQuote(c *gin.Context) {
	dexName := c.Param("name")
	inputMint := c.Query("inputMint")
	outputMint := c.Query("outputMint")
	amountInStr := c.Query("amountIn")

	// 验证必需参数
	if dexName == "" || inputMint == "" || outputMint == "" || amountInStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Missing required parameters",
			Details: "dex name, inputMint, outputMint, and amountIn are required",
		})
		return
	}

	// 解析金额
	amountIn, err := parseUint64Param(amountInStr, "amountIn")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid amountIn parameter",
			Details: err.Error(),
		})
		return
	}

	// 获取报价
	quote, err := dh.dexService.GetQuote(dexName, inputMint, outputMint, amountIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to get quote",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    quote,
		Message: "Quote retrieved successfully",
	})
}

// ValidateSwapRequest 验证交换请求
// @Summary 验证交换请求
// @Description 验证交换请求参数的有效性
// @Tags DEX交易
// @Accept json
// @Produce json
// @Param name path string true "DEX名称"
// @Param request body types.SwapRequest true "交换请求参数"
// @Success 200 {object} types.SuccessResponse "验证成功"
// @Failure 400 {object} types.ErrorResponse "验证失败"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name}/validate/swap [post]
func (dh *DEXHandler) ValidateSwapRequest(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	var req types.SwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 验证请求
	if err := dh.dexService.ValidateDEXRequest(dexName, &req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Request validation failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Request validation passed",
	})
}

// ValidateLiquidityRequest 验证流动性请求
// @Summary 验证流动性请求
// @Description 验证流动性请求参数的有效性
// @Tags DEX交易
// @Accept json
// @Produce json
// @Param name path string true "DEX名称"
// @Param request body types.LiquidityRequest true "流动性请求参数"
// @Success 200 {object} types.SuccessResponse "验证成功"
// @Failure 400 {object} types.ErrorResponse "验证失败"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/dex/{name}/validate/liquidity [post]
func (dh *DEXHandler) ValidateLiquidityRequest(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	var req types.LiquidityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// 验证请求
	if err := dh.dexService.ValidateDEXRequest(dexName, &req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Request validation failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Request validation passed",
	})
}