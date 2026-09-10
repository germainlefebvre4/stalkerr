package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

// DiskSpace represents available disk space information
type DiskSpace struct {
	Available uint64  // Available bytes for unprivileged users
	Free      uint64  // Free bytes on filesystem
	Total     uint64  // Total bytes on filesystem
	UsedPct   float64 // Percentage of space used
}

// nearestExistingAncestor resolves path to an absolute path and walks up its
// parent directories until it finds one that exists, returning that ancestor.
func nearestExistingAncestor(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	checkPath := absPath
	for {
		if _, err := os.Stat(checkPath); err == nil {
			return checkPath, nil
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			// Reached root
			return "", fmt.Errorf("no existing directory found in path")
		}
		checkPath = parent
	}
}

// DeviceID resolves path to its nearest existing ancestor and returns that
// ancestor's device id, so callers can tell whether two configured paths sit
// on the same mounted volume without comparing path strings.
func DeviceID(path string) (uint64, error) {
	checkPath, err := nearestExistingAncestor(path)
	if err != nil {
		return 0, err
	}

	info, err := os.Stat(checkPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat %s: %w", checkPath, err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to read device id for %s", checkPath)
	}

	return stat.Dev, nil
}

// GetDiskSpace returns disk space information for the given path
func GetDiskSpace(path string) (*DiskSpace, error) {
	checkPath, err := nearestExistingAncestor(path)
	if err != nil {
		return nil, err
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(checkPath, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate space in bytes
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	usedPct := float64(used) / float64(total) * 100

	return &DiskSpace{
		Available: available,
		Free:      free,
		Total:     total,
		UsedPct:   usedPct,
	}, nil
}

// HasEnoughSpace checks if there's enough available disk space for the given size
func HasEnoughSpace(path string, requiredBytes uint64) (bool, *DiskSpace, error) {
	space, err := GetDiskSpace(path)
	if err != nil {
		return false, nil, err
	}

	return space.Available >= requiredBytes, space, nil
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// NamedPath is one configured storage path to report disk usage for, labeled
// so a caller can tell which configured setting(s) a merged entry backs.
type NamedPath struct {
	Label string
	Path  string
}

// DiskUsageEntry is the deduplicated disk usage for one mounted volume,
// covering one or more configured NamedPaths that resolved to the same
// device id.
type DiskUsageEntry struct {
	Labels      []string
	Space       *DiskSpace // nil when Unavailable
	Unavailable bool
	Reason      string // set when Unavailable
}

// GroupDiskUsage resolves each NamedPath to its backing device id and returns
// one DiskUsageEntry per distinct device, merging paths that share a volume
// instead of reporting them redundantly. A path that cannot be resolved or
// read is reported as its own unavailable entry rather than failing the
// whole call.
func GroupDiskUsage(paths []NamedPath) []DiskUsageEntry {
	return groupDiskUsage(paths, DeviceID, GetDiskSpace)
}

// groupDiskUsage backs GroupDiskUsage with injectable device-id/disk-space
// lookups, so the grouping/merge logic can be unit tested against synthetic
// device ids instead of depending on the test environment's real mount
// layout to produce two distinct filesystems.
func groupDiskUsage(
	paths []NamedPath,
	deviceIDFn func(string) (uint64, error),
	diskSpaceFn func(string) (*DiskSpace, error),
) []DiskUsageEntry {
	var entries []DiskUsageEntry
	indexByDevice := make(map[uint64]int)

	for _, p := range paths {
		device, err := deviceIDFn(p.Path)
		if err != nil {
			entries = append(entries, DiskUsageEntry{Labels: []string{p.Label}, Unavailable: true, Reason: "unavailable"})
			continue
		}

		if idx, ok := indexByDevice[device]; ok {
			entries[idx].Labels = append(entries[idx].Labels, p.Label)
			continue
		}

		entry := DiskUsageEntry{Labels: []string{p.Label}}
		if space, err := diskSpaceFn(p.Path); err != nil {
			entry.Unavailable = true
			entry.Reason = "unavailable"
		} else {
			entry.Space = space
		}
		indexByDevice[device] = len(entries)
		entries = append(entries, entry)
	}

	return entries
}

// CheckDiskSpaceBeforeDownload validates there's enough space before starting a download
func CheckDiskSpaceBeforeDownload(destPath string, estimatedSize uint64, minFreeSpaceBytes uint64) error {
	space, err := GetDiskSpace(destPath)
	if err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	// Add buffer (10% or minimum free space requirement)
	requiredSpace := estimatedSize
	if minFreeSpaceBytes > 0 && space.Available < minFreeSpaceBytes+requiredSpace {
		return fmt.Errorf(
			"insufficient disk space: available=%s, required=%s (download) + %s (min free) = %s",
			FormatBytes(space.Available),
			FormatBytes(estimatedSize),
			FormatBytes(minFreeSpaceBytes),
			FormatBytes(estimatedSize+minFreeSpaceBytes),
		)
	}

	if space.Available < requiredSpace {
		return fmt.Errorf(
			"insufficient disk space: available=%s, required=%s",
			FormatBytes(space.Available),
			FormatBytes(requiredSpace),
		)
	}

	return nil
}
