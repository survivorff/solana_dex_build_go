package handlers

import (
	"net/http"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/services"
	"solana-dex-service/internal/types"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	configService *services.ConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *services.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetConfig 获取完整配置
// @Summary 获取系统配置
// @Description 获取系统的完整配置信息
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "配置获取成功"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config [get]
func (ch *ConfigHandler) GetConfig(c *gin.Context) {
	config := ch.configService.GetConfig()

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    config,
		Message: "Configuration retrieved successfully",
	})
}

// UpdateConfig 更新完整配置
// @Summary 更新系统配置
// @Description 更新系统的完整配置信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param config body config.Config true "配置信息"
// @Success 200 {object} types.SuccessResponse "配置更新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config [put]
func (ch *ConfigHandler) UpdateConfig(c *gin.Context) {
	var newConfig config.Config

	// 绑定请求参数
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid configuration parameters",
			Details: err.Error(),
		})
		return
	}

	// 更新配置
	if err := ch.configService.UpdateConfig(&newConfig); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to update configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Configuration updated successfully",
	})
}

// GetDEXConfig 获取DEX配置
// @Summary 获取DEX配置
// @Description 获取所有DEX的配置信息
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "DEX配置获取成功"
// @Router /api/v1/config/dex [get]
func (ch *ConfigHandler) GetDEXConfig(c *gin.Context) {
	dexConfig := ch.configService.GetDEXConfig()

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    dexConfig,
		Message: "DEX configuration retrieved successfully",
	})
}

// UpdateDEXConfig 更新DEX配置
// @Summary 更新DEX配置
// @Description 更新所有DEX的配置信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param dexConfig body []config.DEXConfig true "DEX配置列表"
// @Success 200 {object} types.SuccessResponse "DEX配置更新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/dex [put]
func (ch *ConfigHandler) UpdateDEXConfig(c *gin.Context) {
	var dexConfigs []config.DEXConfig

	// 绑定请求参数
	if err := c.ShouldBindJSON(&dexConfigs); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid DEX configuration parameters",
			Details: err.Error(),
		})
		return
	}

	// 更新DEX配置
	if err := ch.configService.UpdateDEXConfig(dexConfigs); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to update DEX configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "DEX configuration updated successfully",
	})
}

// AddDEXConfig 添加DEX配置
// @Summary 添加DEX配置
// @Description 添加新的DEX配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param dexConfig body config.DEXConfig true "DEX配置"
// @Success 201 {object} types.SuccessResponse "DEX配置添加成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/dex [post]
func (ch *ConfigHandler) AddDEXConfig(c *gin.Context) {
	var dexConfig config.DEXConfig

	// 绑定请求参数
	if err := c.ShouldBindJSON(&dexConfig); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid DEX configuration parameters",
			Details: err.Error(),
		})
		return
	}

	// 添加DEX配置
	if err := ch.configService.AddDEXConfig(dexConfig); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Failed to add DEX configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.SuccessResponse{
		Success: true,
		Message: "DEX configuration added successfully",
	})
}

// RemoveDEXConfig 移除DEX配置
// @Summary 移除DEX配置
// @Description 根据名称移除DEX配置
// @Tags 配置管理
// @Produce json
// @Param name path string true "DEX名称"
// @Success 200 {object} types.SuccessResponse "DEX配置移除成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/dex/{name} [delete]
func (ch *ConfigHandler) RemoveDEXConfig(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	// 移除DEX配置
	if err := ch.configService.RemoveDEXConfig(dexName); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Failed to remove DEX configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "DEX configuration removed successfully",
	})
}

// EnableDEX 启用DEX
// @Summary 启用DEX
// @Description 启用指定的DEX
// @Tags 配置管理
// @Produce json
// @Param name path string true "DEX名称"
// @Success 200 {object} types.SuccessResponse "DEX启用成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/dex/{name}/enable [post]
func (ch *ConfigHandler) EnableDEX(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	// 启用DEX
	if err := ch.configService.EnableDEX(dexName); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Failed to enable DEX",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "DEX enabled successfully",
	})
}

// DisableDEX 禁用DEX
// @Summary 禁用DEX
// @Description 禁用指定的DEX
// @Tags 配置管理
// @Produce json
// @Param name path string true "DEX名称"
// @Success 200 {object} types.SuccessResponse "DEX禁用成功"
// @Failure 400 {object} types.ErrorResponse "DEX不存在"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/dex/{name}/disable [post]
func (ch *ConfigHandler) DisableDEX(c *gin.Context) {
	dexName := c.Param("name")
	if dexName == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "DEX name is required",
			Details: "Please provide a valid DEX name",
		})
		return
	}

	// 禁用DEX
	if err := ch.configService.DisableDEX(dexName); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Failed to disable DEX",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "DEX disabled successfully",
	})
}

// GetServerConfig 获取服务器配置
// @Summary 获取服务器配置
// @Description 获取HTTP服务器的配置信息
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "服务器配置获取成功"
// @Router /api/v1/config/server [get]
func (ch *ConfigHandler) GetServerConfig(c *gin.Context) {
	serverConfig := ch.configService.GetServerConfig()

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    serverConfig,
		Message: "Server configuration retrieved successfully",
	})
}

// UpdateServerConfig 更新服务器配置
// @Summary 更新服务器配置
// @Description 更新HTTP服务器的配置信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param serverConfig body config.ServerConfig true "服务器配置"
// @Success 200 {object} types.SuccessResponse "服务器配置更新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/server [put]
func (ch *ConfigHandler) UpdateServerConfig(c *gin.Context) {
	var serverConfig config.ServerConfig

	// 绑定请求参数
	if err := c.ShouldBindJSON(&serverConfig); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid server configuration parameters",
			Details: err.Error(),
		})
		return
	}

	// 更新服务器配置
	if err := ch.configService.UpdateServerConfig(&serverConfig); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to update server configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Server configuration updated successfully",
	})
}

// GetSolanaConfig 获取Solana配置
// @Summary 获取Solana配置
// @Description 获取Solana网络的配置信息
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "Solana配置获取成功"
// @Router /api/v1/config/solana [get]
func (ch *ConfigHandler) GetSolanaConfig(c *gin.Context) {
	solanaConfig := ch.configService.GetSolanaConfig()

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    solanaConfig,
		Message: "Solana configuration retrieved successfully",
	})
}

// UpdateSolanaConfig 更新Solana配置
// @Summary 更新Solana配置
// @Description 更新Solana网络的配置信息
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param solanaConfig body config.SolanaConfig true "Solana配置"
// @Success 200 {object} types.SuccessResponse "Solana配置更新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/solana [put]
func (ch *ConfigHandler) UpdateSolanaConfig(c *gin.Context) {
	var solanaConfig config.SolanaConfig

	// 绑定请求参数
	if err := c.ShouldBindJSON(&solanaConfig); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Invalid Solana configuration parameters",
			Details: err.Error(),
		})
		return
	}

	// 更新Solana配置
	if err := ch.configService.UpdateSolanaConfig(&solanaConfig); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to update Solana configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Solana configuration updated successfully",
	})
}

// ReloadConfig 重新加载配置
// @Summary 重新加载配置
// @Description 从配置文件重新加载系统配置
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "配置重新加载成功"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/reload [post]
func (ch *ConfigHandler) ReloadConfig(c *gin.Context) {
	if err := ch.configService.ReloadConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to reload configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Configuration reloaded successfully",
	})
}

// BackupConfig 备份配置
// @Summary 备份配置
// @Description 备份当前的系统配置
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "配置备份成功"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/v1/config/backup [post]
func (ch *ConfigHandler) BackupConfig(c *gin.Context) {
	if err := ch.configService.BackupConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Failed to backup configuration",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Configuration backed up successfully",
	})
}

// ValidateConfig 验证配置
// @Summary 验证配置
// @Description 验证当前配置的有效性
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "配置验证成功"
// @Failure 400 {object} types.ErrorResponse "配置验证失败"
// @Router /api/v1/config/validate [post]
func (ch *ConfigHandler) ValidateConfig(c *gin.Context) {
	if err := ch.configService.ValidateConfig(); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "Configuration validation failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Configuration validation passed",
	})
}

// GetConfigSummary 获取配置摘要
// @Summary 获取配置摘要
// @Description 获取系统配置的摘要信息
// @Tags 配置管理
// @Produce json
// @Success 200 {object} types.SuccessResponse "配置摘要获取成功"
// @Router /api/v1/config/summary [get]
func (ch *ConfigHandler) GetConfigSummary(c *gin.Context) {
	summary := ch.configService.GetConfigSummary()

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Data:    summary,
		Message: "Configuration summary retrieved successfully",
	})
}