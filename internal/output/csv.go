package output

import (
	"encoding/csv"
	"fmt"
	"time"

	"github.com/lwmacct/251203-mc-metrics/internal/vmapi"
)

// csvWriter CSV 格式输出
type csvWriter struct {
	opts Options
}

// NewCSVWriter 创建 CSV 输出 Writer
func NewCSVWriter(opts Options) Writer {
	return &csvWriter{opts: opts}
}

// WriteQueryResult 输出查询结果
func (w *csvWriter) WriteQueryResult(result *vmapi.QueryResult) error {
	cw := csv.NewWriter(w.opts.Writer)
	defer cw.Flush()

	switch result.ResultType {
	case "vector":
		return w.writeVector(cw, result.Samples)
	case "matrix":
		return w.writeMatrix(cw, result.Samples)
	case "scalar":
		if result.Scalar != nil {
			if !w.opts.NoHeaders {
				_ = cw.Write([]string{"value", "timestamp"})
			}
			_ = cw.Write([]string{
				fmt.Sprintf("%v", result.Scalar.Value),
				result.Scalar.Timestamp.Format(time.RFC3339),
			})
		}
	case "string":
		if result.String != nil {
			if !w.opts.NoHeaders {
				_ = cw.Write([]string{"value", "timestamp"})
			}
			_ = cw.Write([]string{
				result.String.Value,
				result.String.Timestamp.Format(time.RFC3339),
			})
		}
	}

	return cw.Error()
}

// writeVector 输出 instant query 结果
func (w *csvWriter) writeVector(cw *csv.Writer, samples []vmapi.Sample) error {
	if !w.opts.NoHeaders {
		_ = cw.Write([]string{"metric", "value", "timestamp"})
	}

	for _, s := range samples {
		_ = cw.Write([]string{
			formatMetric(s.Metric),
			fmt.Sprintf("%v", s.Value.Value),
			s.Value.Timestamp.Format(time.RFC3339),
		})
	}

	return cw.Error()
}

// writeMatrix 输出 range query 结果
func (w *csvWriter) writeMatrix(cw *csv.Writer, samples []vmapi.Sample) error {
	if !w.opts.NoHeaders {
		_ = cw.Write([]string{"metric", "value", "timestamp"})
	}

	for _, s := range samples {
		metric := formatMetric(s.Metric)
		for _, v := range s.Values {
			_ = cw.Write([]string{
				metric,
				fmt.Sprintf("%v", v.Value),
				v.Timestamp.Format(time.RFC3339),
			})
		}
	}

	return cw.Error()
}

// WriteStrings 输出字符串列表
func (w *csvWriter) WriteStrings(items []string) error {
	cw := csv.NewWriter(w.opts.Writer)
	defer cw.Flush()

	if !w.opts.NoHeaders {
		_ = cw.Write([]string{"name"})
	}

	for _, item := range items {
		_ = cw.Write([]string{item})
	}

	return cw.Error()
}

// WriteSeries 输出时间序列
func (w *csvWriter) WriteSeries(series []vmapi.LabelSet) error {
	cw := csv.NewWriter(w.opts.Writer)
	defer cw.Flush()

	if !w.opts.NoHeaders {
		_ = cw.Write([]string{"series"})
	}

	for _, s := range series {
		_ = cw.Write([]string{formatMetric(s)})
	}

	return cw.Error()
}
