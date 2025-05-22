package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToSwagger(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}