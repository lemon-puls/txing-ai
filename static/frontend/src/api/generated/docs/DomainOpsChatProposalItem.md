# TxingAiApi.DomainOpsChatProposalItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**message** | **String** |  | [optional] 
**proposal** | **Object** | Proposal 结构化提案原始 JSON（website/model/channel/preset 各自的扁平结构， 前端按 type 字段分发渲染卡片。 internal/tool/ops 反向依赖了 domain，这里不能 import tool 包，由 controller 层透传原始 JSON） swaggertype:object 让 swag 将 json.RawMessage 识别为任意 JSON 对象 | [optional] 
**status** | **String** |  | [optional] 


