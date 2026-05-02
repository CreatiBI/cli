---
name: cbi-repo
description: 使用 CreatiBI CLI（cbi）管理素材库文件、文件夹、标签、关联产品、视频理解信号、AI 视频分析结果和爆点片段；在用户提到上传到素材库、查询文件列表或详情、查看信号或分镜表、维护素材元数据、先初始化 cbi 或先登录再操作等场景时使用。
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
depends_on:
  - cbi-shared
---

# CreatiBI CLI 素材库管理

负责素材库的查询、上传与元数据维护。详情字段中的 `signals` 和 `analysis` 定义见 [references/video-intelligence.md](references/video-intelligence.md)。

## 交互规范

- 对用户只输出结果、状态和下一步，不展示命令、参数或错误原文。
- 先处理 `cbi-shared` 前置条件，再执行素材库操作。
- 涉及重复上传、删除、覆盖等操作时，先用自然语言确认。
- 读取详情时，优先返回摘要；只有在用户明确需要时才展开原始结构。

## 处理顺序

1. 先确认已初始化并登录。
2. 再按任务类型选择查询、上传或修改。
3. 文件详情优先看基础信息，再按需展开 `signals`、`analysis` 或爆点片段。
4. 部分失败时继续处理后续条目，最后汇总成功、失败和原因。

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

## 查询规则

- `--keyword` 搜索名称和 `signals`，不是搜索 `analysis`。
- `--has-signals` 只判断是否存在视频理解信号。
- 有条件筛选时保持单一维度，不要把多个筛选模式混在一起。

## 写入操作

- 上传、改名、改备注、改评分、加标签、加文件夹、加产品前，先确认目标素材库和文件 ID。
- 重复上传默认跳过；需要强制写入时先向用户确认。
- 删除类操作默认按最小风险表达，明确说明影响范围。

## 内部命令骨架

```bash
cbi repository list
cbi repository file-list --repository-id <id>
cbi repository file-detail <file-id>
cbi repository file-check --repository-id <id> --file <path>
cbi repository file-create --repository-id <id> --file <path>
cbi repository highlight-clip-list --repository-id <id>
cbi repository highlight-clip-detail <clip-id>
```

## 参考

- [references/video-intelligence.md](references/video-intelligence.md)
- 仓库根目录 `README.md` 的素材库章节
