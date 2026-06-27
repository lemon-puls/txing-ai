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

> UtilsResponse apiAgentExecStreamPost(agentType, opts)

基于 SSE 调用智能体

使用 Server-Sent Events 流式调用智能体

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.AgentApi();
let agentType = "agentType_example"; // String | 智能体类型
let opts = {
  'content': "content_example", // String | 请求内容
  'file': "/path/to/file" // File | 上传文件
};
apiInstance.apiAgentExecStreamPost(agentType, opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **agentType** | **String**| 智能体类型 | 
 **content** | **String**| 请求内容 | [optional] 
 **file** | **File**| 上传文件 | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: text/event-stream

