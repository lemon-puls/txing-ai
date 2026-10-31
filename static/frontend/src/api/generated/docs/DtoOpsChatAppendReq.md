# TxingAiApi.DtoOpsChatAppendReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**content** | **String** |  | [optional] 
**markProposalConfirmed** | **Boolean** | 是否将未确认的提案标记为已确认 | [optional] 
**proposalName** | **String** | 被确认提案的 name（批量录入时精确匹配；为空时回退为\&quot;最后一条未确认提案\&quot;） | [optional] 
**rewindLastRound** | **Boolean** | 回退最后一轮问答（重新生成/失败重试前调用）：移除末尾的助手消息与其对应用户消息； 为 true 时 Content 可为空（不追加任何消息） | [optional] 
**role** | **String** |  | 



## Enum: RoleEnum


* `assistant` (value: `"assistant"`)




