# TxingAiApi.AgentApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiAgentExecPost**](AgentApi.md#apiAgentExecPost) | **POST** /api/agent/exec | 调用智能体
[**apiAgentExecStreamPost**](AgentApi.md#apiAgentExecStreamPost) | **POST** /api/agent/exec/stream | 基于 SSE 调用智能体



## apiAgentExecPost

> UtilsResponse apiAgentExecPost(data)

调用智能体

调用智能体

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.AgentApi();
let data = new TxingAiApi.DtoAgentExecReq(); // DtoAgentExecReq | 请求信息
apiInstance.apiAgentExecPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoAgentExecReq**](DtoAgentExecReq.md)| 请求信息 | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiAgentExecStreamPost

> UtilsResponse apiAgentExecStreamPost(data)

基于 SSE 调用智能体

使用 Server-Sent Events 流式调用智能体

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.AgentApi();
let data = new TxingAiApi.DtoAgentExecReq(); // DtoAgentExecReq | 请求信息
apiInstance.apiAgentExecStreamPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoAgentExecReq**](DtoAgentExecReq.md)| 请求信息 | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: text/event-stream

