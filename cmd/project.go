package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/output"
)

// projectCmd 代表 project 命令组
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "专案管理",
	Long:  `管理专案，包括查看专案列表、创建专案、脚本列表、素材列表。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// projectListCmd 专案列表
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可访问的专案",
	Long: `获取权限范围内的专案列表。

示例：
  cbi project list
  cbi project list --keyword "品牌"
  cbi project list --scope 1 --page 1 --pageSize 20`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		teamIdsStr, _ := cmd.Flags().GetString("team-ids")
		portfolioIdsStr, _ := cmd.Flags().GetString("portfolio-ids")
		scope, _ := cmd.Flags().GetInt("scope")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		// 解析 teamIds
		var teamIds []int64
		if teamIdsStr != "" {
			ids, err := parseIDList(teamIdsStr, "team-ids")
			if err != nil {
				return err
			}
			teamIds = ids
		}

		// 解析 portfolioIds
		var portfolioIds []int64
		if portfolioIdsStr != "" {
			ids, err := parseIDList(portfolioIdsStr, "portfolio-ids")
			if err != nil {
				return err
			}
			portfolioIds = ids
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListProjects(ctx, &client.ProjectListRequest{
			Page:         page,
			PageSize:     pageSize,
			Keyword:      keyword,
			TeamIds:      teamIds,
			PortfolioIds: portfolioIds,
			Scope:        scope,
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
			printProjectListTable(cmd, result)
			return nil
		}
	},
}

// printProjectListTable 表格输出专案列表
func printProjectListTable(cmd *cobra.Command, result *client.ProjectListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Projects) == 0 {
		fmt.Fprintln(w, "无专案")
		return
	}

	t := output.NewTableWriter(w)
	t.AppendHeader("ID", "名称", "创建者", "创建时间")

	for _, p := range result.Projects {
		creator := "-"
		if p.Creator != nil {
			creator = p.Creator.Name
		}
		t.AppendRow(
			strconv.FormatInt(p.ID, 10),
			p.Name,
			creator,
			formatDate(p.CreatedAt),
		)
	}

	t.Render()
}

// totalPages 计算总页数
func totalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	return pages
}

// formatDate 格式化日期（截取 YYYY-MM-DD 部分）
func formatDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)

	// projectListCmd 参数
	projectListCmd.Flags().String("keyword", "", "搜索关键词")
	projectListCmd.Flags().String("team-ids", "", "团队 ID 列表（逗号分隔）")
	projectListCmd.Flags().String("portfolio-ids", "", "作品集 ID 列表（逗号分隔）")
	projectListCmd.Flags().Int("scope", 0, "范围筛选（0=所有可见, 1=我加入的）")
	projectListCmd.Flags().Int("page", 1, "页码")
	projectListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
}