package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/CreatiBI/cli/internal/client"
	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// configCmd 代表 config 命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  `管理 CLI 本地配置，包括初始化、查看配置等。`,
}

// configInitCmd 初始化配置
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化本地配置",
	Long: `初始化 CLI 本地配置，创建应用凭证配置文件。

支持两种模式：
1. 回调模式（默认）：本地浏览器授权，自动回传凭证
2. 设备码模式：远程浏览器授权，适用于 VPS/服务器环境

流程：
1. 选择初始化模式（回调/设备码）
2. 获取应用凭证（自动或手动）
3. 写入配置文件 ~/.cbi/config.json

首次使用需要：
1. 在 CreatiBI 开放平台创建应用
2. 获取 client_id 和 client_secret
3. 使用此命令初始化本地配置`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查是否需要强制覆盖
		newFlag, _ := cmd.Flags().GetBool("new")

		// 检查配置是否已存在
		if config.Exists() && !newFlag {
			return cliErr.NewCLIError("CONFIG_EXISTS", fmt.Sprintf("配置文件已存在: %s\n若需要重新初始化，请使用: cbi config init --new", config.GetConfigFile()))
		}

		// 检查是否指定设备码模式
		deviceFlag, _ := cmd.Flags().GetBool("device")

		// 选择初始化模式
		var credentials *Credential
		var err error

		if deviceFlag {
			// 直接使用设备码模式
			credentials, err = initWithDeviceCode(cmd)
		} else {
			// 交互式选择模式
			mode := selectConfigMode(cmd)
			switch mode {
			case "callback":
				credentials = tryAutoFetchCredentials(cmd)
				if credentials == nil {
					credentials = promptForCredentials(cmd)
				}
			case "device":
				credentials, err = initWithDeviceCode(cmd)
			default:
				return cliErr.NewCLIError("INVALID_MODE", "无效的模式选择")
			}
		}

		if err != nil {
			return err
		}

		if credentials == nil {
			return cliErr.NewCLIError("CREDENTIAL_FAILED", "获取凭证失败")
		}

		// 写入配置
		cfg := &config.AppConfig{
			BaseURL:          credentials.BaseURL,
			ClientID:         credentials.ClientID,
			ClientSecret:     credentials.ClientSecret,
			DefaultWorkspace: credentials.DefaultWorkspace,
		}

		if err := config.SaveAppConfig(cfg); err != nil {
			return cliErr.WrapError(err, cliErr.ErrConfigWriteFailed)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 配置初始化成功")
		fmt.Fprintf(cmd.OutOrStdout(), "配置文件: %s\n", config.GetConfigFile())
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "下一步:")
		fmt.Fprintln(cmd.OutOrStdout(), "  cbi auth login  # 使用 OAuth 登录")
		return nil
	},
}

// selectConfigMode 选择配置初始化模式
func selectConfigMode(cmd *cobra.Command) string {
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "请选择配置初始化模式:")
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "  1. 回调模式 (推荐) - 本地浏览器授权，自动回传凭证")
	fmt.Fprintln(cmd.OutOrStdout(), "  2. 设备码模式     - 远程浏览器授权，适用于 VPS/服务器")
	fmt.Fprintln(cmd.OutOrStdout(), "")

	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(cmd.OutOrStdout(), "请输入选项 [1/2]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1", "":
		return "callback"
	case "2":
		return "device"
	default:
		fmt.Fprintln(cmd.ErrOrStderr(), "无效选项，使用默认回调模式")
		return "callback"
	}
}

// initWithDeviceCode 使用设备码模式获取凭证
func initWithDeviceCode(cmd *cobra.Command) (*Credential, error) {
	// 创建上下文，支持 Ctrl+C 取消
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n取消初始化...")
		cancel()
	}()

	// 启动设备码流程
	oauthClient := client.NewOAuthClient(nil)
	tokenResp, err := oauthClient.StartCredentialDeviceCodeFlow(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为凭证格式
	return &Credential{
		ClientID:     tokenResp.AppID,
		ClientSecret: tokenResp.AppSecret,
		BaseURL:      config.GetBaseURL(),
	}, nil
}

// Credential 凭证信息
type Credential struct {
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	BaseURL          string `json:"base_url"`
	DefaultWorkspace string `json:"default_workspace"`
}

// tryAutoFetchCredentials 尝试自动获取凭证（回调模式）
func tryAutoFetchCredentials(cmd *cobra.Command) *Credential {
	fmt.Fprintln(cmd.OutOrStdout(), "正在打开 CreatiBI 开放平台...")
	fmt.Fprintln(cmd.OutOrStdout(), "")

	callbackPath := "/config/callback"

	// 尝试多个端口
	port := 8080
	maxPort := 8090
	var server *http.Server
	var callbackURL string

	for ; port <= maxPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		// 尝试绑定端口
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			// 端口可用
			callbackURL = fmt.Sprintf("http://localhost:%d%s", port, callbackPath)
			server = &http.Server{Addr: addr}
			break
		}
	}

	if server == nil {
		if verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "无法找到可用端口 (8080-8090)\n")
		}
		return nil
	}

	platformURL := fmt.Sprintf("https://open.creatibi.cn/page/cli?form=cli&callback=%s", callbackURL)

	// 创建上下文，支持 Ctrl+C 取消
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n取消初始化...")
		cancel()
	}()

	// 凭证接收通道
	credChan := make(chan *Credential, 1)
	errChan := make(chan error, 1)

	http.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// 解析凭证参数
		clientID := r.URL.Query().Get("client_id")
		clientSecret := r.URL.Query().Get("client_secret")
		baseURL := r.URL.Query().Get("base_url")
		defaultWorkspace := r.URL.Query().Get("default_workspace")

		if clientID != "" && clientSecret != "" {
			credChan <- &Credential{
				ClientID:         clientID,
				ClientSecret:     clientSecret,
				BaseURL:          baseURL,
				DefaultWorkspace: defaultWorkspace,
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("✓ 凭证已接收，CLI 将自动完成配置"))
		} else {
			errChan <- fmt.Errorf("缺少凭证参数")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("错误: 缺少凭证参数"))
		}
	})

	// 在后台启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 打开浏览器
	if err := openBrowser(platformURL); err != nil {
		if verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "无法自动打开浏览器: %s\n", err.Error())
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "请在浏览器中完成应用创建:")
	fmt.Fprintf(cmd.OutOrStdout(), "  授权 URL: %s\n", platformURL)
	fmt.Fprintln(cmd.OutOrStdout(), "")

	// 倒计时显示
	remaining := 30
	done := make(chan bool, 1)

	// 启动倒计时显示 goroutine
	go func() {
		fmt.Fprintf(cmd.OutOrStdout(), "等待凭证回传... 剩余 %2d 秒 (按 Ctrl+C 取消)", remaining)
		for {
			select {
			case <-done:
				return
			case <-time.After(1 * time.Second):
				remaining--
				if remaining > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "\r等待凭证回传... 剩余 %2d 秒 (按 Ctrl+C 取消)", remaining)
				}
			}
		}
	}()

	// 等待凭证或超时
	select {
	case cred := <-credChan:
		done <- true
		shutdownServer(ctx, server)
		fmt.Fprintln(cmd.OutOrStdout(), "\r                                    ")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 自动获取凭证成功")
		if cred.BaseURL == "" {
			cred.BaseURL = "https://open.creatibi.cn"
		}
		return cred

	case err := <-errChan:
		done <- true
		shutdownServer(ctx, server)
		fmt.Fprintln(cmd.OutOrStdout(), "\r                                    ")
		if verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "自动获取失败: %s\n", err.Error())
		}
		return nil

	case <-ctx.Done():
		done <- true
		shutdownServer(ctx, server)
		fmt.Fprintln(cmd.OutOrStdout(), "\r                                    ")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "超时，切换到手动输入模式")
		return nil
	}
}

// shutdownServer 关闭服务器
func shutdownServer(ctx context.Context, server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}

// openBrowser 打开浏览器
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// promptForCredentials 手动输入凭证
func promptForCredentials(cmd *cobra.Command) *Credential {
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "初始化 CreatiBI CLI 配置")
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "请先在开放平台创建应用并获取凭证:")
	fmt.Fprintln(cmd.OutOrStdout(), "  平台地址: https://open.creatibi.cn")
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "创建应用后，请准备好以下信息:")
	fmt.Fprintln(cmd.OutOrStdout(), "  - client_id (应用 ID)")
	fmt.Fprintln(cmd.OutOrStdout(), "  - client_secret (应用密钥)")
	fmt.Fprintln(cmd.OutOrStdout(), "")

	reader := bufio.NewReader(os.Stdin)

	// 输入 client_id
	fmt.Fprint(cmd.OutOrStdout(), "请输入 client_id: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "错误: client_id 不能为空")
		return nil
	}

	// 输入 client_secret
	fmt.Fprint(cmd.OutOrStdout(), "请输入 client_secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)
	if clientSecret == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "错误: client_secret 不能为空")
		return nil
	}

	// 输入 base_url（可选，有默认值）
	fmt.Fprint(cmd.OutOrStdout(), "请输入 base_url (默认: https://open.creatibi.cn): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://open.creatibi.cn"
	}

	// 输入 default_workspace（可选）
	fmt.Fprint(cmd.OutOrStdout(), "请输入 default_workspace (可选，留空跳过): ")
	defaultWorkspace, _ := reader.ReadString('\n')
	defaultWorkspace = strings.TrimSpace(defaultWorkspace)

	return &Credential{
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		BaseURL:          baseURL,
		DefaultWorkspace: defaultWorkspace,
	}
}

// configShowCmd 显示当前配置
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Long:  `显示当前生效的配置信息（敏感字段脱敏显示）。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查配置是否存在
		if !config.Exists() {
			return cliErr.ErrConfigNotFound
		}

		// 读取配置
		cfg, err := config.LoadAppConfig()
		if err != nil {
			return err
		}

		// 显示配置（脱敏）
		fmt.Fprintln(cmd.OutOrStdout(), "当前配置:")
		fmt.Fprintf(cmd.OutOrStdout(), "  配置文件: %s\n", config.GetConfigFile())
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintf(cmd.OutOrStdout(), "  base_url:          %s\n", cfg.BaseURL)
		fmt.Fprintf(cmd.OutOrStdout(), "  client_id:         %s\n", cfg.ClientID)
		fmt.Fprintf(cmd.OutOrStdout(), "  client_secret:     %s\n", maskSecret(cfg.ClientSecret))
		if cfg.DefaultWorkspace != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  default_workspace: %s\n", cfg.DefaultWorkspace)
		}

		// 显示登录状态
		if cfg.APIKey != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "登录状态:")
			fmt.Fprintf(cmd.OutOrStdout(), "  access_token:      %s\n", maskSecret(cfg.APIKey))
			if cfg.RefreshToken != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  refresh_token:     %s\n", maskSecret(cfg.RefreshToken))
			}
			if !cfg.TokenExpiresAt.IsZero() {
				fmt.Fprintf(cmd.OutOrStdout(), "  token_expires_at:  %s\n", cfg.TokenExpiresAt.Format("2006-01-02 15:04:05"))
			}
		}

		// JSON 格式输出
		if format == "json" {
			outputConfigJSON(cmd, cfg)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)

	// init 命令参数
	configInitCmd.Flags().Bool("new", false, "强制重新初始化（覆盖已有配置）")
	configInitCmd.Flags().Bool("device", false, "使用设备码模式（适用于 VPS/服务器）")
}

// maskSecret 脱敏显示敏感字段（显示前后缀）
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// outputConfigJSON JSON 格式输出配置（脱敏）
func outputConfigJSON(cmd *cobra.Command, cfg *config.AppConfig) {
	output := map[string]interface{}{
		"config_file":       config.GetConfigFile(),
		"base_url":          cfg.BaseURL,
		"client_id":         cfg.ClientID,
		"client_secret":     maskSecret(cfg.ClientSecret),
		"default_workspace": cfg.DefaultWorkspace,
	}

	if cfg.APIKey != "" {
		output["access_token"] = maskSecret(cfg.APIKey)
		if cfg.RefreshToken != "" {
			output["refresh_token"] = maskSecret(cfg.RefreshToken)
		}
		if !cfg.TokenExpiresAt.IsZero() {
			output["token_expires_at"] = cfg.TokenExpiresAt.Format("2006-01-02 15:04:05")
		}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
}
