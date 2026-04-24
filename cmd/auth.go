package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// authCmd 代表 auth 命令
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "认证管理",
	Long:  `管理 CLI 认证状态，包括登录、查看身份、退出登录。`,
}

// authLoginCmd 登录命令（OAuth）
var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "OAuth 登录",
	Long: `使用 OAuth 协议登录 CreatiBI 平台。

流程：
1. CLI 启动本地回调服务器（端口 8080）
2. 自动打开浏览器访问授权页面
3. 用户在浏览器中完成授权
4. 服务端回调到 CLI 并返回授权码
5. CLI 用授权码换取 access_token
6. Token 存储在 ~/.cbi/config.json

前提条件：
  需要先使用 cbi config init 初始化应用凭证`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否已初始化配置
		if !config.Exists() {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 应用凭证未配置")
			fmt.Fprintln(cmd.ErrOrStderr(), "")
			fmt.Fprintln(cmd.ErrOrStderr(), "请先初始化配置:")
			fmt.Fprintln(cmd.ErrOrStderr(), "  cbi config init")
			os.Exit(1)
		}

		// 使用 OAuth 登录
		loginWithOAuth(cmd)
	},
}

// loginWithOAuth OAuth 登录
func loginWithOAuth(cmd *cobra.Command) {
	// 初始化 OAuth 客户端
	oauthClient := client.NewOAuthClient(nil)

	// 检查 client_secret
	if config.GetClientSecret() == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "错误: client_secret 未配置")
		fmt.Fprintln(cmd.ErrOrStderr(), "")
		fmt.Fprintln(cmd.ErrOrStderr(), "请重新初始化配置:")
		fmt.Fprintln(cmd.ErrOrStderr(), "  cbi config init --new")
		os.Exit(1)
	}

	// 创建上下文，支持 Ctrl+C 取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n取消登录...")
		cancel()
	}()

	// 启动 OAuth 流程
	if err := oauthClient.StartOAuthFlow(ctx); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		os.Exit(1)
	}

	// 登录成功后自动获取用户信息
	accessToken := config.GetAPIKey()
	userInfo, err := oauthClient.GetUserInfo(accessToken)
	if err == nil {
		fmt.Println()
		fmt.Println("用户信息:")
		fmt.Fprintf(cmd.OutOrStdout(), "  用户: %s\n", userInfo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  邮箱: %s\n", userInfo.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "  ID: %d\n", userInfo.ID)
	} else {
		if verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取用户信息失败: %s\n", err.Error())
		}
	}

	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "配置文件: %s\n", config.GetConfigFile())
	}
}

// authWhoamiCmd 查看当前身份
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "查看当前身份信息",
	Long:  `显示当前登录用户的身份信息。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否登录
		if !config.IsLoggedIn() {
			fmt.Fprintln(cmd.ErrOrStderr(), cliErr.ErrAuthRequired.Error())
			return
		}

		// 初始化 OAuth 客户端获取用户信息
		oauthClient := client.NewOAuthClient(nil)
		accessToken := config.GetAPIKey()

		// 获取用户信息
		userInfo, err := oauthClient.GetUserInfo(accessToken)
		if err != nil {
			// 获取用户信息失败
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "获取用户信息失败: %s\n", err.Error())
			}
			fmt.Fprintln(cmd.OutOrStdout(), "当前身份:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Access Token: %s\n", maskToken(accessToken))
			fmt.Fprintf(cmd.OutOrStdout(), "  配置文件: %s\n", config.GetConfigFile())
			return
		}

		// 显示用户信息
		fmt.Fprintln(cmd.OutOrStdout(), "当前身份:")
		fmt.Fprintf(cmd.OutOrStdout(), "  用户: %s\n", userInfo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  邅箱: %s\n", userInfo.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "  ID: %d\n", userInfo.ID)
		if userInfo.Avatar != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  头像: %s\n", userInfo.Avatar)
		}

		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "  Access Token: %s\n", maskToken(accessToken))
			fmt.Fprintf(cmd.OutOrStdout(), "  配置文件: %s\n", config.GetConfigFile())

			// Token 过期信息
			expiresAt := config.GetTokenExpiresAt()
			if !expiresAt.IsZero() {
				fmt.Fprintf(cmd.OutOrStdout(), "  Token 过期时间: %s\n", expiresAt.Format("2006-01-02 15:04:05"))
			}

			// Refresh token 信息
			refreshToken := config.GetRefreshToken()
			if refreshToken != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Refresh Token: %s\n", maskToken(refreshToken))
			}
		}
	},
}

// authLogoutCmd 退出登录
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	Long:  `清除本地存储的认证信息。`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 清除登录凭证
		if err := config.Clear(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "错误: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 已退出登录")
		fmt.Fprintf(cmd.OutOrStdout(), "配置文件: %s\n", config.GetConfigFile())
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authLogoutCmd)
}

// maskToken 隐藏 Token 的中间部分
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}