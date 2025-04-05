package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"io/ioutil"
	"strings"
	"txing-ai/internal/global"
	"txing-ai/internal/global/config"
	"txing-ai/internal/global/logging/log"
)

func ValidateError(c *gin.Context, err error) {
	// 打印出失败请求的 body
	body, _ := ioutil.ReadAll(c.Request.Body)
	log.Error("validate error", zap.Error(err), zap.String("body", string(body)))
	errors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 非参数校验错误
		log.Error("not a validate error", zap.Error(errors))
		ErrorWithMsg(c, err.Error(), err)
		return
	}

	// 参数校验错误
	errMsgMap := removeStructPrefix(errors.Translate(config.Trans))
	log.Error("validate error", zap.Any("errors", errMsgMap))
	ErrorWithData(c, global.CodeInvalidParams, errMsgMap, err)
}

// 错误信息中去除类型名 如：User.Name 转为 Name
func removeStructPrefix(errMsg map[string]string) map[string]string {
	res := make(map[string]string)
	for name, msg := range errMsg {
		res[name[strings.Index(name, ".")+1:]] = msg
	}
	return res
}
