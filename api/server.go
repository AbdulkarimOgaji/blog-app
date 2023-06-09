package api

import (
	"log"
	"net/http"
	"time"

	"github.com/abdulkarimogaji/blognado/api/lambda"
	v1 "github.com/abdulkarimogaji/blognado/api/v1"
	"github.com/abdulkarimogaji/blognado/config"
	"github.com/abdulkarimogaji/blognado/db"
	"github.com/abdulkarimogaji/blognado/worker"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
)

type Server interface {
	Start() error
}

type GinServer struct {
	DbService       db.DBService
	Router          *gin.Engine
	TaskDistributor worker.TaskDistributor
}

func NewServer(db db.DBService, taskDistributor worker.TaskDistributor, cloudinaryInstance *cloudinary.Cloudinary) Server {
	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("none", func(fl validator.FieldLevel) bool { return true })
	}

	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Content-Length", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	}))

	lambda.ConfigureRoutes(r.Group("/api/lambda/"), cloudinaryInstance)
	v1.ConfigureRoutes(r.Group("/v1/api/"), db, taskDistributor)

	r.GET("/health", func(c *gin.Context) {
		err := db.PingDB()

		if err != nil {
			log.Println("error here", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to ping the database",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	return &GinServer{Router: r, DbService: db, TaskDistributor: taskDistributor}
}

func (s *GinServer) Start() error {
	return s.Router.Run(":" + config.AppConfig.PORT)
}
