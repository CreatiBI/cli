package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/CreatiBI/cli/internal/config"
	cliErr "github.com/CreatiBI/cli/internal/errors"
)

// Platform appId constants
const (
	AppOceanengine = 1
	AppQianchuan   = 5
	AppTencent     = 6
)

// Object type constants
const (
	ObjTypeAdvertiser = 1 // 广告主
	ObjTypeCampaign   = 2 // 广告计划（仅巨量引擎）
	ObjTypeAdgroup    = 3 // 广告组（仅腾讯广告）
	ObjTypeAds        = 4 // 广告（巨量引擎/千川）/ 创意（腾讯广告）
	ObjTypeCreative   = 5 // 素材
)

// Platform alias -> appId mapping
var appAliasMap = map[string]int{
	"oceanengine": 1, "oe": 1, "juiliang": 1,
	"qianchuan": 5, "qc": 5,
	"tencent": 6, "tx": 6, "guangdian": 6,
}

// Object type alias -> objType mapping
var objTypeAliasMap = map[string]int{
	"advertiser": 1, "campaign": 2, "adgroup": 3, "ads": 4, "creative": 5,
}

// ResolveAppID parses appId from string, supports both numbers ("1") and
// aliases ("oceanengine", "oe", "juiliang", "qianchuan", "qc", "tencent",
// "tx", "guangdian"). Case-insensitive.
func ResolveAppID(s string) (int, error) {
	if v, err := strconv.Atoi(s); err == nil {
		return v, nil
	}
	if v, ok := appAliasMap[strings.ToLower(s)]; ok {
		return v, nil
	}
	return 0, cliErr.NewCLIErrorWithDetail("INVALID_APP_ID",
		fmt.Sprintf("无效的广告平台标识: %s", s),
		"支持: oceanengine/oe/juiliang(1), qianchuan/qc(5), tencent/tx/guangdian(6)")
}

// ResolveObjType parses objType from string, supports both numbers ("2") and
// aliases ("advertiser", "campaign", "adgroup", "ads", "creative").
// Case-insensitive.
func ResolveObjType(s string) (int, error) {
	if v, err := strconv.Atoi(s); err == nil {
		return v, nil
	}
	if v, ok := objTypeAliasMap[strings.ToLower(s)]; ok {
		return v, nil
	}
	return 0, cliErr.NewCLIErrorWithDetail("INVALID_OBJ_TYPE",
		fmt.Sprintf("无效的对象类型标识: %s", s),
		"支持: advertiser(1), campaign(2), adgroup(3), ads(4), creative(5)")
}

// AdProduct 广告产品
type AdProduct struct {
	ID            uint32   `json:"id"`
	Name          string   `json:"name"`
	Type          uint32   `json:"type"`
	AppIds        []uint32 `json:"appIds"`
	SyncState     string   `json:"syncState"`
	HasAuthorized bool     `json:"hasAuthorized"`
}

// AdAccount 广告账户
type AdAccount struct {
	ID             uint64 `json:"id"`
	AdAccountID    string `json:"adAccountId"`
	AdAccountName  string `json:"adAccountName"`
	AppId          uint32 `json:"appId"`
	AdAccountType  uint32 `json:"adAccountType"`
	AuthStatus     uint32 `json:"authStatus"`
	Active         uint32 `json:"active"`
	ActiveLevel    uint32 `json:"activeLevel"`
	Balance        string `json:"balance"`
	Last7dCost     string `json:"last7dCost"`
	Last30dCost    string `json:"last30dCost"`
	CompanyName    string `json:"companyName"`
	Industry       string `json:"industry"`
	ParentID       uint64 `json:"parentId"`
	ParentName     string `json:"parentName"`
}

// AdObject 广告对象（广告主/计划/广告组/广告）
type AdObject struct {
	ObjID       uint64 `json:"objId"`
	ObjType     uint32 `json:"objType"`
	AppId       uint32 `json:"appId"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	OptStatus   string `json:"optStatus"`
	AdAccountID uint64 `json:"adAccountId"`
	ParentID    uint64 `json:"parentId"`
	Attrs       string `json:"attrs"`
	Metrics     string `json:"metrics"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// AdMaterial 广告素材
type AdMaterial struct {
	MaterialID   uint64   `json:"materialId"`
	AppId        uint32   `json:"appId"`
	MaterialType uint32   `json:"materialType"`
	CoverUrl     string   `json:"coverUrl"`
	PlayUrl      string   `json:"playUrl"`
	MaterialName string   `json:"materialName"`
	HasAd        bool     `json:"hasAd"`
	AverageRating string  `json:"averageRating"`
	Tags         []string `json:"tags"`
	AiTags       []string `json:"aiTags"`
	CreatedAt    string   `json:"createdAt"`
}

// AdMaterialRelation 素材关联链
type AdMaterialRelation struct {
	OriginalMaterialID uint64   `json:"originalMaterialId"`
	DerivativeIDs      []uint64 `json:"derivativeIds"`
	FissionIDs         []uint64 `json:"fissionIds"`
	ParentMaterialID   uint64   `json:"parentMaterialId"`
}

// AdPlatformLevel 平台层级
type AdPlatformLevel struct {
	ObjType      uint32 `json:"objType"`
	Name         string `json:"name"`
	ChildObjType uint32 `json:"childObjType"`
}

// AdPlatformSchema 平台 Schema
type AdPlatformSchema struct {
	AppId     uint32            `json:"appId"`
	AppName   string            `json:"appName"`
	Levels    []AdPlatformLevel `json:"levels"`
	FieldKeys []string          `json:"fieldKeys"`
}

// AdObjectField 字段元数据
type AdObjectField struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	DataType   string `json:"dataType"`
	IsAttr     bool   `json:"isAttr"`
	IsMetric   bool   `json:"isMetric"`
	Filterable bool   `json:"filterable"`
	Sortable   bool   `json:"sortable"`
}

// ListAdProductsRequest 广告产品列表请求
type ListAdProductsRequest struct {
	Keyword  string
	Page     uint32
	PageSize uint32
}

// ListAdAccountsRequest 广告账户列表请求
type ListAdAccountsRequest struct {
	ProductID uint32
	AppId     uint32
	Keyword   string
	Page      uint32
	PageSize  uint32
}

// ListAdObjectsRequest 广告对象列表请求
type ListAdObjectsRequest struct {
	AppId     uint32
	ObjType   uint32
	ProductID uint32
	ParentID  uint64
	Keyword   string
	Page      uint32
	PageSize  uint32
}

// ListMaterialsRequest 素材列表请求
type ListMaterialsRequest struct {
	ProductID    uint32
	AppId        uint32
	MaterialType uint32
	Keyword      string
	Page         uint32
	PageSize     uint32
}

// ListAdProductsResult 广告产品列表结果
type ListAdProductsResult struct {
	Products         []AdProduct `json:"products"`
	Total            uint32     `json:"total"`
	CanCreateProduct bool       `json:"canCreateProduct"`
	IsEcomIndustry   bool       `json:"isEcomIndustry"`
}

// ListAdAccountsResult 广告账户列表结果
type ListAdAccountsResult struct {
	Accounts []AdAccount `json:"accounts"`
	Total    uint32     `json:"total"`
}

// ListAdObjectsResult 广告对象列表结果
type ListAdObjectsResult struct {
	Objects []AdObject `json:"objects"`
	Total   uint32    `json:"total"`
}

// ListMaterialsResult 素材列表结果
type ListMaterialsResult struct {
	Materials []AdMaterial `json:"materials"`
	Total     uint32      `json:"total"`
}

// GetMaterialInfoResult 素材详情结果
type GetMaterialInfoResult struct {
	Material AdMaterial        `json:"material"`
	Relation AdMaterialRelation `json:"relation"`
}

// AdClient 广告 API 客户端
type AdClient struct {
	client *resty.Client
}

// NewAdClient 创建广告客户端
func NewAdClient() *AdClient {
	baseURL := config.GetBaseURL()
	return &AdClient{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(30 * time.Second),
	}
}

// ListAdProducts 获取广告产品列表
func (c *AdClient) ListAdProducts(ctx context.Context, req *ListAdProductsRequest) (*ListAdProductsResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.Page > 0 {
		body["page"] = req.Page
	}
	if req.PageSize > 0 {
		body["pageSize"] = req.PageSize
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/product/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	// 处理 500 错误（可能 token 过期）
	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_PRODUCT_LIST_ERROR",
			fmt.Sprintf("获取广告产品列表失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	listResult := &ListAdProductsResult{
		Total:            uint32(data.Get("total").Int()),
		CanCreateProduct: data.Get("canCreateProduct").Bool(),
		IsEcomIndustry:   data.Get("isEcomIndustry").Bool(),
	}

	data.Get("products").ForEach(func(_, value gjson.Result) bool {
		product := AdProduct{
			ID:            uint32(value.Get("id").Int()),
			Name:          value.Get("name").String(),
			Type:          uint32(value.Get("type").Int()),
			SyncState:     value.Get("syncState").String(),
			HasAuthorized: value.Get("hasAuthorized").Bool(),
		}
		value.Get("appIds").ForEach(func(_, appId gjson.Result) bool {
			product.AppIds = append(product.AppIds, uint32(appId.Int()))
			return true
		})
		listResult.Products = append(listResult.Products, product)
		return true
	})

	return listResult, nil
}

// GetAdPlatformSchema 获取广告平台 Schema
func (c *AdClient) GetAdPlatformSchema(ctx context.Context, appId int) (*AdPlatformSchema, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"appId": appId,
		}).
		Post("/openapi/v1/ad/platform/schema")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_PLATFORM_SCHEMA_ERROR",
			fmt.Sprintf("获取广告平台 Schema 失败 (%d)", codeVal), message)
	}

	schemaData := result.Get("data.schema")
	schema := &AdPlatformSchema{
		AppId:   uint32(schemaData.Get("appId").Int()),
		AppName: schemaData.Get("appName").String(),
	}

	schemaData.Get("levels").ForEach(func(_, level gjson.Result) bool {
		schema.Levels = append(schema.Levels, AdPlatformLevel{
			ObjType:      uint32(level.Get("objType").Int()),
			Name:         level.Get("name").String(),
			ChildObjType: uint32(level.Get("childObjType").Int()),
		})
		return true
	})

	schemaData.Get("fieldKeys").ForEach(func(_, key gjson.Result) bool {
		schema.FieldKeys = append(schema.FieldKeys, key.String())
		return true
	})

	return schema, nil
}

// ListAdAccounts 获取广告账户列表
func (c *AdClient) ListAdAccounts(ctx context.Context, req *ListAdAccountsRequest) (*ListAdAccountsResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{}
	if req.ProductID > 0 {
		body["productId"] = req.ProductID
	}
	if req.AppId > 0 {
		body["appId"] = req.AppId
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.Page > 0 {
		body["page"] = req.Page
	}
	if req.PageSize > 0 {
		body["pageSize"] = req.PageSize
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/account/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_ACCOUNT_LIST_ERROR",
			fmt.Sprintf("获取广告账户列表失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	listResult := &ListAdAccountsResult{
		Total: uint32(data.Get("total").Int()),
	}

	data.Get("accounts").ForEach(func(_, value gjson.Result) bool {
		listResult.Accounts = append(listResult.Accounts, AdAccount{
			ID:            uint64(value.Get("id").Int()),
			AdAccountID:   value.Get("adAccountId").String(),
			AdAccountName: value.Get("adAccountName").String(),
			AppId:         uint32(value.Get("appId").Int()),
			AdAccountType: uint32(value.Get("adAccountType").Int()),
			AuthStatus:    uint32(value.Get("authStatus").Int()),
			Active:        uint32(value.Get("active").Int()),
			ActiveLevel:   uint32(value.Get("activeLevel").Int()),
			Balance:       value.Get("balance").String(),
			Last7dCost:    value.Get("last7dCost").String(),
			Last30dCost:   value.Get("last30dCost").String(),
			CompanyName:   value.Get("companyName").String(),
			Industry:      value.Get("industry").String(),
			ParentID:      uint64(value.Get("parentId").Int()),
			ParentName:    value.Get("parentName").String(),
		})
		return true
	})

	return listResult, nil
}

// ListAdObjects 获取广告对象列表
func (c *AdClient) ListAdObjects(ctx context.Context, req *ListAdObjectsRequest) (*ListAdObjectsResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{}
	if req.AppId > 0 {
		body["appId"] = req.AppId
	}
	if req.ObjType > 0 {
		body["objType"] = req.ObjType
	}
	if req.ProductID > 0 {
		body["productId"] = req.ProductID
	}
	if req.ParentID > 0 {
		body["parentId"] = req.ParentID
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.Page > 0 {
		body["page"] = req.Page
	}
	if req.PageSize > 0 {
		body["pageSize"] = req.PageSize
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/object/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_OBJECT_LIST_ERROR",
			fmt.Sprintf("获取广告对象列表失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	listResult := &ListAdObjectsResult{
		Total: uint32(data.Get("total").Int()),
	}

	data.Get("objects").ForEach(func(_, value gjson.Result) bool {
		listResult.Objects = append(listResult.Objects, AdObject{
			ObjID:       uint64(value.Get("objId").Int()),
			ObjType:     uint32(value.Get("objType").Int()),
			AppId:       uint32(value.Get("appId").Int()),
			Name:        value.Get("name").String(),
			Status:      value.Get("status").String(),
			OptStatus:   value.Get("optStatus").String(),
			AdAccountID: uint64(value.Get("adAccountId").Int()),
			ParentID:    uint64(value.Get("parentId").Int()),
			Attrs:       value.Get("attrs").String(),
			Metrics:     value.Get("metrics").String(),
			CreatedAt:   value.Get("createdAt").String(),
			UpdatedAt:   value.Get("updatedAt").String(),
		})
		return true
	})

	return listResult, nil
}

// GetAdObjectDetail 获取广告对象详情
func (c *AdClient) GetAdObjectDetail(ctx context.Context, appId int, objType int, objID uint64) (*AdObject, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"appId":   appId,
			"objType": objType,
			"objId":   objID,
		}).
		Post("/openapi/v1/ad/object/detail")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_OBJECT_DETAIL_ERROR",
			fmt.Sprintf("获取广告对象详情失败 (%d)", codeVal), message)
	}

	objData := result.Get("data.object")
	obj := &AdObject{
		ObjID:       uint64(objData.Get("objId").Int()),
		ObjType:     uint32(objData.Get("objType").Int()),
		AppId:       uint32(objData.Get("appId").Int()),
		Name:        objData.Get("name").String(),
		Status:      objData.Get("status").String(),
		OptStatus:   objData.Get("optStatus").String(),
		AdAccountID: uint64(objData.Get("adAccountId").Int()),
		ParentID:    uint64(objData.Get("parentId").Int()),
		Attrs:       objData.Get("attrs").String(),
		Metrics:     objData.Get("metrics").String(),
		CreatedAt:   objData.Get("createdAt").String(),
		UpdatedAt:   objData.Get("updatedAt").String(),
	}

	return obj, nil
}

// ListMaterials 获取素材列表
func (c *AdClient) ListMaterials(ctx context.Context, req *ListMaterialsRequest) (*ListMaterialsResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	body := map[string]interface{}{}
	if req.ProductID > 0 {
		body["productId"] = req.ProductID
	}
	if req.AppId > 0 {
		body["appId"] = req.AppId
	}
	if req.MaterialType > 0 {
		body["materialType"] = req.MaterialType
	}
	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}
	if req.Page > 0 {
		body["page"] = req.Page
	}
	if req.PageSize > 0 {
		body["pageSize"] = req.PageSize
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/openapi/v1/ad/material/list")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_MATERIAL_LIST_ERROR",
			fmt.Sprintf("获取素材列表失败 (%d)", codeVal), message)
	}

	data := result.Get("data")
	listResult := &ListMaterialsResult{
		Total: uint32(data.Get("total").Int()),
	}

	data.Get("materials").ForEach(func(_, value gjson.Result) bool {
		material := AdMaterial{
			MaterialID:    uint64(value.Get("materialId").Int()),
			AppId:         uint32(value.Get("appId").Int()),
			MaterialType:  uint32(value.Get("materialType").Int()),
			CoverUrl:      value.Get("coverUrl").String(),
			PlayUrl:       value.Get("playUrl").String(),
			MaterialName:  value.Get("materialName").String(),
			HasAd:         value.Get("hasAd").Bool(),
			AverageRating: value.Get("averageRating").String(),
			CreatedAt:     value.Get("createdAt").String(),
		}
		value.Get("tags").ForEach(func(_, tag gjson.Result) bool {
			material.Tags = append(material.Tags, tag.String())
			return true
		})
		value.Get("aiTags").ForEach(func(_, tag gjson.Result) bool {
			material.AiTags = append(material.AiTags, tag.String())
			return true
		})
		listResult.Materials = append(listResult.Materials, material)
		return true
	})

	return listResult, nil
}

// GetMaterialInfo 获取素材详情
func (c *AdClient) GetMaterialInfo(ctx context.Context, appId int, materialID uint64) (*GetMaterialInfoResult, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"appId":      appId,
			"materialId": materialID,
		}).
		Post("/openapi/v1/ad/material/info")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_MATERIAL_INFO_ERROR",
			fmt.Sprintf("获取素材详情失败 (%d)", codeVal), message)
	}

	data := result.Get("data")

	// Parse material
	matData := data.Get("material")
	material := AdMaterial{
		MaterialID:    uint64(matData.Get("materialId").Int()),
		AppId:         uint32(matData.Get("appId").Int()),
		MaterialType:  uint32(matData.Get("materialType").Int()),
		CoverUrl:      matData.Get("coverUrl").String(),
		PlayUrl:       matData.Get("playUrl").String(),
		MaterialName:  matData.Get("materialName").String(),
		HasAd:         matData.Get("hasAd").Bool(),
		AverageRating: matData.Get("averageRating").String(),
		CreatedAt:     matData.Get("createdAt").String(),
	}
	matData.Get("tags").ForEach(func(_, tag gjson.Result) bool {
		material.Tags = append(material.Tags, tag.String())
		return true
	})
	matData.Get("aiTags").ForEach(func(_, tag gjson.Result) bool {
		material.AiTags = append(material.AiTags, tag.String())
		return true
	})

	// Parse relation
	relationData := data.Get("relation")
	relation := AdMaterialRelation{
		OriginalMaterialID: uint64(relationData.Get("originalMaterialId").Int()),
		ParentMaterialID:   uint64(relationData.Get("parentMaterialId").Int()),
	}
	relationData.Get("derivativeIds").ForEach(func(_, id gjson.Result) bool {
		relation.DerivativeIDs = append(relation.DerivativeIDs, uint64(id.Int()))
		return true
	})
	relationData.Get("fissionIds").ForEach(func(_, id gjson.Result) bool {
		relation.FissionIDs = append(relation.FissionIDs, uint64(id.Int()))
		return true
	})

	return &GetMaterialInfoResult{
		Material: material,
		Relation: relation,
	}, nil
}

// ListAdObjectFields 获取广告对象字段列表
func (c *AdClient) ListAdObjectFields(ctx context.Context, appId int, objType int) ([]AdObjectField, error) {
	accessToken := config.GetAPIKey()
	if accessToken == "" {
		return nil, cliErr.ErrAuthRequired
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("user-access-token", accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"appId":   appId,
			"objType": objType,
		}).
		Post("/openapi/v1/ad/object/fields")

	if err != nil {
		return nil, cliErr.WrapError(err, cliErr.ErrNetworkError)
	}

	if resp.StatusCode() == 500 {
		return nil, handle500Error(resp.Body())
	}

	result := gjson.ParseBytes(resp.Body())

	codeVal := result.Get("code").Int()
	if codeVal != 0 {
		message := result.Get("message").String()
		if isTokenExpiredError(codeVal, message) {
			return nil, cliErr.ErrTokenExpired
		}
		return nil, cliErr.NewCLIErrorWithDetail("AD_OBJECT_FIELDS_ERROR",
			fmt.Sprintf("获取广告对象字段列表失败 (%d)", codeVal), message)
	}

	fields := []AdObjectField{}
	result.Get("data.fields").ForEach(func(_, value gjson.Result) bool {
		fields = append(fields, AdObjectField{
			Key:        value.Get("key").String(),
			Name:       value.Get("name").String(),
			DataType:   value.Get("dataType").String(),
			IsAttr:     value.Get("isAttr").Bool(),
			IsMetric:   value.Get("isMetric").Bool(),
			Filterable: value.Get("filterable").Bool(),
			Sortable:   value.Get("sortable").Bool(),
		})
		return true
	})

	return fields, nil
}
