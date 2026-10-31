# TxingAiApi.DomainOpsChatMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**confirmed** | **Boolean** | Confirmed 标记提案确认完成的提示消息（前端本地产生，经 append 接口持久化） | [optional] 
**content** | **String** |  | [optional] 
**durationMs** | **Number** | DurationMs 本次回复总耗时（毫秒），随消息持久化，回放时可展示 | [optional] 
**error** | **String** |  | [optional] 
**interrupted** | **Boolean** | Interrupted 表示本次回复被用户主动中断（内容为已生成的部分） | [optional] 
**proposal** | **Object** | Proposal/ProposalStatus/ProposalMessage 为旧版单提案字段，仅用于反序列化历史数据； QueryOpsChatSessionById 加载时经 NormalizeProposals 统一并入 Proposals，保存后即落为新结构 | [optional] 
**proposalMessage** | **String** |  | [optional] 
**proposalStatus** | **String** |  | [optional] 
**proposals** | [**[DomainOpsChatProposalItem]**](DomainOpsChatProposalItem.md) | Proposals 本条消息携带的全部提案卡片 | [optional] 
**reasoning** | **String** |  | [optional] 
**role** | **String** | user/assistant | [optional] 
**toolCalls** | [**[DomainOpsChatToolCall]**](DomainOpsChatToolCall.md) |  | [optional] 


