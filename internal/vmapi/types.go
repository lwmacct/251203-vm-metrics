// Package vmapi 提供 VictoriaMetrics HTTP API 客户端
package vmapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// APIResponse VictoriaMetrics API 统一响应结构
type APIResponse struct {
	Status    string          `json:"status"`              // "success" | "error"
	Data      json.RawMessage `json:"data,omitempty"`      // 响应数据
	ErrorType string          `json:"errorType,omitempty"` // 错误类型
	Error     string          `json:"error,omitempty"`     // 错误信息
	Warnings  []string        `json:"warnings,omitempty"`  // 警告信息
}

// IsSuccess 检查响应是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.Status == "success"
}

// QueryData 查询结果数据
type QueryData struct {
	ResultType string          `json:"resultType"` // "vector" | "matrix" | "scalar" | "string"
	Result     json.RawMessage `json:"result"`     // 结果数据
}

// Sample 单个样本点
type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  SampleValue       `json:"value"`            // Instant query
	Values []SampleValue     `json:"values,omitempty"` // Range query
}

// SampleValue 样本值 [timestamp, value]
type SampleValue struct {
	Timestamp time.Time
	Value     float64
}

// UnmarshalJSON 自定义 JSON 解析
func (sv *SampleValue) UnmarshalJSON(data []byte) error {
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("invalid sample value: expected 2 elements, got %d", len(raw))
	}

	// 解析时间戳 (Unix timestamp)
	switch ts := raw[0].(type) {
	case float64:
		sv.Timestamp = time.Unix(int64(ts), int64((ts-float64(int64(ts)))*1e9))
	default:
		return fmt.Errorf("invalid timestamp type: %T", raw[0])
	}

	// 解析值 (字符串形式的浮点数)
	switch v := raw[1].(type) {
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("invalid value: %w", err)
		}
		sv.Value = val
	default:
		return fmt.Errorf("invalid value type: %T", raw[1])
	}

	return nil
}

// QueryResult 查询结果
type QueryResult struct {
	ResultType string   // "vector" | "matrix" | "scalar" | "string"
	Samples    []Sample // 样本数据
	Scalar     *SampleValue
	String     *StringResult
}

// StringResult 字符串结果
type StringResult struct {
	Timestamp time.Time
	Value     string
}

// LabelSet 标签集合
type LabelSet map[string]string

// SeriesResult 系列查询结果
type SeriesResult struct {
	Series []LabelSet
}

// LabelsResult 标签名称列表
type LabelsResult struct {
	Labels []string
}

// LabelValuesResult 标签值列表
type LabelValuesResult struct {
	Values []string
}
