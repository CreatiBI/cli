package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// RepositoryClient 素材库 API 客户端
type RepositoryClient struct {
	client *resty.Client
}

// NewRepositoryClient 创建素材库客户端
func NewRepositoryClient() *RepositoryClient {
	baseURL := config.GetBaseURL()
	return &RepositoryClient{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(60 * 5 * 1000000000), // 5 分钟，用于大文件上传
	}
}

// Repository 素材库信息
type Repository struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Desc      string `json:"desc"`
	IsDefault bool   `json:"isDefault"`
	Perm      string `json:"perm"`
}

// Folder 文件夹信息
type Folder struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Statistic *Statistic `json:"statistic,omitempty"`
}

// Statistic 统计信息
type Statistic struct {
	FileCount int64 `json:"fileCount"`
}

// FileCheckResult 文件查重结果
type FileCheckResult struct {
	Existed bool `json:"existed"`
}

// FileCreateInfo 文件创建结果
type FileCreateInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FileDetail 文件详情
type FileDetail struct {
	ID             int64          `json:"id"`
	RepositoryID   int64          `json:"repositoryId"`
	Name           string         `json:"name"`
	FileType       int            `json:"fileType"`
	Format         string         `json:"format"`
	Cover          string         `json:"cover"`
	FileOriginUrl  string         `json:"fileOriginUrl"`
	FileViewUrl    string         `json:"fileViewUrl"`
	Size           string         `json:"size"`
	SizeInByte     int64          `json:"sizeInByte"`
	Duration       string         `json:"duration"`
	Resolution     string         `json:"resolution"`
	Ratio          string         `json:"ratio"`
	FrameRate      string         `json:"frameRate"`
	Hash           string         `json:"hash"`
	Score          int            `json:"score"`
	Notes          string         `json:"notes"`
	FileSourceUrl  string         `json:"fileSourceUrl"`
	SourcePlatform string         `json:"sourcePlatform"`
	Statistic      string         `json:"statistic"`
	Extra          interface{}    `json:"extra"`
	CreatedAt      int64          `json:"createdAt"`
	UpdatedAt      int64          `json:"updatedAt"`
	Products       []Product      `json:"products"`
	Tags           []Tag          `json:"tags"`
	Folders        []FolderInfo   `json:"folders"`
	Creator        *CreatorInfo   `json:"creator"`
	Signals        []Signal       `json:"signals"`
}

// Product 关联产品
type Product struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Img  string `json:"img"`
	URL  string `json:"url"`
}

// Tag 标签
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// FolderInfo 文件夹信息（简化版）
type FolderInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreatorInfo 创建者信息
type CreatorInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

// Signal 视频理解信号
type Signal struct {
	SignalID      string   `json:"signalId"`
	SignalName    string   `json:"signalName"`
	SignalContent string   `json:"signalContent"`
	SignalTags    []string `json:"signalTags"`
}

// ListRepositories 获取权限范围内素材库列表
func (c *RepositoryClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody("{}").
		Post("/openapi/v1/repository/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("REPOSITORY_LIST_ERROR",
			fmt.Sprintf("获取素材库列表失败 (%d)", codeVal), message)
	}

	repositories := []Repository{}
	result.Get("data.repositories").ForEach(func(_, value gjson.Result) bool {
		repositories = append(repositories, Repository{
			ID:        value.Get("id").Int(),
			Name:      value.Get("name").String(),
			Desc:      value.Get("desc").String(),
			IsDefault: value.Get("isDefault").Bool(),
			Perm:      value.Get("perm").String(),
		})
		return true
	})

	return repositories, nil
}

// ListFolders 获取素材库文件夹列表
func (c *RepositoryClient) ListFolders(ctx context.Context, repositoryID int64, parentFolderID int64, withStatistic bool) ([]Folder, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"repositoryId":   repositoryID,
			"parentFolderId": parentFolderID,
			"withStatistic":  withStatistic,
		}).
		Post("/openapi/v1/repository/folder/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FOLDER_LIST_ERROR",
			fmt.Sprintf("获取文件夹列表失败 (%d)", codeVal), message)
	}

	folders := []Folder{}
	result.Get("data.folders").ForEach(func(_, value gjson.Result) bool {
		folder := Folder{
			ID:   value.Get("id").Int(),
			Name: value.Get("name").String(),
		}
		if withStatistic {
			folder.Statistic = &Statistic{
				FileCount: value.Get("statistic.fileCount").Int(),
			}
		}
		folders = append(folders, folder)
		return true
	})

	return folders, nil
}

// CheckFile 检查文件是否已存在（通过 MD5）
func (c *RepositoryClient) CheckFile(ctx context.Context, repositoryID int64, fileMD5 string) (*FileCheckResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"repositoryId": repositoryID,
			"fileMd5":      fileMD5,
		}).
		Post("/openapi/v1/repository/file/check")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FILE_CHECK_ERROR",
			fmt.Sprintf("文件查重失败 (%d)", codeVal), message)
	}

	return &FileCheckResult{
		Existed: result.Get("data.existed").Bool(),
	}, nil
}

// CreateFile 在素材库创建文件（上传）
// 文件大小限制：< 100MB
func (c *RepositoryClient) CreateFile(ctx context.Context, req *CreateFileRequest) (*FileCreateInfo, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	// 检查文件大小
	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("FILE_NOT_FOUND", "文件不存在", req.FilePath)
	}

	// 100MB = 100 * 1024 * 1024 bytes
	if fileInfo.Size() > 100*1024*1024 {
		return nil, cliErr.NewCLIError("FILE_TOO_LARGE", "文件大小超过 100MB 限制")
	}

	// 打开文件
	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("FILE_OPEN_ERROR", "无法打开文件", err.Error())
	}
	defer file.Close()

	// 构建请求
	r := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetFile("file", req.FilePath).
		SetFormData(map[string]string{
			"repositoryId": fmt.Sprintf("%d", req.RepositoryID),
		})

	// 可选参数
	if req.FolderID > 0 {
		r.SetFormData(map[string]string{"folderId": fmt.Sprintf("%d", req.FolderID)})
	}
	if req.Name != "" {
		r.SetFormData(map[string]string{"name": req.Name})
	}
	if req.Note != "" {
		r.SetFormData(map[string]string{"note": req.Note})
	}
	if req.Rating > 0 {
		r.SetFormData(map[string]string{"rating": fmt.Sprintf("%d", req.Rating)})
	}
	if req.SourceURL != "" {
		r.SetFormData(map[string]string{"sourceUrl": req.SourceURL})
	}
	if req.Tags != "" {
		r.SetFormData(map[string]string{"tags": req.Tags})
	}

	resp, err := r.Post("/openapi/v1/repository/file/create")
	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 检查 HTTP 状态码（nginx 可能返回 413 等错误）
	if resp.StatusCode() == 413 {
		return nil, cliErr.NewCLIError("FILE_TOO_LARGE",
			"文件大小超过服务器限制（nginx 413），请尝试上传较小的文件或联系管理员调整服务器配置")
	}
	if resp.StatusCode() >= 400 {
		// 尝试解析响应体中的错误信息
		result := gjson.ParseBytes(resp.Body())
		message := result.Get("message").String()
		if message == "" {
			// 如果无法解析为 JSON，检查是否是 HTML 错误页面
			bodyStr := string(resp.Body())
			if strings.Contains(bodyStr, "<html>") || strings.Contains(bodyStr, "<title>") {
				// 提取标题中的错误信息
				if strings.Contains(bodyStr, "413 Request Entity Too Large") {
					return nil, cliErr.NewCLIError("FILE_TOO_LARGE",
						"文件大小超过服务器限制（nginx 413），请尝试上传较小的文件或联系管理员调整服务器配置")
				}
				return nil, cliErr.NewCLIErrorWithDetail("HTTP_ERROR",
					fmt.Sprintf("HTTP %d 错误", resp.StatusCode()), bodyStr)
			}
			return nil, cliErr.NewCLIError("HTTP_ERROR",
				fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status()))
		}
		return nil, cliErr.NewCLIErrorWithDetail("HTTP_ERROR",
			fmt.Sprintf("HTTP %d", resp.StatusCode()), message)
	}

	// verbose 模式打印完整响应
	if os.Getenv("CBI_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "API Response: %s\n", string(resp.Body()))
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FILE_CREATE_ERROR",
			fmt.Sprintf("文件创建失败 (%d)", codeVal), message)
	}

	fileData := result.Get("data.file")
	fileID := fileData.Get("id").Int()
	fileName := fileData.Get("name").String()

	// 验证返回的文件信息是否有效
	if fileID == 0 {
		// 检查是否有错误信息
		message := result.Get("message").String()
		if message != "" && message != "success" {
			return nil, cliErr.NewCLIErrorWithDetail("FILE_CREATE_ERROR",
				"文件创建失败", message)
		}
		// 打印完整响应便于调试
		return nil, cliErr.NewCLIErrorWithDetail("FILE_CREATE_ERROR",
			"服务器返回无效的文件信息", string(resp.Body()))
	}

	return &FileCreateInfo{
		ID:   fileID,
		Name: fileName,
	}, nil
}

// CreateFileRequest 文件创建请求
type CreateFileRequest struct {
	RepositoryID int64
	FolderID     int64
	FilePath     string
	Name         string
	Note         string
	Rating       int    // 0-5
	SourceURL    string
	Tags         string // 逗号分隔
}

// FileListItem 文件列表项
type FileListItem struct {
	ID           int64        `json:"id"`
	RepositoryID int64        `json:"repositoryId"`
	Name         string       `json:"name"`
	FileType     int          `json:"fileType"`
	Cover        string       `json:"cover"`
	FileViewUrl  string       `json:"fileViewUrl"`
	Score        int          `json:"score"`
	CreatedAt    int64        `json:"createdAt"`
	UpdatedAt    int64        `json:"updatedAt"`
	Tags         []Tag        `json:"tags"`
	Folders      []FolderInfo `json:"folders"`
}

// FileListResult 文件列表结果
type FileListResult struct {
	Files    []FileListItem `json:"files"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
PageSize int            `json:"pageSize"`
}

// ListFilesRequest 文件列表请求参数
type ListFilesRequest struct {
RepositoryID int64
	FolderID    int64
TagID        int64
Keyword      string
	Page        int
PageSize     int
}

// ListFiles 获取素材库文件列表
func (c *RepositoryClient) ListFiles(ctx context.Context, req *ListFilesRequest) (*FileListResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	// 构建请求体
	body := map[string]interface{}{
		"repositoryId": req.RepositoryID,
	}

	// 添加可选参数（oneof 筛选模式）
	if req.FolderID > 0 {
		body["folderId"] = req.FolderID
	}
	if req.TagID > 0 {
		body["tagId"] = req.TagID
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.Page > 0 {
		body["page"] = req.Page
	}
	if req.PageSize > 0 {
		body["pageSize"] = req.PageSize
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/repository/file/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FILE_LIST_ERROR",
			fmt.Sprintf("获取文件列表失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	filesResult := &FileListResult{
		Total:    data.Get("total").Int(),
		Page:     int(data.Get("page").Int()),
		PageSize: int(data.Get("pageSize").Int()),
	}

	// 解析 files 列表
	data.Get("files").ForEach(func(_, value gjson.Result) bool {
		file := FileListItem{
			ID:           value.Get("id").Int(),
			RepositoryID: value.Get("repositoryId").Int(),
			Name:         value.Get("name").String(),
			FileType:     int(value.Get("fileType").Int()),
			Cover:        value.Get("cover").String(),
			FileViewUrl:  value.Get("fileViewUrl").String(),
			Score:        int(value.Get("score").Int()),
			CreatedAt:    value.Get("createdAt").Int(),
			UpdatedAt:    value.Get("updatedAt").Int(),
		}

		// 解析 tags
		value.Get("tags").ForEach(func(_, tag gjson.Result) bool {
			file.Tags = append(file.Tags, Tag{
				ID:    tag.Get("id").Int(),
				Name:  tag.Get("name").String(),
				Color: tag.Get("color").String(),
			})
			return true
		})

		// 解析 folders
		value.Get("folders").ForEach(func(_, folder gjson.Result) bool {
			file.Folders = append(file.Folders, FolderInfo{
				ID:   folder.Get("id").Int(),
				Name: folder.Get("name").String(),
			})
			return true
		})

		filesResult.Files = append(filesResult.Files, file)
		return true
	})

	return filesResult, nil
}

// GetFileDetail 获取素材文件详情
func (c *RepositoryClient) GetFileDetail(ctx context.Context, fileID int64) (*FileDetail, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"fileId": fileID,
		}).
		Post("/openapi/v1/repository/file/detail")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FILE_DETAIL_ERROR",
			fmt.Sprintf("获取文件详情失败 (%d)", codeVal), message)
	}

	fileData := result.Get("data.file")
	detail := &FileDetail{
		ID:             fileData.Get("id").Int(),
		RepositoryID:   fileData.Get("repositoryId").Int(),
		Name:           fileData.Get("name").String(),
		FileType:       int(fileData.Get("fileType").Int()),
		Format:         fileData.Get("format").String(),
		Cover:          fileData.Get("cover").String(),
		FileOriginUrl:  fileData.Get("fileOriginUrl").String(),
		FileViewUrl:    fileData.Get("fileViewUrl").String(),
		Size:           fileData.Get("size").String(),
		SizeInByte:     fileData.Get("sizeInByte").Int(),
		Duration:       fileData.Get("duration").String(),
		Resolution:     fileData.Get("resolution").String(),
		Ratio:          fileData.Get("ratio").String(),
		FrameRate:      fileData.Get("frameRate").String(),
		Hash:           fileData.Get("hash").String(),
		Score:          int(fileData.Get("score").Int()),
		Notes:          fileData.Get("notes").String(),
		FileSourceUrl:  fileData.Get("fileSourceUrl").String(),
		SourcePlatform: fileData.Get("sourcePlatform").String(),
		Statistic:      fileData.Get("statistic").String(),
		CreatedAt:      fileData.Get("createdAt").Int(),
		UpdatedAt:      fileData.Get("updatedAt").Int(),
	}

	// 解析 products
	fileData.Get("products").ForEach(func(_, value gjson.Result) bool {
		detail.Products = append(detail.Products, Product{
			ID:   value.Get("id").Int(),
			Name: value.Get("name").String(),
			Img:  value.Get("img").String(),
			URL:  value.Get("url").String(),
		})
		return true
	})

	// 解析 tags
	fileData.Get("tags").ForEach(func(_, value gjson.Result) bool {
		detail.Tags = append(detail.Tags, Tag{
			ID:    value.Get("id").Int(),
			Name:  value.Get("name").String(),
			Color: value.Get("color").String(),
		})
		return true
	})

	// 解析 folders
	fileData.Get("folders").ForEach(func(_, value gjson.Result) bool {
		detail.Folders = append(detail.Folders, FolderInfo{
			ID:   value.Get("id").Int(),
			Name: value.Get("name").String(),
		})
		return true
	})

	// 解析 creator
	creatorData := fileData.Get("creator")
	if creatorData.Exists() {
		detail.Creator = &CreatorInfo{
			ID:     creatorData.Get("id").Int(),
			Name:   creatorData.Get("name").String(),
			Email:  creatorData.Get("email").String(),
			Avatar: creatorData.Get("avatar").String(),
		}
	}

	// 解析 signals
	fileData.Get("signals").ForEach(func(_, value gjson.Result) bool {
		signal := Signal{
			SignalID:      value.Get("signalId").String(),
			SignalName:    value.Get("signalName").String(),
			SignalContent: value.Get("signalContent").String(),
		}
		// 解析 signalTags
		value.Get("signalTags").ForEach(func(_, tag gjson.Result) bool {
			signal.SignalTags = append(signal.SignalTags, tag.String())
			return true
		})
		detail.Signals = append(detail.Signals, signal)
		return true
	})

	return detail, nil
}