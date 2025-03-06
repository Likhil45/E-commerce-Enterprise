package main

import (
	"e-commerce/gateway/controller"
	"e-commerce/gateway/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/signup", controller.CreateUser)
	r.POST("/login", controller.Login)
	r.GET("/all", controller.GetAllUsers)

	authorized := r.Group("/user")
	authorized.Use(middleware.AuthMiddleware())
	r.GET("/:id", controller.GetUser)
	authorized.POST("/order/create", controller.CreateOrder)

	r.Run(":8080")
}
