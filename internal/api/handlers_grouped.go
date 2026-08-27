package api

import (
	"database/sql/driver"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/database"
)

// flexTime scans a timestamp column regardless of whether the underlying driver
// returns a native time.Time (e.g. postgres) or a formatted string (sqlite,
// notably for values produced by an aggregate expression like MAX(created_at),
// which loses the column's declared type information).
type flexTime struct {
	time.Time
}

var flexTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",
}

func (ft *flexTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ft.Time = v
		return nil
	case string:
		for _, layout := range flexTimeLayouts {
			if t, err := time.Parse(layout, v); err == nil {
				ft.Time = t
				return nil
			}
		}
		return fmt.Errorf("flexTime: cannot parse time value %q", v)
	case []byte:
		return ft.Scan(string(v))
	default:
		return fmt.Errorf("flexTime: unsupported source type %T", value)
	}
}

func (ft flexTime) Value() (driver.Value, error) {
	return ft.Time, nil
}

// itemGroupRow mirrors the column shape produced by the UNION ALL of the four
// grouped-listing aggregates (movies, tvshows, unmatched movies, unmatched tvshows).
type itemGroupRow struct {
	Type           string   `gorm:"column:type"`
	MovieID        *uint    `gorm:"column:movie_id"`
	TMDBID         *int     `gorm:"column:tmdb_id"`
	Title          *string  `gorm:"column:title"`
	Year           *int     `gorm:"column:year"`
	SeasonStart    *int     `gorm:"column:season_start"`
	SeasonEnd      *int     `gorm:"column:season_end"`
	LatestActivity flexTime `gorm:"column:latest_activity"`
}

func (r itemGroupRow) toResponse() ItemGroupResponse {
	return ItemGroupResponse{
		Type:           r.Type,
		MovieID:        r.MovieID,
		TMDBID:         r.TMDBID,
		Title:          r.Title,
		Year:           r.Year,
		SeasonStart:    r.SeasonStart,
		SeasonEnd:      r.SeasonEnd,
		LatestActivity: r.LatestActivity.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// listItemGroups returns a paginated, server-side aggregation of processed_lines
// collapsed to one row per distinct movie or TV show (plus two fixed pseudo-groups
// for unmatched items), ordered by each group's most recent underlying item
// activity descending. See openspec/changes/playlist-grouped-media-view/design.md.
func (s *Server) listItemGroups(c *gin.Context) {
	db := database.Get()
	limit, offset := parsePagination(c)

	filterSQL, filterArgs := buildItemFilterConditions(c, db)
	filterClause := ""
	if filterSQL != "" {
		filterClause = "AND " + filterSQL
	}

	moviesArm := fmt.Sprintf(`
		SELECT
			'movie' AS type,
			m.id AS movie_id,
			CAST(NULL AS INTEGER) AS tmdb_id,
			m.tmdb_title AS title,
			m.tmdb_year AS year,
			CAST(NULL AS INTEGER) AS season_start,
			CAST(NULL AS INTEGER) AS season_end,
			MAX(pl.created_at) AS latest_activity
		FROM processed_lines pl
		JOIN movies m ON m.id = pl.movie_id
		WHERE pl.content_type = 'movies' AND pl.movie_id IS NOT NULL %s
		GROUP BY m.id, m.tmdb_title, m.tmdb_year
	`, filterClause)

	tvshowsArm := fmt.Sprintf(`
		SELECT
			'tvshow' AS type,
			CAST(NULL AS INTEGER) AS movie_id,
			t.tmdb_id AS tmdb_id,
			MAX(t.tmdb_title) AS title,
			MAX(t.tmdb_year) AS year,
			MIN(t.season) AS season_start,
			MAX(t.season) AS season_end,
			MAX(pl.created_at) AS latest_activity
		FROM processed_lines pl
		JOIN tvshows t ON t.id = pl.tv_show_id
		WHERE pl.content_type = 'tvshows' AND pl.tv_show_id IS NOT NULL %s
		GROUP BY t.tmdb_id
	`, filterClause)

	unmatchedMoviesArm := fmt.Sprintf(`
		SELECT
			'unmatched_movies' AS type,
			CAST(NULL AS INTEGER) AS movie_id,
			CAST(NULL AS INTEGER) AS tmdb_id,
			CAST(NULL AS TEXT) AS title,
			CAST(NULL AS INTEGER) AS year,
			CAST(NULL AS INTEGER) AS season_start,
			CAST(NULL AS INTEGER) AS season_end,
			MAX(pl.created_at) AS latest_activity
		FROM processed_lines pl
		WHERE pl.content_type = 'movies' AND pl.movie_id IS NULL %s
		HAVING COUNT(*) > 0
	`, filterClause)

	unmatchedTVShowsArm := fmt.Sprintf(`
		SELECT
			'unmatched_tvshows' AS type,
			CAST(NULL AS INTEGER) AS movie_id,
			CAST(NULL AS INTEGER) AS tmdb_id,
			CAST(NULL AS TEXT) AS title,
			CAST(NULL AS INTEGER) AS year,
			CAST(NULL AS INTEGER) AS season_start,
			CAST(NULL AS INTEGER) AS season_end,
			MAX(pl.created_at) AS latest_activity
		FROM processed_lines pl
		WHERE pl.content_type = 'tvshows' AND pl.tv_show_id IS NULL %s
		HAVING COUNT(*) > 0
	`, filterClause)

	// SQLite's compound-select grammar rejects parenthesized arms (each arm
	// must be a bare SELECT), unlike postgres which allows either form, so the
	// arms are joined without wrapping parens to stay portable across both.
	unionSQL := fmt.Sprintf("%s UNION ALL %s UNION ALL %s UNION ALL %s",
		moviesArm, tvshowsArm, unmatchedMoviesArm, unmatchedTVShowsArm)

	// Each arm references the same filter placeholders, in the same order, so
	// the shared filter args are repeated once per arm.
	var unionArgs []interface{}
	for i := 0; i < 4; i++ {
		unionArgs = append(unionArgs, filterArgs...)
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS grouped", unionSQL)
	if err := db.Raw(countSQL, unionArgs...).Scan(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to count grouped items",
		})
		return
	}

	dataArgs := append(append([]interface{}{}, unionArgs...), limit, offset)
	dataSQL := fmt.Sprintf("SELECT * FROM (%s) AS grouped ORDER BY latest_activity DESC LIMIT ? OFFSET ?", unionSQL)

	var rows []itemGroupRow
	if err := db.Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to fetch grouped items",
		})
		return
	}

	responses := make([]ItemGroupResponse, len(rows))
	for i, row := range rows {
		responses[i] = row.toResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	})
}
