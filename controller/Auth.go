package controller

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("secret-codetech") // ganti dengan secret yang aman

func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	var user model.User
	err := config.DB.QueryRow("SELECT id, email, password FROM users WHERE email = @p1", sql.Named("p1", email)).
		Scan(&user.Id, &user.Email, &user.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token expired in 1 hour
	expirationTime := time.Now().Add(1 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	})

	// Gunakan secret key yang sesuai dengan middleware
	tokenString, err := token.SignedString([]byte("secret-codetech"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successfully",
		"token":   tokenString,
	})
}
