package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/output"
)

// ---- Root command ----

var adCmd = &cobra.Command{
	Use:   "ad",
	Short: "广告数据查询",
	Long:  `查询广告投放数据，支持巨量引擎、千川、腾讯广告三大平台。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsLoggedIn() {
			return cliErr.ErrAuthRequired
		}
		return nil
	},
}

// ---- Helper functions ----

func appDisplayName(appId uint32) string {
	switch appId {
	case client.AppOceanengine:
		return "巨量引擎"
	case client.AppQianchuan:
		return "巨量千川"
	case client.AppTencent:
		return "腾讯广告"
	default:
		return fmt.Sprintf("未知(%d)", appId)
	}
}

func authStatusText(status uint32) string {
	switch status {
	case 0:
		return "无效"
	case 1:
		return "有效"
	case 2:
		return "过期"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}

func materialTypeText(t uint32) string {
	switch t {
	case 1:
		return "视频"
	case 2:
		return "图片"
	case 3:
		return "组图"
	case 4:
		return "文字"
	default:
		return fmt.Sprintf("未知(%d)", t)
	}
}

func objTypeName(objType uint32) string {
	switch objType {
	case client.ObjTypeAdvertiser:
		return "广告主"
	case client.ObjTypeCampaign:
		return "计划"
	case client.ObjTypeAdgroup:
		return "广告组"
	case client.ObjTypeAds:
		return "广告"
	case client.ObjTypeCreative:
		return "素材"
	default:
		return fmt.Sprintf("未知(%d)", objType)
	}
}

// resolveRequiredApp resolves and validates the --app flag, returning the appId int.
func resolveRequiredApp(cmd *cobra.Command) (int, error) {
	appStr, _ := cmd.Flags().GetString("app")
	if appStr == "" {
		return 0, cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	}
	return client.ResolveAppID(appStr)
}

// pagePagination returns page and pageSize from flags.
func pagePagination(cmd *cobra.Command) (uint32, uint32) {
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return uint32(page), uint32(pageSize)
}

// adTotalPages calculates total pages from total count and page size.
func adTotalPages(total, pageSize uint32) uint32 {
	if pageSize == 0 {
		return 1
	}
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}

// ---- Browse command ----

var adBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "智能入口（检测行业+缓存）",
	Long: `智能浏览广告数据入口，自动检测电商/非电商行业。

电商行业（如千川）无产品概念，直接进入账户列表；非电商行业先展示产品列表。
行业检测结果会被缓存，使用 --reset 清除缓存重新检测。`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reset, _ := cmd.Flags().GetBool("reset")
		appStr, _ := cmd.Flags().GetString("app")

		if reset {
			if err := config.ClearAdIsEcomIndustry(); err != nil {
				return cliErr.NewCLIErrorWithDetail("CLEAR_CACHE_ERROR", "清除行业缓存失败", err.Error())
			}
			fmt.Fprintln(cmd.OutOrStdout(), "清除行业缓存")
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		isEcom := config.GetAdIsEcomIndustry()

		if isEcom == nil {
			// No cache, need to detect
			result, err := adClient.ListAdProducts(ctx, &client.ListAdProductsRequest{})
			if err != nil {
				return err
			}
			isEcomVal := result.IsEcomIndustry
			isEcom = &isEcomVal
			if err := config.SetAdIsEcomIndustry(isEcomVal); err != nil {
				// non-fatal, just log
				fmt.Fprintln(cmd.ErrOrStderr(), "警告: 行业缓存保存失败")
			}
		}

		if *isEcom {
			fmt.Fprintln(cmd.OutOrStdout(), "电商行业（无产品概念），直接进入广告账户")

			req := &client.ListAdAccountsRequest{}
			if appStr != "" {
				appId, err := client.ResolveAppID(appStr)
				if err != nil {
					return err
				}
				req.AppId = uint32(appId)
			}

			result, err := adClient.ListAdAccounts(ctx, req)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, result)
			}
			if format == "json" {
				return outputData(cmd, result)
			}
			printAdAccountTable(cmd, result)
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "非电商行业")

		req := &client.ListAdProductsRequest{}
		if appStr != "" {
			// app filter not directly supported in product list, ignore silently
		}

		result, err := adClient.ListAdProducts(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdProductTable(cmd, result)
		return nil
	},
}

// ---- Product sub-group ----

var adProductCmd = &cobra.Command{
	Use:   "product",
	Short: "广告产品管理",
}

var adProductListCmd = &cobra.Command{
	Use:   "list",
	Short: "产品列表",
	Long:  `获取广告产品列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdProducts(ctx, &client.ListAdProductsRequest{
			Keyword:  keyword,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdProductTable(cmd, result)
		return nil
	},
}

// ---- Channel sub-group ----

var adChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "广告平台频道",
}

var adChannelSchemaCmd = &cobra.Command{
	Use:   "schema <appId>",
	Short: "平台层级结构",
	Long: `获取广告平台的层级结构和字段定义。

appId 支持数字或别名: oceanengine/oe(1), qianchuan/qc(5), tencent/tx(6)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := client.ResolveAppID(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		schema, err := adClient.GetAdPlatformSchema(ctx, appId)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, schema)
		}
		if format == "json" {
			return outputData(cmd, schema)
		}
		printAdPlatformSchemaTable(cmd, schema)
		return nil
	},
}

// ---- Account sub-group ----

var adAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "广告账户",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If an accountId is provided, show account detail
		if len(args) > 0 {
			accountID := args[0]
			product, _ := cmd.Flags().GetUint32("product")

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()

			// First try keyword search (may not work for account IDs)
			result, err := adClient.ListAdAccounts(ctx, &client.ListAdAccountsRequest{
				Keyword:   accountID,
				ProductID: product,
			})
			if err != nil {
				return err
			}

			// Find matching account
			var found *client.AdAccount
			for i := range result.Accounts {
				if result.Accounts[i].AdAccountID == accountID || fmt.Sprintf("%d", result.Accounts[i].ID) == accountID {
					found = &result.Accounts[i]
					break
				}
			}
			// If keyword didn't match but we got one result, use it
			if found == nil && len(result.Accounts) == 1 {
				found = &result.Accounts[0]
			}

			// If still not found, try getting accounts without keyword and search by ID
			if found == nil {
				searchResult, err := adClient.ListAdAccounts(ctx, &client.ListAdAccountsRequest{
					ProductID: product,
					Page:      1,
					PageSize:   500,
				})
				if err != nil {
					return err
				}
				for i := range searchResult.Accounts {
					if searchResult.Accounts[i].AdAccountID == accountID || fmt.Sprintf("%d", searchResult.Accounts[i].ID) == accountID {
						found = &searchResult.Accounts[i]
						break
					}
				}
			}

			if found == nil {
				return cliErr.NewCLIError("ACCOUNT_NOT_FOUND", fmt.Sprintf("未找到账户: %s", accountID))
			}

			if quiet {
				return outputData(cmd, found)
			}
			if format == "json" {
				return outputData(cmd, found)
			}
			printAdAccountDetail(cmd, found)
			return nil
		}
		// No args, show help
		return cmd.Help()
	},
}

var adAccountListCmd = &cobra.Command{
	Use:   "list",
	Short: "账户列表",
	Long:  `获取广告账户列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appStr, _ := cmd.Flags().GetString("app")
		product, _ := cmd.Flags().GetUint32("product")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		req := &client.ListAdAccountsRequest{
			ProductID: product,
			Keyword:   keyword,
			Page:      page,
			PageSize:  pageSize,
		}

		if appStr != "" {
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}
			req.AppId = uint32(appId)
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdAccounts(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdAccountTable(cmd, result)
		return nil
	},
}

// ---- Advertiser sub-group ----

var adAdvertiserCmd = &cobra.Command{
	Use:   "advertiser",
	Short: "广告主",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If an objId is provided, show advertiser detail
		if len(args) > 0 {
			appStr, _ := cmd.Flags().GetString("app")
			if appStr == "" {
				return cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
			}
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}

			objID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return cliErr.NewCLIError("INVALID_OBJ_ID", fmt.Sprintf("无效的对象 ID: %s", args[0]))
			}

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()
			obj, err := adClient.GetAdObjectDetail(ctx, appId, client.ObjTypeAdvertiser, objID)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, obj)
			}
			if format == "json" {
				return outputData(cmd, obj)
			}
			printAdObjectDetail(cmd, obj)
			return nil
		}
		// No args, show help
		return cmd.Help()
	},
}

var adAdvertiserListCmd = &cobra.Command{
	Use:   "list",
	Short: "广告主列表",
	Long:  `获取广告主列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}
		product, _ := cmd.Flags().GetUint32("product")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdObjects(ctx, &client.ListAdObjectsRequest{
			AppId:     uint32(appId),
			ObjType:   client.ObjTypeAdvertiser,
			ProductID: product,
			Keyword:   keyword,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdObjectTable(cmd, result)
		return nil
	},
}

// ---- Campaign sub-group ----

var adCampaignCmd = &cobra.Command{
	Use:   "campaign",
	Short: "广告计划",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			appStr, _ := cmd.Flags().GetString("app")
			if appStr == "" {
				return cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
			}
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}

			objID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return cliErr.NewCLIError("INVALID_OBJ_ID", fmt.Sprintf("无效的对象 ID: %s", args[0]))
			}

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()
			obj, err := adClient.GetAdObjectDetail(ctx, appId, client.ObjTypeCampaign, objID)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, obj)
			}
			if format == "json" {
				return outputData(cmd, obj)
			}
			printAdObjectDetail(cmd, obj)
			return nil
		}
		return cmd.Help()
	},
}

var adCampaignListCmd = &cobra.Command{
	Use:   "list",
	Short: "计划列表",
	Long:  `获取广告计划列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}
		advertiser, _ := cmd.Flags().GetUint64("advertiser")
		product, _ := cmd.Flags().GetUint32("product")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdObjects(ctx, &client.ListAdObjectsRequest{
			AppId:     uint32(appId),
			ObjType:   client.ObjTypeCampaign,
			ParentID:  advertiser,
			ProductID: product,
			Keyword:   keyword,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdObjectTable(cmd, result)
		return nil
	},
}

var adCampaignFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "计划字段定义",
	Long:  `获取广告计划的字段定义。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		fields, err := adClient.ListAdObjectFields(ctx, appId, client.ObjTypeCampaign)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, fields)
		}
		if format == "json" {
			return outputData(cmd, fields)
		}
		printAdObjectFieldsTable(cmd, fields)
		return nil
	},
}

// ---- Adgroup sub-group ----

var adAdgroupCmd = &cobra.Command{
	Use:   "adgroup",
	Short: "广告组",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			appStr, _ := cmd.Flags().GetString("app")
			if appStr == "" {
				return cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
			}
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}

			objID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return cliErr.NewCLIError("INVALID_OBJ_ID", fmt.Sprintf("无效的对象 ID: %s", args[0]))
			}

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()
			obj, err := adClient.GetAdObjectDetail(ctx, appId, client.ObjTypeAdgroup, objID)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, obj)
			}
			if format == "json" {
				return outputData(cmd, obj)
			}
			printAdObjectDetail(cmd, obj)
			return nil
		}
		return cmd.Help()
	},
}

var adAdgroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "广告组列表",
	Long:  `获取广告组列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}
		advertiser, _ := cmd.Flags().GetUint64("advertiser")
		product, _ := cmd.Flags().GetUint32("product")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdObjects(ctx, &client.ListAdObjectsRequest{
			AppId:     uint32(appId),
			ObjType:   client.ObjTypeAdgroup,
			ParentID:  advertiser,
			ProductID: product,
			Keyword:   keyword,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdObjectTable(cmd, result)
		return nil
	},
}

var adAdgroupFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "广告组字段定义",
	Long:  `获取广告组的字段定义。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		fields, err := adClient.ListAdObjectFields(ctx, appId, client.ObjTypeAdgroup)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, fields)
		}
		if format == "json" {
			return outputData(cmd, fields)
		}
		printAdObjectFieldsTable(cmd, fields)
		return nil
	},
}

// ---- Ads sub-group ----

var adAdsCmd = &cobra.Command{
	Use:   "ads",
	Short: "广告",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			appStr, _ := cmd.Flags().GetString("app")
			if appStr == "" {
				return cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
			}
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}

			objID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return cliErr.NewCLIError("INVALID_OBJ_ID", fmt.Sprintf("无效的对象 ID: %s", args[0]))
			}

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()
			obj, err := adClient.GetAdObjectDetail(ctx, appId, client.ObjTypeAds, objID)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, obj)
			}
			if format == "json" {
				return outputData(cmd, obj)
			}
			printAdObjectDetail(cmd, obj)
			return nil
		}
		return cmd.Help()
	},
}

var adAdsListCmd = &cobra.Command{
	Use:   "list",
	Short: "广告列表",
	Long:  `获取广告列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}
		campaign, _ := cmd.Flags().GetUint64("campaign")
		adgroup, _ := cmd.Flags().GetUint64("adgroup")
		product, _ := cmd.Flags().GetUint32("product")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		// --campaign takes precedence over --adgroup
		parentID := uint64(0)
		if campaign > 0 {
			parentID = campaign
		} else if adgroup > 0 {
			parentID = adgroup
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListAdObjects(ctx, &client.ListAdObjectsRequest{
			AppId:     uint32(appId),
			ObjType:   client.ObjTypeAds,
			ParentID:  parentID,
			ProductID: product,
			Keyword:   keyword,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdObjectTable(cmd, result)
		return nil
	},
}

var adAdsFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "广告字段定义",
	Long:  `获取广告的字段定义。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appId, err := resolveRequiredApp(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		fields, err := adClient.ListAdObjectFields(ctx, appId, client.ObjTypeAds)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, fields)
		}
		if format == "json" {
			return outputData(cmd, fields)
		}
		printAdObjectFieldsTable(cmd, fields)
		return nil
	},
}

// ---- Creative sub-group ----

var adCreativeCmd = &cobra.Command{
	Use:   "creative",
	Short: "广告素材",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			appStr, _ := cmd.Flags().GetString("app")
			if appStr == "" {
				return cliErr.NewCLIError("MISSING_APP", "必须指定 --app（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
			}
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}

			materialID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return cliErr.NewCLIError("INVALID_MATERIAL_ID", fmt.Sprintf("无效的素材 ID: %s", args[0]))
			}

			withRelations, _ := cmd.Flags().GetBool("with-relations")

			ctx, cancel := newSignalCtx()
			defer cancel()

			adClient := client.NewAdClient()
			result, err := adClient.GetMaterialInfo(ctx, appId, materialID)
			if err != nil {
				return err
			}

			if quiet {
				return outputData(cmd, result)
			}
			if format == "json" {
				return outputData(cmd, result)
			}
			printAdMaterialDetail(cmd, result, withRelations, adClient)
			return nil
		}
		return cmd.Help()
	},
}

var adCreativeListCmd = &cobra.Command{
	Use:   "list",
	Short: "素材列表",
	Long:  `获取广告素材列表。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		appStr, _ := cmd.Flags().GetString("app")
		product, _ := cmd.Flags().GetUint32("product")
		materialType, _ := cmd.Flags().GetUint32("material-type")
		keyword, _ := cmd.Flags().GetString("keyword")
		page, pageSize := pagePagination(cmd)

		req := &client.ListMaterialsRequest{
			ProductID:    product,
			MaterialType: materialType,
			Keyword:      keyword,
			Page:         page,
			PageSize:     pageSize,
		}

		if appStr != "" {
			appId, err := client.ResolveAppID(appStr)
			if err != nil {
				return err
			}
			req.AppId = uint32(appId)
		}

		ctx, cancel := newSignalCtx()
		defer cancel()

		adClient := client.NewAdClient()
		result, err := adClient.ListMaterials(ctx, req)
		if err != nil {
			return err
		}

		if quiet {
			return outputData(cmd, result)
		}
		if format == "json" {
			return outputData(cmd, result)
		}
		printAdMaterialTable(cmd, result)
		return nil
	},
}

// ---- Output / Print functions ----

func printAdProductTable(cmd *cobra.Command, result *client.ListAdProductsResult) {
	out := cmd.OutOrStdout()

	if len(result.Products) == 0 {
		fmt.Fprintln(out, "无广告产品")
		return
	}

	fmt.Fprintf(out, "共 %d 条\n\n", result.Total)

	t := output.NewTableWriter(out)
	t.AppendHeader("ID", "名称", "关联平台", "同步状态", "已授权")

	for _, p := range result.Products {
		appNames := make([]string, 0, len(p.AppIds))
		for _, aid := range p.AppIds {
			appNames = append(appNames, appDisplayName(aid))
		}
		authorized := "-"
		if p.HasAuthorized {
			authorized = "Y"
		}
		t.AppendRow(
			fmt.Sprintf("%d", p.ID),
			p.Name,
			strings.Join(appNames, ", "),
			p.SyncState,
			authorized,
		)
	}

	t.Render()
}

func printAdAccountTable(cmd *cobra.Command, result *client.ListAdAccountsResult) {
	out := cmd.OutOrStdout()

	if len(result.Accounts) == 0 {
		fmt.Fprintln(out, "无广告账户")
		return
	}

	page, pageSize := uint32(1), uint32(20)
	fmt.Fprintf(out, "共 %d 条，第 %d/%d 页\n\n", result.Total, page, adTotalPages(result.Total, pageSize))

	t := output.NewTableWriter(out)
	t.AppendHeader("账户ID", "账户名", "平台", "授权状态", "余额", "近7日消耗", "近30日消耗")

	for _, a := range result.Accounts {
		t.AppendRow(
			a.AdAccountID,
			a.AdAccountName,
			appDisplayName(a.AppId),
			authStatusText(a.AuthStatus),
			a.Balance,
			a.Last7dCost,
			a.Last30dCost,
		)
	}

	t.Render()
}

func printAdAccountDetail(cmd *cobra.Command, account *client.AdAccount) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "账户 ID:     %s\n", account.AdAccountID)
	fmt.Fprintf(out, "账户名:     %s\n", account.AdAccountName)
	fmt.Fprintf(out, "平台:       %s\n", appDisplayName(account.AppId))
	fmt.Fprintf(out, "授权状态:   %s\n", authStatusText(account.AuthStatus))
	fmt.Fprintf(out, "余额:       %s\n", account.Balance)
	fmt.Fprintf(out, "近7日消耗:  %s\n", account.Last7dCost)
	fmt.Fprintf(out, "近30日消耗: %s\n", account.Last30dCost)
	if account.CompanyName != "" {
		fmt.Fprintf(out, "公司名:     %s\n", account.CompanyName)
	}
	if account.Industry != "" {
		fmt.Fprintf(out, "行业:       %s\n", account.Industry)
	}
	if account.ParentName != "" {
		fmt.Fprintf(out, "上级:       %s (ID: %d)\n", account.ParentName, account.ParentID)
	}
}

func printAdObjectTable(cmd *cobra.Command, result *client.ListAdObjectsResult) {
	out := cmd.OutOrStdout()

	if len(result.Objects) == 0 {
		fmt.Fprintln(out, "无数据")
		return
	}

	page, pageSize := uint32(1), uint32(20)
	fmt.Fprintf(out, "共 %d 条，第 %d/%d 页\n\n", result.Total, page, adTotalPages(result.Total, pageSize))

	t := output.NewTableWriter(out)
	t.AppendHeader("ID", "名称", "投放状态", "状态")

	for _, o := range result.Objects {
		t.AppendRow(
			fmt.Sprintf("%d", o.ObjID),
			o.Name,
			o.OptStatus,
			o.Status,
		)
	}

	t.Render()
}

func printAdObjectDetail(cmd *cobra.Command, obj *client.AdObject) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID:         %d\n", obj.ObjID)
	fmt.Fprintf(out, "类型:       %s\n", objTypeName(obj.ObjType))
	fmt.Fprintf(out, "名称:       %s\n", obj.Name)
	fmt.Fprintf(out, "投放状态:   %s\n", obj.OptStatus)
	fmt.Fprintf(out, "状态:       %s\n", obj.Status)
	fmt.Fprintf(out, "平台:       %s\n", appDisplayName(obj.AppId))
	fmt.Fprintf(out, "账户 ID:    %d\n", obj.AdAccountID)
	if obj.ParentID > 0 {
		fmt.Fprintf(out, "上级 ID:    %d\n", obj.ParentID)
	}

	// Parse and display attrs
	if obj.Attrs != "" {
		var attrs map[string]interface{}
		if err := json.Unmarshal([]byte(obj.Attrs), &attrs); err == nil {
			// Show well-known fields individually
			knownKeys := []string{
				"budget", "budgetMode", "marketingGoal", "optStatus",
				"deliveryMode", "bidType", "bid", "roiGoal",
				"deepBidType", "cpaBid", "startDate", "endDate",
			}
			shown := map[string]bool{}
			for _, key := range knownKeys {
				if val, ok := attrs[key]; ok {
					fmt.Fprintf(out, "%s: %v\n", padKey(key), val)
					shown[key] = true
				}
			}
			// Show remaining attrs as JSON block
			remaining := map[string]interface{}{}
			for k, v := range attrs {
				if !shown[k] {
					remaining[k] = v
				}
			}
			if len(remaining) > 0 {
				remainingJSON, _ := json.MarshalIndent(remaining, "", "  ")
				fmt.Fprintf(out, "其他属性:\n%s\n", string(remainingJSON))
			}
		} else {
			fmt.Fprintf(out, "属性:       %s\n", obj.Attrs)
		}
	}

	// Display metrics if present
	if obj.Metrics != "" && obj.Metrics != "{}" && obj.Metrics != "null" {
		fmt.Fprintf(out, "指标:       %s\n", obj.Metrics)
	}

	if obj.CreatedAt != "" {
		fmt.Fprintf(out, "创建时间:   %s\n", obj.CreatedAt)
	}
	if obj.UpdatedAt != "" {
		fmt.Fprintf(out, "更新时间:   %s\n", obj.UpdatedAt)
	}
}

func padKey(key string) string {
	maxLen := 12
	s := key + ":"
	for len(s) < maxLen {
		s += " "
	}
	// Add indent
	return "  " + s
}

func printAdMaterialTable(cmd *cobra.Command, result *client.ListMaterialsResult) {
	out := cmd.OutOrStdout()

	if len(result.Materials) == 0 {
		fmt.Fprintln(out, "无素材")
		return
	}

	page, pageSize := uint32(1), uint32(20)
	fmt.Fprintf(out, "共 %d 条，第 %d/%d 页\n\n", result.Total, page, adTotalPages(result.Total, pageSize))

	t := output.NewTableWriter(out)
	t.AppendHeader("ID", "名称", "平台", "类型", "评分", "标签")

	for _, m := range result.Materials {
		tags := strings.Join(m.Tags, ", ")
		t.AppendRow(
			fmt.Sprintf("%d", m.MaterialID),
			m.MaterialName,
			appDisplayName(m.AppId),
			materialTypeText(m.MaterialType),
			m.AverageRating,
			tags,
		)
	}

	t.Render()
}

func printAdMaterialDetail(cmd *cobra.Command, result *client.GetMaterialInfoResult, withRelations bool, adClient *client.AdClient) {
	out := cmd.OutOrStdout()
	m := result.Material

	fmt.Fprintf(out, "素材 ID:    %d\n", m.MaterialID)
	fmt.Fprintf(out, "名称:       %s\n", m.MaterialName)
	fmt.Fprintf(out, "平台:       %s\n", appDisplayName(m.AppId))
	fmt.Fprintf(out, "类型:       %s\n", materialTypeText(m.MaterialType))
	fmt.Fprintf(out, "评分:       %s\n", m.AverageRating)
	if len(m.Tags) > 0 {
		fmt.Fprintf(out, "标签:       %s\n", strings.Join(m.Tags, ", "))
	}
	if len(m.AiTags) > 0 {
		fmt.Fprintf(out, "AI 标签:    %s\n", strings.Join(m.AiTags, ", "))
	}
	if m.CoverUrl != "" {
		fmt.Fprintf(out, "封面 URL:   %s\n", m.CoverUrl)
	}
	if m.PlayUrl != "" {
		fmt.Fprintf(out, "播放 URL:   %s\n", m.PlayUrl)
	}
	if m.CreatedAt != "" {
		fmt.Fprintf(out, "创建时间:   %s\n", m.CreatedAt)
	}

	if withRelations {
		r := result.Relation
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "=== 关联链 ===")

		if r.OriginalMaterialID > 0 {
			name := fetchMaterialName(adClient, int(m.AppId), r.OriginalMaterialID)
			fmt.Fprintf(out, "原始素材:   %d (%s)\n", r.OriginalMaterialID, name)
		}
		if r.ParentMaterialID > 0 {
			name := fetchMaterialName(adClient, int(m.AppId), r.ParentMaterialID)
			fmt.Fprintf(out, "父素材:     %d (%s)\n", r.ParentMaterialID, name)
		}
		if len(r.DerivativeIDs) > 0 {
			fmt.Fprintf(out, "衍生素材:   ")
			for i, id := range r.DerivativeIDs {
				if i > 0 {
					fmt.Fprint(out, ", ")
				}
				name := fetchMaterialName(adClient, int(m.AppId), id)
				fmt.Fprintf(out, "%d (%s)", id, name)
			}
			fmt.Fprintln(out)
		}
		if len(r.FissionIDs) > 0 {
			fmt.Fprintf(out, "裂变素材:   ")
			for i, id := range r.FissionIDs {
				if i > 0 {
					fmt.Fprint(out, ", ")
				}
				name := fetchMaterialName(adClient, int(m.AppId), id)
				fmt.Fprintf(out, "%d (%s)", id, name)
			}
			fmt.Fprintln(out)
		}
	}
}

// fetchMaterialName fetches a material's name/type for relation display.
// Returns "获取失败" on error to avoid blocking detail output.
func fetchMaterialName(adClient *client.AdClient, appId int, materialID uint64) string {
	ctx, cancel := newSignalCtx()
	defer cancel()

	info, err := adClient.GetMaterialInfo(ctx, appId, materialID)
	if err != nil {
		return "获取失败"
	}
	return info.Material.MaterialName
}

func printAdPlatformSchemaTable(cmd *cobra.Command, schema *client.AdPlatformSchema) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "平台: %s (%s)\n\n", schema.AppName, appDisplayName(schema.AppId))

	if len(schema.Levels) > 0 {
		fmt.Fprintln(out, "层级结构:")
		t := output.NewTableWriter(out)
		t.AppendHeader("对象类型", "名称", "子层级类型")

		for _, l := range schema.Levels {
			childType := "-"
			if l.ChildObjType > 0 {
				childType = objTypeName(l.ChildObjType)
			}
			t.AppendRow(
				objTypeName(l.ObjType),
				l.Name,
				childType,
			)
		}
		t.Render()
	}

	if len(schema.FieldKeys) > 0 {
		fmt.Fprintf(out, "\n字段列表: %s\n", strings.Join(schema.FieldKeys, ", "))
	}
}

func printAdObjectFieldsTable(cmd *cobra.Command, fields []client.AdObjectField) {
	out := cmd.OutOrStdout()

	if len(fields) == 0 {
		fmt.Fprintln(out, "无字段定义")
		return
	}

	t := output.NewTableWriter(out)
	t.AppendHeader("key", "名称", "类型", "可筛选", "可排序")

	for _, f := range fields {
		filterable := "-"
		if f.Filterable {
			filterable = "Y"
		}
		sortable := "-"
		if f.Sortable {
			sortable = "Y"
		}
		t.AppendRow(f.Key, f.Name, f.DataType, filterable, sortable)
	}

	t.Render()
}

// ---- init() registration ----

func init() {
	rootCmd.AddCommand(adCmd)

	// Sub-command groups
	adCmd.AddCommand(adBrowseCmd)
	adCmd.AddCommand(adProductCmd)
	adCmd.AddCommand(adChannelCmd)
	adCmd.AddCommand(adAccountCmd)
	adCmd.AddCommand(adAdvertiserCmd)
	adCmd.AddCommand(adCampaignCmd)
	adCmd.AddCommand(adAdgroupCmd)
	adCmd.AddCommand(adAdsCmd)
	adCmd.AddCommand(adCreativeCmd)

	// Leaf commands under sub-groups
	adProductCmd.AddCommand(adProductListCmd)
	adChannelCmd.AddCommand(adChannelSchemaCmd)
	adAccountCmd.AddCommand(adAccountListCmd)
	adAdvertiserCmd.AddCommand(adAdvertiserListCmd)
	adCampaignCmd.AddCommand(adCampaignListCmd)
	adCampaignCmd.AddCommand(adCampaignFieldsCmd)
	adAdgroupCmd.AddCommand(adAdgroupListCmd)
	adAdgroupCmd.AddCommand(adAdgroupFieldsCmd)
	adAdsCmd.AddCommand(adAdsListCmd)
	adAdsCmd.AddCommand(adAdsFieldsCmd)
	adCreativeCmd.AddCommand(adCreativeListCmd)

	// Browse flags
	adBrowseCmd.Flags().Bool("reset", false, "清除行业缓存并重新检测")
	adBrowseCmd.Flags().String("app", "", "广告平台（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")

	// Product list flags
	adProductListCmd.Flags().String("keyword", "", "搜索关键词")
	adProductListCmd.Flags().Int("page", 1, "页码")
	adProductListCmd.Flags().Int("page-size", 20, "每页条数")

	// Account parent flags (for detail)
	adAccountCmd.Flags().Uint32("product", 0, "产品 ID（用于筛选账户详情）")
	// Account list flags
	adAccountListCmd.Flags().String("app", "", "广告平台（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adAccountListCmd.Flags().Uint32("product", 0, "产品 ID")
	adAccountListCmd.Flags().String("keyword", "", "搜索关键词")
	adAccountListCmd.Flags().Int("page", 1, "页码")
	adAccountListCmd.Flags().Int("page-size", 20, "每页条数")

	// Advertiser parent flags (for detail)
	adAdvertiserCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	// Advertiser list flags
	adAdvertiserListCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adAdvertiserListCmd.Flags().Uint32("product", 0, "产品 ID")
	adAdvertiserListCmd.Flags().String("keyword", "", "搜索关键词")
	adAdvertiserListCmd.Flags().Int("page", 1, "页码")
	adAdvertiserListCmd.Flags().Int("page-size", 20, "每页条数")

	// Campaign parent flags (for detail)
	adCampaignCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	// Campaign list flags
	adCampaignListCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adCampaignListCmd.Flags().Uint64("advertiser", 0, "广告主 ID（作为上级筛选）")
	adCampaignListCmd.Flags().Uint32("product", 0, "产品 ID")
	adCampaignListCmd.Flags().String("keyword", "", "搜索关键词")
	adCampaignListCmd.Flags().Int("page", 1, "页码")
	adCampaignListCmd.Flags().Int("page-size", 20, "每页条数")
	// Campaign fields flags
	adCampaignFieldsCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")

	// Adgroup parent flags (for detail)
	adAdgroupCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	// Adgroup list flags
	adAdgroupListCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adAdgroupListCmd.Flags().Uint64("advertiser", 0, "广告主 ID（作为上级筛选）")
	adAdgroupListCmd.Flags().Uint32("product", 0, "产品 ID")
	adAdgroupListCmd.Flags().String("keyword", "", "搜索关键词")
	adAdgroupListCmd.Flags().Int("page", 1, "页码")
	adAdgroupListCmd.Flags().Int("page-size", 20, "每页条数")
	// Adgroup fields flags
	adAdgroupFieldsCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")

	// Ads parent flags (for detail)
	adAdsCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	// Ads list flags
	adAdsListCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adAdsListCmd.Flags().Uint64("campaign", 0, "计划 ID（作为上级筛选）")
	adAdsListCmd.Flags().Uint64("adgroup", 0, "广告组 ID（作为上级筛选）")
	adAdsListCmd.Flags().Uint32("product", 0, "产品 ID")
	adAdsListCmd.Flags().String("keyword", "", "搜索关键词")
	adAdsListCmd.Flags().Int("page", 1, "页码")
	adAdsListCmd.Flags().Int("page-size", 20, "每页条数")
	// Ads fields flags
	adAdsFieldsCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")

	// Creative parent flags (for detail)
	adCreativeCmd.Flags().String("app", "", "广告平台（必填，1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adCreativeCmd.Flags().Bool("with-relations", false, "显示素材关联链")
	// Creative list flags
	adCreativeListCmd.Flags().String("app", "", "广告平台（1/5/6 或别名: oceanengine/oe, qianchuan/qc, tencent/tx）")
	adCreativeListCmd.Flags().Uint32("product", 0, "产品 ID")
	adCreativeListCmd.Flags().Uint32("material-type", 0, "素材类型（1=视频, 2=图片）")
	adCreativeListCmd.Flags().String("keyword", "", "搜索关键词")
	adCreativeListCmd.Flags().Int("page", 1, "页码")
	adCreativeListCmd.Flags().Int("page-size", 20, "每页条数")
}
