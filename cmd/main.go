package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/handlers"
	"solana-dex-service/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化服务
	transactionService := services.NewTransactionService(cfg)
	dexService := services.NewDEXService(cfg)
	configService := services.NewConfigService(cfg)

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 设置CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 初始化处理器
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	dexHandler := handlers.NewDEXHandler(dexService)
	configHandler := handlers.NewConfigHandler(configService)

	// 设置路由
	setupRoutes(router, transactionHandler, dexHandler, configHandler)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭服务器，等待5秒钟完成现有请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// setupRoutes 设置所有路由
func setupRoutes(router *gin.Engine, transactionHandler *handlers.TransactionHandler, dexHandler *handlers.DEXHandler, configHandler *handlers.ConfigHandler) {
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
		}

		// 配置管理相关路由
		config := v1.Group("/config")
		{
			config.GET("/", configHandler.GetConfig)
			config.PUT("/", configHandler.UpdateConfig)
			config.GET("/dex", configHandler.GetDEXConfig)
			config.PUT("/dex", configHandler.UpdateDEXConfig)
		}
	}
}