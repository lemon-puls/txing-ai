# TxingAiApi.COSApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiCosPresignedUrlPost**](COSApi.md#apiCosPresignedUrlPost) | **POST** /api/cos/presigned-url | 获取预签名URL



## apiCosPresignedUrlPost

> ApiCosPresignedUrlPost200Response apiCosPresignedUrlPost(data)

获取预签名URL

获取文件上传或下载的预签名URL

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.COSApi();
let data = new TxingAiApi.DtoGetPresignedURLReq(); // DtoGetPresignedURLReq | 请求参数
apiInstance.apiCosPresignedUrlPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoGetPresignedURLReq**](DtoGetPresignedURLReq.md)| 请求参数 | 

### Return type

[**ApiCosPresignedUrlPost200Response**](ApiCosPresignedUrlPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

