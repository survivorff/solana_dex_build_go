package tests

import (
	"os"
	"testing"
	"time"

	"solana-dex-service/internal/config"
	"solana-dex-service/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLoad 测试配置加载
func TestConfigLoad(t *testing.T) {
	// 创建临时配置文件
	tempConfigFile := createTempConfigFile(t)
	defer os.Remove(tempConfigFile)

	// 加载配置
	cfg, err := config.LoadConfig(tempConfigFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 验证配置内容
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "mainnet", cfg.Solana.Network)
	assert.True(t, len(cfg.DEXes) > 0)
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Server: config.ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
					Mode: "debug",
				},
				Solana: config.SolanaConfig{
					RPCURL:  "https://api.mainnet-beta.solana.com",
					Network: "mainnet",
				},
				DEXes: []config.DEXConfig{
					{
						Name:      "test-dex",
						ProgramID: "11111111111111111111111111111112",
						Enabled:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid port",
			config: &config.Config{
				Server: config.ServerConfig{
					Port: -1,
				},
				Solana: config.SolanaConfig{
					RPCURL:  "https://api.mainnet-beta.solana.com",
					Network: "mainnet",
				},
			},
			expectError: true,
		},
		{
			name: "missing rpc url",
			config: &config.Config{
				Server: config.ServerConfig{
					Port: 8080,
				},
				Solana: config.SolanaConfig{
					Network: "mainnet",
				},
			},
			expectError: true,
		},
		{
			name: "missing dex name",
			config: &config.Config{
				Server: config.ServerConfig{
					Port: 8080,
				},
				Solana: config.SolanaConfig{
					RPCURL:  "https://api.mainnet-beta.solana.com",
					Network: "mainnet",
				},
				DEXes: []config.DEXConfig{
					{
						ProgramID: "11111111111111111111111111111112",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfigService 测试配置服务
func TestConfigService(t *testing.T) {
	// 创建临时配置文件
	tempConfigFile := createTempConfigFile(t)
	defer os.Remove(tempConfigFile)

	// 加载配置
	cfg, err := config.LoadConfig(tempConfigFile)
	require.NoError(t, err)

	// 创建配置服务
	configService := services.NewConfigService(cfg)
	configService.SetConfigPath(tempConfigFile)

	// 测试获取配置
	gotConfig := configService.GetConfig()
	assert.NotNil(t, gotConfig)
	assert.Equal(t, cfg.Server.Port, gotConfig.Server.Port)

	// 测试获取DEX配置
	dexConfigs := configService.GetDEXConfig()
	assert.True(t, len(dexConfigs) > 0)

	// 测试添加DEX配置
	newDEXConfig := config.DEXConfig{
		Name:      "test-new-dex",
		ProgramID: "22222222222222222222222222222222",
		Enabled:   true,
		Timeout:   30 * time.Second,
		RetryCount: 3,
	}

	err = configService.AddDEXConfig(newDEXConfig)
	assert.NoError(t, err)

	// 验证DEX已添加
	updatedDEXConfigs := configService.GetDEXConfig()
	assert.Equal(t, len(dexConfigs)+1, len(updatedDEXConfigs))

	// 测试启用/禁用DEX
	err = configService.DisableDEX("test-new-dex")
	assert.NoError(t, err)

	err = configService.EnableDEX("test-new-dex")
	assert.NoError(t, err)

	// 测试移除DEX配置
	err = configService.RemoveDEXConfig("test-new-dex")
	assert.NoError(t, err)

	// 验证DEX已移除
	finalDEXConfigs := configService.GetDEXConfig()
	assert.Equal(t, len(dexConfigs), len(finalDEXConfigs))
}

// TestConfigDefaults 测试默认值设置
func TestConfigDefaults(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		Solana: config.SolanaConfig{
			RPCURL:  "https://api.mainnet-beta.solana.com",
			Network: "mainnet",
		},
		DEXes: []config.DEXConfig{
			{
				Name:      "test-dex",
				ProgramID: "11111111111111111111111111111112",
				Enabled:   true,
			},
		},
	}

	// 设置默认值
	cfg.SetDefaults()

	// 验证默认值
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)

	assert.Equal(t, 30*time.Second, cfg.Solana.Timeout)
	assert.Equal(t, 3, cfg.Solana.RetryCount)
	assert.Equal(t, "confirmed", cfg.Solana.Commitment)

	assert.Equal(t, 30*time.Second, cfg.DEXes[0].Timeout)
	assert.Equal(t, 3, cfg.DEXes[0].RetryCount)
	assert.False(t, cfg.DEXes[0].CreatedAt.IsZero())
	assert.False(t, cfg.DEXes[0].UpdatedAt.IsZero())

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)

	assert.Equal(t, 100, cfg.Security.RateLimitRPS)
	assert.Equal(t, int64(1024*1024), cfg.Security.MaxRequestSize)
}

// TestGetDEXConfig 测试获取DEX配置
func TestGetDEXConfig(t *testing.T) {
	cfg := &config.Config{
		DEXes: []config.DEXConfig{
			{
				Name:    "raydium",
				Enabled: true,
			},
			{
				Name:    "pumpfun",
				Enabled: false,
			},
			{
				Name:    "pumpswap",
				Enabled: true,
			},
		},
	}

	// 测试获取存在且启用的DEX
	dexConfig, err := cfg.GetDEXConfig("raydium")
	assert.NoError(t, err)
	assert.NotNil(t, dexConfig)
	assert.Equal(t, "raydium", dexConfig.Name)

	// 测试获取存在但禁用的DEX
	dexConfig, err = cfg.GetDEXConfig("pumpfun")
	assert.Error(t, err)
	assert.Nil(t, dexConfig)

	// 测试获取不存在的DEX
	dexConfig, err = cfg.GetDEXConfig("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, dexConfig)

	// 测试获取所有启用的DEX
	enabledDEXes := cfg.GetEnabledDEXes()
	assert.Len(t, enabledDEXes, 2)
	assert.Equal(t, "raydium", enabledDEXes[0].Name)
	assert.Equal(t, "pumpswap", enabledDEXes[1].Name)
}

// createTempConfigFile 创建临时配置文件用于测试
func createTempConfigFile(t *testing.T) string {
	configContent := `
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"
  read_timeout: 30s
  write_timeout: 30s

solana:
  rpc_url: "https://api.mainnet-beta.solana.com"
  ws_url: "wss://api.mainnet-beta.solana.com"
  network: "mainnet"
  timeout: 30s
  retry_count: 3
  commitment: "confirmed"

dexes:
  - name: "raydium"
    program_id: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
    router_address: "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS"
    endpoints:
      swap: "https://api.raydium.io/v2/sdk/swap"
      pools: "https://api.raydium.io/v2/sdk/liquidity/mainnet.json"
    enabled: true
    timeout: 30s
    retry_count: 3

  - name: "pumpfun"
    program_id: "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
    router_address: "39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg"
    endpoints:
      api: "https://pumpportal.fun/api"
    enabled: true
    timeout: 30s
    retry_count: 3

logging:
  level: "info"
  format: "json"
  output: "stdout"

security:
  enable_https: false
  rate_limit_rps: 100
  max_request_size: 1048576
`

	tempFile, err := os.CreateTemp("", "config_test_*.yaml")
	require.NoError(t, err)

	_, err = tempFile.WriteString(configContent)
	require.NoError(t, err)

	err = tempFile.Close()
	require.NoError(t, err)

	return tempFile.Name()
}