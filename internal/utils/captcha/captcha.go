package captcha

import (
	"context"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

// Store 实现验证码存储接口
// Implement captcha storage interface
type Store struct {
	redis  *redis.Client
	expire time.Duration
}

var _ base64Captcha.Store = (*Store)(nil)

// NewStore 创建验证码存储实例
// Create captcha storage instance
func NewStore(redis *redis.Client, expire time.Duration) *Store {
	return &Store{
		redis:  redis,
		expire: expire,
	}
}

// Set 设置验证码
// Set captcha
func (s *Store) Set(id string, value string) error {
	ctx := context.Background()
	return s.redis.Set(ctx, "captcha:"+id, value, s.expire).Err()
}

// Get 获取验证码
// Get captcha
func (s *Store) Get(id string, clear bool) string {
	ctx := context.Background()
	key := "captcha:" + id
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.redis.Del(ctx, key)
	}
	return val
}

// Verify 验证验证码
// Verify captcha
func (s *Store) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}

var store *Store

// InitStore 初始化验证码存储
// Initialize captcha storage
func InitStore(redis *redis.Client) {
	store = NewStore(redis, 5*time.Minute)
}

// Generate 生成验证码
// Generate captcha
// 生成验证码
func Generate() (id, b64s string, err error) {
	// 创建一个数字验证码驱动，参数分别为宽度、高度、字符数量、字符间距、字体大小
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	// 创建一个验证码实例，参数为驱动和存储
	c := base64Captcha.NewCaptcha(driver, store)
	// 生成验证码，返回验证码id、验证码内容、验证码图片、错误信息
	id, content, _, err := c.Generate()
	// 如果生成验证码出错，返回空字符串、空字符串、错误信息
	if err != nil {
		return "", "", err
	}
	// 返回验证码id、验证码内容、空错误信息
	return id, content, nil
}

// Verify 验证验证码
// Verify captcha
func Verify(id string, value string) bool {
	return store.Verify(id, value, true)
}
