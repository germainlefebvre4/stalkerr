# Downloads Display UI

## Purpose

Enhanced frontend Downloads tab that displays enriched download information with TMDB metadata, file technical specifications, and visual indicators for quality issues. Prioritizes content titles over technical paths for better user experience.

## Requirements

### Requirement: Download List Summary Row
The Downloads tab SHALL fetch enriched downloads from `GET /api/v1/downloads` on initial tab activation and when the user clicks the manual refresh control, and SHALL NOT poll or auto-refresh the list in the background. The tab SHALL show a loading state during a fetch, same as before. Instead of a detailed card, each download SHALL render as a compact summary row: on viewports at or above the mobile breakpoint, a table row; below it, a list card matching the Playlist tab's `mobile-list-card` pattern. Each summary row SHALL display, at minimum: the content type icon (🎬/📺/🔗) and title (falling back to the file name or URL when no `content.title` is available) with the year in parentheses when known, the status badge, and a compact progress indicator (percentage and/or size) for downloads whose status is `downloading` or `retrying`. The summary row SHALL NOT render file paths, technical specification chips, validation badges, genres, or error messages inline — that detail is available in the sidepanel (capability `downloads-details-sidepanel`). Each summary row SHALL be clickable/tappable to open that sidepanel for the corresponding download.

On mobile (viewport narrower than the mobile breakpoint), the status badge SHALL show its emoji alone, with no accompanying word, for the `completed`, `pending`, `downloading`, and `failed` (no retries) statuses; each of these four statuses SHALL use a distinct emoji so that none of them can be confused for another without relying on text. A `failed` download that has been retried SHALL show its emoji followed by the retry count in parentheses (e.g. `❌ (3×)`), without the word "Échec"/"Failed". On desktop (viewport at or above the mobile breakpoint), the status badge SHALL show that same emoji followed by its status text label (e.g. `✅ Complété`, `❌ Échec`), and a retried `failed` download SHALL show the emoji, the "Échec"/"Failed" text, and the retry count together (e.g. `❌ Échec (3×)`). The `retrying` status badge is unchanged on both viewports (emoji plus its word label). Regardless of status or viewport, the status badge SHALL render on a single line and SHALL NOT wrap its content onto multiple lines at any viewport width, even when the download's title is long enough to compress the space available to the badge.

The status, type, and problem filter dropdowns SHALL keep updating query parameters, re-fetching downloads with the new filters, and resetting to page 1 on change, unchanged from before.

#### Scenario: Desktop table row shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport at or above the mobile breakpoint
- **THEN** each download SHALL appear as one table row showing its type icon, title, year, status badge, and (for `downloading`/`retrying` items) a compact progress indicator, with no inline file path, technical specs, validation badges, genres, or error message

#### Scenario: Mobile card shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport narrower than the mobile breakpoint
- **THEN** each download SHALL appear as one list card showing its type icon, title, year, and status badge, matching the visual pattern used by the Playlist tab's mobile list cards

#### Scenario: Missing content title falls back gracefully
- **WHEN** a download has no `content.title`
- **THEN** the summary row SHALL display the download's file name (derived from `download_path`) or, failing that, its `url`, instead of leaving the title blank

#### Scenario: Clicking a row opens the sidepanel
- **WHEN** the user clicks (or taps) a download's summary row
- **THEN** the frontend SHALL open the details sidepanel for that download (capability `downloads-details-sidepanel`)

#### Scenario: Manual refresh re-fetches the current page
- **WHEN** the user clicks the refresh control
- **THEN** the frontend SHALL re-fetch the current page of downloads with the active filters and `limit`/`offset`, without relying on any background timer

#### Scenario: No background auto-refresh
- **WHEN** the Downloads tab is active and the user takes no action
- **THEN** the frontend SHALL NOT issue any additional fetch to `/api/v1/downloads` after the initial load, until the user clicks refresh, changes a filter, changes the page, or changes the items-per-page value

#### Scenario: Completed, pending, downloading, and no-retry failed badges show emoji only
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is `completed`, `pending`, `downloading`, or `failed` with a retry count of zero
- **THEN** its status badge SHALL display only that status's emoji, with no text label, and that emoji SHALL be different from the emoji used for each of the other three statuses

#### Scenario: Completed, pending, downloading, and no-retry failed badges show emoji and text on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and a download's status is `completed`, `pending`, `downloading`, or `failed` with a retry count of zero
- **THEN** its status badge SHALL display that status's emoji followed by its text label (e.g. `✅ Complété`, `❌ Échec`)

#### Scenario: Failed badge with retries shows the retry count instead of the status word
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is `failed` and its retry count is greater than zero (e.g. 3)
- **THEN** its status badge SHALL display the failed emoji followed by the retry count in parentheses (e.g. `❌ (3×)`), and SHALL NOT display the word "Échec"/"Failed"

#### Scenario: Failed badge with retries shows the status word and retry count on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and a download's status is `failed` and its retry count is greater than zero (e.g. 3)
- **THEN** its status badge SHALL display the failed emoji, the "Échec"/"Failed" text, and the retry count together (e.g. `❌ Échec (3×)`)

#### Scenario: Status badge never wraps even with a long title
- **WHEN** a download's title is long enough to compress the horizontal space left for the status badge, on any viewport width
- **THEN** the status badge SHALL remain on a single line, with its emoji (and text, when present) never breaking onto a second line

### Requirement: Pagination Controls
The Downloads tab SHALL paginate the download list using the `limit`, `offset`, `total`, and `total_pages` fields returned by `GET /api/v1/downloads`, and SHALL render page navigation controls plus an items-per-page selector below the list.

#### Scenario: Navigating to another page
- **WHEN** the user navigates to a different page via the pagination controls
- **THEN** the frontend SHALL re-fetch downloads with the `offset` corresponding to that page for the current `limit`, using the active filters, and update the displayed list

#### Scenario: Changing items per page
- **WHEN** the user selects a different items-per-page value
- **THEN** the frontend SHALL re-fetch downloads with the new `limit`, resetting to page 1, using the active filters

### Requirement: Filter Persistence Across Refresh
The frontend SHALL reflect the Downloads tab's status, type, and problem filters in the browser URL's query string under the parameters `dlStatus`, `dlType`, and `dlProblem` respectively, and SHALL restore this exact filter selection from the URL on page load or refresh. These parameter names SHALL NOT be reused by any other tab's filters (e.g. the Playlist tab's `type` and `state` parameters), so that navigating between tabs or sharing a URL cannot cross-apply one tab's filter values to another.

#### Scenario: Restore Downloads filters from the URL after a refresh
- **WHEN** the user has selected the "Échoués" status filter and the "Séries" type filter on the Downloads tab, and refreshes the browser
- **THEN** the frontend SHALL read `dlStatus` and `dlType` from the URL query string and re-render the Downloads list with the same filters applied, instead of resetting to "Tous".

#### Scenario: Downloads and Playlist filters do not collide
- **WHEN** the user has an active Playlist content-type filter (`type=movies`) and switches to the Downloads tab and selects the "Séries" type filter
- **THEN** the frontend SHALL write the Downloads type filter to `dlType=tvshows` without altering or removing the Playlist's `type=movies` query parameter.

### Requirement: Downloads Filter Option Sets
The Downloads tab SHALL offer three filter controls — status, type, and problem — each with a fixed set of options mapping to specific backend query values: the status filter maps "Tous" / "Complétés" / "En cours" / "Échoués" to no filter / `completed` / `downloading` / `failed`; the type filter maps "Tous" / "Films" / "Séries" to no filter / `movies` / `tvshows`; the problem filter maps "Aucun" / "Année manquante" / "Année incorrecte" / "Format inconnu" / "Basse qualité" to no filter / `missing_year` / `year_mismatch` / `unknown_format` / `low_quality`.

#### Scenario: Selecting a status filter option applies the mapped backend value
- **WHEN** the user selects "Échoués" in the status filter
- **THEN** the frontend SHALL query downloads with `status=failed`

#### Scenario: Selecting a problem filter option applies the mapped backend value
- **WHEN** the user selects "Année manquante" in the problem filter
- **THEN** the frontend SHALL query downloads with `problem=missing_year`

#### Scenario: "Tous" resets a filter to no constraint
- **WHEN** the user selects "Tous" in the status or type filter
- **THEN** the frontend SHALL query downloads without that filter parameter

### Requirement: Graceful Handling of Missing Enrichment Fields
The Downloads list SHALL render correctly when optional enrichment fields (`content`, `file_info`, and their nested fields) are absent from a download record, without errors or broken layout.

#### Scenario: Download with no content metadata renders without a title
- **WHEN** a download has no `content` object (e.g. an orphan download with no matched Movie/TVShow)
- **THEN** the row/card SHALL render its other fields normally and fall back gracefully in place of the missing title

#### Scenario: Download with no file_info renders without file details
- **WHEN** a download has a null `download_path` and consequently no `file_info`
- **THEN** the row/card SHALL render without file/format details, without erroring

## TypeScript Interfaces

```typescript
interface DownloadEnriched {
  id: number;
  url: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying';
  download_path?: string;
  file_size?: number;
  bytes_downloaded?: number;
  total_bytes?: number;
  retry_count: number;
  error_message?: string;
  updated_at: string;
  
  content?: {
    type: 'movies' | 'tvshows' | 'channels' | 'uncategorized';
    title: string;
    year?: number;
    resolution?: string;
    season?: number;
    episode?: number;
    genres?: string;
    duration?: number;
  };
  
  file_info?: {
    extension: string;
    folder_name: string;
    file_name: string;
    has_year_in_path: boolean;
    year_mismatch: boolean;
    detected_year?: number;
    detected_resolution?: string;
    is_valid_format: boolean;
  };
}
```

## UI Mockup (Implementation Reference)

### Successful Download Card

```
┌────────────────────────────────────────────────────────────┐
│ 🎬 La Cité de Dieu (2002)                     ✅ Complété  │
├────────────────────────────────────────────────────────────┤
│ 📂 Dossier: La.Cite.de.Dieu.2002.1080p.BluRay/            │
│    └─ video.mkv                                            │
│                                                             │
│ 📹 Format: MKV • 1080p • 4.2 GB • Année ✓                 │
│ 🎭 Genres: Crime, Drama                                    │
└────────────────────────────────────────────────────────────┘
```

### Failed Download Card

```
┌────────────────────────────────────────────────────────────┐
│ 📺 Breaking Bad S05E14 (2013)             ❌ Échec (3×)   │
├────────────────────────────────────────────────────────────┤
│ 🔴 Erreur: HTTP 403 Forbidden - Link expired              │
│                                                             │
│ 📂 Téléchargement partiel:                                 │
│    Breaking.Bad.S05E14.1080p.incomplete                   │
│    1.2 GB / 1.8 GB (66%)                                   │
└────────────────────────────────────────────────────────────┘
```

### Download with Issues

```
┌────────────────────────────────────────────────────────────┐
│ 🎬 Avatar                                  ⚠️ Complété     │
├────────────────────────────────────────────────────────────┤
│ 📂 Dossier: Avatar.BluRay.1080p/                           │
│    └─ avatar.mkv                                            │
│                                                             │
│ 📹 Format: MKV • 1080p • 6.8 GB • ⚠️ Année manquante       │
│ 🎭 Genres: Action, Adventure, Fantasy                      │
└────────────────────────────────────────────────────────────┘
```

## Implementation Details

**Implementation Notes**:
- The TypeScript interface SHALL match backend `DownloadEnrichedResponse` exactly.
- Filters SHALL use standard HTML select elements (no third-party dependencies).

**File Location**: `frontend/src/App.tsx`

**State Updates**:
```typescript
// Replace DownloadInfo with DownloadEnriched
const [downloads, setDownloads] = useState<DownloadEnriched[]>([]);

// Add filter states
const [statusFilter, setStatusFilter] = useState<string>('');
const [typeFilter, setTypeFilter] = useState<string>('');
const [problemFilter, setProblemFilter] = useState<string>('');
```

**Fetch Function**:
```typescript
const fetchDownloads = () => {
  setDownloadsLoading(true);
  
  let url = '/api/v1/downloads?limit=20';
  if (statusFilter) url += `&status=${statusFilter}`;
  if (typeFilter) url += `&type=${typeFilter}`;
  if (problemFilter) url += `&problem=${problemFilter}`;
  
  fetch(url)
    .then(res => res.json())
    .then((data: PaginatedResponse<DownloadEnriched>) => {
      setDownloads(data.data || []);
    })
    .finally(() => setDownloadsLoading(false));
};
```

**Card Rendering**:
```tsx
{downloads.map(item => {
  const title = item.content?.title || 'Inconnu';
  const year = item.content?.year ? `(${item.content.year})` : '';
  const icon = item.content?.type === 'movies' ? '🎬' : '📺';
  const hasYearIssue = item.file_info && !item.file_info.has_year_in_path;
  const hasYearMismatch = item.file_info?.year_mismatch;
  const isCompleted = item.status === 'completed';
  
  return (
    <div key={item.id} className="download-card">
      <div className="download-header">
        <h3>{icon} {title} {year}</h3>
        <StatusBadge status={item.status} />
      </div>
      
      {item.file_info && (
        <div className="file-info">
          <div>📂 {item.file_info.folder_name}/</div>
          <div>   └─ {item.file_info.file_name}</div>
          <div>
            📹 {item.file_info.extension.toUpperCase()} • 
            {item.file_info.detected_resolution || 'Unknown'} • 
            {formatFileSize(item.file_size)} • 
            {hasYearIssue ? '⚠️ Année manquante' : 
             hasYearMismatch ? '⚠️ Année incorrecte' : 
             'Année ✓'}
          </div>
        </div>
      )}
      
      {item.content?.genres && (
        <div>🎭 {item.content.genres}</div>
      )}
      
      {item.error_message && (
        <div className="error-message">
          🔴 {item.error_message}
        </div>
      )}
    </div>
  );
})}
```

**Filter UI**:
```tsx
<div className="filters">
  <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
    <option value="">Statut: Tous</option>
    <option value="completed">Complétés</option>
    <option value="downloading">En cours</option>
    <option value="failed">Échoués</option>
  </select>
  
  <select value={typeFilter} onChange={e => setTypeFilter(e.target.value)}>
    <option value="">Type: Tous</option>
    <option value="movies">Films</option>
    <option value="tvshows">Séries</option>
  </select>
  
  <select value={problemFilter} onChange={e => setProblemFilter(e.target.value)}>
    <option value="">Problèmes: Aucun</option>
    <option value="missing_year">Année manquante</option>
    <option value="year_mismatch">Année incorrecte</option>
    <option value="unknown_format">Format inconnu</option>
    <option value="low_quality">Basse qualité</option>
  </select>
</div>
```

## Styling Additions

**New CSS Classes** (in `index.css`):
```css
.download-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  background: #fff;
  box-shadow: var(--shadow-sm);
  margin-bottom: 1rem;
}

.download-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.75rem;
}

.download-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--primary-slate);
  margin: 0;
}

.file-info {
  background: var(--bg-app);
  padding: 0.75rem;
  border-radius: var(--radius-sm);
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
  line-height: 1.6;
}

.error-message {
  padding: 0.75rem;
  background: var(--status-failed-bg);
  color: var(--status-failed-text);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  margin-top: 0.5rem;
}

.filters {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.filters select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  background: #fff;
}
```

## Testing

**Manual Tests**:
1. Navigate to Downloads tab
2. Verify downloads load with enriched data
3. Test status filter (completed, failed, downloading)
4. Test type filter (movies, tvshows)
5. Test problem filter (missing_year, etc.)
6. Verify auto-refresh every 5 seconds
7. Check responsive behavior on mobile
8. Verify graceful degradation when content/file_info is null

**Visual Validation**:
- Title is most prominent element
- Icons help identify content type quickly
- Error messages are clearly visible
- Problem indicators (⚠️) stand out
- File path is readable but secondary

## Edge Cases

**No Content Metadata**:
- Display URL as fallback title
- Show "Type: Inconnu"

**No File Info**:
- Skip file info section
- Display URL instead

**Partial Download**:
- Show progress bar
- Display bytes_downloaded / total_bytes

**Missing Year**:
- Show ⚠️ Année manquante

**Year Mismatch**:
- Show ⚠️ Année incorrecte with tooltip

## Accessibility

- Use semantic HTML (h3 for titles, proper button elements)
- Ensure color contrast meets WCAG AA standards
- Provide text alternatives for emoji icons (aria-label)
- Keyboard navigation support for filters
