package main

import (
	"fmt"
)

// formatBytes converts a byte count to a human-readable string (e.g. "1.23 MB").
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// valueOrEmpty returns the dereferenced string or an empty string if the pointer is nil.
func valueOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
