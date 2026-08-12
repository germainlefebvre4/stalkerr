# Downloads Display UI

## Description

Enhanced frontend Downloads tab that displays enriched download information with TMDB metadata, file technical specifications, and visual indicators for quality issues. Prioritizes content titles over technical paths for better user experience.

## Requirements

### Functional Requirements

**GIVEN** the Downloads tab is active  
**WHEN** the component loads  
**THEN** the system SHALL:
- Fetch enriched downloads from GET /api/v1/downloads
- Display each download as a card with content title, technical specs, and status
- Auto-refresh every 5 seconds
- Show loading state during initial fetch

**GIVEN** a download with content metadata  
**WHEN** rendering the download card  
**THEN** the system SHALL display:
- Content title (large, bold) from content.title
- Year in parentheses from content.year
- Content type icon (🎬 for movies, 📺 for tvshows)
- Status badge (✅ Complété, ❌ Échec, ⏳ En cours)

**GIVEN** a download with file_info metadata  
**WHEN** rendering the download card  
**THEN** the system SHALL:
- Display the folder/file names inside a compact monospace box.
- Extract and display technical specifications (Format, Resolution, Size, and optional Duration in minutes with 🕒 icon) inline below the monospace box.
- Render validation status/indicators as structured, colored badge chips (Year validity: `✅ Année OK` or `⚠️ Année manquante`/`⚠️ Année incorrecte`, Format validity: `✅ Format OK` or `⚠️ Format inconnu`) next to the technical specs line.

**GIVEN** the download's detected resolution is 480p or 360p  
**WHEN** rendering the download card  
**THEN** the system SHALL display a warning chip `⚠️ Basse qualité (<resolution>)` using `badge badge-pending` next to the validation indicators.

**GIVEN** a download with genres  
**WHEN** rendering the download card  
**THEN** the system SHALL display genres with 🎭 icon

**GIVEN** a failed download  
**WHEN** rendering the download card  
**THEN** the system SHALL:
- Show error message in a highlighted error section
- Display retry count if > 0
- Show partial download progress if available

**GIVEN** a download with year_mismatch flag  
**WHEN** rendering the download card  
**THEN** the system SHALL show ⚠️ indicator with tooltip "Année dans le path différente de TMDB"

**GIVEN** filter dropdowns (status, type, problem)  
**WHEN** a filter is changed  
**THEN** the system SHALL:
- Update query parameters
- Re-fetch downloads with new filters
- Reset to page 1

### Requirement: Filter Persistence Across Refresh
The frontend SHALL reflect the Downloads tab's status, type, and problem filters in the browser URL's query string under the parameters `dlStatus`, `dlType`, and `dlProblem` respectively, and SHALL restore this exact filter selection from the URL on page load or refresh. These parameter names SHALL NOT be reused by any other tab's filters (e.g. the Playlist tab's `type` and `state` parameters), so that navigating between tabs or sharing a URL cannot cross-apply one tab's filter values to another.

#### Scenario: Restore Downloads filters from the URL after a refresh
- **WHEN** the user has selected the "Échoués" status filter and the "Séries" type filter on the Downloads tab, and refreshes the browser
- **THEN** the frontend SHALL read `dlStatus` and `dlType` from the URL query string and re-render the Downloads list with the same filters applied, instead of resetting to "Tous".

#### Scenario: Downloads and Playlist filters do not collide
- **WHEN** the user has an active Playlist content-type filter (`type=movies`) and switches to the Downloads tab and selects the "Séries" type filter
- **THEN** the frontend SHALL write the Downloads type filter to `dlType=tvshows` without altering or removing the Playlist's `type=movies` query parameter.

### Display Priority

1. **Title** (most prominent) - 1.25rem, bold
2. **Folder path** (secondary) - 0.85rem, code style
3. **Technical specs** (tertiary) - 0.8rem, inline
4. **Status/badges** (corner) - badge component
5. **Genres** (optional) - 0.8rem, muted

### Filter Requirements

**Status Filter**:
- Options: Tous, Complétés, En cours, Échoués
- Maps to: null, "completed", "downloading", "failed"

**Type Filter**:
- Options: Tous, Films, Séries
- Maps to: null, "movies", "tvshows"

**Problem Filter**:
- Options: Aucun, Année manquante, Année incorrecte, Format inconnu, Basse qualité
- Maps to: null, "missing_year", "year_mismatch", "unknown_format", "low_quality"

### Non-Functional Requirements

- TypeScript interface SHALL match backend DownloadEnrichedResponse exactly
- Component SHALL handle missing/null fields gracefully
- Filters SHALL use standard HTML select elements (no third-party dependencies)
- UI SHALL be responsive and work on mobile/tablet

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
