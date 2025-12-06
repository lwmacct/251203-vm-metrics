package output

import (
	"fmt"

	"github.com/guptarohit/asciigraph"
	"github.com/lwmacct/251203-vm-metrics/internal/vmapi"
)

// graphWriter ASCII 图表输出
type graphWriter struct {
	opts Options
}

// NewGraphWriter 创建 ASCII 图表输出 Writer
func NewGraphWriter(opts Options) Writer {
	return &graphWriter{opts: opts}
}

// WriteQueryResult 输出查询结果为 ASCII 图表
// 仅支持 matrix 类型 (range query)，其他类型回退到 table 格式
func (w *graphWriter) WriteQueryResult(result *vmapi.QueryResult) error {
	if result.ResultType != "matrix" || len(result.Samples) == 0 {
		// 非 matrix 类型或无数据，回退到 table
		return NewTableWriter(w.opts).WriteQueryResult(result)
	}

	// 为每个时间序列绘制图表
	for _, sample := range result.Samples {
		if len(sample.Values) == 0 {
			continue
		}

		// 提取数值
		data := make([]float64, len(sample.Values))
		for i, v := range sample.Values {
			data[i] = v.Value
		}

		// 绘制图表
		metric := formatMetric(sample.Metric)
		graph := asciigraph.Plot(data,
			asciigraph.Caption(metric),
			asciigraph.Height(10),
			asciigraph.Width(60),
		)

		_, _ = fmt.Fprintln(w.opts.Writer, graph)
		_, _ = fmt.Fprintln(w.opts.Writer) // 空行分隔
	}

	return nil
}

// WriteStrings 字符串列表不支持图表，回退到 table
func (w *graphWriter) WriteStrings(items []string) error {
	return NewTableWriter(w.opts).WriteStrings(items)
}

// WriteSeries 时间序列列表不支持图表，回退到 table
func (w *graphWriter) WriteSeries(series []vmapi.LabelSet) error {
	return NewTableWriter(w.opts).WriteSeries(series)
}
