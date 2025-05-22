package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"youtube-music-video-api/internal/handlers"
)

func TestAPIIntegration_CompleteSearchFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handlers.SearchHandler)
	
	// Test successful search
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?title=Euphoria&artists=Loreen", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response handlers.SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response structure
	if response.Input.Title != "Euphoria" {
		t.Errorf("Expected title 'Euphoria', got '%s'", response.Input.Title)
	}
	
	if len(response.Input.Artists) != 1 || response.Input.Artists[0] != "Loreen" {
		t.Errorf("Expected artists ['Loreen'], got %v", response.Input.Artists)
	}
	
	// Should have found a video
	if response.Video == nil {
		t.Error("Expected to find a video")
	} else {
		if response.Video.ID == "" {
			t.Error("Expected video ID to be present")
		}
		if response.Video.URL == "" {
			t.Error("Expected video URL to be present")
		}
		expectedURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", response.Video.ID)
		if response.Video.URL != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, response.Video.URL)
		}
	}
}

func TestAPIIntegration_CachingBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handlers.SearchHandler)
	
	searchQuery := "title=Bohemian%20Rhapsody&artists=Queen"
	
	// First request - should hit YouTube API
	start1 := time.Now()
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/search?"+searchQuery, nil)
	router.ServeHTTP(w1, req1)
	duration1 := time.Since(start1)
	
	if w1.Code != http.StatusOK {
		t.Errorf("First request failed with status %d", w1.Code)
	}
	
	var response1 handlers.SearchResponse
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	if err != nil {
		t.Errorf("Failed to unmarshal first response: %v", err)
	}
	
	// Second request - should hit cache
	start2 := time.Now()
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/search?"+searchQuery, nil)
	router.ServeHTTP(w2, req2)
	duration2 := time.Since(start2)
	
	if w2.Code != http.StatusOK {
		t.Errorf("Second request failed with status %d", w2.Code)
	}
	
	var response2 handlers.SearchResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	if err != nil {
		t.Errorf("Failed to unmarshal second response: %v", err)
	}
	
	// Both responses should be identical
	if response1.Video == nil || response2.Video == nil {
		t.Error("Expected both responses to have videos")
	} else {
		if response1.Video.ID != response2.Video.ID {
			t.Errorf("Expected same video ID, got '%s' and '%s'", response1.Video.ID, response2.Video.ID)
		}
	}
	
	// Second request should be significantly faster (cache hit)
	if duration2 > duration1/2 {
		t.Logf("Warning: Second request (%v) wasn't significantly faster than first (%v)", duration2, duration1)
		// Note: This might not always be reliable due to network variance
	}
}

func TestAPIIntegration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handlers.SearchHandler)
	
	const numRequests = 3 // Reduced to avoid rate limiting
	var wg sync.WaitGroup
	results := make([]handlers.SearchResponse, numRequests)
	errors := make([]error, numRequests)
	
	// Make multiple concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/search?title=Shape%20of%20You&artists=Ed%20Sheeran", nil)
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors[index] = fmt.Errorf("request %d failed with status %d", index, w.Code)
				return
			}
			
			var response handlers.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				errors[index] = fmt.Errorf("failed to unmarshal response %d: %v", index, err)
				return
			}
			
			results[index] = response
		}(i)
	}
	
	wg.Wait()
	
	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
	
	// All successful responses should have the same video (due to caching)
	var firstVideoID string
	for i, result := range results {
		if errors[i] != nil {
			continue // Skip failed requests
		}
		
		if result.Video == nil {
			t.Errorf("Request %d returned no video", i)
			continue
		}
		
		if firstVideoID == "" {
			firstVideoID = result.Video.ID
		} else if result.Video.ID != firstVideoID {
			t.Errorf("Request %d returned different video ID: expected %s, got %s", i, firstVideoID, result.Video.ID)
		}
	}
}

func TestAPIIntegration_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handlers.SearchHandler)
	
	// Test missing title
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?artists=Test", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing title, got %d", http.StatusBadRequest, w.Code)
	}
	
	var errorResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}
	
	expectedError := "The title can't be empty."
	if errorResponse["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse["error"])
	}
}

func TestAPIIntegration_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handlers.HealthHandler)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestAPIIntegration_ParameterEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handlers.SearchHandler)
	
	testCases := []struct {
		name        string
		query       string
		expectCode  int
		expectError bool
	}{
		{
			name:        "Only whitespace title",
			query:       "title=%20%20%20&artists=Artist",
			expectCode:  http.StatusBadRequest,
			expectError: true,
		},
		// Note: Reduced test cases to avoid rate limiting
		// In production, you'd want to mock these or use a dedicated test environment
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/search?"+tc.query, nil)
			router.ServeHTTP(w, req)
			
			if w.Code != tc.expectCode {
				t.Errorf("Expected status %d, got %d", tc.expectCode, w.Code)
			}
			
			if tc.expectError {
				var errorResponse map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if errorResponse["error"] == "" {
					t.Error("Expected error field to be present")
				}
			} else {
				var response handlers.SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal success response: %v", err)
				}
			}
		})
	}
}