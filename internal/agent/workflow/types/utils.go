package types

import (
	"encoding/json"
	"regexp"

	"go.uber.org/zap"
	"txing-ai/internal/global/logging/log"
)

// ReplaceNestedVars 替换 {{input.xxx}} 和 {{output.xxx}} 格式的嵌套变量
// ReplaceNestedVars replaces nested variables like {{input.xxx}} and {{output.xxx}}
// 支持 JSON 字段访问：如果 input/output 是 JSON 字符串，会解析并提取对应字段
// Supports JSON field access: parses input/output as JSON if it's a JSON string
func ReplaceNestedVars(s, input, output string) string {
	re := regexp.MustCompile(`\{\{(input|output)(?:\.(\w+))?\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// 提取变量名和字段名 / Extract variable name and field name
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		varName := sub[1]
		fieldName := sub[2]

		// 选择源字符串 / Select source string
		var source string
		if varName == "input" {
			source = input
		} else {
			source = output
		}

		// 如果没有字段名，返回整个字符串 / If no field name, return whole string
		if fieldName == "" {
			return source
		}

		// 尝试解析为 JSON 并提取字段 / Try to parse as JSON and extract field
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(source), &data); err == nil {
			if val, ok := data[fieldName]; ok {
				return toString(val)
			}
		}

		// 回退策略：当 source 不是 JSON 时（如纯文本输入），用整个 source 替换
		// 这样 {{input.question}} 等同于 {{input}}
		// Fallback: if source is not JSON (e.g. plain text user input),
		// treat {{input.xxx}} as {{input}} to match user expectations
		log.Debug("嵌套变量按整体输入回退 / Nested var falls back to whole input",
			zap.String("variable", varName),
			zap.String("field", fieldName),
			zap.String("source", source))
		return source
	})
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
