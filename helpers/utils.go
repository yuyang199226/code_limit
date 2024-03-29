package helpers

import (
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"math/rand"
)

func IsStrIn(sub string, ls []string) bool {
	for _, v := range ls {
		if sub == v {
			return true
		}
	}
	return false
}

func GenerateRandomString(num int) string {
	randomBytes := make([]byte, num/2)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(randomBytes)
}

type BaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func RenderJsonSucc(ctx *gin.Context, v any) {
	resp := BaseResponse{
		Code: 200,
		Msg:  "success",
		Data: v,
	}
	r := render.JSON{resp}
	r.Render(ctx.Writer)
}

func RenderJsonFail(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(200, err)
}
