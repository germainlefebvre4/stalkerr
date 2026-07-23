package fileparser

import (
	"testing"
)

func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		downloadPath string
		tmdbYear     *int
		expected     *FileInfo
	}{
		{
			name:         "Empty path",
			downloadPath: "",
			tmdbYear:     intPtr(2020),
			expected:     nil,
		},
		{
			name:         "Complete Movie Path",
			downloadPath: "/media/movies/Matrix.1999.1080p.BluRay.x264/matrix.mkv",
			tmdbYear:     intPtr(1999),
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "Matrix.1999.1080p.BluRay.x264",
				FileName:      "matrix.mkv",
				HasYearInPath: true,
				YearMismatch:  false,
				DetectedYear:  intPtr(1999),
				DetectedRes:   strPtr("1080p"),
				IsValidFormat: true,
			},
		},
		{
			name:         "Missing Year",
			downloadPath: "/media/movies/Avatar.BluRay.1080p/avatar.mkv",
			tmdbYear:     intPtr(2009),
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "Avatar.BluRay.1080p",
				FileName:      "avatar.mkv",
				HasYearInPath: false,
				YearMismatch:  false,
				DetectedYear:  nil,
				DetectedRes:   strPtr("1080p"),
				IsValidFormat: true,
			},
		},
		{
			name:         "Year Mismatch",
			downloadPath: "/media/movies/Avatar.2010.1080p/avatar.mkv",
			tmdbYear:     intPtr(2009),
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "Avatar.2010.1080p",
				FileName:      "avatar.mkv",
				HasYearInPath: true,
				YearMismatch:  true,
				DetectedYear:  intPtr(2010),
				DetectedRes:   strPtr("1080p"),
				IsValidFormat: true,
			},
		},
		{
			name:         "Ambiguous Year",
			downloadPath: "/media/movies/2001.A.Space.Odyssey.1968.720p/movie.mkv",
			tmdbYear:     intPtr(1968),
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "2001.A.Space.Odyssey.1968.720p",
				FileName:      "movie.mkv",
				HasYearInPath: true,
				YearMismatch:  true, // 2001 != 1968
				DetectedYear:  intPtr(2001),
				DetectedRes:   strPtr("720p"),
				IsValidFormat: true,
			},
		},
		{
			name:         "TV Show without TMDB year",
			downloadPath: "/media/tvshows/Breaking.Bad/Season.05/Breaking.Bad.S05E14.1080p.mkv",
			tmdbYear:     nil,
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "Season.05",
				FileName:      "Breaking.Bad.S05E14.1080p.mkv",
				HasYearInPath: false,
				YearMismatch:  false,
				DetectedYear:  nil,
				DetectedRes:   strPtr("1080p"),
				IsValidFormat: true,
			},
		},
		{
			name:         "Case Variations (uppercase MKV)",
			downloadPath: "/media/movies/Avatar/avatar.MKV",
			tmdbYear:     nil,
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "Avatar",
				FileName:      "avatar.MKV",
				HasYearInPath: false,
				YearMismatch:  false,
				DetectedYear:  nil,
				DetectedRes:   nil,
				IsValidFormat: true,
			},
		},
		{
			name:         "Unknown/invalid Format",
			downloadPath: "/media/movies/Avatar/avatar.txt",
			tmdbYear:     nil,
			expected: &FileInfo{
				Extension:     ".txt",
				FolderName:    "Avatar",
				FileName:      "avatar.txt",
				HasYearInPath: false,
				YearMismatch:  false,
				DetectedYear:  nil,
				DetectedRes:   nil,
				IsValidFormat: false,
			},
		},
		{
			name:         "Resolution Variations (4K, 2160p, 1080P)",
			downloadPath: "/media/movies/Inception.2010.4K.2160p.1080P/movie.mp4",
			tmdbYear:     intPtr(2010),
			expected: &FileInfo{
				Extension:     ".mp4",
				FolderName:    "Inception.2010.4K.2160p.1080P",
				FileName:      "movie.mp4",
				HasYearInPath: true,
				YearMismatch:  false,
				DetectedYear:  intPtr(2010),
				DetectedRes:   strPtr("4k"), // First match of resolutionRegex is 4K (case normalized)
				IsValidFormat: true,
			},
		},
		{
			name:         "No folder name (flat path)",
			downloadPath: "matrix.mkv",
			tmdbYear:     nil,
			expected: &FileInfo{
				Extension:     ".mkv",
				FolderName:    "",
				FileName:      "matrix.mkv",
				HasYearInPath: false,
				YearMismatch:  false,
				DetectedYear:  nil,
				DetectedRes:   nil,
				IsValidFormat: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.downloadPath, tt.tmdbYear)
			if tt.expected == nil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %+v, got nil", tt.expected)
			}

			if got.Extension != tt.expected.Extension {
				t.Errorf("Extension: expected %q, got %q", tt.expected.Extension, got.Extension)
			}
			if got.FolderName != tt.expected.FolderName {
				t.Errorf("FolderName: expected %q, got %q", tt.expected.FolderName, got.FolderName)
			}
			if got.FileName != tt.expected.FileName {
				t.Errorf("FileName: expected %q, got %q", tt.expected.FileName, got.FileName)
			}
			if got.HasYearInPath != tt.expected.HasYearInPath {
				t.Errorf("HasYearInPath: expected %v, got %v", tt.expected.HasYearInPath, got.HasYearInPath)
			}
			if got.YearMismatch != tt.expected.YearMismatch {
				t.Errorf("YearMismatch: expected %v, got %v", tt.expected.YearMismatch, got.YearMismatch)
			}
			if (got.DetectedYear == nil) != (tt.expected.DetectedYear == nil) {
				t.Errorf("DetectedYear presence mismatch")
			} else if got.DetectedYear != nil && *got.DetectedYear != *tt.expected.DetectedYear {
				t.Errorf("DetectedYear: expected %d, got %d", *tt.expected.DetectedYear, *got.DetectedYear)
			}
			if (got.DetectedRes == nil) != (tt.expected.DetectedRes == nil) {
				t.Errorf("DetectedRes presence mismatch")
			} else if got.DetectedRes != nil && *got.DetectedRes != *tt.expected.DetectedRes {
				t.Errorf("DetectedRes: expected %s, got %s", *tt.expected.DetectedRes, *got.DetectedRes)
			}
			if got.IsValidFormat != tt.expected.IsValidFormat {
				t.Errorf("IsValidFormat: expected %v, got %v", tt.expected.IsValidFormat, got.IsValidFormat)
			}
		})
	}
}
