package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/Dbriane208/stable-market/db"
	"github.com/Dbriane208/stable-market/models"
	"github.com/Dbriane208/stable-market/networks"
	"github.com/Dbriane208/stablebase-go-sdk/merchant"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// getMerchantClient returns the merchant client instance
func getMerchantClient() *merchant.Merchant {
	baseClient := networks.GetBaseClient()
	if baseClient == nil {
		return nil
	}
	return merchant.New(baseClient)
}

// RegisterMerchant handles merchant registration
func RegisterMerchant(ctx *gin.Context) {
	var info models.MerchantInfo

	if err := ctx.ShouldBindJSON(&info); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	m := getMerchantClient()
	if m == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Network client not initialized",
		})
		return
	}

	bgCtx := context.Background()
	_, merchantId, receipt, err := m.RegisterMerchant(bgCtx, info.PayoutWalletAddress, info.MetadataURI)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
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

	dbMerchant := models.MerchantDB{
		MerchantName:        info.MerchantName,
		MerchantId:          common.BytesToHash(merchantId[:]).Hex(),
		PayoutWalletAddress: info.PayoutWalletAddress.Hex(),
		MetadataURI:         info.MetadataURI,
		TransactionHash:     receipt.TxHash.Hex(),
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("merchants").Insert(dbMerchant).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create merchant: " + err.Error(),
		})
		return
	}

	response := models.MerchantResponse{
		MerchantId:      merchantId,
		Message:         "Merchant registered successfully",
		MetadataURI:     info.MetadataURI,
		WalletAddress:   info.PayoutWalletAddress,
		MerchantName:    info.MerchantName,
		TransactionHash: receipt.TxHash,
		ExplorerURL:     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + receipt.TxHash.Hex(),
	}

	ctx.JSON(http.StatusOK, response)
}

// GetMerchantInfoById
func GetMerchantInfoById(ctx *gin.Context) {
	merchantId := ctx.Param("merchantId")

	if merchantId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "merchantId is required",
		})
		return
	}

	var merchants []models.MerchantDB
	if err := db.Supabase.DB.From("merchants").Select("*").Eq("merchantId", merchantId).Execute(&merchants); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not get merchant details: " + err.Error(),
		})
		return
	}

	if len(merchants) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Merchant not found",
		})
		return
	}

	info := merchants[0]
	ctx.JSON(http.StatusOK, &models.MerchantResponse{
		MerchantId:         common.HexToHash(info.MerchantId),
		Message:            "Success",
		MetadataURI:        info.MetadataURI,
		WalletAddress:      common.HexToAddress(info.PayoutWalletAddress),
		MerchantName:       info.MerchantName,
		TransactionHash:    common.HexToHash(info.TransactionHash),
		ExplorerURL:        networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + info.TransactionHash,
	})
}

func DeleteMerchant(ctx *gin.Context) {
	merchantId := ctx.Param("merchantId")

	if merchantId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "merchantId is required",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("merchants").Delete().Eq("merchantId", merchantId).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update merchant: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"merchantId": merchantId,
		"message":    "Merchant deleted successfully",
	})
}

func GetMerchantBalance(ctx *gin.Context) {
	merchantId := ctx.Param("merchantId")

	var data *models.TokenBalance

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	m := getMerchantClient()
	if m == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Network client not initialized",
		})
		return
	}

	bgCtx := context.Background()
	balance, err := m.GetMerchantTokenBalance(bgCtx, data.WalletAddress, data.TokenAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
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

	dbTokenBalance := &models.TokenBalanceDB{
		MerchantId:    merchantId,
		WalletAddress: data.WalletAddress.Hex(),
		TokenAddress:  data.TokenAddress.Hex(),
		TokenBalance:  balance.String(),
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.From("tokenBalance").Insert(dbTokenBalance).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create token balance: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, &models.TokenBalanceDB{
		MerchantId:    merchantId,
		WalletAddress: data.WalletAddress.Hex(),
		TokenAddress:  data.TokenAddress.Hex(),
		TokenBalance:  balance.String(),
	})
}

func IsMerchantVerified(ctx *gin.Context) {
	merchantId := ctx.Param("merchantId")

	if merchantId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "merchantId is required",
		})
		return
	}

	m := getMerchantClient()
	bgCtx := context.Background()
	isVerified, err := m.IsMerchantVerified(bgCtx, common.HexToHash(merchantId))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"isVerified": isVerified,
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ============================================
// FRONTEND SIGNING ENDPOINTS
// ============================================

func getMerchantRegistryABI() (abi.ABI, error) {
	// ABI for updateMerchant function: updateMerchant(bytes32 _merchantId, address _payoutWalletAdPayoutWalletAddress, string _metadataUri)
	const abiJSON = `[{"type":"function","name":"updateMerchant","inputs":[{"name":"_merchantId","type":"bytes32"},{"name":"_payoutWalletAdPayoutWalletAddress","type":"address"},{"name":"_metadataUri","type":"string"}],"outputs":[],"stateMutability":"nonpayable"}]`
	return abi.JSON(strings.NewReader(abiJSON))
}

func getPaymentProcessorRegistryABI() (abi.ABI, error) {
	const abiJSON = `[{"type":"function","name":"refundOrder","inputs":[{"name":"_orderId","type":"bytes32"},{"name":"_amount","type":"uint256"}],"outputs":[],"stateMutability":"nonpayable"}]`
	return abi.JSON(strings.NewReader(abiJSON))
}

func PrepareUpdateMerchant(ctx *gin.Context) {
	merchantIdParam := ctx.Param("merchantId")
	if merchantIdParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "merchantId is required",
		})
		return
	}

	var input models.MerchantUpdateRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if input.PayoutWalletAddress == nil && input.MetadataURI == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "payoutWalletAdPayoutWalletAddress or metadataURI is required for blockchain update",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	var merchants []models.MerchantDB
	if err := db.Supabase.DB.From("merchants").Select("*").Eq("merchantId", merchantIdParam).Execute(&merchants); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not fetch merchant data: " + err.Error(),
		})
		return
	}

	if len(merchants) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Merchant not found",
		})
		return
	}

	currentMerchant := merchants[0]

	var payoutWalletAdPayoutWalletAddressStr string
	if input.PayoutWalletAddress != nil && *input.PayoutWalletAddress != "" {
		payoutWalletAdPayoutWalletAddressStr = *input.PayoutWalletAddress
	} else {
		payoutWalletAdPayoutWalletAddressStr = currentMerchant.PayoutWalletAddress
	}

	if payoutWalletAdPayoutWalletAddressStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "payoutWalletAdPayoutWalletAddress is required",
		})
		return
	}

	metadataURI := deref(input.MetadataURI)
	if metadataURI == "" {
		metadataURI = currentMerchant.MetadataURI
	}

	if metadataURI == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "metadataURI is required",
		})
		return
	}

	contractABI, err := getMerchantRegistryABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse contract ABI: " + err.Error(),
		})
		return
	}

	merchantIdBytes := common.HexToHash(merchantIdParam)
	payoutAddress := common.HexToAddress(payoutWalletAdPayoutWalletAddressStr)

	callData, err := contractABI.Pack("updateMerchant", merchantIdBytes, payoutAddress, metadataURI)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	contractAddress := networks.BaseSepoliaConfig.MerchantRegistryAddress

	ctx.JSON(http.StatusOK, models.PrepareUpdateResponse{
		TransactionData: models.TransactionData{
			To:       contractAddress.Hex(),
			Data:     "0x" + common.Bytes2Hex(callData),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 200000,
		},
		MerchantId:          merchantIdParam,
		PayoutWalletAddress: payoutWalletAdPayoutWalletAddressStr,
		MetadataURI:         metadataURI,
		Message:             "Sign this transaction with your wallet to update merchant",
	})
}

func ConfirmMerchantUpdate(ctx *gin.Context) {
	merchantIdParam := ctx.Param("merchantId")
	if merchantIdParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "merchantId is required",
		})
		return
	}

	var input models.ConfirmTransactionRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
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

	// Build updates map
	updates := map[string]interface{}{
		"transactionHash": input.TransactionHash,
	}

	if input.MerchantName != "" {
		updates["merchantName"] = input.MerchantName
	}
	if input.PayoutWalletAddress != "" {
		updates["payoutWalletAdPayoutWalletAddress"] = input.PayoutWalletAddress
	}
	if input.MetadataURI != "" {
		updates["metadataURI"] = input.MetadataURI
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.
		From("merchants").
		Update(updates).
		Eq("merchantId", merchantIdParam).
		Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update merchant: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"merchantId":      merchantIdParam,
		"message":         "Merchant updated successfully",
		"transactionHash": input.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + input.TransactionHash,
	})
}

func PrepareRefundOrderMerchant(ctx *gin.Context) {
	orderId := ctx.Param("orderId")

	if orderId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "orderId is required",
		})
		return
	}

	contractABI, err := getPaymentProcessorRegistryABI()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse contract ABI: " + err.Error(),
		})
		return
	}

	orderIdBytes := common.HexToHash(orderId)

	callData, err := contractABI.Pack("refundOrder", orderIdBytes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode transaction data: " + err.Error(),
		})
		return
	}

	contractAddress := networks.BaseSepoliaConfig.PaymentProcessorAddress

	ctx.JSON(http.StatusOK, models.PrepareRefundResponse{
		TransactionData: models.TransactionData{
			To:       contractAddress.Hex(),
			Data:     "0x" + common.Bytes2Hex(callData),
			ChainId:  networks.BaseSepoliaConfig.ChainID.Int64(),
			Value:    "0",
			GasLimit: 150000,
		},
		OrderId: orderId,
		Message: "Sign this transaction with your wallet to refund the order",
	})
}

func ConfirmRefundOrderMerchant(ctx *gin.Context) {
	orderId := ctx.Param("orderId")

	var input models.ConfirmRefundRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
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

	updates := map[string]interface{}{
		"status":          "refunded",
		"transactionHash": input.TransactionHash,
	}

	var result []map[string]interface{}
	if err := db.Supabase.DB.
		From("orders").
		Update(updates).
		Eq("orderId", orderId).
		Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update order status: " + err.Error(),
		})
		return
	}

	response := gin.H{
		"orderId":         orderId,
		"message":         "Order refunded successfully",
		"status":          "refunded",
		"transactionHash": input.TransactionHash,
		"explorerUrl":     networks.BaseSepoliaConfig.ExplorerURL + "/tx/" + input.TransactionHash,
	}

	if len(result) > 0 {
		response["order"] = result[0]
	}

	ctx.JSON(http.StatusOK, response)
}
