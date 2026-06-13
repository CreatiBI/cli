package adplatform

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	cliErr "github.com/CreatiBI/cli/internal/errors"
)

const oceanEngineBaseURL = "https://api.oceanengine.com"

// OceanEngineClient 巨量引擎客户端
type OceanEngineClient struct {
	client      *resty.Client
	accessToken string // 平台凭证（实时获取，不落盘）
}

// NewOceanEngineClient 创建巨量引擎客户端
func NewOceanEngineClient(accessToken string) *OceanEngineClient {
	return &OceanEngineClient{
		client: resty.New().
			SetBaseURL(oceanEngineBaseURL).
			SetTimeout(60 * 1000000000), // 60 秒
		accessToken: accessToken,
	}
}

// UploadVideo 异步上传视频到巨量引擎
func (c *OceanEngineClient) UploadVideo(ctx context.Context, req *VideoUploadRequest) (*VideoUploadResult, error) {
	if c.accessToken == "" {
		return nil, cliErr.NewCLIError("PLATFORM_AUTH_REQUIRED",
			"巨量引擎凭证为空，请确认后端已配置该平台授权")
	}

	if req.Filename == "" {
		return nil, cliErr.NewCLIError("MISSING_FILENAME", "巨量引擎视频上传必须指定 --filename")
	}

	body := map[string]interface{}{
		"account_id":   req.AccountId,
		"account_type": req.AccountType,
		"filename":     req.Filename,
		"video_url":    req.VideoURL,
	}

	if len(req.Labels) > 0 {
		body["labels"] = req.Labels
	}
	if req.IsAIGC {
		body["is_aigc"] = true
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Access-Token", c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/open_api/2/file/upload_task/create/")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("OCEANENGINE_VIDEO_UPLOAD_ERROR",
			fmt.Sprintf("巨量引擎视频上传失败 (%d)", codeVal), message)
	}

	return &VideoUploadResult{
		Platform: PlatformOceanEngine,
		TaskId:   result.Get("data.task_id").Int(),
	}, nil
}

// UploadImage 上传图片到巨量引擎
func (c *OceanEngineClient) UploadImage(ctx context.Context, req *ImageUploadRequest) (*ImageUploadResult, error) {
	if c.accessToken == "" {
		return nil, cliErr.NewCLIError("PLATFORM_AUTH_REQUIRED",
			"巨量引擎凭证为空，请确认后端已配置该平台授权")
	}

	// 巨量引擎图片上传支持 UPLOAD_BY_URL，直接用 TOS URL
	body := map[string]interface{}{
		"advertiser_id": req.AccountId,
		"upload_type":   "UPLOAD_BY_URL",
		"image_url":     req.ImageURL,
	}

	if req.Filename != "" {
		body["filename"] = req.Filename
	}
	if req.IsAIGC {
		body["is_aigc"] = true
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Access-Token", c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/open_api/2/file/image/ad/")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("OCEANENGINE_IMAGE_UPLOAD_ERROR",
			fmt.Sprintf("巨量引擎图片上传失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	return &ImageUploadResult{
		Platform:   PlatformOceanEngine,
		ImageId:    data.Get("id").String(),
		Size:       data.Get("size").Int(),
		Width:      int(data.Get("width").Int()),
		Height:     int(data.Get("height").Int()),
		URL:        data.Get("url").String(),
		Format:     data.Get("format").String(),
		Signature:  data.Get("signature").String(),
		MaterialId: data.Get("material_id").Int(),
	}, nil
}

// GetUploadStatus 查询巨量引擎异步上传任务状态
func (c *OceanEngineClient) GetUploadStatus(ctx context.Context, req *UploadStatusRequest) (*UploadStatusResult, error) {
	if c.accessToken == "" {
		return nil, cliErr.NewCLIError("PLATFORM_AUTH_REQUIRED",
			"巨量引擎凭证为空，请确认后端已配置该平台授权")
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Access-Token", c.accessToken).
		SetQueryParam("account_id", fmt.Sprintf("%d", req.AccountId)).
		SetQueryParam("account_type", req.AccountType).
		SetQueryParam("task_ids", formatTaskIds(req.TaskIds)).
		Get("/open_api/2/file/video/upload_task/list/")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		return nil, cliErr.NewCLIErrorWithDetail("OCEANENGINE_UPLOAD_STATUS_ERROR",
			fmt.Sprintf("巨量引擎上传状态查询失败 (%d)", codeVal), message)
	}

	tasks := []UploadTaskInfo{}
	result.Get("data.list").ForEach(func(_, value gjson.Result) bool {
		task := UploadTaskInfo{
			TaskId:     value.Get("task_id").Int(),
			Status:     UploadTaskStatus(value.Get("status").String()),
			ErrorMsg:   value.Get("error_msg").String(),
			CreateTime: value.Get("create_time").String(),
		}

		// 解析 video_info（仅 SUCCESS 时存在）
		videoInfo := value.Get("video_info")
		if videoInfo.Exists() {
			task.VideoId = videoInfo.Get("video_id").String()
			task.MaterialId = videoInfo.Get("material_id").Int()
			task.Size = videoInfo.Get("size").Int()
			task.Width = int(videoInfo.Get("width").Int())
			task.Height = int(videoInfo.Get("height").Int())
			task.VideoURL = videoInfo.Get("video_url").String()
			task.Duration = videoInfo.Get("duration").Float()
			task.VideoSignature = videoInfo.Get("video_signature").String()
		}

		tasks = append(tasks, task)
		return true
	})

	return &UploadStatusResult{Tasks: tasks}, nil
}

// formatTaskIds 格式化任务 ID 列表为 JSON 数组字符串
func formatTaskIds(ids []int64) string {
	arr := "["
	for i, id := range ids {
		if i > 0 {
			arr += ","
		}
		arr += fmt.Sprintf("%d", id)
	}
	arr += "]"
	return arr
}