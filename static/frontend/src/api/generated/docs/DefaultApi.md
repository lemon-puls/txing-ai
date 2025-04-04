# TxingAiApi.DefaultApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiUserListGet**](DefaultApi.md#apiUserListGet) | **GET** /api/user/list | 获取用户列表



## apiUserListGet

> UtilsResponse apiUserListGet(page, limit, opts)

获取用户列表

获取用户列表，支持分页和条件筛选

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let page = 56; // Number | 页码
let limit = 56; // Number | 每页数量
let opts = {
  'orderBy': "orderBy_example", // String | 排序字段
  'order': "order_example", // String | 排序方式(asc/desc)
  'userId': "userId_example", // String | 用户ID
  'status': 56 // Number | 状态(0:禁用 1:启用)
};
apiInstance.apiUserListGet(page, limit, opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **Number**| 页码 | 
 **limit** | **Number**| 每页数量 | 
 **orderBy** | **String**| 排序字段 | [optional] 
 **order** | **String**| 排序方式(asc/desc) | [optional] 
 **userId** | **String**| 用户ID | [optional] 
 **status** | **Number**| 状态(0:禁用 1:启用) | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

