# promql-cli 命令详解

<!--TOC-->

- [全局参数](#全局参数) `:25+23`
  - [TLS 配置参数](#tls-配置参数) `:38+10`
- [主命令 - 查询](#主命令-查询) `:48+79`
  - [Instant Query (即时查询)](#instant-query-即时查询) `:50+35`
  - [Range Query (范围查询)](#range-query-范围查询) `:85+42`
- [metrics 命令](#metrics-命令) `:127+34`
- [labels 命令](#labels-命令) `:161+29`
- [meta 命令](#meta-命令) `:190+27`
- [输出格式](#输出格式) `:217+37`
  - [Table (默认)](#table-默认) `:219+10`
  - [JSON](#json) `:229+10`
  - [CSV](#csv) `:239+11`
  - [ASCII Graph](#ascii-graph) `:250+4`
- [从文件执行查询](#从文件执行查询) `:254+14`

<!--TOC-->

<!--TOC-->

## 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--host` | Prometheus 服务器 URL | `http://0.0.0.0:9090` |
| `--config` | 配置文件路径 | `$HOME/.promql-cli.yaml` |
| `--timeout` | 查询超时时间 (秒) | `10` |
| `--output` | 输出格式 (json/csv) | 根据查询类型自动选择 |
| `--no-headers` | 禁用表头 | `false` |
| `--auth-type` | 认证类型 (Basic/Bearer) | - |
| `--auth-credentials` | 认证凭证字符串 | - |
| `--auth-credentials-file` | 认证凭证文件路径 | - |

### TLS 配置参数

| 参数 | 说明 |
|------|------|
| `--tls_config.ca_cert_file` | CA 证书文件路径 |
| `--tls_config.cert_file` | 客户端证书文件路径 |
| `--tls_config.key_file` | 客户端密钥文件路径 |
| `--tls_config.servername` | 服务器名称 |
| `--tls_config.insecure_skip_verify` | 跳过 TLS 验证 |

## 主命令 - 查询

### Instant Query (即时查询)

```bash
promql [query_string]
```

执行即时查询，返回当前时间点的指标值。

**参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--time` | 查询时间点 (now 或 RFC3339) | `now` |

**示例：**

```bash
# 查询所有 up 指标
promql 'up'

# 按 job 聚合
promql 'sum(up) by (job)'

# 指定时间点
promql 'up' --time "2024-01-15T10:00:00Z"
```

**输出示例：**

```
INSTANCE                VALUE                 TIMESTAMP
123.456.789.123:6443    14.868565474122951    2020-09-27T09:34:22-04:00
234.567.891.234:6443    9.148373277758477     2020-09-27T09:34:22-04:00
```

### Range Query (范围查询)

```bash
promql [query_string] --start <duration>
```

执行范围查询，返回时间序列数据。

**参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--start` | 开始时间 (lookback 如 1h 或 RFC3339) | - |
| `--end` | 结束时间 (now 或 RFC3339) | `now` |
| `--step` | 数据点间隔 | `1m` |

**示例：**

```bash
# 过去 1 小时
promql 'sum(up) by (job)' --start 1h

# 指定时间范围
promql 'up' --start "2024-01-15T00:00:00Z" --end "2024-01-15T12:00:00Z"

# 自定义步长
promql 'rate(http_requests_total[5m])' --start 24h --step 5m
```

**输出示例 (ASCII 图表)：**

```
##################################################
# TIME_RANGE: Sep 26 09:37:35 -> Sep 27 09:37:35 #
# METRIC: {job="apiserver"}                      #
##################################################
 38.31 ┤          ╭╮
 38.04 ┤          ││   ╭╮                            ╭╮╭╮
 37.76 ┤        ╭─╯╰╮ ╭╯╰─╮                   ╭╮╭───╮││││
 ...
```

## metrics 命令

```bash
promql metrics [query_string]
```

列出所有可用的指标名称。

**API 端点：** `/api/v1/series`

**示例：**

```bash
# 列出所有指标
promql metrics

# 按模式过滤
promql metrics '{__name__=~".+gc.+"}'

# JSON 输出
promql metrics --output json
```

**输出示例：**

```
METRICS
go_gc_duration_seconds
go_gc_duration_seconds_count
go_gc_duration_seconds_sum
go_goroutines
go_info
```

## labels 命令

```bash
promql labels [query_string]
```

获取给定查询结果中的所有标签名。

**API 端点：** `/api/v1/query` (解析结果中的标签)

**示例：**

```bash
promql labels apiserver_request_total
```

**输出示例：**

```
LABELS
__name__
client
code
component
endpoint
instance
job
```

## meta 命令

```bash
promql meta [metric_name]
```

获取指标的类型和帮助元数据。

**API 端点：** `/api/v1/metadata`

**示例：**

```bash
# 查询特定指标元数据
promql meta go_goroutines

# 查询所有元数据
promql meta
```

**输出示例：**

```
METRIC           TYPE     HELP                                          UNIT
go_goroutines    gauge    Number of goroutines that currently exist.
```

## 输出格式

### Table (默认)

Instant 查询默认使用表格格式：

```
INSTANCE    VALUE    TIMESTAMP
node1       1.0      2024-01-15T10:00:00Z
node2       1.0      2024-01-15T10:00:00Z
```

### JSON

```bash
promql 'up' --output json
```

```json
[{"metric":{"instance":"node1"},"value":[1705312800,"1"]}]
```

### CSV

```bash
promql 'up' --output csv
```

```csv
instance,value,timestamp
node1,1.0,2024-01-15T10:00:00Z
```

### ASCII Graph

Range 查询默认使用 ASCII 图表，由 [asciigraph](https://github.com/guptarohit/asciigraph) 生成。

## 从文件执行查询

支持从文件读取复杂查询：

```bash
# my-query.promql
sum(
  rate(http_requests_total{job="api"}[5m])
) by (status_code)
```

```bash
promql "$(cat ./my-query.promql)" --start 1h
```
