package utils

import (
	"context"
)

// GetFromContext 从Context中获取指定类型的值
func GetFromContext[T any](ctx context.Context, key any) T {
	return ctx.Value(key).(T)
}

// GetFromContextWithDefault 从Context中获取指定类型的值，如果不存在则返回默认值
func GetFromContextWithDefault[T any](ctx context.Context, key any, defaultValue T) T {
	value := ctx.Value(key)
	if value == nil {
		return defaultValue
	}
	return value.(T)
}

// GetFromContextWithOK 从Context中获取指定类型的值，并返回是否存在
func GetFromContextWithOK[T any](ctx context.Context, key any) (T, bool) {
	var defaultValue T
	value := ctx.Value(key)
	if value == nil {
		return defaultValue, false
	}
	return value.(T), true
}

// 以下是基于泛型函数的具体应用

// GetDBFromContext 获取数据库连接
func GetDBFromContext[T any](ctx context.Context) T {
	return GetFromContext[T](ctx, "db")
}

// GetUIDFromContext 获取用户ID
func GetUIDFromContext(ctx context.Context) int64 {
	return GetFromContext[int64](ctx, "userId")
}

// GetUIDFromContextAllowEmpty 从 context 中获取 userId，允许为空
func GetUIDFromContextAllowEmpty(ctx context.Context) (int64, bool) {
	return GetFromContextWithOK[int64](ctx, "userId")
}

// GetCosClientFromContext 获取COS客户端
func GetCosClientFromContext[T any](ctx context.Context) T {
	return GetFromContext[T](ctx, "cos")
}

// GetAgentFactoryFromContext 获取Agent工厂
func GetAgentFactoryFromContext[T any](ctx context.Context) T {
	return GetFromContext[T](ctx, "agentFactory")
}

// GetMessageLimiterFromContext 获取消息限制器
func GetMessageLimiterFromContext[T any](ctx context.Context) T {
	return GetFromContext[T](ctx, "messageLimiter")
}

// GetRoleFromContext 获取角色
func GetRoleFromContext(ctx context.Context) int8 {
	return GetFromContext[int8](ctx, "role")
}

// GetIsAdminFromContext 判断是否为管理员
func GetIsAdminFromContext(ctx context.Context) bool {
	// 这里不再依赖enum包，而是使用常量值
	const UserTypeSuper int8 = 1
	return GetFromContext[int8](ctx, "role") == UserTypeSuper
}

// GetRDBFromContext 获取Redis客户端
func GetRDBFromContext[T any](ctx context.Context) T {
	return GetFromContext[T](ctx, "redis")
}

// GetAgentFromContext 获取Agent名称
func GetAgentFromContext(ctx context.Context) string {
	return GetFromContext[string](ctx, "agent")
}
