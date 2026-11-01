package downloader

import (
	"path/filepath"
	"strings"
)

// mediaExtensions are the file extensions a recorded DownloadInfo.DownloadPath
// may legitimately end with once the transfer has produced a file.
var mediaExtensions = map[string]bool{
	".mkv": true, ".mp4": true, ".avi": true, ".mov": true, ".m4v": true,
	".ts": true, ".wmv": true, ".flv": true, ".webm": true, ".mpg": true, ".mpeg": true,
}

// NormalizeRecordedDestPath turns a recorded DownloadInfo.DownloadPath into the
// extension-less base destination path expected by DownloadOptions.BaseDestPath.
// The extension is stripped only when it is a known media extension, so a title
// containing a dot (e.g. "Mr. Robot (2015) [1080p]") is left intact.
func NormalizeRecordedDestPath(path string) string {
	if ext := filepath.Ext(path); mediaExtensions[strings.ToLower(ext)] {
		return strings.TrimSuffix(path, ext)
	}
	return path
}
