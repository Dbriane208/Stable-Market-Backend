package routes

import (
	"github.com/Dbriane208/stable-market/controllers"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(router *gin.Engine) {
	order := router.Group("/api/orders")
	{
		order.POST("/prepare-create", controllers.PrepareCreateOrder)
		order.POST("/confirm-create", controllers.ConfirmCreateOrder)
		order.POST("/prepare-pay-order", controllers.PreparePayOrder)
		order.POST("/confirm-pay-order", controllers.ConfirmPayOrder)
		order.POST("/cancel", controllers.CancelOrder)
	}
}
