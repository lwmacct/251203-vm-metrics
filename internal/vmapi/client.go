// Package vmapi 提供 VictoriaMetrics HTTP API 客户端
package vmapi

import (
	"context"
	"time"
)

// Client VictoriaMetrics API 客户端接口
type Client interface {
	// Query 执行即时查询
	// endpoint: GET /api/v1/query
	Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error)

	// QueryRange 执行范围查询
	// endpoint: GET /api/v1/query_range
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error)

	// Series 获取时间序列
	// endpoint: GET /api/v1/series
	Series(ctx context.Context, match []string, start, end time.Time) (*SeriesResult, error)

	// Labels 获取所有标签名称
	// endpoint: GET /api/v1/labels
	Labels(ctx context.Context, start, end time.Time) (*LabelsResult, error)

	// LabelValues 获取指定标签的所有值
	// endpoint: GET /api/v1/label/<name>/values
	LabelValues(ctx context.Context, label string, start, end time.Time) (*LabelValuesResult, error)
}

// ClientConfig 客户端配置
type ClientConfig struct {
	// 服务器配置
	URL        string
	PathPrefix string // API 路径前缀 (如 /victoria)
	Timeout    time.Duration

	// 认证配置
	AuthType string // "basic" | "bearer"
	User     string
	Password string
	Token    string

	// TLS 配置
	CAPath     string
	CertPath   string
	KeyPath    string
	SkipVerify bool
}
