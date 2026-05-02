# 专案集模块设计

## 概述

新增 `cbi portfolio` 命令组，实现专案集列表、专案集下的专案列表两个功能。

## 命令结构

```
cbi portfolio
├── list              # 获取专案集列表
└── project-list      # 获取专案集下的专案列表
```

## API 接口

### 1. 获取专案集列表
- **端点**: POST `/openapi/v1/portfolio/list`
- **请求参数**: page, pageSize, keyword, scope (0=所有可见, 1=我加入的)
- **响应**: portfolios[], total, page, pageSize

### 2. 获取专案集下的专案列表
- **端点**: POST `/openapi/v1/portfolio/project/list`
- **请求参数**: portfolioId(必填), page, pageSize, keyword
- **响应**: projects[], total, page, pageSize

## 数据结构定义

### internal/client/portfolio.go

```go
// PortfolioClient 专案集 API 客户端
type PortfolioClient struct {
	client *resty.Client
}

// Portfolio 专案集
type Portfolio struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	Color        string       `json:"color"`
	Creator      *CreatorInfo `json:"creator,omitempty"`
	Privacy      int          `json:"privacy"`       // 1=公开, 2=私有
	ProjectCount int64        `json:"projectCount"`  // 包含的专案数量
	CreatedAt    string       `json:"createdAt"`
}

// PortfolioListRequest 专案集列表请求
type PortfolioListRequest struct {
	Page     int
	PageSize int
	Keyword  string
	Scope    int // 0=所有可见, 1=我加入的
}

// PortfolioListResult 专案集列表结果
type PortfolioListResult struct {
	Portfolios []Portfolio `json:"portfolios"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
}

// PortfolioProject 专案集下的专案
type PortfolioProject struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	DeadlineStart string     `json:"deadlineStart"`
	DeadlineEnd   string     `json:"deadlineEnd"`
	Team          *TeamInfo  `json:"team,omitempty"`
	Status        int        `json:"status"`
	Owner         *OwnerInfo `json:"owner,omitempty"`
}

// TeamInfo 团队信息
type TeamInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// OwnerInfo 所属人信息
type OwnerInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// PortfolioProjectListRequest 专案集下的专案列表请求
type PortfolioProjectListRequest struct {
	PortfolioId int64
	Page        int
	PageSize    int
	Keyword     string
}

// PortfolioProjectListResult 专案集下的专案列表结果
type PortfolioProjectListResult struct {
	Projects []PortfolioProject `json:"projects"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}
```

## client 层方法

```go
// ListPortfolios 获取专案集列表
func (c *PortfolioClient) ListPortfolios(ctx context.Context, req *PortfolioListRequest) (*PortfolioListResult, error)

// ListPortfolioProjects 获取专案集下的专案列表
func (c *PortfolioClient) ListPortfolioProjects(ctx context.Context, req *PortfolioProjectListRequest) (*PortfolioProjectListResult, error)
```

## cmd 层命令

### cmd/portfolio.go

新增文件，定义 `portfolioCmd` 及其子命令：

- `portfolioListCmd` — 专案集列表，参数：--keyword, --scope, --page, --pageSize
- `portfolioProjectListCmd` — 专案集下的专案列表，参数：--portfolio-id(必填), --keyword, --page, --pageSize

## 输出格式

- **table**: 表格输出（默认）
- **json**: JSON 格式
- **markdown**: Markdown 格式

### 表格输出示例

**专案集列表**:
```
共 10 条，第 1/1 页

ID    名称          颜色      可见性    专案数    创建者    创建时间
1     品牌A投放     #FF0000   公开      10        张三      2026-5-3
2     品牌B推广     #00FF00   私有      5         李四      2026-5-3
```

示例：
```
cbi portfolio list
cbi portfolio list --keyword "品牌"
cbi portfolio list --scope 1 --page 1 --pageSize 20
```

**专案集下的专案列表**:
```
共 30 条，第 1/1 页

ID    名称          状态        团队      所属人    截止日期
1     专案A         正常进行    团队1     王五      2026-5-1 ~ 2026-5-31
```

## 状态映射

| 值 | 说明 |
|---|------|
| 1 | OnTrack（正常进行） |
| 2 | AtRisk（有风险） |
| 3 | OffTrack（偏离轨道） |
| 4 | OnHold（暂停） |
| 5 | Complete（完成） |
| 6 | NoUpdates（无更新） |

## 可见性映射

| 值 | 说明 |
|---|------|
| 1 | 公开 |
| 2 | 私有 |

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/client/portfolio.go` | 新增：PortfolioClient 和数据结构、API 方法 |
| `cmd/portfolio.go` | 新增：portfolio 命令组和 Cobra 子命令 |

## 认证

所有接口需要 `user-access-token` header，沿用现有认证机制。

## 错误处理

沿用现有模式：
- 500 错误检查 token 过期
- 业务错误返回结构化 CLI 错误
- 参数校验返回 MISSING_* 错误