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
  - "repository tag-delete"
  - "repository product-list"
  - "repository file-name-update"
  - "repository file-notes-update"
  - "repository file-score-update"
  - "repository file-product-add"
  - "repository file-tag-remove"
  - "repository file-product-remove"
  - "repository product-delete"
  - "repository file-delete"
  - "文件查重"
  - "批量添加标签"
  - "批量添加文件夹"
  - "修改文件名称"
  - "修改文件备注"
  - "修改文件评分"
  - "添加关联产品"
  - "关联产品"
  - "移除标签"
  - "移除关联产品"
  - "删除产品"
  - "删除文件"
  - "删除标签"
  - "移入回收站"
  - "爆点片段"
  - "highlight clip"
  - "highlight-clip-list"
  - "highlight-clip-detail"
depends_on:
  - cbi-shared
---

# CreatiBI CLI 素材库管理

素材库文件操作模块，依赖 `cbi-shared` 完成配置初始化与认证。

## 交互规范

**核心原则：对用户隐藏命令细节，只呈现结果。**

- 执行 cbi 命令时，**不要**向用户展示正在运行的命令（如 `cbi repository list`）
- **不要**在回复中引用命令语法、参数列表或代码块
- 只用自然语言告诉用户操作结果：如"已找到 3 个素材库"、"文件上传成功，文件 ID: 123"
- 如果操作失败，用自然语言说明原因和建议，**不要**展示错误原文或命令
- 需要用户确认的操作（如重复文件上传），用自然语言提问而非展示命令参数
- 前置条件检查（未登录/未初始化）也用自然语言引导，不展示命令

**前置条件：** 执行素材库操作前，需确保已完成配置初始化与 OAuth 登录（依赖 `cbi-shared` skill 自动处理）。

---

## 常用命令速查

> 以下命令仅供 AI 内部执行参考，**不要**向用户展示命令本身。

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
| 删除标签 | `cbi repository tag-delete --repository-id <id> --tag-ids <ids>` |
| 产品列表 | `cbi repository product-list --repository-id <id>` |
| 修改文件名称 | `cbi repository file-name-update --repository-id <id> --file-id <fid> --name "新名称"` |
| 修改文件备注 | `cbi repository file-notes-update --repository-id <id> --file-id <fid> --notes "备注"` |
| 修改文件评分 | `cbi repository file-score-update --repository-id <id> --file-id <fid> --score 4` |
| 批量添加标签 | `cbi repository file-tag-add --repository-id <id> --file-ids <ids> --tags "标签1,标签2"` |
| 批量添加到文件夹 | `cbi repository file-folder-add --repository-id <id> --file-ids <ids> --folder-ids <ids>` |
| 添加关联产品 | `cbi repository file-product-add --repository-id <id> --file-id <fid> --products "产品A,产品B"` |
| 移除文件标签 | `cbi repository file-tag-remove --repository-id <id> --file-id <fid> --tag-ids <ids>` |
| 移除关联产品 | `cbi repository file-product-remove --repository-id <id> --file-id <fid> --product-ids <ids>` |
| 删除产品 | `cbi repository product-delete --repository-id <id> --product-ids <ids>` |
| 删除文件到回收站 | `cbi repository file-delete --repository-id <id> --file-ids <ids>` |
| 爆点片段列表 | `cbi repository highlight-clip-list --repository-id <id>` |
| 爆点片段详情 | `cbi repository highlight-clip-detail <clip-id>` |

---

## 素材库列表

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

### 删除标签

```bash
# 删除档案库标签（软删除）
cbi repository tag-delete --repository-id <id> --tag-ids 5,10,15
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--tag-ids`: 标签 ID 列表（逗号分隔，必填）

**注意：**
- 删除后标签不再可用（软删除）
- 已关联的文件标签记录保留
- 需要档案库编辑权限

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

## 文件属性修改

### 修改文件名称

```bash
cbi repository file-name-update --repository-id <id> --file-id 123 --name "新文件名称"
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--name`: 新文件名称（必填）

### 修改文件备注

```bash
# 设置备注
cbi repository file-notes-update --repository-id <id> --file-id 123 --notes "这是备注内容"

# 清空备注（空字符串）
cbi repository file-notes-update --repository-id <id> --file-id 123 --notes ""
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--notes`: 备注内容（空字符串表示清空）

### 修改文件评分

```bash
cbi repository file-score-update --repository-id <id> --file-id 123 --score 4
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--score`: 评分（1-5，必填）

---

## 产品管理

### 列出产品

```bash
# 列出素材库所有已关联产品
cbi repository product-list --repository-id <id>

# JSON 格式
cbi repository product-list --repository-id <id> --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）

**产品类型说明：**
- 1 = 应用
- 2 = 游戏
- 3 = 商品

### 添加关联产品

```bash
# 添加单个产品
cbi repository file-product-add --repository-id <id> --file-id 123 --products "产品A"

# 批量添加多个产品
cbi repository file-product-add --repository-id <id> --file-id 123 --products "产品A,产品B"

# 指定产品类型和 URL
cbi repository file-product-add --repository-id <id> --file-id 123 --products "游戏A" --product-type 2 --product-url "https://example.com"
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--products`: 产品名称列表（逗号分隔，必填）
- `--product-type`: 产品类型（1=应用，2=游戏，3=商品，默认 2）
- `--product-url`: 产品 URL（可选）
- `--product-img`: 产品图片 URL（可选）
- `--product-desc`: 产品描述（可选）

**业务逻辑：**
- 根据产品名称在档案库中查找
- 存在同名产品 → 直接关联已存在的产品
- 不存在 → 创建新产品后关联
- 已关联的产品不会重复关联（自动去重）

### 移除文件标签

```bash
cbi repository file-tag-remove --repository-id <id> --file-id 123 --tag-ids 5,10
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--tag-ids`: 标签 ID 列表（逗号分隔，必填）

### 移除文件关联产品

```bash
cbi repository file-product-remove --repository-id <id> --file-id 123 --product-ids 10,15
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-id`: 文件 ID（必填）
- `--product-ids`: 产品 ID 列表（逗号分隔，必填）

### 删除产品

```bash
cbi repository product-delete --repository-id <id> --product-ids 10,15,20
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--product-ids`: 产品 ID 列表（逗号分隔，必填）

---

## 文件删除

### 删除文件到回收站

```bash
cbi repository file-delete --repository-id <id> --file-ids 123,124,125
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--file-ids`: 文件 ID 列表（逗号分隔，必填）

**注意：** 文件删除后进入回收站，可在回收站恢复或彻底删除。

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

### API 500 错误 / Token 过期 / 认证过期

当命令返回 500 错误、TOKEN_EXPIRED 或"认证已过期"等错误时，**通常是授权登录过期导致**。处理流程：

1. 用自然语言告知用户"登录已过期，我来帮你重新登录"
2. 自动调用 `cbi auth login` 重新授权（依赖 `cbi-shared` 处理）
3. 登录成功后，**自动重试**刚才失败的命令
4. 将重试结果用自然语言反馈给用户

如果重试后仍然失败，再排查其他原因。

### 权限不足/未登录

遇到权限错误时，自动引导用户完成认证（依赖 `cbi-shared` 处理），用自然语言提示如"你需要先登录，我来帮你完成"。

---

## 爆点片段管理

### 获取爆点片段列表

```bash
# 列出素材库所有爆点片段
cbi repository highlight-clip-list --repository-id <id>

# 搜索关键词
cbi repository highlight-clip-list --repository-id <id> --keyword "高光"

# 筛选指定来源视频
cbi repository highlight-clip-list --repository-id <id> --source-video-id 456

# 分页查询
cbi repository highlight-clip-list --repository-id <id> --page 2 --pageSize 30

# JSON 格式
cbi repository highlight-clip-list --repository-id <id> --format json
```

**参数：**
- `--repository-id`: 素材库 ID（必填）
- `--keyword`: 搜索关键词（匹配爆点片段名称）
- `--source-video-id`: 来源视频 ID（筛选指定视频下的爆点片段）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

**输出字段：**
- ID、名称、播放地址、封面、时长
- 分析信息（analysisInfo）
- 生成方式（generateType）：1=AI生成，0=手动
- 来源视频信息（可能为空）
- 片段范围（clipStartSec、clipEndSec）
- 创建者信息（可能为空）

### 获取爆点片段详情

```bash
cbi repository highlight-clip-detail <clip-id>
cbi repository highlight-clip-detail <clip-id> --format json
```

**输出信息包括：**
- 基本信息：ID、名称、格式、大小、时长、分辨率、比例、帧率、评分、备注
- 爆点特有：爆点标题、爆点分析、生成方式、片段范围（秒和帧）
- 来源视频信息（可能为空）
- 创建者信息（可能为空）
- 标签列表
- 关联产品列表
- 各种 URL：封面、原始文件、预览

**注意：**
- clipId 必须是爆点片段类型的文件
- 如果文件不是爆点片段，会返回错误