package main

import (
	"e-commerce/gateway/controller"
	"e-commerce/gateway/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := gin.Default()
	r.POST("/signup", controller.CreateUser)
	r.POST("/login", controller.Login)
	r.GET("/all", controller.GetAllUsers)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authorized := r.Group("/user")
	authorized.Use(middleware.AuthMiddleware())
	r.GET("/:id", controller.GetUser)
	authorized.POST("/order/create", controller.CreateOrder)
	authorized.GET("/order", controller.GetOrder)
	authorized.GET("/test", controller.TestHandler)
	r.GET("/update/pd", controller.AddPaymentDetails)

	r.Run(":8080")
}
