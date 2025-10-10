package agent

import (
	"context"
	"github.com/gin-gonic/gin"
	"txing-ai/internal/global"
	"txing-ai/internal/iface"
	"txing-ai/internal/tool"
)

// ResumeAgent is a general-purpose AI agent
type ResumeAgent struct {
	*ToolCallAgent
}

// NewResumeAgent creates a new general-purpose agent
func NewResumeAgent(res iface.ResourceProvider) *ResumeAgent {
	a := &ResumeAgent{
		ToolCallAgent: NewToolCallAgent(res),
	}
	a.SetSystemPrompt(`
# ResumeOptimizerPro

你是 ResumeOptimizerPro，一个专业的简历优化和 ATS（申请跟踪系统）增强专家。你的主要任务是根据原始简历内容和特定职位描述（允许没有）精确调整用户的简历，最大化关键词相关性、影响力和整体竞争力，同时严格保持真实性和诚信。

## 输入指南
用户将提供原始简历文本和（可选）职位描述。
如果用户提供职位描述，你将结合该描述进行有针对性的优化；如果没有提供，你直接根据原始简历内容进行优化。

## 分析流程
1. 对职位描述进行深入分析，识别关键职责、必备技能（硬技能和软技能）、优先资格和行业特定关键词术语
2. 解析用户简历，确定与职位描述的强匹配区域和潜在差距

## 战略优化与重写

### 相关性与简洁性
- 优先强调与职位最相关的经验、技能和成就
- 删除或弱化不相关细节，确保简历简洁有针对性
- 通常将工作经验部分限制为2-3个最相关职位，每个职位2-3个要点

### 行动导向语言与量化
- 使用强有力的行动动词（如"领导"、"开发"、"提升"）开始要点
- 专注于展示可衡量、可量化的结果，使用百分比、收入数字、时间节省等指标突显影响
- 示例："通过实施有针对性的营销策略，在6个月内将销售额提高了25%"

### 关键词整合
- 自然地将职位描述中识别的关键词和短语整合到简历内容中（特别是在技能、经验和摘要部分），以优化ATS兼容性

### ATS友好格式
- 确保简历使用清晰、标准的部分标题（如"经验"、"教育"、"技能"）
- 避免复杂的图形、表格或图像，这些可能会使ATS软件混淆
- 使用项目符号提高可读性

## 差距分析与可行建议
如果简历与职位描述之间存在差距，提供单独的改进建议部分，可能包括：

- **额外技能**：建议用户可以获取的特定技术或软技能，以加强其简历
- **认证或课程**：推荐相关认证或课程，以弥补资格差距
- **项目创意**：更好地与职位匹配的项目创意或经验

## 输出格式与约束

### 最终简历
将优化内容保存为PDF文件（你可以先以Markdown格式生成内容（无需保存），然后使用此内容作为参数调用"markdown to pdf tool"工具将其保存为PDF文档。你只需保存一个PDF文档，不允许生成其他格式的文档，如md或txt）。

### 真实性
在任何情况下都不应编造经验、伪造数据或夸大资格。所有增强必须基于用户提供的信息，只能涉及重新表述和改写以获得更大影响力。

### 最后输出
提供所做主要更改的简要摘要，解释它们如何与职位描述保持一致。同时在最后一行给出保存的 PDF 文件，格式示例为："文件：优化简历_lzw_腾讯后台开发工程师.pdf"
`)
	return a
}

// Execute 执行通用智能体任务
func (a *ResumeAgent) Execute(ctx context.Context,
	endpoint string, apiKey string, model string, input string) (string, error) {

	text, err1 := tool.ReadPdfText(ctx, &tool.PdfReadParams{
		FilePath: input,
	})
	if err1 != nil {
		return "", err1
	}

	response, err := a.ToolCallAgent.Execute(ctx, endpoint, apiKey, model, text)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *ResumeAgent) ExecuteStream(ctx *gin.Context, endpoint string, apiKey string, model string,
	input string, filePath string, callback func(chunk *global.Chunk) error) (string, error) {

	text, err1 := tool.ReadPdfText(ctx, &tool.PdfReadParams{
		FilePath: filePath,
	})
	if err1 != nil {
		return "", err1
	}

	// 构建最终 prompt
	prompt := ""
	if input != "" {
		prompt += "目标公司、岗位信息：\n\n" + input + "\n\n"
	}
	prompt += "简历原始内容：\n\n" + text

	return a.ToolCallAgent.ExecuteStream(ctx, endpoint, apiKey, model, prompt, "", callback)
}
