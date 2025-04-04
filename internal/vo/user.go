package vo

import (
	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"txing-ai/internal/domain"
)

type UserVO struct {
	domain.BaseModel
	UserId   int64  `json:"userId"`
	Username string `json:"userName"`
	Password string `json:"-"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   int8   `json:"gender"`
	Status   int8   `json:"status"`
	Age      int8   `json:"age"`
	Avatar   string `json:"avatar"`
}

func ToUserVO(user domain.User) UserVO {
	var userVo UserVO
	copier.Copy(&userVo, &user)
	return userVo
}

func ToUserVOs(users []domain.User) []UserVO {
	userVos := lo.Map(users, func(user domain.User, _ int) UserVO {
		return ToUserVO(user)
	})

	return userVos
}
