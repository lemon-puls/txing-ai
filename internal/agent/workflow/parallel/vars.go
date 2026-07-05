package parallel

import (
	"encoding/json"
	"strings"

	"txing-ai/internal/agent/workflow/types"
)


// escapeJSON 转义 JSON 特殊字符
// escapeJSON escapes JSON special characters
func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}


// ReplaceVarsInParams 递归替换参数 map 中的变量占位符
// 支持 {{input}}、{{output}}、{{input.field}} 等格式
// 并将处理后的值统一转为 string 以便传递给 InvokableRun
func ReplaceVarsInParams(params map[string]interface{}, input, output string) map[string]interface{} {
	if params == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = replaceVarValue(v, input, output)
	}
	return result
}


// replaceVarValue 替换单个值中的变量占位符，支持递归处理 map/slice
func replaceVarValue(v interface{}, input, output string) interface{} {
	switch val := v.(type) {
	case string:
		replaced := val
		replaced = strings.ReplaceAll(replaced, "{{input}}", input)
		replaced = strings.ReplaceAll(replaced, "{{output}}", output)
		replaced = types.ReplaceNestedVars(replaced, input, output)
		return replaced
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k2, v2 := range val {
			m[k2] = replaceVarValue(v2, input, output)
		}
		return m
	case []interface{}:
		arr := make([]interface{}, len(val))
		for i, elem := range val {
			arr[i] = replaceVarValue(elem, input, output)
		}
		return arr
	default:
		return val
	}
}


// toString 将任意类型转为字符串
// toString converts any value to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case nil:
		return ""
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

