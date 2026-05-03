# Script Format

## script-save 参数

- `--script-id`：脚本任务 ID（必填）
- `--project-id`：专案 ID（可选）
- `--format`：可选；不传则自动推导
- `--name`：脚本名称（可选）
- `--script`：JSON 脚本内容（普通/剪辑）
- `--markdown`：Markdown 内容（普通）
- `--product-ids`：关联产品 ID（逗号分隔）
- `--app-ids`：关联渠道应用 ID（逗号分隔）
- `--ratios`：关联尺寸（逗号分隔）
- `--ref-repo-file-ids`：引用仓库文件 ID（逗号分隔）

## 格式自动推导

当未传 `--format` 时：

- 传入 `script` JSON：
  - 含 `CbiClipItem` 节点 -> `format=4`（剪辑）
  - 无以上节点 -> `format=1`（普通）
- 传入 `markdown`：
  - `format=1`（普通）

## 约束

- 已取消分镜脚本与口播脚本，不使用 `format=2` / `format=3`。
- 新写入默认只使用：
  - `format=1`（普通）
  - `format=4`（剪辑）
- 如读取到历史数据包含 `format=2/3`，按存量兼容读取，不新增这两类内容。

## 建议流程

1. 先 `script-get` 读取当前脚本和关联信息。
2. 再 `script-save` 做最小必要变更。
3. 变更后再 `script-get` 校验结果。
