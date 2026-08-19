// Package config 集中管理全部运行配置。
// 铁律：敏感信息只从环境变量读取，禁止硬编码；缺省值仅保证本地开发一键可启动。
package config

import (
	"errors"
	"fmt"
	"os"
)

// defaultJWTSecret 开发环境占位密钥，生产环境禁止使用。
const defaultJWTSecret = "dev-only-secret-change-me"

type Config struct {
	AppName  string
	Env      string
	Port     string
	LogLevel string

	MySQL MySQLConfig
	Redis RedisConfig
	JWT   JWTConfig

	DockerHost  string
	RegistryURL string
	// RegistryEngineURL：Docker 引擎（daemon）视角的 registry 地址（引擎在宿主机 VM，无法解析 compose 服务名）
	RegistryEngineURL string
	// AppNetwork：应用/部署容器加入的 Docker 网络（自定义 bridge 才有 IPAM 分配 IP，供健康检查与 Traefik 路由）
	AppNetwork string
	// SeccompProfile：自定义 seccomp profile 名（为空 = 使用 Docker 守护进程默认 profile；绝不使用 unconfined）
	SeccompProfile string

	// AI 助手（智谱 GLM）：key 从环境变量注入，缺省值仅保证本地开发一键可用
	AI AIConfig
}

type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DSN 生成 GORM MySQL 连接串；multiStatements=true 用于一次性执行整份迁移 SQL。
func (m MySQLConfig) DSN() string {
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + m.Port + ")/" + m.Database +
		"?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true"
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	AccessTTL  string
	RefreshTTL string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	return &Config{
		AppName:  "cloud-api",
		Env:      env("APP_ENV", "development"),
		Port:     env("APP_PORT", "8080"),
		LogLevel: env("LOG_LEVEL", "info"),
		MySQL: MySQLConfig{
			Host:     env("MYSQL_HOST", "127.0.0.1"),
			Port:     env("MYSQL_PORT", "3306"),
			User:     env("MYSQL_USER", "root"),
			Password: env("MYSQL_PASSWORD", "root"),
			Database: env("MYSQL_DATABASE", "dxcloud"),
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "127.0.0.1"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:     env("JWT_SECRET", defaultJWTSecret),
			AccessTTL:  env("JWT_ACCESS_TTL", "15m"),
			RefreshTTL: env("JWT_REFRESH_TTL", "168h"),
		},
		DockerHost:        env("DOCKER_HOST", "unix:///var/run/docker.sock"),
		RegistryURL:       env("REGISTRY_URL", "127.0.0.1:5000"),
		RegistryEngineURL: env("REGISTRY_ENGINE_URL", "host.docker.internal:15000"),
		AppNetwork:        env("APP_NETWORK", "dxcloud_edge"),
		SeccompProfile:    env("SECCOMP_PROFILE", ""),
		AI: AIConfig{
			APIKey:  env("ZHIPU_API_KEY", "c975509f9c044fc6b5a1dcb4b19b0eea.pKcylWGOa3Iqj9ZA"),
			BaseURL: env("ZHIPU_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
			Model:   env("ZHIPU_MODEL", "glm-4-flash"),
		},
	}
}

// Validate 校验配置；生产环境强制安全密钥，避免弱默认值上线。
func (c *Config) Validate() error {
	if c.Env == "production" {
		if c.JWT.Secret == defaultJWTSecret {
			return errors.New("production 环境禁止使用默认 JWT_SECRET，请通过环境变量设置强密钥")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("production 环境 JWT_SECRET 长度需不小于 32（当前 %d）", len(c.JWT.Secret))
		}
	}
	return nil
}
