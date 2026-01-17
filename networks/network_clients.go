package networks

import (
	"os"
	"sync"

	"github.com/Dbriane208/stablebase-go-sdk/client"
)

// Clients holds the network clients for both supported networks
type Clients struct {
	BaseClient    *client.Client
	PolygonClient *client.Client
}

// Singleton instance
var (
	instance *Clients
	once     sync.Once
	initErr  error
)

// InitClients initializes and returns clients for both networks (singleton)
func InitClients() (*Clients, error) {
	once.Do(func() {
		// Get your private key
		privateKey := os.Getenv("DEPLOYER_PRIVATE_KEY")

		// Create client for Base Sepolia network
		baseClient, err := client.NewClient(privateKey, BaseSepoliaConfig)
		if err != nil {
			initErr = err
			return
		}

		// Create client for Polygon Amoy network
		polygonClient, err := client.NewClient(privateKey, PolygonAmoyConfig)
		if err != nil {
			initErr = err
			return
		}

		instance = &Clients{
			BaseClient:    baseClient,
			PolygonClient: polygonClient,
		}
	})

	return instance, initErr
}

// GetClients returns the singleton clients instance (must call InitClients first)
func GetClients() *Clients {
	return instance
}

// GetBaseClient returns the Base Sepolia client
func GetBaseClient() *client.Client {
	if instance == nil {
		return nil
	}
	return instance.BaseClient
}

// GetPolygonClient returns the Polygon Amoy client
func GetPolygonClient() *client.Client {
	if instance == nil {
		return nil
	}
	return instance.PolygonClient
}
