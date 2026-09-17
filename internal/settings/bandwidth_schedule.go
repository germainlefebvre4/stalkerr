package settings

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// Schedule/policy action values, ordered from least to most restrictive.
// Shared by the weekly bandwidth schedule and, via MostRestrictive, by the
// adaptive-download-throttling effective-policy combination. See the
// download-bandwidth-schedule spec's "Resolving the Active Window".
const (
	ActionNone     = "none"
	ActionThrottle = "throttle"
	ActionStop     = "stop"
)

// ErrScheduleWindowNotFound is returned when updating or deleting a
// schedule window id that does not exist.
var ErrScheduleWindowNotFound = errors.New("schedule window not found")

var validActions = map[string]bool{ActionNone: true, ActionThrottle: true, ActionStop: true}

var validDaysOfWeek = map[string]bool{
	"sunday": true, "monday": true, "tuesday": true, "wednesday": true,
	"thursday": true, "friday": true, "saturday": true,
}

// InvalidScheduleWindowError is returned when a schedule window input fails
// validation (no days, an unparseable time, or an unknown action).
type InvalidScheduleWindowError struct {
	Reason string
}

func (e *InvalidScheduleWindowError) Error() string {
	return e.Reason
}

// ScheduleWindowInput is the create/update payload for one bandwidth
// schedule window.
type ScheduleWindowInput struct {
	DaysOfWeek []string
	StartTime  string
	EndTime    string
	Action     string
}

func (in ScheduleWindowInput) validate() error {
	if len(in.DaysOfWeek) == 0 {
		return &InvalidScheduleWindowError{Reason: "days_of_week must include at least one day"}
	}
	for _, d := range in.DaysOfWeek {
		if !validDaysOfWeek[strings.ToLower(d)] {
			return &InvalidScheduleWindowError{Reason: fmt.Sprintf("invalid day of week: %q", d)}
		}
	}
	if !isValidTimeOfDay(in.StartTime) {
		return &InvalidScheduleWindowError{Reason: fmt.Sprintf("invalid start_time: %q", in.StartTime)}
	}
	if !isValidTimeOfDay(in.EndTime) {
		return &InvalidScheduleWindowError{Reason: fmt.Sprintf("invalid end_time: %q", in.EndTime)}
	}
	if !validActions[in.Action] {
		return &InvalidScheduleWindowError{Reason: fmt.Sprintf("invalid action: %q", in.Action)}
	}
	return nil
}

func isValidTimeOfDay(s string) bool {
	_, err := time.Parse("15:04", s)
	return err == nil
}

// ListScheduleWindows returns every stored bandwidth schedule window,
// ordered by id.
func ListScheduleWindows() ([]models.BandwidthScheduleWindow, error) {
	var rows []models.BandwidthScheduleWindow
	if err := database.Get().Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateScheduleWindow stores a new bandwidth schedule window. See
// "Managing Schedule Windows Independently".
func CreateScheduleWindow(input ScheduleWindowInput) (models.BandwidthScheduleWindow, error) {
	if err := input.validate(); err != nil {
		return models.BandwidthScheduleWindow{}, err
	}

	row := models.BandwidthScheduleWindow{
		DaysOfWeek: normalizeDays(input.DaysOfWeek),
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
		Action:     input.Action,
	}
	if err := database.Get().Create(&row).Error; err != nil {
		return models.BandwidthScheduleWindow{}, err
	}
	return row, nil
}

// UpdateScheduleWindow replaces the stored fields of schedule window id,
// leaving every other window unchanged.
func UpdateScheduleWindow(id uint, input ScheduleWindowInput) (models.BandwidthScheduleWindow, error) {
	if err := input.validate(); err != nil {
		return models.BandwidthScheduleWindow{}, err
	}

	var row models.BandwidthScheduleWindow
	if err := database.Get().First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.BandwidthScheduleWindow{}, ErrScheduleWindowNotFound
		}
		return models.BandwidthScheduleWindow{}, err
	}

	row.DaysOfWeek = normalizeDays(input.DaysOfWeek)
	row.StartTime = input.StartTime
	row.EndTime = input.EndTime
	row.Action = input.Action

	if err := database.Get().Save(&row).Error; err != nil {
		return models.BandwidthScheduleWindow{}, err
	}
	return row, nil
}

// DeleteScheduleWindow removes schedule window id. See "Managing Schedule
// Windows Independently": every other window is left unchanged.
func DeleteScheduleWindow(id uint) error {
	result := database.Get().Delete(&models.BandwidthScheduleWindow{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrScheduleWindowNotFound
	}
	return nil
}

// MostRestrictive returns the more restrictive of two actions, using the
// ordering stop > throttle > none. See "Resolving the Active Window" and
// adaptive-download-throttling's "Effective Policy Combination".
func MostRestrictive(a, b string) string {
	if actionRank(a) >= actionRank(b) {
		return a
	}
	return b
}

func actionRank(a string) int {
	switch a {
	case ActionStop:
		return 2
	case ActionThrottle:
		return 1
	default:
		return 0
	}
}

// ActiveScheduleAction resolves the currently active schedule action against
// windows, evaluated at instant t (server local time). It is a pure,
// in-memory computation with no I/O, so a caller loads windows once (e.g.
// via ListScheduleWindows) and re-evaluates it cheaply on every check. When
// no window is active - including when windows is empty - it returns
// ActionNone. When more than one window is simultaneously active, the most
// restrictive of their actions is returned. See "Resolving the Active
// Window", "Overnight Windows", and "Default Schedule Is Unrestricted".
func ActiveScheduleAction(windows []models.BandwidthScheduleWindow, t time.Time) string {
	active := ActionNone
	for _, w := range windows {
		if windowActive(w, t) {
			active = MostRestrictive(active, w.Action)
		}
	}
	return active
}

func windowActive(w models.BandwidthScheduleWindow, t time.Time) bool {
	start, err := time.Parse("15:04", w.StartTime)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", w.EndTime)
	if err != nil {
		return false
	}

	nowMinutes := t.Hour()*60 + t.Minute()
	startMinutes := start.Hour()*60 + start.Minute()
	endMinutes := end.Hour()*60 + end.Minute()

	today := strings.ToLower(t.Weekday().String())
	yesterday := strings.ToLower(t.AddDate(0, 0, -1).Weekday().String())

	if endMinutes > startMinutes {
		// Ordinary same-day window.
		return containsDay(w.DaysOfWeek, today) && nowMinutes >= startMinutes && nowMinutes < endMinutes
	}

	// Overnight window (end time earlier than or equal to start time):
	// spans from its start time on its configured day(s) through its end
	// time on the following day. See "Overnight Windows".
	if containsDay(w.DaysOfWeek, today) && nowMinutes >= startMinutes {
		return true
	}
	if containsDay(w.DaysOfWeek, yesterday) && nowMinutes < endMinutes {
		return true
	}
	return false
}

func containsDay(days models.StringList, day string) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func normalizeDays(days []string) models.StringList {
	out := make(models.StringList, len(days))
	for i, d := range days {
		out[i] = strings.ToLower(d)
	}
	return out
}
