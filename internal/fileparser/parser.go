package fileparser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type FileInfo struct {
	Extension     string  `json:"extension"`
	FolderName    string  `json:"folder_name"`
	FileName      string  `json:"file_name"`
	HasYearInPath bool    `json:"has_year_in_path"`
	YearMismatch  bool    `json:"year_mismatch"`
	DetectedYear  *int    `json:"detected_year,omitempty"`
	DetectedRes   *string `json:"detected_resolution,omitempty"`
	IsValidFormat bool    `json:"is_valid_format"`
}

var yearRegex = regexp.MustCompile(`\b(19|20)\d{2}\b`)
var resolutionRegex = regexp.MustCompile(`(?i)\b(2160p|4K|1080p|720p|480p|360p)\b`)

var validExtensions = map[string]bool{
	".mkv":  true,
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".m4v":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
}

// Parse extracts metadata from download path
func Parse(downloadPath string, tmdbYear *int) *FileInfo {
	if downloadPath == "" {
		return nil
	}

	cleanPath := filepath.Clean(downloadPath)
	dir := filepath.Dir(cleanPath)
	
	folderName := ""
	if dir != "." && dir != "/" && dir != string(filepath.Separator) {
		folderName = filepath.Base(dir)
	}
	
	fileName := filepath.Base(cleanPath)
	ext := strings.ToLower(filepath.Ext(cleanPath))
	isValidFormat := validExtensions[ext]

	var detectedYear *int
	var hasYearInPath bool
	var yearMismatch bool

	yearStr := yearRegex.FindString(downloadPath)
	if yearStr != "" {
		var y int
		_, err := fmt.Sscanf(yearStr, "%d", &y)
		if err == nil {
			detectedYear = &y
			hasYearInPath = true
			if tmdbYear != nil && y != *tmdbYear {
				yearMismatch = true
			}
		}
	}

	var detectedRes *string
	resStr := resolutionRegex.FindString(downloadPath)
	if resStr != "" {
		lowerRes := strings.ToLower(resStr)
		detectedRes = &lowerRes
	}

	return &FileInfo{
		Extension:     ext,
		FolderName:    folderName,
		FileName:      fileName,
		HasYearInPath: hasYearInPath,
		YearMismatch:  yearMismatch,
		DetectedYear:  detectedYear,
		DetectedRes:   detectedRes,
		IsValidFormat: isValidFormat,
	}
}
