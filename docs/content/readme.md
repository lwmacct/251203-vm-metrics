# mc-metrics

VictoriaMetrics 命令行工具集，支持 MetricsQL 查询、数据导入导出。

## 安装

```bash
go install github.com/lwmacct/251203-mc-metrics/cmd/mc-metrics@latest

# 或单独安装
go install github.com/lwmacct/251203-mc-metrics/cmd/mc-vmquery@latest
go install github.com/lwmacct/251203-mc-metrics/cmd/mc-vmexport@latest
go install github.com/lwmacct/251203-mc-metrics/cmd/mc-vmimport@latest
```

## 快速开始

```bash
# 统一命令
mc-metrics query 'up{job="prometheus"}'
mc-metrics export '{job="node"}' -o data.json
mc-metrics import data.json

# 使用别名
mc-metrics q -o json 'up'           # query
mc-metrics e '{job="node"}'         # export
mc-metrics i data.json              # import

# 独立命令 (等效)
mc-vmquery 'up{job="prometheus"}'
mc-vmexport '{job="node"}'
mc-vmimport data.json
```

## 命令结构

```
mc-metrics                      # 统一入口
├── query (q)                   # MetricsQL 查询
│   ├── metrics                 # 列出所有指标
│   ├── labels                  # 列出所有标签
│   ├── label-values <label>    # 获取标签值
│   └── series <match>          # 列出时间序列
├── export (e)                  # 数据导出
│   ├── json                    # JSON Line 格式
│   ├── csv                     # CSV 格式
│   └── native                  # 原生二进制格式
├── import (i)                  # 数据导入
│   ├── json                    # JSON Line 格式
│   ├── csv                     # CSV 格式
│   ├── native                  # 原生二进制格式
│   └── prometheus              # Prometheus 格式
└── version                     # 版本信息
```

## 输出格式

```bash
# 表格 (默认)
mc-metrics query 'up'

# JSON
mc-metrics query -o json 'up'

# CSV
mc-metrics query -o csv 'up'

# ASCII 图表 (仅 range query)
mc-metrics query -o graph --range 1h 'rate(http_requests_total[5m])'
```

## 配置

### 配置文件

```yaml
# ~/.mc-vmquery.yaml
server:
  url: http://localhost:8428
  path_prefix: /victoria # 可选，API 路径前缀
  timeout: 30s

auth:
  type: bearer
  token: my-secret-token

output:
  format: table
```

### 环境变量

```bash
export MC_VMQUERY_SERVER_URL=http://vm:8428
export MC_VMQUERY_SERVER_PATH_PREFIX=/victoria
export MC_VMQUERY_AUTH_TOKEN=my-token
```

### 命令行参数

```bash
mc-metrics query --server-url http://vm:8428 --server-path-prefix /victoria 'up'
```

## 开发

### 初始化开发环境

```bash
pre-commit install
```

### 构建

```bash
go build ./cmd/...
```

### 运行测试

```bash
go test ./...
```

### 查看所有任务

```bash
task -a
```

## 相关链接

- [设计文档](docs/content/guide/design.md)
- [VictoriaMetrics 文档](https://docs.victoriametrics.com/)
- [MetricsQL 文档](https://docs.victoriametrics.com/victoriametrics/metricsql/)

## 工具链

- [Taskfile](https://taskfile.dev) - 项目 CLI 管理
- [Pre-commit](https://pre-commit.com/) - Git 钩子管理
