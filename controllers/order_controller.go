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
	"github.com/Dbriane208/stablebase-go-sdk/order"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func getOrderClient() *order.Order {
	baseClient := networks.GetBaseClient()
	if baseClient == nil {
		return nil
	}
	return order.New(baseClient)
}

func PrepareApproveToken(ctx *gin.Context) {
	var input models.ApproveTokenRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tokenAddress := common.HexToAddress(input.TokenAddress)
	if tokenAddress == (common.Address{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token address",
		})
		return
	}

	amount := new(big.Int)
	amount, ok := amount.SetString(input.Amount, 10)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid amount",
		})
		return
	}

	contractABI, err := abi.GetERC20ABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse ERC20 ABI: " + err.Error(),
		})
		return
	}

	spenderAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	callData, err := contractABI.Pack("approve", spenderAddress, amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.PrepareApproveResponse{
		TransactionData: models.TransactionData{
			To:       tokenAddress.Hex(),
			Data:     "0x" + common.Bytes2Hex(callData),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 100000,
		},
		TokenAddress: input.TokenAddress,
		Spender:      spenderAddress.Hex(),
		Amount:       input.Amount,
		Message:      "Sign this transaction to approve PaymentProcessor to spend your tokens",
	})
}

func ConfirmApproveToken(ctx *gin.Context) {

	var input models.ConfirmApproveRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "Token approval confirmed",
		"transactionHash": input.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + input.TransactionHash,
	})
}

// PrepareCreateOrder prepares an unsigned transaction for creating an order
func PrepareCreateOrder(ctx *gin.Context) {
	var req models.CreateOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var merchantId [32]byte
	merchantIdBytes, err := hex.DecodeString(strings.TrimPrefix(req.MerchantId, "0x"))
	if err != nil || len(merchantIdBytes) != 32 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid merchantId format (must be 32 bytes hex string)",
		})
		return
	}
	copy(merchantId[:], merchantIdBytes)

	if !common.IsHexAddress(req.TokenAddress) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token address",
		})
		return
	}
	tokenAddress := common.HexToAddress(req.TokenAddress)

	amount := new(big.Int)
	amount, ok := amount.SetString(req.Amount, 10)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid amount format",
		})
		return
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Amount must be greater than zero",
		})
		return
	}

	if req.MetadataURI == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Metadata URI is required",
		})
		return
	}

	contractABI, err := abi.GetPaymentProcessorABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PaymentProcessor ABI: " + err.Error(),
		})
		return
	}

	data, err := contractABI.Pack("createOrder", merchantId, tokenAddress, amount, req.MetadataURI)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	paymentProcessorAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	// Return unsigned transaction data
	response := models.PrepareCreateOrderResponse{
		TransactionData: models.TransactionData{
			To:       paymentProcessorAddress.Hex(),
			Data:     "0x" + hex.EncodeToString(data),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 300000,
		},
		MerchantId:   req.MerchantId,
		TokenAddress: req.TokenAddress,
		Amount:       req.Amount,
		MetadataURI:  req.MetadataURI,
		Message:      "Please sign with your wallet and submit the transaction hash to confirm.",
	}

	ctx.JSON(http.StatusOK, response)
}

func ConfirmCreateOrder(ctx *gin.Context) {
	var req models.ConfirmCreateOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !common.IsHexAddress(req.TransactionHash) && len(req.TransactionHash) != 66 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction hash format",
		})
		return
	}

	if !common.IsHexAddress(req.PayerAddress) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payer address",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	merchantIdHex := req.MerchantId
	if !strings.HasPrefix(merchantIdHex, "0x") {
		merchantIdHex = "0x" + merchantIdHex
	}

	var merchants []map[string]interface{}
	err := db.Supabase.DB.From("merchants").Select("*").Eq("merchantId", merchantIdHex).Execute(&merchants)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch merchant: " + err.Error(),
		})
		return
	}

	if len(merchants) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Merchant not found",
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

	contractABI, err := abi.GetPaymentProcessorABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PaymentProcessor ABI: " + err.Error(),
		})
		return
	}

	var orderId string
	for _, log := range receipt.Logs {
		event := struct {
			OrderId [32]byte
		}{}

		err := contractABI.UnpackIntoInterface(&event, "OrderCreated", log.Data)
		if err == nil {
			orderId = "0x" + hex.EncodeToString(event.OrderId[:])
			break
		}
	}

	if orderId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Order created but orderId not found in transaction logs",
		})
		return
	}

	dbOrder := models.OrderDB{
		OrderId:         orderId,
		MerchantId:      merchantIdHex,
		PayerAddress:    req.PayerAddress,
		TokenAddress:    req.TokenAddress,
		Amount:          req.Amount,
		Status:          "created",
		MetadataURI:     req.MetadataURI,
		TransactionHash: req.TransactionHash,
	}

	var result []models.OrderDB
	if err := db.Supabase.DB.From("orders").Insert(dbOrder).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Order not saved to database: " + err.Error(),
		})
		return
	}

	if len(result) > 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success":     true,
			"order":       result[0],
			"explorerUrl": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"success":     true,
			"order":       dbOrder,
			"explorerUrl": networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
		})
	}
}

func PreparePayOrder(ctx *gin.Context) {
	var req models.PrepareOrder

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

	contractABI, err := abi.GetPaymentProcessorABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PaymentProcessor ABI: " + err.Error(),
		})
		return
	}

	data, err := contractABI.Pack("payOrder", orderId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	paymentProcessorAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	// Return unsigned transaction data
	response := models.PreparePayOrderResponse{
		TransactionData: models.TransactionData{
			To:       paymentProcessorAddress.Hex(),
			Data:     "0x" + hex.EncodeToString(data),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 300000,
		},
		OrderId: req.OrderId,
		Message: "Please sign with your wallet to Pay the order.",
	}

	ctx.JSON(http.StatusOK, response)
}

func ConfirmPayOrder(ctx *gin.Context) {
	var req models.ConfirmPayOrderRequest

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
		"status":          "paid",
		"transactionHash": req.TransactionHash,
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("orders").Update(updates).Eq("orderId", orderIdHex).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update order status: " + err.Error(),
		})
		return
	}

	response := gin.H{
		"success":         true,
		"orderId":         orderIdHex,
		"message":         "Order paid successfully",
		"status":          "paid",
		"transactionHash": req.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + req.TransactionHash,
	}

	ctx.JSON(http.StatusOK, response)
}

func CancelOrder(ctx *gin.Context) {
	var req models.PrepareOrder

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

	o := getOrderClient()
	bgCtx := context.Background()
	_, receipt, err := o.CancelOrder(bgCtx, orderId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"status":          "cancelled",
		"transactionHash": receipt.TxHash.Hex(),
	}

	var result []map[string]interface{}
	err = db.Supabase.DB.From("orders").Update(updates).Eq("orderId", orderIdHex).Execute(&result)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update order status: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orderId":         orderIdHex,
		"status":          "cancelled",
		"transactionHash": receipt.TxHash.Hex(),
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	})
}
