package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	fmt.Println("Hello from Go")

	r.Run()
}
