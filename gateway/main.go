package main

import (
	"e-commerce/gateway/controller"
	"e-commerce/gateway/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/signup", controller.CreateUser)
	r.GET("/login", controller.Login)

	authorized := r.Group("/user")
	authorized.Use(middleware.AuthMiddleware())
	r.GET("/:id", controller.GetUser)

	r.Run(":8080")
}
