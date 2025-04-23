package main

import (
	"e-commerce/logger"
	"e-commerce/product-service/handler"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.InitLogger("product-service")
	handler.Init()
	rout := gin.Default()

	rout.Use(func(c *gin.Context) {
		c.Next()
		handler.HttpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(c.Writer.Status())).Inc()
	})

	r := rout.Group("/prod")

	r.PUT("/update", handler.UpdateProduct)
	r.POST("/create", handler.CreateProduct)
	r.DELETE("/delete/:id", handler.DeleteProduct)
	r.GET("/all", handler.GetAllProducts)
	r.GET("/:id", handler.GetProduct)

	rout.GET("/metrics", gin.WrapH(promhttp.Handler()))

	rout.Run(":8081")

}
