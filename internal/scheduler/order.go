package scheduler

// SortSeasonItems orders a series-season stream's items so that ascending
// episode number is the primary key, with items carrying ResumeInfo (an
// existing, incomplete download) ordered ahead of not-yet-attempted items —
// see design.md decision 2 on merging incomplete downloads into their owning
// stream. Movie streams (single item) are unaffected.
func SortSeasonItems(items []Item) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && itemLess(items[j], items[j-1]); j-- {
			items[j-1], items[j] = items[j], items[j-1]
		}
	}
}

func itemLess(a, b Item) bool {
	aResumed := a.ResumeInfo != nil
	bResumed := b.ResumeInfo != nil
	if aResumed != bResumed {
		return aResumed
	}
	return a.Episode < b.Episode
}
