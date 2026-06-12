package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/adplatform"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/output"
)

// deliverCmd 交付命令组
var deliverCmd = &cobra.Command{
	Use:   "deliver",
	Short: "素材交付到广告平台",
	Long: `将素材上传到广告投放平台（巨量引擎、腾讯广告）。

素材文件存储在火山引擎 TOS 上，需先通过 cbi CLI 获取文件 URL，再上传到投放平台。
平台凭证通过 CreatiBI 后端实时获取，不存储到本地，保障信息安全。

支持的平台：
  oceanengine  巨量引擎（抖音/头条/TikTok）
  tencentads   腾讯广告（广点通）

示例：
  cbi deliver upload-video --platform oceanengine --account-id 123 --video-url <tos-url> --filename video.mp4
  cbi deliver upload-image --platform tencentads --account-id 456 --image-url <tos-url>`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// deliverUploadVideoCmd 上传视频到广告平台
var deliverUploadVideoCmd = &cobra.Command{
	Use:   "upload-video",
	Short: "上传视频到广告平台",
	Long: `将视频素材上传到广告投放平台。

巨量引擎：异步上传，提交 TOS URL 后返回 task_id，可通过 upload-status 查询结果。
腾讯广告：同步上传，CLI 自动从 TOS 下载文件后 multipart 上传。

平台凭证通过 CreatiBI 后端实时获取，不落盘。

示例：
  cbi deliver upload-video --platform oceanengine --account-id 123 --video-url <tos-url> --filename video.mp4
  cbi deliver upload-video --platform tencentads --account-id 456 --video-url <tos-url>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		accountId, _ := cmd.Flags().GetInt64("account-id")

		videoURL, _ := cmd.Flags().GetString("video-url")
		if videoURL == "" {
			return cliErr.NewCLIError("MISSING_VIDEO_URL", "必须指定 --video-url（TOS 视频文件 URL）")
		}

		filename, _ := cmd.Flags().GetString("filename")
		labelsStr, _ := cmd.Flags().GetString("labels")
		isAIGC, _ := cmd.Flags().GetBool("is-aigc")
		description, _ := cmd.Flags().GetString("description")

		// 解析标签
		var labels []string
		if labelsStr != "" {
			for _, l := range strings.Split(labelsStr, ",") {
				l = strings.TrimSpace(l)
				if l != "" {
					labels = append(labels, l)
				}
			}
		}

		// 获取账户类型
		accountType, _ := cmd.Flags().GetString("account-type")
		if accountType == "" {
			accountType = "ADVERTISER"
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台凭证（不落盘）
		credential, err := adplatform.GetPlatformCredential(ctx, platform)
		if err != nil {
			return err
		}

		// 使用后端返回的 account_id（如果 CLI 未指定）
		if accountId == 0 && credential.AccountId > 0 {
			accountId = credential.AccountId
		}
		if accountId == 0 {
			return cliErr.NewCLIError("MISSING_ACCOUNT_ID", "必须指定 --account-id 或后端配置默认账户 ID")
		}

		// 使用后端返回的 account_type（如果 CLI 未指定且后端有值）
		if accountType == "ADVERTISER" && credential.AccountType != "" {
			accountType = credential.AccountType
		}

		// 2. 构建请求
		req := &adplatform.VideoUploadRequest{
			Platform:    platform,
			AccountId:   accountId,
			AccountType: accountType,
			VideoURL:    videoURL,
			Filename:    filename,
			Labels:      labels,
			IsAIGC:      isAIGC,
			Description: description,
		}

		// 3. 获取平台客户端（传入凭证）
		client := getPlatformClient(platform, credential.AccessToken)

		result, err := client.UploadVideo(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch result.Platform {
		case adplatform.PlatformOceanEngine:
			fmt.Fprintln(cmd.OutOrStdout(), "✓ 视频上传任务已提交（巨量引擎异步处理）")
			fmt.Fprintf(cmd.OutOrStdout(), "  任务 ID: %d\n", result.TaskId)
			fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %d\n", accountId)
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "💡 使用以下命令查询上传结果：")
			fmt.Fprintf(cmd.OutOrStdout(), "  cbi deliver upload-status --platform oceanengine --account-id %d --task-ids %d\n", accountId, result.TaskId)
		case adplatform.PlatformTencentAds:
			fmt.Fprintln(cmd.OutOrStdout(), "✓ 视频上传成功（腾讯广告）")
			fmt.Fprintf(cmd.OutOrStdout(), "  视频 ID: %d\n", result.VideoId)
			fmt.Fprintf(cmd.OutOrStdout(), "  封面图 ID: %d\n", result.CoverImageId)
			fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %d\n", accountId)
		}
		return nil
	},
}

// deliverUploadImageCmd 上传图片到广告平台
var deliverUploadImageCmd = &cobra.Command{
	Use:   "upload-image",
	Short: "上传图片到广告平台",
	Long: `将图片素材上传到广告投放平台。

巨量引擎：通过 URL 直传图片（UPLOAD_BY_URL）。
腾讯广告：CLI 自动从 TOS 下载文件后 multipart 上传。

平台凭证通过 CreatiBI 后端实时获取，不落盘。

示例：
  cbi deliver upload-image --platform oceanengine --account-id 123 --image-url <tos-url>
  cbi deliver upload-image --platform tencentads --account-id 456 --image-url <tos-url>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		accountId, _ := cmd.Flags().GetInt64("account-id")

		imageURL, _ := cmd.Flags().GetString("image-url")
		if imageURL == "" {
			return cliErr.NewCLIError("MISSING_IMAGE_URL", "必须指定 --image-url（TOS 图片文件 URL）")
		}

		filename, _ := cmd.Flags().GetString("filename")
		isAIGC, _ := cmd.Flags().GetBool("is-aigc")

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台凭证（不落盘）
		credential, err := adplatform.GetPlatformCredential(ctx, platform)
		if err != nil {
			return err
		}

		// 使用后端返回的 account_id（如果 CLI 未指定）
		if accountId == 0 && credential.AccountId > 0 {
			accountId = credential.AccountId
		}
		if accountId == 0 {
			return cliErr.NewCLIError("MISSING_ACCOUNT_ID", "必须指定 --account-id 或后端配置默认账户 ID")
		}

		// 2. 构建请求
		req := &adplatform.ImageUploadRequest{
			Platform:  platform,
			AccountId: accountId,
			ImageURL:  imageURL,
			Filename:  filename,
			IsAIGC:    isAIGC,
		}

		// 3. 获取平台客户端（传入凭证）
		client := getPlatformClient(platform, credential.AccessToken)

		result, err := client.UploadImage(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		platformName := adplatform.PlatformNames[platform]
		fmt.Fprintf(cmd.OutOrStdout(), "✓ 图片上传成功（%s）\n", platformName)
		fmt.Fprintf(cmd.OutOrStdout(), "  图片 ID: %s\n", result.ImageId)
		fmt.Fprintf(cmd.OutOrStdout(), "  尺寸: %d x %d\n", result.Width, result.Height)
		fmt.Fprintf(cmd.OutOrStdout(), "  格式: %s\n", result.Format)
		fmt.Fprintf(cmd.OutOrStdout(), "  大小: %s\n", formatFileSize(result.Size))
		if result.MaterialId > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  素材 ID: %d\n", result.MaterialId)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %d\n", accountId)
		return nil
	},
}

// deliverUploadStatusCmd 查询上传任务状态
var deliverUploadStatusCmd = &cobra.Command{
	Use:   "upload-status",
	Short: "查询上传任务状态",
	Long: `查询巨量引擎异步上传视频任务的处理状态。

视频上传为异步处理，正常会在 3 分钟内完成。
平台凭证通过 CreatiBI 后端实时获取，不落盘。

任务状态：
  PROCESS  处理中
  SUCCESS  成功
  FAILED   失败

示例：
  cbi deliver upload-status --platform oceanengine --account-id 123 --task-ids 789,790`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		if platform != adplatform.PlatformOceanEngine {
			return cliErr.NewCLIError("UNSUPPORTED_OPERATION",
				fmt.Sprintf("%s 不支持异步上传状态查询，上传为同步返回", adplatform.PlatformNames[platform]))
		}

		accountId, _ := cmd.Flags().GetInt64("account-id")

		taskIdsStr, _ := cmd.Flags().GetString("task-ids")
		if taskIdsStr == "" {
			return cliErr.NewCLIError("MISSING_TASK_IDS", "必须指定 --task-ids（任务 ID 列表，逗号分隔）")
		}

		// 解析任务 ID
		var taskIds []int64
		for _, s := range strings.Split(taskIdsStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				id, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return cliErr.NewCLIErrorWithDetail("INVALID_TASK_ID",
						fmt.Sprintf("无效的任务 ID: %s", s), err.Error())
				}
				taskIds = append(taskIds, id)
			}
		}

		// 获取账户类型
		accountType, _ := cmd.Flags().GetString("account-type")
		if accountType == "" {
			accountType = "ADVERTISER"
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台凭证
		credential, err := adplatform.GetPlatformCredential(ctx, platform)
		if err != nil {
			return err
		}

		if accountId == 0 && credential.AccountId > 0 {
			accountId = credential.AccountId
		}
		if accountId == 0 {
			return cliErr.NewCLIError("MISSING_ACCOUNT_ID", "必须指定 --account-id 或后端配置默认账户 ID")
		}

		// 2. 构建请求
		req := &adplatform.UploadStatusRequest{
			Platform:    platform,
			AccountId:   accountId,
			AccountType: accountType,
			TaskIds:     taskIds,
		}

		// 3. 获取平台客户端
		client := getPlatformClient(platform, credential.AccessToken)

		result, err := client.GetUploadStatus(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printUploadStatusTable(cmd, result)
			return nil
		}
	},
}

// requirePlatform 校验并获取平台参数
func requirePlatform(cmd *cobra.Command) (adplatform.Platform, error) {
	platformStr, _ := cmd.Flags().GetString("platform")
	if platformStr == "" {
		return "", cliErr.NewCLIError("MISSING_PLATFORM", "必须指定 --platform（oceanengine 或 tencentads）")
	}

	platform := adplatform.Platform(platformStr)
	if !adplatform.IsValidPlatform(platform) {
		return "", cliErr.NewCLIError("INVALID_PLATFORM",
			fmt.Sprintf("不支持的平台: %s，可选值: oceanengine、tencentads", platformStr))
	}

	return platform, nil
}

// getPlatformClient 获取平台客户端（传入实时获取的凭证）
func getPlatformClient(platform adplatform.Platform, accessToken string) adplatform.PlatformClient {
	switch platform {
	case adplatform.PlatformOceanEngine:
		return adplatform.NewOceanEngineClient(accessToken)
	case adplatform.PlatformTencentAds:
		return adplatform.NewTencentAdsClient(accessToken)
	default:
		return nil
	}
}

// printUploadStatusTable 表格输出上传状态
func printUploadStatusTable(cmd *cobra.Command, result *adplatform.UploadStatusResult) {
	w := cmd.OutOrStdout()

	if len(result.Tasks) == 0 {
		fmt.Fprintln(w, "未找到上传任务")
		return
	}

	fmt.Fprintf(w, "共 %d 个任务\n\n", len(result.Tasks))

	t := output.NewTableWriter(w)
	t.AppendHeader("任务ID", "状态", "错误信息", "视频ID", "素材ID", "创建时间")

	for _, task := range result.Tasks {
		statusName := string(task.Status)
		switch task.Status {
		case adplatform.TaskStatusProcessing:
			statusName = "处理中"
		case adplatform.TaskStatusSuccess:
			statusName = "成功"
		case adplatform.TaskStatusFailed:
			statusName = "失败"
		}

		errorMsg := "-"
		if task.ErrorMsg != "" {
			errorMsg = task.ErrorMsg
		}

		videoId := "-"
		if task.VideoId != "" {
			videoId = task.VideoId
		}

		materialId := "-"
		if task.MaterialId > 0 {
			materialId = strconv.FormatInt(task.MaterialId, 10)
		}

		t.AppendRow(
			strconv.FormatInt(task.TaskId, 10),
			statusName,
			errorMsg,
			videoId,
			materialId,
			task.CreateTime,
		)
	}

	t.Render()
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	if size <= 0 {
		return "-"
	}
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

func init() {
	rootCmd.AddCommand(deliverCmd)
	deliverCmd.AddCommand(deliverUploadVideoCmd)
	deliverCmd.AddCommand(deliverUploadImageCmd)
	deliverCmd.AddCommand(deliverUploadStatusCmd)

	// deliverUploadVideoCmd 参数
	deliverUploadVideoCmd.Flags().String("platform", "", "广告平台（oceanengine 或 tencentads，必填）")
	deliverUploadVideoCmd.Flags().Int64("account-id", 0, "广告账户 ID（必填，如后端配置了默认账户可省略）")
	deliverUploadVideoCmd.Flags().String("video-url", "", "TOS 视频文件 URL（必填）")
	deliverUploadVideoCmd.Flags().String("filename", "", "文件名（巨量引擎必填）")
	deliverUploadVideoCmd.Flags().String("labels", "", "标签列表（逗号分隔，巨量引擎可选）")
	deliverUploadVideoCmd.Flags().Bool("is-aigc", false, "是否为 AIGC 生成的素材")
	deliverUploadVideoCmd.Flags().String("description", "", "视频描述（腾讯广告可选）")
	deliverUploadVideoCmd.Flags().String("account-type", "ADVERTISER", "账户类型（ADVERTISER 或 AGENT）")

	// deliverUploadImageCmd 参数
	deliverUploadImageCmd.Flags().String("platform", "", "广告平台（oceanengine 或 tencentads，必填）")
	deliverUploadImageCmd.Flags().Int64("account-id", 0, "广告账户 ID（必填，如后端配置了默认账户可省略）")
	deliverUploadImageCmd.Flags().String("image-url", "", "TOS 图片文件 URL（必填）")
	deliverUploadImageCmd.Flags().String("filename", "", "文件名（可选）")
	deliverUploadImageCmd.Flags().Bool("is-aigc", false, "是否为 AIGC 生成的素材")

	// deliverUploadStatusCmd 参数
	deliverUploadStatusCmd.Flags().String("platform", "", "广告平台（目前仅支持 oceanengine）")
	deliverUploadStatusCmd.Flags().Int64("account-id", 0, "广告账户 ID（必填，如后端配置了默认账户可省略）")
	deliverUploadStatusCmd.Flags().String("task-ids", "", "任务 ID 列表（逗号分隔，必填）")
	deliverUploadStatusCmd.Flags().String("account-type", "ADVERTISER", "账户类型（ADVERTISER 或 AGENT）")
}