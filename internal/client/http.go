package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

var (
	httpClient *Client
)

// Client HTTP 客户端
type Client struct {
	client  *resty.Client
	baseURL string
}

// Init 初始化 HTTP 客户端
func Init() {
	httpClient = NewClient(config.GetBaseURL())
}

// NewClient 创建新的 HTTP 客户端
func NewClient(baseURL string) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(2)
	client.SetRetryWaitTime(1 * time.Second)

	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

// GetClient 获取全局客户端
func GetClient() *Client {
	if httpClient == nil {
		Init()
	}
	return httpClient
}

// Request 请求参数
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// Response 响应结果
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Duration   time.Duration
	RequestID  string
}

// Do 执行请求
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// 获取 API Key
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		return nil, cliErr.ErrAuthRequired
	}

	// 构建 resty 请求
	r := c.client.R()
	r.SetContext(ctx)
	r.SetHeader("Authorization", "Bearer "+apiKey)
	r.SetHeader("Content-Type", "application/json")

	// 添加额外 Headers
	for k, v := range req.Headers {
		r.SetHeader(k, v)
	}

	// 设置请求体
	if req.Body != nil {
		r.SetBody(req.Body)
	}

	// 执行请求
	var resp *resty.Response
	var err error

	switch req.Method {
	case http.MethodGet:
		resp, err = r.Get(req.Path)
	case http.MethodPost:
		resp, err = r.Post(req.Path)
	case http.MethodPut:
		resp, err = r.Put(req.Path)
	case http.MethodDelete:
		resp, err = r.Delete(req.Path)
	default:
		return nil, cliErr.NewCLIError("INVALID_METHOD", "不支持的请求方法")
	}

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 构建响应
	response := &Response{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
		Duration:   resp.Time(),
		RequestID:  resp.Header().Get("X-Request-ID"),
	}

	// 检查错误状态码
	if resp.StatusCode() >= 400 {
		return response, parseErrorResponse(resp.StatusCode(), resp.Body())
	}

	return response, nil
}

// parseErrorResponse 解析错误响应
func parseErrorResponse(statusCode int, body []byte) error {
	// 从响应中提取错误信息
	bodyStr := string(body)
	reason := gjson.Get(bodyStr, "reason").String()
	message := gjson.Get(bodyStr, "message").String()

	switch statusCode {
	case http.StatusUnauthorized:
		if reason == "TOKEN_EXPIRED" {
			return cliErr.ErrTokenExpired
		}
		return cliErr.ErrAuthRequired
	case http.StatusForbidden:
		return cliErr.ErrPermissionDenied
	case http.StatusNotFound:
		return cliErr.NewCLIErrorWithDetail("NOT_FOUND", "资源不存在", message)
	case http.StatusBadRequest:
		return cliErr.NewCLIErrorWithDetail("INVALID_ARGUMENT", "参数错误", message)
	case http.StatusServiceUnavailable:
		return cliErr.ErrUpstreamTimeout
	default:
		if statusCode >= 500 {
			return cliErr.ErrUpstreamError
		}
		return cliErr.NewCLIErrorWithDetail("API_ERROR", fmt.Sprintf("API 错误 (%d)", statusCode), message)
	}
}

// GetJSON 获取 JSON 结果
func (c *Client) GetJSON(ctx context.Context, path string) (gjson.Result, error) {
	resp, err := c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
	})
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(resp.Body), nil
}

// PostJSON POST JSON 并获取结果
func (c *Client) PostJSON(ctx context.Context, path string, body interface{}) (gjson.Result, error) {
	resp, err := c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(resp.Body), nil
}

// PostStream POST 并流式返回
func (c *Client) PostStream(ctx context.Context, path string, body interface{}) (io.ReadCloser, error) {
	// 获取 API Key
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		return nil, cliErr.ErrAuthRequired
	}

	r := c.client.R()
	r.SetContext(ctx)
	r.SetHeader("Authorization", "Bearer "+apiKey)
	r.SetHeader("Content-Type", "application/json")
	r.SetBody(body)
	r.SetDoNotParseResponse(true) // 不自动解析响应

	resp, err := r.Post(path)
	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() >= 400 {
		defer resp.RawBody().Close()
		return nil, parseErrorResponse(resp.StatusCode(), []byte{})
	}

	return resp.RawBody(), nil
}

// RawJSON 原始 JSON 字节数组
func (r *Response) RawJSON() []byte {
	return r.Body
}

// ParseJSON 解析 JSON
func (r *Response) ParseJSON() gjson.Result {
	return gjson.ParseBytes(r.Body)
}

// ParseJSONTo 解析 JSON 到结构体
func (r *Response) ParseJSONTo(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}