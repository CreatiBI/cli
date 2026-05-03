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
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	State            int               `json:"state"`
	Creator          *CreatorInfo      `json:"creator,omitempty"`
	AssignedWriter   *CreatorInfo      `json:"assignedWriter,omitempty"`
	AssignedDesigner *CreatorInfo      `json:"assignedDesigner,omitempty"`
	DueDate          string            `json:"dueDate"`
	CreatedAt        string            `json:"createdAt"`
	ParentId         int64             `json:"parentId"`
	CurrentVersionNo int               `json:"currentVersionNo"`
	TableIdValue     int64             `json:"tableIdValue"`
	AiGenerate       int               `json:"aiGenerate"`
	CustomFields     map[string]string `json:"customFields"` // 自定义字段值，key=fieldName，value=JSON字符串
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
	ID           int64             `json:"id"`
	Name         string            `json:"name"`
	FileType     int               `json:"fileType"`    // 1=视频, 2=图片
	Format       string            `json:"format"`
	Duration     string            `json:"duration"`
	Resolution   string            `json:"resolution"`
	Cover        string            `json:"cover"`
	PlayUrl      string            `json:"playUrl"`
	Ratio        float64           `json:"ratio"`
	FileSize     int64             `json:"fileSize"`
	Rating       int               `json:"rating"`
	Status       int               `json:"status"`
	ScriptId     int64             `json:"scriptId"`
	Creator      *CreatorInfo      `json:"creator,omitempty"`
	Producer     *CreatorInfo      `json:"producer,omitempty"`
	Tags         []Tag             `json:"tags"`
	CreatedAt    string            `json:"createdAt"`
	CustomFields map[string]string `json:"customFields"` // 自定义字段值，key=fieldName，value=JSON字符串
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
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
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

	// 处理 500 错误（可能 token 过期）
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
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
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
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
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
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

// CreateScriptTaskRequest 创建脚本任务请求
type CreateScriptTaskRequest struct {
	ProjectId    int64  // 必填
	Name         string // 必填
	ParentId     int64  // 可选，父任务ID
	SourceObject string // 可选，来源对象
}

// CreateScriptTaskResult 创建脚本任务结果
type CreateScriptTaskResult struct {
	ScriptId int64  `json:"scriptId"`
	Name     string `json:"name"`
}

// CreateScriptTask 创建脚本任务
func (c *ProjectClient) CreateScriptTask(ctx context.Context, req *CreateScriptTaskRequest) (*CreateScriptTaskResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId": req.ProjectId,
		"name":      req.Name,
	}
	if req.ParentId > 0 {
		body["parentId"] = req.ParentId
	}
	if req.SourceObject != "" {
		body["sourceObject"] = req.SourceObject
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/script/create")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("SCRIPT_CREATE_ERROR",
			fmt.Sprintf("创建脚本任务失败 (%d)", codeVal), message)
	}

	return &CreateScriptTaskResult{
		ScriptId: result.Get("data.scriptId").Int(),
		Name:     result.Get("data.name").String(),
	}, nil
}

// CreateFissionMaterialFromTaskRequest 从任务创建裂变素材请求
type CreateFissionMaterialFromTaskRequest struct {
	ProjectId int64  // 必填
	ScriptId  int64  // 必填
	Name      string // 必填
}

// CreateFissionMaterialFromTaskResult 从任务创建裂变素材结果
type CreateFissionMaterialFromTaskResult struct {
	MaterialId int64  `json:"materialId"`
	Name       string `json:"name"`
}

// CreateFissionMaterialFromTask 从任务创建裂变素材（父子关系）
func (c *ProjectClient) CreateFissionMaterialFromTask(ctx context.Context, req *CreateFissionMaterialFromTaskRequest) (*CreateFissionMaterialFromTaskResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId": req.ProjectId,
		"scriptId":  req.ScriptId,
		"name":      req.Name,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/material/fission-from-task")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FISSION_MATERIAL_FROM_TASK_ERROR",
			fmt.Sprintf("从任务创建裂变素材失败 (%d)", codeVal), message)
	}

	return &CreateFissionMaterialFromTaskResult{
		MaterialId: result.Get("data.materialId").Int(),
		Name:       result.Get("data.name").String(),
	}, nil
}

// CreateDerivativeMaterialFromTaskRequest 从任务创建衍生素材请求
type CreateDerivativeMaterialFromTaskRequest struct {
	ProjectId int64  // 必填
	ScriptId  int64  // 必填
	Name      string // 必填
}

// CreateDerivativeMaterialFromTaskResult 从任务创建衍生素材结果
type CreateDerivativeMaterialFromTaskResult struct {
	MaterialId int64  `json:"materialId"`
	Name       string `json:"name"`
}

// CreateDerivativeMaterialFromTask 从任务创建衍生素材（同级关系）
func (c *ProjectClient) CreateDerivativeMaterialFromTask(ctx context.Context, req *CreateDerivativeMaterialFromTaskRequest) (*CreateDerivativeMaterialFromTaskResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId": req.ProjectId,
		"scriptId":  req.ScriptId,
		"name":      req.Name,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/material/derivative-from-task")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("DERIVATIVE_MATERIAL_FROM_TASK_ERROR",
			fmt.Sprintf("从任务创建衍生素材失败 (%d)", codeVal), message)
	}

	return &CreateDerivativeMaterialFromTaskResult{
		MaterialId: result.Get("data.materialId").Int(),
		Name:       result.Get("data.name").String(),
	}, nil
}

// CreateFissionMaterialFromMaterialRequest 从素材创建裂变素材请求
type CreateFissionMaterialFromMaterialRequest struct {
	ProjectId  int64  // 必填
	MaterialId int64  // 必填
	Name       string // 必填
}

// CreateFissionMaterialFromMaterialResult 从素材创建裂变素材结果
type CreateFissionMaterialFromMaterialResult struct {
	MaterialId int64  `json:"materialId"`
	Name       string `json:"name"`
}

// CreateFissionMaterialFromMaterial 从素材创建裂变素材（父子关系）
func (c *ProjectClient) CreateFissionMaterialFromMaterial(ctx context.Context, req *CreateFissionMaterialFromMaterialRequest) (*CreateFissionMaterialFromMaterialResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId":  req.ProjectId,
		"materialId": req.MaterialId,
		"name":       req.Name,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/material/fission-from-material")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FISSION_MATERIAL_FROM_MATERIAL_ERROR",
			fmt.Sprintf("从素材创建裂变素材失败 (%d)", codeVal), message)
	}

	return &CreateFissionMaterialFromMaterialResult{
		MaterialId: result.Get("data.materialId").Int(),
		Name:       result.Get("data.name").String(),
	}, nil
}

// CreateDerivativeMaterialFromMaterialRequest 从素材创建衍生素材请求
type CreateDerivativeMaterialFromMaterialRequest struct {
	ProjectId  int64  // 必填
	MaterialId int64  // 必填
	Name       string // 必填
}

// CreateDerivativeMaterialFromMaterialResult 从素材创建衍生素材结果
type CreateDerivativeMaterialFromMaterialResult struct {
	MaterialId int64  `json:"materialId"`
	Name       string `json:"name"`
}

// CreateDerivativeMaterialFromMaterial 从素材创建衍生素材（同级关系）
func (c *ProjectClient) CreateDerivativeMaterialFromMaterial(ctx context.Context, req *CreateDerivativeMaterialFromMaterialRequest) (*CreateDerivativeMaterialFromMaterialResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"projectId":  req.ProjectId,
		"materialId": req.MaterialId,
		"name":       req.Name,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/project/material/derivative-from-material")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误
	if resp.StatusCode() == 500 {
		if handle500Error(resp.Body()) == cliErr.ErrTokenExpired {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIError("SERVER_ERROR", "服务器内部错误，请稍后重试")
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("DERIVATIVE_MATERIAL_FROM_MATERIAL_ERROR",
			fmt.Sprintf("从素材创建衍生素材失败 (%d)", codeVal), message)
	}

	return &CreateDerivativeMaterialFromMaterialResult{
		MaterialId: result.Get("data.materialId").Int(),
		Name:       result.Get("data.name").String(),
	}, nil
}