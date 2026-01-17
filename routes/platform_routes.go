package routes

import (
	"github.com/Dbriane208/stable-market/controllers"
	"github.com/gin-gonic/gin"
)

// SetupPlatformRoutes configures platform-related routes
func SetupPlatformRoutes(router *gin.Engine) {
	platform := router.Group("/api/platform")
	{
		platform.POST("/emergency-withdrawal", controllers.EmergencyWithdraw)
		platform.POST("/enable-emergency-withdrawal", controllers.SetEmergencyWithdrawalEnabled)
		platform.POST("/update-merchant-registry", controllers.UpdateMerchantRegistry)
		platform.POST("/set-token-support", controllers.SetTokenSupport)
		platform.POST("/merchant-verification-status", controllers.UpdateMerchantVerificationStatus)
		platform.GET("/token-balance", controllers.GetPlatformTokenBalance)
		platform.GET("/contract-token-balance", controllers.GetContractTokenBalance)

		// Token approval with frontend signing
		platform.POST("/approve-token", controllers.PrepareApproveToken)
		platform.POST("/confirm-approve", controllers.ConfirmApproveToken)

		// Settle order with frontend signing
		platform.POST("/prepare-settle", controllers.PrepareSettleOrder)
		platform.POST("/confirm-settle", controllers.ConfirmSettleOrder)

		// Refund order with frontend signing
		platform.POST("/prepare-refund", controllers.PrepareRefundOrder)
		platform.POST("/confirm-refund", controllers.ConfirmRefundOrder)
	}
}
