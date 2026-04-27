# CreatiBI CLI

CreatiBI 命令行工具，用于素材库管理。

支持 OAuth 登录和素材库文件操作，通过命令行安全、便捷地管理创意素材。

## 安装

### 通过 npm 安装（推荐）

```bash
# 全局安装
npm install -g @creatibi/cbi-cli@latest

# 安装 CLI Skill（推荐）
npx skills add CreatiBI/cli -y -g
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
```

## 安装 CLI Skill（必需）

为了让 Agent 能够自动调用 cbi CLI，需要安装对应的 Skill：

```bash
npx skills add CreatiBI/cli -y -g
```

安装成功后，Claude Code 将自动识别以下触发词：
- `cbi`、`素材库`、`上传素材`、`上传到素材库`
- `cbi 上传`、`cbi repository`
- `repository file-create`、`file-list`、`file-detail`
- `文件查重`、`批量添加标签`、`批量添加文件夹`

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
    ├── folder-create         # 创建文件夹
    ├── tag-list              # 标签列表
    ├── tag-delete             # 删除标签
    ├── product-list          # 产品列表
    ├── product-delete        # 删除产品
    ├── file-list             # 文件列表（支持筛选）
    ├── file-detail           # 获取文件详情
    ├── file-check            # 文件查重（MD5）
    ├── file-create           # 上传文件
    ├── file-delete           # 删除文件到回收站
    ├── file-name-update      # 更新文件名称
    ├── file-notes-update     # 更新文件备注
    ├── file-score-update     # 更新文件评分
    ├── file-tag-add          # 批量添加标签
    ├── file-tag-remove       # 移除文件标签
    ├── file-folder-add       # 批量添加文件到文件夹
    ├── file-product-add      # 添加关联产品
    └── file-product-remove   # 移除关联产品
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

---

## 认证模块 (auth)

### OAuth 登录

```bash
cbi auth login
```

前提条件：已执行 `cbi config init` 配置应用凭证。

**支持两种登录模式：**

| 模式 | 适用场景 | 说明 |
|------|---------|------|
| 授权码模式 | 桌面环境（macOS/Windows/Linux） | 本地浏览器授权，自动回调 |
| 设备码模式 | VPS/服务器环境 | 远程浏览器授权，手动输入验证码 |

**授权码模式流程：**
1. CLI 启动本地回调服务器（端口 8080）
2. 自动打开浏览器访问授权页面
3. 用户在浏览器中完成授权
4. 服务端回调到 CLI 并返回授权码
5. CLI 用授权码换取 access_token
6. Token 存储在 ~/.cbi/config.json

**设备码模式流程：**
1. CLI 向服务端请求设备码
2. CLI 显示验证 URL 和验证码（如 `550e-8400`）
3. 用户在任意浏览器访问验证 URL 并输入验证码
4. CLI 轮询等待用户授权（默认 15 分钟有效期）
5. 授权成功后获取 access_token

```bash
# 直接使用设备码模式
cbi auth login --device

# 或通过环境变量设置
export CBI_LOGIN_MODE=device
cbi auth login
```

**设备码模式示例输出：**
```
========================================
   设备码登录
========================================

请在浏览器中访问以下地址:
  https://open.creatibi.cn/device/verify?user_code=550e-8400

或手动输入验证码:
  验证码: 550e-8400

----------------------------------------
有效期: 15 分钟
----------------------------------------

等待授权...
✓ 登录成功
Token 已存储到: ~/.cbi/config.json
```

### 查看当前身份

```bash
cbi auth whoami

# 详细模式
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

参数：
- `--repository-id`: 素材库 ID（必填）
- `--parent-folder-id`: 父文件夹 ID（0 表示根目录）
- `--with-statistic`: 包含统计信息（文件数量）

### 创建文件夹

```bash
# 创建根目录文件夹
cbi repository folder-create --repository-id 1 --name "新文件夹"

# 创建子文件夹
cbi repository folder-create --repository-id 1 --name "子文件夹" --parent-folder-id 100

# JSON 格式输出
cbi repository folder-create --repository-id 1 --name "新文件夹" --format json
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--name`: 文件夹名称（必填）
- `--parent-folder-id`: 父文件夹 ID（可选，默认为根目录）

### 列出标签

```bash
# 列出素材库所有标签
cbi repository tag-list --repository-id 1

# 包含使用次数统计
cbi repository tag-list --repository-id 1 --with-refcnt

# JSON 格式
cbi repository tag-list --repository-id 1 --format json
```

参数：
- `--repository-id`: 材库 ID（必填）
- `--with-refcnt`: 包含标签使用次数

### 删除标签

```bash
cbi repository tag-delete --repository-id 1 --tag-ids 5,10,15
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--tag-ids`: 标签 ID 列表（逗号分隔，必填）

注意：
- 删除后标签不再可用（软删除）
- 已关联的文件标签记录保留
- 需要档案库编辑权限

### 列出产品

```bash
# 列出素材库所有已关联产品
cbi repository product-list --repository-id 1

# JSON 格式
cbi repository product-list --repository-id 1 --format json
```

参数：
- `--repository-id`: 素材库 ID（必填）

产品类型：
- 1 = 应用
- 2 = 游戏
- 3 = 商品

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

参数：
- `--repository-id`: 素材库 ID（必填）
- `--folder-id`: 文件夹 ID（文件夹筛选模式）
- `--tag-id`: 标签 ID（标签筛选模式）
- `--keyword`: 搜索关键词（搜索名称+signals）
- `--has-signals`: 按视频理解信号筛选（true/false）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

筛选模式（oneof，不可组合）：
- `--folder-id`: 按文件夹筛选
- `--tag-id`: 按标签筛选
- `--keyword`: 按关键词搜索
- `--has-signals`: 按是否有视频理解信号筛选

### 获取文件详情

```bash
# 获取文件详情
cbi repository file-detail 123

# JSON 格式输出
cbi repository file-detail 123 --format json

# 静默模式（只输出 JSON）
cbi repository file-detail 123 -q
```

输出信息包括：
- 基本信息：ID、名称、格式、大小、时长、分辨率等
- 来源信息：来源平台、来源 URL
- 关联产品、标签、文件夹
- 创建者信息
- 视频理解信号（signals）
- 各种 URL：封面、原始文件、预览

示例输出：
```
文件详情:
  ID:           123
  名称:         广告视频.mp4
  格式:         mp4
  大小:         15.2MB (15200000 bytes)
  时长:         00:30
  分辨率:       1920x1080
  比例:         16:9
  帧率:         30fps
  评分:         85
  备注:         优质素材
  来源平台:     douyin
  创建时间:     2023-10-01 10:00:00

关联产品:
  - 产品A (ID: 1)

标签:
  - 游戏 (ID: 5)

所在文件夹:
  - 视频素材 (ID: 1749)

创建者:
  姓名:   张三
  邮箱:   zhang@example.com
  ID:     100

视频理解信号:
  [开头] 视频开头分析
    标签: 剧情, 情感
    内容:
      视频开头展示主角在城市街头奔跑...
  [人物] 主要人物特征
    标签: 年轻男性, 运动风格
```

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

### 批量添加标签

```bash
# 为多个文件添加标签（标签不存在时自动创建）
cbi repository file-tag-add --repository-id 1 --file-ids 123,456,789 --tags "游戏,新素材"

# JSON 格式输出
cbi repository file-tag-add --repository-id 1 --file-ids 123,456 --tags "优质" --format json
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）
- `--tags`: 标签名称列表（逗号分隔，必填）

### 移除文件标签

```bash
cbi repository file-tag-remove --repository-id 1 --file-id 123 --tag-ids 5,10
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--tag-ids`: 标签 ID 列表（逗号分隔，必填）

### 批量添加文件到文件夹

```bash
# 将多个文件添加到多个文件夹
cbi repository file-folder-add --repository-id 1 --file-ids 123,456,789 --folder-ids 100,101

# JSON 格式输出
cbi repository file-folder-add --repository-id 1 --file-ids 123,456 --folder-ids 100 --format json
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）
- `--folder-ids`: 文件夹 ID 列表（逗号分隔，必填）

### 更新文件名称

```bash
cbi repository file-name-update --repository-id 1 --file-id 123 --name "新名称"
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--name`: 新文件名称（必填）

### 更新文件备注

```bash
# 设置备注
cbi repository file-notes-update --repository-id 1 --file-id 123 --notes "这是备注内容"

# 清空备注
cbi repository file-notes-update --repository-id 1 --file-id 123 --notes ""
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--notes`: 备注内容（空字符串表示清空）

### 更新文件评分

```bash
cbi repository file-score-update --repository-id 1 --file-id 123 --score 4
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--score`: 评分（1-5，必填）

### 添加关联产品

```bash
# 添加单个产品
cbi repository file-product-add --repository-id 1 --file-id 123 --products "产品A"

# 批量添加多个产品
cbi repository file-product-add --repository-id 1 --file-id 123 --products "产品A,产品B"

# 指定产品类型和 URL
cbi repository file-product-add --repository-id 1 --file-id 123 --products "游戏A" --product-type 2 --product-url "https://example.com"
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--products`: 产品名称列表（逗号分隔，必填）
- `--product-type`: 产品类型（1=应用，2=游戏，3=商品，默认 2）
- `--product-url`: 产品 URL（可选）
- `--product-img`: 产品图片 URL（可选）
- `--product-desc`: 产品描述（可选）

业务逻辑：
- 根据产品名称在档案库中查找
- 存在同名产品 → 直接关联
- 不存在 → 创建新产品后关联
- 已关联的产品不会重复关联

### 移除关联产品

```bash
cbi repository file-product-remove --repository-id 1 --file-id 123 --product-ids 10,15
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--product-ids`: 产品 ID 列表（逗号分隔，必填）

### 删除产品

```bash
cbi repository product-delete --repository-id 1 --product-ids 10,15,20
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--product-ids`: 产品 ID 列表（逗号分隔，必填）

### 删除文件到回收站

```bash
cbi repository file-delete --repository-id 1 --file-ids 123,124,125
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）

注意：文件删除后进入回收站，可在回收站恢复或彻底删除。

### 上传文件

```bash
# 基本上传（默认跳过重复文件）
cbi repository file-create --repository-id 1 --file ./image.png

# 上传到指定文件夹
cbi repository file-create --repository-id 1 --file ./video.mp4 --folder-id 123

# 强制上传重复文件（需用户确认）
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
- `--force`: 强制上传重复文件

上传流程：
1. 检查文件大小（限制 100MB）
2. 计算 MD5 进行查重
3. **如果文件重复，默认跳过上传**
4. 如需上传重复文件，询问用户确认后使用 `--force` 参数

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
| `json` | JSON 格式 | 程序处理、API 对接 |

---

## 配置文件

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

## 开发

```bash
make build      # 构建当前平台
make build-all  # 跨平台编译（用于发布）
make run        # 运行开发模式
make test       # 运行测试
```

---

## 许可证

Copyright © CreatiBI
