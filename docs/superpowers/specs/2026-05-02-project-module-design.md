# 专案模块设计

## 概述

新增 `cbi project` 命令组，实现专案列表、创建、脚本列表、素材列表四个功能。

## 命令结构

```
cbi project
├── list              # 获取专案列表
├── create            # 创建专案
├── script-list       # 获取专案脚本列表
└── material-list     # 获取专案素材列表
```

## API 接口

### 1. 获取专案列表
- **端点**: POST `/openapi/v1/project/list`
- **请求参数**: page, pageSize, keyword, teamIds[], portfolioIds[], scope (0=所有可见, 1=我加入的)
- **响应**: projects[], total, page, pageSize

### 2. 创建专案
- **端点**: POST `/openapi/v1/project/create`
- **请求参数**: teamId(必填), name(必填), privacy, description, templateId, deadlineStart, deadlineEnd
- **响应**: projectId, name

### 3. 获取专案脚本列表
- **端点**: POST `/openapi/v1/project/script/list`
- **请求参数**: projectId(必填), page, pageSize, keyword, state, parentId, isArchived
- **响应**: scripts[], total, page, pageSize, fields[]
- **脚本字段**: id, name, state, creator, assignedWriter, assignedDesigner, dueDate, createdAt, parentId, currentVersionNo, tableIdValue, aiGenerate

### 4. 获取专案素材列表
- **端点**: POST `/openapi/v1/project/material/list`
- **请求参数**: projectId(必填), page, pageSize, keyword
- **响应**: materials[], total, page, pageSize
- **素材字段**: id, name, fileType(1=视频,2=图片), format, duration, resolution, cover, playUrl, ratio, fileSize, rating, status, scriptId, creator, producer, tags[], createdAt

## 数据结构定义

### internal/client/project.go

```go
// Project 专案
type Project struct {
    ID        int64        `json:"id"`
    Name      string       `json:"name"`
    Creator   *CreatorInfo `json:"creator,omitempty"`
    CreatedAt string       `json:"createdAt"`
}

// ProjectListRequest 专案列表请求
type ProjectListRequest struct {
    Page         int
    PageSize     int
    Keyword      string
    TeamIds      []int64
    PortfolioIds []int64
    Scope        int  // 0=所有可见, 1=我加入的
}

// ProjectListResult 专案列表结果
type ProjectListResult struct {
    Projects   []Project `json:"projects"`
    Total      int64     `json:"total"`
    Page       int       `json:"page"`
    PageSize   int       `json:"pageSize"`
}

// ProjectCreateRequest 创建专案请求
type ProjectCreateRequest struct {
    TeamId        int64
    Name          string
    Privacy       int    // 1=公开(默认), 2=私有
    Description   string
    TemplateId    int64
    DeadlineStart string  // YYYY-MM-DD
    DeadlineEnd   string  // YYYY-MM-DD
}

// ProjectCreateResult 创建专案结果
type ProjectCreateResult struct {
    ProjectId int64  `json:"projectId"`
    Name      string `json:"name"`
}

// Script 脚本
type Script struct {
    ID              int64              `json:"id"`
    Name            string             `json:"name"`
    State           int                `json:"state"`
    Creator         *CreatorInfo       `json:"creator,omitempty"`
    AssignedWriter  *CreatorInfo       `json:"assignedWriter,omitempty"`
    AssignedDesigner *CreatorInfo      `json:"assignedDesigner,omitempty"`
    DueDate         string             `json:"dueDate"`
    CreatedAt       string             `json:"createdAt"`
    ParentId        int64              `json:"parentId"`
    CurrentVersionNo int               `json:"currentVersionNo"`
    TableIdValue    int64              `json:"tableIdValue"`
    AiGenerate      int                `json:"aiGenerate"`
    CustomFields     map[string]string `json:"customFields"`  // 自定义字段值，key=fieldName，value=JSON字符串
}

// ScriptListRequest 脚本列表请求
type ScriptListRequest struct {
    ProjectId  int64
    Page       int
    PageSize   int
    Keyword    string
    State      int  // 任务状态筛选
    ParentId   int64  // 父任务筛选，0=不过滤
    IsArchived int  // 0=不过滤, 1=档案, 2=非档案
}

// ScriptListResult 脚本列表结果
type ScriptListResult struct {
    Scripts   []Script    `json:"scripts"`
    Total     int64       `json:"total"`
    Page      int         `json:"page"`
    PageSize  int         `json:"pageSize"`
    Fields    []FieldDef  `json:"fields"`  // 字段定义列表
}

// FieldDef 字段定义（脚本和素材共用）
type FieldDef struct {
    FieldName     string `json:"fieldName"`     // 字段名称
    ViewName      string `json:"viewName"`      // 显示名称
    FieldType     int    `json:"fieldType"`     // 字段类型
    Classify      int    `json:"classify"`      // 分类：1=固定字段，2=固有字段，3=自定义字段
    IsShow        int    `json:"isShow"`        // 是否显示
    FieldSettings string `json:"fieldSettings"` // 字段配置（JSON字符串）
    IsLazy        int    `json:"isLazy"`        // 是否懒加载
}

// Material 素材
type Material struct {
    ID           int64              `json:"id"`
    Name         string             `json:"name"`
    FileType     int                `json:"fileType"`    // 1=视频, 2=图片
    Format       string             `json:"format"`
    Duration     string             `json:"duration"`
    Resolution   string             `json:"resolution"`
    Cover        string             `json:"cover"`
    PlayUrl      string             `json:"playUrl"`
    Ratio        float64            `json:"ratio"`
    FileSize     int64              `json:"fileSize"`
    Rating       int                `json:"rating"`
    Status       int                `json:"status"`
    ScriptId     int64              `json:"scriptId"`
    Creator      *CreatorInfo       `json:"creator,omitempty"`
    Producer     *CreatorInfo       `json:"producer,omitempty"`
    Tags         []Tag              `json:"tags"`
    CreatedAt    string             `json:"createdAt"`
    CustomFields map[string]string  `json:"customFields"` // 自定义字段值，key=fieldName，value=JSON字符串
}

// MaterialListRequest 素材列表请求
type MaterialListRequest struct {
    ProjectId int64
    Page      int
    PageSize  int
    Keyword   string
}

// MaterialListResult 素材列表结果
type MaterialListResult struct {
    Materials []Material  `json:"materials"`
    Total     int64       `json:"total"`
    Page      int         `json:"page"`
    PageSize  int         `json:"pageSize"`
    Fields    []FieldDef  `json:"fields"`  // 字段定义列表
}
```

## client 层方法

```go
// ListProjects 获取专案列表
func (c *ProjectClient) ListProjects(ctx context.Context, req *ProjectListRequest) (*ProjectListResult, error)

// CreateProject 创建专案
func (c *ProjectClient) CreateProject(ctx context.Context, req *ProjectCreateRequest) (*ProjectCreateResult, error)

// ListScripts 获取专案脚本列表
func (c *ProjectClient) ListScripts(ctx context.Context, req *ScriptListRequest) (*ScriptListResult, error)

// ListMaterials 获取专案素材列表
func (c *ProjectClient) ListMaterials(ctx context.Context, req *MaterialListRequest) (*MaterialListResult, error)
```

## cmd 层命令

### cmd/project.go

新增文件，定义 `projectCmd` 及其子命令：

- `projectListCmd` — 专案列表，参数：--keyword, --team-ids, --portfolio-ids, --scope, --page, --pageSize
- `projectCreateCmd` — 创建专案，参数：--team-id(必填), --name(必填), --privacy, --description, --template-id, --deadline-start, --deadline-end
- `projectScriptListCmd` — 脚本列表，参数：--project-id(必填), --keyword, --state, --parent-id, --is-archived, --page, --pageSize
- `projectMaterialListCmd` — 素材列表，参数：--project-id(必填), --keyword, --page, --pageSize

## 输出格式

- **table**: 表格输出（默认）
- **json**: JSON 格式
- **markdown**: Markdown 格式

### 表格输出示例

**专案列表**:
```
共 10 条，第 1/1 页

ID    名称          创建者    创建时间
1     品牌A投放     张三      2026-5-1
2     品牌B推广     李四      2026-5-2
```

**脚本列表**:
```
共 5 条，第 1/1 页

ID    名称      状态    编剧    设计师    截止日期
1     脚本A     进行中  李四    王五      2026-5-15
```

**素材列表**:
```
共 8 条，第 1/1 页

ID    名称      类型    格式    时长      创建者
1     素材A     视频    mp4    00:00:30  张三
```

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/client/project.go` | 新增：ProjectClient 和数据结构、API 方法 |
| `cmd/project.go` | 新增：project 命令组和 Cobra 子命令 |

## 认证

所有接口需要 `user-access-token` header，OAuth2 scope: `project`（或沿用 repository scope）

## 错误处理

沿用现有模式：
- 500 错误检查 token 过期
- 业务错误返回结构化 CLI 错误
- 参数校验返回 MISSING_* 错误