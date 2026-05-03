# Portfolio

## 范围

`cbi portfolio` 负责专案集维度能力：专案集列表、专案集下的专案列表。

## 常用命令

```bash
cbi portfolio list
cbi portfolio project-list --portfolio-id <id>
```

## 参数要点

### portfolio list

- `--keyword`：关键词搜索
- `--scope`：范围筛选（0=所有可见，1=我加入的）
- `--page` / `--pageSize`：分页

可见性常见枚举：
- `1`：公开
- `2`：私有

### portfolio project-list

- 必填：`--portfolio-id`
- 可选：`--keyword`、`--page`、`--pageSize`

专案状态常见枚举：
- `1`：正常进行
- `2`：有风险
- `3`：偏离轨道
- `4`：暂停
- `5`：完成
- `6`：无更新

## 使用规则

- 当用户提到“我加入的专案集”，将 `scope` 设为 `1`。
- 当用户只给专案集名称但未给 ID，先用 `portfolio list --keyword` 定位，再获取 `portfolio-id`。
