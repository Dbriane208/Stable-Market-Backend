package controllers

import (
	"net/http"
	"strconv"

	"github.com/Dbriane208/stable-market/db"
	"github.com/Dbriane208/stable-market/models"
	"github.com/Dbriane208/stable-market/utils"
	"github.com/gin-gonic/gin"
)

func CreateProduct(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(100 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form: " + err.Error(),
		})
		return
	}

	name := ctx.PostForm("name")
	priceStr := ctx.PostForm("price")
	description := ctx.PostForm("description")
	merchantId := ctx.PostForm("merchantId")

	if name == "" || priceStr == "" || merchantId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Name, price and description are required",
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Price must be a valid number",
		})
		return
	}

	if db.Supabase == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database client not initialized",
		})
		return
	}

	uploadResult, err := utils.UploadImageToCloudinary(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to upload image: " + err.Error(),
		})
		return
	}
	if uploadResult == nil {
		return
	}

	dbProduct := models.Products{
		Name:        name,
		Price:       price,
		ImageUrl:    uploadResult.ImageURL,
		Description: description,
		MerchantId:  merchantId,
	}

	var result []models.Products
	if err := db.Supabase.DB.From("products").Insert(dbProduct).Execute(&result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not add product: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Product added successfully",
		"data":    result,
	})
}
