---
name: cbi-repo
description: 使用 CreatiBI CLI（cbi）将本地文件上传到素材库，并在需要时完成首次配置初始化与 OAuth 登录。用户提到"上传到素材库""cbi 上传文件""repository file-create""先初始化 cbi""先登录再上传"等场景时使用；适用于图片、视频、文档等需要入库的本地文件上传任务。
trigger:
  - "cbi"
  - "素材库"
  - "上传素材"
  - "上传到素材库"
  - "cbi 上传"
  - "cbi repository"
  - "repository file-create"
  - "repository file-list"
  - "repository file-detail"
  - "repository folder-create"
  - "repository tag-list"
  - "文件查重"
  - "批量添加标签"
  - "批量添加文件夹"
depends_on:
  - cbi-shared
---

# CreatiBI CLI 索材库管理

素材库文件操作模块，依赖 `cbi-shared` 完成配置初始化与认证。

**前置条件：** 执行素材库操作前，需确保已完成：
```bash
cbi config init    # 初始化配置
cbi auth login     # 登录授权
```

---

## 常用命令速查

| 场景 | 命令 |
|------|------|
| 素材库列表 | `cbi repository list` |
| 文件列表 | `cbi repository file-list --repository-id <id>` |
| 上传文件 | `cbi repository file-create --repository-id <id> --file <path>`（重复默认跳过） |
| 文件详情 | `cbi repository file-detail <file-id>` |
| 文件查重 | `cbi repository file-check --repository-id <id> --file <path>` |
| 文件夹列表 | `cbi repository folders --repository-id <id>` |
| 创建文件夹 | `cbi repository folder-create --repository-id <id> --name "名称"` |
| 标签列表 | `cbi repository tag-list --repository-id <id>` |
| 批量添加标签 | `cbi repository file-tag-add --repository-id <id> --file-ids <ids> --tag-ids <ids>` |
| 批量添加到文件夹 | `cbi repository file-folder-add --repository-id <id> --file-ids <ids> --folder-id <id>` |

---

## 索材库列表

```bash
cbi repository list
cbi repo list --format json   # JSON 格式
```

---

## 文件夹管理

### 列出文件夹

```bash
cbi repository folders --repository-id <id>
cbi repository folders --repository-id <id> --with-statistic   # 含文件数统计
cbi repository folders --repository-id <id> --parent-folder-id 100   # 指定父文件夹
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--parent-folder-id`: 父文件夹 ID（0 表示根目录）
- `--with-statistic`: 包含统计信息（文件数量）

### 创建文件夹

```bash
# 创建根目录文件夹
cbi repository folder-create --repository-id <id> --name "新文件夹"

# 创建子文件夹
cbi repository folder-create --repository-id <id> --name "子文件夹" --parent-folder-id 100

# JSON 格式输出
cbi repository folder-create --repository-id <id> --name "新文件夹" --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--name`: 文件夹名称（必填）
- `--parent-folder-id`: 父文件夹 ID（可选，默认根目录）

---

## 标签管理

```bash
# 列出素材库所有标签
cbi repository tag-list --repository-id <id>

# 包含使用次数统计
cbi repository tag-list --repository-id <id> --with-refcnt

# JSON 格式
cbi repository tag-list --repository-id <id> --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--with-refcnt`: 包含标签使用次数

---

## 文件列表查询

```bash
# 列出素材库所有文件
cbi repository file-list --repository-id <id>

# 筛选模式（oneof，不可组合）
cbi repository file-list --repository-id <id> --folder-id 10      # 按文件夹
cbi repository file-list --repository-id <id> --tag-id 5          # 按标签
cbi repository file-list --repository-id <id> --keyword "广告"    # 搜索关键词（名称+signals）
cbi repository file-list --repository-id <id> --has-signals true  # 按视频理解信号筛选

# 分页查询
cbi repository file-list --repository-id <id> --page 2 --pageSize 30

# JSON 格式
cbi repository file-list --repository-id <id> --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

**筛选模式说明：**
- `--folder-id`: 按文件夹筛选
- `--tag-id`: 按标签筛选
- `--keyword`: 搜索名称和 signals
- `--has-signals`: 按是否有视频理解信号筛选

---

## 文件详情

```bash
cbi repository file-detail <file-id>
cbi repository file-detail <file-id> --format json
cbi repository file-detail <file-id> -q   # 静默模式（只输出 JSON）
```

**输出信息包括：**
- 基本信息：ID、名称、格式、大小、时长、分辨率、比例、帧率、评分、备注
- 来源信息：来源平台、来源 URL
- 关联产品、标签、文件夹
- 创建者信息
- 视频理解信号（signals）
- 各种 URL：封面、原始文件、预览

---

## 文件查重

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

---

## 上传文件

```bash
# 基本上传（默认跳过重复文件）
cbi repository file-create --repository-id <id> --file ./image.png

# 上传到指定文件夹
cbi repository file-create --repository-id <id> --file ./video.mp4 --folder-id 123

# 强制上传重复文件（需用户确认）
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
- `--force`: 强制上传重复文件

**上传流程：**
1. 检查文件大小（限制 100MB）
2. 计算 MD5 进行查重
3. **如果文件重复，默认跳过上传**
4. 如需上传重复文件，询问用户确认后使用 `--force` 参数

---

## 批量操作

### 批量添加标签

```bash
cbi repository file-tag-add --repository-id <id> --file-ids 123,456,789 --tag-ids 5,10

# JSON 格式输出
cbi repository file-tag-add --repository-id <id> --file-ids 123,456 --tag-ids 5 --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）
- `--tag-ids`: 标签 ID 列表（逗号分隔，必填）

### 批量添加文件到文件夹

```bash
cbi repository file-folder-add --repository-id <id> --file-ids 123,456,789 --folder-id 100

# JSON 格式输出
cbi repository file-folder-add --repository-id <id> --file-ids 123,456 --folder-id 100 --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）
- `--folder-id`: 目标文件夹 ID（必填）

---

## 通用参数

| 参数 | 说明 |
|------|------|
| `--config <path>` | 配置文件路径（默认 ~/.cbi/config.json） |
| `-f, --format` | 输出格式：json / table |
| `-o, --output` | 输出到文件 |
| `-q, --quiet` | 只输出数据，无日志 |
| `-v, --verbose` | 显示详细信息 |

---

## 错误处理

### 权限不足/未登录

遇到权限错误时，先执行认证流程（参见 `cbi-shared`）：
```bash
cbi config show    # 检查配置
cbi config init    # 初始化（如不存在）
cbi auth login     # 登录
```

### Token 过期

```bash
cbi auth login   # 重新登录
```