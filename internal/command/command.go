// Package command 提供 CLI 命令的公共组件
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/lwmacct/251203-vm-metrics/internal/config"
	"github.com/lwmacct/251207-go-mod-version/pkg/version"

	"github.com/lwmacct/251203-vm-metrics/internal/vmapi"
	"github.com/urfave/cli/v3"
)

// Defaults 默认配置 - 单一来源 (Single Source of Truth)
var Defaults = config.DefaultConfig()

// MetaKeyConfig 配置在 Metadata 中的 key
const MetaKeyConfig = "config"

// BeforeLoadConfig 在 Action 执行前加载配置
// 将配置存入 cmd.Metadata 供后续 Action 使用
func BeforeLoadConfig(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	cfg, err := config.Load(cmd, cmd.String("config"), version.GetAppRawName())
	if err != nil {
		return ctx, err
	}

	if cmd.Metadata == nil {
		cmd.Metadata = make(map[string]any)
	}
	cmd.Metadata[MetaKeyConfig] = cfg
	return ctx, nil
}

// GetConfig 从 cmd.Metadata 获取已加载的配置
func GetConfig(cmd *cli.Command) *config.Config {
	if cmd.Metadata == nil {
		return nil
	}
	if cfg, ok := cmd.Metadata[MetaKeyConfig].(*config.Config); ok {
		return cfg
	}
	return nil
}

// BaseFlags 返回所有命令共享的基础 flags
// 包括：配置文件、服务器、认证、TLS 配置
func BaseFlags() []cli.Flag {
	return []cli.Flag{
		// 配置文件
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "配置文件路径",
		},
		// 服务器配置
		&cli.StringFlag{
			Name:  "server-url",
			Usage: "VictoriaMetrics 服务器地址",
			Value: Defaults.Server.URL,
		},
		&cli.StringFlag{
			Name:  "server-path-prefix",
			Usage: "API 路径前缀 (如 /victoria)",
		},
		&cli.DurationFlag{
			Name:  "server-timeout",
			Usage: "请求超时时间",
			Value: Defaults.Server.Timeout,
		},
		// 认证配置
		&cli.StringFlag{
			Name:  "auth-type",
			Usage: "认证类型: basic, bearer",
		},
		&cli.StringFlag{
			Name:  "auth-user",
			Usage: "Basic 认证用户名",
		},
		&cli.StringFlag{
			Name:  "auth-password",
			Usage: "Basic 认证密码",
		},
		&cli.StringFlag{
			Name:  "auth-token",
			Usage: "Bearer Token",
		},
		// TLS 配置
		&cli.StringFlag{
			Name:  "tls-ca",
			Usage: "CA 证书路径",
		},
		&cli.StringFlag{
			Name:  "tls-cert",
			Usage: "客户端证书路径",
		},
		&cli.StringFlag{
			Name:  "tls-key",
			Usage: "客户端密钥路径",
		},
		&cli.BoolFlag{
			Name:  "tls-skip-verify",
			Usage: "跳过证书验证",
			Value: Defaults.TLS.SkipVerify,
		},
	}
}

// NewClient 从配置创建 vmapi 客户端
func NewClient(cfg *config.Config) (vmapi.Client, error) {
	return vmapi.NewClient(&vmapi.ClientConfig{
		URL:        cfg.Server.URL,
		PathPrefix: cfg.Server.PathPrefix,
		Timeout:    cfg.Server.Timeout,
		AuthType:   cfg.Auth.Type,
		User:       cfg.Auth.User,
		Password:   cfg.Auth.Password,
		Token:      cfg.Auth.Token,
		CAPath:     cfg.TLS.CA,
		CertPath:   cfg.TLS.Cert,
		KeyPath:    cfg.TLS.Key,
		SkipVerify: cfg.TLS.SkipVerify,
	})
}

// ParseTime 解析时间字符串，支持多种格式
// - 空字符串: 返回零值
// - "now": 返回当前时间
// - RFC3339: 如 "2024-01-01T00:00:00Z"
// - Unix 时间戳: 如 "1704067200"
func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if s == "now" {
		return time.Now(), nil
	}
	// 尝试 RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// 尝试 Unix 时间戳
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err == nil {
		return time.Unix(ts, 0), nil
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s (use RFC3339 or Unix timestamp)", s)
}
