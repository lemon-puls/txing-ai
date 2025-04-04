package user

import (
	"gorm.io/gorm"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils/page"
)

func ListUser(queryDto dto.UserListRequest, db *gorm.DB) *page.PageVo[domain.User] {

	var users []domain.User

	query := db

	if queryDto.UserName != "" {
		query = query.Where("user_name like ?", "%"+queryDto.UserName+"%")
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
