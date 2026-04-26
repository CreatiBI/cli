---
name: cbi-shared
description: CreatiBI CLI 共享基础：应用配置初始化、认证登录（auth login）、身份查看（auth whoami）。当用户需要第一次配置、使用登录授权、遇到权限不足或首次使用 cbi-cli 时触发。
trigger:
  - "cbi config init"
  - "cbi auth login"
  - "cbi auth whoami"
  - "cbi auth logout"
  - "cbi config show"
  - "初始化 cbi"
  - "配置 cbi"
  - "登录 cbi"
  - "cbi 登录"
  - "CreatiBI 登录"
  - "首次使用 cbi"
  - "权限不足"
  - "Token 过期"
  - "重新登录"
---

# CreatiBI CLI 基础配置与认证

CreatiBI 命令行工具的初始化与认证模块。

## 安装

```bash
# npm 全局安装
npm install -g @creatibi/cbi-cli@latest
```

支持平台：macOS (amd64/arm64)、Linux (amd64/arm64)、Windows (amd64/arm64)

---

## 快速开始

```bash
# 1. 初始化配置（首次使用）
cbi config init

# 2. OAuth 登录
cbi auth login

# 3. 确认登录成功
cbi auth whoami
```

---

## 常用命令速查

| 场景 | 命令 |
|------|------|
| 初始化配置 | `cbi config init` |
| 强制重新初始化 | `cbi config init --new` |
| 查看配置 | `cbi config show` |
| 登录授权 | `cbi auth login` |
| 查看身份 | `cbi auth whoami` |
| 退出登录 | `cbi auth logout` |

---

## 配置模块 (config)

### 初始化配置

首次使用需要初始化应用凭证配置：

```bash
# 初始化配置（交互式输入）
cbi config init

# 强制重新初始化（覆盖已有配置）
cbi config init --new
```

**流程：**
1. 检查配置是否已存在
2. 若已存在，需使用 `--new` 强制覆盖
3. 引导输入应用凭证信息：
   - `client_id`（应用 ID）- 必填
   - `client_secret`（应用密钥）- 必填
   - `base_url`（默认 https://open.creatibi.cn）
   - `default_workspace`（可选）
4. 配置写入 `~/.cbi/config.json`

**前提条件：** 在 CreatiBI 开放平台创建应用，获取 client_id 和 client_secret

### 显示当前配置

```bash
cbi config show              # 显示配置（敏感字段脱敏）
cbi config show -v           # 详细模式（显示登录凭证）
cbi config show --format json   # JSON 格式输出
```

---

## 认证模块 (auth)

### OAuth 登录

```bash
cbi auth login
```

**前提条件：** 已执行 `cbi config init` 配置应用凭证

**流程：**
1. CLI 启动本地回调服务器（端口 8080）
2. 自动打开浏览器访问授权页面
3. 用户在浏览器中完成授权
4. 服务端回调到 CLI 并返回授权码
5. CLI 用授权码换取 access_token
6. Token 存储在 ~/.cbi/config.json

### 查看当前身份

```bash
cbi auth whoami      # 查看当前登录身份
cbi auth whoami -v   # 详细模式（显示 token 信息）
```

### 退出登录

```bash
cbi auth logout
```

---

## 错误处理

### 权限不足/未登录

当遇到 `permission denied` 或 `auth required` 错误时：

```bash
# 1. 检查配置是否存在
cbi config show

# 2. 如果配置不存在，先初始化
cbi config init

# 3. 登录授权
cbi auth login

# 4. 确认登录成功
cbi auth whoami
```

### Token 过期

当遇到 `expired access token` 错误时：

```bash
cbi auth login   # 重新登录
```

### 配置文件问题

```bash
cbi config init --new   # 强制重新初始化配置
cbi auth login          # 重新登录
```

---

## 配置文件位置

配置存储在 `~/.cbi/config.json`：

```json
{
  "base_url": "https://open.creatibi.cn",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "api_key": "YOUR_ACCESS_TOKEN",
  "refresh_token": "YOUR_REFRESH_TOKEN",
  "token_expires_at": "2026-04-24T00:28:54Z"
}
```

---

## 通用参数

| 参数 | 说明 |
|------|------|
| `--config <path>` | 配置文件路径（默认 ~/.cbi/config.json） |
| `-f, --format` | 输出格式：json / table |
| `-v, --verbose` | 显示详细信息 |