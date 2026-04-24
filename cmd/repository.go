package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

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
		// 检查认证
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		// 创建上下文，支持 Ctrl+C 取消
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// 调用 API
		repoClient := client.NewRepositoryClient()
		repositories, err := repoClient.ListRepositories(ctx)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// 输出结果
		if quiet {
			// 静默模式，只输出数据
			outputData(cmd, repositories)
			return
		}

		// 根据格式输出
		switch format {
		case "json":
			outputData(cmd, repositories)
		case "table":
			printRepositoryTable(cmd, repositories)
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
		// 检查认证
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		// 获取参数
		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		parentFolderID, _ := cmd.Flags().GetInt64("parent-folder-id")
		withStatistic, _ := cmd.Flags().GetBool("with-statistic")

		// 创建上下文
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// 调用 API
		repoClient := client.NewRepositoryClient()
		folders, err := repoClient.ListFolders(ctx, repositoryID, parentFolderID, withStatistic)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// 输出结果
		if quiet {
			outputData(cmd, folders)
			return
		}

		switch format {
		case "json":
			outputData(cmd, folders)
		case "table":
			printFolderTable(cmd, folders, withStatistic)
		default:
			printFolderTable(cmd, folders, withStatistic)
		}
	},
}

// repositoryFileCheckCmd 文件查重
var repositoryFileCheckCmd = &cobra.Command{
	Use:   "file-check",
	Short: "检查文件是否已存在",
	Long:  `通过 MD5 检查文件是否已存在于素材库中。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查认证
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		// 获取参数
		repositoryID, _ := cmd.Flags().GetInt64("repository-id")
		if repositoryID == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 必须指定 --repository-id")
			os.Exit(1)
		}

		filePath, _ := cmd.Flags().GetString("file")
		fileMD5, _ := cmd.Flags().GetString("file-md5")

		// 如果提供了文件路径，计算 MD5
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

		// 创建上下文
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// 调用 API
		repoClient := client.NewRepositoryClient()
		result, err := repoClient.CheckFile(ctx, repositoryID, fileMD5)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// 输出结果
		if quiet {
			outputData(cmd, result)
			return
		}

		if result.Existed {
			fmt.Fprintln(cmd.OutOrStdout(), "文件已存在 (重复)")
			if filePath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  文件: %s\n", filePath)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  MD5: %s\n", fileMD5)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "文件不存在 (可上传)")
			if filePath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  文件: %s\n", filePath)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  MD5: %s\n", fileMD5)
		}
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
  cbi repository file-create --repository-id 1 --file ./video.mp4 --folder-id 123
  cbi repository file-create --repository-id 1 --file ./image.png --force`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查认证
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			os.Exit(1)
		}

		// 获取参数
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

		// 检查文件是否存在
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "错误: 文件不存在 - %s\n", filePath)
			os.Exit(1)
		}

		// 检查文件大小（100MB 限制）
		if fileInfo.Size() > 100*1024*1024 {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 文件大小超过 100MB 限制")
			os.Exit(1)
		}

		skipCheck, _ := cmd.Flags().GetBool("skip-check")
		force, _ := cmd.Flags().GetBool("force")

		// 创建上下文
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

		// 查重检查（除非跳过）
		if !skipCheck {
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "正在计算文件 MD5...\n")
			}

			fileMD5, err := calculateFileMD5(filePath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "计算 MD5 失败: %s\n", err.Error())
				os.Exit(1)
			}

			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "文件 MD5: %s\n", fileMD5)
				fmt.Fprintf(cmd.ErrOrStderr(), "正在检查文件是否重复...\n")
			}

			result, err := repoClient.CheckFile(ctx, repositoryID, fileMD5)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
				os.Exit(1)
			}

			if result.Existed && !force {
				fmt.Fprintln(cmd.ErrOrStderr(), "错误: 文件已存在于素材库中（重复）")
				fmt.Fprintln(cmd.ErrOrStderr(), "")
				fmt.Fprintln(cmd.ErrOrStderr(), "选项:")
				fmt.Fprintln(cmd.ErrOrStderr(), "  --force      强制上传（覆盖重复文件）")
				fmt.Fprintln(cmd.ErrOrStderr(), "  --skip-check 跳过查重检查")
				os.Exit(1)
			}

			if result.Existed && force {
				fmt.Fprintln(cmd.ErrOrStderr(), "警告: 文件重复，将强制上传")
			}
		}

		// 构建请求
		req := &client.CreateFileRequest{
			RepositoryID: repositoryID,
			FilePath:     filePath,
		}

		// 可选参数
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

		// 上传文件
		if !quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "正在上传: %s\n", filePath)
		}

		fileInfoResult, err := repoClient.CreateFile(ctx, req)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// 输出结果
		if quiet {
			outputData(cmd, fileInfoResult)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 文件上传成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  文件 ID: %d\n", fileInfoResult.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  文件名: %s\n", fileInfoResult.Name)
	},
}

func init() {
	rootCmd.AddCommand(repositoryCmd)
	repositoryCmd.AddCommand(repositoryListCmd)
	repositoryCmd.AddCommand(repositoryFoldersCmd)
	repositoryCmd.AddCommand(repositoryFileCheckCmd)
	repositoryCmd.AddCommand(repositoryFileCreateCmd)

	// folders 命令参数
	repositoryFoldersCmd.Flags().Int64("repository-id", 0, "素材库 ID（必填）")
	repositoryFoldersCmd.Flags().Int64("parent-folder-id", 0, "父文件夹 ID（0 表示根目录）")
	repositoryFoldersCmd.Flags().Bool("with-statistic", false, "包含统计信息（文件数量）")

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
	repositoryFileCreateCmd.Flags().Bool("skip-check", false, "跳过查重检查")
	repositoryFileCreateCmd.Flags().Bool("force", false, "强制上传（即使文件重复）")
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