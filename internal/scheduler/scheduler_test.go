package scheduler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
)

func movieStream(id int, tier Tier) *Stream {
	return &Stream{
		Tier:      tier,
		SourceKey: fmt.Sprintf("movie:%d", id),
		Items:     []Item{{DisplayName: fmt.Sprintf("Movie %d", id)}},
	}
}

func seasonStream(seriesID, season int, tier Tier, episodes ...int) *Stream {
	items := make([]Item, 0, len(episodes))
	for _, ep := range episodes {
		items = append(items, Item{
			DisplayName: fmt.Sprintf("Series %d S%02dE%02d", seriesID, season, ep),
			Episode:     ep,
		})
	}
	SortSeasonItems(items)
	return &Stream{
		Tier:      tier,
		SourceKey: fmt.Sprintf("series:%d", seriesID),
		SeriesID:  seriesID,
		Season:    season,
		Items:     items,
	}
}

// 1.1 Stream/Item construction from fixture data.
func TestStreamConstruction_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		stream        *Stream
		wantTier      Tier
		wantItemCount int
	}{
		{"single movie", movieStream(1, Tier1), Tier1, 1},
		{"season with three episodes", seasonStream(10, 1, Tier1, 1, 2, 3), Tier1, 3},
		{"tier2 upgrade movie", movieStream(2, Tier2), Tier2, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.stream.Tier != tc.wantTier {
				t.Errorf("expected tier %v, got %v", tc.wantTier, tc.stream.Tier)
			}
			if len(tc.stream.Items) != tc.wantItemCount {
				t.Errorf("expected %d items, got %d", tc.wantItemCount, len(tc.stream.Items))
			}
		})
	}
}

// 1.5 Episode-ascending ordering within a season stream, from out-of-order input.
func TestSortSeasonItems_OutOfOrderEpisodes(t *testing.T) {
	items := []Item{{Episode: 6}, {Episode: 3}, {Episode: 5}}
	SortSeasonItems(items)

	want := []int{3, 5, 6}
	for i, ep := range want {
		if items[i].Episode != ep {
			t.Errorf("index %d: expected episode %d, got %d", i, ep, items[i].Episode)
		}
	}
}

func TestSortSeasonItems_ResumedItemFirst(t *testing.T) {
	items := []Item{
		{Episode: 3},
		{Episode: 5, ResumeInfo: &models.DownloadInfo{}},
	}
	SortSeasonItems(items)

	if items[0].Episode != 5 {
		t.Fatalf("expected resumed episode 5 first, got episode %d", items[0].Episode)
	}
	if items[1].Episode != 3 {
		t.Fatalf("expected fresh episode 3 second, got episode %d", items[1].Episode)
	}
}

// 1.2 ClaimNext: empirical draw ratio and tier-1-only-when-tier-2-empty.
func TestClaimNext_TierWeightedDraw(t *testing.T) {
	const iterations = 20000
	const p = 0.1

	tier1Count, tier2Count := 0, 0
	for i := 0; i < iterations; i++ {
		streams := []*Stream{movieStream(1, Tier1), movieStream(2, Tier2)}
		s := NewScheduler(streams, p)
		stream, ok := s.ClaimNext()
		if !ok {
			t.Fatal("expected a claimable stream")
		}
		if stream.Tier == Tier1 {
			tier1Count++
		} else {
			tier2Count++
		}
		// put it back for the next iteration by rebuilding is simpler than
		// mutating pools directly; nothing else to do here.
	}

	ratio := float64(tier2Count) / float64(iterations)
	if ratio < p-0.03 || ratio > p+0.03 {
		t.Errorf("expected tier2 draw ratio near %.2f, got %.3f (%d/%d)", p, ratio, tier2Count, iterations)
	}
}

func TestClaimNext_Tier1AlwaysWhenTier2Empty(t *testing.T) {
	streams := []*Stream{movieStream(1, Tier1), movieStream(2, Tier1)}
	s := NewScheduler(streams, 0.5)

	for i := 0; i < 2; i++ {
		stream, ok := s.ClaimNext()
		if !ok {
			t.Fatal("expected a claimable stream")
		}
		if stream.Tier != Tier1 {
			t.Errorf("expected tier1, got %v", stream.Tier)
		}
	}

	if _, ok := s.ClaimNext(); ok {
		t.Error("expected no more claimable streams")
	}
}

// 1.3 Release: season N+1 not claimable until season N fully released.
func TestRelease_SeasonOrder(t *testing.T) {
	streams := []*Stream{
		seasonStream(1, 1, Tier1, 1, 2),
		seasonStream(1, 2, Tier1, 1, 2),
	}
	s := NewScheduler(streams, 0)

	if s.Remaining() != 1 {
		t.Fatalf("expected only season 1 claimable, got %d claimable", s.Remaining())
	}

	claimed, ok := s.ClaimNext()
	if !ok {
		t.Fatal("expected to claim season 1")
	}
	if claimed.Season != 1 {
		t.Fatalf("expected to claim season 1, got season %d", claimed.Season)
	}

	if _, ok := s.ClaimNext(); ok {
		t.Fatal("season 2 should not be claimable before season 1 is released")
	}

	s.Release(1)

	claimed, ok = s.ClaimNext()
	if !ok {
		t.Fatal("expected season 2 to become claimable after release")
	}
	if claimed.Season != 2 {
		t.Fatalf("expected season 2, got season %d", claimed.Season)
	}
}

// media-download-scheduling spec, "Random selection among claimable tier-1
// streams": selection is not source-ordered — movie streams and series-season
// streams are equally likely to be drawn first, not "all movies then all
// series" or vice versa.
func TestClaimNext_SelectionNotSourceOrdered(t *testing.T) {
	const trials = 2000
	movieFirst, seriesFirst := 0, 0

	for i := 0; i < trials; i++ {
		streams := []*Stream{movieStream(1, Tier1), seasonStream(1, 1, Tier1, 1)}
		s := NewScheduler(streams, 0)
		stream, ok := s.ClaimNext()
		if !ok {
			t.Fatal("expected a claimable stream")
		}
		if stream.SeriesID == 0 {
			movieFirst++
		} else {
			seriesFirst++
		}
	}

	ratio := float64(movieFirst) / float64(trials)
	if ratio < 0.4 || ratio > 0.6 {
		t.Errorf("expected roughly even split between movie-first and series-first draws, got movie=%d series=%d", movieFirst, seriesFirst)
	}
}

// 1.4 Concurrent-access: N workers hammering ClaimNext/drain/Release against a
// fixed stream set. No stream is double-claimed and every stream is drained
// exactly once.
func TestScheduler_ConcurrentClaimDrainRelease(t *testing.T) {
	const numSeries = 20
	const seasonsPerSeries = 3
	const numMovies = 30
	const numWorkers = 8

	var streams []*Stream
	for m := 0; m < numMovies; m++ {
		streams = append(streams, movieStream(m, Tier1))
	}
	for series := 1; series <= numSeries; series++ {
		for season := 1; season <= seasonsPerSeries; season++ {
			streams = append(streams, seasonStream(series, season, Tier1, 1, 2))
		}
	}

	s := NewScheduler(streams, 0.1)

	var mu sync.Mutex
	drainedCount := make(map[string]int)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				stream, ok := s.ClaimNext()
				if !ok {
					return
				}

				key := fmt.Sprintf("%s|season:%d", stream.SourceKey, stream.Season)
				mu.Lock()
				drainedCount[key]++
				mu.Unlock()

				if stream.SeriesID != 0 {
					s.Release(stream.SeriesID)
				}
			}
		}()
	}
	wg.Wait()

	if len(drainedCount) != len(streams) {
		t.Fatalf("expected %d distinct streams drained, got %d", len(streams), len(drainedCount))
	}
	for key, count := range drainedCount {
		if count != 1 {
			t.Errorf("stream %s claimed %d times, expected exactly 1 (double-claim)", key, count)
		}
	}
}
