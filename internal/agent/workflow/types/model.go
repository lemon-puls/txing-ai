package types

// ModelInfo 模型信息（包含端点和密钥）
type ModelInfo struct {
	Endpoint string
	APIKey   string
	Model    string // 映射后的模型名称
}

// ModelResolver 模型解析器接口，用于根据模型名称获取对应的端点和密钥
type ModelResolver interface {
	Resolve(modelName string) (*ModelInfo, error)
}
