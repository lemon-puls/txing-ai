package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"go.uber.org/zap"
)

// 临时密钥
type TempCredential struct {
	SecretID     string
	SecretKey    string
	SessionToken string
	StartTime    int64
	ExpiredTime  int64
}

// COS 客户端
type COSClient struct {
	client *cos.Client
	config *global.CosConfig
}

// 上传文件
func (c *COSClient) PutFromURL(ctx context.Context, key string, url string) (string, error) {
	uploadResult, response, err := c.client.Object.PutFromURL(ctx, key, url, nil)
	if err != nil {
		log.Error("PutFromURL failed", zap.String("key", key), zap.String("url", url),
			zap.Any("uploadResult", uploadResult), zap.Any("response", response), zap.Error(err))
		return "", err
	}
	return uploadResult.Key, err
}

// 获取临时密钥
func (c *COSClient) GetTempCredential() (*TempCredential, error) {
	stsClient := sts.NewClient(
		// 通过环境变量获取密钥, os.Getenv 方法表示获取环境变量
		c.config.AccessKey, // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考https://cloud.tencent.com/document/product/598/37140
		c.config.SecretKey, // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考https://cloud.tencent.com/document/product/598/37140
		nil,
		// sts.Host("sts.internal.tencentcloudapi.com"), // 设置域名, 默认域名sts.tencentcloudapi.com
		// sts.Scheme("http"),      // 设置协议, 默认为https，公有云sts获取临时密钥不允许走http，特殊场景才需要设置http
	)
	// 策略概述 https://cloud.tencent.com/document/product/436/18023
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(time.Hour.Seconds()),
		Region:          c.config.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
					Action: []string{
						// 简单上传
						"name/cos:PostObject",
						"name/cos:PutObject",
						// 删除对象
						"name/cos:DeleteObject",
						// 查询对象
						"name/cos:GetObject",
						"name/cos:HeadObject",
						// 分片上传
						"name/cos:InitiateMultipartUpload",
						"name/cos:ListMultipartUploads",
						"name/cos:ListParts",
						"name/cos:UploadPart",
						"name/cos:CompleteMultipartUpload",
						// 列出对象
						"name/cos:GetBucket",
						"name/cos:HeadBucket",
					},
					Effect: "allow",
					Resource: []string{
						// 这里改成允许的路径前缀，可以根据自己网站的用户登录态判断允许上传的具体路径，例子： a.jpg 或者 a/* 或者 * (使用通配符*存在重大安全风险, 请谨慎评估使用)
						// 存储桶的命名格式为 BucketName-APPID，此处填写的 bucket 必须为此格式
						//"qcs::cos:" + region + ":uid/" + appid + ":" + bucket + "/exampleobject",
						"*",
					},
					// 开始构建生效条件 condition
					// 关于 condition 的详细设置规则和COS支持的condition类型可以参考https://cloud.tencent.com/document/product/436/71306
					//Condition: map[string]map[string]interface{}{
					//	"ip_equal": map[string]interface{}{
					//		"qcs:ip": []string{
					//			"*",
					//		},
					//	},
					//},
				},
			},
		},
	}

	// 请求临时密钥
	resp, err := stsClient.GetCredential(opt)
	if err != nil {
		log.Error("GetCredential failed", zap.Error(err))
		return nil, err
	}

	return &TempCredential{
		SecretID:     resp.Credentials.TmpSecretID,
		SecretKey:    resp.Credentials.TmpSecretKey,
		SessionToken: resp.Credentials.SessionToken,
		StartTime:    time.Now().Unix(),
		ExpiredTime:  time.Now().Add(2 * time.Hour).Unix(),
	}, nil
}

// 创建 COS 客户端
func NewCOSClient(config *global.CosConfig) (*COSClient, error) {
	// 存储桶URL
	//bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region))
	bucketURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse bucket url failed: %v", err)
	}

	// 服务URL
	serviceURL, err := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", config.Region))
	if err != nil {
		return nil, fmt.Errorf("parse service url failed: %v", err)
	}

	// 初始化客户端
	b := &cos.BaseURL{BucketURL: bucketURL, ServiceURL: serviceURL}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKey,
			SecretKey: config.SecretKey,
		},
	})

	return &COSClient{
		client: client,
		config: config,
	}, nil
}

// 生成文件上传预签名URL
func (c *COSClient) GenerateUploadPresignedURL(key string) (string, error) {
	ctx := context.Background()

	// 获取临时密钥
	//tempCred, err := c.GetTempCredential()
	//if err != nil {
	//	return "", fmt.Errorf("get temp credential failed: %v", err)
	//}

	//tak := tempCred.SecretID
	//tsk := tempCred.SecretKey
	//token := &URLToken{
	//	SessionToken: tempCred.SessionToken,
	//}
	//u, _ := url.Parse("https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com")
	//b := &cos.BaseURL{BucketURL: u}
	//c := cos.NewClient(b, &http.Client{})

	// 方法2 通过 tag 设置 x-cos-security-token
	// 获取预签名
	presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodPut, key, c.config.AccessKey,
		c.config.SecretKey, c.config.SignExpire*time.Second, nil)
	if err != nil {
		log.Error("GetPresignedURL failed", zap.Error(err))
		return "", err
	}

	return presignedURL.String(), nil
}

// 通过 tag 的方式，用户可以将请求参数或者请求头部放进签名中。
type URLToken struct {
	SessionToken string `url:"x-cos-security-token,omitempty" header:"-"`
}

// 生成文件下载预签名URL
func (c *COSClient) GenerateDownloadPresignedURL(key string) (string, error) {

	if key == "" {
		return "", fmt.Errorf("key is empty")
	}

	// 如果非 COS 对象存储对象，则直接返回原路径
	if strings.HasPrefix(key, "https://") {
		return key, nil
	}

	ctx := context.Background()

	// 获取临时密钥
	//tempCred, err := c.GetTempCredential()
	//if err != nil {
	//	return "", fmt.Errorf("get temp credential failed: %v", err)
	//}

	// 使用临时密钥生成预签名URL
	//token := &URLToken{
	//	SessionToken: tempCred.SessionToken,
	//}

	// 注意：这里使用 GET 方法而不是 PUT 方法
	presignedURL, err := c.client.Object.GetPresignedURL(
		ctx,
		http.MethodGet, // 使用 GET 方法进行下载
		key,
		c.config.AccessKey, // TODO 经测试发现，这里若使用临时密钥，当使用的是自定义域名时，会出现签名错误问题，而使用子账号密钥则不会出现此问题，后面再进一步排查原因。
		c.config.SecretKey,
		c.config.SignExpire*time.Second,
		nil,
	)
	if err != nil {
		log.Error("GetPresignedURL failed", zap.Error(err))
		return "", err
	}

	// 把前面部分替换为 CDN 加速地址
	url := strings.Replace(presignedURL.String(), c.config.BaseURL, c.config.CDNURL, 1)

	return url, nil
}

// 删除文件
func (c *COSClient) DeleteObject(key string) error {
	ctx := context.Background()

	_, err := c.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("delete object failed: %v", err)
	}

	return nil
}

// 批量删除文件
func (c *COSClient) DeleteObjects(keys []string) error {
	ctx := context.Background()

	obs := []cos.Object{}
	for _, key := range keys {
		obs = append(obs, cos.Object{Key: key})
	}

	opt := &cos.ObjectDeleteMultiOptions{
		Objects: obs,
		Quiet:   true,
	}

	_, _, err := c.client.Object.DeleteMulti(ctx, opt)
	if err != nil {
		return fmt.Errorf("batch delete objects failed: %v", err)
	}

	return nil
}

// 检查文件是否存在
func (c *COSClient) IsObjectExist(key string) (bool, error) {
	ctx := context.Background()

	_, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("check object exist failed: %v", err)
	}

	return true, nil
}

// 获取文件大小
func (c *COSClient) GetObjectSize(key string) (int64, error) {
	ctx := context.Background()

	resp, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		return 0, fmt.Errorf("get object size failed: %v", err)
	}

	return resp.ContentLength, nil
}

// ConvertObjectPath 将文件路径转换为存储至数据库中的路径
// 例如：https://www.example.com/exampleobject/1745647348066-761.jpg?q-sign-algorithm=sha1&q-ak=AKIDc6MDsKXWGm38z432-7823gGhv9D4jANM7e094m
// 转换为：exampleobject/1745647348066-761.jpg
func (c *COSClient) ConvertObjectPath(path string) string {
	// 只有是本系统 COS 的地址才需要转换，其他的地址不转换
	if !strings.HasPrefix(path, c.config.CDNURL) {
		return path
	}

	if path == "" {
		return ""
	}
	u, _ := url.Parse(path)
	return u.Path[1:]
}

// ConvertSliceFieldToPresignedURL 将切片中每个元素的指定字段转换为预签名URL
// slice: 要处理的切片
// fieldName: 要转换的字段名
// client: COS客户端实例
func ConvertSliceFieldToPresignedURL[T any](slice []T, fieldName string, client *COSClient) []T {
	if len(slice) == 0 {
		return slice
	}

	// 使用反射获取字段
	for i := range slice {
		item := &slice[i]
		v := reflect.ValueOf(item).Elem()

		// 获取字段
		field := v.FieldByName(fieldName)
		if !field.IsValid() || field.Kind() != reflect.String {
			continue
		}

		// 获取字段值
		value := field.String()
		if value == "" {
			continue
		}

		// 生成预签名URL
		presignedURL, err := client.GenerateDownloadPresignedURL(value)
		if err != nil {
			log.Error("generate presigned url failed",
				zap.String("field", fieldName),
				zap.String("value", value),
				zap.Error(err))
			continue
		}

		// 设置新值
		field.SetString(presignedURL)
	}

	return slice
}
