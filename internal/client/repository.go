package client

import (
	"context"
	"fmt"
	"os"

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

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("FILE_CREATE_ERROR",
			fmt.Sprintf("文件创建失败 (%d)", codeVal), message)
	}

	fileData := result.Get("data.file")
	return &FileCreateInfo{
		ID:   fileData.Get("id").Int(),
		Name: fileData.Get("name").String(),
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