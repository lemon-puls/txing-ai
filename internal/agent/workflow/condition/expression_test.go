package condition

import (
	"testing"
)

func TestExpressionEvaluator_Contains(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"包含成功", "{{output}} contains '成功'", "操作成功", true},
		{"不包含", "{{output}} contains '失败'", "操作成功", false},
		{"包含数字", "{{output}} contains '123'", "结果是123", true},
		{"空输入", "{{output}} contains 'test'", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_Equals(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"精确匹配", "{{output}} equals 'hello'", "hello", true},
		{"不匹配", "{{output}} equals 'world'", "hello", false},
		{"大小写敏感", "{{output}} equals 'Hello'", "hello", false},
		{"空字符串", "{{output}} equals ''", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_StartsWith(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"前缀匹配", "{{output}} starts_with 'prefix'", "prefix_value", true},
		{"不匹配", "{{output}} starts_with 'suffix'", "prefix_value", false},
		{"空前缀", "{{output}} starts_with ''", "any", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_EndsWith(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"后缀匹配", "{{output}} ends_with 'value'", "prefix_value", true},
		{"不匹配", "{{output}} ends_with 'prefix'", "prefix_value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_Matches(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"正则匹配数字", "{{output}} matches '^\\d+$'", "12345", true},
		{"正则不匹配", "{{output}} matches '^\\d+$'", "123abc", false},
		{"正则匹配邮箱", "{{output}} matches '^[\\w.]+@[\\w.]+$'", "test@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_GreaterThan(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"大于", "{{output}} greater_than 10", "15", true},
		{"等于", "{{output}} greater_than 10", "10", false},
		{"小于", "{{output}} greater_than 10", "5", false},
		{"浮点数", "{{output}} greater_than 10.5", "10.6", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_LessThan(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"小于", "{{output}} less_than 10", "5", true},
		{"等于", "{{output}} less_than 10", "10", false},
		{"大于", "{{output}} less_than 10", "15", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_ContainsAny(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"包含其中一个", "{{output}} contains_any '成功,完成,通过'", "操作成功了", true},
		{"包含多个", "{{output}} contains_any '成功,完成'", "操作成功完成", true},
		{"不包含", "{{output}} contains_any '失败,错误'", "操作成功", false},
		{"JSON数组", "{{output}} contains_any '[\"成功\",\"完成\"]'", "操作成功", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_NotEquals(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{"不相等", "{{output}} not_equals 'hello'", "world", true},
		{"相等", "{{output}} not_equals 'hello'", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_OutputContainsOperator(t *testing.T) {
	eval := NewExpressionEvaluator()

	// 当输出内容包含运算符子串时，不应错误匹配
	// 这些测试直接用替换后的表达式（模拟 EvaluateWithVars 的 replaceVars 结果）
	tests := []struct {
		name     string
		expr     string
		input    string
		expected bool
	}{
		{
			"输出包含equals子串-not_equals空串",
			`{"status":"equals","data":"test"} not_equals ''`,
			`{"status":"equals","data":"test"}`,
			true,
		},
		{
			"输出包含contains子串-not_equals空串",
			`{"msg":"contains some data"} not_equals ''`,
			`{"msg":"contains some data"}`,
			true,
		},
		{
			"空输出-not_equals空串",
			` not_equals ''`,
			``,
			false,
		},
		{
			"输出包含not_equals子串-contains判断",
			`{"result":"not_equals","status":"success"} contains 'success'`,
			`{"result":"not_equals","status":"success"}`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.expr, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_EmptyExpression(t *testing.T) {
	eval := NewExpressionEvaluator()

	result := eval.Evaluate("", "test input")
	if !result.Result {
		t.Error("空表达式应返回 true")
	}
}

func TestExpressionEvaluator_InvalidExpression(t *testing.T) {
	eval := NewExpressionEvaluator()

	result := eval.Evaluate("invalid expression", "test input")
	if result.Error == nil {
		t.Error("无效表达式应返回错误")
	}
}

func TestExpressionEvaluator_Batch(t *testing.T) {
	eval := NewExpressionEvaluator()

	tests := []struct {
		name        string
		expressions []string
		input       string
		expected    bool
	}{
		{
			"全部满足",
			[]string{
				"{{output}} contains '成功'",
				"{{output}} starts_with '操作'",
			},
			"操作成功",
			true,
		},
		{
			"部分满足",
			[]string{
				"{{output}} contains '成功'",
				"{{output}} contains '失败'",
			},
			"操作成功",
			false,
		},
		{
			"空列表",
			[]string{},
			"test",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.EvaluateBatch(tt.expressions, tt.input)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestExpressionEvaluator_WithVars(t *testing.T) {
	eval := NewExpressionEvaluator()

	vars := map[string]string{
		"output": "操作成功",
		"code":   "200",
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{"变量替换", "{{output}} contains '成功'", true},
		{"多变量", "{{code}} equals '200'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.EvaluateWithVars(tt.expr, vars)
			if result.Result != tt.expected {
				t.Errorf("期望 %v, 得到 %v, 原因: %s", tt.expected, result.Result, result.Reason)
			}
		})
	}
}

func TestConditionResult(t *testing.T) {
	tests := []struct {
		name           string
		result         bool
		expectedBranch string
	}{
		{"true 结果", true, "true"},
		{"false 结果", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewConditionResult(tt.result, "测试原因")
			if result.Result != tt.result {
				t.Errorf("期望 Result=%v, 得到 %v", tt.result, result.Result)
			}
			if result.Branch != tt.expectedBranch {
				t.Errorf("期望 Branch=%s, 得到 %s", tt.expectedBranch, result.Branch)
			}
		})
	}
}

func TestConditionConfig_GetHandles(t *testing.T) {
	config := DefaultConditionConfig()

	if config.GetTrueHandle() != "true" {
		t.Errorf("期望 true handle='true', 得到 %s", config.GetTrueHandle())
	}

	if config.GetFalseHandle() != "false" {
		t.Errorf("期望 false handle='false', 得到 %s", config.GetFalseHandle())
	}

	config.TrueHandle = "yes"
	config.FalseHandle = "no"

	if config.GetTrueHandle() != "yes" {
		t.Errorf("期望 true handle='yes', 得到 %s", config.GetTrueHandle())
	}

	if config.GetFalseHandle() != "no" {
		t.Errorf("期望 false handle='no', 得到 %s", config.GetFalseHandle())
	}
}

func TestNewConditionError(t *testing.T) {
	config := DefaultConditionConfig()
	config.FailureAction = FailureActionDefaultFalse

	err := &ExpressionError{Msg: "test error"}
	result := NewConditionError(err, config)

	if result.Error == nil {
		t.Error("应包含错误信息")
	}

	if result.Branch != "false" {
		t.Errorf("默认错误处理应走 false 分支, 得到 %s", result.Branch)
	}
}

func TestNewConditionError_Terminate(t *testing.T) {
	config := DefaultConditionConfig()
	config.FailureAction = FailureActionTerminate

	err := &ExpressionError{Msg: "test error"}
	result := NewConditionError(err, config)

	if result.Branch != "" {
		t.Errorf("终止策略应返回空分支, 得到 %s", result.Branch)
	}
}

func TestNewConditionError_Configurable(t *testing.T) {
	config := DefaultConditionConfig()
	config.FailureAction = FailureActionConfigurable
	config.FailureBranch = "true"

	err := &ExpressionError{Msg: "test error"}
	result := NewConditionError(err, config)

	if result.Branch != "true" {
		t.Errorf("可配置策略应返回配置的分支, 得到 %s", result.Branch)
	}
}
