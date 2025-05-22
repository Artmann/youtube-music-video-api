package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockHTTPClient allows us to mock HTTP responses
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}

// HTTPClient interface for dependency injection
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Modified YouTubeService to accept HTTPClient interface
type TestableYouTubeService struct {
	client HTTPClient
	cache  *LRUCache
}

func NewTestableYouTubeService(client HTTPClient) *TestableYouTubeService {
	return &TestableYouTubeService{
		client: client,
		cache:  NewLRUCache(100),
	}
}

func (ys *TestableYouTubeService) SearchVideos(title string, artists []string) ([]string, error) {
	query := ys.buildSearchQuery(title, artists)
	cacheKey := ys.buildCacheKey(title, artists)
	
	// Check cache first
	if videoID, found := ys.cache.Get(cacheKey); found {
		return []string{videoID}, nil
	}
	
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.youtube.com/results?search_query=%s", query), nil)
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
		ys.cache.Put(cacheKey, videoIDs[0])
	}
	
	return videoIDs, nil
}

func (ys *TestableYouTubeService) buildSearchQuery(title string, artists []string) string {
	parts := []string{title}
	parts = append(parts, artists...)
	return strings.Join(parts, " ")
}

func (ys *TestableYouTubeService) buildCacheKey(title string, artists []string) string {
	parts := []string{title}
	parts = append(parts, artists...)
	return strings.Join(parts, " ")
}

func (ys *TestableYouTubeService) extractVideoIDs(html string) ([]string, error) {
	// Reuse the same logic from the original service
	service := NewYouTubeService()
	return service.extractVideoIDs(html)
}

func TestYouTubeService_MockSuccessfulResponse(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock a successful YouTube response
			html := `{"videoId":"dQw4w9WgXcQ"} some other content {"videoId":"oHg5SJYRHA0"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	result, err := service.SearchVideos("Never Gonna Give You Up", []string{"Rick Astley"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(result) == 0 {
		t.Error("Expected at least one video ID")
	}
	
	if result[0] != "dQw4w9WgXcQ" {
		t.Errorf("Expected video ID 'dQw4w9WgXcQ', got '%s'", result[0])
	}
}

func TestYouTubeService_MockHTTPError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Test", []string{"Artist"})
	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}
	
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected error to mention status 500, got: %v", err)
	}
}

func TestYouTubeService_MockNetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network error")
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Test", []string{"Artist"})
	if err == nil {
		t.Error("Expected error for network failure")
	}
	
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("Expected error to mention network error, got: %v", err)
	}
}

func TestYouTubeService_MockEmptyResponse(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock response with no video IDs
			html := `<html><body>No videos found</body></html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Nonexistent Song", []string{"Unknown Artist"})
	if err == nil {
		t.Error("Expected error when no video IDs are found")
	}
	
	if !strings.Contains(err.Error(), "no video IDs found") {
		t.Errorf("Expected error about no video IDs, got: %v", err)
	}
}

func TestYouTubeService_MockMalformedHTML(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock response with malformed JSON
			html := `{"videoId":"incomplete`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Test", []string{"Artist"})
	if err == nil {
		t.Error("Expected error when no valid video IDs are found")
	}
}

func TestYouTubeService_MockCacheBehavior(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			html := `{"videoId":"cached123"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	title := "Test Title"
	artists := []string{"Test Artist"}
	
	// First call should hit the mock client
	result1, err := service.SearchVideos(title, artists)
	if err != nil {
		t.Errorf("Unexpected error on first call: %v", err)
	}
	
	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call, got %d", callCount)
	}
	
	// Second call should hit the cache, not the client
	result2, err := service.SearchVideos(title, artists)
	if err != nil {
		t.Errorf("Unexpected error on second call: %v", err)
	}
	
	if callCount != 1 {
		t.Errorf("Expected still 1 HTTP call (cache hit), got %d", callCount)
	}
	
	// Results should be identical
	if len(result1) != len(result2) || result1[0] != result2[0] {
		t.Errorf("Expected identical results, got %v and %v", result1, result2)
	}
}

func TestYouTubeService_MockUserAgent(t *testing.T) {
	var capturedUserAgent string
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			capturedUserAgent = req.Header.Get("User-Agent")
			html := `{"videoId":"test123"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Test", []string{"Artist"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	expectedUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	if capturedUserAgent != expectedUserAgent {
		t.Errorf("Expected User-Agent '%s', got '%s'", expectedUserAgent, capturedUserAgent)
	}
}

func TestYouTubeService_MockRequestURL(t *testing.T) {
	var capturedURL string
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			capturedURL = req.URL.String()
			html := `{"videoId":"test123"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		},
	}
	
	service := NewTestableYouTubeService(mockClient)
	
	_, err := service.SearchVideos("Hello World", []string{"Test Artist"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(capturedURL, "youtube.com/results") {
		t.Errorf("Expected URL to contain 'youtube.com/results', got '%s'", capturedURL)
	}
	
	if !strings.Contains(capturedURL, "search_query=") {
		t.Errorf("Expected URL to contain 'search_query=', got '%s'", capturedURL)
	}
}