package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

// projectMaterialCmd 素材操作命令组
var projectMaterialCmd = &cobra.Command{
	Use:   "material",
	Short: "素材操作",
	Long:  `素材操作命令组，包括从脚本创建素材、从素材创建子素材等。`,
}

// projectScriptCreateCmd 创建脚本任务
var projectScriptCreateCmd = &cobra.Command{
	Use:   "script-create",
	Short: "创建脚本任务",
	Long: `创建新的脚本任务。

示例：
  cbi project script-create --project-id 1 --name "脚本任务名称"
  cbi project script-create --project-id 1 --name "子任务" --parent-id 100`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		parentId, _ := cmd.Flags().GetInt64("parent-id")
		sourceObject, _ := cmd.Flags().GetString("source-object")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateScriptTask(ctx, &client.CreateScriptTaskRequest{
			ProjectId:    projectId,
			Name:         name,
			ParentId:     parentId,
			SourceObject: sourceObject,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 脚本任务创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  脚本 ID: %d\n", result.ScriptId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

// projectScriptGetCmd 获取脚本内容
var projectScriptGetCmd = &cobra.Command{
	Use:   "script-get",
	Short: "获取脚本内容",
	Long: `获取脚本的完整内容，包括格式、JSON/Markdown 内容和关联信息。

示例：
  cbi project script-get --script-id 37110
  cbi project script-get --script-id 37110 --project-id 2359`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptId, _ := cmd.Flags().GetInt64("script-id")
		if scriptId == 0 {
			return cliErr.NewCLIError("MISSING_SCRIPT_ID", "必须指定 --script-id")
		}

		projectId, _ := cmd.Flags().GetInt64("project-id")

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.GetScriptContent(ctx, &client.GetScriptContentRequest{
			ScriptId:  scriptId,
			ProjectId: projectId,
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
			printScriptContent(cmd, result)
			return nil
		}
	},
}

// printScriptContent 表格输出脚本内容
func printScriptContent(cmd *cobra.Command, result *client.ScriptContent) {
	w := cmd.OutOrStdout()

	fmt.Fprintf(w, "脚本内容:\n")
	fmt.Fprintf(w, "  ID:         %d\n", result.ScriptId)
	fmt.Fprintf(w, "  专案 ID:    %d\n", result.ProjectId)
	fmt.Fprintf(w, "  名称:       %s\n", result.Name)
	fmt.Fprintf(w, "  格式:       %s\n", scriptFormatName(result.Format))
	fmt.Fprintf(w, "  创建时间:   %s\n", result.CreatedAt)
	fmt.Fprintf(w, "  更新时间:   %s\n", result.UpdatedAt)

	if len(result.ProductIds) > 0 {
		fmt.Fprintf(w, "  关联产品:   %v\n", result.ProductIds)
	}
	if len(result.AppIds) > 0 {
		fmt.Fprintf(w, "  关联应用:   %v\n", result.AppIds)
	}
	if len(result.Ratios) > 0 {
		fmt.Fprintf(w, "  关联尺寸:   %v\n", result.Ratios)
	}
	if len(result.RefRepoFileIds) > 0 {
		fmt.Fprintf(w, "  引用文件:   %v\n", result.RefRepoFileIds)
	}

	// 输出内容
	if result.Markdown != "" {
		fmt.Fprintf(w, "\nMarkdown 内容:\n%s\n", result.Markdown)
	}
	if result.Script != "" {
		fmt.Fprintf(w, "\n脚本 JSON:\n%s\n", result.Script)
	}
}

// scriptFormatName 获取脚本格式名称
func scriptFormatName(format int) string {
	switch format {
	case 1:
		return "普通(Markdown)"
	case 2:
		return "分镜(JSON)"
	case 3:
		return "口播(JSON)"
	case 4:
		return "剪辑(JSON)"
	default:
		return fmt.Sprintf("未知(%d)", format)
	}
}

// projectScriptSaveCmd 保存脚本内容
var projectScriptSaveCmd = &cobra.Command{
	Use:   "script-save",
	Short: "保存脚本内容",
	Long: `保存或修改脚本内容，系统自动从 JSON 内容推导格式。

示例：
  # 保存分镜格式脚本
  cbi project script-save --script-id 37110 --script '{"type":"doc","content":[...]}'

  # 保存普通 Markdown 脚本
  cbi project script-save --script-id 37110 --markdown "# 标题\n正文内容"

  # 更新脚本名称和关联信息
  cbi project script-save --script-id 37110 --name "新名称" --product-ids 1,2`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptId, _ := cmd.Flags().GetInt64("script-id")
		if scriptId == 0 {
			return cliErr.NewCLIError("MISSING_SCRIPT_ID", "必须指定 --script-id")
		}

		projectId, _ := cmd.Flags().GetInt64("project-id")
		name, _ := cmd.Flags().GetString("name")
		script, _ := cmd.Flags().GetString("script")
		markdown, _ := cmd.Flags().GetString("markdown")
		formatVal, _ := cmd.Flags().GetInt("format")
		productIdsStr, _ := cmd.Flags().GetString("product-ids")
		appIdsStr, _ := cmd.Flags().GetString("app-ids")
		ratiosStr, _ := cmd.Flags().GetString("ratios")
		refRepoFileIdsStr, _ := cmd.Flags().GetString("ref-repo-file-ids")

		// 解析 ID 列表
		var productIds, appIds, refRepoFileIds []int64
		var ratios []int32
		if productIdsStr != "" {
			productIds = parseIDListToInt64(productIdsStr)
		}
		if appIdsStr != "" {
			appIds = parseIDListToInt64(appIdsStr)
		}
		if ratiosStr != "" {
			ratios = parseIDListToInt32(ratiosStr)
		}
		if refRepoFileIdsStr != "" {
			refRepoFileIds = parseIDListToInt64(refRepoFileIdsStr)
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.SaveScriptContent(ctx, &client.SaveScriptContentRequest{
			ScriptId:       scriptId,
			ProjectId:      projectId,
			Format:         formatVal,
			Name:           name,
			Script:         script,
			Markdown:       markdown,
			ProductIds:     productIds,
			AppIds:         appIds,
			Ratios:         ratios,
			RefRepoFileIds: refRepoFileIds,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 脚本内容保存成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  脚本 ID: %d\n", result.ScriptId)
		fmt.Fprintf(cmd.OutOrStdout(), "  格式:   %s\n", scriptFormatName(result.Format))
		fmt.Fprintf(cmd.OutOrStdout(), "  名称:   %s\n", result.Name)
		return nil
	},
}

// parseIDListToInt64 解析逗号分隔的 ID 列表为 int64
func parseIDListToInt64(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := []int64{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

// parseIDListToInt32 解析逗号分隔的 ID 列表为 int32
func parseIDListToInt32(s string) []int32 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := []int32{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 32)
		if err == nil {
			result = append(result, int32(id))
		}
	}
	return result
}

// projectMaterialFissionFromTaskCmd 从脚本创建裂变素材
var projectMaterialFissionFromTaskCmd = &cobra.Command{
	Use:   "fission-from-task",
	Short: "从脚本创建裂变素材",
	Long: `从脚本任务创建裂变素材，素材与脚本为父子关系。

示例：
  cbi project material fission-from-task --project-id 1 --script-id 100 --name "裂变素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		scriptId, _ := cmd.Flags().GetInt64("script-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}
		if scriptId == 0 {
			return cliErr.NewCLIError("MISSING_SCRIPT_ID", "必须指定 --script-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateFissionMaterialFromTask(ctx, &client.CreateFissionMaterialFromTaskRequest{
			ProjectId: projectId,
			ScriptId:  scriptId,
			Name:      name,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 裂变素材创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  素材 ID: %d\n", result.MaterialId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

// projectMaterialDerivativeFromTaskCmd 从脚本创建衍生素材
var projectMaterialDerivativeFromTaskCmd = &cobra.Command{
	Use:   "derivative-from-task",
	Short: "从脚本创建衍生素材",
	Long: `从脚本任务创建衍生素材，素材与脚本为平级关系，可跨专案。

示例：
  cbi project material derivative-from-task --project-id 1 --script-id 100 --name "衍生素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		scriptId, _ := cmd.Flags().GetInt64("script-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}
		if scriptId == 0 {
			return cliErr.NewCLIError("MISSING_SCRIPT_ID", "必须指定 --script-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateDerivativeMaterialFromTask(ctx, &client.CreateDerivativeMaterialFromTaskRequest{
			ProjectId: projectId,
			ScriptId:  scriptId,
			Name:      name,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 衍生素材创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  素材 ID: %d\n", result.MaterialId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

// projectMaterialFissionFromMaterialCmd 从素材创建裂变子素材
var projectMaterialFissionFromMaterialCmd = &cobra.Command{
	Use:   "fission-from-material",
	Short: "从素材创建裂变子素材",
	Long: `从已有素材创建裂变子素材，新素材与原素材为父子关系。

示例：
  cbi project material fission-from-material --project-id 1 --material-id 100 --name "裂变子素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		materialId, _ := cmd.Flags().GetInt64("material-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}
		if materialId == 0 {
			return cliErr.NewCLIError("MISSING_MATERIAL_ID", "必须指定 --material-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateFissionMaterialFromMaterial(ctx, &client.CreateFissionMaterialFromMaterialRequest{
			ProjectId:  projectId,
			MaterialId: materialId,
			Name:       name,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 裂变子素材创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  素材 ID: %d\n", result.MaterialId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

// projectMaterialDerivativeFromMaterialCmd 从素材创建衍生子素材
var projectMaterialDerivativeFromMaterialCmd = &cobra.Command{
	Use:   "derivative-from-material",
	Short: "从素材创建衍生子素材",
	Long: `从已有素材创建衍生子素材，新素材与原素材为平级关系，可跨专案。

示例：
  cbi project material derivative-from-material --project-id 1 --material-id 100 --name "衍生子素材"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectId, _ := cmd.Flags().GetInt64("project-id")
		materialId, _ := cmd.Flags().GetInt64("material-id")
		name, _ := cmd.Flags().GetString("name")

		// 必填参数验证
		if projectId == 0 {
			return cliErr.NewCLIError("MISSING_PROJECT_ID", "必须指定 --project-id")
		}
		if materialId == 0 {
			return cliErr.NewCLIError("MISSING_MATERIAL_ID", "必须指定 --material-id")
		}
		if name == "" {
			return cliErr.NewCLIError("MISSING_NAME", "必须指定 --name")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		projectClient := client.NewProjectClient()
		result, err := projectClient.CreateDerivativeMaterialFromMaterial(ctx, &client.CreateDerivativeMaterialFromMaterialRequest{
			ProjectId:  projectId,
			MaterialId: materialId,
			Name:       name,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 衍生子素材创建成功")
		fmt.Fprintf(cmd.OutOrStdout(), "  素材 ID: %d\n", result.MaterialId)
		fmt.Fprintf(cmd.OutOrStdout(), "  名称: %s\n", result.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectScriptListCmd)
	projectCmd.AddCommand(projectMaterialListCmd)
	projectCmd.AddCommand(projectScriptCreateCmd)
	projectCmd.AddCommand(projectScriptGetCmd)
	projectCmd.AddCommand(projectScriptSaveCmd)
	projectCmd.AddCommand(projectMaterialCmd)
	projectMaterialCmd.AddCommand(projectMaterialFissionFromTaskCmd)
	projectMaterialCmd.AddCommand(projectMaterialDerivativeFromTaskCmd)
	projectMaterialCmd.AddCommand(projectMaterialFissionFromMaterialCmd)
	projectMaterialCmd.AddCommand(projectMaterialDerivativeFromMaterialCmd)

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

	// projectScriptListCmd 参数
	projectScriptListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectScriptListCmd.Flags().String("keyword", "", "搜索关键词")
	projectScriptListCmd.Flags().Int("state", 0, "任务状态筛选")
	projectScriptListCmd.Flags().Int64("parent-id", 0, "父任务筛选")
	projectScriptListCmd.Flags().Int("is-archived", 0, "档案筛选（0=不过滤, 1=档案, 2=非档案）")
	projectScriptListCmd.Flags().Int("page", 1, "页码")
	projectScriptListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")
	// projectMaterialListCmd 参数
	projectMaterialListCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialListCmd.Flags().String("keyword", "", "搜索关键词")
	projectMaterialListCmd.Flags().Int("page", 1, "页码")
	projectMaterialListCmd.Flags().Int("pageSize", 20, "每页条数（最大 50）")

	// projectScriptCreateCmd 参数
	projectScriptCreateCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectScriptCreateCmd.Flags().String("name", "", "脚本任务名称（必填）")
	projectScriptCreateCmd.Flags().Int64("parent-id", 0, "父任务 ID")
	projectScriptCreateCmd.Flags().String("source-object", "", "来源对象")

	// projectScriptGetCmd 参数
	projectScriptGetCmd.Flags().Int64("script-id", 0, "脚本任务 ID（必填）")
	projectScriptGetCmd.Flags().Int64("project-id", 0, "专案 ID（可选）")

	// projectScriptSaveCmd 参数
	projectScriptSaveCmd.Flags().Int64("script-id", 0, "脚本任务 ID（必填）")
	projectScriptSaveCmd.Flags().Int64("project-id", 0, "专案 ID（可选）")
	projectScriptSaveCmd.Flags().Int("format", 0, "脚本格式（可选，不传自动推导：1=普通 2=分镜 3=口播 4=剪辑）")
	projectScriptSaveCmd.Flags().String("name", "", "脚本名称（可选）")
	projectScriptSaveCmd.Flags().String("script", "", "脚本内容 JSON（分镜/口播/剪辑格式）")
	projectScriptSaveCmd.Flags().String("markdown", "", "Markdown 内容（普通格式）")
	projectScriptSaveCmd.Flags().String("product-ids", "", "关联产品 ID（逗号分隔）")
	projectScriptSaveCmd.Flags().String("app-ids", "", "关联渠道应用 ID（逗号分隔）")
	projectScriptSaveCmd.Flags().String("ratios", "", "关联尺寸（逗号分隔）")
	projectScriptSaveCmd.Flags().String("ref-repo-file-ids", "", "引用仓库文件 ID（逗号分隔）")

	// projectMaterialFissionFromTaskCmd 参数
	projectMaterialFissionFromTaskCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialFissionFromTaskCmd.Flags().Int64("script-id", 0, "脚本任务 ID（必填）")
	projectMaterialFissionFromTaskCmd.Flags().String("name", "", "素材名称（必填）")

	// projectMaterialDerivativeFromTaskCmd 参数
	projectMaterialDerivativeFromTaskCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialDerivativeFromTaskCmd.Flags().Int64("script-id", 0, "脚本任务 ID（必填）")
	projectMaterialDerivativeFromTaskCmd.Flags().String("name", "", "素材名称（必填）")

	// projectMaterialFissionFromMaterialCmd 参数
	projectMaterialFissionFromMaterialCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialFissionFromMaterialCmd.Flags().Int64("material-id", 0, "素材 ID（必填）")
	projectMaterialFissionFromMaterialCmd.Flags().String("name", "", "素材名称（必填）")

	// projectMaterialDerivativeFromMaterialCmd 参数
	projectMaterialDerivativeFromMaterialCmd.Flags().Int64("project-id", 0, "专案 ID（必填）")
	projectMaterialDerivativeFromMaterialCmd.Flags().Int64("material-id", 0, "素材 ID（必填）")
	projectMaterialDerivativeFromMaterialCmd.Flags().String("name", "", "素材名称（必填）")
}