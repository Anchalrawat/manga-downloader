package ui

import "mangadl/internal/domain"

type MangaFetchedMsg *domain.MangaDetails
type ErrMsg error
type ProgressMsg struct {
	Done    int // -1 for increment
	Total   int
	Message string
}
type DownloadCompleteMsg struct{}
