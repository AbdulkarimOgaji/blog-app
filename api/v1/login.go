package v1

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/abdulkarimogaji/blognado/api/middleware/auth"
	"github.com/abdulkarimogaji/blognado/db"
	"github.com/abdulkarimogaji/blognado/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	User  user   `json:"user"`
	Token string `json:"token"`
}

func login(c *gin.Context) {
	var body loginRequest
	err := c.ShouldBindBodyWith(&body, binding.JSON)

	// request validation
	if err != nil {
		validationResponse(err, c)
		return
	}

	// get user
	var resp loginResponse
	row := db.DbConn.QueryRow("SELECT id, first_name, last_name, password, email, created_at, updated_at from users WHERE email = ?", body.Email)
	err = row.Scan(&resp.User.Id, &resp.User.FirstName, &resp.User.LastName, &resp.User.Password, &resp.User.Email, &resp.User.CreatedAt, &resp.User.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
				"error":   true,
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}

	//check password
	ok := util.VerifyPassword(body.Password, resp.User.Password)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Incorrect email or password",
			"error":   true,
			"data":    nil,
		})
		return
	}

	// create token
	maker, err := auth.NewJwtMaker()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}

	token, err := maker.CreateToken(resp.User.Id, time.Minute*5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "server error",
			"error":   err.Error(),
			"data":    nil,
		})
		return
	}

	resp.Token = token
	resp.User.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
		"error":   nil,
		"data":    resp,
	})
}