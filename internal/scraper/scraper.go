package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"mangadl/internal/config"
	"mangadl/internal/domain"

	"github.com/PuerkitoBio/goquery"
)

var (
	httpClient *http.Client
)

func init() {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient = &http.Client{Transport: transport, Timeout: config.DefaultHTTPTimeout}
}

// FetchMangaDetails fetches the title and chapters for a manga URL.
func FetchMangaDetails(mangaURL string) (*domain.MangaDetails, error) {
	doc, err := fetchPage(mangaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	title := doc.Find("h1.heading").Text()
	if title == "" {
		title = doc.Find("title").Text()
	}

	chapters := []domain.Chapter{}
	seenChapters := make(map[string]bool)
	chapterRegex := regexp.MustCompile(`/manga/.*/c\d+`)

	doc.Find(".chapters a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// If generic link regex match fails, skip
		if !chapterRegex.MatchString(href) {
			return
		}

		chapterURL := href
		if !strings.HasPrefix(href, "http") {
			chapterURL = "https://mangakatana.com" + href
		}
		if seenChapters[chapterURL] {
			return
		}
		chapterName := strings.TrimSpace(s.Text())
		if chapterName != "" {
			chapters = append(chapters, domain.Chapter{Name: chapterName, URL: chapterURL})
			seenChapters[chapterURL] = true
		}
	})

	// If no chapters found with specific selector, try generic (fallback)
	if len(chapters) == 0 {
		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || !chapterRegex.MatchString(href) {
				return
			}
			chapterURL := href
			if !strings.HasPrefix(href, "http") {
				chapterURL = "https://mangakatana.com" + href
			}
			if seenChapters[chapterURL] {
				return
			}
			chapterName := strings.TrimSpace(s.Text())
			if chapterName != "" {
				chapters = append(chapters, domain.Chapter{Name: chapterName, URL: chapterURL})
				seenChapters[chapterURL] = true
			}
		})
	}

	// Sort chapters by number (ascending)
	sort.Slice(chapters, func(i, j int) bool {
		return parseChapterNumber(chapters[i].Name) < parseChapterNumber(chapters[j].Name)
	})

	return &domain.MangaDetails{Title: strings.TrimSpace(title), Chapters: chapters}, nil
}

func parseChapterNumber(name string) float64 {
	// Extract number from "Chapter 10.5" or similar
	re := regexp.MustCompile(`(?i)(?:chapter|ch\.?)\s*(\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(name)
	if len(matches) > 1 {
		val, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return val
		}
	}
	// Fallback: try to find any float in the string
	re = regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(name)
	if len(matches) > 1 {
		val, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return val
		}
	}
	return 0
}

// ExtractImageURLs finds all image URLs on a chapter page.
func ExtractImageURLs(doc *goquery.Document) []string {
	imageURLs := []string{}
	seen := make(map[string]bool)

	// Strategy 1: Find direct image tags
	doc.Find("div#imgs img").Each(func(i int, s *goquery.Selection) {
		if url := getImageURL(s); url != "" && !seen[url] {
			imageURLs = append(imageURLs, url)
			seen[url] = true
		}
	})

	// Strategy 2: Find embedded image data in scripts
	if len(imageURLs) == 0 {
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			content := s.Text()
			if strings.Contains(content, "ytaw") || strings.Contains(content, "images") {
				regex := regexp.MustCompile(`https?://[^"']+\.(?:jpg|png|jpeg|webp)`)
				matches := regex.FindAllString(content, -1)
				for _, url := range matches {
					if !seen[url] {
						imageURLs = append(imageURLs, url)
						seen[url] = true
					}
				}
			}
		})
	}
	return imageURLs
}

// fetchPage is a helper to get a goquery document from a URL.
func fetchPage(url string) (*goquery.Document, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", config.DefaultUserAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

// getImageURL extracts the best available source URL from an image element.
func getImageURL(s *goquery.Selection) string {
	url, _ := s.Attr("data-src")
	if url == "" {
		url, _ = s.Attr("src")
	}
	if url == "" {
		url, _ = s.Attr("data-lazy-src")
	}
	if url == "" || url == "#" || strings.HasPrefix(url, "data:image") {
		return ""
	}
	if url != "" && !strings.HasPrefix(url, "http") {
		if strings.HasPrefix(url, "//") {
			url = "https:" + url
		} else {
			url = "https://mangakatana.com" + url
		}
	}
	lower := strings.ToLower(url)
	if !strings.Contains(lower, ".jpg") && !strings.Contains(lower, ".jpeg") &&
		!strings.Contains(lower, ".png") && !strings.Contains(lower, ".webp") {
		return ""
	}
	return url
}
