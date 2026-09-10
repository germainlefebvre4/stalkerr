package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDiskSpace(t *testing.T) {
	// Use current directory for testing
	wd, err := os.Getwd()
	require.NoError(t, err)

	space, err := GetDiskSpace(wd)
	require.NoError(t, err)
	assert.NotNil(t, space)

	// Basic sanity checks
	assert.Greater(t, space.Total, uint64(0))
	assert.Greater(t, space.Available, uint64(0))
	assert.LessOrEqual(t, space.Available, space.Total)
	assert.GreaterOrEqual(t, space.UsedPct, 0.0)
	assert.LessOrEqual(t, space.UsedPct, 100.0)
}

func TestGetDiskSpace_NonExistentPath(t *testing.T) {
	// Test with non-existent path - should check parent directory
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "does", "not", "exist", "yet")

	space, err := GetDiskSpace(nonExistent)
	require.NoError(t, err)
	assert.NotNil(t, space)
	assert.Greater(t, space.Available, uint64(0))
}

func TestGetDiskSpace_TempDir(t *testing.T) {
	tempDir := t.TempDir()

	space, err := GetDiskSpace(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, space)
	assert.Greater(t, space.Available, uint64(0))
}

func TestHasEnoughSpace(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		requiredBytes uint64
		expectEnough  bool
	}{
		{
			name:          "1 KB should have space",
			requiredBytes: 1024,
			expectEnough:  true,
		},
		{
			name:          "1 MB should have space",
			requiredBytes: 1024 * 1024,
			expectEnough:  true,
		},
		{
			name:          "extremely large should not have space",
			requiredBytes: 1024 * 1024 * 1024 * 1024 * 1024, // 1 PB
			expectEnough:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSpace, space, err := HasEnoughSpace(tempDir, tt.requiredBytes)
			require.NoError(t, err)
			assert.NotNil(t, space)
			assert.Equal(t, tt.expectEnough, hasSpace)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1536 * 1024 * 1024, "1.5 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckDiskSpaceBeforeDownload(t *testing.T) {
	tempDir := t.TempDir()

	// Get actual available space
	space, err := GetDiskSpace(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name              string
		estimatedSize     uint64
		minFreeSpaceBytes uint64
		expectError       bool
	}{
		{
			name:              "small download with no minimum",
			estimatedSize:     1024 * 1024, // 1 MB
			minFreeSpaceBytes: 0,
			expectError:       false,
		},
		{
			name:              "small download with reasonable minimum",
			estimatedSize:     1024 * 1024,       // 1 MB
			minFreeSpaceBytes: 100 * 1024 * 1024, // 100 MB
			expectError:       space.Available < (101 * 1024 * 1024),
		},
		{
			name:              "unreasonably large download",
			estimatedSize:     1024 * 1024 * 1024 * 1024 * 1024, // 1 PB
			minFreeSpaceBytes: 0,
			expectError:       true,
		},
		{
			name:              "unreasonably large minimum free space",
			estimatedSize:     1024,                             // 1 KB
			minFreeSpaceBytes: 1024 * 1024 * 1024 * 1024 * 1024, // 1 PB
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDiskSpaceBeforeDownload(tempDir, tt.estimatedSize, tt.minFreeSpaceBytes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceID_SameFilesystem(t *testing.T) {
	tempDir := t.TempDir()
	subDirA := filepath.Join(tempDir, "a")
	subDirB := filepath.Join(tempDir, "b")
	require.NoError(t, os.MkdirAll(subDirA, 0o755))
	require.NoError(t, os.MkdirAll(subDirB, 0o755))

	deviceA, err := DeviceID(subDirA)
	require.NoError(t, err)
	deviceB, err := DeviceID(subDirB)
	require.NoError(t, err)

	assert.Equal(t, deviceA, deviceB)
}

func TestDeviceID_NonExistentPath(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "does", "not", "exist")

	device, err := DeviceID(nonExistent)
	require.NoError(t, err)
	assert.Greater(t, device, uint64(0))
}

// fakeDiskSpaceLookup returns canned device ids/disk space per path, so the
// grouping/merge logic in groupDiskUsage can be tested against a controlled
// set of "volumes" instead of depending on the test host's real mount layout.
type fakeDiskSpaceLookup struct {
	devices map[string]uint64
	errs    map[string]error
}

func (f *fakeDiskSpaceLookup) deviceID(path string) (uint64, error) {
	if err, ok := f.errs[path]; ok {
		return 0, err
	}
	return f.devices[path], nil
}

func (f *fakeDiskSpaceLookup) diskSpace(path string) (*DiskSpace, error) {
	if err, ok := f.errs[path]; ok {
		return nil, err
	}
	return &DiskSpace{Available: 100, Free: 100, Total: 200, UsedPct: 50}, nil
}

func TestGroupDiskUsage_SharedVolumeMerged(t *testing.T) {
	fake := &fakeDiskSpaceLookup{devices: map[string]uint64{
		"/movies":  1,
		"/tvshows": 1,
	}}

	entries := groupDiskUsage([]NamedPath{
		{Label: "movies", Path: "/movies"},
		{Label: "tvshows", Path: "/tvshows"},
	}, fake.deviceID, fake.diskSpace)

	require.Len(t, entries, 1)
	assert.Equal(t, []string{"movies", "tvshows"}, entries[0].Labels)
	assert.False(t, entries[0].Unavailable)
	require.NotNil(t, entries[0].Space)
}

func TestGroupDiskUsage_DistinctVolumesSeparate(t *testing.T) {
	fake := &fakeDiskSpaceLookup{devices: map[string]uint64{
		"/movies":  1,
		"/archive": 2,
	}}

	entries := groupDiskUsage([]NamedPath{
		{Label: "movies", Path: "/movies"},
		{Label: "archive", Path: "/archive"},
	}, fake.deviceID, fake.diskSpace)

	require.Len(t, entries, 2)
	assert.Equal(t, []string{"movies"}, entries[0].Labels)
	assert.Equal(t, []string{"archive"}, entries[1].Labels)
}

func TestGroupDiskUsage_MissingPathIsUnavailable(t *testing.T) {
	fake := &fakeDiskSpaceLookup{
		devices: map[string]uint64{"/movies": 1},
		errs:    map[string]error{"/missing": fmt.Errorf("no such file or directory")},
	}

	entries := groupDiskUsage([]NamedPath{
		{Label: "movies", Path: "/movies"},
		{Label: "archive", Path: "/missing"},
	}, fake.deviceID, fake.diskSpace)

	require.Len(t, entries, 2)
	archiveEntry := entries[1]
	assert.Equal(t, []string{"archive"}, archiveEntry.Labels)
	assert.True(t, archiveEntry.Unavailable)
	assert.NotEmpty(t, archiveEntry.Reason)
	assert.Nil(t, archiveEntry.Space)
}

func TestGroupDiskUsage_RealPaths(t *testing.T) {
	tempDir := t.TempDir()
	subDirA := filepath.Join(tempDir, "movies")
	subDirB := filepath.Join(tempDir, "tvshows")
	require.NoError(t, os.MkdirAll(subDirA, 0o755))
	require.NoError(t, os.MkdirAll(subDirB, 0o755))

	entries := GroupDiskUsage([]NamedPath{
		{Label: "movies", Path: subDirA},
		{Label: "tvshows", Path: subDirB},
	})

	require.Len(t, entries, 1)
	assert.ElementsMatch(t, []string{"movies", "tvshows"}, entries[0].Labels)
	assert.False(t, entries[0].Unavailable)
	require.NotNil(t, entries[0].Space)
	assert.Greater(t, entries[0].Space.Total, uint64(0))
}

func TestCheckDiskSpaceBeforeDownload_NonExistentPath(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "future", "download", "path")

	// Should not error even if path doesn't exist yet
	err := CheckDiskSpaceBeforeDownload(nonExistent, 1024*1024, 0)
	assert.NoError(t, err)
}
