# TxingAiApi.DomainOpsChatMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**confirmed** | **Boolean** | Confirmed 标记提案确认完成的提示消息（前端本地产生，经 append 接口持久化） | [optional] 
**content** | **String** |  | [optional] 
**error** | **String** |  | [optional] 
**interrupted** | **Boolean** | Interrupted 表示本次回复被用户主动中断（内容为已生成的部分） | [optional] 
**proposal** | [**DomainOpsChatProposal**](DomainOpsChatProposal.md) |  | [optional] 
**proposalMessage** | **String** |  | [optional] 
**proposalStatus** | **String** |  | [optional] 
**reasoning** | **String** |  | [optional] 
**role** | **String** | user/assistant | [optional] 
**toolCalls** | [**[DomainOpsChatToolCall]**](DomainOpsChatToolCall.md) |  | [optional] 


