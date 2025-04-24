package userservice

import (
	"go.uber.org/zap"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"gorm.io/gorm"
)

func ListUser(queryDto dto.ListUserReq, db *gorm.DB) *page.PageVo[domain.User] {

	var users []domain.User

	query := db

	if queryDto.Username != "" {
		query = query.Where("user_name like ?", "%"+queryDto.Username+"%")
	}

	if queryDto.Status != nil {
		query = query.Where("status = ?", queryDto.Status)
	}

	paginate, err := page.Paginate[domain.User](query, queryDto.PageRequest, &users)
	if err != nil {
		panic(err)
	}

	return paginate
}

func RefreshToken(refreshToken string, db *gorm.DB) (*vo.TokenPair, error) {
	// 验证刷新令牌
	claims, err := utils.VerifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	user := GetUserByUserId(claims.UserID, db)

	accessToken, refreshToken, err := utils.GenerateTokenPair(user.Id, user.Role)

	if err != nil {
		log.Error("generate token pair error", zap.Error(err))
		return nil, err
	}

	return &vo.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserByUserId 根据用户ID获取用户信息
func GetUserByUserId(userId int64, db *gorm.DB) *domain.User {
	var user domain.User
	result := db.Where("id = ?", userId).First(&user)
	if result.Error != nil {
		log.Error("get user by id error", zap.Error(result.Error))
		panic(result.Error)
	}
	return &user
}
