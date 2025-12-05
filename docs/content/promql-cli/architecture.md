# promql-cli 架构分析

<!--TOC-->

- [项目结构](#项目结构) `:22+22`
- [核心模块](#核心模块) `:44+112`
  - [cmd 包 - CLI 命令定义](#cmd-包-cli-命令定义) `:46+30`
  - [pkg/promql - API 客户端](#pkgpromql-api-客户端) `:76+33`
  - [pkg/writer - 输出格式化](#pkgwriter-输出格式化) `:109+37`
  - [pkg/util - 工具函数](#pkgutil-工具函数) `:146+10`
- [依赖关系](#依赖关系) `:156+13`
- [数据流](#数据流) `:169+33`
- [设计特点](#设计特点) `:202+15`
  - [优点](#优点) `:204+7`
  - [可改进点](#可改进点) `:211+6`

<!--TOC-->

<!--TOC-->

## 项目结构

```
promql-cli/
├── main.go                 # 程序入口
├── cmd/
│   ├── root.go            # 主命令 + Instant/Range 查询
│   ├── labels.go          # labels 子命令
│   ├── metrics.go         # metrics 子命令
│   └── meta.go            # meta 子命令
├── pkg/
│   ├── promql/
│   │   └── promql.go      # Prometheus API 客户端封装
│   ├── util/
│   │   └── util.go        # 工具函数
│   └── writer/
│       ├── writer.go      # 输出格式化
│       └── writer_test.go
├── go.mod
└── Makefile
```

## 核心模块

### cmd 包 - CLI 命令定义

基于 [spf13/cobra](https://github.com/spf13/cobra) 实现命令行解析。

#### root.go 关键代码

```go
// 主命令结构
var rootCmd = &cobra.Command{
    Version: "v0.2.1",
    Use:     "promql [query_string]",
    Short:   "Query prometheus from the command line",
    Args:    cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // 根据 --start 参数判断查询类型
        if pql.Start != "" {
            // Range Query
            result, warnings, err := pql.RangeQuery(query)
            r := writer.RangeResult{Matrix: result}
            writer.WriteRange(&r, pql.Output, pql.NoHeaders)
        } else {
            // Instant Query
            result, warnings, err := pql.InstantQuery(query)
            r := writer.InstantResult{Vector: result}
            writer.WriteInstant(&r, pql.Output, pql.NoHeaders)
        }
    },
}
```

### pkg/promql - API 客户端

封装 Prometheus HTTP API 调用，支持认证和 TLS。

#### 核心结构体

```go
type PromQL struct {
    Host            string              // Prometheus 服务器地址
    Step            string              // 查询步长
    Output          string              // 输出格式
    TimeoutDuration time.Duration       // 超时时间
    CfgFile         string              // 配置文件路径
    Time            time.Time           // 查询时间点
    Start           string              // 范围查询开始时间
    End             string              // 范围查询结束时间
    NoHeaders       bool                // 禁用表头
    Auth            config.Authorization // 认证配置
    Client          v1.API              // Prometheus API 客户端
    TLSConfig       config.TLSConfig    // TLS 配置
}
```

#### API 方法

| 方法 | 功能 | Prometheus API |
|------|------|----------------|
| `InstantQuery()` | 即时查询 | `/api/v1/query` |
| `RangeQuery()` | 范围查询 | `/api/v1/query_range` |
| `LabelsQuery()` | 标签查询 | `/api/v1/query` |
| `MetaQuery()` | 元数据查询 | `/api/v1/metadata` |
| `SeriesQuery()` | 系列查询 | `/api/v1/series` |

### pkg/writer - 输出格式化

定义 Writer 接口，支持多种输出格式。

#### 接口设计

```go
// 基础 Writer 接口
type Writer interface {
    Json() (bytes.Buffer, error)
    Csv(noHeaders bool) (bytes.Buffer, error)
}

// Range 查询专用
type RangeWriter interface {
    Writer
    Graph(dim util.TermDimensions) (bytes.Buffer, error)
}

// Instant 查询专用
type InstantWriter interface {
    Writer
    Table(noHeaders bool) (bytes.Buffer, error)
}
```

#### 结果类型

| 类型 | 描述 | 实现接口 |
|------|------|----------|
| `RangeResult` | 范围查询结果 (model.Matrix) | RangeWriter |
| `InstantResult` | 即时查询结果 (model.Vector) | InstantWriter |
| `LabelsResult` | 标签查询结果 | InstantWriter |
| `MetricsResult` | 指标名列表 | InstantWriter |
| `MetaResult` | 元数据结果 | InstantWriter |
| `SeriesResult` | 系列查询结果 | - |

### pkg/util - 工具函数

```go
// 提取唯一标签名
func UniqLabels(result model.Value) ([]model.LabelName, error)

// 获取终端尺寸 (用于 ASCII 图表)
func TerminalSize() (TermDimensions, error)
```

## 依赖关系

```
main.go
    └── cmd/root.go
            ├── pkg/promql/promql.go
            │       └── prometheus/client_golang (v1.14.0)
            │       └── prometheus/common (v0.37.0)
            └── pkg/writer/writer.go
                    ├── pkg/util/util.go
                    └── guptarohit/asciigraph
```

## 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户输入                                   │
│  promql 'sum(up) by (job)' --start 1h --output json            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    cmd/root.go                                   │
│  1. 解析命令行参数                                                │
│  2. 加载配置文件/环境变量                                         │
│  3. 创建 Prometheus API 客户端                                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 pkg/promql/promql.go                            │
│  1. 构建 HTTP 请求                                               │
│  2. 发送到 Prometheus /api/v1/query_range                       │
│  3. 解析响应为 model.Matrix                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 pkg/writer/writer.go                            │
│  1. 包装为 RangeResult                                          │
│  2. 根据 output 参数选择格式化方法                                │
│  3. 输出到 stdout                                                │
└─────────────────────────────────────────────────────────────────┘
```

## 设计特点

### 优点

1. **模块化设计** - 清晰的包分离
2. **接口抽象** - Writer 接口便于扩展输出格式
3. **配置灵活** - 支持配置文件、环境变量、命令行参数
4. **认证完善** - 支持 Basic/Bearer 认证和 TLS

### 可改进点

1. **依赖较重** - 依赖完整的 prometheus/client_golang
2. **终端尺寸获取** - 使用 `stty size` 命令，可改用 golang.org/x/term
3. **错误处理** - 可以更细粒度的错误类型
4. **测试覆盖** - 只有 writer_test.go
