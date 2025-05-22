package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchInput struct {
	Title   string   `json:"title"`
	Artists []string `json:"artists"`
}

type SearchVideo struct {
	ID string `json:"id"`
}

type SearchResponse struct {
	Input  SearchInput   `json:"input"`
	Videos []SearchVideo `json:"videos"`
}

// SearchHandler godoc
// @Summary Search for music videos
// @Description Returns a list of music video search results
// @Tags search
// @Produce json
// @Param title query string true "Title to search for"
// @Param artists query string false "Artist name or comma-separated list of artists"
// @Success 200 {object} SearchResponse
// @Failure 400 {object} map[string]string
// @Router /search [get]
func SearchHandler(c *gin.Context) {
	title := strings.TrimSpace(c.Query("title"))
	artistsParam := strings.TrimSpace(c.Query("artists"))
	
	if title == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "The title can't be empty."})
		return
	}
	
	var artists []string
	if artistsParam != "" {
		artistsList := strings.Split(artistsParam, ",")
		for _, artist := range artistsList {
			trimmed := strings.TrimSpace(artist)
			if trimmed != "" {
				artists = append(artists, trimmed)
			}
		}
	}
	
	response := SearchResponse{
		Input: SearchInput{
			Title:   title,
			Artists: artists,
		},
		Videos: []SearchVideo{
			{ID: "123"},
		},
	}
	c.JSON(http.StatusOK, response)
}