package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type YouTubeService struct {
	client *http.Client
	cache  *LRUCache
}

func NewYouTubeService() *YouTubeService {
	return &YouTubeService{
		client: &http.Client{},
		cache:  NewLRUCache(5000), // Cache up to 5000 search results
	}
}

func (ys *YouTubeService) SearchVideos(title string, artists []string) ([]string, error) {
	query := ys.buildSearchQuery(title, artists)
	cacheKey := ys.buildCacheKey(title, artists)
	
	// Check cache first
	if videoID, found := ys.cache.Get(cacheKey); found {
		log.Printf("Cache HIT for key: %s", cacheKey)
		return []string{videoID}, nil
	}
	log.Printf("Cache MISS for key: %s", cacheKey)
	
	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", url.QueryEscape(query))
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	resp, err := ys.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube search results: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch YouTube search results: status %d", resp.StatusCode)
	}
	
	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	videoIDs, err := ys.extractVideoIDs(string(html))
	if err != nil {
		return nil, fmt.Errorf("failed to extract video IDs: %w", err)
	}
	
	// Cache the first video ID if found
	if len(videoIDs) > 0 {
		log.Printf("Caching video ID %s for key: %s", videoIDs[0], cacheKey)
		ys.cache.Put(cacheKey, videoIDs[0])
	}
	
	return videoIDs, nil
}

func (ys *YouTubeService) buildSearchQuery(title string, artists []string) string {
	parts := []string{title}
	parts = append(parts, artists...)
	return strings.Join(parts, " ")
}

func (ys *YouTubeService) buildCacheKey(title string, artists []string) string {
	parts := []string{title}
	parts = append(parts, artists...)
	return strings.Join(parts, " ")
}

func (ys *YouTubeService) extractVideoIDs(html string) ([]string, error) {
	pattern := `"videoId":"([^"]+)"`
	re := regexp.MustCompile(pattern)
	
	matches := re.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no video IDs found in search results")
	}
	
	var videoIDs []string
	seen := make(map[string]bool)
	
	// Return up to 10 unique video IDs
	for _, match := range matches {
		if len(videoIDs) >= 10 {
			break
		}
		
		videoID := match[1]
		if videoID != "" && !seen[videoID] {
			videoIDs = append(videoIDs, videoID)
			seen[videoID] = true
		}
	}
	
	if len(videoIDs) == 0 {
		return nil, fmt.Errorf("failed to extract any valid video IDs")
	}
	
	return videoIDs, nil
}