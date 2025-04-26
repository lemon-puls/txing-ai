# TxingAiApi.DefaultApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiAdminChannelIdDelete**](DefaultApi.md#apiAdminChannelIdDelete) | **DELETE** /api/admin/channel/{id} | 删除渠道
[**apiAdminChannelIdGet**](DefaultApi.md#apiAdminChannelIdGet) | **GET** /api/admin/channel/{id} | 获取渠道详情
[**apiAdminChannelIdPut**](DefaultApi.md#apiAdminChannelIdPut) | **PUT** /api/admin/channel/{id} | 更新渠道
[**apiAdminChannelListGet**](DefaultApi.md#apiAdminChannelListGet) | **GET** /api/admin/channel/list | 获取渠道列表
[**apiAdminChannelPost**](DefaultApi.md#apiAdminChannelPost) | **POST** /api/admin/channel | 创建渠道
[**apiAdminModelIdDelete**](DefaultApi.md#apiAdminModelIdDelete) | **DELETE** /api/admin/model/{id} | 删除模型
[**apiAdminModelIdPut**](DefaultApi.md#apiAdminModelIdPut) | **PUT** /api/admin/model/{id} | 更新模型
[**apiAdminModelPost**](DefaultApi.md#apiAdminModelPost) | **POST** /api/admin/model | 创建模型
[**apiAdminUserListGet**](DefaultApi.md#apiAdminUserListGet) | **GET** /api/admin/user/list | 获取用户列表
[**apiAdminUserStatusIdPut**](DefaultApi.md#apiAdminUserStatusIdPut) | **PUT** /api/admin/user/status/{id} | 切换用户状态
[**apiCaptchaGet**](DefaultApi.md#apiCaptchaGet) | **GET** /api/captcha | 生成验证码
[**apiCosPresignedUrlPost**](DefaultApi.md#apiCosPresignedUrlPost) | **POST** /api/cos/presigned-url | 获取预签名URL
[**apiModelIdGet**](DefaultApi.md#apiModelIdGet) | **GET** /api/model/{id} | 获取模型详情
[**apiModelListGet**](DefaultApi.md#apiModelListGet) | **GET** /api/model/list | 获取模型列表
[**apiPresetIdDelete**](DefaultApi.md#apiPresetIdDelete) | **DELETE** /api/preset/{id} | 删除预设
[**apiPresetIdGet**](DefaultApi.md#apiPresetIdGet) | **GET** /api/preset/{id} | 获取预设详情
[**apiPresetIdPut**](DefaultApi.md#apiPresetIdPut) | **PUT** /api/preset/{id} | 更新预设
[**apiPresetListGet**](DefaultApi.md#apiPresetListGet) | **GET** /api/preset/list | 获取预设列表
[**apiPresetPost**](DefaultApi.md#apiPresetPost) | **POST** /api/preset | 创建预设
[**apiUserInfoGet**](DefaultApi.md#apiUserInfoGet) | **GET** /api/user/info | 获取当前用户信息
[**apiUserLoginPost**](DefaultApi.md#apiUserLoginPost) | **POST** /api/user/login | 用户登录
[**apiUserLogoutPost**](DefaultApi.md#apiUserLogoutPost) | **POST** /api/user/logout | 退出登录
[**apiUserPasswordPut**](DefaultApi.md#apiUserPasswordPut) | **PUT** /api/user/password | 修改密码
[**apiUserProfilePut**](DefaultApi.md#apiUserProfilePut) | **PUT** /api/user/profile | 更新个人信息
[**apiUserRefreshPost**](DefaultApi.md#apiUserRefreshPost) | **POST** /api/user/refresh | 刷新访问令牌
[**apiUserRegisterPost**](DefaultApi.md#apiUserRegisterPost) | **POST** /api/user/register | 用户注册
[**apiUserResetPasswordPost**](DefaultApi.md#apiUserResetPasswordPost) | **POST** /api/user/reset-password | 重置密码



## apiAdminChannelIdDelete

> UtilsResponse apiAdminChannelIdDelete(id)

删除渠道

删除指定渠道

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 渠道ID
apiInstance.apiAdminChannelIdDelete(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 渠道ID | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiAdminChannelIdGet

> ApiAdminChannelPost200Response apiAdminChannelIdGet(id)

获取渠道详情

获取指定渠道的详细信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 渠道ID
apiInstance.apiAdminChannelIdGet(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 渠道ID | 

### Return type

[**ApiAdminChannelPost200Response**](ApiAdminChannelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiAdminChannelIdPut

> ApiAdminChannelPost200Response apiAdminChannelIdPut(id, data)

更新渠道

更新渠道信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 渠道ID
let data = new TxingAiApi.DtoUpdateChannelReq(); // DtoUpdateChannelReq | 渠道信息
apiInstance.apiAdminChannelIdPut(id, data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 渠道ID | 
 **data** | [**DtoUpdateChannelReq**](DtoUpdateChannelReq.md)| 渠道信息 | 

### Return type

[**ApiAdminChannelPost200Response**](ApiAdminChannelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiAdminChannelListGet

> UtilsResponse apiAdminChannelListGet(page, limit, opts)

获取渠道列表

获取渠道列表，支持分页

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let page = 56; // Number | 页码
let limit = 56; // Number | 每页数量
let opts = {
  'orderBy': "orderBy_example", // String | 排序字段
  'order': "order_example", // String | 排序方式(asc/desc)
  'type': "type_example", // String | 渠道类型
  'status': true, // Boolean | 状态
  'name': "name_example" // String | 渠道名称
};
apiInstance.apiAdminChannelListGet(page, limit, opts).then((data) => {
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
 **type** | **String**| 渠道类型 | [optional] 
 **status** | **Boolean**| 状态 | [optional] 
 **name** | **String**| 渠道名称 | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiAdminChannelPost

> ApiAdminChannelPost200Response apiAdminChannelPost(data)

创建渠道

创建新的渠道

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoCreateChannelReq(); // DtoCreateChannelReq | 渠道信息
apiInstance.apiAdminChannelPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoCreateChannelReq**](DtoCreateChannelReq.md)| 渠道信息 | 

### Return type

[**ApiAdminChannelPost200Response**](ApiAdminChannelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiAdminModelIdDelete

> UtilsResponse apiAdminModelIdDelete(id)

删除模型

删除指定模型

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 模型ID
apiInstance.apiAdminModelIdDelete(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 模型ID | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiAdminModelIdPut

> ApiAdminModelPost200Response apiAdminModelIdPut(id, data)

更新模型

更新模型信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 模型ID
let data = new TxingAiApi.DtoUpdateModelReq(); // DtoUpdateModelReq | 模型信息
apiInstance.apiAdminModelIdPut(id, data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 模型ID | 
 **data** | [**DtoUpdateModelReq**](DtoUpdateModelReq.md)| 模型信息 | 

### Return type

[**ApiAdminModelPost200Response**](ApiAdminModelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiAdminModelPost

> ApiAdminModelPost200Response apiAdminModelPost(data)

创建模型

创建新的模型

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoCreateModelReq(); // DtoCreateModelReq | 模型信息
apiInstance.apiAdminModelPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoCreateModelReq**](DtoCreateModelReq.md)| 模型信息 | 

### Return type

[**ApiAdminModelPost200Response**](ApiAdminModelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiAdminUserListGet

> UtilsResponse apiAdminUserListGet(page, limit, opts)

获取用户列表

获取用户列表，支持分页

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let page = 56; // Number | 页码
let limit = 56; // Number | 每页数量
let opts = {
  'orderBy': "orderBy_example", // String | 排序字段
  'order': "order_example", // String | 排序方式(asc/desc)
  'username': "username_example", // String | 用户名
  'status': 56 // Number | 状态(0:启用, 1:禁用)
};
apiInstance.apiAdminUserListGet(page, limit, opts).then((data) => {
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
 **username** | **String**| 用户名 | [optional] 
 **status** | **Number**| 状态(0:启用, 1:禁用) | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiAdminUserStatusIdPut

> UtilsResponse apiAdminUserStatusIdPut(id)

切换用户状态

启用或禁用指定用户

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 用户ID
apiInstance.apiAdminUserStatusIdPut(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 用户ID | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiCaptchaGet

> UtilsResponse apiCaptchaGet()

生成验证码

生成图片验证码

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
apiInstance.apiCaptchaGet().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiCosPresignedUrlPost

> ApiCosPresignedUrlPost200Response apiCosPresignedUrlPost(data)

获取预签名URL

获取文件上传或下载的预签名URL

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
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


## apiModelIdGet

> ApiAdminModelPost200Response apiModelIdGet(id)

获取模型详情

获取指定模型的详细信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 模型ID
apiInstance.apiModelIdGet(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 模型ID | 

### Return type

[**ApiAdminModelPost200Response**](ApiAdminModelPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiModelListGet

> UtilsResponse apiModelListGet(page, limit, opts)

获取模型列表

获取模型列表，支持分页

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let page = 56; // Number | 页码
let limit = 56; // Number | 每页数量
let opts = {
  'orderBy': "orderBy_example", // String | 排序字段
  'order': "order_example", // String | 排序方式(asc/desc)
  'tag': "tag_example", // String | 标签
  '_default': true, // Boolean | 是否默认
  'name': "name_example" // String | 模型名称
};
apiInstance.apiModelListGet(page, limit, opts).then((data) => {
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
 **tag** | **String**| 标签 | [optional] 
 **_default** | **Boolean**| 是否默认 | [optional] 
 **name** | **String**| 模型名称 | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiPresetIdDelete

> UtilsResponse apiPresetIdDelete(id)

删除预设

删除指定预设

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 预设ID
apiInstance.apiPresetIdDelete(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 预设ID | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiPresetIdGet

> ApiPresetPost200Response apiPresetIdGet(id)

获取预设详情

获取指定预设的详细信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 预设ID
apiInstance.apiPresetIdGet(id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 预设ID | 

### Return type

[**ApiPresetPost200Response**](ApiPresetPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiPresetIdPut

> ApiPresetPost200Response apiPresetIdPut(id, data)

更新预设

更新预设信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let id = 56; // Number | 预设ID
let data = new TxingAiApi.DtoUpdatePresetReq(); // DtoUpdatePresetReq | 预设信息
apiInstance.apiPresetIdPut(id, data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| 预设ID | 
 **data** | [**DtoUpdatePresetReq**](DtoUpdatePresetReq.md)| 预设信息 | 

### Return type

[**ApiPresetPost200Response**](ApiPresetPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiPresetListGet

> UtilsResponse apiPresetListGet(page, limit, opts)

获取预设列表

获取预设列表，支持分页

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let page = 56; // Number | 页码
let limit = 56; // Number | 每页数量
let opts = {
  'orderBy': "orderBy_example", // String | 排序字段
  'order': "order_example", // String | 排序方式(asc/desc)
  'official': true, // Boolean | 是否官方预设
  'userId': 56, // Number | 用户ID
  'name': "name_example", // String | 预设名称
  'tags': "tags_example" // String | 预设标签
};
apiInstance.apiPresetListGet(page, limit, opts).then((data) => {
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
 **official** | **Boolean**| 是否官方预设 | [optional] 
 **userId** | **Number**| 用户ID | [optional] 
 **name** | **String**| 预设名称 | [optional] 
 **tags** | **String**| 预设标签 | [optional] 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiPresetPost

> ApiPresetPost200Response apiPresetPost(data)

创建预设

创建新的预设

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoCreatePresetReq(); // DtoCreatePresetReq | 预设信息
apiInstance.apiPresetPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoCreatePresetReq**](DtoCreatePresetReq.md)| 预设信息 | 

### Return type

[**ApiPresetPost200Response**](ApiPresetPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiUserInfoGet

> ApiUserInfoGet200Response apiUserInfoGet()

获取当前用户信息

获取当前登录用户的详细信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
apiInstance.apiUserInfoGet().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**ApiUserInfoGet200Response**](ApiUserInfoGet200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiUserLoginPost

> ApiUserLoginPost200Response apiUserLoginPost(data)

用户登录

用户登录并返回token

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoLoginReq(); // DtoLoginReq | 登录信息
apiInstance.apiUserLoginPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoLoginReq**](DtoLoginReq.md)| 登录信息 | 

### Return type

[**ApiUserLoginPost200Response**](ApiUserLoginPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiUserLogoutPost

> UtilsResponse apiUserLogoutPost()

退出登录

清除用户登录状态

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
apiInstance.apiUserLogoutPost().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiUserPasswordPut

> UtilsResponse apiUserPasswordPut(data)

修改密码

修改当前登录用户的密码

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoUpdatePasswordReq(); // DtoUpdatePasswordReq | 密码信息
apiInstance.apiUserPasswordPut(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoUpdatePasswordReq**](DtoUpdatePasswordReq.md)| 密码信息 | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiUserProfilePut

> ApiUserInfoGet200Response apiUserProfilePut(data)

更新个人信息

更新当前登录用户的个人信息

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoUpdateProfileReq(); // DtoUpdateProfileReq | 个人信息
apiInstance.apiUserProfilePut(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoUpdateProfileReq**](DtoUpdateProfileReq.md)| 个人信息 | 

### Return type

[**ApiUserInfoGet200Response**](ApiUserInfoGet200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiUserRefreshPost

> ApiUserRefreshPost200Response apiUserRefreshPost(authorization)

刷新访问令牌

使用刷新令牌获取新的访问令牌

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let authorization = "authorization_example"; // String | Bearer 刷新令牌
apiInstance.apiUserRefreshPost(authorization).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **String**| Bearer 刷新令牌 | 

### Return type

[**ApiUserRefreshPost200Response**](ApiUserRefreshPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## apiUserRegisterPost

> ApiUserInfoGet200Response apiUserRegisterPost(data)

用户注册

新用户注册

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoRegisterReq(); // DtoRegisterReq | 注册信息
apiInstance.apiUserRegisterPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoRegisterReq**](DtoRegisterReq.md)| 注册信息 | 

### Return type

[**ApiUserInfoGet200Response**](ApiUserInfoGet200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## apiUserResetPasswordPost

> UtilsResponse apiUserResetPasswordPost(data)

重置密码

通过邮箱重置密码

### Example

```javascript
import TxingAiApi from 'txing_ai_api';

let apiInstance = new TxingAiApi.DefaultApi();
let data = new TxingAiApi.DtoResetPasswordReq(); // DtoResetPasswordReq | 重置密码信息
apiInstance.apiUserResetPasswordPost(data).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**DtoResetPasswordReq**](DtoResetPasswordReq.md)| 重置密码信息 | 

### Return type

[**UtilsResponse**](UtilsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

