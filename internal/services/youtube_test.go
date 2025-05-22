package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestYouTubeService_BuildSearchQuery(t *testing.T) {
	ys := NewYouTubeService()
	
	tests := []struct {
		title    string
		artists  []string
		expected string
	}{
		{"Euphoria", []string{"Loreen"}, "Euphoria Loreen"},
		{"Hello", []string{"Adele"}, "Hello Adele"},
		{"Bohemian Rhapsody", []string{"Queen"}, "Bohemian Rhapsody Queen"},
		{"Shape of You", []string{"Ed", "Sheeran"}, "Shape of You Ed Sheeran"},
		{"Title Only", []string{}, "Title Only"},
		{"Multi Artist", []string{"Artist1", "Artist2", "Artist3"}, "Multi Artist Artist1 Artist2 Artist3"},
	}
	
	for _, test := range tests {
		result := ys.buildSearchQuery(test.title, test.artists)
		if result != test.expected {
			t.Errorf("buildSearchQuery(%q, %v) = %q; want %q", test.title, test.artists, result, test.expected)
		}
	}
}

func TestYouTubeService_BuildCacheKey(t *testing.T) {
	ys := NewYouTubeService()
	
	tests := []struct {
		title    string
		artists  []string
		expected string
	}{
		{"Euphoria", []string{"Loreen"}, "Euphoria Loreen"},
		{"Hello", []string{"Adele"}, "Hello Adele"},
		{"Title Only", []string{}, "Title Only"},
	}
	
	for _, test := range tests {
		result := ys.buildCacheKey(test.title, test.artists)
		if result != test.expected {
			t.Errorf("buildCacheKey(%q, %v) = %q; want %q", test.title, test.artists, result, test.expected)
		}
	}
}

func TestYouTubeService_ExtractVideoIDs(t *testing.T) {
	ys := NewYouTubeService()
	
	tests := []struct {
		name     string
		html     string
		expected []string
		hasError bool
	}{
		{
			name: "Single video ID",
			html: `{"videoId":"dQw4w9WgXcQ"}`,
			expected: []string{"dQw4w9WgXcQ"},
			hasError: false,
		},
		{
			name: "Multiple video IDs",
			html: `{"videoId":"dQw4w9WgXcQ"} some text {"videoId":"oHg5SJYRHA0"}`,
			expected: []string{"dQw4w9WgXcQ", "oHg5SJYRHA0"},
			hasError: false,
		},
		{
			name: "Duplicate video IDs",
			html: `{"videoId":"dQw4w9WgXcQ"} some text {"videoId":"dQw4w9WgXcQ"}`,
			expected: []string{"dQw4w9WgXcQ"},
			hasError: false,
		},
		{
			name: "No video IDs",
			html: `<html>no video ids here</html>`,
			expected: nil,
			hasError: true,
		},
		{
			name: "More than 10 video IDs",
			html: strings.Repeat(`{"videoId":"test123456"} `, 15),
			expected: []string{"test123456"},
			hasError: false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ys.extractVideoIDs(test.html)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d video IDs, got %d", len(test.expected), len(result))
				return
			}
			
			for i, expected := range test.expected {
				if result[i] != expected {
					t.Errorf("Expected video ID %q at index %d, got %q", expected, i, result[i])
				}
			}
		})
	}
}

func TestYouTubeService_CacheIntegration(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"videoId":"cached123"}`))
	}))
	defer server.Close()
	
	ys := NewYouTubeService()
	// Replace the client with one that uses our test server
	ys.client = &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL, "http://")
				return http.DefaultTransport.RoundTrip(req)
			},
		},
	}
	
	title := "Test Title"
	artists := []string{"Test Artist"}
	
	// First call should hit the API
	result1, err := ys.SearchVideos(title, artists)
	if err != nil {
		t.Errorf("Unexpected error on first call: %v", err)
	}
	if len(result1) == 0 || result1[0] != "cached123" {
		t.Errorf("Expected cached123, got %v", result1)
	}
	
	// Second call should hit the cache
	result2, err := ys.SearchVideos(title, artists)
	if err != nil {
		t.Errorf("Unexpected error on second call: %v", err)
	}
	if len(result2) == 0 || result2[0] != "cached123" {
		t.Errorf("Expected cached123 from cache, got %v", result2)
	}
}

// Helper type for mocking HTTP transport
type roundTripperFunc struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f *roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.fn(req)
}

func TestYouTubeService_HTTPError(t *testing.T) {
	ys := NewYouTubeService()
	ys.client = &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
				}, nil
			},
		},
	}
	
	_, err := ys.SearchVideos("Test", []string{"Artist"})
	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}
}

func TestYouTubeService_NetworkError(t *testing.T) {
	ys := NewYouTubeService()
	ys.client = &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				return nil, io.EOF
			},
		},
	}
	
	_, err := ys.SearchVideos("Test", []string{"Artist"})
	if err == nil {
		t.Error("Expected error for network failure")
	}
}