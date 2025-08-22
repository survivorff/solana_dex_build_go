package services

import (
	"fmt"
	"os"
	"time"

	"solana-dex-service/internal/config"

	"gopkg.in/yaml.v3"
)

// ConfigService 配置服务
type ConfigService struct {
	config     *config.Config
	configPath string
}

// NewConfigService 创建配置服务
func NewConfigService(cfg *config.Config) *ConfigService {
	return &ConfigService{
		config:     cfg,
		configPath: "config/config.yaml", // 默认配置文件路径
	}
}

// SetConfigPath 设置配置文件路径
func (cs *ConfigService) SetConfigPath(path string) {
	cs.configPath = path
}

// GetConfig 获取完整配置
func (cs *ConfigService) GetConfig() *config.Config {
	return cs.config
}

// UpdateConfig 更新完整配置
func (cs *ConfigService) UpdateConfig(newConfig *config.Config) error {
	// 验证新配置
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 设置默认值
	newConfig.SetDefaults()

	// 保存到文件
	if err := cs.saveConfigToFile(newConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// 更新内存中的配置
	cs.config = newConfig

	return nil
}

// GetDEXConfig 获取DEX配置
func (cs *ConfigService) GetDEXConfig() []config.DEXConfig {
	return cs.config.DEXes
}

// UpdateDEXConfig 更新DEX配置
func (cs *ConfigService) UpdateDEXConfig(dexConfigs []config.DEXConfig) error {
	// 验证DEX配置
	for i, dex := range dexConfigs {
		if dex.Name == "" {
			return fmt.Errorf("dex[%d] name is required", i)
		}
		if dex.ProgramID == "" {
			return fmt.Errorf("dex[%d] program_id is required", i)
		}
		// 设置更新时间
		dexConfigs[i].UpdatedAt = time.Now()
		if dexConfigs[i].CreatedAt.IsZero() {
			dexConfigs[i].CreatedAt = time.Now()
		}
	}

	// 更新配置
	cs.config.DEXes = dexConfigs

	// 保存到文件
	if err := cs.saveConfigToFile(cs.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// AddDEXConfig 添加新的DEX配置
func (cs *ConfigService) AddDEXConfig(dexConfig config.DEXConfig) error {
	// 检查是否已存在同名DEX
	for _, existing := range cs.config.DEXes {
		if existing.Name == dexConfig.Name {
			return fmt.Errorf("DEX with name '%s' already exists", dexConfig.Name)
		}
	}

	// 验证新配置
	if dexConfig.Name == "" {
		return fmt.Errorf("DEX name is required")
	}
	if dexConfig.ProgramID == "" {
		return fmt.Errorf("DEX program_id is required")
	}

	// 设置时间戳
	dexConfig.CreatedAt = time.Now()
	dexConfig.UpdatedAt = time.Now()

	// 设置默认值
	if dexConfig.Timeout == 0 {
		dexConfig.Timeout = 30 * time.Second
	}
	if dexConfig.RetryCount == 0 {
		dexConfig.RetryCount = 3
	}

	// 添加到配置列表
	cs.config.DEXes = append(cs.config.DEXes, dexConfig)

	// 保存到文件
	if err := cs.saveConfigToFile(cs.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemoveDEXConfig 移除DEX配置
func (cs *ConfigService) RemoveDEXConfig(dexName string) error {
	// 查找要删除的DEX
	index := -1
	for i, dex := range cs.config.DEXes {
		if dex.Name == dexName {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("DEX with name '%s' not found", dexName)
	}

	// 从切片中移除
	cs.config.DEXes = append(cs.config.DEXes[:index], cs.config.DEXes[index+1:]...)

	// 保存到文件
	if err := cs.saveConfigToFile(cs.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// EnableDEX 启用DEX
func (cs *ConfigService) EnableDEX(dexName string) error {
	return cs.setDEXEnabled(dexName, true)
}

// DisableDEX 禁用DEX
func (cs *ConfigService) DisableDEX(dexName string) error {
	return cs.setDEXEnabled(dexName, false)
}

// setDEXEnabled 设置DEX启用状态
func (cs *ConfigService) setDEXEnabled(dexName string, enabled bool) error {
	// 查找DEX
	for i, dex := range cs.config.DEXes {
		if dex.Name == dexName {
			cs.config.DEXes[i].Enabled = enabled
			cs.config.DEXes[i].UpdatedAt = time.Now()

			// 保存到文件
			if err := cs.saveConfigToFile(cs.config); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("DEX with name '%s' not found", dexName)
}

// GetServerConfig 获取服务器配置
func (cs *ConfigService) GetServerConfig() *config.ServerConfig {
	return &cs.config.Server
}

// UpdateServerConfig 更新服务器配置
func (cs *ConfigService) UpdateServerConfig(serverConfig *config.ServerConfig) error {
	// 验证服务器配置
	if serverConfig.Port <= 0 || serverConfig.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", serverConfig.Port)
	}

	// 更新配置
	cs.config.Server = *serverConfig

	// 保存到文件
	if err := cs.saveConfigToFile(cs.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetSolanaConfig 获取Solana配置
func (cs *ConfigService) GetSolanaConfig() *config.SolanaConfig {
	return &cs.config.Solana
}

// UpdateSolanaConfig 更新Solana配置
func (cs *ConfigService) UpdateSolanaConfig(solanaConfig *config.SolanaConfig) error {
	// 验证Solana配置
	if solanaConfig.RPCURL == "" {
		return fmt.Errorf("solana rpc_url is required")
	}
	if solanaConfig.Network == "" {
		return fmt.Errorf("solana network is required")
	}

	// 更新配置
	cs.config.Solana = *solanaConfig

	// 保存到文件
	if err := cs.saveConfigToFile(cs.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ReloadConfig 重新加载配置文件
func (cs *ConfigService) ReloadConfig() error {
	newConfig, err := config.LoadConfig(cs.configPath)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	cs.config = newConfig
	return nil
}

// BackupConfig 备份当前配置
func (cs *ConfigService) BackupConfig() error {
	backupPath := fmt.Sprintf("%s.backup.%d", cs.configPath, time.Now().Unix())
	return cs.saveConfigToPath(cs.config, backupPath)
}

// RestoreConfig 从备份恢复配置
func (cs *ConfigService) RestoreConfig(backupPath string) error {
	// 加载备份配置
	backupConfig, err := config.LoadConfig(backupPath)
	if err != nil {
		return fmt.Errorf("failed to load backup config: %w", err)
	}

	// 保存为当前配置
	if err := cs.saveConfigToFile(backupConfig); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	// 更新内存中的配置
	cs.config = backupConfig

	return nil
}

// ValidateConfig 验证配置
func (cs *ConfigService) ValidateConfig() error {
	return cs.config.Validate()
}

// GetConfigSummary 获取配置摘要
func (cs *ConfigService) GetConfigSummary() map[string]interface{} {
	enabledDEXes := 0
	for _, dex := range cs.config.DEXes {
		if dex.Enabled {
			enabledDEXes++
		}
	}

	return map[string]interface{}{
		"server_port":    cs.config.Server.Port,
		"server_mode":    cs.config.Server.Mode,
		"solana_network": cs.config.Solana.Network,
		"solana_rpc":     cs.config.Solana.RPCURL,
		"total_dexes":    len(cs.config.DEXes),
		"enabled_dexes":  enabledDEXes,
		"log_level":      cs.config.Logging.Level,
	}
}

// saveConfigToFile 保存配置到默认文件
func (cs *ConfigService) saveConfigToFile(cfg *config.Config) error {
	return cs.saveConfigToPath(cfg, cs.configPath)
}

// saveConfigToPath 保存配置到指定路径
func (cs *ConfigService) saveConfigToPath(cfg *config.Config, path string) error {
	// 序列化配置为YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}