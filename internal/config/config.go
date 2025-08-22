package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig `yaml:"server"`
	Solana   SolanaConfig `yaml:"solana"`
	DEXes    []DEXConfig  `yaml:"dexes"`
	Logging  LogConfig    `yaml:"logging"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	Mode         string        `yaml:"mode"` // debug, release, test
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// SolanaConfig Solana网络配置
type SolanaConfig struct {
	RPCURL      string        `yaml:"rpc_url"`
	WSURL       string        `yaml:"ws_url"`
	Network     string        `yaml:"network"` // mainnet, devnet, testnet
	Timeout     time.Duration `yaml:"timeout"`
	RetryCount  int           `yaml:"retry_count"`
	Commitment  string        `yaml:"commitment"` // processed, confirmed, finalized
}

// DEXConfig DEX配置
type DEXConfig struct {
	Name          string            `yaml:"name"`
	ProgramID     string            `yaml:"program_id"`
	RouterAddress string            `yaml:"router_address"`
	Endpoints     map[string]string `yaml:"endpoints"`
	Enabled       bool              `yaml:"enabled"`
	Timeout       time.Duration     `yaml:"timeout"`
	RetryCount    int               `yaml:"retry_count"`
	CreatedAt     time.Time         `yaml:"created_at"`
	UpdatedAt     time.Time         `yaml:"updated_at"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`      // debug, info, warn, error
	Format     string `yaml:"format"`     // json, text
	Output     string `yaml:"output"`     // stdout, file
	FilePath   string `yaml:"file_path"`  // 日志文件路径
	MaxSize    int    `yaml:"max_size"`   // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`    // days
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableHTTPS    bool   `yaml:"enable_https"`
	CertFile       string `yaml:"cert_file"`
	KeyFile        string `yaml:"key_file"`
	RateLimitRPS   int    `yaml:"rate_limit_rps"`
	MaxRequestSize int64  `yaml:"max_request_size"` // bytes
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 设置默认值
	config.SetDefaults()

	return &config, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// 验证Solana配置
	if c.Solana.RPCURL == "" {
		return fmt.Errorf("solana rpc_url is required")
	}

	if c.Solana.Network == "" {
		return fmt.Errorf("solana network is required")
	}

	// 验证DEX配置
	for i, dex := range c.DEXes {
		if dex.Name == "" {
			return fmt.Errorf("dex[%d] name is required", i)
		}
		if dex.ProgramID == "" {
			return fmt.Errorf("dex[%d] program_id is required", i)
		}
	}

	return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
	// 服务器默认值
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}

	// Solana默认值
	if c.Solana.Timeout == 0 {
		c.Solana.Timeout = 30 * time.Second
	}
	if c.Solana.RetryCount == 0 {
		c.Solana.RetryCount = 3
	}
	if c.Solana.Commitment == "" {
		c.Solana.Commitment = "confirmed"
	}

	// DEX默认值
	for i := range c.DEXes {
		if c.DEXes[i].Timeout == 0 {
			c.DEXes[i].Timeout = 30 * time.Second
		}
		if c.DEXes[i].RetryCount == 0 {
			c.DEXes[i].RetryCount = 3
		}
		if c.DEXes[i].CreatedAt.IsZero() {
			c.DEXes[i].CreatedAt = time.Now()
		}
		c.DEXes[i].UpdatedAt = time.Now()
	}

	// 日志默认值
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	// 安全默认值
	if c.Security.RateLimitRPS == 0 {
		c.Security.RateLimitRPS = 100
	}
	if c.Security.MaxRequestSize == 0 {
		c.Security.MaxRequestSize = 1024 * 1024 // 1MB
	}
}

// GetDEXConfig 根据名称获取DEX配置
func (c *Config) GetDEXConfig(name string) (*DEXConfig, error) {
	for _, dex := range c.DEXes {
		if dex.Name == name && dex.Enabled {
			return &dex, nil
		}
	}
	return nil, fmt.Errorf("dex config not found or disabled: %s", name)
}

// GetEnabledDEXes 获取所有启用的DEX配置
func (c *Config) GetEnabledDEXes() []DEXConfig {
	var enabled []DEXConfig
	for _, dex := range c.DEXes {
		if dex.Enabled {
			enabled = append(enabled, dex)
		}
	}
	return enabled
}