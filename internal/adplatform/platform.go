package adplatform

import (
	"context"
)

// Platform 广告投放平台枚举
type Platform string

const (
	PlatformOceanEngine Platform = "oceanengine" // 巨量引擎
	PlatformTencentAds  Platform = "tencentads"  // 腾讯广告
)

// ValidPlatforms 有效平台列表
var ValidPlatforms = []Platform{PlatformOceanEngine, PlatformTencentAds}

// IsValidPlatform 检查平台是否有效
func IsValidPlatform(p Platform) bool {
	for _, vp := range ValidPlatforms {
		if vp == p {
			return true
		}
	}
	return false
}

// PlatformNames 平台显示名称
var PlatformNames = map[Platform]string{
	PlatformOceanEngine: "巨量引擎",
	PlatformTencentAds:  "腾讯广告",
}

// VideoUploadRequest 视频上传请求（公共）
type VideoUploadRequest struct {
	Platform    Platform
	AccountId   int64
	AccountType string // ADVERTISER 或 AGENT（巨量引擎）
	VideoURL    string // TOS 视频文件 URL
	Filename    string // 文件名
	Labels      []string // 标签（巨量引擎）
	IsAIGC      bool   // 是否 AIGC 生成
	Description string // 描述（腾讯广告）
}

// VideoUploadResult 视频上传结果（公共）
type VideoUploadResult struct {
	Platform Platform
	// 巨量引擎异步结果
	TaskId int64
	// 腾讯广告同步结果
	VideoId      int64
	CoverImageId int64
}

// ImageUploadRequest 图片上传请求（公共）
type ImageUploadRequest struct {
	Platform  Platform
	AccountId int64
	ImageURL  string // TOS 图片文件 URL
	Filename  string // 文件名
	IsAIGC    bool   // 是否 AIGC 生成
}

// ImageUploadResult 图片上传结果（公共）
type ImageUploadResult struct {
	Platform Platform
	// 巨量引擎结果
	ImageId    string
	Size       int64
	Width      int
	Height     int
	URL        string
	Format     string
	Signature  string
	MaterialId int64
	// 腾讯广告结果（字段名不同，复用上述字段）
}

// UploadStatusRequest 上传状态查询请求（公共）
type UploadStatusRequest struct {
	Platform    Platform
	AccountId   int64
	AccountType string // ADVERTISER 或 AGENT
	TaskIds     []int64
}

// UploadTaskStatus 上传任务状态
type UploadTaskStatus string

const (
	TaskStatusProcessing UploadTaskStatus = "PROCESS" // 处理中
	TaskStatusSuccess    UploadTaskStatus = "SUCCESS" // 成功
	TaskStatusFailed     UploadTaskStatus = "FAILED"  // 失败
)

// UploadTaskInfo 上传任务信息
type UploadTaskInfo struct {
	TaskId     int64
	Status     UploadTaskStatus
	ErrorMsg   string
	CreateTime string
	// 成功时的视频信息
	VideoId        string
	MaterialId     int64
	Size           int64
	Width          int
	Height         int
	VideoURL       string
	Duration       float64
	VideoSignature string
}

// UploadStatusResult 上传状态查询结果（公共）
type UploadStatusResult struct {
	Tasks []UploadTaskInfo
}

// PlatformClient 广告平台客户端接口
type PlatformClient interface {
	UploadVideo(ctx context.Context, req *VideoUploadRequest) (*VideoUploadResult, error)
	UploadImage(ctx context.Context, req *ImageUploadRequest) (*ImageUploadResult, error)
	GetUploadStatus(ctx context.Context, req *UploadStatusRequest) (*UploadStatusResult, error)
}