package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// OAuthConfig OAuth 配置
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	RedirectURL  string
	Scope        string
	UserInfoURL  string
}

// DefaultOAuthConfig 默认 OAuth 配置（从本地配置读取）
func DefaultOAuthConfig() *OAuthConfig {
	baseURL := config.GetBaseURL()

	return &OAuthConfig{
		ClientID:     config.GetClientID(),
		ClientSecret: config.GetClientSecret(),
		AuthorizeURL: "https://app.creatibi.cn/oauth/authorize",
		TokenURL:     baseURL + "/openapi/v1/authen/oauth/token",
		RedirectURL:  "http://localhost:8080/callback",
		Scope:        "user:profile repository",
		UserInfoURL:  baseURL + "/openapi/v1/user/info",
	}
}

// OAuthClient OAuth 客户端
type OAuthClient struct {
	config *OAuthConfig
	client *resty.Client
}

// NewOAuthClient 创建 OAuth 客户端
func NewOAuthClient(cfg *OAuthConfig) *OAuthClient {
	if cfg == nil {
		cfg = DefaultOAuthConfig()
	}
	return &OAuthClient{
		config: cfg,
		client: resty.New().SetTimeout(30 * time.Second),
	}
}

// StartOAuthFlow 启动 OAuth 登录流程
func (c *OAuthClient) StartOAuthFlow(ctx context.Context) error {
	// 生成随机 state 防止 CSRF
	state, err := generateState()
	if err != nil {
		return err
	}

	// 尝试多个端口
	port := 8080
	maxPort := 8090
	var server *http.Server
	var callbackURL string

	for ; port <= maxPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			callbackURL = fmt.Sprintf("http://localhost:%d/callback", port)
			server = &http.Server{Addr: addr}
			break
		}
	}

	if server == nil {
		return cliErr.NewCLIError("PORT_UNAVAILABLE", "无法找到可用端口 (8080-8090)")
	}

	// 更新 redirect URL
	c.config.RedirectURL = callbackURL

	// 构建授权 URL
	authURL := c.buildAuthorizeURL(state)

	fmt.Println("正在启动 OAuth 登录...")
	fmt.Println("请在浏览器中完成授权")
	fmt.Println()
	fmt.Println("授权 URL:")
	fmt.Println(authURL)
	fmt.Println()

	// 启动本地回调服务器
	tokenChan := make(chan string, 1)
	errChan := make(chan error, 1)

	http.HandleFunc("/callback", c.handleCallback(state, tokenChan, errChan))

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 尝试自动打开浏览器
	openBrowser(authURL)

	fmt.Println("等待授权回调... (按 Ctrl+C 取消)")

	// 等待回调或超时
	select {
	case token := <-tokenChan:
		// 关闭服务器
		server.Shutdown(ctx)
		// 存储 token
		if err := config.SetAPIKey(token); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("✓ 登录成功")
		fmt.Printf("Token 已存储到: %s\n", config.GetConfigFile())
		return nil

	case err := <-errChan:
		server.Shutdown(ctx)
		return err

	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return cliErr.NewCLIError("OAUTH_TIMEOUT", "OAuth 授权超时")

	case <-ctx.Done():
		server.Shutdown(ctx)
		return ctx.Err()
	}
}

// buildAuthorizeURL 构建授权 URL
func (c *OAuthClient) buildAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", c.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("scope", c.config.Scope)

	return c.config.AuthorizeURL + "?" + params.Encode()
}

// handleCallback 处理 OAuth 回调
func (c *OAuthClient) handleCallback(expectedState string, tokenChan chan string, errChan chan error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取授权码
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		// 验证 state
		if state != expectedState {
			errChan <- cliErr.NewCLIError("OAUTH_STATE_MISMATCH", "OAuth state 不匹配")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "<html><body><h2>授权失败</h2><p>state 不匹配</p><script>window.close();</script></body></html>")
			return
		}

		if code == "" {
			err := r.URL.Query().Get("error")
			errDesc := r.URL.Query().Get("error_description")
			errChan <- cliErr.NewCLIErrorWithDetail("OAUTH_DENIED", "用户拒绝授权", errDesc)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h2>授权失败</h2><p>%s - %s</p><script>window.close();</script></body></html>\n", err, errDesc)
			return
		}

		// 用授权码换取 token
		token, err := c.exchangeCodeForToken(code)
		if err != nil {
			errChan <- err
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "<html><body><h2>获取 Token 失败</h2><p>"+err.Error()+"</p><script>window.close();</script></body></html>")
			return
		}

		// 成功 - 返回自动关闭的页面
		tokenChan <- token
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `<html>
<head>
<title>授权成功</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding: 50px; }
.success { color: #28a745; }
</style>
</head>
<body>
<h2 class="success">✓ 授权成功</h2>
<p>CLI 登录已完成，此页面将自动关闭...</p>
<script>
setTimeout(function() {
  window.close();
  document.body.innerHTML = '<h2 class="success">✓ 授权成功</h2><p>CLI 登录已完成，请手动关闭此页面。</p>';
}, 1000);
</script>
</body>
</html>`)
	}
}

// exchangeCodeForToken 用授权码换取 token
func (c *OAuthClient) exchangeCodeForToken(code string) (string, error) {
	// 检查 client_secret
	if c.config.ClientSecret == "" {
		return "", cliErr.NewCLIError("OAUTH_CLIENT_SECRET_REQUIRED",
			"缺少 client_secret，请先配置应用密钥")
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  c.config.RedirectURL,
			"client_id":     c.config.ClientID,
			"client_secret": c.config.ClientSecret,
		}).
		Post(c.config.TokenURL)

	if err != nil {
		return "", cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 解析响应
	result := gjson.ParseBytes(resp.Body())

	// 检查错误码 (code != 0 表示失败)
	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		errorMsg := result.Get("error").String()
		errorDesc := result.Get("error_description").String()
		return "", cliErr.NewCLIErrorWithDetail("OAUTH_TOKEN_ERROR", errorMsg, errorDesc)
	}

	// 获取 access_token
	accessToken := result.Get("access_token").String()
	if accessToken == "" {
		return "", cliErr.NewCLIError("OAUTH_TOKEN_MISSING", "响应中未包含 access_token")
	}

	// 存储 refresh_token 和有效期信息
	refreshToken := result.Get("refresh_token").String()
	if refreshToken != "" {
		config.SetRefreshToken(refreshToken)
	}

	// 存储 token 过期时间
	expiresIn := result.Get("expires_in").Int()
	if expiresIn > 0 {
		config.SetTokenExpiresAt(time.Now().Add(time.Duration(expiresIn) * time.Second))
	}

	// 存储 refresh_token 过期时间
	refreshExpiresIn := result.Get("refresh_token_expires_in").Int()
	if refreshExpiresIn > 0 {
		config.SetRefreshTokenExpiresAt(time.Now().Add(time.Duration(refreshExpiresIn) * time.Second))
	}

	return accessToken, nil
}

// generateState 生成随机 state
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// openBrowser 尝试打开浏览器
func openBrowser(url string) {
	var cmd string
	var args []string

	switch {
	case isCommandAvailable("open"): // macOS
		cmd = "open"
		args = []string{url}
	case isCommandAvailable("xdg-open"): // Linux
		cmd = "xdg-open"
		args = []string{url}
	case isCommandAvailable("start"): // Windows
		cmd = "start"
		args = []string{url}
	default:
		fmt.Println("无法自动打开浏览器，请手动复制上述 URL 到浏览器中打开")
		return
	}

	exec.Command(cmd, args...).Start()
	fmt.Printf("已尝试打开浏览器\n")
}

// isCommandAvailable 检查命令是否可用
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RefreshToken 刷新 token
func (c *OAuthClient) RefreshToken(refreshToken string) (string, error) {
	if c.config.ClientSecret == "" {
		return "", cliErr.NewCLIError("OAUTH_CLIENT_SECRET_REQUIRED",
			"缺少 client_secret，无法刷新 token")
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
			"client_id":     c.config.ClientID,
			"client_secret": c.config.ClientSecret,
		}).
		Post(c.config.TokenURL)

	if err != nil {
		return "", cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	// 检查错误码
	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		return "", cliErr.ErrTokenExpired
	}

	accessToken := result.Get("access_token").String()

	// 更新 refresh_token（如果返回了新的）
	newRefreshToken := result.Get("refresh_token").String()
	if newRefreshToken != "" {
		config.SetRefreshToken(newRefreshToken)
	}

	// 更新过期时间
	expiresIn := result.Get("expires_in").Int()
	if expiresIn > 0 {
		config.SetTokenExpiresAt(time.Now().Add(time.Duration(expiresIn) * time.Second))
	}

	return accessToken, nil
}

// GetUserInfo 获取用户信息
func (c *OAuthClient) GetUserInfo(accessToken string) (*UserInfo, error) {
	resp, err := c.client.R().
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody("{}").
		Post(c.config.UserInfoURL)

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	// 检查错误码
	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("USER_INFO_ERROR", "获取用户信息失败", message)
	}

	userData := result.Get("data.user")
	return &UserInfo{
		ID:     userData.Get("id").Int(),
		Name:   userData.Get("name").String(),
		Email:  userData.Get("email").String(),
		Avatar: userData.Get("avatar").String(),
	}, nil
}

// UserInfo 用户信息
type UserInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}