package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"txing-ai/internal/global"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func response(ctx *gin.Context, httpCode int, response *Response) {
	ctx.JSON(httpCode, response)
}

// FailWithMsg returns a failed response with a message
func ErrorWithMsg(ctx *gin.Context, msg string, err error) {

	fields := []zap.Field{
		zap.String("uri:", ctx.Request.RequestURI),
		zap.Int("code:", global.CodeServerInternalError),
		zap.String("msg:", msg),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	zap.L().Error("return error", fields...)

	response(ctx, http.StatusOK, &Response{
		Code: global.CodeServerInternalError,
		Msg:  msg,
		Data: nil,
	})
}

// FailWithCode returns a failed response with a code
func ErrorWithCode(ctx *gin.Context, code global.Code, err error) {
	fields := []zap.Field{
		zap.String("uri:", ctx.Request.RequestURI),
		zap.Int("code:", global.CodeServerInternalError),
		zap.String("msg:", code.ToMsg()),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	zap.L().Error("return error", fields...)

	response(ctx, http.StatusOK, &Response{
		Code: int(code),
		Msg:  code.ToMsg(),
		Data: nil,
	})
}

func ErrorWithCodeAndMsg(ctx *gin.Context, code global.Code, msg string, err error) {

	fields := []zap.Field{
		zap.String("uri:", ctx.Request.RequestURI),
		zap.Int("code:", int(code)),
		zap.String("msg:", msg),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	zap.L().Error("return error", fields...)

	response(ctx, http.StatusOK, &Response{
		Code: int(code),
		Msg:  msg,
		Data: nil,
	})
}

func ErrorWithData(ctx *gin.Context, code global.Code, data interface{}, err error) {

	fields := []zap.Field{
		zap.String("uri:", ctx.Request.RequestURI),
		zap.Int("code:", int(code)),
		zap.String("msg:", code.ToMsg()),
		zap.Any("data", data),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	zap.L().Error("return error", fields...)

	response(ctx, http.StatusOK, &Response{
		Code: int(code),
		Msg:  code.ToMsg(),
		Data: data,
	})
}

// Ok returns a successful response
func OkWithMsg(ctx *gin.Context, msg string) {
	response(ctx, http.StatusOK, &Response{
		Code: global.CodeSuccess,
		Msg:  msg,
		Data: nil,
	})
}

// OkWithData returns a successful response with data
func OkWithData(ctx *gin.Context, data interface{}) {
	response(ctx, http.StatusOK, &Response{
		Code: global.CodeSuccess,
		Msg:  global.Code(global.CodeSuccess).ToMsg(),
		Data: data,
	})
}
