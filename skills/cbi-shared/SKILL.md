---
name: cbi-shared
description: 上传和查询 CreatiBI 素材库文件。CreatiBI CLI 基础：应用配置初始化、认证登录（auth login）、身份查看（auth whoami）等。当用户需要第一次配置、使用登录授权、遇到权限不足或首次使用 cbi-cli 时触发。
trigger:
  - "cbi"
  - "素材库"
  - "上传素材"
  - "cbi config init"
  - "cbi auth login"
  - "cbi auth whoami"
  - "cbi repository"
  - "CreatiBI CLI"
  - "权限不足"
  - "首次使用"
  - "config init --new"
  - "file-create"
  - "file-list"
  - "file-detail"
---

# CreatiBI CLI 素材库管理

CreatiBI 命令行工具，用于素材库管理。支持 OAuth 登录和素材库文件操作。

## 安装

```bash
# npm 全局安装（推荐）
npm install -g @creatibi/cbi-cli@latest

# 或使用 npx（无需安装）
npx @creatibi/cbi-cli --help
```

支持平台：macOS (amd64, arm64)、Linux (amd64, arm64)、Windows (amd64, arm64)

---

## 快速开始

```bash
# 1. 初始化配置（首次使用）
cbi config init

# 2. OAuth 登录
cbi auth login

# 3. 查看素材库列表
cbi repository list

# 4. 查看素材文件列表
cbi repository file-list --repository-id 1

# 5. 获取文件详情
cbi repository file-detail 123

# 6. 上传文件到素材库
cbi repository file-create --repository-id 1 --file ./image.png
```

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
1. CLI 自动打开浏览器访问 https://open.creatibi.cn
2. 用户在开放平台创建应用并获取凭证
3. 凭证自动回传到 CLI（30秒倒计时，超时则手动输入）
4. 配置写入 `~/.cbi/config.json`

**手动输入字段：**
- `client_id`（应用 ID）- 必填
- `client_secret`（应用密钥）- 必填
- `base_url`（默认 https://open.creatibi.cn）
- `default_workspace`（可选）

### 显示当前配置

```bash
# 显示配置（敏感字段脱敏）
cbi config show

# 详细模式
cbi config show -v

# JSON 格式输出
cbi config show --format json
```

---

## 认证模块 (auth)

### OAuth 登录

```bash
cbi auth login
```

**前提条件：** 已执行 `cbi config init` 配置应用凭证

**流程：**
1. CLI 启动本地回调服务器（端口 8080-8090，自动尝试可用端口）
2. 自动打开浏览器访问授权页面
3. 用户在浏览器中完成授权
4. 服务端回调到 CLI 并返回授权码
5. CLI 用授权码换取 access_token
6. Token 存储在 ~/.cbi/config.json

### 查看当前身份

```bash
# 查看当前登录身份
cbi auth whoami

# 详细模式（显示 token 信息）
cbi auth whoami -v
```

### 退出登录

```bash
cbi auth logout
```

---

## 素材库模块 (repository/repo)

### 列出可访问素材库

```bash
# 表格格式
cbi repository list

# JSON 格式
cbi repository list --format json

# 使用别名
cbi repo list
```

### 列出文件夹

```bash
# 根目录文件夹
cbi repository folders --repository-id 1

# 包含文件数量统计
cbi repository folders --repository-id 1 --with-statistic

# 指定父文件夹
cbi repository folders --repository-id 1 --parent-folder-id 100
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--parent-folder-id`: 父文件夹 ID（0 表示根目录）
- `--with-statistic`: 包含统计信息（文件数量）

### 查询素材文件列表

```bash
# 列出素材库所有文件
cbi repository file-list --repository-id 1

# 按文件夹筛选
cbi repository file-list --repository-id 1 --folder-id 10

# 按标签筛选
cbi repository file-list --repository-id 1 --tag-id 5

# 搜索关键词（名称 + signals）
cbi repository file-list --repository-id 1 --keyword "广告"

# 分页查询
cbi repository file-list --repository-id 1 --page 2 --pageSize 30

# JSON 格式
cbi repository file-list --repository-id 1 --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--folder-id`: 文件夹 ID（文件夹筛选模式）
- `--tag-id`: 标签 ID（标签筛选模式）
- `--keyword`: 搜索关键词（搜索名称+signals）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

**筛选模式（oneof，不可组合）：**
- `--folder-id`: 按文件夹筛选
- `--tag-id`: 按标签筛选
- `--keyword`: 按关键词搜索

### 获取文件详情

```bash
# 获取文件详情
cbi repository file-detail <file-id>

# JSON 格式输出
cbi repository file-detail <file-id> --format json

# 静默模式（只输出 JSON）
cbi repository file-detail <file-id> -q
```

**输出信息包括：**
- 基本信息：ID、名称、格式、大小、时长、分辨率等
- 来源信息：来源平台、来源 URL
- 关联产品、标签、文件夹
- 创建者信息
- 视频理解信号（signals）
- 各种 URL：封面、原始文件、预览

### 文件查重

```bash
# 通过文件路径自动计算 MD5
cbi repository file-check --repository-id <id> --file ./image.png

# 直接提供 MD5
cbi repository file-check --repository-id <id> --file-md5 abc123def456

# 详细模式（显示 MD5 值）
cbi repository file-check --repository-id <id> --file ./image.png -v
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file`: 本地文件路径（用于计算 MD5）
- `--file-md5`: 文件 MD5 值

### 上传文件

```bash
# 基本上传
cbi repository file-create --repository-id <id> --file ./image.png

# 上传到指定文件夹
cbi repository file-create --repository-id <id> --file ./video.mp4 --folder-id 123

# 跳过查重检查
cbi repository file-create --repository-id <id> --file ./image.png --skip-check

# 强制上传（即使重复）
cbi repository file-create --repository-id <id> --file ./image.png --force

# 完整参数示例
cbi repository file-create \
  --repository-id 1 \
  --file ./image.png \
  --folder-id 100 \
  --name "创意素材" \
  --note "用于春节期间投放" \
  --rating 5 \
  --source-url "https://example.com/source" \
  --tags "春节,促销,创意"
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file`: 本地文件路径（必填）
- `--folder-id`: 目标文件夹 ID
- `--name`: 文件名（默认使用原文件名）
- `--note`: 备注
- `--rating`: 评分（1-5）
- `--source-url`: 来源 URL
- `--tags`: 标签（逗号分隔）
- `--skip-check`: 跳过查重检查
- `--force`: 强制上传（即使文件重复）

**上传流程：**
1. 检查文件大小（限制 100MB）
2. 计算 MD5 进行查重（除非 `--skip-check`）
3. 如果重复，提示用户（除非 `--force`）
4. 上传文件

---

## 通用参数

```bash
--config <path>              # 配置文件路径（默认 ~/.cbi/config.json）
-f, --format <format>        # 输出格式：json | table
-o, --output <file>          # 输出到文件
-q, --quiet                  # 只输出数据，无日志
-v, --verbose                # 显示详细信息
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
# 重新登录
cbi auth login
```

### 端口占用

`cbi config init` 和 `cbi auth login` 会自动尝试端口 8080-8090，无需手动处理。

### 配置文件问题

```bash
# 强制重新初始化配置
cbi config init --new

# 重新登录
cbi auth login
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