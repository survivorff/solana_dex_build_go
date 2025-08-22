package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"solana-dex-service/internal/handlers"
	"solana-dex-service/internal/services"
	"solana-dex-service/internal/types"
)

// TestAPIIntegration 测试API集成
func TestAPIIntegration(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := createTestConfig()

	// 创建服务
	transactionService := services.NewTransactionService(cfg)
	dexService := services.NewDEXService(cfg)
	configService := services.NewConfigService(cfg)

	// 设置服务依赖
	dexService.SetTransactionService(transactionService)

	// 创建处理器
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dexHandler := handlers.NewDEXHandler(dexService)
	configHandler := handlers.NewConfigHandler(configService)

	// 创建路由
	router := setupTestRouter(transactionHandler, dexHandler, configHandler)

	// 运行测试
	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t, router)
	})

	t.Run("ListDEXes", func(t *testing.T) {
		testListDEXes(t, router)
	})

	t.Run("GetDEX", func(t *testing.T) {
		testGetDEX(t, router)
	})

	t.Run("EncodeSwap", func(t *testing.T) {
		testEncodeSwap(t, router)
	})

	t.Run("EncodeLiquidity", func(t *testing.T) {
		testEncodeLiquidity(t, router)
	})

	t.Run("GetConfig", func(t *testing.T) {
		testGetConfig(t, router)
	})

	t.Run("ValidateSwapRequest", func(t *testing.T) {
		testValidateSwapRequest(t, router)
	})
}

// setupTestRouter 设置测试路由
func setupTestRouter(transactionHandler *handlers.TransactionHandler, dexHandler *handlers.DEXHandler, configHandler *handlers.ConfigHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 交易编码相关路由
		encode := v1.Group("/encode")
		{
			encode.POST("/swap", transactionHandler.EncodeSwap)
			encode.POST("/liquidity", transactionHandler.EncodeLiquidity)
		}

		// 交易测试相关路由
		test := v1.Group("/test")
		{
			test.POST("/transaction", transactionHandler.TestTransaction)
			test.POST("/simulate", transactionHandler.SimulateTransaction)
		}

		// DEX相关路由
		dex := v1.Group("/dex")
		{
			dex.GET("/list", dexHandler.ListDEXes)
			dex.GET("/:name", dexHandler.GetDEX)
			dex.GET("/:name/pools", dexHandler.GetPools)
			dex.GET("/:name/status", dexHandler.CheckDEXStatus)
			dex.GET("/:name/quote", dexHandler.GetQuote)
			dex.POST("/:name/validate/swap", dexHandler.ValidateSwapRequest)
			dex.POST("/:name/validate/liquidity", dexHandler.ValidateLiquidityRequest)
		}

		// 配置管理相关路由
		config := v1.Group("/config")
		{
			config.GET("/", configHandler.GetConfig)
			config.PUT("/", configHandler.UpdateConfig)
			config.GET("/dex", configHandler.GetDEXConfig)
			config.PUT("/dex", configHandler.UpdateDEXConfig)
			config.GET("/summary", configHandler.GetConfigSummary)
			config.POST("/validate", configHandler.ValidateConfig)
		}
	}

	return router
}

// testHealthCheck 测试健康检查
func testHealthCheck(t *testing.T, router *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
}

// testListDEXes 测试获取DEX列表
func testListDEXes(t *testing.T, router *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dex/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// 验证返回的DEX列表
	dexes, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.True(t, len(dexes) > 0)
}

// testGetDEX 测试获取指定DEX
func testGetDEX(t *testing.T, router *gin.Engine) {
	// 测试获取存在的DEX
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dex/raydium", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// 测试获取不存在的DEX
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/dex/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var errorResponse types.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "DEX not found")
}

// testEncodeSwap 测试编码交换交易
func testEncodeSwap(t *testing.T, router *gin.Engine) {
	// 创建有效的交换请求
	swapReq := types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112", // SOL
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountIn:   1000000000, // 1 SOL
		Slippage:   0.005,      // 0.5%
		UserWallet: "11111111111111111111111111111112",
	}

	reqBody, err := json.Marshal(swapReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/encode/swap", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.TransactionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应
	if response.Success {
		assert.NotEmpty(t, response.Transaction)
		assert.NotEmpty(t, response.RequestID)
	} else {
		assert.NotEmpty(t, response.Error)
	}

	// 测试无效请求
	invalidReq := types.SwapRequest{
		DEXType: "invalid-dex",
	}

	reqBody, err = json.Marshal(invalidReq)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/encode/swap", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// testEncodeLiquidity 测试编码流动性交易
func testEncodeLiquidity(t *testing.T, router *gin.Engine) {
	// 创建有效的流动性请求
	liquidityReq := types.LiquidityRequest{
		DEXType:     "raydium",
		Operation:   "add",
		TokenAMint:  "So11111111111111111111111111111111111111112", // SOL
		TokenBMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		AmountA:     1000000000, // 1 SOL
		AmountB:     1000000,    // 1 USDC
		Slippage:    0.005,      // 0.5%
		UserWallet:  "11111111111111111111111111111112",
	}

	reqBody, err := json.Marshal(liquidityReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/encode/liquidity", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.TransactionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应
	if response.Success {
		assert.NotEmpty(t, response.Transaction)
		assert.NotEmpty(t, response.RequestID)
	} else {
		assert.NotEmpty(t, response.Error)
	}
}

// testGetConfig 测试获取配置
func testGetConfig(t *testing.T, router *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/config/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// 测试获取配置摘要
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/config/summary", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// 验证摘要内容
	summary, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, summary, "server_port")
	assert.Contains(t, summary, "solana_network")
	assert.Contains(t, summary, "total_dexes")
}

// testValidateSwapRequest 测试验证交换请求
func testValidateSwapRequest(t *testing.T, router *gin.Engine) {
	// 创建有效的交换请求
	swapReq := types.SwapRequest{
		DEXType:    "raydium",
		InputMint:  "So11111111111111111111111111111111111111112",
		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		AmountIn:   1000000000,
		Slippage:   0.005,
		UserWallet: "11111111111111111111111111111112",
	}

	reqBody, err := json.Marshal(swapReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/dex/raydium/validate/swap", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response types.SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// 测试无效请求
	invalidReq := types.SwapRequest{
		DEXType:   "raydium",
		InputMint: "", // 空的输入代币
	}

	reqBody, err = json.Marshal(invalidReq)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/dex/raydium/validate/swap", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var errorResponse types.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "validation failed")
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := createTestConfig()

	// 创建服务
	transactionService := services.NewTransactionService(cfg)
	dexService := services.NewDEXService(cfg)
	configService := services.NewConfigService(cfg)

	// 设置服务依赖
	dexService.SetTransactionService(transactionService)

	// 创建处理器
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dexHandler := handlers.NewDEXHandler(dexService)
	configHandler := handlers.NewConfigHandler(configService)

	// 创建路由
	router := setupTestRouter(transactionHandler, dexHandler, configHandler)

	// 测试无效的JSON请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/encode/swap", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var errorResponse types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "Invalid request parameters")

	// 测试不存在的路由
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	// 测试不支持的HTTP方法
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/encode/swap", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 405, w.Code)
}

// TestConcurrentRequests 测试并发请求
func TestConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := createTestConfig()

	// 创建服务
	transactionService := services.NewTransactionService(cfg)
	dexService := services.NewDEXService(cfg)
	configService := services.NewConfigService(cfg)

	// 设置服务依赖
	dexService.SetTransactionService(transactionService)

	// 创建处理器
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dexHandler := handlers.NewDEXHandler(dexService)
	configHandler := handlers.NewConfigHandler(configService)

	// 创建路由
	router := setupTestRouter(transactionHandler, dexHandler, configHandler)

	// 并发测试
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()

			// 测试健康检查
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)

			// 测试获取DEX列表
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/dex/list", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)

			// 测试获取配置摘要
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/v1/config/summary", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
}