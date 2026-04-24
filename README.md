# cbi - CreatiBI CLI

CreatiBI 命令行工具，用于素材库管理。

支持 OAuth 登录和素材库文件操作，通过命令行安全、便捷地管理创意素材。

## 安装

### 通过 npm 安装（推荐）

```bash
# 安装到项目
npm install @creatibi/cbi-cli

# 或全局安装
npm install -g @creatibi/cbi-cli

# 使用 npx（无需安装）
npx @creatibi/cbi-cli --help
```

支持平台：
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64, arm64)

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/CreatiBI/cli.git

# 构建
make build

# 安装到系统
sudo make install

# 跨平台编译（用于发布）
make build-all
```

## 快速开始

```bash
# 1. 初始化配置（首次使用）
cbi config init

# 2. OAuth 登录
cbi auth login

# 3. 查看素材库列表
cbi repository list

# 4. 上传文件到素材库
cbi repository file-create --repository-id 1 --file ./image.png
```

---

## 命令概览

```
cbi
├── config                    # 配置管理
│   ├── init --new            # 初始化配置
│   └── show                  # 显示当前配置
├── auth                      # 认证管理
│   ├── login                 # OAuth 登录
│   ├── whoami                # 查看当前身份
│   └── logout                # 退出登录
└── repository (repo)         # 素材库管理
    ├── list                  # 素材库列表
    ├── folders               # 文件夹列表
    ├── file-check            # 文件查重（MD5）
    └── file-create           # 上传文件
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

流程：
1. 检查配置是否已存在
2. 若已存在，需使用 `--new` 强制覆盖
3. 引导输入应用凭证信息：
   - client_id（应用 ID）
   - client_secret（应用密钥）
   - base_url（默认 https://open.creatibi.cn）
   - default_workspace（可选）
4. 配置写入 `~/.cbi/config.json`

前提条件：
- 在 CreatiBI 开放平台创建应用
- 获取 client_id 和 client_secret

### 显示当前配置

```bash
# 显示配置（敏感字段脱敏）
cbi config show

# 详细模式（显示登录凭证）
cbi config show -v

# JSON 格式输出
cbi config show --format json
```

脱敏规则：
- client_secret: 显示前 4 位 + **** + 后 4 位
- access_token: 显示前 4 位 + **** + 后 4 位
- refresh_token: 显示前 4 位 + **** + 后 4 位

示例输出：
```
当前配置:
  配置文件: /Users/xxx/.cbi/config.json

  base_url:          https://open.creatibi.cn
  client_id:         c674ced8ed314c1e9651af5cc23c4ad7
  client_secret:     7c9b****125b

登录状态:
  access_token:      MTG0****YTG1
  refresh_token:     ZDJI****OTHL
  token_expires_at:  2026-04-24 00:28:54
```

---

## 认证模块 (auth)

### OAuth 登录

```bash
# 使用已配置的凭证登录
cbi auth login
```

前提条件：已执行 `cbi config init` 配置应用凭证。

流程：
1. CLI 启动本地回调服务器（端口 8080）
2. 自动打开浏览器访问授权页面
3. 用户在浏览器中完成授权
4. 服务端回调到 CLI 并返回授权码
5. CLI 用授权码换取 access_token
6. Token 存储在 ~/.cbi/config.json

### 查看当前身份

```bash
cbi auth whoami

# 详细模式（显示 Token 信息）
cbi auth whoami -v
```

### 退出登录

```bash
cbi auth logout
```

清除登录凭证，保留应用配置。

---

## 素材库模块 (repository/repo)

### 列出可访问素材库

```bash
cbi repository list

# JSON 格式
cbi repository list --format json

# 使用别名
cbi repo list
```

输出字段：ID、名称、描述、默认、权限

### 列出文件夹

```bash
# 根目录文件夹
cbi repository folders --repository-id 1

# 包含文件数量统计
cbi repository folders --repository-id 1 --with-statistic

# 指定父文件夹
cbi repository folders --repository-id 1 --parent-folder-id 100
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--parent-folder-id`: 父文件夹 ID（0 表示根目录）
- `--with-statistic`: 包含统计信息（文件数量）

### 文件查重

```bash
# 通过文件路径自动计算 MD5
cbi repository file-check --repository-id 1 --file ./image.png

# 直接提供 MD5
cbi repository file-check --repository-id 1 --file-md5 abc123def456

# 详细模式（显示 MD5 值）
cbi repository file-check --repository-id 1 --file ./image.png -v
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file`: 本地文件路径（用于计算 MD5）
- `--file-md5`: 文件 MD5 值

### 上传文件

```bash
# 基本上传
cbi repository file-create --repository-id 1 --file ./image.png

# 上传到指定文件夹
cbi repository file-create --repository-id 1 --file ./video.mp4 --folder-id 123

# 跳过查重检查
cbi repository file-create --repository-id 1 --file ./image.png --skip-check

# 强制上传（即使重复）
cbi repository file-create --repository-id 1 --file ./image.png --force

# 完整参数
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

参数：
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

上传流程：
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

## 输出格式

| 格式 | 说明 | 适用场景 |
|------|------|----------|
| `table` | 表格格式（默认） | 人工查看 |
| `json` | JSON 格式 | 程序处理 |

---

## 配置文件

配置存储在 `~/.cbi/config.json`：

```json
{
  "base_url": "https://open.creatibi.cn",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "default_workspace": "",
  "api_key": "YOUR_ACCESS_TOKEN",
  "refresh_token": "YOUR_REFRESH_TOKEN",
  "token_expires_at": "2026-04-24T00:28:54Z"
}
```

配置说明：
- `base_url`: API 基础地址
- `client_id`: OAuth 应用 ID
- `client_secret`: OAuth 应用密钥（登录凭证）
- `default_workspace`: 默认 workspace（可选）
- `api_key`: 登录后获取的 access_token
- `refresh_token`: 用于刷新 access_token
- `token_expires_at`: token 过期时间

---

## 开发

```bash
make build      # 构建当前平台
make build-all  # 跨平台编译（用于发布）
make run        # 运行开发模式
make test       # 运行测试
make package    # 构建 + 打包 npm
```

### 发布新版本

```bash
# 1. 更新版本号
# 编辑 package.json 中的 version 字段

# 2. 构建并打包
make build-all
npm pack

# 3. 发布到 npm（需要 npm 账号）
npm publish --access restricted
```

---

## 许可证

Copyright © CreatiBI