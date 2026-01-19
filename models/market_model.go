package models

type Products struct {
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	ImageUrl    string  `json:"imageUrl" binding:"required"`
	Description string  `json:"description" binding:"required"`
	MerchantId  string  `json:"merchantId"`
}

type ProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Description string  `json:"description"`
	MerchantId  string  `json:"merchantId" binding:"required"`
}
