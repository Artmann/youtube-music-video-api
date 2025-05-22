package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSearchHandler_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Mock request with valid parameters
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Euphoria&artists=Loreen",
		},
	}
	c.Request = req
	
	// Note: This test will make a real YouTube request
	// In a production environment, you'd want to mock the YouTube service
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Input.Title != "Euphoria" {
		t.Errorf("Expected title 'Euphoria', got '%s'", response.Input.Title)
	}
	
	if len(response.Input.Artists) != 1 || response.Input.Artists[0] != "Loreen" {
		t.Errorf("Expected artists ['Loreen'], got %v", response.Input.Artists)
	}
}

func TestSearchHandler_EmptyTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=&artists=Loreen",
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}
	
	expectedError := "The title can't be empty."
	if response["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, response["error"])
	}
}

func TestSearchHandler_WhitespaceTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=%20%20%20&artists=Loreen", // URL encoded spaces
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_MissingTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "artists=Loreen",
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_MultipleArtists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Test&artists=Artist1,Artist2,Artist3",
		},
	}
	c.Request = req
	
	// Note: This test will make a real YouTube request
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	expectedArtists := []string{"Artist1", "Artist2", "Artist3"}
	if len(response.Input.Artists) != len(expectedArtists) {
		t.Errorf("Expected %d artists, got %d", len(expectedArtists), len(response.Input.Artists))
	}
	
	for i, expected := range expectedArtists {
		if response.Input.Artists[i] != expected {
			t.Errorf("Expected artist '%s' at index %d, got '%s'", expected, i, response.Input.Artists[i])
		}
	}
}

func TestSearchHandler_ArtistsWithSpaces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Test&artists=%20Artist1%20,%20Artist2%20,%20Artist3%20", // URL encoded spaces
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	expectedArtists := []string{"Artist1", "Artist2", "Artist3"}
	for i, expected := range expectedArtists {
		if response.Input.Artists[i] != expected {
			t.Errorf("Expected trimmed artist '%s' at index %d, got '%s'", expected, i, response.Input.Artists[i])
		}
	}
}

func TestSearchHandler_EmptyArtists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Test&artists=",
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if len(response.Input.Artists) != 0 {
		t.Errorf("Expected empty artists array, got %v", response.Input.Artists)
	}
}

func TestSearchHandler_NoArtistsParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Test",
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if len(response.Input.Artists) != 0 {
		t.Errorf("Expected empty artists array, got %v", response.Input.Artists)
	}
}

func TestSearchHandler_TitleTrimming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=%20%20Test%20Title%20%20", // URL encoded spaces around title
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	expectedTitle := "Test Title"
	if response.Input.Title != expectedTitle {
		t.Errorf("Expected trimmed title '%s', got '%s'", expectedTitle, response.Input.Title)
	}
}

func TestSearchHandler_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "title=Test&artists=Artist",
		},
	}
	c.Request = req
	
	SearchHandler(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	// Check that input field exists and is correct
	if response.Input.Title == "" {
		t.Error("Expected input.title to be present")
	}
	
	// Check that video field exists (can be null if no results)
	// The video field should either be null or contain ID and URL
	if response.Video != nil {
		if response.Video.ID == "" {
			t.Error("Expected video.id to be present when video is not null")
		}
		if response.Video.URL == "" {
			t.Error("Expected video.url to be present when video is not null")
		}
		if !strings.Contains(response.Video.URL, "youtube.com/watch?v=") {
			t.Errorf("Expected YouTube URL format, got %s", response.Video.URL)
		}
	}
}