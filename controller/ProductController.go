package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// get all data
func GetAllProduct(c *gin.Context) {
	var products []model.Product

	rows, err := config.DB.Query(`
		SELECT id, title, description, price, discount, type, icon, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product model.Product
		if err := rows.Scan(
			&product.Id,
			&product.Title,
			&product.Description,
			&product.Price,
			&product.Discount,
			&product.Type,
			&product.Icon,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse data", "detail": err.Error()})
			return
		}
		products = append(products, product)
	}

	// Cek apakah ada data
	if len(products) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No data found", "data": []model.Product{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// create data
func CreateProduct(c *gin.Context) {
	// Ambil form input
	title := c.PostForm("title")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	discountStr := c.PostForm("discount")
	typeProduct := c.PostForm("type")

	var req model.ProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Konversi price
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
		return
	}

	// Konversi discount (jika diisi)
	var discount float64
	if discountStr != "" {
		discount, err = strconv.ParseFloat(discountStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount format"})
			return
		}
	}

	// Proses file upload (icon opsional)
	var icon string
	file, err := c.FormFile("icon")
	if err == nil {
		if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
			return
		}

		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/" + filename

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		icon = "/uploads/" + filename
	}

	// Simpan ke database PostgreSQL
	query := `
		INSERT INTO products (title, description, price, discount, type, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = config.DB.Exec(query, title, description, price, discount, typeProduct, icon, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to insert data",
			"detail": err.Error(),
		})
		return
	}

	// Response sukses
	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
	})
}

// update data
func UpdateProduct(c *gin.Context) {
	// Ambil ID dari path parameter
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.ProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Ambil form data
	title := c.PostForm("title")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	discountStr := c.PostForm("discount")
	typeProduct := c.PostForm("type")

	// Validasi & parsing angka
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
		return
	}

	var discount float64
	if discountStr != "" {
		discount, err = strconv.ParseFloat(discountStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount format"})
			return
		}
	}

	// Ambil data lama (icon lama)
	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM products WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product", "detail": err.Error()})
		return
	}

	iconPath := oldIcon

	// Jika user upload file baru
	file, err := c.FormFile("icon")
	if err == nil {
		// Buat folder upload jika belum ada
		if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads folder"})
			return
		}

		// Simpan file baru
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		iconPath = "/" + savePath

		// Hapus file lama jika berbeda
		if oldIcon != "" {
			_, imageFile := filepath.Split(oldIcon)
			oldFilePath := filepath.Join("uploads", imageFile)
			if _, err := os.Stat(oldFilePath); err == nil {
				_ = os.Remove(oldFilePath) // Error diabaikan agar update tetap lanjut
			}
		}
	}

	// Update data PostgreSQL
	query := `
		UPDATE products
		SET title = $1, description = $2, price = $3, discount = $4, type = $5, icon = $6, updated_at = $7
		WHERE id = $8
	`
	_, err = config.DB.Exec(query, title, description, price, discount, typeProduct, iconPath, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// delete data
func DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Ambil icon lama untuk hapus file-nya
	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM products WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing product", "detail": err.Error()})
		return
	}

	// Hapus file icon lama jika ada
	if oldIcon != "" {
		_, imageFile := filepath.Split(oldIcon)
		imagePath := filepath.Join("uploads", imageFile)
		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image file", "detail": err.Error()})
				return
			}
		}
	}

	// Hapus data dari PostgreSQL
	query := `DELETE FROM products WHERE id = $1`
	_, err = config.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
