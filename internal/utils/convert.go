package utils

import (
	"fmt"
	"strconv"
)

// MustParseInt64 将字符串转换为 int64，如果转换失败则 panic
// Convert string to int64, panic if conversion fails
func MustParseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse int64 from string: %s, error: %v", s, err))
	}
	return i
}

// MustParseInt 将字符串转换为 int，如果转换失败则 panic
// Convert string to int, panic if conversion fails
func MustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse int from string: %s, error: %v", s, err))
	}
	return i
}

// MustParseFloat64 将字符串转换为 float64，如果转换失败则 panic
// Convert string to float64, panic if conversion fails
func MustParseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse float64 from string: %s, error: %v", s, err))
	}
	return f
}

// MustParseBool 将字符串转换为 bool，如果转换失败则 panic
// Convert string to bool, panic if conversion fails
func MustParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse bool from string: %s, error: %v", s, err))
	}
	return b
}

// SafeParseInt64 将字符串转换为 int64，如果转换失败则返回默认值
// Convert string to int64, return default value if conversion fails
func SafeParseInt64(s string, defaultValue int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

// SafeParseInt 将字符串转换为 int，如果转换失败则返回默认值
// Convert string to int, return default value if conversion fails
func SafeParseInt(s string, defaultValue int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return i
}

// SafeParseFloat64 将字符串转换为 float64，如果转换失败则返回默认值
// Convert string to float64, return default value if conversion fails
func SafeParseFloat64(s string, defaultValue float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

// SafeParseBool 将字符串转换为 bool，如果转换失败则返回默认值
// Convert string to bool, return default value if conversion fails
func SafeParseBool(s string, defaultValue bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return defaultValue
	}
	return b
}

// ToString 将任意类型转换为字符串
// Convert any type to string
func ToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}
