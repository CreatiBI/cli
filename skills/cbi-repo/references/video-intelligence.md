# Video Intelligence

## 术语

### signals

`signals` 是视频通过 AI 多模态理解后，再由 AI 整理出来的结构化内容。

常见用途：
- 作为文件列表关键词搜索的扩展命中源
- 作为文件详情里的可读摘要
- 作为后续人工运营、标签整理和筛选的依据

### analysis

`analysis` 是文件级 AI 视频分析结果。
它对应更完整的详细分镜表，通常以 JSON 字符串存储。

常见用途：
- 查看完整镜头/分镜结构
- 了解时间线、场景切换、镜头节奏
- 在文件详情里做更深的分析读取

### highlight clip

`highlight-clip-list` / `highlight-clip-detail` 是独立的爆点片段对象。
它们更偏向高光切片，不等同于文件级 `analysis`。

## 读取规则

- 文件列表的 `--keyword` 主要匹配名称和 `signals`。
- 文件列表的 `--has-signals` 只判断是否存在 `signals`。
- 文件详情里如果 `analysis` 很长，优先让用户查看 JSON 输出或完整详情。
- 当用户说“分镜表”“AI 视频分析结果”“详细镜头分析”时，优先映射到 `analysis`。

## analysis 结构提示

- `analysis` 是完整 AI 分析结果（JSON 字符串），常见字段包括：
  - `overall_analysis`
  - `tags`
  - `shots`
  - `signals`（更完整版本）
  - `creative_strategy.storyboard_reproduction`
  - `script_type`
- `analysis` 可能为空字符串，表示该文件没有 AI 分析结果。
- `analysis` 与外层 `signals` 不同：外层 `signals` 是扁平摘要，`analysis` 是完整原始结构。

## 版本兼容

- V1：`signalContent` 常见为字符串。
- V2：`signalContent` 常见为 JSON 结构体。
- 可用经验判定：
  - `creative_strategy` 非空，或
  - `signalContent` 以 `{` / `[` 开头
  则可按 V2 结构处理。
