package utils

import (
	"errors"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID int64 `json:"userId"`
	Role   int8  `json:"role"`
	jwt.RegisteredClaims
}

var secret string

// access token 过期时间 默认 5 分钟
var accessTokenExpire = 5 * time.Minute

// refresh token 过期时间 默认 7 天
var refreshTokenExpire = 7 * 24 * time.Hour

func InitJwtSecret(authConfig *global.AuthConfig) {
	if authConfig == nil {
		panic("auth config is nil")
	}
	secret = authConfig.JwtSecret
	if authConfig.AccessTokenExpire > 0 {
		accessTokenExpire = authConfig.AccessTokenExpire * time.Second
	}
	if authConfig.RefreshTokenExpire > 0 {
		refreshTokenExpire = authConfig.RefreshTokenExpire * time.Second
	}
}

func generateToken(userId int64, role int8, expire time.Duration) (string, error) {
	if secret == "" {
		log.Error("jwt secret is empty")
		return "", errors.New("jwt secret is empty")
	}
	mySigningKey := []byte(secret)
	// Create the claims
	claims := CustomClaims{
		userId,
		role,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),             // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),             // 生效时间
			Issuer:    "txing-ai",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mySigningKey)
	return tokenString, err
}

func AccessToken(userId int64, role int8) (string, error) {
	return generateToken(userId, role, accessTokenExpire)
}

func RefreshToken(userId int64, role int8) (string, error) {
	return generateToken(userId, role, refreshTokenExpire)
}

// GenerateTokenPair 生成 access token 和 refresh token
func GenerateTokenPair(userId int64, role int8) (accessToken string, refreshToken string, err error) {
	accessToken, err = AccessToken(userId, role)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = RefreshToken(userId, role)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func VerifyToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("解析 token 失败")
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		//fmt.Printf("%v %v", claims.UserID, claims.RegisteredClaims.Issuer)
		return claims, nil
	}
	return nil, errors.New("token 不合法")
}
