package output

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/lwmacct/251203-mc-metrics/internal/vmapi"
)

// tableWriter tabwriter 实现的表格输出
type tableWriter struct {
	opts Options
}

// NewTableWriter 创建表格输出 Writer
func NewTableWriter(opts Options) Writer {
	return &tableWriter{opts: opts}
}

// WriteQueryResult 输出查询结果
func (w *tableWriter) WriteQueryResult(result *vmapi.QueryResult) error {
	tw := tabwriter.NewWriter(w.opts.Writer, 0, 0, 2, ' ', 0)
	defer func() { _ = tw.Flush() }()

	switch result.ResultType {
	case "vector":
		return w.writeVector(tw, result.Samples)
	case "matrix":
		return w.writeMatrix(tw, result.Samples)
	case "scalar":
		if result.Scalar != nil {
			_, _ = fmt.Fprintf(tw, "%v\t@%s\n", result.Scalar.Value, result.Scalar.Timestamp.Format(time.RFC3339))
		}
	case "string":
		if result.String != nil {
			_, _ = fmt.Fprintf(tw, "%s\t@%s\n", result.String.Value, result.String.Timestamp.Format(time.RFC3339))
		}
	}

	return nil
}

// writeVector 输出 instant query 结果
func (w *tableWriter) writeVector(tw *tabwriter.Writer, samples []vmapi.Sample) error {
	if !w.opts.NoHeaders {
		_, _ = fmt.Fprintln(tw, "METRIC\tVALUE\tTIMESTAMP")
	}

	for _, s := range samples {
		metric := formatMetric(s.Metric)
		_, _ = fmt.Fprintf(tw, "%s\t%v\t%s\n",
			metric,
			s.Value.Value,
			s.Value.Timestamp.Format(time.RFC3339),
		)
	}

	return nil
}

// writeMatrix 输出 range query 结果
func (w *tableWriter) writeMatrix(tw *tabwriter.Writer, samples []vmapi.Sample) error {
	if !w.opts.NoHeaders {
		_, _ = fmt.Fprintln(tw, "METRIC\tVALUE\tTIMESTAMP")
	}

	for _, s := range samples {
		metric := formatMetric(s.Metric)
		for _, v := range s.Values {
			_, _ = fmt.Fprintf(tw, "%s\t%v\t%s\n",
				metric,
				v.Value,
				v.Timestamp.Format(time.RFC3339),
			)
		}
	}

	return nil
}

// WriteStrings 输出字符串列表
func (w *tableWriter) WriteStrings(items []string) error {
	for _, item := range items {
		_, _ = fmt.Fprintln(w.opts.Writer, item)
	}
	return nil
}

// WriteSeries 输出时间序列
func (w *tableWriter) WriteSeries(series []vmapi.LabelSet) error {
	tw := tabwriter.NewWriter(w.opts.Writer, 0, 0, 2, ' ', 0)
	defer func() { _ = tw.Flush() }()

	for _, s := range series {
		_, _ = fmt.Fprintln(tw, formatMetric(s))
	}

	return nil
}

// formatMetric 格式化 metric 标签为 Prometheus 格式
// 例如: metric_name{label1="value1", label2="value2"}
func formatMetric(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	name := labels["__name__"]

	// 收集非 __name__ 的标签
	var pairs []string
	for k, v := range labels {
		if k != "__name__" {
			pairs = append(pairs, fmt.Sprintf("%s=%q", k, v))
		}
	}

	// 按字母顺序排序
	sort.Strings(pairs)

	if len(pairs) == 0 {
		return name
	}

	return name + "{" + strings.Join(pairs, ", ") + "}"
}
