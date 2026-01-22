package main

import (
	"log"
	"os"

	"github.com/Dbriane208/stable-market/db"
	"github.com/Dbriane208/stable-market/networks"
	"github.com/Dbriane208/stable-market/routes"
	"github.com/Dbriane208/stable-market/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize db
	if err := db.DatabaseClient(); err != nil {
		log.Fatal("Failed to initialize database connection: ", err)
	}

	// Initialize both network clients
	_, err := networks.InitClients()
	if err != nil {
		log.Fatal("Failed to initialize clients: ", err)
	}

	if err = utils.InitCloudinary(); err != nil {
		log.Fatal("Failed to initialize cloudinary client: ", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "*"}, // Add your frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupMerchantRoutes(router)
	routes.SetupPlatformRoutes(router)
	routes.SetupOrderRoutes(router)
	routes.SetupMarketRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on :" + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}