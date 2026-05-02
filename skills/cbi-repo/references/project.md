# Project

## 范围

`cbi project` 负责专案维度能力：专案列表、创建专案、专案脚本列表、专案素材列表。

## 常用命令

```bash
cbi project list
cbi project create --team-id <id> --name "<name>"
cbi project script-list --project-id <id>
cbi project material-list --project-id <id>
```

## 参数要点

### project list

- `--keyword`：关键词搜索
- `--scope`：范围筛选（0=所有可见，1=我加入的）
- `--team-ids`：团队 ID 列表（逗号分隔）
- `--portfolio-ids`：作品集 ID 列表（逗号分隔）
- `--page` / `--pageSize`：分页

### project create

- 必填：`--team-id`、`--name`
- 可选：`--privacy`（1=公开，2=私有）、`--description`、`--template-id`
- 可选日期：`--deadline-start`、`--deadline-end`（`YYYY-MM-DD`）

### project script-list

- 必填：`--project-id`
- 常用筛选：`--keyword`、`--state`、`--parent-id`、`--is-archived`
- 状态枚举：1=待处理，2=进行中，3=已完成，4=已归档

### project material-list

- 必填：`--project-id`
- 常用筛选：`--keyword`、`--page`、`--pageSize`
- 素材类型常见枚举：1=视频，2=图片

## customFields

脚本与素材列表都可能返回 `customFields`（`map<string,string>`），并随列表返回 `fields` 定义。

`fields[].classify` 常见枚举：
- `1`：固定字段
- `2`：固有字段
- `3`：自定义字段
