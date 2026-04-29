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
func compareVersions(latest, current string) bool {
	// semver 需要 "v" 前缀
	vLatest := "v" + strings.TrimPrefix(latest, "v")
	vCurrent := "v" + strings.TrimPrefix(current, "v")

	// 确保版本号格式正确
	if !semver.IsValid(vLatest) || !semver.IsValid(vCurrent) {
		return false
	}

	return semver.Compare(vLatest, vCurrent) > 0
}

// printUpdateNotice 输出更新提示
func printUpdateNotice(info *UpdateInfo) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "📌 提示: 有新版本可用 (v%s)，当前版本 v%s\n", info.LatestVersion, info.CurrentVersion)
	fmt.Fprintln(os.Stderr, "   运行 npm update -g @creatibi/cbi-cli 更新")
}