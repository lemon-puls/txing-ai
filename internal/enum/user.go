package enum

// 用户相关枚举值
const (
	// 用户类型
	UserTypeNormal int8 = 0 // 普通用户
	UserTypeSuper  int8 = 1 // 超级管理员

	// 用户状态
	UserStatusNormal    int8 = 0 // 正常
	UserStatusForbidden int8 = 1 // 禁用

	// 登录失败次数限制
	//LoginFailLimit = 3 // 登录失败超过此次数需要验证码
)

// 用户角色描述
var UserTypeDesc = map[int8]string{
	UserTypeNormal: "普通用户",
	UserTypeSuper:  "超级管理员",
}

// 用户状态描述
var UserStatusDesc = map[int8]string{
	UserStatusNormal:    "正常",
	UserStatusForbidden: "禁用",
}
