package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

var (
	// ConfigDir 配置目录
	ConfigDir string
	// ConfigFile 配置文件路径
	ConfigFile string
)

// AppConfig 应用配置结构
type AppConfig struct {
	BaseURL          string    `json:"base_url"`
	ClientID         string    `json:"client_id"`
	ClientSecret     string    `json:"client_secret"`
	DefaultWorkspace string    `json:"default_workspace,omitempty"`

	// 登录后的凭证
	APIKey           string    `json:"api_key,omitempty"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	TokenExpiresAt   time.Time `json:"token_expires_at,omitempty"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at,omitempty"`

	// 更新检查缓存
	UpdateLastCheckedAt  time.Time `json:"update_last_checked_at,omitempty"`
	UpdateLatestVersion  string    `json:"update_latest_version,omitempty"`
}

// Init 初始化配置目录
func Init() {
	home, _ := os.UserHomeDir()
	ConfigDir = filepath.Join(home, ".cbi")
	ConfigFile = filepath.Join(ConfigDir, "config.json")

	// 确保目录存在
	_ = os.MkdirAll(ConfigDir, 0755)
}

// Exists 检查配置文件是否存在
func Exists() bool {
	Init()

	if _, err := os.Stat(ConfigFile); err != nil {
		return false
	}

	// 检查文件是否有有效内容
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return false
	}

	// 解析 JSON 检查是否有必要字段
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	return cfg.ClientID != "" && cfg.ClientSecret != ""
}

// LoadAppConfig 加载应用配置
func LoadAppConfig() (*AppConfig, error) {
	Init()

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveAppConfig 保存应用配置
func SaveAppConfig(cfg *AppConfig) error {
	Init()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFile, data, 0600)
}

// GetConfigFile 获取配置文件路径
func GetConfigFile() string {
	Init()
	return ConfigFile
}

// GetClientID 获取 client_id
func GetClientID() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return ""
	}
	return cfg.ClientID
}

// GetClientSecret 获取 client_secret
func GetClientSecret() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return ""
	}
	return cfg.ClientSecret
}

// GetBaseURL 获取 base_url
func GetBaseURL() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return "https://open.creatibi.cn"
	}
	if cfg.BaseURL == "" {
		return "https://open.creatibi.cn"
	}
	return cfg.BaseURL
}

// GetAPIKey 获取 API Key (access_token)
func GetAPIKey() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return ""
	}
	return cfg.APIKey
}

// SetAPIKey 设置 API Key (access_token)
func SetAPIKey(apiKey string) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		// 如果配置不存在，创建空配置
		cfg = &AppConfig{}
	}

	cfg.APIKey = apiKey
	return SaveAppConfig(cfg)
}

// GetRefreshToken 获取 refresh token
func GetRefreshToken() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return ""
	}
	return cfg.RefreshToken
}

// SetRefreshToken 设置 refresh token
func SetRefreshToken(refreshToken string) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		cfg = &AppConfig{}
	}

	cfg.RefreshToken = refreshToken
	return SaveAppConfig(cfg)
}

// SetTokenExpiresAt 设置 token 过期时间
func SetTokenExpiresAt(expiresAt time.Time) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		cfg = &AppConfig{}
	}

	cfg.TokenExpiresAt = expiresAt
	return SaveAppConfig(cfg)
}

// GetTokenExpiresAt 获取 token 过期时间
func GetTokenExpiresAt() time.Time {
	cfg, err := LoadAppConfig()
	if err != nil {
		return time.Time{}
	}
	return cfg.TokenExpiresAt
}

// SetRefreshTokenExpiresAt 设置 refresh token 过期时间
func SetRefreshTokenExpiresAt(expiresAt time.Time) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		cfg = &AppConfig{}
	}

	cfg.RefreshTokenExpiresAt = expiresAt
	return SaveAppConfig(cfg)
}

// IsLoggedIn 检查是否已登录
func IsLoggedIn() bool {
	return GetAPIKey() != ""
}

// Clear 清除登录凭证（保留应用配置）
func Clear() error {
	cfg, err := LoadAppConfig()
	if err != nil {
		return err
	}

	// 只清除登录凭证，保留应用配置
	cfg.APIKey = ""
	cfg.RefreshToken = ""
	cfg.TokenExpiresAt = time.Time{}
	cfg.RefreshTokenExpiresAt = time.Time{}

	return SaveAppConfig(cfg)
}

// GetUpdateLastCheckedAt 获取上次更新检查时间
func GetUpdateLastCheckedAt() time.Time {
	cfg, err := LoadAppConfig()
	if err != nil {
		return time.Time{}
	}
	return cfg.UpdateLastCheckedAt
}

// SetUpdateLastCheckedAt 设置上次更新检查时间
func SetUpdateLastCheckedAt(t time.Time) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		cfg = &AppConfig{}
	}
	cfg.UpdateLastCheckedAt = t
	return SaveAppConfig(cfg)
}

// GetUpdateLatestVersion 获取缓存的最新版本
func GetUpdateLatestVersion() string {
	cfg, err := LoadAppConfig()
	if err != nil {
		return ""
	}
	return cfg.UpdateLatestVersion
}

// SetUpdateLatestVersion 设置缓存的最新版本
func SetUpdateLatestVersion(version string) error {
	cfg, err := LoadAppConfig()
	if err != nil {
		cfg = &AppConfig{}
	}
	cfg.UpdateLatestVersion = version
	return SaveAppConfig(cfg)
}