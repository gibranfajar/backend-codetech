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
	"github.com/gibranfajar/backend-codetech/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// get all data
func GetAllUser(c *gin.Context) {
	var users []model.UserResponse

	rows, err := config.DB.Query("SELECT id, name, email, profile, role, created_at, updated_at FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var user model.UserResponse
		if err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Profile, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// get data where is login user with middleware
func GetUser(c *gin.Context) {
	var user model.UserResponse

	id, ok := c.MustGet("user_id").(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user"})
		return
	}

	err := config.DB.QueryRow("SELECT id, name, email, profile, role, created_at, updated_at FROM users WHERE id = @p1", sql.Named("p1", id)).Scan(
		&user.Id, &user.Name, &user.Email, &user.Profile, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// create data
func CreateUser(c *gin.Context) {
	var req model.UserRequest
	// Validasi request body
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi dengan validator
	err := config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Cek apakah email sudah ada
	var user model.User
	err = config.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&user.Id)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data already exists"})
		return
	}

	// Upload profile
	file, err := c.FormFile("profile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile image is required"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Insert data ke PostgreSQL
	_, err = config.DB.Exec(`
	INSERT INTO users (name, email, password, profile, role, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`, name, email, hashedPassword, "/uploads/"+filename, role, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})

}

// update data
func UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.UserRequestUpdate
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")

	// Cek apakah user dengan ID tersebut ada
	var user model.User
	err = config.DB.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&user.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Cek apakah email sudah digunakan oleh user lain
	var existingID int
	err = config.DB.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", email, id).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Ambil gambar lama
	var oldImage string
	err = config.DB.QueryRow("SELECT profile FROM users WHERE id = $1", id).Scan(&oldImage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get old profile"})
		return
	}

	// Upload file baru jika ada
	file, err := c.FormFile("profile")
	var profilePath string
	if err == nil {
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}
		profilePath = "/uploads/" + filename

		// Hapus file lama jika ada
		if oldImage != "" {
			_, oldFile := filepath.Split(oldImage)
			os.Remove("uploads/" + oldFile)
		}
	} else {
		profilePath = oldImage
	}

	// Build query PostgreSQL
	query := `
		UPDATE users
		SET name = $1, email = $2, profile = $3, role = $4, updated_at = $5`
	args := []interface{}{name, email, profilePath, role, time.Now()}

	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		query += `, password = $6 WHERE id = $7`
		args = append(args, hashedPassword, id)
	} else {
		query += ` WHERE id = $6`
		args = append(args, id)
	}

	_, err = config.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to update data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

// delete data
func DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Cek apakah data ada
	var user model.User
	err = config.DB.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&user.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hapus file lama jika ada
	var oldImage string
	err = config.DB.QueryRow("SELECT profile FROM users WHERE id = $1", id).Scan(&oldImage)
	if err == nil && oldImage != "" {
		_, filename := filepath.Split(oldImage)
		os.Remove("uploads/" + filename)
	}

	// Hapus data dari database
	_, err = config.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to delete data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}

// get data where not admin
func GetUserNotAdmin(c *gin.Context) {
	var users []model.UserResponse

	rows, err := config.DB.Query(`
		SELECT id, name, email, profile, role, created_at, updated_at
		FROM users
		WHERE role NOT IN ('admin', 'superadmin')
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch data",
			"detail": err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user model.UserResponse
		if err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Profile, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to scan data",
				"detail": err.Error(),
			})
			return
		}
		users = append(users, user)
	}

	// Cek error setelah iterasi
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Row iteration error",
			"detail": err.Error(),
		})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}
