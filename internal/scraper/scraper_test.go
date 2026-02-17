package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractImageURLs_DOM(t *testing.T) {
	html := `
		<html>
		<body>
			<div id="imgs">
				<img data-src="https://example.com/1.jpg" />
				<img src="https://example.com/2.png" />
				<img data-lazy-src="https://example.com/3.webp" />
				<img src="#" /> <!-- Invalid -->
			</div>
		</body>
		</html>
	`
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	urls := ExtractImageURLs(doc)
	expected := []string{
		"https://example.com/1.jpg",
		"https://example.com/2.png",
		"https://example.com/3.webp",
	}

	if len(urls) != len(expected) {
		t.Errorf("Expected %d URLs, got %d", len(expected), len(urls))
	}
}

func TestExtractImageURLs_Script(t *testing.T) {
	html := `
		<html>
		<body>
			<script>
				var ytaw = "some content https://example.com/4.jpg inside script";
			</script>
		</body>
		</html>
	`
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	urls := ExtractImageURLs(doc)
	expected := []string{
		"https://example.com/4.jpg",
	}

	if len(urls) != len(expected) {
		t.Errorf("Expected %d URLs, got %d", len(expected), len(urls))
	}
}

func TestGetImageURL(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "data-src",
			html:     `<img data-src="https://example.com/image.jpg">`,
			expected: "https://example.com/image.jpg",
		},
		{
			name:     "src",
			html:     `<img src="https://example.com/image.png">`,
			expected: "https://example.com/image.png",
		},
		{
			name:     "relative protocol",
			html:     `<img src="//example.com/image.jpg">`,
			expected: "https://example.com/image.jpg",
		},
		{
			name:     "relative path",
			html:     `<img src="/image.jpg">`,
			expected: "https://mangakatana.com/image.jpg",
		},
		{
			name:     "invalid extension",
			html:     `<img src="https://example.com/image.txt">`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			sel := doc.Find("img").First()
			got := getImageURL(sel)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
