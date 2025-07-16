package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("main working")
	server := gin.Default()
	server.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "report server in")
	})

	initRoutes(server)
	// fmt.Printf("Server is listening on port: %d\n", 8082)
	server.Run(":8082")
}

func initRoutes(s *gin.Engine) {
	reportHdls := NewReportHanlder()
	reportHdls.RegisterRoutes(s)
}
