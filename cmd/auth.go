package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

支持两种登录模式：
1. 授权码模式（默认） - 本地浏览器授权，适合桌面环境
2. 设备码模式 - 远程浏览器授权，适合 VPS/服务器环境

前提条件：
  需要先使用 cbi config init 初始化应用凭证`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查是否已初始化配置
		if !config.Exists() {
			return cliErr.NewCLIError("CONFIG_NOT_FOUND", "应用凭证未配置，请先初始化: cbi config init")
		}

		// 选择登录模式
		loginMode := selectLoginMode(cmd)

		switch loginMode {
		case "device":
			return loginWithDeviceCode(cmd)
		default:
			return loginWithOAuth(cmd)
		}
	},
}

// selectLoginMode 选择登录模式
func selectLoginMode(cmd *cobra.Command) string {
	// 检查是否有 --device 参数
	useDevice, _ := cmd.Flags().GetBool("device")
	if useDevice {
		return "device"
	}

	// 检查环境变量
	if os.Getenv("CBI_LOGIN_MODE") == "device" {
		return "device"
	}

	// 交互式选择
	fmt.Println()
	fmt.Println("请选择登录模式:")
	fmt.Println()
	fmt.Println("  [1] 授权码模式 - 本地浏览器授权（适合桌面环境）")
	fmt.Println("  [2] 设备码模式 - 远程浏览器授权（适合 VPS/服务器）")
	fmt.Println()
	fmt.Print("请输入选项 (1/2): ")

	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)
	switch input {
	case "1", "":
		return "oauth"
	case "2":
		return "device"
	default:
		fmt.Fprintf(cmd.ErrOrStderr(), "无效选项: %s，使用默认授权码模式\n", input)
		return "oauth"
	}
}

// loginWithOAuth OAuth 登录（授权码模式）
func loginWithOAuth(cmd *cobra.Command) error {
	// 初始化 OAuth 客户端
	oauthClient := client.NewOAuthClient(nil)

	// 检查 client_secret
	if config.GetClientSecret() == "" {
		return cliErr.NewCLIError("CONFIG_INCOMPLETE", "client_secret 未配置，请重新初始化: cbi config init --new")
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
		return err
	}

	// 登录成功后自动获取用户信息
	printUserInfo(oauthClient)

	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "配置文件: %s\n", config.GetConfigFile())
	}
	return nil
}

// loginWithDeviceCode 设备码登录
func loginWithDeviceCode(cmd *cobra.Command) error {
	// 初始化 OAuth 客户端
	oauthClient := client.NewOAuthClient(nil)

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

	// 启动设备码流程
	if err := oauthClient.StartDeviceCodeFlow(ctx); err != nil {
		return err
	}

	// 登录成功后自动获取用户信息
	printUserInfo(oauthClient)

	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "配置文件: %s\n", config.GetConfigFile())
	}
	return nil
}

// printUserInfo 获取并打印用户信息
func printUserInfo(oauthClient *client.OAuthClient) {
	accessToken := config.GetAPIKey()
	userInfo, err := oauthClient.GetUserInfo(accessToken)
	if err == nil {
		fmt.Println()
		fmt.Println("用户信息:")
		fmt.Fprintf(os.Stdout, "  用户: %s\n", userInfo.Name)
		fmt.Fprintf(os.Stdout, "  酅箱: %s\n", userInfo.Email)
		fmt.Fprintf(os.Stdout, "  ID: %d\n", userInfo.ID)
	} else if verbose {
		fmt.Fprintf(os.Stderr, "获取用户信息失败: %s\n", err.Error())
	}
}

// authWhoamiCmd 查看当前身份
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "查看当前身份信息",
	Long:  `显示当前登录用户的身份信息。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		// 初始化 OAuth 客户端获取用户信息
		oauthClient := client.NewOAuthClient(nil)
		accessToken := config.GetAPIKey()

		// 获取用户信息
		userInfo, err := oauthClient.GetUserInfo(accessToken)
		if err != nil {
			if verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "获取用户信息失败: %s\n", err.Error())
			}
			fmt.Fprintln(cmd.OutOrStdout(), "当前身份:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Access Token: %s\n", maskToken(accessToken))
			fmt.Fprintf(cmd.OutOrStdout(), "  配置文件: %s\n", config.GetConfigFile())
			return nil
		}

		// 显示用户信息
		fmt.Fprintln(cmd.OutOrStdout(), "当前身份:")
		fmt.Fprintf(cmd.OutOrStdout(), "  用户: %s\n", userInfo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  酅箱: %s\n", userInfo.Email)
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
		return nil
	},
}

// authLogoutCmd 退出登录
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	Long:  `清除本地存储的认证信息。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return cliErr.WrapError(err, cliErr.NewCLIError("LOGOUT_FAILED", "退出登录失败"))
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✓ 已退出登录")
		fmt.Fprintf(cmd.OutOrStdout(), "配置文件: %s\n", config.GetConfigFile())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authLogoutCmd)

	// 登录命令参数
	authLoginCmd.Flags().Bool("device", false, "使用设备码模式登录（适合 VPS/服务器环境）")
}

// maskToken 隐藏 Token 的中间部分
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
