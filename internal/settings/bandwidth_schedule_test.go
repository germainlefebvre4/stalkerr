package settings

import (
	"errors"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestListScheduleWindows_EmptyWhenNoneCreated(t *testing.T) {
	setupTestDB(t)

	windows, err := ListScheduleWindows()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(windows) != 0 {
		t.Errorf("expected 0 windows, got %d", len(windows))
	}
}

func TestCreateScheduleWindow_StoresAndLists(t *testing.T) {
	setupTestDB(t)

	created, err := CreateScheduleWindow(ScheduleWindowInput{
		DaysOfWeek: []string{"Monday", "Tuesday"},
		StartTime:  "08:00",
		EndTime:    "18:00",
		Action:     ActionThrottle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected a non-zero id to be assigned")
	}
	if len(created.DaysOfWeek) != 2 || created.DaysOfWeek[0] != "monday" {
		t.Errorf("expected normalized lowercase days, got %v", created.DaysOfWeek)
	}

	windows, err := ListScheduleWindows()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
	if windows[0].Action != ActionThrottle {
		t.Errorf("expected action throttle, got %q", windows[0].Action)
	}
}

func TestCreateScheduleWindow_RejectsInvalidInput(t *testing.T) {
	setupTestDB(t)

	cases := []ScheduleWindowInput{
		{DaysOfWeek: nil, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle},
		{DaysOfWeek: []string{"someday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle},
		{DaysOfWeek: []string{"monday"}, StartTime: "not-a-time", EndTime: "18:00", Action: ActionThrottle},
		{DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: "invalid"},
	}
	for _, c := range cases {
		if _, err := CreateScheduleWindow(c); err == nil {
			t.Errorf("expected validation error for input %+v", c)
		}
	}
}

func TestUpdateScheduleWindow_ChangesFields(t *testing.T) {
	setupTestDB(t)

	created, err := CreateScheduleWindow(ScheduleWindowInput{
		DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := UpdateScheduleWindow(created.ID, ScheduleWindowInput{
		DaysOfWeek: []string{"friday", "saturday"}, StartTime: "22:00", EndTime: "07:00", Action: ActionStop,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Action != ActionStop || updated.StartTime != "22:00" || updated.EndTime != "07:00" {
		t.Errorf("unexpected updated window: %+v", updated)
	}
	if len(updated.DaysOfWeek) != 2 {
		t.Errorf("expected 2 days, got %v", updated.DaysOfWeek)
	}
}

func TestUpdateScheduleWindow_NotFound(t *testing.T) {
	setupTestDB(t)

	_, err := UpdateScheduleWindow(999, ScheduleWindowInput{
		DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle,
	})
	if !errors.Is(err, ErrScheduleWindowNotFound) {
		t.Errorf("expected ErrScheduleWindowNotFound, got %v", err)
	}
}

func TestDeleteScheduleWindow_LeavesOthersIntact(t *testing.T) {
	setupTestDB(t)

	a, err := CreateScheduleWindow(ScheduleWindowInput{DaysOfWeek: []string{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := CreateScheduleWindow(ScheduleWindowInput{DaysOfWeek: []string{"friday"}, StartTime: "22:00", EndTime: "07:00", Action: ActionStop})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := DeleteScheduleWindow(a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	windows, err := ListScheduleWindows()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(windows) != 1 || windows[0].ID != b.ID {
		t.Errorf("expected only window b to remain, got %+v", windows)
	}
}

func TestDeleteScheduleWindow_NotFound(t *testing.T) {
	setupTestDB(t)

	if err := DeleteScheduleWindow(999); !errors.Is(err, ErrScheduleWindowNotFound) {
		t.Errorf("expected ErrScheduleWindowNotFound, got %v", err)
	}
}

func mustTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("failed to parse time %q: %v", value, err)
	}
	return parsed
}

func TestActiveScheduleAction_NoWindowsDefined(t *testing.T) {
	now := mustTime(t, "2006-01-02 15:04", "2024-01-01 12:00") // a Monday
	if got := ActiveScheduleAction(nil, now); got != ActionNone {
		t.Errorf("expected ActionNone with no windows, got %q", got)
	}
}

func TestActiveScheduleAction_SingleOrdinaryWindow(t *testing.T) {
	windows := []models.BandwidthScheduleWindow{
		{DaysOfWeek: models.StringList{"monday", "tuesday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle},
	}

	inside := mustTime(t, "2006-01-02 15:04", "2024-01-01 10:00") // Monday 10:00
	if got := ActiveScheduleAction(windows, inside); got != ActionThrottle {
		t.Errorf("expected throttle inside window, got %q", got)
	}

	beforeStart := mustTime(t, "2006-01-02 15:04", "2024-01-01 07:59")
	if got := ActiveScheduleAction(windows, beforeStart); got != ActionNone {
		t.Errorf("expected none before window start, got %q", got)
	}

	atEnd := mustTime(t, "2006-01-02 15:04", "2024-01-01 18:00")
	if got := ActiveScheduleAction(windows, atEnd); got != ActionNone {
		t.Errorf("expected none at window end (exclusive), got %q", got)
	}

	wrongDay := mustTime(t, "2006-01-02 15:04", "2024-01-03 10:00") // Wednesday
	if got := ActiveScheduleAction(windows, wrongDay); got != ActionNone {
		t.Errorf("expected none on a day not in the window, got %q", got)
	}
}

func TestActiveScheduleAction_OvernightWindowCrossingMidnight(t *testing.T) {
	windows := []models.BandwidthScheduleWindow{
		{DaysOfWeek: models.StringList{"friday"}, StartTime: "22:00", EndTime: "07:00", Action: ActionStop},
	}

	fridayNight := mustTime(t, "2006-01-02 15:04", "2024-01-05 23:00") // Friday 23:00
	if got := ActiveScheduleAction(windows, fridayNight); got != ActionStop {
		t.Errorf("expected stop late Friday night, got %q", got)
	}

	saturdayEarly := mustTime(t, "2006-01-02 15:04", "2024-01-06 05:00") // Saturday 05:00
	if got := ActiveScheduleAction(windows, saturdayEarly); got != ActionStop {
		t.Errorf("expected stop early Saturday morning, got %q", got)
	}

	saturdayLate := mustTime(t, "2006-01-02 15:04", "2024-01-06 08:00") // Saturday 08:00, past window end
	if got := ActiveScheduleAction(windows, saturdayLate); got != ActionNone {
		t.Errorf("expected none after the overnight window ends, got %q", got)
	}

	fridayMorning := mustTime(t, "2006-01-02 15:04", "2024-01-05 10:00") // Friday 10:00, before start
	if got := ActiveScheduleAction(windows, fridayMorning); got != ActionNone {
		t.Errorf("expected none before the overnight window starts, got %q", got)
	}
}

func TestActiveScheduleAction_OverlappingWindowsMostRestrictiveWins(t *testing.T) {
	windows := []models.BandwidthScheduleWindow{
		{DaysOfWeek: models.StringList{"monday"}, StartTime: "08:00", EndTime: "18:00", Action: ActionThrottle},
		{DaysOfWeek: models.StringList{"monday"}, StartTime: "12:00", EndTime: "13:00", Action: ActionStop},
	}

	duringOverlap := mustTime(t, "2006-01-02 15:04", "2024-01-01 12:30")
	if got := ActiveScheduleAction(windows, duringOverlap); got != ActionStop {
		t.Errorf("expected stop (most restrictive) during overlap, got %q", got)
	}

	outsideOverlap := mustTime(t, "2006-01-02 15:04", "2024-01-01 09:00")
	if got := ActiveScheduleAction(windows, outsideOverlap); got != ActionThrottle {
		t.Errorf("expected throttle outside the stop window, got %q", got)
	}
}

func TestMostRestrictive_Ordering(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{ActionNone, ActionThrottle, ActionThrottle},
		{ActionThrottle, ActionStop, ActionStop},
		{ActionStop, ActionNone, ActionStop},
		{ActionNone, ActionNone, ActionNone},
	}
	for _, c := range cases {
		if got := MostRestrictive(c.a, c.b); got != c.want {
			t.Errorf("MostRestrictive(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}
