package output

import (
	"encoding/json"

	"github.com/lwmacct/251203-mc-metrics/internal/vmapi"
)

// jsonWriter JSON 格式输出
type jsonWriter struct {
	opts Options
}

// NewJSONWriter 创建 JSON 输出 Writer
func NewJSONWriter(opts Options) Writer {
	return &jsonWriter{opts: opts}
}

// WriteQueryResult 输出查询结果
func (w *jsonWriter) WriteQueryResult(result *vmapi.QueryResult) error {
	return w.writeJSON(result)
}

// WriteStrings 输出字符串列表
func (w *jsonWriter) WriteStrings(items []string) error {
	return w.writeJSON(items)
}

// WriteSeries 输出时间序列
func (w *jsonWriter) WriteSeries(series []vmapi.LabelSet) error {
	return w.writeJSON(series)
}

// writeJSON 通用 JSON 输出
func (w *jsonWriter) writeJSON(v any) error {
	enc := json.NewEncoder(w.opts.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
