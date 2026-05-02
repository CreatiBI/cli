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

// portfolioCmd 代表 portfolio 命令组
var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "专案集管理",
	Long:  `管理专案集，包括查看专案集列表、专案集内专案列表。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// portfolioListCmd 专案集列表
var portfolioListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可访问的专案集",
	Long: `获取权限范围内的专案集列表。

示例：
  cbi portfolio list
  cbi portfolio list --keyword "品牌"
  cbi portfolio list --scope 1 --page 1 --pageSize 20`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		scope, _ := cmd.Flags().GetInt("scope")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		portfolioClient := client.NewPortfolioClient()
		result, err := portfolioClient.ListPortfolios(ctx, &client.PortfolioListRequest{
			Page:     page,
			PageSize: pageSize,
			Keyword:  keyword,
			Scope:    scope,
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
			printPortfolioListTable(cmd, result)
			return nil
		}
	},
}

// printPortfolioListTable 表格输出专案集列表
func printPortfolioListTable(cmd *cobra.Command, result *client.PortfolioListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Portfolios) == 0 {
		fmt.Fprintln(w, "无专案集")
		return
	}

	t := output.NewTableWriter(w)
	t.AppendHeader("ID", "名称", "颜色", "可见性", "专案数", "创建者", "创建时间")

	for _, p := range result.Portfolios {
		creator := "-"
		if p.Creator != nil {
			creator = p.Creator.Name
		}
		t.AppendRow(
			strconv.FormatInt(p.ID, 10),
			p.Name,
			p.Color,
			privacyName(p.Privacy),
			strconv.FormatInt(p.ProjectCount, 10),
			creator,
			formatDate(p.CreatedAt),
		)
	}

	t.Render()
}

// privacyName 获取可见性名称
func privacyName(privacy int) string {
	switch privacy {
	case 1:
		return "公开"
	case 2:
		return "私有"
	default:
		return fmt.Sprintf("未知(%d)", privacy)
	}
}

func init() {
	rootCmd.AddCommand(portfolioCmd)
	portfolioCmd.AddCommand(portfolioListCmd)

	// portfolioListCmd 参数
	portfolioListCmd.Flags().String("keyword", "", "搜索关键词")
	portfolioListCmd.Flags().Int("scope", 0, "范围筛选（0=所有可见, 1=我加入的）")
	portfolioListCmd.Flags().Int("page", 1, "页码")
	portfolioListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
}