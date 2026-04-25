package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/output"
)

// repositoryCmd 代表 repository 命令
var repositoryCmd = &cobra.Command{
	Use:     "repository",
	Short:   "素材库管理",
	Long:    `管理素材库，包括查看素材库列表、文件夹、文件查重、文件上传。`,
	Aliases: []string{"repo"},
}

// repositoryListCmd 列出素材库
var repositoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可访问的素材库",
	Long:  `获取权限范围内的素材库列表。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		repoClient := client.NewRepositoryClient()
		repositories, err := repoClient.ListRepositories(ctx)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, repositories)
			return
		}

		switch format {
		case "json":
			outputData(cmd, repositories)
		default:
			printRepositoryTable(cmd, repositories)
		}
	},
}

// repositoryFoldersCmd 列出文件夹
var repositoryFoldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "列出素材库文件夹",
	Long:  `获取素材库的文件夹列表，支持指定父文件夹和统计信息。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		parentFolderID, _ := cmd.Flags().GetInt64("parent-folder-id")
		withStatistic, _ := cmd.Flags().GetBool("with-statistic")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		repoClient := client.NewRepositoryClient()
		folders, err := repoClient.ListFolders(ctx, repositoryID, parentFolderID, withStatistic)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, folders)
			return
		}

		switch format {
		case "json":
			outputData(cmd, folders)
		default:
			printFolderTable(cmd, folders, withStatistic)
		}
	},
}

// repositoryFolderCreateCmd 创建文件夹
var repositoryFolderCreateCmd = &cobra.Command{
	Use:   "folder-create",
	Short: "创建素材库文件夹",
	Long: `创建素材库文件夹，支持创建子文件夹。

示例：
  cbi repository folder-create --repository-id 1 --name "游戏素材"
  cbi repository folder-create --repository-id 1 --name "子文件夹" --parent-id 10
  cbi repository folder-create --repository-id 1 --name "测试" --color "#FF5733"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --name")
			os.Exit(1)
		}

		parentID, _ := cmd.Flags().GetInt64("parent-id")
		color, _ := cmd.Flags().GetString("color")
		icon, _ := cmd.Flags().GetInt("icon")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.CreateFolder(ctx, &client.CreateFolderRequest{
			RepositoryID: repositoryID,
			Name:         name,
			ParentID:     parentID,
			Color:        color,
			Icon:         icon,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, result)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件夹创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件夹 ID: %d\n", result.FolderID)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		if result.ParentID > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  父文件夹 ID: %d\n", result.ParentID)
		}
	},
}

// repositoryTagListCmd 列出标签
var repositoryTagListCmd = &cobra.Command{
	Use:   "tag-list",
	Short: "列出素材库标签",
	Long:  `获取素材库中所有可用标签。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		tags, err := repoClient.ListTags(ctx, repositoryID)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, tags)
			return
		}

		switch format {
		case "json":
			outputData(cmd, tags)
		default:
			printTagListTable(cmd, tags)
		}
	},
}

// repositoryFileTagAddCmd 给文件添加标签
var repositoryFileTagAddCmd = &cobra.Command{
	Use:   "file-tag-add",
	Short: "给素材文件添加标签",
	Long: `批量给素材文件添加标签，标签不存在时自动创建。

示例：
  cbi repository file-tag-add --repository-id 1 --file-ids 10,20,30 --tags "游戏,新素材"
  cbi repository file-tag-add --repository-id 1 --file-ids 100 --tags "优质"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		fileIDsStr, _ := cmd.Flags().GetString("file-ids")
		if fileIDsStr == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --file-ids")
			os.Exit(1)
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --tags")
			os.Exit(1)
		}

		// 解析 file-ids
		fileIDs := []int64{}
		for _, idStr := range strings.Split(fileIDsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "错误: 无效的文件 ID - %s\n", idStr)
				os.Exit(1)
			}
			fileIDs = append(fileIDs, id)
		}

		// 解析 tags
		tagNames := []string{}
		for _, tag := range strings.Split(tagsStr, ",") {
			tagNames = append(tagNames, strings.TrimSpace(tag))
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.AddFileTags(ctx, &client.AddFileTagsRequest{
			RepositoryID: repositoryID,
			FileIDs:      fileIDs,
			TagNames:     tagNames,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, result)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 标签添加成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  处理文件数: %d\n", result.SuccessCount)
		if len(result.CreatedTags) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  新创建标签:")
			for _, tag := range result.CreatedTags {
				fmt.Fprintf(cmd.OutOrStdout(), "    - %s (ID: %d)\n", tag.Name, tag.ID)
			}
		}
	},
}

// repositoryFileFolderAddCmd 将文件添加到文件夹
var repositoryFileFolderAddCmd = &cobra.Command{
	Use:   "file-folder-add",
	Short: "将素材文件添加到文件夹",
	Long: `批量将素材文件添加到多个文件夹，自动去重已存在的关联。

示例：
  cbi repository file-folder-add --repository-id 1 --file-ids 10,20,30 --folder-ids 5,8`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		fileIDsStr, _ := cmd.Flags().GetString("file-ids")
		if fileIDsStr == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --file-ids")
			os.Exit(1)
		}

		folderIDsStr, _ := cmd.Flags().GetString("folder-ids")
		if folderIDsStr == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --folder-ids")
			os.Exit(1)
		}

		// 解析 file-ids
		fileIDs := []int64{}
		for _, idStr := range strings.Split(fileIDsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "错误: 无效的文件 ID - %s\n", idStr)
				os.Exit(1)
			}
			fileIDs = append(fileIDs, id)
		}

		// 解析 folder-ids
		folderIDs := []int64{}
		for _, idStr := range strings.Split(folderIDsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "错误: 无效的文件夹 ID - %s\n", idStr)
				os.Exit(1)
			}
			folderIDs = append(folderIDs, id)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.AddFilesToFolder(ctx, &client.AddFilesToFolderRequest{
			RepositoryID: repositoryID,
			FileIDs:      fileIDs,
			FolderIDs:    folderIDs,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, result)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件已添加到文件夹")
		fmt.Fprintf(cmd.OutOrStdout(), "  新建关联数: %d\n", result.SuccessCount)
		fmt.Fprintf(cmd.OutOrStdout(), "  处理文件数: %d\n", result.AddedFileCount)
		fmt.Fprintf(cmd.OutOrStdout(), "  添加到文件夹数: %d\n", result.AddedFolderCount)
	},
}

// repositoryFileCheckCmd 文件查重
var repositoryFileCheckCmd = &cobra.Command{
	Use:   "file-check",
	Short: "检查文件是否已存在",
	Long:  `通过 MD5 检查文件是否已存在于素材库中。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		filePath, _ := cmd.Flags().GetString("file")
		fileMD5, _ := cmd.Flags().GetString("file-md5")

		if filePath != "" && fileMD5 == "" {
			md5Val, err := calculateFileMD5(filePath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "计算文件 MD5 失败: %s\n", err.Error())
				os.Exit(1)
			}
			fileMD5 = md5Val
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "文件 MD5: %s\n", fileMD5)
			}
		}

		if fileMD5 == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --file 或 --file-md5")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.CheckFile(ctx, repositoryID, fileMD5)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, result)
			return
		}

		if result.Existed {
			fmt.Fprintln(cmd.OutOrStdout(), "文件已存在 (重复)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "文件不存在 (可上传)")
		}
		if filePath != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  文件: %s\n", filePath)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  MD5: %s\n", fileMD5)
	},
}

// repositoryFileCreateCmd 创建文件（上传）
var repositoryFileCreateCmd = &cobra.Command{
	Use:   "file-create",
	Short: "上传文件到素材库",
	Long: `上传文件到素材库。

流程：
1. 检查文件大小（限制 100MB）
2. 计算 MD5 进行查重（除非 --skip-check）
3. 如果重复，提示用户（除非 --force）
4. 上传文件

示例：
  cbi repository file-create --repository-id 1 --file ./image.png
  cbi repository file-create --repository-id 1 --file ./video.mp4 --folder-id 123`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --file")
			os.Exit(1)
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "错误: 文件不存在 - %s\n", filePath)
			os.Exit(1)
		}

		if fileInfo.Size() > 100*1024*1024 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 文件大小超过 100MB 限制")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\n取消上传...")
			cancel()
		}()

		repoClient := client.NewRepositoryClient()

		req := &client.CreateFileRequest{
			RepositoryID: repositoryID,
			FilePath:     filePath,
		}

		folderID, _ := cmd.Flags().GetInt64("folder-id")
		if folderID > 0 {
			req.FolderID = folderID
		}

		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req.Name = name
		}

		note, _ := cmd.Flags().GetString("note")
		if note != "" {
			req.Note = note
		}

		rating, _ := cmd.Flags().GetInt("rating")
		if rating > 0 && rating <= 5 {
			req.Rating = rating
		}

		sourceURL, _ := cmd.Flags().GetString("source-url")
		if sourceURL != "" {
			req.SourceURL = sourceURL
		}

		tags, _ := cmd.Flags().GetString("tags")
		if tags != "" {
			req.Tags = tags
		}

		if !quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "正在上传: %s\n", filePath)
		}

		fileInfoResult, err := repoClient.CreateFile(ctx, req)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, fileInfoResult)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件上传成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", fileInfoResult.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  文件名: %s\n", fileInfoResult.Name)
	},
}

// repositoryFileListCmd 列出素材文件
var repositoryFileListCmd = &cobra.Command{
	Use:   "file-list",
	Short: "查询素材文件列表",
	Long: `获取素材库中的文件列表，支持多种筛选模式。

筛选模式（oneof，不可组合）：
  --folder-id    按文件夹筛选
  --tag-id       按标签筛选
  --keyword      按关键词搜索（名称 + signals）
  --has-signals  筛选有/无视频理解信号

示例：
  cbi repository file-list --repository-id 1
  cbi repository file-list --repository-id 1 --folder-id 10
  cbi repository file-list --repository-id 1 --has-signals
  cbi repository file-list --repository-id 1 --has-signals=false`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		folderID, _ := cmd.Flags().GetInt64("folder-id")
		tagID, _ := cmd.Flags().GetInt64("tag-id")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		// has-signals 参数
		var hasSignals *bool
		hasSignalsFlag, _ := cmd.Flags().GetString("has-signals")
		if hasSignalsFlag != "" {
			val := hasSignalsFlag == "true" || hasSignalsFlag == "1"
			hasSignals = &val
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.ListFiles(ctx, &client.ListFilesRequest{
			RepositoryID: repositoryID,
			FolderID:     folderID,
			TagID:        tagID,
			Keyword:      keyword,
			HasSignals:   hasSignals,
			Page:         page,
			PageSize:     pageSize,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet {
			outputData(cmd, result)
			return
		}

		switch format {
		case "json":
			outputData(cmd, result)
		default:
			printFileListTable(cmd, result)
		}
	},
}

// repositoryFileDetailCmd 获取文件详情
var repositoryFileDetailCmd = &cobra.Command{
	Use:   "file-detail <file-id>",
	Short: "获取素材文件详情",
	Long: `获取素材文件的完整详情信息。

示例：
  cbi repository file-detail 123
  cbi repository file-detail 123 --format json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		fileID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "错误: 无效的文件 ID - %s\n", args[0])
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		detail, err := repoClient.GetFileDetail(ctx, fileID)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		if quiet || format == "json" {
			outputData(cmd, detail)
			return
		}

		printFileDetail(cmd, detail)
	},
}

func init() {
	rootCmd.AddCommand(repositoryCmd)
	repositoryCmd.AddCommand(repositoryListCmd)
	repositoryCmd.AddCommand(repositoryFoldersCmd)
	repositoryCmd.AddCommand(repositoryFolderCreateCmd)
	repositoryCmd.AddCommand(repositoryTagListCmd)
	repositoryCmd.AddCommand(repositoryFileCheckCmd)
	repositoryCmd.AddCommand(repositoryFileCreateCmd)
	repositoryCmd.AddCommand(repositoryFileListCmd)
	repositoryCmd.AddCommand(repositoryFileDetailCmd)
	repositoryCmd.AddCommand(repositoryFileTagAddCmd)
	repositoryCmd.AddCommand(repositoryFileFolderAddCmd)

	// folders 命令参数
	repositoryFoldersCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFoldersCmd.Flags().Int64("parent-folder-id", 0, "父文件夹 ID（0 表示根目录）")
	repositoryFoldersCmd.Flags().Bool("with-statistic", false, "包含统计信息（文件数量）")

	// folder-create 命令参数
	repositoryFolderCreateCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFolderCreateCmd.Flags().String("name", "", "文件夹名称（必填）")
	repositoryFolderCreateCmd.Flags().Int64("parent-id", 0, "父文件夹 ID（0=顶级文件夹）")
	repositoryFolderCreateCmd.Flags().String("color", "", "颜色（如 #FF5733）")
	repositoryFolderCreateCmd.Flags().Int("icon", 0, "图标编号")

	// tag-list 命令参数
	repositoryTagListCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")

	// file-tag-add 命令参数
	repositoryFileTagAddCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileTagAddCmd.Flags().String("file-ids", "", "文件 ID 列表（逗号分隔）")
	repositoryFileTagAddCmd.Flags().String("tags", "", "标签名称列表（逗号分隔）")

	// file-folder-add 命令参数
	repositoryFileFolderAddCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileFolderAddCmd.Flags().String("file-ids", "", "文件 ID 列表（逗号分隔）")
	repositoryFileFolderAddCmd.Flags().String("folder-ids", "", "文件夹 ID 列表（逗号分隔）")

	// file-check 命令参数
	repositoryFileCheckCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileCheckCmd.Flags().String("file", "", "本地文件路径（用于计算 MD5）")
	repositoryFileCheckCmd.Flags().String("file-md5", "", "文件 MD5 值（若提供 --file 则自动计算）")

	// file-create 命令参数
	repositoryFileCreateCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileCreateCmd.Flags().String("file", "", "本地文件路径（必填）")
	repositoryFileCreateCmd.Flags().Int64("folder-id", 0, "目标文件夹 ID")
	repositoryFileCreateCmd.Flags().String("name", "", "文件名（默认使用原文件名）")
	repositoryFileCreateCmd.Flags().String("note", "", "备注")
	repositoryFileCreateCmd.Flags().Int("rating", 0, "评分（1-5）")
	repositoryFileCreateCmd.Flags().String("source-url", "", "来源 URL")
	repositoryFileCreateCmd.Flags().String("tags", "", "标签（逗号分隔）")

	// file-list 命令参数
	repositoryFileListCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileListCmd.Flags().Int64("folder-id", 0, "文件夹 ID（文件夹筛选模式）")
	repositoryFileListCmd.Flags().Int64("tag-id", 0, "标签 ID（标签筛选模式）")
	repositoryFileListCmd.Flags().String("keyword", "", "搜索关键词（搜索名称+signals）")
	repositoryFileListCmd.Flags().String("has-signals", "", "筛选有无信号（true/false）")
	repositoryFileListCmd.Flags().Int("page", 1, "页码（默认 1）")
	repositoryFileListCmd.Flags().Int("pageSize", 20, "每页条数（默认 20，最大 50）")
}

// calculateFileMD5 计算文件 MD5
func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// printRepositoryTable 打印素材库表格
func printRepositoryTable(cmd *cobra.Command, repositories []client.Repository) {
	if len(repositories) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "无素材库")
		return
	}

	t := output.NewTableWriter(cmd.OutOrStdout())
	t.AppendHeader("ID", "名称", "描述", "默认", "权限")

	for _, repo := range repositories {
		isDefault := ""
		if repo.IsDefault {
			isDefault = "✓"
		}
		t.AppendRow(fmt.Sprintf("%d", repo.ID), repo.Name, repo.Desc, isDefault, repo.Perm)
	}

	t.Render()
}

// printFolderTable 打印文件夹表格
func printFolderTable(cmd *cobra.Command, folders []client.Folder, withStatistic bool) {
	if len(folders) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "无文件夹")
		return
	}

	t := output.NewTableWriter(cmd.OutOrStdout())
	if withStatistic {
		t.AppendHeader("ID", "名称", "文件数")
	} else {
		t.AppendHeader("ID", "名称")
	}

	for _, folder := range folders {
		if withStatistic && folder.Statistic != nil {
			t.AppendRow(fmt.Sprintf("%d", folder.ID), folder.Name, fmt.Sprintf("%d", folder.Statistic.FileCount))
		} else {
			t.AppendRow(fmt.Sprintf("%d", folder.ID), folder.Name)
		}
	}

	t.Render()
}

// printTagListTable 打印标签列表表格
func printTagListTable(cmd *cobra.Command, tags []client.Tag) {
	if len(tags) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "无标签")
		return
	}

	t := output.NewTableWriter(cmd.OutOrStdout())
	t.AppendHeader("ID", "名称", "使用次数")

	for _, tag := range tags {
		t.AppendRow(fmt.Sprintf("%d", tag.ID), tag.Name, fmt.Sprintf("%d", tag.RefCnt))
	}

	t.Render()
}

// printFileDetail 打印文件详情
func printFileDetail(cmd *cobra.Command, detail *client.FileDetail) {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "文件详情:")
	fmt.Fprintf(out, "  ID:           %d\n", detail.ID)
	fmt.Fprintf(out, "  名称:         %s\n", detail.Name)
	fmt.Fprintf(out, "  格式:         %s\n", detail.Format)
	fmt.Fprintf(out, "  大小:         %s (%d bytes)\n", detail.Size, detail.SizeInByte)

	if detail.Duration != "" {
		fmt.Fprintf(out, "  时长:         %s\n", detail.Duration)
	}
	if detail.Resolution != "" {
		fmt.Fprintf(out, "  分辨率:       %s\n", detail.Resolution)
	}
	if detail.Ratio != "" {
		fmt.Fprintf(out, "  比例:         %s\n", detail.Ratio)
	}
	if detail.FrameRate != "" {
		fmt.Fprintf(out, "  帧率:         %s\n", detail.FrameRate)
	}
	if detail.Score > 0 {
		fmt.Fprintf(out, "  评分:         %d\n", detail.Score)
	}
	if detail.Notes != "" {
		fmt.Fprintf(out, "  备注:         %s\n", detail.Notes)
	}
	if detail.SourcePlatform != "" {
		fmt.Fprintf(out, "  来源平台:     %s\n", detail.SourcePlatform)
	}
	if detail.FileSourceUrl != "" {
		fmt.Fprintf(out, "  来源 URL:     %s\n", detail.FileSourceUrl)
	}
	if detail.Hash != "" {
		fmt.Fprintf(out, "  MD5:          %s\n", detail.Hash)
	}

	if detail.CreatedAt > 0 {
		fmt.Fprintf(out, "  创建时间:     %s\n", time.Unix(detail.CreatedAt, 0).Format("2006-01-02 15:04:05"))
	}
	if detail.UpdatedAt > 0 {
		fmt.Fprintf(out, "  更新时间:     %s\n", time.Unix(detail.UpdatedAt, 0).Format("2006-01-02 15:04:05"))
	}

	if detail.Cover != "" {
		fmt.Fprintf(out, "  封面 URL:     %s\n", detail.Cover)
	}
	if detail.FileOriginUrl != "" {
		fmt.Fprintf(out, "  原始文件 URL: %s\n", detail.FileOriginUrl)
	}
	if detail.FileViewUrl != "" {
		fmt.Fprintf(out, "  预览 URL:     %s\n", detail.FileViewUrl)
	}

	if len(detail.Products) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "关联产品:")
		for _, p := range detail.Products {
			fmt.Fprintf(out, "  - %s (ID: %d)\n", p.Name, p.ID)
		}
	}

	if len(detail.Tags) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "标签:")
		for _, t := range detail.Tags {
			fmt.Fprintf(out, "  - %s (ID: %d)\n", t.Name, t.ID)
		}
	}

	if len(detail.Folders) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "所在文件夹:")
		for _, f := range detail.Folders {
			fmt.Fprintf(out, "  - %s (ID: %d)\n", f.Name, f.ID)
		}
	}

	if detail.Creator != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "创建者:")
		fmt.Fprintf(out, "  姓名:   %s\n", detail.Creator.Name)
		fmt.Fprintf(out, "  邮箱:   %s\n", detail.Creator.Email)
		fmt.Fprintf(out, "  ID:     %d\n", detail.Creator.ID)
	}

	if len(detail.Signals) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "视频理解信号:")
		for _, s := range detail.Signals {
			fmt.Fprintf(out, "  [%s] %s\n", s.SignalID, s.SignalName)
			if len(s.SignalTags) > 0 {
				fmt.Fprintf(out, "    标签: %s\n", strings.Join(s.SignalTags, ", "))
			}
			if s.SignalContent != "" {
				fmt.Fprintf(out, "    内容:\n")
				lines := strings.Split(s.SignalContent, "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Fprintf(out, "      %s\n", line)
					}
				}
			}
		}
	}
}

// printFileListTable 打印文件列表表格
func printFileListTable(cmd *cobra.Command, result *client.FileListResult) {
	out := cmd.OutOrStdout()

	if len(result.Files) == 0 {
		fmt.Fprintln(out, "无文件")
		return
	}

	totalPages := result.Total / int64(result.PageSize)
	if result.Total % int64(result.PageSize) > 0 {
		totalPages++
	}
	fmt.Fprintf(out, "共 %d 条，第 %d/%d 页\n\n", result.Total, result.Page, totalPages)

	t := output.NewTableWriter(out)
	t.AppendHeader("ID", "名称", "评分", "标签", "创建时间")

	for _, file := range result.Files {
		tagNames := ""
		if len(file.Tags) > 0 {
			for i, tag := range file.Tags {
				if i > 0 {
					tagNames += ", "
				}
				tagNames += tag.Name
			}
		}

		createTime := ""
		if file.CreatedAt > 0 {
			createTime = time.Unix(file.CreatedAt, 0).Format("2006-01-02 15:04")
		}

		score := ""
		if file.Score > 0 {
			score = fmt.Sprintf("%d", file.Score)
		}

		t.AppendRow(fmt.Sprintf("%d", file.ID), file.Name, score, tagNames, createTime)
	}

	t.Render()
}