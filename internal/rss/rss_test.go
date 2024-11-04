package rss

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test RSS Feed parsing
func TestCheckRSSFeed(t *testing.T) {
	// Mock RSS feed XML
	rssFeedXML := `
		<rss>
			<channel>
				<title>Test Blog</title>
				<item>
					<title>Test Post</title>
					<link>https://example.com/test-post</link>
					<description>This is a test post</description>
				</item>
			</channel>
		</rss>`

	server := mockHTTPServer(rssFeedXML, 200)
	defer server.Close()

	posts, err := CheckRSSFeed(server.URL)
	if err != nil {
		t.Fatalf("Failed to fetch RSS feed: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].Title != "Test Post" {
		t.Errorf("Expected post title 'Test Post', got '%s'", posts[0].Title)
	}
}

// Test hash content function
func TestHashContent(t *testing.T) {
	content := "This is a test post"
	actualHash := HashContent(content)

	expectedHash := [32]byte{171, 214, 38, 231, 215, 166, 144, 206, 157, 133, 112, 100, 123, 136, 149, 247, 102, 45, 79, 114, 7, 254, 136, 203, 103, 200, 223, 156, 18, 75, 167, 165}

	if !bytes.Equal(actualHash[:], expectedHash[:]) {
		t.Errorf("Expected hash '%p', got '%p'", expectedHash[:], actualHash[:])
	}
}

// Helper function to mock an HTTP server
func mockHTTPServer(response string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(response))

	}))
}
