package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// PortfolioClient 作品集 API 客户端
type PortfolioClient struct {
	client *resty.Client
}

// NewPortfolioClient 创建作品集客户端
func NewPortfolioClient() *PortfolioClient {
	baseURL := config.GetBaseURL()
	return &PortfolioClient{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(30 * 1000000000), // 30 秒
	}
}

// Portfolio 作品集
type Portfolio struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	Color        string       `json:"color"`
	Creator      *CreatorInfo `json:"creator,omitempty"`
	Privacy      int          `json:"privacy"`      // 1=公开, 2=私有
	ProjectCount int64        `json:"projectCount"` // 包含的专案数量
	CreatedAt    string       `json:"createdAt"`
}

// PortfolioListRequest 作品集列表请求
type PortfolioListRequest struct {
	Page     int
	PageSize int
	Keyword  string
	Scope    int // 0=所有可见, 1=我加入的
}

// PortfolioListResult 作品集列表结果
type PortfolioListResult struct {
	Portfolios []Portfolio `json:"portfolios"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
}

// PortfolioProject 作品集内的专案
type PortfolioProject struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	DeadlineStart string      `json:"deadlineStart"`
	DeadlineEnd   string      `json:"deadlineEnd"`
	Team         *TeamInfo    `json:"team,omitempty"`
	Status       int          `json:"status"`
	Owner        *OwnerInfo   `json:"owner,omitempty"`
}

// TeamInfo 团队信息
type TeamInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// OwnerInfo 专案负责人信息
type OwnerInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// PortfolioProjectListRequest 作品集专案列表请求
type PortfolioProjectListRequest struct {
	PortfolioId int64
	Page        int
	PageSize    int
	Keyword     string
}

// PortfolioProjectListResult 作品集专案列表结果
type PortfolioProjectListResult struct {
	Projects []PortfolioProject `json:"projects"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// ListPortfolios 获取作品集列表
func (c *PortfolioClient) ListPortfolios(ctx context.Context, req *PortfolioListRequest) (*PortfolioListResult, error) {
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
	if req.Scope > 0 {
		body["scope"] = req.Scope
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/portfolio/list")

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
		return nil, cliErr.NewCLIErrorWithDetail("PORTFOLIO_LIST_ERROR",
			fmt.Sprintf("获取作品集列表失败 (%d)", codeVal), message)
	}

	portfolios := []Portfolio{}
	result.Get("data.portfolios").ForEach(func(_, value gjson.Result) bool {
		portfolio := Portfolio{
			ID:           value.Get("id").Int(),
			Name:         value.Get("name").String(),
			Color:        value.Get("color").String(),
			Privacy:      int(value.Get("privacy").Int()),
			ProjectCount: value.Get("projectCount").Int(),
			CreatedAt:    value.Get("createdAt").String(),
		}
		if creator := value.Get("creator"); creator.Exists() {
			portfolio.Creator = &CreatorInfo{
				ID:     creator.Get("id").Int(),
				Name:   creator.Get("name").String(),
				Email:  creator.Get("email").String(),
				Avatar: creator.Get("avatar").String(),
			}
		}
		portfolios = append(portfolios, portfolio)
		return true
	})

	return &PortfolioListResult{
		Portfolios: portfolios,
		Total:      result.Get("data.total").Int(),
		Page:       int(result.Get("data.page").Int()),
		PageSize:   int(result.Get("data.pageSize").Int()),
	}, nil
}

// ListPortfolioProjects 获取作品集专案列表
func (c *PortfolioClient) ListPortfolioProjects(ctx context.Context, req *PortfolioProjectListRequest) (*PortfolioProjectListResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{
		"portfolioId": req.PortfolioId,
		"page":        req.Page,
		"pageSize":    req.PageSize,
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/portfolio/project/list")

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
		return nil, cliErr.NewCLIErrorWithDetail("PORTFOLIO_PROJECT_LIST_ERROR",
			fmt.Sprintf("获取作品集专案列表失败 (%d)", codeVal), message)
	}

	projects := []PortfolioProject{}
	result.Get("data.projects").ForEach(func(_, value gjson.Result) bool {
		project := PortfolioProject{
			ID:            value.Get("id").Int(),
			Name:          value.Get("name").String(),
			DeadlineStart: value.Get("deadlineStart").String(),
			DeadlineEnd:   value.Get("deadlineEnd").String(),
			Status:        int(value.Get("status").Int()),
		}
		// 解析 team
		if team := value.Get("team"); team.Exists() {
			project.Team = parseTeamInfo(team)
		}
		// 解析 owner
		if owner := value.Get("owner"); owner.Exists() {
			project.Owner = parseOwnerInfo(owner)
		}
		projects = append(projects, project)
		return true
	})

	return &PortfolioProjectListResult{
		Projects: projects,
		Total:    result.Get("data.total").Int(),
		Page:     int(result.Get("data.page").Int()),
		PageSize: int(result.Get("data.pageSize").Int()),
	}, nil
}

// parseTeamInfo 解析 TeamInfo
func parseTeamInfo(value gjson.Result) *TeamInfo {
	return &TeamInfo{
		ID:   value.Get("id").Int(),
		Name: value.Get("name").String(),
	}
}

// parseOwnerInfo 解析 OwnerInfo
func parseOwnerInfo(value gjson.Result) *OwnerInfo {
	return &OwnerInfo{
		ID:     value.Get("id").Int(),
		Name:   value.Get("name").String(),
		Avatar: value.Get("avatar").String(),
	}
}