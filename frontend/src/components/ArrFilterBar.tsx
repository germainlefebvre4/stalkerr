import { useTranslation } from 'react-i18next';
import { MatchStatusFilter, SuiviFilter, FileFilter } from '../types';
import { MultiSelectDropdown, MultiSelectDropdownOption } from './MultiSelectDropdown';

interface ArrFilterBarProps {
  service: 'radarr' | 'sonarr';
  searchPlaceholder: string;
  searchValue: string;
  onSearchChange: (value: string) => void;
  matchStatus: MatchStatusFilter;
  onMatchStatusChange: (value: MatchStatusFilter) => void;
  suivi: SuiviFilter;
  onSuiviChange: (value: SuiviFilter) => void;
  suiviOptions: MultiSelectDropdownOption<'monitored' | 'unmonitored'>[];
  file: FileFilter;
  onFileChange: (value: FileFilter) => void;
  fileOptions: MultiSelectDropdownOption<'downloaded' | 'partial' | 'none'>[];
  missingActive: boolean;
  onMissing: () => void;
  isDefault: boolean;
  onReset: () => void;
}

// Filter bar shared by the Films and Séries sections: the search field on its
// own line, then labelled groups (Radarr/Sonarr, Playlist) and a reset button.
// Purely presentational - all state and URL handling stays in the caller.
export function ArrFilterBar({
  service, searchPlaceholder, searchValue, onSearchChange,
  matchStatus, onMatchStatusChange,
  suivi, onSuiviChange, suiviOptions, file, onFileChange, fileOptions,
  missingActive, onMissing, isDefault, onReset,
}: ArrFilterBarProps) {
  const { t } = useTranslation('radarrSonarr');
  const selectId = `arr-match-status-${service}`;

  return (
    <div className="arr-filter-bar">
      <input
        type="text"
        placeholder={searchPlaceholder}
        value={searchValue}
        onChange={e => onSearchChange(e.target.value)}
        className="custom-input arr-filter-search"
      />
      <div className="arr-filter-groups">
        <div className="arr-filter-group" role="group" aria-labelledby={`${selectId}-arr-caption`}>
          <span className="arr-filter-caption" id={`${selectId}-arr-caption`}>{t(`filterBar.group.${service}`)}</span>
          <div className="arr-filter-controls">
            <MultiSelectDropdown
              label={t('filterStatus.suivi.label')}
              value={suivi}
              onChange={onSuiviChange}
              options={suiviOptions}
              noneLabel={t('filterStatus.none')}
            />
            <MultiSelectDropdown
              label={t('filterStatus.fichier.label')}
              value={file}
              onChange={onFileChange}
              options={fileOptions}
              noneLabel={t('filterStatus.none')}
            />
            <button
              type="button"
              className={`btn-secondary arr-filter-control missing-preset${missingActive ? ' missing-preset--active' : ''}`}
              aria-pressed={missingActive}
              onClick={onMissing}
            >
              {t('filterStatus.missingPreset')}
            </button>
          </div>
        </div>
        <div className="arr-filter-group" role="group" aria-labelledby={`${selectId}-playlist-caption`}>
          <span className="arr-filter-caption" id={`${selectId}-playlist-caption`}>{t('filterBar.group.playlist')}</span>
          <div className="arr-filter-controls">
            <label className="arr-filter-select-label" htmlFor={selectId}>{t('filterBar.correspondance')}</label>
            <select
              id={selectId}
              className="arr-filter-control arr-filter-select"
              value={matchStatus}
              onChange={e => onMatchStatusChange(e.target.value as MatchStatusFilter)}
            >
              <option value="">{t('filterStatus.all')}</option>
              <option value="matched">{t('filterStatus.matched')}</option>
              <option value="no_match">{t('filterStatus.noMatch')}</option>
            </select>
          </div>
        </div>
        <button
          type="button"
          className="btn-secondary arr-filter-control arr-filter-reset"
          disabled={isDefault}
          onClick={onReset}
        >
          {t('filterBar.reset')}
        </button>
      </div>
    </div>
  );
}
