package config

import "time"

const (
	// DefaultUserAgent used for all HTTP requests
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// HTTP Timeouts
	DefaultHTTPTimeout  = 60 * time.Second
	DefaultHeadTimeout  = 10 * time.Second
	DefaultChunkTimeout = 30 * time.Second

	// Concurrency Limits
	MaxChapterWorkers = 20
	MaxImageWorkers   = 100

	// Chunk settings
	MinChunkSize     = 100 * 1024 // 100KB
	DefaultNumChunks = 4

	// Directory settings
	DefaultOutputDir = "output"
)
