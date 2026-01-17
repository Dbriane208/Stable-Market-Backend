package abi

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// GetPaymentProcessorABI returns the ABI for PaymentProcessor contract
func GetPaymentProcessorABI() (abi.ABI, error) {
	const abiJSON = `[
		{"type":"function","name":"createOrder","inputs":[{"name":"_merchantId","type":"bytes32"},{"name":"_token","type":"address"},{"name":"_amount","type":"uint256"},{"name":"_metadataUri","type":"string"}],"outputs":[{"name":"_orderId","type":"bytes32"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"payOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"settleOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"refundOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"function","name":"cancelOrder","inputs":[{"name":"_orderId","type":"bytes32"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
		{"type":"event","name":"OrderCreated","inputs":[{"name":"orderId","type":"bytes32","indexed":true},{"name":"payer","type":"address","indexed":true},{"name":"merchantId","type":"bytes32","indexed":true},{"name":"merchantPayout","type":"address","indexed":false},{"name":"token","type":"address","indexed":false},{"name":"amount","type":"uint256","indexed":false},{"name":"status","type":"uint8","indexed":false},{"name":"metadataUri","type":"string","indexed":false}]}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}
