package router

import (
	"code_limit/handler"
	"github.com/gin-gonic/gin"
)

func Api(engine *gin.Engine) {

	engine.GET("/gen_code", handler.GenVerificationCode)
	engine.POST("/verify_code", handler.CheckVerificationCode)
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

}
