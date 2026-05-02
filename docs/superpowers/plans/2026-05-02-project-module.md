# Project Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 `cbi project` 命令组，支持专案列表、创建、脚本列表、素材列表四个功能。

**Architecture:** client 层封装 API 调用（resty + gjson），cmd 层处理 Cobra 命令、参数验证、输出格式。遵循现有 repository 模块架构模式。

**Tech Stack:** Go, Cobra, resty, gjson, go-pretty/table

---

## File Structure

| 文件 | 职责 |
|------|------|
| `internal/client/project.go` | 新增：ProjectClient、数据结构、API 方法 |
| `cmd/project.go` | 新增：project 命令组和 4 个子命令 |

---

### Task 1: Client Layer - 数据结构和 ListProjects 方法

**Files:**
- Create: `internal/client/project.go`

- [ ] **Step 1: 创建 project.go 文件，定义数据结构**

```go
package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// ProjectClient 专案 API 客户端
type ProjectClient struct {
	client *resty.Client
}

// NewProjectClient 创建专案客户端
func NewProjectClient() *ProjectClient {
	baseURL := config.GetBaseURL()
	return &ProjectClient{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(30 * 1000000000), // 30 秒
	}
}

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
	Scope        int // 0=所有可见, 1=我加入的
}

// ProjectListResult 专案列表结果
type ProjectListResult struct {
	Projects  []Project `json:"projects"`
	Total     int64     `json:"total"`
	Page      int       `json:"page"`
	PageSize  int       `json:"pageSize"`
}

// ProjectCreateRequest 创建专案请求
type ProjectCreateRequest struct {
	TeamId        int64
	Name          string
	Privacy       int    // 1=公开(默认), 2=私有
	Description   string
	TemplateId    int64
	DeadlineStart string // YYYY-MM-DD
	DeadlineEnd   string // YYYY-MM-DD
}

// ProjectCreateResult 创建专案结果
type ProjectCreateResult struct {
	ProjectId int64  `json:"projectId"`
	Name      string `json:"name"`
}

// Script 脚本
type Script struct {
	ID               int64              `json:"id"`
	Name             string             `json:"name"`
	State            int                `json:"state"`
	Creator          *CreatorInfo       `json:"creator,omitempty"`
	AssignedWriter   *CreatorInfo       `json:"assignedWriter,omitempty"`
	AssignedDesigner *CreatorInfo       `json:"assignedDesigner,omitempty"`
	DueDate          string             `json:"dueDate"`
	CreatedAt        string             `json:"createdAt"`
	ParentId         int64              `json:"parentId"`
	CurrentVersionNo int                `json:"currentVersionNo"`
	TableIdValue     int64              `json:"tableIdValue"`
	AiGenerate       int                `json:"aiGenerate"`
	CustomFields     map[string]string  `json:"customFields"` // 自定义字段值，key=fieldName，value=JSON字符串
}

// ScriptListRequest 脚本列表请求
type ScriptListRequest struct {
	ProjectId  int64
	Page       int
	PageSize   int
	Keyword    string
	State      int   // 任务状态筛选
	ParentId   int64 // 父任务筛选，0=不过滤
	IsArchived int   // 0=不过滤, 1=档案, 2=非档案
}

// ScriptListResult 脚本列表结果
type ScriptListResult struct {
	Scripts  []Script   `json:"scripts"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Fields   []FieldDef `json:"fields"` // 字段定义列表
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
	Fields    []FieldDef  `json:"fields"` // 字段定义列表
}
```

- [ ] **Step 2: 实现 ListProjects 方法**

```go
// ListProjects 获取专案列表
func (c *ProjectClient) ListProjects(ctx context.Context, req *ProjectListRequest) (*ProjectListResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"page":     req.Page,
		"pageSize": req.PageSize,
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if len(req.TeamIds) > 0 {
		body["teamIds"] = req.TeamIds
	}
	if len(req.PortfolioIds) > 0 {
		body["portfolioIds"] = req.PortfolioIds
	}
	if req.Scope > 0 {
		body["scope"] = req.Scope
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误（可能 token 过期）
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) {
			return nil, cliErr.ErrAuthExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("PROJECT_LIST_ERROR",
			fmt.Sprintf("获取专案列表失败 (%d)", codeVal), message)
	}

	projects := []Project{}
	result.Get("data.projects").ForEach(func(_, value gjson.Result) bool {
		project := Project{
			ID:        value.Get("id").Int(),
			Name:      value.Get("name").String(),
			CreatedAt: value.Get("createdAt").String(),
		}
		if creator := value.Get("creator"); creator.Exists() {
			project.Creator = &CreatorInfo{
				ID:     creator.Get("id").Int(),
				Name:   creator.Get("name").String(),
				Email:  creator.Get("email").String(),
				Avatar: creator.Get("avatar").String(),
			}
		}
		projects = append(projects, project)
		return true
	})

	return &ProjectListResult{
		Projects:  projects,
		Total:     result.Get("data.total").Int(),
		Page:      int(result.Get("data.page").Int()),
		PageSize:  int(result.Get("data.pageSize").Int()),
	}, nil
}
```

- [ ] **Step 3: 运行测试验证编译**

Run: `go build ./internal/client`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add internal/client/project.go
git commit -m "feat(project): 添加 ProjectClient 数据结构和 ListProjects 方法"
```

---

### Task 2: Client Layer - CreateProject 方法

**Files:**
- Modify: `internal/client/project.go`

- [ ] **Step 1: 实现 CreateProject 方法**

在 `internal/client/project.go` 文件末尾添加：

```go
// CreateProject 创建专案
func (c *ProjectClient) CreateProject(ctx context.Context, req *ProjectCreateRequest) (*ProjectCreateResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"teamId": req.TeamId,
		"name":   req.Name,
	}
	if req.Privacy > 0 {
		body["privacy"] = req.Privacy
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	if req.TemplateId > 0 {
		body["templateId"] = req.TemplateId
	}
	if req.DeadlineStart != "" {
		body["deadlineStart"] = req.DeadlineStart
	}
	if req.DeadlineEnd != "" {
		body["deadlineEnd"] = req.DeadlineEnd
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/create")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) {
			return nil, cliErr.ErrAuthExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("PROJECT_CREATE_ERROR",
			fmt.Sprintf("创建专案失败 (%d)", codeVal), message)
	}

	return &ProjectCreateResult{
		ProjectId: result.Get("data.projectId").Int(),
		Name:      result.Get("data.name").String(),
	}, nil
}
```

- [ ] **Step 2: 运行测试验证编译**

Run: `go build ./internal/client`
Expected: 编译成功，无错误

- [ ] **Step 3: Commit**

```bash
git add internal/client/project.go
git commit -m "feat(project): 添加 CreateProject 方法"
```

---

### Task 3: Client Layer - ListScripts 方法

**Files:**
- Modify: `internal/client/project.go`

- [ ] **Step 1: 实现 ListScripts 方法**

在 `internal/client/project.go` 文件末尾添加：

```go
// ListScripts 获取专案脚本列表
func (c *ProjectClient) ListScripts(ctx context.Context, req *ScriptListRequest) (*ScriptListResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId": req.ProjectId,
		"page":      req.Page,
		"pageSize":  req.PageSize,
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.State > 0 {
		body["state"] = req.State
	}
	if req.ParentId > 0 {
		body["parentId"] = req.ParentId
	}
	if req.IsArchived > 0 {
		body["isArchived"] = req.IsArchived
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/script/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) {
			return nil, cliErr.ErrAuthExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("SCRIPT_LIST_ERROR",
			fmt.Sprintf("获取脚本列表失败 (%d)", codeVal), message)
	}

	scripts := []Script{}
	result.Get("data.scripts").ForEach(func(_, value gjson.Result) bool {
		script := Script{
			ID:               value.Get("id").Int(),
			Name:             value.Get("name").String(),
			State:            int(value.Get("state").Int()),
			DueDate:          value.Get("dueDate").String(),
			CreatedAt:        value.Get("createdAt").String(),
			ParentId:         value.Get("parentId").Int(),
			CurrentVersionNo: int(value.Get("currentVersionNo").Int()),
			TableIdValue:     value.Get("tableIdValue").Int(),
			AiGenerate:       int(value.Get("aiGenerate").Int()),
		}

		// 解析 creator
		if creator := value.Get("creator"); creator.Exists() {
			script.Creator = parseCreatorInfo(creator)
		}
		// 解析 assignedWriter
		if writer := value.Get("assignedWriter"); writer.Exists() {
			script.AssignedWriter = parseCreatorInfo(writer)
		}
		// 解析 assignedDesigner
		if designer := value.Get("assignedDesigner"); designer.Exists() {
			script.AssignedDesigner = parseCreatorInfo(designer)
		}
		// 解析 customFields
		if customFields := value.Get("customFields"); customFields.Exists() {
			script.CustomFields = parseCustomFields(customFields)
		}

		scripts = append(scripts, script)
		return true
	})

	// 解析 fields
	fields := []FieldDef{}
	result.Get("data.fields").ForEach(func(_, value gjson.Result) bool {
		fields = append(fields, FieldDef{
			FieldName:     value.Get("fieldName").String(),
			ViewName:      value.Get("viewName").String(),
			FieldType:     int(value.Get("fieldType").Int()),
			Classify:      int(value.Get("classify").Int()),
			IsShow:        int(value.Get("isShow").Int()),
			FieldSettings: value.Get("fieldSettings").String(),
			IsLazy:        int(value.Get("isLazy").Int()),
		})
		return true
	})

	return &ScriptListResult{
		Scripts:  scripts,
		Total:    result.Get("data.total").Int(),
		Page:     int(result.Get("data.page").Int()),
		PageSize: int(result.Get("data.pageSize").Int()),
		Fields:   fields,
	}, nil
}

// parseCreatorInfo 解析 CreatorInfo
func parseCreatorInfo(value gjson.Result) *CreatorInfo {
	return &CreatorInfo{
		ID:     value.Get("id").Int(),
		Name:   value.Get("name").String(),
		Email:  value.Get("email").String(),
		Avatar: value.Get("avatar").String(),
	}
}

// parseCustomFields 解析 customFields (map[string]string)
func parseCustomFields(value gjson.Result) map[string]string {
	result := make(map[string]string)
	value.ForEach(func(key, val gjson.Result) bool {
		result[key.String()] = val.String()
		return true
	})
	return result
}
```

- [ ] **Step 2: 运行测试验证编译**

Run: `go build ./internal/client`
Expected: 编译成功，无错误

- [ ] **Step 3: Commit**

```bash
git add internal/client/project.go
git commit -m "feat(project): 添加 ListScripts 方法"
```

---

### Task 4: Client Layer - ListMaterials 方法

**Files:**
- Modify: `internal/client/project.go`

- [ ] **Step 1: 实现 ListMaterials 方法**

在 `internal/client/project.go` 文件末尾添加：

```go
// ListMaterials 获取专案素材列表
func (c *ProjectClient) ListMaterials(ctx context.Context, req *MaterialListRequest) (*MaterialListResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId": req.ProjectId,
		"page":      req.Page,
		"pageSize":  req.PageSize,
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/material/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) {
			return nil, cliErr.ErrAuthExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("MATERIAL_LIST_ERROR",
			fmt.Sprintf("获取素材列表失败 (%d)", codeVal), message)
	}

	materials := []Material{}
	result.Get("data.materials").ForEach(func(_, value gjson.Result) bool {
		material := Material{
			ID:         value.Get("id").Int(),
			Name:       value.Get("name").String(),
			FileType:   int(value.Get("fileType").Int()),
			Format:     value.Get("format").String(),
			Duration:   value.Get("duration").String(),
			Resolution: value.Get("resolution").String(),
			Cover:      value.Get("cover").String(),
			PlayUrl:    value.Get("playUrl").String(),
			Ratio:      value.Get("ratio").Float(),
			FileSize:   value.Get("fileSize").Int(),
			Rating:     int(value.Get("rating").Int()),
			Status:     int(value.Get("status").Int()),
			ScriptId:   value.Get("scriptId").Int(),
			CreatedAt:  value.Get("createdAt").String(),
		}

		// 解析 creator
		if creator := value.Get("creator"); creator.Exists() {
			material.Creator = parseCreatorInfo(creator)
		}
		// 解析 producer
		if producer := value.Get("producer"); producer.Exists() {
			material.Producer = parseCreatorInfo(producer)
		}
		// 解析 tags
		if tags := value.Get("tags"); tags.Exists() {
			material.Tags = parseTags(tags)
		}
		// 解析 customFields
		if customFields := value.Get("customFields"); customFields.Exists() {
			material.CustomFields = parseCustomFields(customFields)
		}

		materials = append(materials, material)
		return true
	})

	// 解析 fields
	fields := []FieldDef{}
	result.Get("data.fields").ForEach(func(_, value gjson.Result) bool {
		fields = append(fields, FieldDef{
			FieldName:     value.Get("fieldName").String(),
			ViewName:      value.Get("viewName").String(),
			FieldType:     int(value.Get("fieldType").Int()),
			Classify:      int(value.Get("classify").Int()),
			IsShow:        int(value.Get("isShow").Int()),
			FieldSettings: value.Get("fieldSettings").String(),
			IsLazy:        int(value.Get("isLazy").Int()),
		})
		return true
	})

	return &MaterialListResult{
		Materials: materials,
		Total:     result.Get("data.total").Int(),
		Page:      int(result.Get("data.page").Int()),
		PageSize:  int(result.Get("data.pageSize").Int()),
		Fields:    fields,
	}, nil
}

// parseTags 解析 Tags
func parseTags(value gjson.Result) []Tag {
	tags := []Tag{}
	value.ForEach(func(_, tag gjson.Result) bool {
		tags = append(tags, Tag{
			ID:    tag.Get("id").Int(),
			Name:  tag.Get("name").String(),
			Color: tag.Get("color").String(),
		})
		return true
	})
	return tags
}
```

- [ ] **Step 2: 运行测试验证编译**

Run: `go build ./internal/client`
Expected: 编译成功，无错误

- [ ] **Step 3: Commit**

```bash
git add internal/client/project.go
git commit -m "feat(project): 添加 ListMaterials 方法"
```

---

### Task 5: Cmd Layer - projectCmd 命令组和 projectListCmd

**Files:**
- Create: `cmd/project.go`

- [ ] **Step 1: 创建 project.go 文件，定义 projectCmd 和 projectListCmd**

```go
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// projectCmd 代表 project 命令组
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "专案管理",
	Long:  `管理专案，包括查看专案列表、创建专案、脚本列表、素材列表。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// projectListCmd 专案列表
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可访问的专案",
	Long: `获取权限范围内的专案列表。

示例：
  cbi project list
  cbi project list --keyword "品牌"
  cbi project list --scope 1 --page 1 --pageSize 20`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		teamIdsStr, _ := cmd.Flags().GetString("team-ids")
		portfolioIdsStr, _ := cmd.Flags().GetString("portfolio-ids")
		scope, _ := cmd.Flags().GetInt("scope")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		// 解析 teamIds
		var teamIds []int64
		if teamIdsStr != "" {
			ids, err := parseIDList(teamIdsStr, "team-ids")
			if err != nil {
				return err
			}
			teamIds = ids
		}

		// 解析 portfolioIds
		var portfolioIds []int64
		if portfolioIdsStr != "" {
			ids, err := parseIDList(portfolioIdsStr, "portfolio-ids")
			if err != nil {
				return err
			}
			portfolioIds = ids
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListProjects(ctx, &client.ProjectListRequest{
			Page:         page,
			PageSize:     pageSize,
			Keyword:      keyword,
			TeamIds:      teamIds,
			PortfolioIds: portfolioIds,
			Scope:        scope,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printProjectListTable(cmd, result)
			return nil
		}
	},
}

// printProjectListTable 表格输出专案列表
func printProjectListTable(cmd *cobra.Command, result *client.ProjectListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Projects) == 0 {
		fmt.Fprintln(w, "无专案")
		return
	}

	output.PrintTable(w, []string{"ID", "名称", "创建者", "创建时间"},
		func() [][]string {
			rows := [][]string{}
			for _, p := range result.Projects {
				creator := "-"
				if p.Creator != nil {
					creator = p.Creator.Name
				}
				rows = append(rows, []string{
					strconv.FormatInt(p.ID, 10),
					p.Name,
					creator,
					formatDate(p.CreatedAt),
				})
			}
			return rows
		}())
}

// totalPages 计算总页数
func totalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	pages := int(total) / pageSize
	if int(total) % pageSize > 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	return pages
}

// formatDate 格式化日期（截取 YYYY-MM-DD 部分）
func formatDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
```

- [ ] **Step 2: 添加 projectListCmd 参数**

在 `init()` 函数中添加：

```go
func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)

	// projectListCmd 参数
	projectListCmd.Flags().String("keyword", "", "搜索关键词")
	projectListCmd.Flags().String("team-ids", "", "团队 ID 列表（逗号分隔）")
	projectListCmd.Flags().String("portfolio-ids", "", "作品集 ID 列表（逗号分隔）")
	projectListCmd.Flags().Int("scope", 0, "范围筛选（0=所有可见, 1=我加入的）")
	projectListCmd.Flags().Int("page", 1, "页码")
	projectListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
}
```

- [ ] **Step 3: 运行测试验证编译**

Run: `go build ./cmd`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add cmd/project.go
git commit -m "feat(project): 添加 projectCmd 命令组和 projectListCmd"
```

---

### Task 6: Cmd Layer - projectCreateCmd

**Files:**
- Modify: `cmd/project.go`

- [ ] **Step 1: 实现 projectCreateCmd**

在 `cmd/project.go` 文件中添加 projectCreateCmd 定义：

```go
// projectCreateCmd 创建专案
var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建专案",
	Long: `创建新专案。

示例：
  cbi project create --team-id 1 --name "品牌投放"
  cbi project create --team-id 1 --name "新品推广" --privacy 2 --description "新品上市推广素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		teamId, _ := cmd.Flags().GetInt64("team-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if teamId == 0 {
			return cliErr.NewCLIError("MISSING_TEAM_ID", "必须指定 --team-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		privacy, _ := cmd.Flags().GetInt("privacy")
		description, _ := cmd.Flags().GetString("description")
		templateId, _ := cmd.Flags().GetInt64("template-id")
		deadlineStart, _ := cmd.Flags().GetString("deadline-start")
		deadlineEnd, _ := cmd.Flags().GetString("deadline-end")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateProject(ctx, &client.ProjectCreateRequest{
			TeamId:        teamId,
			Name:          name,
			Privacy:       privacy,
			Description:   description,
			TemplateId:    templateId,
			DeadlineStart: deadlineStart,
			DeadlineEnd:   deadlineEnd,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 专案创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  专案 ID: %d\n", result.ProjectId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}
```

- [ ] **Step 2: 在 init() 中添加 projectCreateCmd 和参数**

在 `init()` 函数中添加：

```go
	projectCmd.AddCommand(projectCreateCmd)

	// projectCreateCmd 参数
	projectCreateCmd.Flags().Int64("team-id", 0, "团队 ID（必填）")
	projectCreateCmd.Flags().String("name", "", "专案名称（必填）")
	projectCreateCmd.Flags().Int("privacy", 1, "隐私设置（1=公开, 2=私有）")
	projectCreateCmd.Flags().String("description", "", "专案描述")
	projectCreateCmd.Flags().Int64("template-id", 0, "模板 ID")
	projectCreateCmd.Flags().String("deadline-start", "", "截止日期开始（YYYY-MM-DD）")
	projectCreateCmd.Flags().String("deadline-end", "", "截止日期结束（YYYY-MM-DD）")
	projectCreateCmd.MarkFlagRequired("team-id")
	projectCreateCmd.MarkFlagRequired("name")
```

- [ ] **Step 3: 运行测试验证编译**

Run: `go build ./cmd`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add cmd/project.go
git commit -m "feat(project): 添加 projectCreateCmd"
```

---

### Task 7: Cmd Layer - projectScriptListCmd

**Files:**
- Modify: `cmd/project.go`

- [ ] **Step 1: 实现 projectScriptListCmd**

在 `cmd/project.go` 文件中添加 projectScriptListCmd 定义：

```go
// projectScriptListCmd 脚本列表
var projectScriptListCmd = &cobra.Command{
	Use:   "script-list",
	Short: "列出专案脚本",
	Long: `获取专案的脚本列表。

示例：
  cbi project script-list --project-id 1
  cbi project script-list --project-id 1 --keyword "广告" --state 1
  cbi project script-list --project-id 1 --page 2 --pageSize 30`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		state, _ := cmd.Flags().GetInt("state")
		parentId, _ := cmd.Flags().GetInt64("parent-id")
		isArchived, _ := cmd.Flags().GetInt("is-archived")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListScripts(ctx, &client.ScriptListRequest{
			ProjectId:  projectId,
			Page:       page,
			PageSize:   pageSize,
			Keyword:    keyword,
			State:      state,
			ParentId:   parentId,
			IsArchived: isArchived,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printScriptListTable(cmd, result)
			return nil
		}
	},
}

// printScriptListTable 表格输出脚本列表
func printScriptListTable(cmd *cobra.Command, result *client.ScriptListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Scripts) == 0 {
		fmt.Fprintln(w, "无脚本")
		return
	}

	output.PrintTable(w, []string{"ID", "名称", "状态", "编剧", "设计师", "截止日期"},
		func() [][]string {
			rows := [][]string{}
			for _, s := range result.Scripts {
				writer := "-"
				if s.AssignedWriter != nil {
					writer = s.AssignedWriter.Name
				}
				designer := "-"
				if s.AssignedDesigner != nil {
					designer = s.AssignedDesigner.Name
				}
				rows = append(rows, []string{
					strconv.FormatInt(s.ID, 10),
					s.Name,
					scriptStateName(s.State),
					writer,
					designer,
					formatDate(s.DueDate),
				})
			}
			return rows
		}())
}

// scriptStateName 获取脚本状态名称
func scriptStateName(state int) string {
	switch state {
	case 1:
		return "待处理"
	case 2:
		return "进行中"
	case 3:
		return "已完成"
	case 4:
		return "已归档"
	default:
		return fmt.Sprintf("未知(%d)", state)
	}
}
```

- [ ] **Step 2: 在 init() 中添加 projectScriptListCmd 和参数**

在 `init()` 函数中添加：

```go
	projectCmd.AddCommand(projectScriptListCmd)

	// projectScriptListCmd 参数
	projectScriptListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectScriptListCmd.Flags().String("keyword", "", "搜索关键词")
	projectScriptListCmd.Flags().Int("state", 0, "任务状态筛选")
	projectScriptListCmd.Flags().Int64("parent-id", 0, "父任务筛选")
	projectScriptListCmd.Flags().Int("is-archived", 0, "档案筛选（0=不过滤, 1=档案, 2=非档案）")
	projectScriptListCmd.Flags().Int("page", 1, "页码")
	projectScriptListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
	projectScriptListCmd.MarkFlagRequired("project-id")
```

- [ ] **Step 3: 运行测试验证编译**

Run: `go build ./cmd`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add cmd/project.go
git commit -m "feat(project): 添加 projectScriptListCmd"
```

---

### Task 8: Cmd Layer - projectMaterialListCmd

**Files:**
- Modify: `cmd/project.go`

- [ ] **Step 1: 实现 projectMaterialListCmd**

在 `cmd/project.go` 文件中添加 projectMaterialListCmd 定义：

```go
// projectMaterialListCmd 素材列表
var projectMaterialListCmd = &cobra.Command{
	Use:   "material-list",
	Short: "列出专案素材",
	Long: `获取专案的素材列表。

示例：
  cbi project material-list --project-id 1
  cbi project material-list --project-id 1 --keyword "视频"
  cbi project material-list --project-id 1 --page 2 --pageSize 30`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListMaterials(ctx, &client.MaterialListRequest{
			ProjectId: projectId,
			Page:      page,
			PageSize:  pageSize,
			Keyword:   keyword,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printMaterialListTable(cmd, result)
			return nil
		}
	},
}

// printMaterialListTable 表格输出素材列表
func printMaterialListTable(cmd *cobra.Command, result *client.MaterialListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Materials) == 0 {
		fmt.Fprintln(w, "无素材")
		return
	}

	output.PrintTable(w, []string{"ID", "名称", "类型", "格式", "时长", "创建者"},
		func() [][]string {
			rows := [][]string{}
			for _, m := range result.Materials {
				creator := "-"
				if m.Creator != nil {
					creator = m.Creator.Name
				}
				rows = append(rows, []string{
					strconv.FormatInt(m.ID, 10),
					m.Name,
					fileTypeName(m.FileType),
					m.Format,
					m.Duration,
					creator,
				})
			}
			return rows
		}())
}

// fileTypeName 获取文件类型名称
func fileTypeName(fileType int) string {
	switch fileType {
	case 1:
		return "视频"
	case 2:
		return "图片"
	default:
		return fmt.Sprintf("未知(%d)", fileType)
	}
}
```

- [ ] **Step 2: 在 init() 中添加 projectMaterialListCmd 和参数**

在 `init()` 函数中添加：

```go
	projectCmd.AddCommand(projectMaterialListCmd)

	// projectMaterialListCmd 参数
	projectMaterialListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialListCmd.Flags().String("keyword", "", "搜索关键词")
	projectMaterialListCmd.Flags().Int("page", 1, "页码")
	projectMaterialListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
	projectMaterialListCmd.MarkFlagRequired("project-id")
```

- [ ] **Step 3: 运行测试验证编译**

Run: `go build ./cmd`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add cmd/project.go
git commit -m "feat(project): 添加 projectMaterialListCmd"
```

---

### Task 9: 集成测试

**Files:**
- None

- [ ] **Step 1: 构建完整项目**

Run: `make build`
Expected: 编译成功，生成 `bin/cbi`

- [ ] **Step 2: 测试命令帮助**

Run: `./bin/cbi project --help`
Expected: 显示 project 命令组帮助

Run: `./bin/cbi project list --help`
Expected: 显示 project list 参数帮助

- [ ] **Step 3: Commit（如有修复）**

如有修复，提交：

```bash
git add cmd/project.go internal/client/project.go
git commit -m "fix(project): 修复集成测试问题"
```

---

### Task 10: 更新 README.md

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 在命令概览中添加 project 命令组**

在 README.md 的命令概览部分添加：

```markdown
├── project                   # 专案管理
│   ├── list                  # 专案列表
│   ├── create                # 创建专案
│   ├── script-list           # 脚本列表
│   └── material-list         # 素材列表
```

- [ ] **Step 2: 添加专案模块文档章节**

在 README.md 中添加新章节：

```markdown
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

**customFields 说明：**

脚本和素材都支持自定义字段（customFields），以 `map<string, string>` 形式返回，key 为字段名称（fieldName），value 为 JSON 字符串。字段定义（fields）随列表返回，包含 classify 字段标识分类：

| Classify | 说明 |
|----------|------|
| 1 | 固定字段 |
| 2 | 固有字段 |
| 3 | 自定义字段 |
```

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: 添加 project 模块文档"
```

---

## Self-Review

**1. Spec coverage:**
- 专案列表: Task 1, 5 ✓
- 创建专案: Task 2, 6 ✓
- 脚本列表: Task 3, 7 ✓
- 素材列表: Task 4, 8 ✓
- customFields/fields: Task 3, 4 ✓
- README.md: Task 10 ✓

**2. Placeholder scan:** 无 TBD/TODO ✓

**3. Type consistency:**
- ProjectClient 方法签名一致 ✓
- Script/Material 结构体字段一致 ✓
- FieldDef 结构体共用 ✓