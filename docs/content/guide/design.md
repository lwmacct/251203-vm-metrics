# vm-metrics 工具集设计方案

<!--TOC-->

- [设计目标](#设计目标) `:41+22`
  - [核心目标](#核心目标) `:43+6`
  - [工具集概览](#工具集概览) `:49+8`
  - [非目标](#非目标) `:57+6`
- [技术选型](#技术选型) `:63+9`
- [配置系统](#配置系统) `:72+83`
  - [加载优先级 (从低到高)](#加载优先级-从低到高) `:74+7`
  - [配置文件搜索路径](#配置文件搜索路径) `:81+9`
  - [配置结构](#配置结构) `:90+19`
  - [配置文件示例 (YAML)](#配置文件示例-yaml) `:109+20`
  - [环境变量映射](#环境变量映射) `:129+13`
  - [CLI 参数映射](#cli-参数映射) `:142+13`
- [命令设计](#命令设计) `:155+121`
  - [mc-vmquery 命令结构](#mc-vmquery-命令结构) `:157+12`
  - [mc-vmexport 命令结构](#mc-vmexport-命令结构) `:169+32`
  - [mc-vmimport 命令结构](#mc-vmimport-命令结构) `:201+31`
  - [全局参数](#全局参数) `:232+27`
  - [使用示例](#使用示例) `:259+17`
- [架构设计](#架构设计) `:276+54`
  - [目录结构](#目录结构) `:278+27`
  - [核心接口](#核心接口) `:305+25`
- [VictoriaMetrics API](#victoriametrics-api) `:330+75`
  - [查询端点 (mc-vmquery)](#查询端点-mc-vmquery) `:332+10`
  - [导出端点 (mc-vmexport)](#导出端点-mc-vmexport) `:342+19`
  - [导入端点 (mc-vmimport)](#导入端点-mc-vmimport) `:361+23`
  - [响应格式](#响应格式) `:384+21`
- [实施路线图](#实施路线图) `:405+39`
  - [Phase 1: 基础架构 (已完成)](#phase-1-基础架构-已完成) `:407+8`
  - [Phase 2: mc-vmquery 完善](#phase-2-mc-vmquery-完善) `:415+10`
  - [Phase 3: mc-vmexport 实现](#phase-3-mc-vmexport-实现) `:425+9`
  - [Phase 4: mc-vmimport 实现](#phase-4-mc-vmimport-实现) `:434+10`
- [与 promql-cli 差异](#与-promql-cli-差异) `:444+15`
- [参考资料](#参考资料) `:459+5`

<!--TOC-->

## 设计目标

### 核心目标

1. **对齐 VictoriaMetrics** - 支持 MetricsQL 查询语言和数据导入/导出
2. **轻量级实现** - 使用 go-resty 替代重量级 Prometheus 客户端
3. **工具集架构** - 三个独立命令：`mc-vmquery`、`mc-vmexport`、`mc-vmimport`

### 工具集概览

| 命令          | 说明                 | 主要用途                  |
| ------------- | -------------------- | ------------------------- |
| `vm-metrics`  | 统一命令行工具       | 整合 query/export/import  |
| `mc-vmquery`  | MetricsQL 查询客户端 | 即时/范围查询、元数据查询 |
| `mc-vmexport` | 数据导出工具         | 导出时序数据到文件/管道   |
| `mc-vmimport` | 数据导入工具         | 从文件/管道导入时序数据   |

### 非目标

- 不实现 PromQL 解析器 (服务端负责)
- 不实现复杂的本地缓存
- 不支持多服务器联邦查询

## 技术选型

| 组件        | 选择                | 理由                             |
| ----------- | ------------------- | -------------------------------- |
| CLI 框架    | `urfave/cli/v3`     | 轻量、功能完整、Context 原生支持 |
| 配置管理    | `koanf`             | 多来源配置合并、类型安全         |
| HTTP 客户端 | `go-resty/resty/v2` | 链式 API、内置认证、自动 JSON    |
| ASCII 图表  | `asciigraph`        | 轻量级、效果好                   |

## 配置系统

### 加载优先级 (从低到高)

1. **默认值** - `DefaultConfig()` 函数定义
2. **配置文件** - 通过 `--config` 指定，或按顺序搜索默认路径
3. **环境变量** - `MC_VMQUERY_*` 前缀
4. **CLI flags** - 最高优先级

### 配置文件搜索路径

当未指定 `--config` 时，按以下顺序搜索：

1. `./config.yaml` (当前目录)
2. `./config/config.yaml` (当前目录下 config 子目录)
3. `$HOME/.mc-vmquery.yaml` (用户主目录)
4. `/etc/mc-vmquery/config.yaml` (系统配置)

### 配置结构

> 详见 `internal/config/config.go`

| 配置项               | 类型     | 说明                            |
| -------------------- | -------- | ------------------------------- |
| `server.url`         | string   | VictoriaMetrics 地址            |
| `server.path_prefix` | string   | API 路径前缀 (如 /victoria)     |
| `server.timeout`     | Duration | 查询超时                        |
| `auth.type`          | string   | 认证类型 (basic/bearer)         |
| `auth.user`          | string   | Basic 用户名                    |
| `auth.password`      | string   | Basic 密码                      |
| `auth.token`         | string   | Bearer Token                    |
| `tls.ca`             | string   | CA 证书路径                     |
| `tls.cert`           | string   | 客户端证书                      |
| `tls.key`            | string   | 客户端密钥                      |
| `tls.skip_verify`    | bool     | 跳过证书验证                    |
| `output.format`      | string   | 输出格式 (table/json/csv/graph) |
| `output.no_headers`  | bool     | 禁用表头                        |

### 配置文件示例 (YAML)

```yaml
server:
  url: http://localhost:8428
  path_prefix: "" # 可选，API 路径前缀
  timeout: 30s

auth:
  type: bearer
  token: my-secret-token

tls:
  ca: /path/to/ca.crt
  skip_verify: false

output:
  format: table
  no_headers: false
```

### 环境变量映射

使用 `MC_VMQUERY_` 前缀，下划线分隔嵌套路径：

| 环境变量                        | 配置路径             |
| ------------------------------- | -------------------- |
| `MC_VMQUERY_SERVER_URL`         | `server.url`         |
| `MC_VMQUERY_SERVER_PATH_PREFIX` | `server.path_prefix` |
| `MC_VMQUERY_SERVER_TIMEOUT`     | `server.timeout`     |
| `MC_VMQUERY_AUTH_TYPE`          | `auth.type`          |
| `MC_VMQUERY_AUTH_TOKEN`         | `auth.token`         |
| `MC_VMQUERY_TLS_SKIP_VERIFY`    | `tls.skip_verify`    |
| `MC_VMQUERY_OUTPUT_FORMAT`      | `output.format`      |

### CLI 参数映射

koanf 标签使用 `snake_case`，CLI 使用 `kebab-case`，自动转换：

| 配置路径             | CLI 参数               |
| -------------------- | ---------------------- |
| `server.url`         | `--server-url`         |
| `server.path_prefix` | `--server-path-prefix` |
| `server.timeout`     | `--server-timeout`     |
| `auth.type`          | `--auth-type`          |
| `auth.token`         | `--auth-token`         |
| `tls.skip_verify`    | `--tls-skip-verify`    |
| `output.format`      | `--output-format`      |

## 命令设计

### mc-vmquery 命令结构

```bash
mc-vmquery [flags] <query>              # Instant 查询 (主命令)
mc-vmquery query <query> [flags]        # 显式查询子命令
mc-vmquery metrics [match] [flags]      # 列出指标
mc-vmquery labels [flags]               # 列出所有标签
mc-vmquery label-values <label> [flags] # 获取标签值
mc-vmquery series [match] [flags]       # 系列数据
mc-vmquery version                      # 版本信息
```

### mc-vmexport 命令结构

```bash
mc-vmexport [flags] <match>             # 导出数据 (默认 JSON Line)
mc-vmexport native <match> [flags]      # Native 二进制格式导出
mc-vmexport csv <match> [flags]         # CSV 格式导出
mc-vmexport json <match> [flags]        # JSON Line 格式导出 (显式)
mc-vmexport version                     # 版本信息
```

**mc-vmexport 特有参数：**

```
--start                   开始时间 (RFC3339/Unix 时间戳)
--end                     结束时间 (RFC3339/Unix 时间戳)
--output, -o              输出文件路径 (默认: stdout)
--gzip                    启用 gzip 压缩
--max-rows-per-line       JSON Line 每行最大样本数

# CSV 格式专用
--csv-format              CSV 列定义 (默认: __name__,__value__,__timestamp__:unix_s)
--csv-timestamp-format    时间戳格式: unix_s, unix_ms, unix_ns, rfc3339
```

**mc-vmexport API 端点 (3 个)：**

| 子命令        | API 端点                | 格式      | 说明               |
| ------------- | ----------------------- | --------- | ------------------ |
| `json` (默认) | `/api/v1/export`        | JSON Line | 每行一个 JSON 对象 |
| `csv`         | `/api/v1/export/csv`    | CSV       | 灵活的列定义       |
| `native`      | `/api/v1/export/native` | Binary    | 最高效率，VM 专用  |

### mc-vmimport 命令结构

```bash
mc-vmimport [flags] <file>              # 导入数据 (自动检测格式)
mc-vmimport native <file> [flags]       # Native 二进制格式导入
mc-vmimport csv <file> [flags]          # CSV 格式导入
mc-vmimport json <file> [flags]         # JSON Line 格式导入
mc-vmimport prometheus <file> [flags]   # Prometheus 格式导入
mc-vmimport version                     # 版本信息
```

**mc-vmimport 特有参数：**

```
--input, -i               输入文件路径 (默认: stdin)
--gzip                    输入为 gzip 压缩格式

# Prometheus 格式专用
--job                     Pushgateway job 标签
--instance                Pushgateway instance 标签
```

**mc-vmimport API 端点 (4 个)：**

| 子命令       | API 端点                    | 格式      | 说明                                   |
| ------------ | --------------------------- | --------- | -------------------------------------- |
| `json`       | `/api/v1/import`            | JSON Line | 每行一个 JSON 对象                     |
| `csv`        | `/api/v1/import/csv`        | CSV       | 需配合导出的 CSV 格式                  |
| `native`     | `/api/v1/import/native`     | Binary    | 最高效率，VM 专用                      |
| `prometheus` | `/api/v1/import/prometheus` | Text      | Prometheus exposition/OpenMetrics 格式 |

### 全局参数

```
# 配置文件
--config, -c              配置文件路径 (可选，优先于默认搜索路径)

# 服务器配置
--server-url              VictoriaMetrics 地址 (默认: http://localhost:8428)
--server-timeout          查询超时 (默认: 30s)

# 输出配置
--output-format, -o       输出格式: table, json, csv, graph (默认: table)
--output-no-headers       禁用表头输出

# 认证配置
--auth-type               认证类型: basic, bearer
--auth-user               Basic 认证用户名
--auth-password           Basic 认证密码
--auth-token              Bearer Token

# TLS 配置
--tls-ca                  CA 证书路径
--tls-cert                客户端证书路径
--tls-key                 客户端密钥路径
--tls-skip-verify         跳过证书验证
```

### 使用示例

```bash
# 使用默认配置
mc-vmquery 'up'

# 指定配置文件
mc-vmquery -c /path/to/config.yaml 'up'

# 命令行覆盖配置
mc-vmquery --server-url http://vm:8428 --output-format json 'up'

# 环境变量
export MC_VMQUERY_SERVER_URL=http://vm:8428
mc-vmquery 'up'
```

## 架构设计

### 目录结构

```
cmd/
├── vm-metrics/main.go              # 统一命令入口
├── mc-vmquery/main.go              # 查询工具入口
├── mc-vmexport/main.go             # 导出工具入口
└── mc-vmimport/main.go             # 导入工具入口

internal/
├── command/                        # CLI 命令定义
│   ├── command.go                  # 共享组件 (BaseFlags, NewClient, ParseTime)
│   ├── query/
│   │   ├── command.go              # 命令定义
│   │   └── action.go               # 业务逻辑
│   ├── export/
│   │   ├── command.go              # 命令定义
│   │   └── action.go               # 业务逻辑
│   └── import/
│       ├── command.go              # 命令定义
│       └── action.go               # 业务逻辑
├── config/                         # 配置管理 (共享)
│   ├── config.go                   # 配置结构定义
│   └── loading.go                  # koanf 配置加载
├── vmapi/                          # VictoriaMetrics API 客户端 (共享)
│   ├── client.go                   # Client 接口定义
│   ├── resty.go                    # go-resty 实现
│   ├── export.go                   # 导出 API
│   ├── import.go                   # 导入 API
│   └── types.go                    # API 响应类型
├── output/                         # 输出格式化
│   ├── writer.go                   # Writer 接口 + 工厂
│   ├── table.go                    # tabwriter 表格输出
│   ├── json.go                     # JSON 输出
│   ├── csv.go                      # CSV 输出
│   └── graph.go                    # asciigraph ASCII 图表
└── version/                        # 版本信息 (共享)
    ├── command.go                  # 版本命令
    └── version.go                  # 版本信息
```

### 核心接口

#### vmapi.Client

> 详见 `internal/vmapi/client.go`

| 方法          | 说明     | API 端点                      |
| ------------- | -------- | ----------------------------- |
| `Query`       | 即时查询 | `/api/v1/query`               |
| `QueryRange`  | 范围查询 | `/api/v1/query_range`         |
| `Series`      | 时间序列 | `/api/v1/series`              |
| `Labels`      | 标签名称 | `/api/v1/labels`              |
| `LabelValues` | 标签值   | `/api/v1/label/<name>/values` |

#### output.Writer

> 详见 `internal/output/writer.go`

| 方法               | 说明                                |
| ------------------ | ----------------------------------- |
| `WriteQueryResult` | 输出查询结果 (vector/matrix/scalar) |
| `WriteStrings`     | 输出字符串列表 (labels/metrics)     |
| `WriteSeries`      | 输出时间序列列表                    |

**支持的输出格式：**

| 格式    | 实现文件   | 说明                          |
| ------- | ---------- | ----------------------------- |
| `table` | `table.go` | tabwriter 美化表格 (默认)     |
| `json`  | `json.go`  | 缩进格式化的 JSON             |
| `csv`   | `csv.go`   | 标准 CSV 格式                 |
| `graph` | `graph.go` | asciigraph ASCII 图表 (range) |

## VictoriaMetrics API

### 查询端点 (mc-vmquery)

| 端点                          | 方法     | 说明         |
| ----------------------------- | -------- | ------------ |
| `/api/v1/query`               | GET/POST | Instant 查询 |
| `/api/v1/query_range`         | GET/POST | Range 查询   |
| `/api/v1/series`              | GET/POST | 系列数据     |
| `/api/v1/labels`              | GET/POST | 标签名列表   |
| `/api/v1/label/<name>/values` | GET      | 标签值列表   |

### 导出端点 (mc-vmexport)

| 端点                    | 方法     | 格式      | 说明                           |
| ----------------------- | -------- | --------- | ------------------------------ |
| `/api/v1/export`        | GET/POST | JSON Line | 每行一个 JSON 对象，支持 gzip  |
| `/api/v1/export/csv`    | GET/POST | CSV       | 灵活列定义，支持多种时间戳格式 |
| `/api/v1/export/native` | GET/POST | Binary    | VM 原生格式，效率最高          |

**导出参数：**

| 参数                | 说明                  | 适用端点      |
| ------------------- | --------------------- | ------------- |
| `match[]`           | 时间序列选择器 (必需) | 全部          |
| `start`             | 开始时间              | 全部          |
| `end`               | 结束时间              | 全部          |
| `max_rows_per_line` | 每行最大样本数        | `/export`     |
| `format`            | CSV 列定义            | `/export/csv` |
| `reduce_mem_usage`  | 跳过去重 (1=启用)     | `/export/csv` |

### 导入端点 (mc-vmimport)

| 端点                        | 方法 | 格式      | 说明                                   |
| --------------------------- | ---- | --------- | -------------------------------------- |
| `/api/v1/import`            | POST | JSON Line | 流式导入，支持 gzip                    |
| `/api/v1/import/csv`        | POST | CSV       | 配合 `/export/csv` 使用                |
| `/api/v1/import/native`     | POST | Binary    | VM 原生格式，效率最高                  |
| `/api/v1/import/prometheus` | POST | Text      | Prometheus exposition/OpenMetrics 格式 |

**导入请求头：**

| Header                   | 说明               |
| ------------------------ | ------------------ |
| `Content-Encoding: gzip` | 发送 gzip 压缩数据 |

**Prometheus 格式变体：**

```
/api/v1/import/prometheus/metrics/job/<job>/instance/<instance>
```

Pushgateway 兼容端点，自动添加 job/instance 标签。

### 响应格式

> 详见 `internal/vmapi/types.go`

| 字段        | 类型     | 说明                           |
| ----------- | -------- | ------------------------------ |
| `status`    | string   | 响应状态 ("success" / "error") |
| `data`      | object   | 响应数据                       |
| `errorType` | string   | 错误类型                       |
| `error`     | string   | 错误信息                       |
| `warnings`  | []string | 警告信息                       |

**查询结果类型：**

| ResultType | 说明     | 数据结构          |
| ---------- | -------- | ----------------- |
| `vector`   | 即时向量 | `[]Sample` (单值) |
| `matrix`   | 范围矩阵 | `[]Sample` (多值) |
| `scalar`   | 标量     | `SampleValue`     |
| `string`   | 字符串   | `StringResult`    |

## 实施路线图

### Phase 1: 基础架构 ✅

- [x] CLI 骨架 (`urfave/cli/v3`)
- [x] 配置系统 (`koanf`)
- [x] 版本命令
- [x] HTTP 客户端 (`go-resty`)
- [x] 认证选项 (Basic/Bearer/TLS)
- [x] 路径前缀支持 (`--server-path-prefix`)

### Phase 2: mc-vmquery ✅

- [x] 核心查询功能 (Query, QueryRange)
- [x] output.Writer 接口
- [x] Table 输出 (`tabwriter`)
- [x] JSON 输出
- [x] CSV 输出
- [x] ASCII Graph 输出 (`asciigraph`)
- [x] 辅助子命令: `metrics`, `labels`, `label-values`, `series`

### Phase 3: mc-vmexport ✅

- [x] vmapi 导出接口
- [x] JSON Line 导出
- [x] CSV 导出
- [x] Native 导出
- [x] gzip 压缩支持

### Phase 4: mc-vmimport ✅

- [x] vmapi 导入接口
- [x] JSON Line 导入
- [x] CSV 导入
- [x] Native 导入
- [x] Prometheus 格式导入
- [x] gzip 解压支持

### Phase 5: 统一命令 ✅

- [x] `vm-metrics` 统一入口
- [x] 子命令别名 (q/e/i)
- [x] 代码重构 (command.go + action.go)

## 与 promql-cli 差异

| 方面           | promql-cli               | mc-vmquery        |
| -------------- | ------------------------ | ----------------- |
| 目标服务       | Prometheus               | VictoriaMetrics   |
| CLI 框架       | cobra                    | urfave/cli/v3     |
| 配置管理       | viper                    | koanf             |
| HTTP 客户端    | prometheus/client_golang | go-resty/resty/v2 |
| 配置文件参数   | `--config`               | `--config, -c`    |
| 服务器地址参数 | `--host`                 | `--server-url`    |
| Range 参数     | `--start`                | `--range`         |
| label-values   | 不支持                   | 支持              |
| 环境变量前缀   | `PROMQL_`                | `MC_VMQUERY_`     |
| 默认端口       | 9090                     | 8428              |

## 参考资料

- [VictoriaMetrics 文档](https://docs.victoriametrics.com/)
- [MetricsQL 文档](https://docs.victoriametrics.com/victoriametrics/metricsql/)
- [promql-cli 源码分析](../promql-cli/architecture.md)
