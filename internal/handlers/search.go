package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"youtube-music-video-api/internal/services"
)

type SearchInput struct {
	Title   string   `json:"title"`
	Artists []string `json:"artists"`
}

type SearchVideo struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type SearchResponse struct {
	Input SearchInput  `json:"input"`
	Video *SearchVideo `json:"video"`
}

var youtubeService = services.NewYouTubeService()

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
	
	videoIDs, err := youtubeService.SearchVideos(title, artists)
	if err != nil {
		log.Printf("Error searching YouTube for title '%s' with artists %v: %v", title, artists, err)
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search YouTube"})
		return
	}
	
	var video *SearchVideo
	if len(videoIDs) > 0 {
		video = &SearchVideo{
			ID:  videoIDs[0],
			URL: "https://www.youtube.com/watch?v=" + videoIDs[0],
		}
	}
	
	response := SearchResponse{
		Input: SearchInput{
			Title:   title,
			Artists: artists,
		},
		Video: video,
	}
	c.JSON(http.StatusOK, response)
}