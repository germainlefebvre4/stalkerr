package api

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/classifier"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/fileparser"
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

// listDownloadsEnriched returns a paginated, metadata-enriched list of downloads with advanced filtering
func (s *Server) listDownloadsEnriched(c *gin.Context) {
	db := database.Get()
	limit, offset := parsePagination(c)
	status := c.Query("status")
	contentType := c.Query("type")
	problem := c.Query("problem")

	query := db.Model(&models.DownloadInfo{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if contentType != "" {
		query = query.Joins("JOIN processed_lines pl ON pl.download_info_id = download_info.id").
			Where("pl.content_type = ?", contentType).
			Distinct()
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
	if err := query.Preload("ProcessedLines.Movie").
		Preload("ProcessedLines.TVShow").
		Order("download_info.updated_at desc").
		Limit(limit).Offset(offset).
		Find(&downloads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to fetch downloads",
		})
		return
	}

	enriched := make([]DownloadEnrichedResponse, 0, len(downloads))
	for _, dl := range downloads {
		resp := enrichDownloadInfo(dl)
		if matchesProblem(resp, problem) {
			enriched = append(enriched, resp)
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       enriched,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	})
}

func enrichDownloadInfo(dl models.DownloadInfo) DownloadEnrichedResponse {
	resp := DownloadEnrichedResponse{
		ID:              dl.ID,
		URL:             dl.URL,
		Status:          dl.Status,
		DownloadPath:    dl.DownloadPath,
		FileSize:        dl.FileSize,
		BytesDownloaded: dl.BytesDownloaded,
		TotalBytes:      dl.TotalBytes,
		RetryCount:      dl.RetryCount,
		ErrorMessage:    dl.ErrorMessage,
		UpdatedAt:       dl.UpdatedAt,
	}

	var contentYear *int
	if len(dl.ProcessedLines) > 0 {
		contentInfo := buildContentInfo(dl.ProcessedLines[0])
		resp.Content = contentInfo
		contentYear = contentInfo.Year
	}

	if dl.DownloadPath != nil && *dl.DownloadPath != "" {
		resp.FileInfo = fileparser.Parse(*dl.DownloadPath, contentYear)
	}

	return resp
}

func buildContentInfo(pl models.ProcessedLine) *ContentInfo {
	info := &ContentInfo{}

	if pl.Resolution != nil {
		info.Resolution = pl.Resolution
	}

	if pl.Movie != nil {
		info.Type = "movies"
		info.Title = pl.Movie.TMDBTitle
		info.Year = &pl.Movie.TMDBYear
		info.Genres = pl.Movie.TMDBGenres
		info.Duration = pl.Movie.Duration
		return info
	}

	if pl.TVShow != nil {
		info.Type = "tvshows"
		title := pl.TVShow.TMDBTitle
		if pl.TVShow.Season != nil && pl.TVShow.Episode != nil {
			title = fmt.Sprintf("%s S%02dE%02d", pl.TVShow.TMDBTitle, *pl.TVShow.Season, *pl.TVShow.Episode)
		}
		info.Title = title
		info.Year = &pl.TVShow.TMDBYear
		info.Genres = pl.TVShow.TMDBGenres
		info.Season = pl.TVShow.Season
		info.Episode = pl.TVShow.Episode
		return info
	}

	info.Type = string(pl.ContentType)
	if info.Type == "" {
		info.Type = "uncategorized"
	}
	info.Title = pl.TvgName
	return info
}

func matchesProblem(resp DownloadEnrichedResponse, filter string) bool {
	if filter == "" {
		return true
	}
	if resp.FileInfo == nil {
		return false
	}
	switch filter {
	case "missing_year":
		return !resp.FileInfo.HasYearInPath
	case "year_mismatch":
		return resp.FileInfo.YearMismatch
	case "unknown_format":
		return !resp.FileInfo.IsValidFormat
	case "low_quality":
		if resp.FileInfo.DetectedRes == nil {
			return false
		}
		res := *resp.FileInfo.DetectedRes
		return res == "480p" || res == "360p"
	default:
		return true
	}
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

// searchTMDBProxy queries movies or TV shows on TMDB safely from the backend.
func (s *Server) searchTMDBProxy(c *gin.Context) {
	if s.tmdbClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "tmdb_disabled",
			Message: "TMDB integration is disabled or not configured",
		})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "query parameter is required",
		})
		return
	}

	mediaType := c.Query("type")
	if mediaType != "movie" && mediaType != "tvshow" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "type parameter must be either movie or tvshow",
		})
		return
	}

	var results []TMDBSearchResult

	if mediaType == "movie" {
		var yearPtr *int
		if yearStr := c.Query("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
				yearPtr = &y
			}
		}

		movies, err := s.tmdbClient.SearchMovies(query, yearPtr)
		if err != nil {
			// If no results found, TMDB client returns error. Return empty array instead of 500.
			if strings.Contains(err.Error(), "no results found") {
				c.JSON(http.StatusOK, []TMDBSearchResult{})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "tmdb_error",
				Message: fmt.Sprintf("failed to search TMDB: %v", err),
			})
			return
		}

		results = make([]TMDBSearchResult, len(movies))
		for i, m := range movies {
			results[i] = TMDBSearchResult{
				ID:            m.ID,
				Title:         m.Title,
				OriginalTitle: m.OriginalTitle,
				ReleaseDate:   m.ReleaseDate,
				Overview:      m.Overview,
				PosterPath:    m.PosterPath,
			}
		}
	} else {
		shows, err := s.tmdbClient.SearchTVShows(query)
		if err != nil {
			if strings.Contains(err.Error(), "no results found") {
				c.JSON(http.StatusOK, []TMDBSearchResult{})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "tmdb_error",
				Message: fmt.Sprintf("failed to search TMDB: %v", err),
			})
			return
		}

		results = make([]TMDBSearchResult, len(shows))
		for i, show := range shows {
			results[i] = TMDBSearchResult{
				ID:            show.ID,
				Title:         show.Name,
				OriginalTitle: show.OriginalName,
				ReleaseDate:   show.FirstAirDate,
				Overview:      show.Overview,
				PosterPath:    show.PosterPath,
			}
		}
	}

	c.JSON(http.StatusOK, results)
}

// overrideItem manually associates a VOD item with a specific TMDB movie or TV show.
func (s *Server) overrideItem(c *gin.Context) {
	if s.tmdbClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "tmdb_disabled",
			Message: "TMDB integration is disabled or not configured",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "invalid item id",
		})
		return
	}

	db := database.Get()
	var item models.ProcessedLine
	if err := db.Preload("Movie").Preload("TVShow").First(&item, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: fmt.Sprintf("failed to fetch item: %v", err),
		})
		return
	}

	var req OverrideItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: fmt.Sprintf("invalid override payload: %v", err),
		})
		return
	}

	var season, episode *int
	now := time.Now()

	if req.Type == "movie" {
		details, err := s.tmdbClient.GetMovieDetails(req.TMDBID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "override_failed",
				Message: fmt.Sprintf("failed to fetch TMDB movie details: %v", err),
			})
			return
		}

		externalIDs, _ := s.tmdbClient.GetMovieExternalIDs(req.TMDBID)

		var movie models.Movie
		tmdbYear := tmdb.ExtractYear(details.ReleaseDate)
		genres := tmdb.FormatGenres(details.Genres)

		var tvdbID *int
		var imdbID *string
		if externalIDs != nil {
			tvdbID = externalIDs.TVDBID
			imdbID = externalIDs.IMDBID
		}

		attrs := models.Movie{
			TMDBID:     details.ID,
			TVDBID:     tvdbID,
			TMDBTitle:  details.Title,
			TMDBYear:   tmdbYear,
			TMDBGenres: &genres,
			Duration:   details.Runtime,
			PosterPath: details.PosterPath,
			Overview:   &details.Overview,
			IMDBID:     imdbID,
		}

		if err := db.Where("tmdb_id = ? AND tmdb_year = ?", details.ID, tmdbYear).Attrs(attrs).FirstOrCreate(&movie).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "override_failed",
				Message: fmt.Sprintf("failed to save movie details: %v", err),
			})
			return
		}

		if externalIDs != nil && externalIDs.TVDBID != nil && movie.TVDBID == nil {
			movie.TVDBID = externalIDs.TVDBID
			db.Save(&movie)
		}

		item.ContentType = models.ContentTypeMovies
		item.MovieID = &movie.ID
		item.TVShowID = nil
		item.ChannelID = nil
		item.UncategorizedID = nil
		item.Movie = &movie
		item.TVShow = nil
	} else {
		if req.Season != nil {
			season = req.Season
		}
		if req.Episode != nil {
			episode = req.Episode
		}
		if season == nil || episode == nil {
			cl := classifier.New()
			extractedSeason, extractedEpisode := cl.ExtractSeasonEpisode(item.TvgName)
			if season == nil {
				season = extractedSeason
			}
			if episode == nil {
				episode = extractedEpisode
			}
		}

		details, err := s.tmdbClient.GetTVShowDetails(req.TMDBID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "override_failed",
				Message: fmt.Sprintf("failed to fetch TMDB TV show details: %v", err),
			})
			return
		}

		externalIDs, _ := s.tmdbClient.GetTVShowExternalIDs(req.TMDBID)

		var tvshow models.TVShow
		tmdbYear := tmdb.ExtractYear(details.FirstAirDate)
		genres := tmdb.FormatGenres(details.Genres)

		var tvdbID *int
		var imdbID *string
		if externalIDs != nil {
			tvdbID = externalIDs.TVDBID
			imdbID = externalIDs.IMDBID
		}

		attrs := models.TVShow{
			TMDBID:     details.ID,
			TVDBID:     tvdbID,
			TMDBTitle:  details.Name,
			TMDBYear:   tmdbYear,
			TMDBGenres: &genres,
			Season:     season,
			Episode:    episode,
			PosterPath: details.PosterPath,
			Overview:   &details.Overview,
			IMDBID:     imdbID,
		}

		query := db.Where("tmdb_id = ?", details.ID)
		if season != nil {
			query = query.Where("season = ?", *season)
		} else {
			query = query.Where("season IS NULL")
		}
		if episode != nil {
			query = query.Where("episode = ?", *episode)
		} else {
			query = query.Where("episode IS NULL")
		}

		if err := query.Attrs(attrs).FirstOrCreate(&tvshow).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "override_failed",
				Message: fmt.Sprintf("failed to save TV show details: %v", err),
			})
			return
		}

		if externalIDs != nil && externalIDs.TVDBID != nil && tvshow.TVDBID == nil {
			tvshow.TVDBID = externalIDs.TVDBID
			db.Save(&tvshow)
		}

		item.ContentType = models.ContentTypeTVShows
		item.TVShowID = &tvshow.ID
		item.MovieID = nil
		item.ChannelID = nil
		item.UncategorizedID = nil
		item.TVShow = &tvshow
		item.Movie = nil
	}

	// Persist the learned mapping for future automatic imports
	var existingMapping models.ManualMapping
	if err := db.Where("tvg_name = ? AND group_title = ?", item.TvgName, item.GroupTitle).First(&existingMapping).Error; err == nil {
		existingMapping.ContentType = item.ContentType
		existingMapping.TMDBID = req.TMDBID
		existingMapping.Season = season
		existingMapping.Episode = episode
		existingMapping.UpdatedAt = now
		db.Save(&existingMapping)
	} else {
		mapping := models.ManualMapping{
			TvgName:     item.TvgName,
			GroupTitle:  item.GroupTitle,
			ContentType: item.ContentType,
			TMDBID:      req.TMDBID,
			Season:      season,
			Episode:     episode,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		db.Create(&mapping)
	}

	overrideBy := "manual"
	item.OverrideBy = &overrideBy
	item.OverrideAt = &now

	if err := db.Select("ContentType", "MovieID", "TVShowID", "ChannelID", "UncategorizedID", "OverrideBy", "OverrideAt").Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "override_failed",
			Message: fmt.Sprintf("failed to save override association: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, toItemResponse(item))
}
