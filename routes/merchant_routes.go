package routes

import (
	"github.com/Dbriane208/stable-market/controllers"
	"github.com/gin-gonic/gin"
)

// SetupMerchantRoutes configures merchant-related routes
func SetupMerchantRoutes(router *gin.Engine) {
	merchant := router.Group("/api/merchants")
	{
		merchant.POST("/register", controllers.RegisterMerchant)
		merchant.GET("/merchant-info/:merchantId", controllers.GetMerchantInfoById)
		merchant.DELETE("/delete/:merchantId", controllers.DeleteMerchant)
		merchant.GET("/balance/:merchantId", controllers.GetMerchantBalance)
		merchant.GET("/merchant-status/:merchantId", controllers.IsMerchantVerified)

		// Frontend signing endpoints for merchant updates
		merchant.POST("/prepare-update/:merchantId", controllers.PrepareUpdateMerchant)
		merchant.POST("/confirm-update/:merchantId", controllers.ConfirmMerchantUpdate)

		// Frontend signing endpoints for order refunds
		merchant.POST("/prepare-refund/:orderId", controllers.PrepareRefundOrderMerchant)
		merchant.POST("/confirm-refund/:orderId", controllers.ConfirmRefundOrderMerchant)
	}
}
