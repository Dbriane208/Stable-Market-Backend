package abi

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// GetMerchantRegistryABI returns the ABI for MerchantRegistry contract
func GetMerchantRegistryABI() (abi.ABI, error) {
	const abiJSON = `[
		{"type":"function","name":"registerMerchant","inputs":[{"name":"_payoutWalletAddress","type":"address"},{"name":"_metadataUri","type":"string"}],"outputs":[{"name":"_merchantId","type":"bytes32"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"updateMerchant","inputs":[{"name":"_merchantId","type":"bytes32"},{"name":"_payoutWalletAddress","type":"address"},{"name":"_metadataUri","type":"string"}],"outputs":[],"stateMutability":"nonpayable"},
		{"type":"event","name":"MerchantRegistered","inputs":[{"name":"merchantId","type":"bytes32","indexed":true},{"name":"owner","type":"address","indexed":true},{"name":"payoutWallet","type":"address","indexed":false},{"name":"metadataUri","type":"string","indexed":false}]}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}
