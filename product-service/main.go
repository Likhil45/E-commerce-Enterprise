package main

import (
	"e-commerce/product-service/handler"

	"github.com/gin-gonic/gin"
)

func main() {

	rout := gin.Default()
	r := rout.Group("/prod")

	r.PUT("/update", handler.UpdateProduct)
	r.POST("/create", handler.CreateProduct)
	r.DELETE("/delete/:id", handler.DeleteProduct)
	r.GET("/all", handler.GetAllProducts)
	r.GET("/:id", handler.GetProduct)

	rout.Run(":8081")

}
