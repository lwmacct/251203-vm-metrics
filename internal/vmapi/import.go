package vmapi

import (
	"context"
	"fmt"
	"io"
)

// ImportFormat 导入格式
type ImportFormat string

const (
	ImportFormatJSON       ImportFormat = "json"
	ImportFormatCSV        ImportFormat = "csv"
	ImportFormatNative     ImportFormat = "native"
	ImportFormatPrometheus ImportFormat = "prometheus"
)

// ImportOptions 导入选项
type ImportOptions struct {
	// Prometheus 格式专用
	Job      string // Pushgateway job 标签
	Instance string // Pushgateway instance 标签
}

// Importer 导入接口
type Importer interface {
	// ImportJSON 导入 JSON Line 格式
	ImportJSON(ctx context.Context, r io.Reader) error

	// ImportCSV 导入 CSV 格式
	ImportCSV(ctx context.Context, r io.Reader) error

	// ImportNative 导入 Native 二进制格式
	ImportNative(ctx context.Context, r io.Reader) error

	// ImportPrometheus 导入 Prometheus exposition 格式
	ImportPrometheus(ctx context.Context, r io.Reader, opts *ImportOptions) error
}

// ImportJSON 导入 JSON Line 格式
func (c *restyClient) ImportJSON(ctx context.Context, r io.Reader) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(r).
		SetHeader("Content-Type", "application/json").
		Post("/api/v1/import")
	if err != nil {
		return fmt.Errorf("import json request failed: %w", err)
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
		return fmt.Errorf("import json failed [%d]: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// ImportCSV 导入 CSV 格式
func (c *restyClient) ImportCSV(ctx context.Context, r io.Reader) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(r).
		SetHeader("Content-Type", "text/csv").
		Post("/api/v1/import/csv")
	if err != nil {
		return fmt.Errorf("import csv request failed: %w", err)
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
		return fmt.Errorf("import csv failed [%d]: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// ImportNative 导入 Native 二进制格式
func (c *restyClient) ImportNative(ctx context.Context, r io.Reader) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(r).
		SetHeader("Content-Type", "application/octet-stream").
		Post("/api/v1/import/native")
	if err != nil {
		return fmt.Errorf("import native request failed: %w", err)
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
		return fmt.Errorf("import native failed [%d]: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// ImportPrometheus 导入 Prometheus exposition 格式
func (c *restyClient) ImportPrometheus(ctx context.Context, r io.Reader, opts *ImportOptions) error {
	var endpoint string
	if opts != nil && opts.Job != "" {
		// Pushgateway 兼容端点
		endpoint = fmt.Sprintf("/api/v1/import/prometheus/metrics/job/%s", opts.Job)
		if opts.Instance != "" {
			endpoint += fmt.Sprintf("/instance/%s", opts.Instance)
		}
	} else {
		endpoint = "/api/v1/import/prometheus"
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(r).
		SetHeader("Content-Type", "text/plain").
		Post(endpoint)
	if err != nil {
		return fmt.Errorf("import prometheus request failed: %w", err)
	}

	if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
		return fmt.Errorf("import prometheus failed [%d]: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
