// Package scheduler groups Radarr- and Sonarr-originated downloadable content into
// atomic Streams and coordinates a shared worker pool that claims, drains, and
// releases them: one stream per movie, one stream per TV series' earliest
// incomplete season, with a small, configurable chance of drawing already-downloaded
// (upgrade-eligible) content instead of new content on any given draw.
package scheduler

import (
	"math/rand"
	"sync"
)

// DefaultForceTierProbability is used when the configured probability is out of range.
const DefaultForceTierProbability = 0.1

// Scheduler holds the mutable pool of claimable streams and the per-series queues
// of not-yet-claimable future seasons. All state mutation happens behind a single
// mutex with no I/O in the critical section, so it is safe to share across worker
// goroutines.
type Scheduler struct {
	mu          sync.Mutex
	probability float64
	rng         *rand.Rand

	tier1 []*Stream
	tier2 []*Stream

	// pendingSeasons holds, per series, the not-yet-claimable season streams in
	// ascending season order. The earliest season for a series is placed directly
	// into tier1/tier2 above at construction time; later seasons live here until
	// Release is called.
	pendingSeasons map[int][]*Stream
}

// NewScheduler builds a Scheduler from the full stream set produced by
// BuildStreams. For each series (SeriesID != 0), only the earliest season stream
// is made claimable; later seasons are queued and only exposed via Release once
// the earlier season is reported complete. Movie streams and Tier2 streams
// (SeriesID == 0) are claimable immediately.
func NewScheduler(streams []*Stream, probability float64) *Scheduler {
	if probability < 0 || probability > 1 {
		probability = DefaultForceTierProbability
	}

	s := &Scheduler{
		probability:    probability,
		rng:            rand.New(rand.NewSource(rand.Int63())),
		pendingSeasons: make(map[int][]*Stream),
	}

	bySeries := make(map[int][]*Stream)
	for _, stream := range streams {
		if stream.SeriesID != 0 {
			bySeries[stream.SeriesID] = append(bySeries[stream.SeriesID], stream)
			continue
		}
		s.addToPool(stream)
	}

	for seriesID, seasonStreams := range bySeries {
		sortStreamsBySeason(seasonStreams)
		s.addToPool(seasonStreams[0])
		if len(seasonStreams) > 1 {
			s.pendingSeasons[seriesID] = seasonStreams[1:]
		}
	}

	return s
}

func sortStreamsBySeason(streams []*Stream) {
	for i := 1; i < len(streams); i++ {
		for j := i; j > 0 && streams[j-1].Season > streams[j].Season; j-- {
			streams[j-1], streams[j] = streams[j], streams[j-1]
		}
	}
}

func (s *Scheduler) addToPool(stream *Stream) {
	if stream.Tier == Tier2 {
		s.tier2 = append(s.tier2, stream)
		return
	}
	s.tier1 = append(s.tier1, stream)
}

// ClaimNext draws the next stream to work on: with probability p it draws from
// tier 2 (whenever tier 2 has a claimable stream), otherwise it draws uniformly
// at random from tier 1. It returns (nil, false) once no stream remains claimable.
func (s *Scheduler) ClaimNext() (*Stream, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tier1) == 0 && len(s.tier2) == 0 {
		return nil, false
	}

	useTier2 := len(s.tier2) > 0 && (len(s.tier1) == 0 || s.rng.Float64() < s.probability)

	pool := &s.tier1
	if useTier2 {
		pool = &s.tier2
	}

	idx := s.rng.Intn(len(*pool))
	stream := (*pool)[idx]

	last := len(*pool) - 1
	(*pool)[idx] = (*pool)[last]
	*pool = (*pool)[:last]

	return stream, true
}

// Release is called once a series-season stream has been fully drained (every
// item reached a terminal state). It exposes that series' next-earliest season,
// if any, as claimable. It is a no-op for series with no further pending season.
func (s *Scheduler) Release(seriesID int) {
	if seriesID == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	queue := s.pendingSeasons[seriesID]
	if len(queue) == 0 {
		return
	}

	next := queue[0]
	if len(queue) > 1 {
		s.pendingSeasons[seriesID] = queue[1:]
	} else {
		delete(s.pendingSeasons, seriesID)
	}

	s.addToPool(next)
}

// Remaining returns the total number of streams currently claimable (not
// counting streams queued behind a pending season). Primarily useful for tests
// and dry-run reporting.
func (s *Scheduler) Remaining() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tier1) + len(s.tier2)
}
