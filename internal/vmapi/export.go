package vmapi

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	ExportFormatJSON   ExportFormat = "json"
	ExportFormatCSV    ExportFormat = "csv"
	ExportFormatNative ExportFormat = "native"
)

// ExportOptions 导出选项
type ExportOptions struct {
	Match           []string      // 时间序列选择器
	Start           time.Time     // 开始时间
	End             time.Time     // 结束时间
	MaxRowsPerLine  int           // JSON Line 每行最大样本数
	CSVFormat       string        // CSV 列定义
	ReduceMemUsage  bool          // 跳过去重
}

// Exporter 导出接口
type Exporter interface {
	// ExportJSON 导出 JSON Line 格式
	ExportJSON(ctx context.Context, w io.Writer, opts *ExportOptions) error

	// ExportCSV 导出 CSV 格式
	ExportCSV(ctx context.Context, w io.Writer, opts *ExportOptions) error

	// ExportNative 导出 Native 二进制格式
	ExportNative(ctx context.Context, w io.Writer, opts *ExportOptions) error
}

// ExportJSON 导出 JSON Line 格式到 Writer
func (c *restyClient) ExportJSON(ctx context.Context, w io.Writer, opts *ExportOptions) error {
	req := c.client.R().
		SetContext(ctx).
		SetDoNotParseResponse(true) // 流式响应

	// 添加 match[] 参数
	for _, m := range opts.Match {
		req.QueryParam.Add("match[]", m)
	}

	if !opts.Start.IsZero() {
		req.SetQueryParam("start", formatTime(opts.Start))
	}
	if !opts.End.IsZero() {
		req.SetQueryParam("end", formatTime(opts.End))
	}
	if opts.MaxRowsPerLine > 0 {
		req.SetQueryParam("max_rows_per_line", fmt.Sprintf("%d", opts.MaxRowsPerLine))
	}

	resp, err := req.Get("/api/v1/export")
	if err != nil {
		return fmt.Errorf("export request failed: %w", err)
	}
	defer func() { _ = resp.RawBody().Close() }()

	if resp.StatusCode() != 200 {
		body, _ := io.ReadAll(resp.RawBody())
		return fmt.Errorf("export failed: %s", string(body))
	}

	_, err = io.Copy(w, resp.RawBody())
	return err
}

// ExportCSV 导出 CSV 格式到 Writer
func (c *restyClient) ExportCSV(ctx context.Context, w io.Writer, opts *ExportOptions) error {
	req := c.client.R().
		SetContext(ctx).
		SetDoNotParseResponse(true)

	// format 参数必需
	if opts.CSVFormat == "" {
		opts.CSVFormat = "__name__,__value__,__timestamp__:unix_s"
	}
	req.SetQueryParam("format", opts.CSVFormat)

	// 添加 match[] 参数
	for _, m := range opts.Match {
		req.QueryParam.Add("match[]", m)
	}

	if !opts.Start.IsZero() {
		req.SetQueryParam("start", formatTime(opts.Start))
	}
	if !opts.End.IsZero() {
		req.SetQueryParam("end", formatTime(opts.End))
	}
	if opts.ReduceMemUsage {
		req.SetQueryParam("reduce_mem_usage", "1")
	}

	resp, err := req.Get("/api/v1/export/csv")
	if err != nil {
		return fmt.Errorf("export csv request failed: %w", err)
	}
	defer func() { _ = resp.RawBody().Close() }()

	if resp.StatusCode() != 200 {
		body, _ := io.ReadAll(resp.RawBody())
		return fmt.Errorf("export csv failed: %s", string(body))
	}

	_, err = io.Copy(w, resp.RawBody())
	return err
}

// ExportNative 导出 Native 二进制格式到 Writer
func (c *restyClient) ExportNative(ctx context.Context, w io.Writer, opts *ExportOptions) error {
	req := c.client.R().
		SetContext(ctx).
		SetDoNotParseResponse(true)

	// 添加 match[] 参数
	for _, m := range opts.Match {
		req.QueryParam.Add("match[]", m)
	}

	if !opts.Start.IsZero() {
		req.SetQueryParam("start", formatTime(opts.Start))
	}
	if !opts.End.IsZero() {
		req.SetQueryParam("end", formatTime(opts.End))
	}

	resp, err := req.Get("/api/v1/export/native")
	if err != nil {
		return fmt.Errorf("export native request failed: %w", err)
	}
	defer func() { _ = resp.RawBody().Close() }()

	if resp.StatusCode() != 200 {
		body, _ := io.ReadAll(resp.RawBody())
		return fmt.Errorf("export native failed: %s", string(body))
	}

	_, err = io.Copy(w, resp.RawBody())
	return err
}
