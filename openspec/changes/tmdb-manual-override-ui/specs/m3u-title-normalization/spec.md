## ADDED Requirements

### Requirement: Consulting persistent manual mappings during processing
Before executing any normal title extraction, normalization, or querying the external TMDB API, the playlist processor MUST query the local database's `manual_mappings` table using the item's original `TvgName` and `GroupTitle`.
If a matching record is found:
- The processor SHALL skip normal title parsing, normalization, and TMDB API search entirely.
- The processor SHALL use the mapped `ContentType` (movies or tvshows).
- The processor SHALL find or create the local `Movie` or `TVShow` GORM record using the stored `TMDBID` (and stored `Season`/`Episode` if it's a TV show) and link the `ProcessedLine` to it.

#### Scenario: Processor uses learned manual mapping for a movie VOD
- **WHEN** an IPTV line with TvgName `"FR: INCEPTION (2010) FHD"` and GroupTitle `"FR: FILMS ACTION"` is being processed, and there is a `ManualMapping` record in the database for this TvgName and GroupTitle pointing to Movie TMDB ID `27205`
- **THEN** the processor SHALL bypass the TMDB API, retrieve/create Movie `27205`, set the line's ContentType to `"movies"`, associate it with Movie `27205`, and successfully complete processing
