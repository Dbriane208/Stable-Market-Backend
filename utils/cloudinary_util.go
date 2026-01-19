package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

const (
	MaxFileSize = 5 * 1024 * 1024
)

var cld *cloudinary.Cloudinary

type CloudinaryUploadResult struct {
	ImageURL    string `json:"imageUrl"`
	PublicID    string `json:"publicId"`
	Format      string `json:"format"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	IsDuplicate bool   `json:"isDuplicate"`
}

func InitCloudinary() error {
	url := os.Getenv("CLOUDINARY_URL")

	var err error
	cld, err = cloudinary.NewFromURL(url)

	if err != nil {
		return err
	}

	cld.Config.URL.Secure = true

	return nil
}

func UploadImageToCloudinary(ctx *gin.Context) (*CloudinaryUploadResult, error) {
	if cld == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Cloudinary client not initialized",
		})
		return nil, nil
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return nil, nil
	}
	defer file.Close()

	if header.Size > MaxFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "File size must be less than 5MB",
		})
		return nil, nil
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "File type must be JPEG or PNG",
		})
		return nil, nil
	}

	fileHash, fileData, err := computeFileHash(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to compute file hash: " + err.Error(),
		})
		return nil, nil
	}

	publicId := "StableMarket/products/" + fileHash[:16]

	bgCtx := context.Background()

	// Check if image already exists
	existingImage, err := cld.Admin.Asset(bgCtx, admin.AssetParams{
		PublicID: publicId,
	})

	// If image exists, return the existing URL without re-uploading
	if err == nil && existingImage != nil && existingImage.SecureURL != "" {
		return &CloudinaryUploadResult{
			ImageURL:    existingImage.SecureURL,
			PublicID:    existingImage.PublicID,
			Format:      existingImage.Format,
			Width:       existingImage.Width,
			Height:      existingImage.Height,
			IsDuplicate: true,
		}, nil
	}

	// Image doesn't exist, upload it using the file data bytes
	uploadResult, err := cld.Upload.Upload(bgCtx, bytes.NewReader(fileData), uploader.UploadParams{
		PublicID: publicId,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error uploading file: " + err.Error(),
		})
		return nil, nil
	}

	return &CloudinaryUploadResult{
		ImageURL:    uploadResult.SecureURL,
		PublicID:    uploadResult.PublicID,
		Format:      uploadResult.Format,
		Width:       uploadResult.Width,
		Height:      uploadResult.Height,
		IsDuplicate: false,
	}, nil
}

func computeFileHash(file io.Reader) (string, []byte, error) {
	hash := sha256.New()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", nil, err
	}

	hash.Write(data)
	hashSum := hex.EncodeToString(hash.Sum(nil))

	return hashSum, data, nil
}
