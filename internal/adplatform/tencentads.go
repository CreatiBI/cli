package adplatform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/upload"
)

const tencentAdsBaseURL = "https://api.e.qq.com/v3.0"

// TencentAdsClient 腾讯广告客户端
type TencentAdsClient struct {
	client      *resty.Client
	accessToken string // 平台凭证（实时获取，不落盘）
}

// NewTencentAdsClient 创建腾讯广告客户端
func NewTencentAdsClient(accessToken string) *TencentAdsClient {
	return &TencentAdsClient{
		client: resty.New().
			SetBaseURL(tencentAdsBaseURL).
			SetTimeout(120 * 1000000000), // 120 秒（大文件上传需要更长时间）
		accessToken: accessToken,
	}
}

// UploadVideo 上传视频到腾讯广告
// 腾讯广告不支持 URL 直传，需要先下载文件再 multipart 上传
func (c *TencentAdsClient) UploadVideo(ctx context.Context, req *VideoUploadRequest) (*VideoUploadResult, error) {
	if c.accessToken == "" {
		return nil, cliErr.NewCLIError("PLATFORM_AUTH_REQUIRED",
			"腾讯广告凭证为空，请确认后端已配置该平台授权")
	}

	// 1. 下载文件到临时目录
	localPath, err := downloadToTempFile(ctx, req.VideoURL, "cbi-video-*")
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("FILE_DOWNLOAD_ERROR",
			"从 TOS 下载视频文件失败", err.Error())
	}
	defer os.Remove(localPath) // 清理临时文件

	// 2. 计算 MD5 签名
	signature, err := upload.ComputeMD5(localPath)
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("MD5_COMPUTE_ERROR",
			"计算视频文件 MD5 失败", err.Error())
	}

	// 3. 构建认证参数
	authParams := buildTencentAdsAuthParams(c.accessToken)

	// 4. 构建 form data
	formData := map[string]string{
		"signature": signature,
	}
	// account_id 或 organization_id
	if req.AccountId > 0 {
		formData["account_id"] = fmt.Sprintf("%d", req.AccountId)
	}

	// 5. multipart 上传
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(authParams).
		SetFormData(formData).
		SetFile("video_file", localPath).
		Post("/videos/add")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("TENCENTADS_VIDEO_UPLOAD_ERROR",
			fmt.Sprintf("腾讯广告视频上传失败 (%d)", codeVal), message)
	}

	return &VideoUploadResult{
		Platform:     PlatformTencentAds,
		VideoId:      result.Get("data.video_id").Int(),
		CoverImageId: result.Get("data.cover_image_id").Int(),
	}, nil
}

// UploadImage 上传图片到腾讯广告
// 腾讯广告图片不支持 URL 直传，需要先下载文件再 multipart 上传
func (c *TencentAdsClient) UploadImage(ctx context.Context, req *ImageUploadRequest) (*ImageUploadResult, error) {
	if c.accessToken == "" {
		return nil, cliErr.NewCLIError("PLATFORM_AUTH_REQUIRED",
			"腾讯广告凭证为空，请确认后端已配置该平台授权")
	}

	// 1. 下载文件到临时目录
	localPath, err := downloadToTempFile(ctx, req.ImageURL, "cbi-image-*")
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("FILE_DOWNLOAD_ERROR",
			"从 TOS 下载图片文件失败", err.Error())
	}
	defer os.Remove(localPath) // 清理临时文件

	// 2. 计算 MD5 签名
	signature, err := upload.ComputeMD5(localPath)
	if err != nil {
		return nil, cliErr.NewCLIErrorWithDetail("MD5_COMPUTE_ERROR",
			"计算图片文件 MD5 失败", err.Error())
	}

	// 3. 构建认证参数
	authParams := buildTencentAdsAuthParams(c.accessToken)

	// 4. 构建 form data
	formData := map[string]string{
		"upload_type": "UPLOAD_TYPE_FILE",
		"signature":   signature,
	}
	if req.AccountId > 0 {
		formData["account_id"] = fmt.Sprintf("%d", req.AccountId)
	}

	// 5. multipart 上传
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(authParams).
		SetFormData(formData).
		SetFile("file", localPath).
		Post("/images/add")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("TENCENTADS_IMAGE_UPLOAD_ERROR",
			fmt.Sprintf("腾讯广告图片上传失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	return &ImageUploadResult{
		Platform:   PlatformTencentAds,
		ImageId:    fmt.Sprintf("%d", data.Get("image_id").Int()),
		Size:       data.Get("size").Int(),
		Width:      int(data.Get("width").Int()),
		Height:     int(data.Get("height").Int()),
		URL:        data.Get("url").String(),
		Format:     data.Get("format").String(),
		Signature:  signature,
		MaterialId: data.Get("material_id").Int(),
	}, nil
}

// GetUploadStatus 腾讯广告不支持异步上传
func (c *TencentAdsClient) GetUploadStatus(ctx context.Context, req *UploadStatusRequest) (*UploadStatusResult, error) {
	return nil, cliErr.NewCLIError("UNSUPPORTED_OPERATION",
		"腾讯广告不支持异步上传状态查询，视频和图片上传均为同步返回结果")
}

// downloadToTempFile 从 URL 下载文件到临时目录
func downloadToTempFile(ctx context.Context, url, pattern string) (string, error) {
	tmpDir := os.TempDir()

	file, err := os.CreateTemp(tmpDir, filepath.Base(pattern))
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求失败: %w", err)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，HTTP 状态码: %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	return file.Name(), nil
}

// buildTencentAdsAuthParams 构建腾讯广告 OAuth2 认证查询参数
func buildTencentAdsAuthParams(accessToken string) map[string]string {
	return map[string]string{
		"access_token": accessToken,
		"timestamp":    fmt.Sprintf("%d", time.Now().Unix()),
		"nonce":        generateNonce(),
	}
}

// generateNonce 生成随机 nonce 字符串（32字节）
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}