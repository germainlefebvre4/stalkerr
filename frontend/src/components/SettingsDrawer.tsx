import { useState, ReactNode } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { Theme } from '../hooks/useTheme';
import { SystemStatusSection } from './SystemStatusSection';
import { FiltersSection } from './FiltersSection';
import { FilterConfig } from '../types';

const LANGUAGES: { code: 'en' | 'fr'; labelKey: string }[] = [
  { code: 'en', labelKey: 'language.en' },
  { code: 'fr', labelKey: 'language.fr' },
];

const PLAYLIST_LIMIT_OPTIONS = [10, 50, 100];

const REPOSITORY_URL = 'https://github.com/germainlefebvre4/Stalkeer';

interface DisclosureSectionProps {
  title: string;
  isExpanded: boolean;
  onToggle: () => void;
  children: ReactNode;
}

function DisclosureSection({ title, isExpanded, onToggle, children }: DisclosureSectionProps) {
  return (
    <div className="settings-section">
      <button
        type="button"
        className="advanced-filters-toggle"
        onClick={onToggle}
        aria-expanded={isExpanded}
      >
        <span className="advanced-filters-toggle-label">{title}</span>
        <span className={`advanced-filters-chevron${isExpanded ? ' is-open' : ''}`}>▾</span>
      </button>
      {isExpanded && <div style={{ paddingTop: '0.5rem' }}>{children}</div>}
    </div>
  );
}

interface ToggleSwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  iconOff?: string;
  iconOn?: string;
  ariaLabel: string;
  title?: string;
}

function ToggleSwitch({ checked, onChange, iconOff, iconOn, ariaLabel, title }: ToggleSwitchProps) {
  return (
    <label className="view-toggle-switch" title={title} aria-label={ariaLabel}>
      {iconOff && (
        <span className={`view-toggle-switch-icon${!checked ? ' is-active' : ''}`} aria-hidden="true">{iconOff}</span>
      )}
      <input
        type="checkbox"
        className="view-toggle-switch-input"
        checked={checked}
        onChange={e => onChange(e.target.checked)}
      />
      <span className="view-toggle-switch-track">
        <span className="view-toggle-switch-knob" />
      </span>
      {iconOn && (
        <span className={`view-toggle-switch-icon${checked ? ' is-active' : ''}`} aria-hidden="true">{iconOn}</span>
      )}
    </label>
  );
}

export interface SettingsDrawerProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;

  theme: Theme;
  onSetTheme: (theme: Theme) => void;
  reduceMotion: boolean;
  onSetReduceMotion: (value: boolean) => void;

  startupTab: string;
  onSetStartupTab: (tab: string) => void;
  tabOptions: { value: string; label: string }[];

  playlistLimit: number;
  onSetPlaylistLimit: (limit: number) => void;
  playlistView: 'items' | 'grouped';
  onSetPlaylistView: (view: 'items' | 'grouped') => void;

  filters: FilterConfig[];
  filtersLoading: boolean;
  onFetchFilters: () => void;
  onDeleteFilter: (id: number) => void;
  onOpenCreateFilter: () => void;
}

export function SettingsDrawer({
  isOpen,
  onOpenChange,
  theme,
  onSetTheme,
  reduceMotion,
  onSetReduceMotion,
  startupTab,
  onSetStartupTab,
  tabOptions,
  playlistLimit,
  onSetPlaylistLimit,
  playlistView,
  onSetPlaylistView,
  filters,
  filtersLoading,
  onFetchFilters,
  onDeleteFilter,
  onOpenCreateFilter,
}: SettingsDrawerProps) {
  const { t, i18n } = useTranslation('settings');
  const activeLanguage = i18n.language.startsWith('fr') ? 'fr' : 'en';

  const [isFiltresExpanded, setIsFiltresExpanded] = useState(false);
  const [isSystemeExpanded, setIsSystemeExpanded] = useState(false);

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="drawer-overlay" />
        <Dialog.Content className="settings-drawer-content">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0 }}>
              {t('title')}
            </Dialog.Title>
            <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
              {t('close')}
            </Dialog.Close>
          </div>
          <Dialog.Description style={{ display: 'none' }}>
            {t('description')}
          </Dialog.Description>

          {/* Apparence */}
          <section className="settings-section">
            <h3 className="settings-section-header">{t('sections.appearance')}</h3>

            <div className="settings-field settings-toggle-row">
              <div>
                <span className="settings-field-label" style={{ marginBottom: 0 }}>{t('appearance.themeLabel')}</span>
              </div>
              <ToggleSwitch
                checked={theme === 'dark'}
                onChange={checked => onSetTheme(checked ? 'dark' : 'light')}
                iconOff="☀️"
                iconOn="🌙"
                ariaLabel={t('appearance.themeLabel')}
                title={theme === 'dark' ? t('appearance.switchToLight') : t('appearance.switchToDark')}
              />
            </div>

            <div className="settings-field settings-toggle-row">
              <div>
                <span className="settings-field-label" style={{ marginBottom: 0 }}>{t('appearance.reduceMotionLabel')}</span>
              </div>
              <ToggleSwitch
                checked={reduceMotion}
                onChange={onSetReduceMotion}
                ariaLabel={t('appearance.reduceMotionLabel')}
              />
            </div>
          </section>

          {/* Langue */}
          <section className="settings-section">
            <h3 className="settings-section-header">{t('sections.language')}</h3>
            <div className="settings-field">
              <label className="settings-field-label" htmlFor="settings-language-select">{t('language.label')}</label>
              <select
                id="settings-language-select"
                className="custom-select"
                value={activeLanguage}
                onChange={e => i18n.changeLanguage(e.target.value)}
              >
                {LANGUAGES.map(({ code, labelKey }) => (
                  <option key={code} value={code}>{t(labelKey)}</option>
                ))}
              </select>
            </div>
          </section>

          {/* Filtres */}
          <DisclosureSection
            title={t('sections.filters')}
            isExpanded={isFiltresExpanded}
            onToggle={() => setIsFiltresExpanded(open => !open)}
          >
            <FiltersSection
              isExpanded={isFiltresExpanded}
              filters={filters}
              filtersLoading={filtersLoading}
              onFetchFilters={onFetchFilters}
              onDeleteFilter={onDeleteFilter}
              onOpenCreate={onOpenCreateFilter}
            />
          </DisclosureSection>

          {/* Préférences */}
          <section className="settings-section">
            <h3 className="settings-section-header">{t('sections.preferences')}</h3>

            <div className="settings-field">
              <label className="settings-field-label" htmlFor="settings-startup-tab-select">{t('preferences.startupTabLabel')}</label>
              <select
                id="settings-startup-tab-select"
                className="custom-select"
                value={startupTab}
                onChange={e => onSetStartupTab(e.target.value)}
              >
                {tabOptions.map(({ value, label }) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
              <p className="settings-field-hint">{t('preferences.startupTabHint')}</p>
            </div>

            <div className="settings-field">
              <label className="settings-field-label" htmlFor="settings-playlist-limit-select">{t('preferences.playlistPageSizeLabel')}</label>
              <select
                id="settings-playlist-limit-select"
                className="custom-select"
                value={playlistLimit}
                onChange={e => onSetPlaylistLimit(Number(e.target.value))}
              >
                {PLAYLIST_LIMIT_OPTIONS.map(limit => (
                  <option key={limit} value={limit}>{limit}</option>
                ))}
              </select>
            </div>

            <div className="settings-field settings-toggle-row">
              <span className="settings-field-label" style={{ marginBottom: 0 }}>{t('preferences.playlistViewLabel')}</span>
              <ToggleSwitch
                checked={playlistView === 'grouped'}
                onChange={checked => onSetPlaylistView(checked ? 'grouped' : 'items')}
                iconOff="☰"
                iconOn="🎬"
                ariaLabel={t('preferences.playlistViewLabel')}
                title={playlistView === 'grouped' ? t('preferences.switchToItems') : t('preferences.switchToGrouped')}
              />
            </div>
          </section>

          {/* Système */}
          <DisclosureSection
            title={t('sections.system')}
            isExpanded={isSystemeExpanded}
            onToggle={() => setIsSystemeExpanded(open => !open)}
          >
            <SystemStatusSection isExpanded={isSystemeExpanded} />
          </DisclosureSection>

          {/* À propos */}
          <section className="settings-section">
            <h3 className="settings-section-header">{t('sections.about')}</h3>
            <div className="settings-field">
              <div style={{ fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.35rem' }}>
                {t('about.appName')}
              </div>
              <a
                href={REPOSITORY_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="btn-secondary"
                style={{ display: 'inline-flex', textDecoration: 'none', fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
              >
                {t('about.repositoryLink')}
              </a>
            </div>
          </section>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
