# promql-cli 项目分析

<!--TOC-->

- [概述](#概述) `:23+4`
- [功能特性](#功能特性) `:27+10`
- [安装](#安装) `:37+16`
  - [从 Release 下载](#从-release-下载) `:39+4`
  - [从源码构建](#从源码构建) `:43+10`
- [快速开始](#快速开始) `:53+25`
  - [基本查询](#基本查询) `:55+10`
  - [输出格式](#输出格式) `:65+13`
- [配置](#配置) `:78+20`
  - [环境变量](#环境变量) `:88+10`
- [认证](#认证) `:98+12`
- [相关文档](#相关文档) `:110+4`

<!--TOC-->

<!--TOC-->

## 概述

[promql-cli](https://github.com/nalbury/promql-cli) 是一个命令行工具，用于从终端直接查询 Prometheus 服务器进行快速数据分析。

## 功能特性

| 功能 | 描述 |
|------|------|
| Instant Query | 执行即时查询，返回当前时间点的指标值 |
| Range Query | 执行范围查询，返回时间序列数据并以 ASCII 图表展示 |
| Metrics | 列出所有可用的指标名称 |
| Labels | 获取给定查询的所有标签 |
| Meta | 获取指标的类型和帮助元数据 |

## 安装

### 从 Release 下载

macOS 和 Linux 的二进制文件可在 [Releases 页面](https://github.com/nalbury/promql-cli/releases) 下载。

### 从源码构建

需要 Go 1.13+ 环境：

```bash
git clone https://github.com/nalbury/promql-cli.git
cd promql-cli/
OS=linux INSTALL_PATH=/usr/local/bin make install
```

## 快速开始

### 基本查询

```bash
# Instant 查询
promql --host "http://prometheus:9090" 'sum(up) by (job)'

# Range 查询 (过去 1 小时)
promql --host "http://prometheus:9090" 'sum(up) by (job)' --start 1h
```

### 输出格式

```bash
# JSON 输出
promql 'up' --output json

# CSV 输出
promql 'up' --output csv

# 禁用表头
promql 'up' --no-headers
```

## 配置

配置文件默认位于 `$HOME/.promql-cli.yaml`：

```yaml
host: https://my.prometheus.server:9090
output: json
step: 5m
```

### 环境变量

支持 `PROMQL_` 前缀的环境变量：

```bash
export PROMQL_HOST="http://prometheus:9090"
export PROMQL_AUTH_TYPE="Bearer"
export PROMQL_AUTH_CREDENTIALS="my-token"
```

## 认证

支持 Basic 和 Bearer 认证：

```bash
# Bearer Token
promql --auth-type Bearer --auth-credentials "my-token" metrics

# Basic Auth (base64 编码)
promql --auth-type Basic --auth-credentials-file ~/.promql_token metrics
```

## 相关文档

- [架构分析](./architecture.md)
- [命令详解](./commands.md)
