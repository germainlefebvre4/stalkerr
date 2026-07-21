package api

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// listProcessingLogs returns a paginated list of background processing logs
func (s *Server) listProcessingLogs(c *gin.Context) {
	db := database.Get()
	limit, offset := parsePagination(c)
	status := c.Query("status")

	query := db.Model(&models.ProcessingLog{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to count processing logs",
		})
		return
	}

	var logs []models.ProcessingLog
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to fetch processing logs",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       logs,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	})
}

// listDownloads returns a paginated list of downloads with status filtering
func (s *Server) listDownloads(c *gin.Context) {
	db := database.Get()
	limit, offset := parsePagination(c)
	status := c.Query("status")

	query := db.Model(&models.DownloadInfo{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to count downloads",
		})
		return
	}

	var downloads []models.DownloadInfo
	if err := query.Order("updated_at desc").Limit(limit).Offset(offset).Find(&downloads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to fetch downloads",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       downloads,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	})
}

// getConfigPaths returns configured default storage directories
func (s *Server) getConfigPaths(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, gin.H{
		"movies_path":  cfg.Downloads.MoviesPath,
		"tvshows_path": cfg.Downloads.TVShowsPath,
	})
}

// MoveDownloadRequest represents folder move payload
type MoveDownloadRequest struct {
	DestinationParentDir string `json:"destination_parent_dir" binding:"required"`
}

// moveMovieFolder moves the entire movie folder to a new parent folder
func (s *Server) moveMovieFolder(c *gin.Context) {
	db := database.Get()
	id := c.Param("id")

	var movie models.Movie
	if err := db.Preload("ProcessedLines").First(&movie, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Movie not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch movie",
		})
		return
	}

	var currentMovieDir string
	var downloadInfos []models.DownloadInfo

	for _, line := range movie.ProcessedLines {
		if line.DownloadInfoID != nil {
			var dl models.DownloadInfo
			if err := db.First(&dl, *line.DownloadInfoID).Error; err == nil && dl.DownloadPath != nil && *dl.DownloadPath != "" {
				downloadInfos = append(downloadInfos, dl)
				if currentMovieDir == "" {
					currentMovieDir = filepath.Dir(*dl.DownloadPath)
				}
			}
		}
	}

	if currentMovieDir == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_completed_downloads",
			Message: "No completed downloads with valid file paths found for this movie",
		})
		return
	}

	var req MoveDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	movieFolderName := filepath.Base(currentMovieDir)
	targetMovieDir := filepath.Join(req.DestinationParentDir, movieFolderName)

	if filepath.Clean(currentMovieDir) == filepath.Clean(targetMovieDir) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "same_directory",
			Message: "Target movie directory is the same as the current directory",
		})
		return
	}

	// Ensure destination base folder exists
	if err := os.MkdirAll(req.DestinationParentDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fs_error",
			Message: "Failed to create target base directory: " + err.Error(),
		})
		return
	}

	// Execute physical movement on disk
	if err := MoveDir(currentMovieDir, targetMovieDir); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "move_failed",
			Message: "Failed to move physical folder: " + err.Error(),
		})
		return
	}

	// Atomically update associated download info records in database
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, dl := range downloadInfos {
			oldPath := *dl.DownloadPath
			newPath := filepath.Join(targetMovieDir, filepath.Base(oldPath))
			if err := tx.Model(&models.DownloadInfo{}).Where("id = ?", dl.ID).Update("download_path", newPath).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_update_failed",
			Message: "Directory moved successfully on disk, but database path updates failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"new_path": targetMovieDir,
	})
}

// moveTVShowFolder moves the entire TV show folder (all seasons) to a new parent folder
func (s *Server) moveTVShowFolder(c *gin.Context) {
	db := database.Get()
	id := c.Param("id")

	var tvShow models.TVShow
	if err := db.Preload("ProcessedLines").First(&tvShow, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "TV show not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch TV show",
		})
		return
	}

	var seriesDir string
	var downloadInfos []models.DownloadInfo

	for _, line := range tvShow.ProcessedLines {
		if line.DownloadInfoID != nil {
			var dl models.DownloadInfo
			if err := db.First(&dl, *line.DownloadInfoID).Error; err == nil && dl.DownloadPath != nil && *dl.DownloadPath != "" {
				downloadInfos = append(downloadInfos, dl)
				if seriesDir == "" {
					parentDir := filepath.Dir(*dl.DownloadPath)
					grandparentDir := filepath.Dir(parentDir)

					if strings.HasPrefix(strings.ToLower(filepath.Base(parentDir)), "season") {
						seriesDir = grandparentDir
					} else {
						seriesDir = parentDir
					}
				}
			}
		}
	}

	if seriesDir == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_completed_downloads",
			Message: "No completed downloads with valid paths found for this TV show",
		})
		return
	}

	var req MoveDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	seriesFolderName := filepath.Base(seriesDir)
	targetSeriesDir := filepath.Join(req.DestinationParentDir, seriesFolderName)

	if filepath.Clean(seriesDir) == filepath.Clean(targetSeriesDir) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "same_directory",
			Message: "Target series directory is the same as the current directory",
		})
		return
	}

	// Ensure destination base folder exists
	if err := os.MkdirAll(req.DestinationParentDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fs_error",
			Message: "Failed to create target base directory: " + err.Error(),
		})
		return
	}

	// Execute physical movement on disk
	if err := MoveDir(seriesDir, targetSeriesDir); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "move_failed",
			Message: "Failed to move physical TV show directory: " + err.Error(),
		})
		return
	}

	// Atomically update associated download info records in database
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, dl := range downloadInfos {
			oldPath := *dl.DownloadPath
			relPath, err := filepath.Rel(seriesDir, oldPath)
			if err != nil {
				return err
			}
			newPath := filepath.Join(targetSeriesDir, relPath)
			if err := tx.Model(&models.DownloadInfo{}).Where("id = ?", dl.ID).Update("download_path", newPath).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_update_failed",
			Message: "Directory moved successfully on disk, but database path updates failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"new_path": targetSeriesDir,
	})
}

// MoveDir moves a directory from src to dst, trying rename first, then copy+verify+delete fallbacks
func MoveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := copyDir(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("recursive copy failed: %w", err)
	}

	srcSize, srcCount, err := getDirInfo(src)
	if err != nil {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("failed to gather source directory information: %w", err)
	}

	dstSize, dstCount, err := getDirInfo(dst)
	if err != nil {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("failed to gather destination directory information: %w", err)
	}

	if srcSize != dstSize || srcCount != dstCount {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("size or count mismatch: srcSize=%d dstSize=%d srcCount=%d dstCount=%d", srcSize, dstSize, srcCount, dstCount)
	}

	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

func getDirInfo(path string) (int64, int, error) {
	var size int64
	var count int

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return size, count, err
}
