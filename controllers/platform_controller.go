package controllers

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/http"
	"strings"

	"github.com/Dbriane208/stable-market/abi"
	"github.com/Dbriane208/stable-market/db"
	"github.com/Dbriane208/stable-market/models"
	"github.com/Dbriane208/stable-market/networks"
	"github.com/Dbriane208/stablebase-go-sdk/platform"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func getPlatformClient() *platform.Platform {
	baseClient := networks.GetBaseClient()
	if baseClient == nil {
		return nil
	}
	return platform.New(baseClient)
}

func PrepareSettleOrder(ctx *gin.Context) {
	var req models.PrepareSettleOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var orderId [32]byte
	orderIdBytes, err := hex.DecodeString(strings.TrimPrefix(req.OrderId, "0x"))
	if err != nil || len(orderIdBytes) != 32 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid orderId format (must be 32 bytes hex string)",
		})
		return
	}
	copy(orderId[:], orderIdBytes)

	orderIdHex := req.OrderId
	if !strings.HasPrefix(orderIdHex, "0x") {
		orderIdHex = "0x" + orderIdHex
	}

	if db.Supabase != nil {

		var existingOrders []map[string]interface{}
		err = db.Supabase.DB.From("orders").Select("*").Eq("orderId", orderIdHex).Execute(&existingOrders)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch order: " + err.Error(),
			})
			return
		}

		if len(existingOrders) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}

		if status, ok := existingOrders[0]["status"].(string); ok && status != "paid" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Order must be in 'paid' status to be settled. Current status: " + status,
			})
			return
		}
	}

	contractABI, err := abi.GetPaymentProcessorABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PaymentProcessor ABI: " + err.Error(),
		})
		return
	}

	data, err := contractABI.Pack("settleOrder", orderId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	paymentProcessorAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	response := models.PrepareSettleOrderResponse{
		TransactionData: models.TransactionData{
			To:       paymentProcessorAddress.Hex(),
			Data:     "0x" + hex.EncodeToString(data),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 300000,
		},
		OrderId: orderIdHex,
		Message: "Please sign with your wallet to settle the order and transfer funds to merchant.",
	}

	ctx.JSON(http.StatusOK, response)
}

func ConfirmSettleOrder(ctx *gin.Context) {
	var req models.ConfirmSettleOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(req.TransactionHash) != 66 || !strings.HasPrefix(req.TransactionHash, "0x") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction hash format",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	orderIdHex := req.OrderId
	if !strings.HasPrefix(orderIdHex, "0x") {
		orderIdHex = "0x" + orderIdHex
	}

	var existingOrders []map[string]interface{}
	err := db.Supabase.DB.From("orders").Select("*").Eq("orderId", orderIdHex).Execute(&existingOrders)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch order: " + err.Error(),
		})
		return
	}

	if len(existingOrders) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	sdkClient := networks.GetBaseClient()
	if sdkClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize blockchain client",
		})
		return
	}

	txHash := common.HexToHash(req.TransactionHash)
	bgCtx := context.Background()
	receipt, err := sdkClient.EthClient.TransactionReceipt(bgCtx, txHash)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Transaction not found or not yet mined: " + err.Error(),
			"message": "Please wait for the transaction to be confirmed and try again",
		})
		return
	}

	if receipt.Status == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction failed on blockchain",
		})
		return
	}

	updates := map[string]interface{}{
		"status":          "settled",
		"transactionHash": req.TransactionHash,
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("orders").Update(updates).Eq("orderId", orderIdHex).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update order status: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":         true,
		"orderId":         orderIdHex,
		"message":         "Order settled successfully. Funds transferred to merchant.",
		"status":          "settled",
		"transactionHash": req.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
	})
}

func PrepareRefundOrder(ctx *gin.Context) {
	var req models.PrepareRefundOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var orderId [32]byte
	orderIdBytes, err := hex.DecodeString(strings.TrimPrefix(req.OrderId, "0x"))
	if err != nil || len(orderIdBytes) != 32 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid orderId format (must be 32 bytes hex string)",
		})
		return
	}
	copy(orderId[:], orderIdBytes)

	orderIdHex := req.OrderId
	if !strings.HasPrefix(orderIdHex, "0x") {
		orderIdHex = "0x" + orderIdHex
	}

	if db.Supabase != nil {
		var existingOrders []map[string]interface{}

		err = db.Supabase.DB.From("orders").Select("*").Eq("orderId", orderIdHex).Execute(&existingOrders)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch order: " + err.Error(),
			})
			return
		}

		if len(existingOrders) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}

		if status, ok := existingOrders[0]["status"].(string); ok {
			if status != "paid" && status != "settled" {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": "Order must be in 'paid' or 'settled' status to be refunded. Current status: " + status,
				})
				return
			}
		}
	}

	contractABI, err := abi.GetPaymentProcessorABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PaymentProcessor ABI: " + err.Error(),
		})
		return
	}

	data, err := contractABI.Pack("refundOrder", orderId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	paymentProcessorAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	response := models.PrepareRefundResponse{
		TransactionData: models.TransactionData{
			To:       paymentProcessorAddress.Hex(),
			Data:     "0x" + hex.EncodeToString(data),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 300000,
		},
		OrderId: orderIdHex,
		Message: "Please sign with your wallet to refund the order to the payer.",
	}

	ctx.JSON(http.StatusOK, response)
}

func ConfirmRefundOrder(ctx *gin.Context) {
	var req models.ConfirmRefundOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(req.TransactionHash) != 66 || !strings.HasPrefix(req.TransactionHash, "0x") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction hash format",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	orderIdHex := req.OrderId
	if !strings.HasPrefix(orderIdHex, "0x") {
		orderIdHex = "0x" + orderIdHex
	}

	var existingOrders []map[string]interface{}
	err := db.Supabase.DB.From("orders").Select("*").Eq("orderId", orderIdHex).Execute(&existingOrders)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch order: " + err.Error(),
		})
		return
	}

	if len(existingOrders) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	sdkClient := networks.GetBaseClient()
	if sdkClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize blockchain client",
		})
		return
	}

	txHash := common.HexToHash(req.TransactionHash)
	bgCtx := context.Background()
	receipt, err := sdkClient.EthClient.TransactionReceipt(bgCtx, txHash)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Transaction not found or not yet mined: " + err.Error(),
			"message": "Please wait for the transaction to be confirmed and try again",
		})
		return
	}

	if receipt.Status == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction failed on blockchain",
		})
		return
	}

	updates := map[string]interface{}{
		"status":          "refunded",
		"transactionHash": req.TransactionHash,
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("orders").Update(updates).Eq("orderId", orderIdHex).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update order status: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":         true,
		"orderId":         orderIdHex,
		"message":         "Order refunded successfully. Funds returned to payer.",
		"status":          "refunded",
		"transactionHash": req.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
	})
}

func EmergencyWithdraw(ctx *gin.Context) {
	var req models.EmergencyWithdraw

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tokenAddress := common.HexToAddress(req.TokenAddress)
	if tokenAddress == (common.Address{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token address",
		})
		return
	}

	receiverAddress := common.HexToAddress(req.RecieverAddress)
	if tokenAddress == (common.Address{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid receiver address",
		})
		return
	}

	amount := new(big.Int)
	amount, ok := amount.SetString(req.Amount, 10)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid amount",
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()
	_, receipt, err := p.EmergencyWithdraw(bgCtx, tokenAddress, receiverAddress, amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to withraw on emergency: " + err.Error(),
		})
		return
	}

	req.TransactionHash = receipt.TxHash.Hex()

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	sdkClient := networks.GetBaseClient()
	if sdkClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize blockchain client",
		})
		return
	}

	dbEmergencyWithrawal := models.EmergencyWithdraw{
		TokenAddress:    req.TokenAddress,
		RecieverAddress: req.RecieverAddress,
		Amount:          req.Amount,
		SenderAddress:   sdkClient.PaymentProcessorAddress.Hex(),
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("emergencyWithdrawal").Insert(dbEmergencyWithrawal).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Emergency Withdrawal not saved in database: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"withdrawal":  dbEmergencyWithrawal,
		"explorerUrl": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
	})
}

func SetEmergencyWithdrawalEnabled(ctx *gin.Context) {
	var req models.WithdrawalStatus

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	_, receipt, err := p.SetEmergencyWithdrawalEnabled(bgCtx, *req.IsWithdrawalEnabled)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enable emergency withdrawal: " + err.Error(),
		})
		return
	}

	message := "Emergency withdrawal disabled"
	if *req.IsWithdrawalEnabled {
		message = "Emergency withdrawal enabled"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         message,
		"transactionHash": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	})
}

func UpdateMerchantRegistry(ctx *gin.Context) {
	var req models.MerchantRegistryUpdate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	newRegistryAddress := common.HexToAddress(req.NewRegistryAddress)

	_, receipt, err := p.UpdateMerchantRegistry(bgCtx, newRegistryAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update the merchantRegistry",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "Merchant Registry Updated Succesfully",
		"transactionHash": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	})
}

func SetTokenSupport(ctx *gin.Context) {
	var req models.TokenSupport

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	var statusValue *big.Int
	switch req.StatusValue {
	case "disabled":
		statusValue = big.NewInt(0)
	case "enabled":
		statusValue = big.NewInt(1)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status value. Must be: enabled or disabled",
		})
		return
	}

	tokenAddress := common.HexToAddress(req.TokenAddress)

	_, receipt, err := p.SetTokenSupport(bgCtx, tokenAddress, statusValue)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enable token support",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "Token Enabled Successfully",
		"transactionHash": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	})
}

func GetPlatformTokenBalance(ctx *gin.Context) {
	var req models.PlatformBalanceCheck

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	platformWallet := common.HexToAddress(req.PlatformWallet)
	tokenAddress := common.HexToAddress(req.TokenAddress)

	balance, err := p.GetPlatformTokenBalance(bgCtx, platformWallet, tokenAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get platform balance: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Balance retrieved successfully",
		"balance": balance.String(),
	})
}

func GetContractTokenBalance(ctx *gin.Context) {
	var req models.ContractBalanceCheck

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	tokenAddress := common.HexToAddress(req.TokenAddress)

	balance, err := p.GetContractTokenBalance(bgCtx, tokenAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get contract balance: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Balance retrieved successfully",
		"balance": balance.String(),
	})
}

func UpdateMerchantVerificationStatus(ctx *gin.Context) {
	var req models.UpdateMerchantVerificationStatus

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	var existingMerchant []map[string]interface{}
	err := db.Supabase.DB.From("merchants").Select("*").Eq("merchantId", req.MerchantId).Execute(&existingMerchant)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch merchant: " + err.Error(),
		})
		return
	}

	if len(existingMerchant) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Merchant not found",
		})
		return
	}

	p := getPlatformClient()
	if p == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize platform client",
		})
		return
	}
	bgCtx := context.Background()

	var merchantId [32]byte
	merchantIdBytes, _ := hex.DecodeString(strings.TrimPrefix(req.MerchantId, "0x"))
	copy(merchantId[:], merchantIdBytes)

	var verificationStatus uint8
	switch req.VerificationStatus {
	case "pending":
		verificationStatus = 0
	case "verified":
		verificationStatus = 1
	case "rejected":
		verificationStatus = 2
	case "suspended":
		verificationStatus = 3
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid verification status. Must be: pending, verified, rejected, or suspended",
		})
		return
	}

	_, receipt, err := p.UpdateMerchantVerificationStatus(bgCtx, merchantId, verificationStatus)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update merchant verification: " + err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"verificationStatus": req.VerificationStatus,
	}

	var result []map[string]interface{}
	err = db.Supabase.DB.From("merchants").Update(updates).Eq("merchantId", req.MerchantId).Execute(&result)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update merchant status: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"merchantId":         req.MerchantId,
		"verificationStatus": req.VerificationStatus,
		"explorerUrl":        networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	})
}