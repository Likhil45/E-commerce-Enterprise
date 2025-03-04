package main

import (
	"e-commerce/order-service/orderhand"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.POST("/order/create", orderhand.CreateOrder)
	r.Run(":8083")

}
