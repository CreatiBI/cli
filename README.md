# CreatiBI CLI

CreatiBI 命令行工具，用于广告素材驱动的买量解决方案。

支持 OAuth 登录、素材库管理、专案管理和专案集管理，通过命令行安全、便捷地管理创意素材与投放流程。

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
- `cbi project`、`专案`、`专案列表`
- `cbi portfolio`、`专案集`、`专案集列表`

## 快速开始

```bash
# 1. 初始化配置（首次使用）
cbi config init

# 2. OAuth 登录
cbi auth login

# 3. 查看素材库列表
cbi repository list

# 4. 查看专案集列表
cbi portfolio list

# 5. 查看专案列表
cbi project list

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
    ├── file-product-remove   # 移除关联产品
    ├── highlight-clip-list   # 爆点片段列表
    └── highlight-clip-detail # 爆点片段详情
├── project                   # 专案管理
│   ├── list                  # 专案列表
│   ├── create                # 创建专案
│   ├── script-list           # 脚本列表
│   ├── script-create         # 创建脚本任务
│   ├── material-list         # 素材列表
│   └── material              # 素材操作
│       ├── fission-from-task        # 从脚本创建裂变素材
│       ├── derivative-from-task     # 从脚本创建衍生素材
│       ├── fission-from-material    # 从素材创建裂变子素材
│       └── derivative-from-material # 从素材创建衍生子素材
├── portfolio                 # 专案集管理
│   ├── list                  # 专案集列表
│   └── project-list          # 专案集下的专案列表
```

---

## 配置模块 (config)

### 初始化配置

首次使用需要初始化应用凭证配置：

```bash
# 初始化配置（交互式选择模式）
cbi config init

# 强制重新初始化（覆盖已有配置）
cbi config init --new

# 直接使用设备码模式（适用于 VPS/服务器）
cbi config init --device
```

**支持两种初始化模式：**

| 模式 | 适用场景 | 说明 |
|------|---------|------|
| 回调模式 | 桌面环境（macOS/Windows/Linux） | 本地浏览器创建凭证，自动回传 |
| 设备码模式 | VPS/服务器环境 | 远程浏览器创建凭证，轮询获取 |

**回调模式流程：**
1. CLI 启动本地回调服务器（端口 8080）
2. 自动打开浏览器访问开放平台
3. 用户在开放平台创建/选择应用
4. 凭证自动回传到 CLI
5. 配置写入 `~/.cbi/config.json`

**设备码模式流程：**
1. CLI 向平台请求设备码
2. CLI 显示验证 URL 和验证码（如 `550e-8400`）
3. 用户在任意浏览器访问验证 URL 并确认授权
4. CLI 轮询等待授权（15 分钟有效期）
5. 授权成功后获取 app_id/app_secret
6. 配置写入 `~/.cbi/config.json`

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
- AI 视频分析结果（analysis）
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

#### AI 视频分析结果（analysis）

`analysis` 字段是 AI 视频理解的完整分析结果（JSON 字符串），通过 `--format json` 可查看完整结构：

```json
{
  "overall_analysis": "整体分析文本",
  "tags": {
    "industry": ["电商"],
    "style": ["真人实拍"],
    "emotion": ["温馨"]
  },
  "shots": [
    {
      "start": "00:00:00",
      "end": "00:00:03",
      "tags": {...}
    }
  ],
  "signals": [
    {
      "signalName": "脚本结构",
      "signalType": 2,
      "signalContent": "{...}"
    }
  ],
  "creative_strategy": {
    "storyboard_reproduction": [
      {
        "segment_start_time": "00:00:00",
        "segment_end_time": "00:00:03",
        "visual_description": "...",
        "audio_info": "...",
        "action_description": "...",
        "creation_prompt": "...",
        "video_id": 0,
        "cover_url": "https://..."
      }
    ]
  },
  "script_type": "短视频"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `overall_analysis` | string | 整体视频分析描述 |
| `tags` | object | 视频标签（行业、风格、情感等） |
| `shots` | array | 镜头分段信息（时间区间 + 标签） |
| `signals` | array | 信号列表，比外层 signals 字段更完整，含 signalType、结构化 signalContent |
| `signals[].signalType` | int32 | 信号类型编号（1=品牌产品/画面/声音，2=脚本结构，3=前三秒分析等） |
| `signals[].signalContent` | string/json | 信号内容，V2 格式为 JSON 结构体 |
| `creative_strategy` | object | 创意策略，含分镜复现（storyboard_reproduction） |
| `creative_strategy.storyboard_reproduction` | array | 分镜复现列表，每个包含时间区间、画面描述、音频、动作、创作提示词、封面 |
| `script_type` | string | 脚本类型 |

**注意事项：**
- `analysis` 字段来源：`ai_content_analysis` 表，通过 `repository_file.hash` → `ai_content_analysis.file_hash` 关联
- 当文件没有 AI 分析结果时，`analysis` 返回空字符串
- `analysis` 与外层 `signals` 的区别：`signals` 是从 `repository_file_analysis.signals` 解析的扁平化信号摘要，`analysis` 是完整的原始 AI 分析结果，包含 signals 的更丰富版本、creative_strategy、shots 等完整数据
- V1/V2 格式兼容：V1 的 `signalContent` 为 string，V2 为 JSON 结构体。V2 判定依据是 `creative_strategy` 非空或 `signalContent` 以 `{` 或 `[` 开头

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

### 获取爆点片段列表

```bash
# 列出素材库所有爆点片段
cbi repository highlight-clip-list --repository-id 1

# 搜索关键词（匹配爆点片段名称）
cbi repository highlight-clip-list --repository-id 1 --keyword "高光"

# 筛选指定来源视频的爆点片段
cbi repository highlight-clip-list --repository-id 1 --source-video-id 456

# 分页查询
cbi repository highlight-clip-list --repository-id 1 --page 2 --pageSize 30

# JSON 格式
cbi repository highlight-clip-list --repository-id 1 --format json
```

参数：
- `--repository-id`: 素材库 ID（必填）
- `--keyword`: 搜索关键词（匹配爆点片段名称）
- `--source-video-id`: 来源视频 ID（筛选指定视频下的爆点片段）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

输出字段：
- ID、名称、播放地址、封面、时长
- 分析信息（analysisInfo）
- 生成方式（generateType）：1=AI生成，0=手动
- 来源视频信息（可能为空）
- 片段范围（clipStartSec、clipEndSec）
- 创建者信息（可能为空）

### 获取爆点片段详情

```bash
# 获取爆点片段详情
cbi repository highlight-clip-detail 123

# JSON 格式输出
cbi repository highlight-clip-detail 123 --format json
```

输出信息包括：
- 基本信息：ID、名称、格式、大小、时长、分辨率、比例、帧率、评分、备注
- 爆点特有：爆点标题、爆点分析、生成方式、片段范围（秒和帧）
- 来源视频信息（可能为空）
- 创建者信息（可能为空）
- 标签列表
- 关联产品列表
- 各种 URL：封面、原始文件、预览

注意：
- clipId 必须是爆点片段类型的文件
- 如果文件不是爆点片段，会返回错误

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

## 专案模块 (project)

### 列出专案

```bash
# 列出所有可见专案
cbi project list

# 搜索关键词
cbi project list --keyword "品牌"

# 筛选我加入的专案
cbi project list --scope 1

# 分页查询
cbi project list --page 2 --pageSize 30

# JSON 格式
cbi project list --format json
```

参数：
- `--keyword`: 搜索关键词
- `--team-ids`: 团队 ID 列表（逗号分隔）
- `--portfolio-ids`: 作品集 ID 列表（逗号分隔）
- `--scope`: 范围筛选（0=所有可见, 1=我加入的）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

### 创建专案

```bash
# 创建公开专案
cbi project create --team-id 1 --name "品牌投放"

# 创建私有专案
cbi project create --team-id 1 --name "新品推广" --privacy 2

# 完整参数
cbi project create \
  --team-id 1 \
  --name "春节活动" \
  --privacy 1 \
  --description "春节期间投放素材" \
  --deadline-start 2026-01-15 \
  --deadline-end 2026-02-15
```

参数：
- `--team-id`: 团队 ID（必填）
- `--name`: 专案名称（必填）
- `--privacy`: 隐私设置（1=公开, 2=私有，默认 1）
- `--description`: 专案描述
- `--template-id`: 模板 ID
- `--deadline-start`: 截止日期开始（YYYY-MM-DD）
- `--deadline-end`: 截止日期结束（YYYY-MM-DD）

### 列出脚本

```bash
# 列出专案所有脚本
cbi project script-list --project-id 1

# 搜索关键词
cbi project script-list --project-id 1 --keyword "广告"

# 筛选状态
cbi project script-list --project-id 1 --state 2

# 分页查询
cbi project script-list --project-id 1 --page 2 --pageSize 30

# JSON 格式
cbi project script-list --project-id 1 --format json
```

参数：
- `--project-id`: 专案 ID（必填）
- `--keyword`: 搜索关键词
- `--state`: 任务状态筛选
- `--parent-id`: 父任务筛选
- `--is-archived`: 档案筛选（0=不过滤, 1=档案, 2=非档案）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

脚本状态：
- 1 = 待处理
- 2 = 进行中
- 3 = 已完成
- 4 = 已归档

### 列出素材

```bash
# 列出专案所有素材
cbi project material-list --project-id 1

# 搜索关键词
cbi project material-list --project-id 1 --keyword "视频"

# 分页查询
cbi project material-list --project-id 1 --page 2 --pageSize 30

# JSON 格式
cbi project material-list --project-id 1 --format json
```

参数：
- `--project-id`: 专案 ID（必填）
- `--keyword`: 搜索关键词
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

素材类型：
- 1 = 视频
- 2 = 图片

### 创建脚本任务

```bash
# 创建脚本任务
cbi project script-create --project-id 1 --name "脚本任务名称"

# 创建子任务（裂变场景）
cbi project script-create --project-id 1 --name "子任务" --parent-id 100

# 创建衍生任务
cbi project script-create --project-id 1 --name "衍生任务" --source-object "原任务ID"
```

参数：
- `--project-id`: 专案 ID（必填）
- `--name`: 脚本任务名称（必填）
- `--parent-id`: 父任务 ID（可选，裂变场景）
- `--source-object`: 来源对象（可选，衍生场景）

### 从脚本创建裂变素材

裂变素材与脚本为父子关系，素材的 parentId 指向父脚本 ID，必须在同一专案内。

```bash
cbi project material fission-from-task --project-id 1 --script-id 100 --name "裂变素材"
```

参数：
- `--project-id`: 专案 ID（必填）
- `--script-id`: 来源脚本 ID（必填）
- `--name`: 素材名称（必填）

### 从脚本创建衍生素材

衍生素材与脚本为平级关系，素材的 parentId 为 0，sourceObject 指向原脚本 ID，可跨专案。

```bash
# 同专案衍生
cbi project material derivative-from-task --project-id 1 --script-id 100 --name "衍生素材"

# 跨专案衍生
cbi project material derivative-from-task --project-id 2 --script-id 100 --name "衍生素材"
```

参数：
- `--project-id`: 目标专案 ID（必填，可不同于来源专案）
- `--script-id`: 来源脚本 ID（必填）
- `--name`: 素材名称（必填）

### 从素材创建裂变子素材

裂变子素材与原素材为父子关系，parentId 指向原素材 ID，必须在同一专案内。

```bash
cbi project material fission-from-material --project-id 1 --material-id 200 --name "裂变子素材"
```

参数：
- `--project-id`: 专案 ID（必填，同原素材所在专案）
- `--material-id`: 来源素材 ID（必填）
- `--name`: 新素材名称（必填）

### 从素材创建衍生子素材

衍生子素材与原素材为平级关系，parentId 为 0，sourceObject 指向原素材 ID，可跨专案。

```bash
# 同专案衍生
cbi project material derivative-from-material --project-id 1 --material-id 200 --name "衍生子素材"

# 跨专案衍生
cbi project material derivative-from-material --project-id 2 --material-id 200 --name "衍生子素材"
```

参数：
- `--project-id`: 目标专案 ID（必填，可跨专案）
- `--material-id`: 来源素材 ID（必填）
- `--name`: 新素材名称（必填）

**customFields 说明：**

脚本和素材都支持自定义字段（customFields），以 `map<string, string>` 形式返回，key 为字段名称（fieldName），value 为 JSON 字符串。字段定义（fields）随列表返回，包含 classify 字段标识分类：

| Classify | 说明 |
|----------|------|
| 1 | 固定字段 |
| 2 | 固有字段 |
| 3 | 自定义字段 |

---

## 专案集模块 (portfolio)

### 列出专案集

```bash
# 列出所有可见专案集
cbi portfolio list

# 搜索关键词
cbi portfolio list --keyword "品牌"

# 筛选我加入的专案集
cbi portfolio list --scope 1

# 分页查询
cbi portfolio list --page 2 --pageSize 30

# JSON 格式
cbi portfolio list --format json
```

参数：
- `--keyword`: 搜索关键词
- `--scope`: 范围筛选（0=所有可见, 1=我加入的）
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

可见性：
- 1 = 公开
- 2 = 私有

### 列出专案集下的专案

```bash
# 列出专案集所有专案
cbi portfolio project-list --portfolio-id 1

# 搜索关键词
cbi portfolio project-list --portfolio-id 1 --keyword "投放"

# 分页查询
cbi portfolio project-list --portfolio-id 1 --page 2 --pageSize 30

# JSON 格式
cbi portfolio project-list --portfolio-id 1 --format json
```

参数：
- `--portfolio-id`: 专案集 ID（必填）
- `--keyword`: 搜索关键词
- `--page`: 页码（默认 1）
- `--pageSize`: 每页条数（默认 20，最大 50）

专案状态：
- 1 = 正常进行
- 2 = 有风险
- 3 = 偏离轨道
- 4 = 暂停
- 5 = 完成
- 6 = 无更新

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
