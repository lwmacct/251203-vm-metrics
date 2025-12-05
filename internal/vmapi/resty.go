package vmapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// restyClient go-resty 实现的 VictoriaMetrics 客户端
type restyClient struct {
	client  *resty.Client
	baseURL string
}

// NewClient 创建新的 VictoriaMetrics 客户端
func NewClient(cfg *ClientConfig) (Client, error) {
	// 构建 baseURL，支持路径前缀
	baseURL := strings.TrimSuffix(cfg.URL, "/")
	if cfg.PathPrefix != "" {
		baseURL = baseURL + "/" + strings.Trim(cfg.PathPrefix, "/")
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(cfg.Timeout).
		SetHeader("Accept", "application/json").
		SetDisableWarn(true)

	// 配置认证
	switch cfg.AuthType {
	case "basic":
		client.SetBasicAuth(cfg.User, cfg.Password)
	case "bearer":
		client.SetAuthToken(cfg.Token)
	}

	// 配置 TLS
	if cfg.CAPath != "" || cfg.CertPath != "" || cfg.SkipVerify {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	return &restyClient{
		client:  client,
		baseURL: cfg.URL,
	}, nil
}

// buildTLSConfig 构建 TLS 配置
func buildTLSConfig(cfg *ClientConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify, //nolint:gosec // 用户明确请求跳过验证
	}

	// 加载 CA 证书
	if cfg.CAPath != "" {
		caCert, err := os.ReadFile(cfg.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 加载客户端证书
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// Query 执行即时查询
func (c *restyClient) Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error) {
	params := map[string]string{
		"query": query,
	}
	if !ts.IsZero() {
		params["time"] = formatTime(ts)
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(params).
		Get("/api/v1/query")
	if err != nil {
		return nil, fmt.Errorf("query request failed: %w", err)
	}

	return parseQueryResponse(resp.Body())
}

// QueryRange 执行范围查询
func (c *restyClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error) {
	params := map[string]string{
		"query": query,
		"start": formatTime(start),
		"end":   formatTime(end),
		"step":  formatDuration(step),
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(params).
		Get("/api/v1/query_range")
	if err != nil {
		return nil, fmt.Errorf("query_range request failed: %w", err)
	}

	return parseQueryResponse(resp.Body())
}

// Series 获取时间序列
func (c *restyClient) Series(ctx context.Context, match []string, start, end time.Time) (*SeriesResult, error) {
	req := c.client.R().SetContext(ctx)

	// match[] 参数可以有多个
	for _, m := range match {
		req.QueryParam.Add("match[]", m)
	}

	if !start.IsZero() {
		req.SetQueryParam("start", formatTime(start))
	}
	if !end.IsZero() {
		req.SetQueryParam("end", formatTime(end))
	}

	resp, err := req.Get("/api/v1/series")
	if err != nil {
		return nil, fmt.Errorf("series request failed: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.ErrorType, apiResp.Error)
	}

	var series []LabelSet
	if err := json.Unmarshal(apiResp.Data, &series); err != nil {
		return nil, fmt.Errorf("failed to parse series data: %w", err)
	}

	return &SeriesResult{Series: series}, nil
}

// Labels 获取所有标签名称
func (c *restyClient) Labels(ctx context.Context, start, end time.Time) (*LabelsResult, error) {
	req := c.client.R().SetContext(ctx)

	if !start.IsZero() {
		req.SetQueryParam("start", formatTime(start))
	}
	if !end.IsZero() {
		req.SetQueryParam("end", formatTime(end))
	}

	resp, err := req.Get("/api/v1/labels")
	if err != nil {
		return nil, fmt.Errorf("labels request failed: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.ErrorType, apiResp.Error)
	}

	var labels []string
	if err := json.Unmarshal(apiResp.Data, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse labels data: %w", err)
	}

	return &LabelsResult{Labels: labels}, nil
}

// LabelValues 获取指定标签的所有值
func (c *restyClient) LabelValues(ctx context.Context, label string, start, end time.Time) (*LabelValuesResult, error) {
	req := c.client.R().SetContext(ctx)

	if !start.IsZero() {
		req.SetQueryParam("start", formatTime(start))
	}
	if !end.IsZero() {
		req.SetQueryParam("end", formatTime(end))
	}

	resp, err := req.Get(fmt.Sprintf("/api/v1/label/%s/values", label))
	if err != nil {
		return nil, fmt.Errorf("label_values request failed: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.ErrorType, apiResp.Error)
	}

	var values []string
	if err := json.Unmarshal(apiResp.Data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse label values data: %w", err)
	}

	return &LabelValuesResult{Values: values}, nil
}

// parseQueryResponse 解析查询响应
func parseQueryResponse(body []byte) (*QueryResult, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.ErrorType, apiResp.Error)
	}

	var queryData QueryData
	if err := json.Unmarshal(apiResp.Data, &queryData); err != nil {
		return nil, fmt.Errorf("failed to parse query data: %w", err)
	}

	result := &QueryResult{
		ResultType: queryData.ResultType,
	}

	switch queryData.ResultType {
	case "vector", "matrix":
		var samples []Sample
		if err := json.Unmarshal(queryData.Result, &samples); err != nil {
			return nil, fmt.Errorf("failed to parse samples: %w", err)
		}
		result.Samples = samples

	case "scalar":
		var sv SampleValue
		if err := json.Unmarshal(queryData.Result, &sv); err != nil {
			return nil, fmt.Errorf("failed to parse scalar: %w", err)
		}
		result.Scalar = &sv

	case "string":
		var raw []any
		if err := json.Unmarshal(queryData.Result, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse string result: %w", err)
		}
		if len(raw) == 2 {
			ts, _ := raw[0].(float64)
			val, _ := raw[1].(string)
			result.String = &StringResult{
				Timestamp: time.Unix(int64(ts), 0),
				Value:     val,
			}
		}
	}

	return result, nil
}

// formatTime 格式化时间为 Unix 时间戳字符串
func formatTime(t time.Time) string {
	return strconv.FormatFloat(float64(t.UnixNano())/1e9, 'f', 3, 64)
}

// formatDuration 格式化 Duration 为秒数字符串
func formatDuration(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 0, 64)
}
