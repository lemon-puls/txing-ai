package config

import (
	"github.com/gin-gonic/gin/binding"
	en "github.com/go-playground/locales/en"
	zh "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"
	"reflect"
	"strings"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	Trans    ut.Translator
)

func InitTranslator(language string) error {
	// 拿到 gin 里面的默认校验器
	validate = binding.Validator.Engine().(*validator.Validate)

	// 注册自定义的 tag name
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	zh := zh.New()
	en := en.New()

	uni = ut.New(zh, zh, en)

	// 获取到指定的语言翻译器
	var ok bool
	Trans, ok = uni.GetTranslator(language)
	if !ok {
		zap.L().Error("uni.GetTranslator failed", zap.String("language", language))
		Trans = uni.GetFallback()
		zap.L().Info("uni.GetFallback", zap.String("language", Trans.Locale()))
	}

	// 注册翻译器到 gin 校验器
	switch language {
	case "zh":
		zh_translations.RegisterDefaultTranslations(validate, Trans)
	case "en":
		en_translations.RegisterDefaultTranslations(validate, Trans)
	default:
		zh_translations.RegisterDefaultTranslations(validate, Trans)
	}
	return nil
}
