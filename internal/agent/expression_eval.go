package agent

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"txing-ai/internal/global/logging/log"
	"go.uber.org/zap"
)

// ExpressionOperator 表达式运算符
type ExpressionOperator string

const (
	OpContains    ExpressionOperator = "contains"
	OpEquals      ExpressionOperator = "equals"
	OpStartsWith  ExpressionOperator = "starts_with"
	OpEndsWith    ExpressionOperator = "ends_with"
	OpMatches     ExpressionOperator = "matches"
	OpGreaterThan ExpressionOperator = "greater_than"
	OpLessThan    ExpressionOperator = "less_than"
	OpNotEquals   ExpressionOperator = "not_equals"
	OpContainsAny ExpressionOperator = "contains_any"
)

// ExpressionEvaluator 表达式求值器
type ExpressionEvaluator struct{}

// NewExpressionEvaluator 创建表达式求值器
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{}
}

// Evaluate 执行表达式判断
// 表达式格式: {{output}} operator value
// 例如: {{output}} contains "成功"
func (e *ExpressionEvaluator) Evaluate(expression string, input string) *ConditionResult {
	// 空表达式默认返回 true
	if expression == "" {
		return NewConditionResult(true, "表达式为空，默认通过")
	}

	// 解析表达式
	op, value, err := e.parseExpression(expression)
	if err != nil {
		return NewConditionError(err, DefaultConditionConfig())
	}

	// 执行判断
	result := e.executeOperator(op, input, value)
	return result
}

// parseExpression 解析表达式，提取运算符和值
func (e *ExpressionEvaluator) parseExpression(expression string) (ExpressionOperator, string, error) {
	// 去除空白
	expression = strings.TrimSpace(expression)

	// 支持 {{output}} 或 ${output} 或直接 output 变量名
	// 格式: {{output}} operator value

	// 定义运算符列表（按优先级排序，长的先匹配）
	operators := []ExpressionOperator{
		OpStartsWith,
		OpEndsWith,
		OpGreaterThan,
		OpLessThan,
		OpContainsAny,
		OpContains,
		OpMatches,
		OpEquals,
		OpNotEquals,
	}

	for _, op := range operators {
		// 从右向左查找运算符位置，避免匹配到输出内容中的子串
		// 运算符前后应该有空格，如 `{{output}} not_equals ''`
		opStr := " " + string(op) + " "
		idx := strings.LastIndex(strings.ToLower(expression), opStr)
		if idx == -1 {
			// 也尝试匹配值为引号开头的情况，如 `not_equals ''`
			opQuote := " " + string(op) + " '"
			idx = strings.LastIndex(strings.ToLower(expression), opQuote)
		}
		if idx == -1 {
			opQuote := " " + string(op) + " \""
			idx = strings.LastIndex(strings.ToLower(expression), opQuote)
		}
		if idx != -1 {
			// 提取运算符后的值（跳过前导空格 + 运算符 + 空格）
			valuePart := strings.TrimSpace(expression[idx+len(op)+2:])
			// 去除引号
			value := e.extractValue(valuePart)
			return op, value, nil
		}
	}

	return "", "", &ExpressionError{Msg: "无法解析表达式: " + expression}
}

// extractValue 从值部分提取实际值
func (e *ExpressionEvaluator) extractValue(valuePart string) string {
	valuePart = strings.TrimSpace(valuePart)

	// 去除首尾引号（支持单引号和双引号）
	if len(valuePart) >= 2 {
		if (valuePart[0] == '"' && valuePart[len(valuePart)-1] == '"') ||
			(valuePart[0] == '\'' && valuePart[len(valuePart)-1] == '\'') {
			return valuePart[1 : len(valuePart)-1]
		}
	}

	return valuePart
}

// executeOperator 执行运算符判断
func (e *ExpressionEvaluator) executeOperator(op ExpressionOperator, input string, value string) *ConditionResult {
	var result bool
	var reason string

	switch op {
	case OpContains:
		result = strings.Contains(input, value)
		reason = e.formatReason("包含", input, value, result)

	case OpEquals:
		result = input == value
		reason = e.formatReason("等于", input, value, result)

	case OpNotEquals:
		result = input != value
		reason = e.formatReason("不等于", input, value, result)

	case OpStartsWith:
		result = strings.HasPrefix(input, value)
		reason = e.formatReason("开始于", input, value, result)

	case OpEndsWith:
		result = strings.HasSuffix(input, value)
		reason = e.formatReason("结束于", input, value, result)

	case OpMatches:
		// 正则匹配
		re, err := regexp.Compile(value)
		if err != nil {
			return NewConditionError(&ExpressionError{Msg: "正则表达式无效: " + err.Error()}, DefaultConditionConfig())
		}
		result = re.MatchString(input)
		reason = e.formatReason("匹配正则", input, value, result)

	case OpGreaterThan:
		// 数值比较
		inputNum, err1 := e.parseNumber(input)
		valueNum, err2 := e.parseNumber(value)
		if err1 != nil || err2 != nil {
			return NewConditionError(&ExpressionError{Msg: "数值比较需要有效的数字"}, DefaultConditionConfig())
		}
		result = inputNum > valueNum
		reason = e.formatReasonNumeric("大于", inputNum, valueNum, result)

	case OpLessThan:
		// 数值比较
		inputNum, err1 := e.parseNumber(input)
		valueNum, err2 := e.parseNumber(value)
		if err1 != nil || err2 != nil {
			return NewConditionError(&ExpressionError{Msg: "数值比较需要有效的数字"}, DefaultConditionConfig())
		}
		result = inputNum < valueNum
		reason = e.formatReasonNumeric("小于", inputNum, valueNum, result)

	case OpContainsAny:
		// 包含任意一个（value 是 JSON 数组或逗号分隔）
		values := e.parseArrayOrCSV(value)
		result = e.containsAny(input, values)
		reason = e.formatReasonArray("包含任意", input, values, result)

	default:
		return NewConditionError(&ExpressionError{Msg: "未知运算符: " + string(op)}, DefaultConditionConfig())
	}

	return NewConditionResult(result, reason)
}

// parseNumber 解析数字
func (e *ExpressionEvaluator) parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// parseArrayOrCSV 解析数组或逗号分隔值
func (e *ExpressionEvaluator) parseArrayOrCSV(value string) []string {
	// 尝试解析为 JSON 数组
	var arr []string
	if err := json.Unmarshal([]byte(value), &arr); err == nil {
		return arr
	}

	// 作为逗号分隔处理
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// containsAny 检查输入是否包含数组中的任意一个值
func (e *ExpressionEvaluator) containsAny(input string, values []string) bool {
	for _, v := range values {
		if strings.Contains(input, v) {
			return true
		}
	}
	return false
}

// formatReason 格式化原因
func (e *ExpressionEvaluator) formatReason(op string, input string, value string, result bool) string {
	truncateLen := 50
	inputDisplay := input
	if len(input) > truncateLen {
		inputDisplay = input[:truncateLen] + "..."
	}

	if result {
		return inputDisplay + " " + op + " '" + value + "'"
	}
	return inputDisplay + " 不" + op + " '" + value + "'"
}

// formatReasonNumeric 格式化数值比较原因
func (e *ExpressionEvaluator) formatReasonNumeric(op string, input, value float64, result bool) string {
	if result {
		return strconv.FormatFloat(input, 'f', -1, 64) + " " + op + " " + strconv.FormatFloat(value, 'f', -1, 64)
	}
	return strconv.FormatFloat(input, 'f', -1, 64) + " 不" + op + " " + strconv.FormatFloat(value, 'f', -1, 64)
}

// formatReasonArray 格式化数组判断原因
func (e *ExpressionEvaluator) formatReasonArray(op string, input string, values []string, result bool) string {
	truncateLen := 50
	inputDisplay := input
	if len(input) > truncateLen {
		inputDisplay = input[:truncateLen] + "..."
	}

	valuesStr := strings.Join(values, ", ")
	if len(valuesStr) > 30 {
		valuesStr = valuesStr[:30] + "..."
	}

	if result {
		return inputDisplay + " " + op + " [" + valuesStr + "]"
	}
	return inputDisplay + " 不" + op + " [" + valuesStr + "]"
}

// ExpressionError 表达式错误
type ExpressionError struct {
	Msg string
}

func (e *ExpressionError) Error() string {
	return "表达式错误: " + e.Msg
}

// EvaluateBatch 批量执行多个表达式（AND 连接）
func (e *ExpressionEvaluator) EvaluateBatch(expressions []string, input string) *ConditionResult {
	if len(expressions) == 0 {
		return NewConditionResult(true, "无表达式，默认通过")
	}

	for _, expr := range expressions {
		result := e.Evaluate(expr, input)
		if !result.Result {
			return result // 短路返回第一个失败的结果
		}
	}

	return NewConditionResult(true, "所有表达式都满足")
}

// EvaluateWithVars 使用变量上下文执行表达式
// vars 是变量名到值的映射，例如: {"output": "操作成功", "code": "200"}
func (e *ExpressionEvaluator) EvaluateWithVars(expression string, vars map[string]string) *ConditionResult {
	// 替换变量
	input := e.replaceVars(expression, vars)
	log.Debug("表达式变量替换", zap.String("original", expression), zap.String("replaced", input))

	// 提取实际的表达式部分（去掉变量占位符后）
	// 重新解析以获取运算符和值
	return e.Evaluate(input, vars["output"])
}

// replaceVars 替换表达式中的变量
func (e *ExpressionEvaluator) replaceVars(expression string, vars map[string]string) string {
	result := expression

	// 替换 {{var}} 格式
	for name, value := range vars {
		result = strings.ReplaceAll(result, "{{"+name+"}}", value)
		result = strings.ReplaceAll(result, "${"+name+"}", value)
	}

	return result
}
