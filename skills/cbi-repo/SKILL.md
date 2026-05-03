---
name: cbi-repo
description: 使用 CreatiBI CLI（cbi）管理素材库、专案和专案集能力，包括文件、文件夹、标签、关联产品、视频理解信号、AI 视频分析结果、爆点片段、专案列表与创建、专案脚本/素材列表、专案集列表与专案集下专案列表；在用户提到上传到素材库、查询文件或专案、查看信号或分镜表、维护素材元数据、查看专案集、先初始化 cbi 或先登录再操作等场景时使用。
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
  - "视频理解信号"
  - "signals"
  - "AI 视频分析"
  - "视频分析结果"
  - "分镜表"
  - "详细分镜"
  - "爆点片段"
  - "highlight clip"
  - "highlight-clip-list"
  - "highlight-clip-detail"
  - "专案"
  - "项目"
  - "project"
  - "cbi project"
  - "project list"
  - "project create"
  - "project script-list"
  - "project script-get"
  - "project script-save"
  - "project script-create"
  - "project material-list"
  - "project material fission-from-task"
  - "project material derivative-from-task"
  - "project material fission-from-material"
  - "project material derivative-from-material"
  - "专案列表"
  - "创建专案"
  - "创建任务"
  - "获取脚本内容"
  - "保存脚本内容"
  - "脚本编辑"
  - "脚本保存"
  - "创建脚本任务"
  - "衍生"
  - "裂变"
  - "裂变素材"
  - "衍生素材"
  - "专案脚本"
  - "专案素材"
  - "专案集"
  - "专案集列表"
  - "portfolio"
  - "cbi portfolio"
  - "portfolio list"
  - "portfolio project-list"
depends_on:
  - cbi-shared
---

# CreatiBI CLI 素材库管理

负责素材库、专案、专案集能力的查询、上传与元数据维护。`signals` 和 `analysis` 定义见 [references/video-intelligence.md](references/video-intelligence.md)，专案字段见 [references/project.md](references/project.md)，专案集字段见 [references/portfolio.md](references/portfolio.md)。

## 交互规范

- 对用户只输出结果、状态和下一步，不展示命令、参数或错误原文。
- 先处理 `cbi-shared` 前置条件，再执行素材库操作。
- 涉及重复上传、删除、覆盖等操作时，先用自然语言确认。
- 读取详情时，优先返回摘要；只有在用户明确需要时才展开原始结构。

## 处理顺序

1. 先确认已初始化并登录。
2. 再按任务类型选择查询、上传或修改。
3. 资产域任务先确认 `repository-id`，专案域任务先确认 `project-id` 或 `team-id`。
4. 文件详情优先看基础信息，再按需展开 `signals`、`analysis` 或爆点片段。
5. 部分失败时继续处理后续条目，最后汇总成功、失败和原因。

## 术语

- `signals`：视频经过 AI 多模态理解后，再由 AI 整理出的结构化信号，用于搜索、筛选和详情展示。
- `analysis`：文件级 AI 视频分析结果，也就是更完整的详细分镜表，通常是 JSON 结构。
- `highlight-clip-*`：独立的爆点片段对象，不等同于文件级 `analysis`。

## 常用能力

| 场景 | 处理要点 |
|------|------|
| 素材库列表 | 选择可写入的目标素材库 |
| 文件列表 | 按文件夹、标签、关键词或是否有信号筛选 |
| 文件详情 | 查看基础信息、来源、标签、信号、AI 分镜表 |
| 文件查重 | 上传前确认是否已存在 |
| 文件上传 | 将本地文件入库 |
| 爆点片段 | 查看 AI 生成或人工维护的高光片段 |
| 专案列表 | 按关键词、可见范围、团队或作品集筛选 |
| 创建专案 | 按团队创建公开或私有专案，并可设置日期范围 |
| 专案脚本列表 | 按状态、关键词、父任务和归档状态筛选 |
| 专案素材列表 | 按关键词检索专案下视频/图片素材 |
| 创建脚本任务 | 创建普通任务、子任务（裂变）或来源任务（衍生） |
| 获取脚本内容 | 读取脚本任务内容与关联信息 |
| 保存脚本内容 | 保存脚本 JSON/Markdown，并可更新关联信息 |
| 脚本转裂变素材 | 从脚本创建同专案父子关系裂变素材 |
| 脚本转衍生素材 | 从脚本创建可跨专案的平级衍生素材 |
| 素材转裂变子素材 | 从素材创建同专案父子关系裂变子素材 |
| 素材转衍生子素材 | 从素材创建可跨专案的平级衍生子素材 |
| 专案集列表 | 按关键词、可见范围和分页筛选 |
| 专案集下专案列表 | 查看某个专案集下的专案并按关键词筛选 |

## 查询规则

- `--keyword` 搜索名称和 `signals`，不是搜索 `analysis`。
- `--has-signals` 只判断是否存在视频理解信号。
- 有条件筛选时保持单一维度，不要把多个筛选模式混在一起。
- 当用户要求“完整 AI 分析 JSON”时，优先使用 JSON 输出查看 `analysis` 原始结构。

## 写入操作

- 上传、改名、改备注、改评分、加标签、加文件夹、加产品前，先确认目标素材库和文件 ID。
- 重复上传默认跳过；需要强制写入时先向用户确认。
- 删除类操作默认按最小风险表达，明确说明影响范围。
- 创建专案前先确认团队 ID、专案名称和隐私级别（1=公开，2=私有）。
- 上传文件可按需带 `name`、`note`、`rating`、`source-url`、`tags` 等扩展参数。

## 专案规则

- `project list` 支持关键词、`scope`、`team-ids`、`portfolio-ids` 和分页。
- `project create` 需要 `team-id` 与 `name`，可选 `privacy`、`description`、`template-id`、`deadline-start`、`deadline-end`。
- `project script-list` 重点筛选字段：`state`、`parent-id`、`is-archived`。
- `project material-list` 重点筛选字段：`keyword` 与分页；素材类型通常为 1=视频、2=图片。
- `project script-create` 支持 `parent-id`（裂变子任务）与 `source-object`（衍生任务来源）。
- `project script-get` 用于读取脚本内容，支持可选 `project-id` 做权限验证。
- `project script-save` 支持 `script`（JSON）与 `markdown` 两种内容写入方式。
- `project script-save` 可同时更新 `name`、`product-ids`、`app-ids`、`ratios`、`ref-repo-file-ids`。
- `project script-save` 在未显式传 `format` 时会按内容自动推导格式。
- `project material fission-from-task`：必须同专案，素材 parentId 指向来源脚本。
- `project material derivative-from-task`：可跨专案，素材 parentId=0，sourceObject 指向来源脚本。
- `project material fission-from-material`：必须同专案，子素材 parentId 指向来源素材。
- `project material derivative-from-material`：可跨专案，子素材 parentId=0，sourceObject 指向来源素材。
- 当用户提到“我加入的专案”，将 `scope` 设为 `1`；未指定时默认全可见范围。

## 专案集规则

- `portfolio list` 支持 `keyword`、`scope`、分页，`scope=1` 表示“我加入的专案集”。
- `portfolio project-list` 需要 `portfolio-id`，可选 `keyword` 与分页。
- 专案集可见性常见枚举：1=公开，2=私有。
- 专案状态常见枚举：1=正常进行，2=有风险，3=偏离轨道，4=暂停，5=完成，6=无更新。

## 内部命令骨架

```bash
cbi repository list
cbi repository file-list --repository-id <id>
cbi repository file-detail <file-id>
cbi repository file-check --repository-id <id> --file <path>
cbi repository file-create --repository-id <id> --file <path>
cbi repository highlight-clip-list --repository-id <id>
cbi repository highlight-clip-detail <clip-id>
cbi project list
cbi project create --team-id <id> --name "<name>"
cbi project script-list --project-id <id>
cbi project material-list --project-id <id>
cbi project script-create --project-id <id> --name "<name>"
cbi project script-get --script-id <id>
cbi project script-save --script-id <id> --script '<json>'
cbi project material fission-from-task --project-id <id> --script-id <id> --name "<name>"
cbi project material derivative-from-task --project-id <id> --script-id <id> --name "<name>"
cbi project material fission-from-material --project-id <id> --material-id <id> --name "<name>"
cbi project material derivative-from-material --project-id <id> --material-id <id> --name "<name>"
cbi portfolio list
cbi portfolio project-list --portfolio-id <id>
```

## 参考

- [references/video-intelligence.md](references/video-intelligence.md)
- [references/project.md](references/project.md)
- [references/portfolio.md](references/portfolio.md)
- 仓库根目录 `README.md` 的素材库章节
