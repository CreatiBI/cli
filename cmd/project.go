package cmd

import (
	"context"
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

// projectCreateCmd 创建专案
var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建专案",
	Long: `创建新专案。

示例：
  cbi project create --team-id 1 --name "品牌投放"
  cbi project create --team-id 1 --name "新品推广" --privacy 2 --description "新品上市推广素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		teamId, _ := cmd.Flags().GetInt64("team-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if teamId == 0 {
			return cliErr.NewCLIError("MISSING_TEAM_ID", "必须指定 --team-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		privacy, _ := cmd.Flags().GetInt("privacy")
		description, _ := cmd.Flags().GetString("description")
		templateId, _ := cmd.Flags().GetInt64("template-id")
		deadlineStart, _ := cmd.Flags().GetString("deadline-start")
		deadlineEnd, _ := cmd.Flags().GetString("deadline-end")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateProject(ctx, &client.ProjectCreateRequest{
			TeamId:        teamId,
			Name:          name,
			Privacy:       privacy,
			Description:   description,
			TemplateId:    templateId,
			DeadlineStart: deadlineStart,
			DeadlineEnd:   deadlineEnd,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 专案创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  专案 ID: %d\n", result.ProjectId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

// projectScriptListCmd 脚本列表
var projectScriptListCmd = &cobra.Command{
	Use:   "script-list",
	Short: "列出专案脚本",
	Long: `获取专案的脚本列表。

示例：
  cbi project script-list --project-id 1
  cbi project script-list --project-id 1 --keyword "广告" --state 1
  cbi project script-list --project-id 1 --page 2 --pageSize 30`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		state, _ := cmd.Flags().GetInt("state")
		parentId, _ := cmd.Flags().GetInt64("parent-id")
		isArchived, _ := cmd.Flags().GetInt("is-archived")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListScripts(ctx, &client.ScriptListRequest{
			ProjectId:  projectId,
			Page:       page,
			PageSize:   pageSize,
			Keyword:    keyword,
			State:      state,
			ParentId:   parentId,
			IsArchived: isArchived,
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
			printScriptListTable(cmd, result)
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

// printScriptListTable 表格输出脚本列表
func printScriptListTable(cmd *cobra.Command, result *client.ScriptListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Scripts) == 0 {
		fmt.Fprintln(w, "无脚本")
		return
	}

	t := output.NewTableWriter(w)
	t.AppendHeader("ID", "名称", "状态", "编剧", "设计师", "截止日期")

	for _, s := range result.Scripts {
		writer := "-"
		if s.AssignedWriter != nil {
			writer = s.AssignedWriter.Name
		}
		designer := "-"
		if s.AssignedDesigner != nil {
			designer = s.AssignedDesigner.Name
		}
		t.AppendRow(
			strconv.FormatInt(s.ID, 10),
			s.Name,
			scriptStateName(s.State),
			writer,
			designer,
			formatDate(s.DueDate),
		)
	}

	t.Render()
}

// scriptStateName 获取脚本状态名称
func scriptStateName(state int) string {
	switch state {
	case 1:
		return "待处理"
	case 2:
		return "进行中"
	case 3:
		return "已完成"
	case 4:
		return "已归档"
	default:
		return fmt.Sprintf("未知(%d)", state)
	}
}

// projectMaterialListCmd 素材列表
var projectMaterialListCmd = &cobra.Command{
	Use:   "material-list",
	Short: "列出专案素材",
	Long: `获取专案的素材列表。

示例：
  cbi project material-list --project-id 1
  cbi project material-list --project-id 1 --keyword "视频"
  cbi project material-list --project-id 1 --page 2 --pageSize 30`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("pageSize")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.ListMaterials(ctx, &client.MaterialListRequest{
			ProjectId: projectId,
			Page:      page,
			PageSize:  pageSize,
			Keyword:   keyword,
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
			printMaterialListTable(cmd, result)
			return nil
		}
	},
}

// printMaterialListTable 表格输出素材列表
func printMaterialListTable(cmd *cobra.Command, result *client.MaterialListResult) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "共 %d 条，第 %d/%d 页\n\n",
		result.Total, result.Page, totalPages(result.Total, result.PageSize))

	if len(result.Materials) == 0 {
		fmt.Fprintln(w, "无素材")
		return
	}

	t := output.NewTableWriter(w)
	t.AppendHeader("ID", "名称", "类型", "格式", "时长", "创建者")

	for _, m := range result.Materials {
		creator := "-"
		if m.Creator != nil {
			creator = m.Creator.Name
		}
		t.AppendRow(
			strconv.FormatInt(m.ID, 10),
			m.Name,
			fileTypeName(m.FileType),
			m.Format,
			m.Duration,
			creator,
		)
	}

	t.Render()
}

// fileTypeName 获取文件类型名称
func fileTypeName(fileType int) string {
	switch fileType {
	case 1:
		return "视频"
	case 2:
		return "图片"
	default:
		return fmt.Sprintf("未知(%d)", fileType)
	}
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectScriptListCmd)
	projectCmd.AddCommand(projectMaterialListCmd)

	// projectListCmd 参数
	projectListCmd.Flags().String("keyword", "", "搜索关键词")
	projectListCmd.Flags().String("team-ids", "", "团队 ID 列表（逗号分隔）")
	projectListCmd.Flags().String("portfolio-ids", "", "作品集 ID 列表（逗号分隔）")
	projectListCmd.Flags().Int("scope", 0, "范围筛选（0=所有可见, 1=我加入的）")
	projectListCmd.Flags().Int("page", 1, "页码")
	projectListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")

	// projectCreateCmd 参数
	projectCreateCmd.Flags().Int64("team-id", 0, "团队 ID（必填）")
	projectCreateCmd.Flags().String("name", "", "专案名称（必填）")
	projectCreateCmd.Flags().Int("privacy", 1, "隐私设置（1=公开, 2=私有）")
	projectCreateCmd.Flags().String("description", "", "专案描述")
	projectCreateCmd.Flags().Int64("template-id", 0, "模板 ID")
	projectCreateCmd.Flags().String("deadline-start", "", "截止日期开始（YYYY-MM-DD）")
	projectCreateCmd.Flags().String("deadline-end", "", "截止日期结束（YYYY-MM-DD）")
	projectCreateCmd.MarkFlagRequired("team-id")
	projectCreateCmd.MarkFlagRequired("name")

	// projectScriptListCmd 参数
	projectScriptListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectScriptListCmd.Flags().String("keyword", "", "搜索关键词")
	projectScriptListCmd.Flags().Int("state", 0, "任务状态筛选")
	projectScriptListCmd.Flags().Int64("parent-id", 0, "父任务筛选")
	projectScriptListCmd.Flags().Int("is-archived", 0, "档案筛选（0=不过滤, 1=档案, 2=非档案）")
	projectScriptListCmd.Flags().Int("page", 1, "页码")
	projectScriptListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
	projectScriptListCmd.MarkFlagRequired("project-id")

	// projectMaterialListCmd 参数
	projectMaterialListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialListCmd.Flags().String("keyword", "", "搜索关键词")
	projectMaterialListCmd.Flags().Int("page", 1, "页码")
	projectMaterialListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
	projectMaterialListCmd.MarkFlagRequired("project-id")
}