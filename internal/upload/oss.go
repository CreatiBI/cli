package upload

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	tos "github.com/volcengine/ve-tos-golang-sdk/tos"
	"github.com/tidwall/gjson"
)

// STSCredential STS 临时凭证
type STSCredential struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SecurityToken   string `json:"securityToken"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"` // TOS 用
	Expiration      string `json:"expiration"`
}

// OSSUploader OSS 上传器
type OSSUploader struct {
	credential *STSCredential
	ossPath    string
	region     string // cn 或 en
	storage    int    // 1=火山引擎 TOS, 2=阿里云 OSS
}

// NewOSSUploader 创建 OSS 上传器
func NewOSSUploader(uploadToken, ossPath, region string, storage int) (*OSSUploader, error) {
	// 解析 STS 凭证
	cred := parseSTSCredential(uploadToken)
	if cred == nil {
		return nil, fmt.Errorf("无法解析 STS 凭证")
	}

	return &OSSUploader{
		credential: cred,
		ossPath:    ossPath,
		region:     region,
		storage:    storage,
	}, nil
}

// parseSTSCredential 解析 STS 凭证 JSON
func parseSTSCredential(tokenJSON string) *STSCredential {
	result := gjson.Parse(tokenJSON)
	if !result.IsObject() {
		return nil
	}

	return &STSCredential{
		AccessKeyID:     result.Get("accessKeyId").String(),
		AccessKeySecret: result.Get("accessKeySecret").String(),
		SecurityToken:   result.Get("securityToken").String(),
		Bucket:          result.Get("bucket").String(),
		Endpoint:        result.Get("endpoint").String(),
		Region:          result.Get("region").String(),
		Expiration:      result.Get("expiration").String(),
	}
}

// Upload 上传文件到 OSS
func (u *OSSUploader) Upload(localPath string) (string, error) {
	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件名
	fileName := filepath.Base(localPath)
	// 构建完整的 OSS 路径
	ossFullPath := u.ossPath + fileName

	// 根据 storage 类型选择 SDK
	switch u.storage {
	case 1:
		// 火山引擎 TOS
		err = u.uploadToTOS(file, ossFullPath)
	case 2:
		// 阿里云 OSS
		err = u.uploadToOSS(file, ossFullPath)
	default:
		return "", fmt.Errorf("未知的存储类型: %d", u.storage)
	}

	if err != nil {
		return "", err
	}

	return ossFullPath, nil
}

// uploadToTOS 上传到火山引擎 TOS
func (u *OSSUploader) uploadToTOS(file *os.File, ossPath string) error {
	// 创建 TOS 客户端
	endpoint := u.credential.Endpoint
	if endpoint == "" {
		// 默认使用 cn-beijing
		endpoint = "tos-cn-beijing.volces.com"
	}

	// 创建凭证
	cred := tos.NewStaticCredentials(u.credential.AccessKeyID, u.credential.AccessKeySecret)
	cred.WithSecurityToken(u.credential.SecurityToken)

	// 确定 region
	region := u.credential.Region
	if region == "" {
		region = "cn-beijing"
	}

	// 创建客户端
	client, err := tos.NewClient(endpoint, tos.WithCredentials(cred), tos.WithRegion(region))
	if err != nil {
		return fmt.Errorf("创建 TOS 客户端失败: %w", err)
	}

	// 获取 bucket
	bucket, err := client.Bucket(u.credential.Bucket)
	if err != nil {
		return fmt.Errorf("获取 bucket 失败: %w", err)
	}

	// 上传对象
	ctx := context.Background()
	_, err = bucket.PutObject(ctx, ossPath, file)
	if err != nil {
		return fmt.Errorf("上传到 TOS 失败: %w", err)
	}

	return nil
}

// uploadToOSS 上传到阿里云 OSS
func (u *OSSUploader) uploadToOSS(file *os.File, ossPath string) error {
	// 创建 OSS 客户端
	client, err := oss.New(u.credential.Endpoint, u.credential.AccessKeyID, u.credential.AccessKeySecret,
		oss.SecurityToken(u.credential.SecurityToken))
	if err != nil {
		return fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	// 获取 bucket
	bucket, err := client.Bucket(u.credential.Bucket)
	if err != nil {
		return fmt.Errorf("获取 bucket 失败: %w", err)
	}

	// 上传对象
	err = bucket.PutObject(ossPath, file)
	if err != nil {
		return fmt.Errorf("上传到 OSS 失败: %w", err)
	}

	return nil
}

// ComputeMD5 计算文件的 MD5 hash
func ComputeMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算 MD5 失败: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ComputeMD5Batch 批量计算文件 MD5
func ComputeMD5Batch(filePaths []string) (map[string]string, error) {
	hashes := make(map[string]string)
	for _, path := range filePaths {
		hash, err := ComputeMD5(path)
		if err != nil {
			return nil, fmt.Errorf("计算 %s MD5 失败: %w", path, err)
		}
		hashes[path] = hash
	}
	return hashes, nil
}