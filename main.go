package main

import (
	"code_limit/helpers"
	router2 "code_limit/router"
	"github.com/gin-gonic/gin"
)

func main() {
	eng := gin.New()
	router2.Api(eng)
	helpers.InitRedis()
	eng.Run(":8080")

}
