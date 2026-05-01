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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// requireRepositoryID 获取并验证 repository-id 参数
func requireRepositoryID(cmd *cobra.Command) (int64, error) {
	id, _ := cmd.Flags().GetInt64("repository-id")
	if id == 0 {
		return 0, cliErr.NewCLIError("MISSING_REPOSITORY_ID", "必须指定 --repository-id")
	}
	return id, nil
}

// requireFileID 获取并验证 file-id 参数
func requireFileID(cmd *cobra.Command) (int64, error) {
	id, _ := cmd.Flags().GetInt64("file-id")
	if id == 0 {
		return 0, cliErr.NewCLIError("MISSING_FILE_ID", "必须指定 --file-id")
	}
	return id, nil
}

// parseIDList 解析逗号分隔的 ID 列表
func parseIDList(s, label string) ([]int64, error) {
	if s == "" {
		return nil, cliErr.NewCLIError("MISSING_IDS", fmt.Sprintf("必须指定 --%s", label))
	}
	var ids []int64
	for _, idStr := range strings.Split(s, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			return nil, cliErr.NewCLIError("INVALID_ID", fmt.Sprintf("无效的 %s - %s", label, idStr))
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// newSignalCtx 创建支持信号取消的 context
func newSignalCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()
	return ctx, cancel
}

// repositoryListCmd 列出素材库
var repositoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可访问的素材库",
	Long:  `获取权限范围内的素材库列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		repositories, err := repoClient.ListRepositories(ctx)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, repositories)
		}

		switch format {
		case "json":
			return outputData(cmd, repositories)
		default:
			printRepositoryTable(cmd, repositories)
			return nil
		}
	},
}

// repositoryFoldersCmd 列出文件夹
var repositoryFoldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "列出素材库文件夹",
	Long:  `获取素材库的文件夹列表，支持指定父文件夹和统计信息。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		parentFolderID, _ := cmd.Flags().GetInt64("parent-folder-id")
		withStatistic, _ := cmd.Flags().GetBool("with-statistic")

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		folders, err := repoClient.ListFolders(ctx, repositoryID, parentFolderID, withStatistic)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, folders)
		}

		switch format {
		case "json":
			return outputData(cmd, folders)
		default:
			printFolderTable(cmd, folders, withStatistic)
			return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
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
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件夹创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件夹 ID: %d\n", result.FolderID)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		if result.ParentID > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  父文件夹 ID: %d\n", result.ParentID)
		}
		return nil
	},
}

// repositoryTagListCmd 列出标签
var repositoryTagListCmd = &cobra.Command{
	Use:   "tag-list",
	Short: "列出素材库标签",
	Long:  `获取素材库中所有可用标签。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		tags, err := repoClient.ListTags(ctx, repositoryID)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, tags)
		}

		switch format {
		case "json":
			return outputData(cmd, tags)
		default:
			printTagListTable(cmd, tags)
			return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileIDsStr, _ := cmd.Flags().GetString("file-ids")
		fileIDs, err := parseIDList(fileIDsStr, "file-ids")
		if err != nil {
			return err
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr == "" {
			return cliErr.NewCLIError("MISSING_TAGS", "必须指定 --tags")
		}
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
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 标签添加成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  处理文件数: %d\n", result.SuccessCount)
		if len(result.CreatedTags) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  新创建标签:")
			for _, tag := range result.CreatedTags {
				fmt.Fprintf(cmd.OutOrStdout(), "    - %s (ID: %d)\n", tag.Name, tag.ID)
			}
		}
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileIDsStr, _ := cmd.Flags().GetString("file-ids")
		fileIDs, err := parseIDList(fileIDsStr, "file-ids")
		if err != nil {
			return err
		}

		folderIDsStr, _ := cmd.Flags().GetString("folder-ids")
		folderIDs, err := parseIDList(folderIDsStr, "folder-ids")
		if err != nil {
			return err
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
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件已添加到文件夹")
		fmt.Fprintf(cmd.OutOrStdout(), "  新建关联数: %d\n", result.SuccessCount)
		fmt.Fprintf(cmd.OutOrStdout(), "  处理文件数: %d\n", result.AddedFileCount)
		fmt.Fprintf(cmd.OutOrStdout(), "  添加到文件夹数: %d\n", result.AddedFolderCount)
		return nil
	},
}

// repositoryFileCheckCmd 文件查重
var repositoryFileCheckCmd = &cobra.Command{
	Use:   "file-check",
	Short: "检查文件是否已存在",
	Long:  `通过 MD5 检查文件是否已存在于素材库中。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		filePath, _ := cmd.Flags().GetString("file")
		fileMD5, _ := cmd.Flags().GetString("file-md5")

		if filePath != "" && fileMD5 == "" {
			md5Val, err := calculateFileMD5(filePath)
			if err != nil {
				return cliErr.NewCLIErrorWithDetail("MD5_FAILED", "计算文件 MD5 失败", err.Error())
			}
			fileMD5 = md5Val
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "文件 MD5: %s\n", fileMD5)
			}
		}

		if fileMD5 == "" {
			return cliErr.NewCLIError("MISSING_FILE", "必须指定 --file 或 --file-md5")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.CheckFile(ctx, repositoryID, fileMD5)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
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
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return cliErr.NewCLIError("MISSING_FILE", "必须指定 --file")
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return cliErr.NewCLIError("FILE_NOT_FOUND", fmt.Sprintf("文件不存在 - %s", filePath))
		}

		if fileInfo.Size() > 100*1024*1024 {
			return cliErr.NewCLIError("FILE_TOO_LARGE", "文件大小超过 100MB 限制")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

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
			return err
		}

		if quiet {
			return outputData(cmd, fileInfoResult)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件上传成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", fileInfoResult.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  文件名: %s\n", fileInfoResult.Name)
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
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

		ctx, cancel := newSignalCtx()
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
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		switch format {
		case "json":
			return outputData(cmd, result)
		default:
			printFileListTable(cmd, result)
			return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		fileID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return cliErr.NewCLIError("INVALID_FILE_ID", fmt.Sprintf("无效的文件 ID - %s", args[0]))
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		detail, err := repoClient.GetFileDetail(ctx, fileID)
		if err != nil {
			return err
		}

		if quiet || format == "json" {
			return outputData(cmd, detail)
		}

		printFileDetail(cmd, detail)
		return nil
	},
}

// repositoryFileNameUpdateCmd 更新文件名称
var repositoryFileNameUpdateCmd = &cobra.Command{
	Use:   "file-name-update",
	Short: "更新文件名称",
	Long: `更新素材文件的名称。

示例：
  cbi repository file-name-update --repository-id 1 --file-id 123 --name "新名称"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.UpdateFileName(ctx, &client.UpdateFileNameRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			Name:         name,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件名称更新成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", result.FileID)
		fmt.Fprintf(cmd.OutOrStdout(), "  新名称: %s\n", result.Name)
		return nil
	},
}

// repositoryFileNotesUpdateCmd 更新文件备注
var repositoryFileNotesUpdateCmd = &cobra.Command{
	Use:   "file-notes-update",
	Short: "更新文件备注",
	Long: `更新素材文件的备注，空字符串表示清空备注。

示例：
  cbi repository file-notes-update --repository-id 1 --file-id 123 --notes "这是备注"
  cbi repository file-notes-update --repository-id 1 --file-id 123 --notes ""  # 清空备注`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		notes, _ := cmd.Flags().GetString("notes")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.UpdateFileNotes(ctx, &client.UpdateFileNotesRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			Notes:        notes,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件备注更新成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", result.FileID)
		if result.Notes != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  备注: %s\n", result.Notes)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  备注: (已清空)")
		}
		return nil
	},
}

// repositoryFileScoreUpdateCmd 更新文件评分
var repositoryFileScoreUpdateCmd = &cobra.Command{
	Use:   "file-score-update",
	Short: "更新文件评分",
	Long: `更新素材文件的评分（1-5）。

示例：
  cbi repository file-score-update --repository-id 1 --file-id 123 --score 4`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		score, _ := cmd.Flags().GetInt("score")
		if score < 1 || score > 5 {
			return cliErr.NewCLIError("INVALID_SCORE", "评分必须在 1-5 范围内")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.UpdateFileScore(ctx, &client.UpdateFileScoreRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			Score:        score,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件评分更新成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", result.FileID)
		fmt.Fprintf(cmd.OutOrStdout(), "  评分: %d\n", result.Score)
		return nil
	},
}

// repositoryProductListCmd 查询产品列表
var repositoryProductListCmd = &cobra.Command{
	Use:   "product-list",
	Short: "查询档案库产品列表",
	Long: `获取档案库中所有已关联的产品列表。

示例：
  cbi repository product-list --repository-id 1
  cbi repository product-list --repository-id 1 --format json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		products, err := repoClient.ListProducts(ctx, repositoryID)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, products)
		}

		switch format {
		case "json":
			return outputData(cmd, products)
		default:
			printProductListTable(cmd, products)
			return nil
		}
	},
}

// repositoryFileProductAddCmd 给文件添加关联产品
var repositoryFileProductAddCmd = &cobra.Command{
	Use:   "file-product-add",
	Short: "给文件添加关联产品",
	Long: `给素材文件添加关联产品，支持批量添加。

产品类型：
  1 = 应用
  2 = 游戏（默认）
  3 = 商品

业务逻辑：
  - 根据产品名称在档案库中查找
  - 存在同名产品 → 直接关联，isNew=false
  - 不存在 → 创建新产品后关联，isNew=true
  - 已关联的产品不会重复关联

示例：
  cbi repository file-product-add --repository-id 1 --file-id 123 --products "产品A,产品B"
  cbi repository file-product-add --repository-id 1 --file-id 123 --products "产品A" --product-type 2`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		productsStr, _ := cmd.Flags().GetString("products")
		if productsStr == "" {
			return cliErr.NewCLIError("MISSING_PRODUCTS", "必须指定 --products")
		}

		productType, _ := cmd.Flags().GetInt("product-type")
		productURL, _ := cmd.Flags().GetString("product-url")
		productImg, _ := cmd.Flags().GetString("product-img")
		productDesc, _ := cmd.Flags().GetString("product-desc")

		// 解析产品名称列表
		productInputs := []client.ProductInput{}
		for _, name := range strings.Split(productsStr, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				p := client.ProductInput{
					Name: name,
					Type: productType,
				}
				if productURL != "" {
					p.URL = productURL
				}
				if productImg != "" {
					p.Img = productImg
				}
				if productDesc != "" {
					p.Description = productDesc
				}
				productInputs = append(productInputs, p)
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.AddFileProducts(ctx, &client.AddFileProductsRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			Products:     productInputs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 产品关联添加成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  成功关联数: %d\n", result.SuccessCount)
		if len(result.Products) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  产品列表:")
			for _, p := range result.Products {
				newStr := ""
				if p.IsNew {
					newStr = " (新创建)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "    - %s (ID: %d)%s\n", p.Name, p.ID, newStr)
			}
		}
		return nil
	},
}

// repositoryFileTagRemoveCmd 移除文件标签
var repositoryFileTagRemoveCmd = &cobra.Command{
	Use:   "file-tag-remove",
	Short: "移除文件的标签",
	Long: `移除素材文件的标签。

示例：
  cbi repository file-tag-remove --repository-id 1 --file-id 123 --tag-ids 5,10`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		tagIDsStr, _ := cmd.Flags().GetString("tag-ids")
		tagIDs, err := parseIDList(tagIDsStr, "tag-ids")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.RemoveFileTags(ctx, &client.RemoveFileTagsRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			TagIDs:       tagIDs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 标签移除成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  移除数量: %d\n", result.SuccessCount)
		return nil
	},
}

// repositoryFileProductRemoveCmd 移除文件关联产品
var repositoryFileProductRemoveCmd = &cobra.Command{
	Use:   "file-product-remove",
	Short: "移除文件的关联产品",
	Long: `移除素材文件的关联产品。

示例：
  cbi repository file-product-remove --repository-id 1 --file-id 123 --product-ids 10,15`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileID, err := requireFileID(cmd)
		if err != nil {
			return err
		}

		productIDsStr, _ := cmd.Flags().GetString("product-ids")
		productIDs, err := parseIDList(productIDsStr, "product-ids")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.RemoveFileProducts(ctx, &client.RemoveFileProductsRequest{
			RepositoryID: repositoryID,
			FileID:       fileID,
			ProductIDs:   productIDs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 关联产品移除成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  移除数量: %d\n", result.SuccessCount)
		return nil
	},
}

// repositoryProductDeleteCmd 删除产品
var repositoryProductDeleteCmd = &cobra.Command{
	Use:   "product-delete",
	Short: "删除产品",
	Long: `删除档案库中的产品。

示例：
  cbi repository product-delete --repository-id 1 --product-ids 10,15,20`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		productIDsStr, _ := cmd.Flags().GetString("product-ids")
		productIDs, err := parseIDList(productIDsStr, "product-ids")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.DeleteProducts(ctx, &client.DeleteProductsRequest{
			RepositoryID: repositoryID,
			ProductIDs:   productIDs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 产品删除成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  删除数量: %d\n", result.SuccessCount)
		return nil
	},
}

// repositoryFileDeleteCmd 删除文件到回收站
var repositoryFileDeleteCmd = &cobra.Command{
	Use:   "file-delete",
	Short: "删除文件到回收站",
	Long: `将素材文件移入回收站。

示例：
  cbi repository file-delete --repository-id 1 --file-ids 123,124,125`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		fileIDsStr, _ := cmd.Flags().GetString("file-ids")
		fileIDs, err := parseIDList(fileIDsStr, "file-ids")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.DeleteFiles(ctx, &client.DeleteFilesRequest{
			RepositoryID: repositoryID,
			FileIDs:      fileIDs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件已移入回收站")
		fmt.Fprintf(cmd.OutOrStdout(), "  删除数量: %d\n", result.SuccessCount)
		return nil
	},
}

// repositoryTagDeleteCmd 删除档案库标签
var repositoryTagDeleteCmd = &cobra.Command{
	Use:   "tag-delete",
	Short: "删除档案库标签",
	Long: `删除档案库中的标签（软删除）。

删除后标签不再可用，但已关联的文件标签记录保留。
需要档案库编辑权限。

示例：
  cbi repository tag-delete --repository-id 1 --tag-ids 5,10,15`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		tagIDsStr, _ := cmd.Flags().GetString("tag-ids")
		tagIDs, err := parseIDList(tagIDsStr, "tag-ids")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.DeleteTags(ctx, &client.DeleteTagsRequest{
			RepositoryID: repositoryID,
			TagIDs:       tagIDs,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 标签删除成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  删除数量: %d\n", result.SuccessCount)
		return nil
	},
}

// repositoryHighlightClipListCmd 爆点片段列表
var repositoryHighlightClipListCmd = &cobra.Command{
	Use:   "highlight-clip-list",
	Short: "获取爆点片段列表",
	Long: `获取素材库中的爆点片段列表。

筛选参数：
  --keyword       搜索关键词（匹配爆点片段名称）
  --source-video-id 筛选指定来源视频的爆点片段

示例：
  cbi repository highlight-clip-list --repository-id 1
  cbi repository highlight-clip-list --repository-id 1 --keyword "高光"
  cbi repository highlight-clip-list --repository-id 1 --source-video-id 456`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryID, err := requireRepositoryID(cmd)
		if err != nil {
			return err
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		sourceVideoID, _ := cmd.Flags().GetInt64("source-video-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		repoClient := client.NewRepositoryClient()
		result, err := repoClient.ListHighlightClips(ctx, &client.HighlightClipListRequest{
			RepositoryID:  repositoryID,
			Keyword:       keyword,
			SourceVideoID: sourceVideoID,
			Page:          page,
			PageSize:      pageSize,
		})
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
			printHighlightClipListTable(cmd, result)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(repositoryCmd)
	repositoryCmd.AddCommand(repositoryListCmd)
	repositoryCmd.AddCommand(repositoryFoldersCmd)
	repositoryCmd.AddCommand(repositoryFolderCreateCmd)
	repositoryCmd.AddCommand(repositoryTagListCmd)
	repositoryCmd.AddCommand(repositoryProductListCmd)
	repositoryCmd.AddCommand(repositoryFileCheckCmd)
	repositoryCmd.AddCommand(repositoryFileCreateCmd)
	repositoryCmd.AddCommand(repositoryFileListCmd)
	repositoryCmd.AddCommand(repositoryFileDetailCmd)
	repositoryCmd.AddCommand(repositoryFileTagAddCmd)
	repositoryCmd.AddCommand(repositoryFileFolderAddCmd)
	repositoryCmd.AddCommand(repositoryFileNameUpdateCmd)
	repositoryCmd.AddCommand(repositoryFileNotesUpdateCmd)
	repositoryCmd.AddCommand(repositoryFileScoreUpdateCmd)
	repositoryCmd.AddCommand(repositoryFileProductAddCmd)
	repositoryCmd.AddCommand(repositoryFileTagRemoveCmd)
	repositoryCmd.AddCommand(repositoryFileProductRemoveCmd)
	repositoryCmd.AddCommand(repositoryProductDeleteCmd)
	repositoryCmd.AddCommand(repositoryFileDeleteCmd)
	repositoryCmd.AddCommand(repositoryTagDeleteCmd)
	repositoryCmd.AddCommand(repositoryHighlightClipListCmd)

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

	// product-list 命令参数
	repositoryProductListCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")

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

	// file-name-update 命令参数
	repositoryFileNameUpdateCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileNameUpdateCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileNameUpdateCmd.Flags().String("name", "", "新文件名称（必填）")

	// file-notes-update 命令参数
	repositoryFileNotesUpdateCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileNotesUpdateCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileNotesUpdateCmd.Flags().String("notes", "", "备注内容（空字符串表示清空）")

	// file-score-update 命令参数
	repositoryFileScoreUpdateCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileScoreUpdateCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileScoreUpdateCmd.Flags().Int("score", 0, "评分（1-5，必填）")

	// file-product-add 命令参数
	repositoryFileProductAddCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileProductAddCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileProductAddCmd.Flags().String("products", "", "产品名称列表（逗号分隔，必填）")
	repositoryFileProductAddCmd.Flags().Int("product-type", 2, "产品类型（1=应用，2=游戏，3=商品）")
	repositoryFileProductAddCmd.Flags().String("product-url", "", "产品 URL（可选，应用到所有产品）")
	repositoryFileProductAddCmd.Flags().String("product-img", "", "产品图片 URL（可选，应用到所有产品）")
	repositoryFileProductAddCmd.Flags().String("product-desc", "", "产品描述（可选，应用到所有产品）")

	// file-tag-remove 命令参数
	repositoryFileTagRemoveCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileTagRemoveCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileTagRemoveCmd.Flags().String("tag-ids", "", "标签 ID 列表（逗号分隔，必填）")

	// file-product-remove 命令参数
	repositoryFileProductRemoveCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileProductRemoveCmd.Flags().Int64("file-id", 0, "文件 ID（必填）")
	repositoryFileProductRemoveCmd.Flags().String("product-ids", "", "产品 ID 列表（逗号分隔，必填）")

	// product-delete 命令参数
	repositoryProductDeleteCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryProductDeleteCmd.Flags().String("product-ids", "", "产品 ID 列表（逗号分隔，必填）")

	// file-delete 命令参数
	repositoryFileDeleteCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFileDeleteCmd.Flags().String("file-ids", "", "文件 ID 列表（逗号分隔，必填）")

	// tag-delete 命令参数
	repositoryTagDeleteCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryTagDeleteCmd.Flags().String("tag-ids", "", "标签 ID 列表（逗号分隔，必填）")

	// highlight-clip-list 命令参数
	repositoryHighlightClipListCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryHighlightClipListCmd.Flags().String("keyword", "", "搜索关键词（匹配名称）")
	repositoryHighlightClipListCmd.Flags().Int64("source-video-id", 0, "来源视频 ID")
	repositoryHighlightClipListCmd.Flags().Int("page", 1, "页码（默认 1）")
	repositoryHighlightClipListCmd.Flags().Int("pageSize", 20, "每页条数（默认 20，最大 50）")
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
	if result.Total%int64(result.PageSize) > 0 {
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

// printProductListTable 打印产品列表表格
func printProductListTable(cmd *cobra.Command, products []client.Product) {
	if len(products) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "无产品")
		return
	}

	t := output.NewTableWriter(cmd.OutOrStdout())
	t.AppendHeader("ID", "名称", "类型", "URL", "描述")

	for _, p := range products {
		typeStr := ""
		switch p.Type {
		case 1:
			typeStr = "应用"
		case 2:
			typeStr = "游戏"
		case 3:
			typeStr = "商品"
		}
		t.AppendRow(fmt.Sprintf("%d", p.ID), p.Name, typeStr, p.URL, p.Description)
	}

	t.Render()
}

// printHighlightClipListTable 打印爆点片段列表表格
func printHighlightClipListTable(cmd *cobra.Command, result *client.HighlightClipListResult) {
	out := cmd.OutOrStdout()

	if len(result.Clips) == 0 {
		fmt.Fprintln(out, "无爆点片段")
		return
	}

	totalPages := result.Total / int64(result.PageSize)
	if result.Total%int64(result.PageSize) > 0 {
		totalPages++
	}
	fmt.Fprintf(out, "共 %d 条，第 %d/%d 页\n\n", result.Total, result.Page, totalPages)

	t := output.NewTableWriter(out)
	t.AppendHeader("ID", "名称", "时长", "来源视频", "创建时间")

	for _, clip := range result.Clips {
		sourceVideoName := ""
		if clip.SourceVideo != nil {
			sourceVideoName = clip.SourceVideo.Name
		}

		createTime := ""
		if clip.CreatedAt > 0 {
			createTime = time.Unix(clip.CreatedAt, 0).Format("2006-01-02 15:04")
		}

		t.AppendRow(fmt.Sprintf("%d", clip.ID), clip.Name, clip.Duration, sourceVideoName, createTime)
	}

	t.Render()
}
