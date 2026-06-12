package adplatform

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// PlatformCredential 广告平台凭证（从 CreatiBI 后端实时获取，不落盘）
type PlatformCredential struct {
	AccessToken    string // 平台 Access-Token
	AccountId      int64  // 广告账户 ID
	OrganizationId int64  // 业务单元 ID（腾讯广告）
	AccountType    string // ADVERTISER 或 AGENT
}

// GetPlatformCredential 从 CreatiBI 后端 OpenAPI 获取广告平台凭证
// 凭证仅在使用时获取，不存储到本地配置文件
func GetPlatformCredential(ctx context.Context, platform Platform) (*PlatformCredential, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	baseURL := config.GetBaseURL()
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * 1000000000) // 30 秒

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"platform": string(platform),
		}).
		Post("/openapi/v1/ad/platform/credential/get")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("PLATFORM_CREDENTIAL_ERROR",
			fmt.Sprintf("获取 %s 平台凭证失败 (%d)", PlatformNames[platform], codeVal), message)
	}

	data := result.Get("data")
	return &PlatformCredential{
		AccessToken:    data.Get("access_token").String(),
		AccountId:      data.Get("account_id").Int(),
		OrganizationId: data.Get("organization_id").Int(),
		AccountType:    data.Get("account_type").String(),
	}, nil
}