package routes

import (
	"github.com/Dbriane208/stable-market/controllers"
	"github.com/gin-gonic/gin"
)

func SetupMarketRoutes(router *gin.Engine) {
	market := router.Group("/api/market")
	{
		market.POST("/add-product", controllers.CreateProduct)
	}
}
