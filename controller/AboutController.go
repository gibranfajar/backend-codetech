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

// getAllDate
func GetAllAbout(c *gin.Context) {
	var about model.About

	err := config.DB.QueryRow("SELECT id, title, description, image, created_at, updated_at FROM abouts").Scan(
		&about.Id, &about.Title, &about.Description, &about.Image, &about.CreatedAt, &about.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"message": "No data found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": about,
	})
}

// create data
func CreateAbout(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")

	var req model.AboutRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err := config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// check apakah sudah ada data di database atau belum, jika sudah maka tidak bisa menambahkan data lagi
	var about model.About
	err = config.DB.QueryRow("SELECT id FROM abouts").Scan(&about.Id)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data already exists"})
		return
	}

	_, err = config.DB.Exec(`
		INSERT INTO abouts (title, description, image, created_at, updated_at)
		VALUES (@p1, @p2, @p3, @p4, @p5)
	`, title, description, "/uploads/"+filename, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update
func UpdateAbout(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.AboutRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")

	// Ambil data lama untuk dapatkan image lama
	var oldImage string
	err = config.DB.QueryRow("SELECT image FROM abouts WHERE id = @p1", sql.Named("p1", id)).Scan(&oldImage)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "About not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing about", "detail": err.Error()})
		return
	}

	imagePath := oldImage // default: gunakan image lama

	file, err := c.FormFile("image")
	if err == nil {
		// Jika ada file baru, upload dan ganti
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// Hapus file lama jika ada
		if oldImage != "" {
			_, oldFile := filepath.Split(oldImage)
			oldFilePath := filepath.Join("uploads", oldFile)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}

		imagePath = "/uploads/" + filename
	}

	_, err = config.DB.Exec(`
		UPDATE abouts
		SET title = @p1, description = @p2, image = @p3, updated_at = @p4
		WHERE id = @p5
	`, title, description, imagePath, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete
func DeleteAbout(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var about model.About
	// Langsung ambil id dan image dalam satu query
	err = config.DB.QueryRow("SELECT id, image FROM abouts WHERE id = @p1", sql.Named("p1", id)).Scan(&about.Id, &about.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		}
		return
	}

	// Hapus file gambar jika ada
	if about.Image != "" {
		_, imageFile := filepath.Split(about.Image)
		imagePath := filepath.Join("uploads", imageFile)
		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image", "detail": err.Error()})
				return
			}
		}
	}

	// Hapus data dari database
	_, err = config.DB.Exec("DELETE FROM abouts WHERE id = @p1", sql.Named("p1", id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
