package domain

// Chapter represents a manga chapter.
type Chapter struct {
	Name string
	URL  string
	ID   int // preserve original index
}

// Title returns the chapter name for the list interface.
func (c Chapter) Title() string { return c.Name }

// Description returns the chapter URL for the list interface.
func (c Chapter) Description() string { return c.URL }

// FilterValue returns the chapter name for filtering.
func (c Chapter) FilterValue() string { return c.Name }

// MangaDetails contains information about a manga.
type MangaDetails struct {
	Title    string
	Chapters []Chapter
	CoverURL string
}

// ChunkDownload represents a downloaded chunk of a file.
type ChunkDownload struct {
	Data  []byte
	Start int64
	End   int64
}

// DownloadStatus represents the current state of a download.
type DownloadStatus int

const (
	StatusPending DownloadStatus = iota
	StatusDownloading
	StatusCompleted
	StatusFailed
)
