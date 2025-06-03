package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 结构体定义配置结构
type Config struct {
	App struct {
		Name             string `yaml:"name"`
		Env              string `yaml:"env"`
		Port             int    `yaml:"port"`
		Version          string `yaml:"version"`
		BaseDomainUrl    string `yaml:"basedomainurl"`
		BaseCachePageURL string `yaml:"basecachepageurl"`
		BaseCacheDataURL string `yaml:"basecachedataurl"`
		BaseCacheLogURL  string `yaml:"basecachelogurl"`
		BaseWaterPageUrl string `yaml:"basewaterpageurl"`
		TotalPages       int    `yaml:"totalpages"`
		StartPage        int    `yaml:"startpage"`
	} `yaml:"app"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

// 单例配置实例
var instance *Config

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error) {
	if instance != nil {
		return instance, nil
	}

	// 读取配置文件
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 应用环境变量覆盖
	overrideWithEnv(&config)

	// 保存单例
	instance = &config
	return instance, nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if instance == nil {
		panic("配置未初始化，请先调用LoadConfig")
	}
	return instance
}

// overrideWithEnv 使用环境变量覆盖配置
func overrideWithEnv(config *Config) {
	if env := os.Getenv("APP_ENV"); env != "" {
		config.App.Env = env
	}
	if port := os.Getenv("APP_PORT"); port != "" {
		// 类型转换
		// 实际项目中应添加错误处理
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
}
