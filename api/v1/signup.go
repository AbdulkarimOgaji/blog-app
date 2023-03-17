package v1

import (
	"net/http"
	"time"

	"github.com/abdulkarimogaji/blognado/db"
	"github.com/abdulkarimogaji/blognado/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_ "github.com/go-playground/validator/v10"
)

type signUpRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	City        string `json:"city" binding:"required"`
	Country     string `json:"country"`
	Photo       string `json:"photo" binding:"omitempty,url"`
	DateOfBirth string `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	About       string `json:"about"`
	Settings    string `json:"settings" binding:"omitempty,json"`
	Socials     string `json:"socials" binding:"omitempty,json"`
}

func signUp(c *gin.Context) {
	var body signUpRequest
	err := c.ShouldBindBodyWith(&body, binding.JSON)
	if err != nil {
		validationResponse(err, c)
		return
	}

	hashedPassword, err := util.HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}
	stmt, err := db.DbConn.Prepare("INSERT INTO users (first_name, last_name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?);")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	createdAt := time.Now()
	updatedAt := time.Now()

	result, err := stmt.Exec(body.FirstName, body.LastName, body.Email, hashedPassword, createdAt, updatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	// insert to profile table
	stmt, err = db.DbConn.Prepare("INSERT INTO profile (user_id, date_of_birth, about, photo, city, country, settings, socials, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	result, err = stmt.Exec(id, body.DateOfBirth, body.About, body.Photo, body.City, body.Country, body.Settings, body.Socials, createdAt, updatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	_, err = result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": int(id),
		"error":   false,
	})
}