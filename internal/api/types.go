package api

import (
	"time"

	"github.com/glefebvre/stalkeer/internal/fileparser"
)

type DownloadEnrichedResponse struct {
	ID              uint                 `json:"id"`
	URL             string               `json:"url"`
	Status          string               `json:"status"`
	DownloadPath    *string              `json:"download_path,omitempty"`
	FileSize        *int64               `json:"file_size,omitempty"`
	BytesDownloaded *int64               `json:"bytes_downloaded,omitempty"`
	TotalBytes      *int64               `json:"total_bytes,omitempty"`
	RetryCount      int                  `json:"retry_count"`
	ErrorMessage    *string              `json:"error_message,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at"`
	Content         *ContentInfo         `json:"content,omitempty"`
	FileInfo        *fileparser.FileInfo `json:"file_info,omitempty"`
}

type ContentInfo struct {
	Type       string  `json:"type"`
	Title      string  `json:"title"`
	Year       *int    `json:"year,omitempty"`
	Resolution *string `json:"resolution,omitempty"`
	Season     *int    `json:"season,omitempty"`
	Episode    *int    `json:"episode,omitempty"`
	Genres     *string `json:"genres,omitempty"`
	Duration   *int    `json:"duration,omitempty"`
}
