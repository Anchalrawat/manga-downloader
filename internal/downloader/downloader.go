package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"go-manga-ripper/internal/config"
	"go-manga-ripper/internal/domain"
	"go-manga-ripper/internal/scraper"

	"github.com/PuerkitoBio/goquery"
	"github.com/valyala/fasthttp"
)

var (
	fasthttpClient *fasthttp.Client
	imageSemaphore = make(chan struct{}, config.MaxImageWorkers)
)

func init() {
	fasthttpClient = &fasthttp.Client{MaxConnsPerHost: 1000}
}

// DownloadChapter handles the full download process for a single chapter.
func DownloadChapter(chapterURL, chapterName, mangaDir string) error {
	safeName := SanitizeFilename(chapterName)
	outputDir := filepath.Join(config.DefaultOutputDir, mangaDir, safeName)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	doc, err := fetchPage(chapterURL)
	if err != nil {
		return err
	}

	imageURLs := scraper.ExtractImageURLs(doc)
	if len(imageURLs) == 0 {
		return fmt.Errorf("no images found")
	}

	if err := downloadImagesChunked(imageURLs, outputDir); err != nil {
		return err
	}

	zipName := filepath.Join(config.DefaultOutputDir, mangaDir, safeName+".cbz")
	return createCBZ(outputDir, zipName)
}

// fetchPage is a local helper, duplicated from scraper to avoid circular dependency if needed,
// or better yet, just use http client here as well.
// For simplicity, let's just reuse the logic but with standard http client for the page itself.
func fetchPage(url string) (*goquery.Document, error) {
	client := &http.Client{Timeout: config.DefaultHTTPTimeout}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", config.DefaultUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

func downloadImagesChunked(imageURLs []string, outputDir string) error {
	var wg sync.WaitGroup
	for i, url := range imageURLs {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			imageSemaphore <- struct{}{}
			defer func() { <-imageSemaphore }()
			DownloadImageInChunks(u, outputDir, idx+1)
		}(i, url)
	}
	wg.Wait()
	return nil
}

// DownloadImageInChunks downloads a single image, splitting it into chunks if supported.
func DownloadImageInChunks(url, outputDir string, index int) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("HEAD")
	req.Header.Set("User-Agent", config.DefaultUserAgent)
	req.Header.Set("Referer", "https://mangakatana.com/")

	if err := fasthttpClient.DoTimeout(req, resp, config.DefaultHeadTimeout); err != nil {
		return downloadImageFast(url, outputDir, index)
	}

	acceptRanges := string(resp.Header.Peek("Accept-Ranges"))
	contentLength := string(resp.Header.Peek("Content-Length"))

	if acceptRanges != "bytes" || contentLength == "" {
		return downloadImageFast(url, outputDir, index)
	}

	fileSize, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil || fileSize == 0 || fileSize < int64(config.MinChunkSize) {
		return downloadImageFast(url, outputDir, index)
	}

	numChunks := config.DefaultNumChunks
	chunkSize := fileSize / int64(numChunks)
	var wg sync.WaitGroup
	var chunkMux sync.Mutex
	var chunks []domain.ChunkDownload
	var downloadErr error

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkIdx int) {
			defer wg.Done()
			start := int64(chunkIdx) * chunkSize
			end := start + chunkSize - 1
			if chunkIdx == numChunks-1 {
				end = fileSize - 1
			}
			data, err := downloadChunk(url, start, end)
			if err != nil {
				chunkMux.Lock()
				if downloadErr == nil {
					downloadErr = err
				}
				chunkMux.Unlock()
				return
			}
			chunkMux.Lock()
			chunks = append(chunks, domain.ChunkDownload{Data: data, Start: start, End: end})
			chunkMux.Unlock()
		}(i)
	}
	wg.Wait()

	if downloadErr != nil {
		return downloadImageFast(url, outputDir, index)
	}

	// Sort chunks
	sortedChunks := make([]domain.ChunkDownload, len(chunks))
	for _, chunk := range chunks {
		idx := int(chunk.Start / chunkSize)
		if idx >= len(sortedChunks) {
			idx = len(sortedChunks) - 1
		}
		sortedChunks[idx] = chunk
	}

	filename := filepath.Join(outputDir, fmt.Sprintf("%03d.jpg", index))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, chunk := range sortedChunks {
		if _, err := file.Write(chunk.Data); err != nil {
			return err
		}
	}
	return nil
}

func downloadChunk(url string, start, end int64) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("User-Agent", config.DefaultUserAgent)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	if err := fasthttpClient.DoTimeout(req, resp, config.DefaultChunkTimeout); err != nil {
		return nil, err
	}
	if resp.StatusCode() != fasthttp.StatusPartialContent && resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode())
	}
	data := make([]byte, len(resp.Body()))
	copy(data, resp.Body())
	return data, nil
}

func downloadImageFast(url, outputDir string, index int) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	if err := fasthttpClient.DoTimeout(req, resp, config.DefaultChunkTimeout); err != nil {
		return err
	}
	filename := filepath.Join(outputDir, fmt.Sprintf("%03d.jpg", index))
	return os.WriteFile(filename, resp.Body(), 0644)
}

func createCBZ(src, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		header, _ := zip.FileInfoHeader(info)
		header.Method = zip.Deflate
		header.Name, _ = filepath.Rel(src, path)
		writer, _ := w.CreateHeader(header)
		file, _ := os.Open(path)
		defer file.Close()
		io.Copy(writer, file)
		return nil
	})
}

func SanitizeFilename(name string) string {
	return regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(name, "_")
}
