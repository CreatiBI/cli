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
平台凭证通过 CreatiBI 后端实时获取（不落盘），保障信息安全。

支持的平台：
  oceanengine  巨量引擎（抖音/头条/TikTok）
  tencentads   腾讯广告（广点通）

使用流程：
  1. cbi deliver auth-list --app-id <ID>        查看可用授权
  2. cbi deliver account-list --app-id <ID>     查看投放账户
  3. cbi deliver upload-video --app-id <ID> --ad-account-id <ID> --video-url <url>  上传视频
  4. cbi deliver upload-image --app-id <ID> --ad-account-id <ID> --image-url <url>  上传图片`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// deliverAuthListCmd 查看广告授权列表
var deliverAuthListCmd = &cobra.Command{
	Use:   "auth-list",
	Short: "查看广告授权账户列表",
	Long: `获取广告平台授权概要列表（不含 Token 等敏感信息）。

需要 Token 时，上传操作会自动通过后端获取，不在此展示。

授权状态：
  1 = 有效
  0 = 无效
  2 = 过期

示例：
  cbi deliver auth-list --app-id 5
  cbi deliver auth-list --app-id 5 --team-id 10`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, _ := cmd.Flags().GetInt64("app-id")
		if appId == 0 {
			return cliErr.NewCLIError("MISSING_APP_ID", "必须指定 --app-id（广告平台应用 ID）")
		}

		teamId, _ := cmd.Flags().GetInt64("team-id")

		ctx, cancel := newSignalCtx()
		defer cancel()

		auths, err := adplatform.ListAdAuthorizations(ctx, appId, teamId)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, auths)
		}

		switch format {
		case "json":
			return outputData(cmd, auths)
		default:
			printAuthListTable(cmd, auths)
			return nil
		}
	},
}

// deliverAccountListCmd 查看投放账户列表
var deliverAccountListCmd = &cobra.Command{
	Use:   "account-list",
	Short: "查看广告投放账户列表",
	Long: `获取广告投放账户列表，用于指定上传素材的目标账户。

账户类型：
  1 = 广告主
  2 = 代理商
  3 = 媒体

示例：
  cbi deliver account-list --app-id 5
  cbi deliver account-list --app-id 5 --authorization-id 45`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, _ := cmd.Flags().GetInt64("app-id")
		if appId == 0 {
			return cliErr.NewCLIError("MISSING_APP_ID", "必须指定 --app-id（广告平台应用 ID）")
		}

		authorizationId, _ := cmd.Flags().GetInt64("authorization-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		accounts, total, err := adplatform.ListAdPlatformAccounts(ctx, appId, authorizationId, page, pageSize)
		if err != nil {
			return err
		}

		result := map[string]interface{}{
			"accounts": accounts,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printAccountListTable(cmd, accounts, total, page, pageSize)
			return nil
		}
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
  cbi deliver upload-video --app-id 5 --ad-account-id "12345" --video-url <tos-url> --filename video.mp4
  cbi deliver upload-video --app-id 5 --ad-account-id "12345" --video-url <tos-url> --platform tencentads`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, _ := cmd.Flags().GetInt64("app-id")
		if appId == 0 {
			return cliErr.NewCLIError("MISSING_APP_ID", "必须指定 --app-id（广告平台应用 ID）")
		}

		adAccountId, _ := cmd.Flags().GetString("ad-account-id")

		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		videoURL, _ := cmd.Flags().GetString("video-url")
		if videoURL == "" {
			return cliErr.NewCLIError("MISSING_VIDEO_URL", "必须指定 --video-url（TOS 视频文件 URL）")
		}

		filename, _ := cmd.Flags().GetString("filename")
		labelsStr, _ := cmd.Flags().GetString("labels")
		isAIGC, _ := cmd.Flags().GetBool("is-aigc")
		description, _ := cmd.Flags().GetString("description")
		accountType, _ := cmd.Flags().GetString("account-type")

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

		if accountType == "" {
			accountType = "ADVERTISER"
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台 Token（不落盘）
		token, err := adplatform.GetAdAuthorizationToken(ctx, appId, adAccountId)
		if err != nil {
			return err
		}

		// 2. 构建请求
		req := &adplatform.VideoUploadRequest{
			Platform:    platform,
			AccountId:   0, // 由平台客户端自行处理
			AccountType: accountType,
			VideoURL:    videoURL,
			Filename:    filename,
			Labels:      labels,
			IsAIGC:      isAIGC,
			Description: description,
		}

		// 巨量引擎需要 accountId
		if platform == adplatform.PlatformOceanEngine && adAccountId != "" {
			id, _ := strconv.ParseInt(adAccountId, 10, 64)
			req.AccountId = id
		}
		// 腾讯广告需要 accountId
		if platform == adplatform.PlatformTencentAds && adAccountId != "" {
			id, _ := strconv.ParseInt(adAccountId, 10, 64)
			req.AccountId = id
		}

		// 3. 获取平台客户端（传入实时获取的 Token）
		client := getPlatformClient(platform, token.AccessToken)

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
			fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %s\n", adAccountId)
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "💡 使用以下命令查询上传结果：")
			fmt.Fprintf(cmd.OutOrStdout(), "  cbi deliver upload-status --app-id %d --task-ids %d\n", appId, result.TaskId)
		case adplatform.PlatformTencentAds:
			fmt.Fprintln(cmd.OutOrStdout(), "✓ 视频上传成功（腾讯广告）")
			fmt.Fprintf(cmd.OutOrStdout(), "  视频 ID: %d\n", result.VideoId)
			fmt.Fprintf(cmd.OutOrStdout(), "  封面图 ID: %d\n", result.CoverImageId)
			fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %s\n", adAccountId)
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
  cbi deliver upload-image --app-id 5 --ad-account-id "12345" --image-url <tos-url> --platform oceanengine
  cbi deliver upload-image --app-id 5 --ad-account-id "12345" --image-url <tos-url> --platform tencentads`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, _ := cmd.Flags().GetInt64("app-id")
		if appId == 0 {
			return cliErr.NewCLIError("MISSING_APP_ID", "必须指定 --app-id（广告平台应用 ID）")
		}

		adAccountId, _ := cmd.Flags().GetString("ad-account-id")

		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		imageURL, _ := cmd.Flags().GetString("image-url")
		if imageURL == "" {
			return cliErr.NewCLIError("MISSING_IMAGE_URL", "必须指定 --image-url（TOS 图片文件 URL）")
		}

		filename, _ := cmd.Flags().GetString("filename")
		isAIGC, _ := cmd.Flags().GetBool("is-aigc")

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台 Token
		token, err := adplatform.GetAdAuthorizationToken(ctx, appId, adAccountId)
		if err != nil {
			return err
		}

		// 2. 构建请求
		accountId := int64(0)
		if adAccountId != "" {
			accountId, _ = strconv.ParseInt(adAccountId, 10, 64)
		}

		req := &adplatform.ImageUploadRequest{
			Platform:  platform,
			AccountId: accountId,
			ImageURL:  imageURL,
			Filename:  filename,
			IsAIGC:    isAIGC,
		}

		// 3. 获取平台客户端
		client := getPlatformClient(platform, token.AccessToken)

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
		fmt.Fprintf(cmd.OutOrStdout(), "  账户 ID: %s\n", adAccountId)
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
  cbi deliver upload-status --app-id 5 --platform oceanengine --task-ids 789,790`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, _ := cmd.Flags().GetInt64("app-id")
		if appId == 0 {
			return cliErr.NewCLIError("MISSING_APP_ID", "必须指定 --app-id（广告平台应用 ID）")
		}

		platform, err := requirePlatform(cmd)
		if err != nil {
			return err
		}

		if platform != adplatform.PlatformOceanEngine {
			return cliErr.NewCLIError("UNSUPPORTED_OPERATION",
				fmt.Sprintf("%s 不支持异步上传状态查询，上传为同步返回", adplatform.PlatformNames[platform]))
		}

		taskIdsStr, _ := cmd.Flags().GetString("task-ids")
		if taskIdsStr == "" {
			return cliErr.NewCLIError("MISSING_TASK_IDS", "必须指定 --task-ids（任务 ID 列表，逗号分隔）")
		}

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

		accountType, _ := cmd.Flags().GetString("account-type")
		if accountType == "" {
			accountType = "ADVERTISER"
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		// 1. 从后端实时获取平台 Token
		token, err := adplatform.GetAdAuthorizationToken(ctx, appId, "")
		if err != nil {
			return err
		}

		// 2. 构建请求
		req := &adplatform.UploadStatusRequest{
			Platform:    platform,
			AccountId:   0, // 巨量引擎状态查询需要 account_id，暂不传
			AccountType: accountType,
			TaskIds:     taskIds,
		}

		// 3. 获取平台客户端
		client := getPlatformClient(platform, token.AccessToken)

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

// getPlatformClient 获取平台客户端（传入实时获取的 Token）
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

// printAuthListTable 表格输出授权列表
func printAuthListTable(cmd *cobra.Command, auths []adplatform.AdAuthorization) {
	w := cmd.OutOrStdout()

	if len(auths) == 0 {
		fmt.Fprintln(w, "无可用授权")
		return
	}

	fmt.Fprintf(w, "共 %d 条授权\n\n", len(auths))

	t := output.NewTableWriter(w)
	t.AppendHeader("授权ID", "授权用户", "平台AppID", "状态", "过期时间", "授权时间")

	for _, a := range auths {
		statusName := authStatusName(a.AuthStatus)
		t.AppendRow(
			strconv.FormatInt(a.ID, 10),
			a.AuthUserName,
			a.AuthAppId,
			statusName,
			a.ExpirationTime,
			a.AuthTime,
		)
	}

	t.Render()
}

// printAccountListTable 表格输出投放账户列表
func printAccountListTable(cmd *cobra.Command, accounts []adplatform.AdPlatformAccount, total int64, page int, pageSize int) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		total, page, totalPages(total, pageSize))

	if len(accounts) == 0 {
		fmt.Fprintln(w, "无投放账户")
		return
	}

	t := output.NewTableWriter(w)
	t.AppendHeader("ID", "账户ID", "账户名称", "类型", "授权状态", "活跃状态")

	accountTypeNames := map[int]string{
		1: "广告主",
		2: "代理商",
		3: "媒体",
	}

	for _, a := range accounts {
		typeName := accountTypeNames[a.AdAccountType]
		if typeName == "" {
			typeName = strconv.Itoa(a.AdAccountType)
		}
		authStatus := authStatusName(a.AuthStatus)
		active := "活跃"
		if a.Active != 1 {
			active = "不活跃"
		}
		t.AppendRow(
			strconv.FormatInt(a.ID, 10),
			a.AdAccountId,
			a.AdAccountName,
			typeName,
			authStatus,
			active,
		)
	}

	t.Render()
}

// authStatusName 授权状态名称
func authStatusName(status int) string {
	switch status {
	case 1:
		return "有效"
	case 0:
		return "无效"
	case 2:
		return "过期"
	default:
		return strconv.Itoa(status)
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
	deliverCmd.AddCommand(deliverAuthListCmd)
	deliverCmd.AddCommand(deliverAccountListCmd)
	deliverCmd.AddCommand(deliverUploadVideoCmd)
	deliverCmd.AddCommand(deliverUploadImageCmd)
	deliverCmd.AddCommand(deliverUploadStatusCmd)

	// deliverAuthListCmd 参数
	deliverAuthListCmd.Flags().Int64("app-id", 0, "广告平台应用 ID（必填）")
	deliverAuthListCmd.Flags().Int64("team-id", 0, "团队 ID（可选，按团队筛选）")

	// deliverAccountListCmd 参数
	deliverAccountListCmd.Flags().Int64("app-id", 0, "广告平台应用 ID（必填）")
	deliverAccountListCmd.Flags().Int64("authorization-id", 0, "授权记录 ID（可选，按授权筛选）")
	deliverAccountListCmd.Flags().Int("page", 1, "页码")
	deliverAccountListCmd.Flags().Int("pageSize", 20, "每页条数（最大 100）")

	// deliverUploadVideoCmd 参数
	deliverUploadVideoCmd.Flags().Int64("app-id", 0, "广告平台应用 ID（必填）")
	deliverUploadVideoCmd.Flags().String("ad-account-id", "", "投放账户 ID（可选，指定则获取其关联授权的 Token）")
	deliverUploadVideoCmd.Flags().String("platform", "", "广告平台（oceanengine 或 tencentads，必填）")
	deliverUploadVideoCmd.Flags().String("video-url", "", "TOS 视频文件 URL（必填）")
	deliverUploadVideoCmd.Flags().String("filename", "", "文件名（巨量引擎必填）")
	deliverUploadVideoCmd.Flags().String("labels", "", "标签列表（逗号分隔，巨量引擎可选）")
	deliverUploadVideoCmd.Flags().Bool("is-aigc", false, "是否为 AIGC 生成的素材")
	deliverUploadVideoCmd.Flags().String("description", "", "视频描述（腾讯广告可选）")
	deliverUploadVideoCmd.Flags().String("account-type", "ADVERTISER", "账户类型（ADVERTISER 或 AGENT）")

	// deliverUploadImageCmd 参数
	deliverUploadImageCmd.Flags().Int64("app-id", 0, "广告平台应用 ID（必填）")
	deliverUploadImageCmd.Flags().String("ad-account-id", "", "投放账户 ID（可选）")
	deliverUploadImageCmd.Flags().String("platform", "", "广告平台（oceanengine 或 tencentads，必填）")
	deliverUploadImageCmd.Flags().String("image-url", "", "TOS 图片文件 URL（必填）")
	deliverUploadImageCmd.Flags().String("filename", "", "文件名（可选）")
	deliverUploadImageCmd.Flags().Bool("is-aigc", false, "是否为 AIGC 生成的素材")

	// deliverUploadStatusCmd 参数
	deliverUploadStatusCmd.Flags().Int64("app-id", 0, "广告平台应用 ID（必填）")
	deliverUploadStatusCmd.Flags().String("platform", "", "广告平台（目前仅支持 oceanengine）")
	deliverUploadStatusCmd.Flags().String("task-ids", "", "任务 ID 列表（逗号分隔，必填）")
	deliverUploadStatusCmd.Flags().String("account-type", "ADVERTISER", "账户类型（ADVERTISER 或 AGENT）")
}