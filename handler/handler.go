package handler

import (
	"code_limit/common"
	"code_limit/helpers"
	"code_limit/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GenVerificationCode(ctx *gin.Context) {
	svc := service.CodeService{}
	var codeSession string
	codeSession, _ = ctx.Cookie(common.VerificationCodeCookie)
	if codeSession == "" {
		codeSession = helpers.GenerateRandomString(16)
		logrus.Infof("set cookie session_value=%s", codeSession)
		ctx.SetCookie(common.VerificationCodeCookie, codeSession, 86400, "/", "", false, false)
	}

	img, err := svc.GenCode(ctx, codeSession)
	if err != nil {
		logrus.Errorf("gen verification code failed, err=%v", err)
		helpers.RenderJsonFail(ctx, err)
		return
	}
	helpers.RenderJsonSucc(ctx, common.GetVerificationCodeResp{ImgBase64: img})
}

func CheckVerificationCode(ctx *gin.Context) {
	svc := service.CodeService{}
	codeSession, err := ctx.Cookie(common.VerificationCodeCookie)
	if err != nil {
		logrus.Warnf("[CheckVerificationCode] get cookie error: %v", err)
		helpers.RenderJsonFail(ctx, err)
		return
	}
	in := common.CheckVerificationCodeReq{}
	if err1 := ctx.ShouldBind(&in); err != nil {
		logrus.Warnf("[CheckVerificationCode] params error: %v", err1)
		helpers.RenderJsonFail(ctx, common.ErrorParamInvalid)
		return
	}
	err = svc.Verify(ctx, codeSession, in.Code)
	if err != nil {
		helpers.RenderJsonFail(ctx, err)
		return
	}
	helpers.RenderJsonSucc(ctx, nil)
}
