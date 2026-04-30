package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
	"github.com/CreatiBI/cli/internal/update"
)

var (
	// 版本信息
	Version = "dev"

	// 全局标志
	cfgFile    string
	format     string
	outputFile string
	quiet      bool
	verbose    bool
)

// rootCmd 代表基础命令
var rootCmd = &cobra.Command{
	Use:   "cbi",
	Short: "CreatiBI CLI - 广告素材驱动的买量解决方案",
	Long: `CBI CLI 是 CreatiBI 的命令行工具，用于广告素材驱动的买量解决方案。

支持 OAuth 登录和素材库文件操作，通过命令行安全、便捷地管理创意素材。

示例:
  cbi config init                    # 初始化配置
  cbi auth login                     # OAuth 登录
  cbi repository list                # 查看素材库
  cbi repository file-create --repository-id 1 --file ./image.png`,
	Version: Version,
}

// Execute 执行根命令
func Execute() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, cliErr.FormatError(err, verbose))
		os.Exit(1)
	}

	// 更新检查（静默，不影响主命令）
	update.CheckAndNotify(Version)
}

func init() {
	// 配置文件标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认 ~/.cbi/config.json)")

	// 输出相关标志
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "输出格式: json|table")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "输出到文件")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "只输出数据，无日志")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细信息")

	// 自定义配置文件路径
	if cfgFile != "" {
		config.ConfigFile = cfgFile
	}
}

// outputData 输出数据（根据全局 format 参数）
func outputData(cmd *cobra.Command, data interface{}) error {
	w := cmd.OutOrStdout()

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return cliErr.NewCLIErrorWithDetail("OUTPUT_ERROR", "数据输出失败", err.Error())
	}
	return nil
}

// requireAuth 检查登录状态，未登录时返回友好错误
func requireAuth() error {
	if !config.IsLoggedIn() {
		return cliErr.ErrAuthRequired
	}
	return nil
}
