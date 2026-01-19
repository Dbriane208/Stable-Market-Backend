package models

type PrepareRefundResponse struct {
	TransactionData TransactionData `json:"transactionData"`
	OrderId         string          `json:"orderId"`
	Message         string          `json:"message"`
}

type ApproveTokenRequest struct {
	TokenAddress string `json:"tokenAddress" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
}

type PrepareApproveResponse struct {
	TransactionData TransactionData `json:"transactionData"`
	TokenAddress    string          `json:"tokenAddress"`
	Spender         string          `json:"spender"`
	Amount          string          `json:"amount"`
	Message         string          `json:"message"`
}

type ConfirmApproveRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
}

type ConfirmRefundRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
}

type CreateOrderRequest struct {
	MerchantId   string `json:"merchantId" binding:"required"`
	TokenAddress string `json:"tokenAddress" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	MetadataURI  string `json:"metadataURI" binding:"required"`
}

type PrepareCreateOrderResponse struct {
	TransactionData TransactionData `json:"transactionData"`
	MerchantId      string          `json:"merchantId"`
	TokenAddress    string          `json:"tokenAddress"`
	Amount          string          `json:"amount"`
	MetadataURI     string          `json:"metadataURI"`
	Message         string          `json:"message"`
}

type ConfirmCreateOrderRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
	MerchantId      string `json:"merchantId" binding:"required"`
	TokenAddress    string `json:"tokenAddress" binding:"required"`
	Amount          string `json:"amount" binding:"required"`
	MetadataURI     string `json:"metadataURI" binding:"required"`
	PayerAddress    string `json:"payerAddress" binding:"required"`
}

type PrepareOrder struct {
	OrderId string `json:"orderId" binding:"required"`
}

type PreparePayOrderResponse struct {
	TransactionData TransactionData `json:"transactionData"`
	OrderId         string          `json:"orderId"`
	MerchantId      string          `json:"merchantId"`
	TokenAddress    string          `json:"tokenAddress"`
	Amount          string          `json:"amount"`
	MetadataURI     string          `json:"metadataURI"`
	Message         string          `json:"message"`
}

type ConfirmPayOrderRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
	OrderId         string `json:"orderId" binding:"required"`
	MerchantId      string `json:"merchantId" binding:"required"`
	PayerAddress    string `json:"payerAddress" binding:"required"`
	TokenAddress    string `json:"tokenAddress" binding:"required"`
	Status          string `json:"status" binding:"required"`
	Amount          string `json:"amount" binding:"required"`
}

type OrderDB struct {
	OrderId         string `json:"orderId"`
	MerchantId      string `json:"merchantId"`
	PayerAddress    string `json:"payerAddress"`
	TokenAddress    string `json:"tokenAddress"`
	Amount          string `json:"amount"`
	Status          string `json:"status"`
	MetadataURI     string `json:"metadataURI"`
	TransactionHash string `json:"transactionHash"`
}

type PrepareSettleOrderRequest struct {
	OrderId string `json:"orderId" binding:"required"`
}

type PrepareSettleOrderResponse struct {
	TransactionData TransactionData `json:"transactionData"`
	OrderId         string          `json:"orderId"`
	Message         string          `json:"message"`
}

type ConfirmSettleOrderRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
	OrderId         string `json:"orderId" binding:"required"`
}

type PrepareRefundOrderRequest struct {
	OrderId string `json:"orderId" binding:"required"`
}

type ConfirmRefundOrderRequest struct {
	TransactionHash string `json:"transactionHash" binding:"required"`
	OrderId         string `json:"orderId" binding:"required"`
}

type EmergencyWithdraw struct {
	TokenAddress    string `json:"tokenAddress" binding:"required"`
	RecieverAddress string `json:"receiverAddress" binding:"required"`
	Amount          string `json:"amount" binding:"required"`
	SenderAddress   string `json:"senderAddress"`
	TransactionHash string `json:"transactionHash"`
}

type WithdrawalStatus struct {
	IsWithdrawalEnabled *bool `json:"isWithdrawalEnabled" binding:"required"`
}

type MerchantRegistryUpdate struct {
	NewRegistryAddress string `json:"newRegistryAddress" binding:"required"`
}

type TokenSupport struct {
	TokenAddress string `json:"tokenAddress" binding:"required"`
	StatusValue  string `json:"statusValue" binding:"required"`
}

type PlatformBalanceCheck struct {
	PlatformWallet string `json:"platformWallet" binding:"required"`
	TokenAddress   string `json:"tokenAddress" binding:"required"`
}

type ContractBalanceCheck struct {
	TokenAddress string `json:"tokenAddress" binding:"required"`
}

type UpdateMerchantVerificationStatus struct {
	MerchantId         string `json:"merchantId" binding:"required"`
	VerificationStatus string `json:"verificationStatus" binding:"required"`
}
