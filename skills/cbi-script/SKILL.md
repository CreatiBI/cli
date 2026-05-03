---
name: cbi-script
description: 使用 CreatiBI CLI（cbi）在专案内创建脚本任务、获取脚本内容、保存脚本内容，并在裂变/衍生场景下管理脚本上下游关系。用户提到“写脚本”“保存脚本内容”“获取脚本内容”“script-save”“script-get”“script-create”“脚本任务”等场景时使用。
trigger:
  - "写脚本"
  - "创建脚本"
  - "脚本任务"
  - "创建脚本任务"
  - "获取脚本内容"
  - "保存脚本内容"
  - "脚本编辑"
  - "脚本保存"
  - "script-create"
  - "script-get"
  - "script-save"
  - "cbi script"
  - "cbi project script"
  - "专案脚本"
  - "脚本衍生"
  - "脚本裂变"
depends_on:
  - cbi-shared
---

# CreatiBI CLI 脚本写作与保存

聚焦专案脚本读写能力：创建任务、读取脚本、保存脚本，并处理脚本在衍生/裂变链路中的关系。

## 交互规范

- 对用户输出结果与下一步，不展示命令细节。
- 写入前先确认目标对象：`project-id` 或 `script-id`。
- 涉及覆盖或批量关联更新时先确认影响范围。

## 处理顺序

1. 先确认登录状态（依赖 `cbi-shared`）。
2. 创建脚本任务时确认 `project-id` 与任务名称。
3. 保存内容前先获取当前脚本内容（避免误覆盖）。
4. 保存后回读脚本，确认格式和关联信息已生效。

## 能力清单

| 场景 | 处理要点 |
|------|------|
| 创建脚本任务 | `script-create`，支持普通任务/子任务/衍生来源任务 |
| 获取脚本内容 | `script-get`，读取基本信息、关联信息与内容 |
| 保存脚本内容 | `script-save`，支持 JSON/Markdown，支持关联信息更新 |

## 关键规则

- `script-create`：
  - 必填：`project-id`、`name`
  - 可选：`parent-id`（裂变子任务）、`source-object`（衍生来源）
- `script-get`：
  - 必填：`script-id`
  - 可选：`project-id`（权限验证）
- `script-save`：
  - 必填：`script-id`
  - 内容参数二选一或组合更新：`script`（JSON）/`markdown`
  - 可选更新：`name`、`product-ids`、`app-ids`、`ratios`、`ref-repo-file-ids`
  - 格式约束：仅允许普通与剪辑格式（不再使用分镜/口播格式）

## 格式规则

- 格式枚举：`1=普通`、`4=剪辑`。
- 未传 `format` 时按内容自动推导（见 [references/script-format.md](references/script-format.md)）。

## 内部命令骨架

```bash
cbi project script-create --project-id <id> --name "<name>"
cbi project script-get --script-id <id>
cbi project script-save --script-id <id> --script '<json>'
cbi project script-save --script-id <id> --markdown "# 标题"
```

## 参考

- [references/script-format.md](references/script-format.md)
- 仓库根目录 `README.md` 的“专案模块 (project)”章节
