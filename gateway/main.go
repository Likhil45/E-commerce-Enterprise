package main

import (
	"e-commerce/gateway/controller"
	"e-commerce/gateway/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/signup", controller.CreateUser)
	r.GET("/login", controller.Login)
	r.GET("/all", controller.GetAllUsers)

	authorized := r.Group("/user")
	authorized.Use(middleware.AuthMiddleware())
	authorized.GET("/:id", controller.GetUser)

	r.Run(":8080")
}
