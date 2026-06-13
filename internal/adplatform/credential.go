package adplatform

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// AdAuthorizationToken 广告授权 Token（从 CreatiBI 后端实时获取，不落盘）
type AdAuthorizationToken struct {
	AccessToken      string // 平台 Access Token
	RefreshToken     string // 平台 Refresh Token
	ExpirationTime   string // Token 过期时间（UTC+8）
	AuthStatus       int    // 授权状态：1=有效, 0=无效, 2=过期
	AuthAppId        string // 平台 app_id
	AuthorizationId  int64  // 授权记录 ID
}

// AdAuthorization 广告授权概要（不含 Token）
type AdAuthorization struct {
	ID              int64  // 授权记录 ID
	AuthUserId      string // 平台授权用户 ID
	AuthUserName    string // 授权用户名称
	AuthAppId       string // 平台 app_id
	AuthStatus      int    // 授权状态：1=有效, 0=无效, 2=过期
	ExpirationTime  string // Token 过期时间（UTC+8）
	AppId           int64  // 系统应用 ID
	TeamId          int64  // 团队 ID
	AuthTime        string // 授权时间（UTC+8）
}

// AdPlatformAccount 广告投放账户
type AdPlatformAccount struct {
	ID               int64  // 投放账户数据库主键
	AdAccountId      string // 平台账户 ID
	AdAccountName    string // 账户名称
	AdAccountType    int    // 账户类型：1=广告主, 2=代理商, 3=媒体
	AuthStatus       int    // 授权状态：1=有效, 0=无效
	AuthorizationIds string // 关联授权 ID 列表（JSON 数组字符串）
	ParentId         int64  // 父账户 ID
	Active           int    // 活跃状态：1=活跃, 2=不活跃
}

// GetAdAuthorizationToken 获取广告授权 Token
// Token 仅在使用时获取，不存储到本地配置文件，保障信息安全
func GetAdAuthorizationToken(ctx context.Context, appId int64, adAccountId string) (*AdAuthorizationToken, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	baseURL := config.GetBaseURL()
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * 1000000000)

	body := map[string]interface{}{
		"appId": appId,
	}
	if adAccountId != "" {
		body["adAccountId"] = adAccountId
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/authorization/token/get")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("AD_AUTH_TOKEN_ERROR",
			fmt.Sprintf("获取广告授权 Token 失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	token := &AdAuthorizationToken{
		AccessToken:     data.Get("accessToken").String(),
		RefreshToken:    data.Get("refreshToken").String(),
		ExpirationTime:  data.Get("expirationTime").String(),
		AuthStatus:      int(data.Get("authStatus").Int()),
		AuthAppId:       data.Get("authAppId").String(),
		AuthorizationId: data.Get("authorizationId").Int(),
	}

	// 检查授权状态
	if token.AuthStatus == 2 {
		return token, cliErr.NewCLIError("TOKEN_EXPIRED",
			fmt.Sprintf("授权 Token 已过期（过期时间: %s），请联系管理员刷新授权", token.ExpirationTime))
	}
	if token.AuthStatus == 0 {
		return token, cliErr.NewCLIError("AUTH_INVALID",
			"授权无效，请联系管理员重新授权")
	}

	return token, nil
}

// ListAdAuthorizations 获取广告授权账户列表（不含 Token）
func ListAdAuthorizations(ctx context.Context, appId int64, teamId int64) ([]AdAuthorization, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	baseURL := config.GetBaseURL()
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * 1000000000)

	body := map[string]interface{}{
		"appId": appId,
	}
	if teamId > 0 {
		body["teamId"] = teamId
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/authorization/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("AD_AUTH_LIST_ERROR",
			fmt.Sprintf("获取广告授权列表失败 (%d)", codeVal), message)
	}

	auths := []AdAuthorization{}
	result.Get("data.list").ForEach(func(_, value gjson.Result) bool {
		auths = append(auths, AdAuthorization{
			ID:             value.Get("id").Int(),
			AuthUserId:     value.Get("authUserId").String(),
			AuthUserName:   value.Get("authUserName").String(),
			AuthAppId:      value.Get("authAppId").String(),
			AuthStatus:     int(value.Get("authStatus").Int()),
			ExpirationTime: value.Get("expirationTime").String(),
			AppId:          value.Get("appId").Int(),
			TeamId:         value.Get("teamId").Int(),
			AuthTime:       value.Get("authTime").String(),
		})
		return true
	})

	return auths, nil
}

// ListAdPlatformAccounts 获取投放账户列表
func ListAdPlatformAccounts(ctx context.Context, appId int64, authorizationId int64, page int, pageSize int) ([]AdPlatformAccount, int64, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, 0, cliErr.ErrAuthRequired
	}

	baseURL := config.GetBaseURL()
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * 1000000000)

	body := map[string]interface{}{
		"appId": appId,
	}
	if authorizationId > 0 {
		body["authorizationId"] = authorizationId
	}
	if page > 0 {
		body["page"] = page
	}
	if pageSize > 0 {
		body["pageSize"] = pageSize
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/platform/account/list")

	if err != nil {
		return nil, 0, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, 0, cliErr.NewCLIErrorWithDetail("AD_ACCOUNT_LIST_ERROR",
			fmt.Sprintf("获取投放账户列表失败 (%d)", codeVal), message)
	}

	accounts := []AdPlatformAccount{}
	result.Get("data.list").ForEach(func(_, value gjson.Result) bool {
		accounts = append(accounts, AdPlatformAccount{
			ID:               value.Get("id").Int(),
			AdAccountId:      value.Get("adAccountId").String(),
			AdAccountName:    value.Get("adAccountName").String(),
			AdAccountType:    int(value.Get("adAccountType").Int()),
			AuthStatus:       int(value.Get("authStatus").Int()),
			AuthorizationIds: value.Get("authorizationIds").String(),
			ParentId:         value.Get("parentId").Int(),
			Active:           int(value.Get("active").Int()),
		})
		return true
	})

	total := result.Get("data.total").Int()

	return accounts, total, nil
}