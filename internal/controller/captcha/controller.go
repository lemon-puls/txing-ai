package captcha

import (
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/captcha"

	"github.com/gin-gonic/gin"
)

// Generate 生成验证码
// @Summary 生成验证码
// @Description 生成图片验证码
// @Tags 验证码
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/captcha [get]
func Generate(ctx *gin.Context) {
	id, b64s, err := captcha.Generate()
	if err != nil {
		utils.ErrorWithMsg(ctx, "生成验证码失败", err)
		return
	}
	utils.OkWithData(ctx, gin.H{
		"id":    id,
		"image": b64s,
	})
}
