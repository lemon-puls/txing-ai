package cos

import (
	"net/http"
	"time"
	"txing-ai/internal/dto"
	"txing-ai/internal/vo"

	"txing-ai/internal/utils"

	"github.com/gin-gonic/gin"
)

// @Summary 获取预签名URL
// @Description 获取文件上传或下载的预签名URL
// @Tags 对象存储相关
// @Accept json
// @Produce json
// @Param data body dto.GetPresignedURLReq true "请求参数"
// @Success 200 {object} utils.Response{data=vo.GetPresignedURLVO} "成功"
// @Router /api/cos/presigned-url [post]
func GetPresignedURL(ctx *gin.Context) {
	var req dto.GetPresignedURLReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	// 默认过期时间 1 小时
	expire := 3600

	// 生成预签名URL
	var (
		url string
		err error
	)

	switch req.Type {
	case "upload":
		url, err = cosClient.GenerateUploadPresignedURL(req.Key)
	case "download":
		url, err = cosClient.GenerateDownloadPresignedURL(req.Key)
	default:
		utils.ErrorWithCode(ctx, http.StatusBadRequest, nil)
		return
	}

	if err != nil {
		utils.ErrorWithCode(ctx, http.StatusInternalServerError, err)
		return
	}

	// 计算过期时间戳
	expireAt := time.Now().Add(time.Duration(expire) * time.Second).Unix()

	utils.OkWithData(ctx, vo.GetPresignedURLVO{
		URL:      url,
		Key:      req.Key,
		ExpireAt: expireAt,
	})
}
