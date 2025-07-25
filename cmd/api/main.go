package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"youtube-music-video-api/internal/handlers"
	_ "youtube-music-video-api/docs"
)

// @title YouTube Music Video API
// @version 1.0
// @description API for searching YouTube music videos
// @host localhost:9898
// @BasePath /
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9898"
	}
	
	r := gin.Default()
	
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}))
	
	r.GET("/", handlers.RedirectToSwagger)
	r.GET("/health", handlers.HealthHandler)
	r.GET("/search", handlers.SearchHandler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	r.Run(":" + port)
}