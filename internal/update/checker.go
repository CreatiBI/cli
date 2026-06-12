package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"github.com/CreatiBI/cli/internal/config"
)

const (
	// npm registry API URL
	npmRegistryURL = "https://registry.npmjs.org/@creatibi/cbi-cli/latest"
	// 检查间隔（24小时）
	checkInterval = 24 * time.Hour
	// HTTP 请求超时
	httpTimeout = 5 * time.Second
)

// UpdateInfo 更新信息
type UpdateInfo struct {
	LatestVersion  string
	CurrentVersion string
	HasUpdate      bool
}

// npmPackage npm 包响应结构
type npmPackage struct {
	Version string `json:"version"`
}

// CheckAndNotify 检查更新并输出提示（静默模式，失败不报错）
func CheckAndNotify(currentVersion string) {
	// 静默模式：所有错误都被忽略，不影响主命令执行
	info, err := checkUpdate(currentVersion)
	if err != nil {
		// 静默失败，下次继续尝试
		return
	}

	if info.HasUpdate {
		printUpdateNotice(info)
	}
}

// checkUpdate 执行更新检查逻辑
func checkUpdate(currentVersion string) (*UpdateInfo, error) {
	// 检查缓存：是否需要重新查询
	lastChecked := config.GetUpdateLastCheckedAt()
	now := time.Now()

	// 如果 24 小时内已检查，使用缓存结果
	if !lastChecked.IsZero() && now.Sub(lastChecked) < checkInterval {
		cachedVersion := config.GetUpdateLatestVersion()
		if cachedVersion != "" {
			hasUpdate := compareVersions(cachedVersion, currentVersion)
			return &UpdateInfo{
				LatestVersion:  cachedVersion,
				CurrentVersion: currentVersion,
				HasUpdate:      hasUpdate,
			}, nil
		}
	}

	// 查询 npm registry
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	latestVersion, err := fetchLatestVersion(ctx)
	if err != nil {
		return nil, err
	}

	// 保存缓存
	config.SetUpdateLastCheckedAt(now)
	config.SetUpdateLatestVersion(latestVersion)

	hasUpdate := compareVersions(latestVersion, currentVersion)
	return &UpdateInfo{
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		HasUpdate:      hasUpdate,
	}, nil
}

// fetchLatestVersion 从 npm registry 获取最新版本
func fetchLatestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, npmRegistryURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned status %d", resp.StatusCode)
	}

	var pkg npmPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return "", err
	}

	return pkg.Version, nil
}

// compareVersions 比较版本号，返回 true 表示 latest > current
// 开发版本号可能包含 git 偏移（如 0.2.1-3-gc255d1d），需要先剥离再比较
func compareVersions(latest, current string) bool {
	// 剥离 git describe 的偏移后缀（如 "-3-gc255d1d" 或 "-dirty"）
	// 只保留基础版本号用于比较
	currentBase := stripGitDescribeSuffix(current)
	latestBase := stripGitDescribeSuffix(latest)

	// semver 需要 "v" 前缀
	vLatest := "v" + strings.TrimPrefix(latestBase, "v")
	vCurrent := "v" + strings.TrimPrefix(currentBase, "v")

	// 确保版本号格式正确
	if !semver.IsValid(vLatest) || !semver.IsValid(vCurrent) {
		return false
	}

	return semver.Compare(vLatest, vCurrent) > 0
}

// stripGitDescribeSuffix 剥离 git describe 生成的偏移后缀
// 例如：0.2.1-3-gc255d1d-dirty → 0.2.1
// 例如：0.2.1 → 0.2.1（无变化）
func stripGitDescribeSuffix(version string) string {
	// git describe 格式：<tag>-<N>-g<hash>[-dirty]
	// 找到第一个 "-<数字>-g" 模式并截断
	for i := 0; i < len(version); i++ {
		if version[i] == '-' && i+1 < len(version) {
			// 检查是否是 git describe 的偏移部分（-N-g<hash>）
			rest := version[i+1:]
			// 偏移部分以数字开头，后面跟着 "-g"
			dashG := strings.Index(rest, "-g")
			if dashG > 0 {
				// 验证偏移数字部分
				offsetPart := rest[:dashG]
				if isDigitString(offsetPart) {
					return version[:i]
				}
			}
		}
	}

	// 也处理 -dirty 后缀
	if strings.HasSuffix(version, "-dirty") {
		return strings.TrimSuffix(version, "-dirty")
	}

	return version
}

// isDigitString 检查字符串是否全为数字
func isDigitString(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// printUpdateNotice 输出更新提示
func printUpdateNotice(info *UpdateInfo) {
	fmt.Fprintln(os.Stderr, "")
	// 确保版本号显示只有一个 v 前缀
	latestDisplay := ensureVPrefix(info.LatestVersion)
	currentDisplay := ensureVPrefix(info.CurrentVersion)
	fmt.Fprintf(os.Stderr, "📌 提示: 有新版本可用 (%s)，当前版本 %s\n",
		latestDisplay, currentDisplay)
	fmt.Fprintln(os.Stderr, "   运行 npm update -g @creatibi/cbi-cli 更新")
}

// ensureVPrefix 确保版本号有且仅有一个 v 前缀
func ensureVPrefix(version string) string {
	v := strings.TrimPrefix(version, "v")
	return "v" + v
}
