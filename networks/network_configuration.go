package networks

import (
	"math/big"

	"github.com/Dbriane208/stablebase-go-sdk/client"
	"github.com/ethereum/go-ethereum/common"
)

var BaseSepoliaConfig = client.NetworkConfig{
	NetworkName:             "base-sepolia",
	ChainID:                 big.NewInt(84532),
	RPCURL:                  "https://sepolia.base.org",
	USDCAddress:             common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	PaymentProcessorAddress: common.HexToAddress("0x7c39408AC96a1b9a2722056eDE90b54D2B260380"),
	MerchantRegistryAddress: common.HexToAddress("0x93e93Dfa36C87De32B9118CA5D9BAd1Db892002d"),
	ExplorerURL:             "https://sepolia.basescan.org",
}

var PolygonAmoyConfig = client.NetworkConfig{
	NetworkName:             "polygon-amoy",
	ChainID:                 big.NewInt(80002),
	RPCURL:                  "https://rpc-amoy.polygon.technology",
	USDCAddress:             common.HexToAddress("0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"),
	PaymentProcessorAddress: common.HexToAddress("0x3B08Be115E1672cE8A6618D932a97B2Cc251d853"),
	MerchantRegistryAddress: common.HexToAddress("0xE664919f8a195d44c8a137C71cBeb967A71eD3DF"),
	ExplorerURL:             "https://amoy.polygonscan.com",
}

// NetworkConfigs provides a map of all supported network configurations
var NetworkConfigs = map[string]client.NetworkConfig{
	"base-sepolia": BaseSepoliaConfig,
	"polygon-amoy": PolygonAmoyConfig,
}

// GetNetworkConfig returns the network configuration for the given network name
func GetNetworkConfig(networkName string) (client.NetworkConfig, bool) {
	config, exists := NetworkConfigs[networkName]
	return config, exists
}

// GetAllNetworkConfigs returns all supported network configurations
func GetAllNetworkConfigs() []client.NetworkConfig {
	return []client.NetworkConfig{BaseSepoliaConfig, PolygonAmoyConfig}
}
