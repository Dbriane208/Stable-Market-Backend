package abi

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// GetERC20ABI returns the ABI for ERC20 token standard
func GetERC20ABI() (abi.ABI, error) {
	const abiJSON = `[
		{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"transferFrom","inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
		{"type":"function","name":"allowance","inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}

// GetPlatformABI returns the combined ABI for platform operations
func GetPlatformABI() (abi.ABI, error) {
	const abiJSON = `[
		{"type":"function","name":"createOrder","inputs":[{"name":"_merchantId","type":"bytes32"},{"name":"_token","type":"address"},{"name":"_amount","type":"uint256"},{"name":"_metadataUri","type":"string"}],"outputs":[{"name":"_orderId","type":"bytes32"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"payOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"refundOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"cancelOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"registerMerchant","inputs":[{"name":"_payoutWalletAddress","type":"address"},{"name":"_metadataUri","type":"string"}],"outputs":[{"name":"_merchantId","type":"bytes32"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"updateMerchant","inputs":[{"name":"_merchantId","type":"bytes32"},{"name":"_payoutWalletAddress","type":"address"},{"name":"_metadataUri","type":"string"}],"outputs":[],"stateMutability":"nonpayable"},
		{"type":"event","name":"OrderCreated","inputs":[{"name":"orderId","type":"bytes32","indexed":true},{"name":"payer","type":"address","indexed":true},{"name":"merchantId","type":"bytes32","indexed":true},{"name":"merchantPayout","type":"address","indexed":false},{"name":"token","type":"address","indexed":false},{"name":"amount","type":"uint256","indexed":false},{"name":"status","type":"uint8","indexed":false},{"name":"metadataUri","type":"string","indexed":false}]},
		{"type":"event","name":"MerchantRegistered","inputs":[{"name":"merchantId","type":"bytes32","indexed":true},{"name":"owner","type":"address","indexed":true},{"name":"payoutWallet","type":"address","indexed":false},{"name":"metadataUri","type":"string","indexed":false}]}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}
