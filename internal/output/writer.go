// Package output 提供查询结果的格式化输出
package output

import (
	"fmt"
	"io"
	"os"

	"github.com/lwmacct/251203-mc-metrics/internal/vmapi"
)

// Writer 输出格式化接口
type Writer interface {
	// WriteQueryResult 输出查询结果 (vector/matrix/scalar)
	WriteQueryResult(result *vmapi.QueryResult) error

	// WriteStrings 输出字符串列表 (labels, metrics, label-values)
	WriteStrings(items []string) error

	// WriteSeries 输出时间序列列表
	WriteSeries(series []vmapi.LabelSet) error
}

// Options 输出选项
type Options struct {
	Writer    io.Writer // 输出目标，默认 os.Stdout
	NoHeaders bool      // 禁用表头 (table/csv)
	NoColor   bool      // 禁用颜色 (table)
}

// DefaultOptions 返回默认输出选项
func DefaultOptions() Options {
	return Options{
		Writer:    os.Stdout,
		NoHeaders: false,
		NoColor:   false,
	}
}

// New 根据格式创建 Writer
func New(format string, opts Options) (Writer, error) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch format {
	case "json":
		return NewJSONWriter(opts), nil
	case "csv":
		return NewCSVWriter(opts), nil
	case "graph":
		return NewGraphWriter(opts), nil
	case "table", "":
		return NewTableWriter(opts), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
