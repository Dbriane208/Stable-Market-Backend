package models

import (
	"github.com/ethereum/go-ethereum/common"
)

type MerchantResponse struct {
	MerchantId         common.Hash    `json:"merchantId"`
	Message            string         `json:"message"`
	MetadataURI        string         `json:"metadataURI"`
	WalletAddress      common.Address `json:"walletAddress"`
	MerchantName       string         `json:"merchantName"`
	TransactionHash    common.Hash    `json:"transactionHash"`
	ExplorerURL        string         `json:"explorerUrl"`
	IsMerchantVerified bool           `json:"isMerchantVerified"`
}

type MerchantInfo struct {
	MerchantName        string         `json:"merchantName" binding:"required"`
	MerchantId          common.Hash    `json:"merchantId"`
	PayoutWalletAddress common.Address `json:"payoutWalletAddress" binding:"required"`
	MetadataURI         string         `json:"metadataURI" binding:"required"`
	IsMerchantVerified  bool           `json:"isMerchantVerified"`
	TransactionHash     common.Hash    `json:"transactionHash"`
}

type MerchantDB struct {
	MerchantName        string `json:"merchantName"`
	MerchantId          string `json:"merchantId"`
	PayoutWalletAddress string `json:"payoutWalletAddress"`
	MetadataURI         string `json:"metadataURI"`
	TransactionHash     string `json:"transactionHash"`
}

type MerchantUpdateRequest struct {
	MerchantName        *string `json:"merchantName,omitempty"`
	PayoutWalletAddress *string `json:"payoutWalletAddress,omitempty"`
	MetadataURI         *string `json:"metadataURI,omitempty"`
}

type TokenBalance struct {
	WalletAddress common.Address `json:"walletAddress" binding:"required"`
	TokenAddress  common.Address `json:"tokenAddress" binding:"required"`
}

type TokenBalanceDB struct {
	MerchantId    string `json:"merchantId"`
	WalletAddress string `json:"walletAddress"`
	TokenAddress  string `json:"tokenAddress"`
	TokenBalance  string `json:"tokenBalance"`
}

type TransactionData struct {
	To       string `json:"to"`
	Data     string `json:"data"`
	ChainId  int64  `json:"chainId"`
	Value    string `json:"value"`
	GasLimit uint64 `json:"gasLimit"`
}

type PrepareUpdateResponse struct {
	TransactionData     TransactionData `json:"transactionData"`
	MerchantId          string          `json:"merchantId"`
	PayoutWalletAddress string          `json:"payoutWalletAddress"`
	MetadataURI         string          `json:"metadataURI"`
	Message             string          `json:"message"`
}

type ConfirmTransactionRequest struct {
	TransactionHash     string `json:"transactionHash" binding:"required"`
	MerchantName        string `json:"merchantName,omitempty"`
	PayoutWalletAddress string `json:"payoutWalletAddress,omitempty"`
	MetadataURI         string `json:"metadataURI,omitempty"`
}
