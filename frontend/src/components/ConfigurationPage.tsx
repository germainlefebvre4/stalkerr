import { useEffect, useState } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Theme } from '../hooks/useTheme';
import { useIsMobile } from '../hooks/useMediaQuery';
import { useAppSettings } from '../hooks/useAppSettings';
import { useM3uSources } from '../hooks/useM3uSources';
import { useURLState, URLStateSchema } from '../hooks/useURLState';
import { ToggleSwitch } from './ToggleSwitch';
import { SystemStatusSection } from './SystemStatusSection';
import { FiltersSection } from './FiltersSection';
import { IntegrationsSection } from './IntegrationsSection';
import { NotificationsSection } from './NotificationsSection';
import { AdvancedSection } from './AdvancedSection';
import { M3uSourcesSection } from './M3uSourcesSection';
import { FilterConfig, FilterOriginEntry } from '../types';
import { INTEGRATIONS_SETTINGS_KEYS, NOTIFICATIONS_SETTINGS_KEYS, ADVANCED_SETTINGS_KEYS } from '../settingsFieldGroups';

const LANGUAGES: { code: 'en' | 'fr'; labelKey: string }[] = [
  { code: 'en', labelKey: 'language.en' },
  { code: 'fr', labelKey: 'language.fr' },
];

const PLAYLIST_LIMIT_OPTIONS = [10, 50, 100];

const REPOSITORY_URL = 'https://github.com/germainlefebvre4/Stalkeer';

function GithubIcon({ size = 14 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      focusable="false"
    >
      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
    </svg>
  );
}

const SETTINGS_TAB_IDS = ['general', 'integrations', 'content', 'notifications', 'advanced', 'system'] as const;
type SettingsTabId = typeof SETTINGS_TAB_IDS[number];

const SETTINGS_TAB_STORAGE_KEY = 'stalkeer_configuration_tab';

function isSettingsTabId(value: string | null): value is SettingsTabId {
  return !!value && (SETTINGS_TAB_IDS as readonly string[]).includes(value);
}

// Precedence: URL `settingsTab` param > `stalkeer_configuration_tab`
// localStorage > "general". See the Configuration Page Container requirement.
function readInitialSettingsTab(): SettingsTabId {
  const urlTab = new URLSearchParams(window.location.search).get('settingsTab');
  if (isSettingsTabId(urlTab)) return urlTab;

  const storedTab = localStorage.getItem(SETTINGS_TAB_STORAGE_KEY);
  if (isSettingsTabId(storedTab)) return storedTab;

  return 'general';
}

const SETTINGS_TAB_URL_SCHEMA = {
  settingsTab: {
    default: 'general' as SettingsTabId,
    parse: (raw: string) => raw as SettingsTabId,
    serialize: (v: SettingsTabId) => v,
    isValid: isSettingsTabId,
  },
} satisfies URLStateSchema;

export interface ConfigurationPageProps {
  onBack: () => void;

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
  filterOrigin: FilterOriginEntry[];
  filtersLoading: boolean;
  onFetchFilters: () => void;
  onDeleteFilter: (id: number) => void;
  onOpenCreateFilter: () => void;
}

// A dedicated Configuration page, reachable from the header's settings icon,
// distinct from the Home/Downloads/Playlist/Logs tabs, organized into six
// tabs (Général, Intégrations, Contenu, Notifications, Avancé, Système). See
// frontend-configuration-page and frontend-app-settings-management.
export function ConfigurationPage({
  onBack,
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
  filterOrigin,
  filtersLoading,
  onFetchFilters,
  onDeleteFilter,
  onOpenCreateFilter,
}: ConfigurationPageProps) {
  const { t, i18n } = useTranslation('settings');
  const activeLanguage = i18n.language.startsWith('fr') ? 'fr' : 'en';
  const isMobile = useIsMobile();

  const [activeSettingsTab, setActiveSettingsTabState] = useState<SettingsTabId>(readInitialSettingsTab);
  const [, patchSettingsTabURL] = useURLState(SETTINGS_TAB_URL_SCHEMA);
  const [searchQuery, setSearchQuery] = useState('');

  const setActiveSettingsTab = (tabId: SettingsTabId) => {
    setActiveSettingsTabState(tabId);
    localStorage.setItem(SETTINGS_TAB_STORAGE_KEY, tabId);
    patchSettingsTabURL({ settingsTab: tabId });
  };

  const { settings, loading: settingsLoading, fetchSettings, setSetting, clearSetting } = useAppSettings();
  const {
    sources, originNames, loading: sourcesLoading, fetchSources, createSource, updateSource, deleteSource,
  } = useM3uSources();

  // Prefetch settings, M3U sources (+ origin), and filters (+ origin) once
  // on mount - independently of which tab is initially active and of
  // useSystemStatus (which stays fetch-on-tab-activation-only) - so the
  // summary banner and per-tab badges below are accurate as soon as the
  // page opens. See the "Prefetch enough to make the summary banner true on
  // open" design decision.
  useEffect(() => {
    fetchSettings();
    fetchSources();
    onFetchFilters();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const isOverridden = (key: string) => settings.some(f => f.key === key && f.origin === 'interface');
  const isRestartRequiredOverride = (key: string) => settings.some(f => f.key === key && f.origin === 'interface' && f.restart_required);

  const integrationsOverrideCount = INTEGRATIONS_SETTINGS_KEYS.filter(isOverridden).length;
  const integrationsRestartRequired = INTEGRATIONS_SETTINGS_KEYS.some(isRestartRequiredOverride);

  const notificationsOverrideCount = NOTIFICATIONS_SETTINGS_KEYS.filter(isOverridden).length;
  const notificationsRestartRequired = NOTIFICATIONS_SETTINGS_KEYS.some(isRestartRequiredOverride);

  const advancedOverrideCount = ADVANCED_SETTINGS_KEYS.filter(isOverridden).length;
  const advancedRestartRequired = ADVANCED_SETTINGS_KEYS.some(isRestartRequiredOverride);

  const m3uOverrideCount = sources.filter(s => s.is_runtime).length;
  const filtersOverrideCount = filters.length;
  const contentOverrideCount = m3uOverrideCount + filtersOverrideCount;

  const settingsOverrideCount = settings.filter(f => f.origin === 'interface').length;
  const totalOverrideCount = settingsOverrideCount + contentOverrideCount;
  const restartRequiredCount = settings.filter(f => f.origin === 'interface' && f.restart_required).length;

  const query = searchQuery.trim().toLowerCase();
  const matchesQuery = (label: string) => !query || label.toLowerCase().includes(query);

  const appearanceFields = [
    {
      label: t('appearance.themeLabel'),
      node: (
        <div className="settings-field settings-toggle-row" key="theme">
          <span className="settings-field-label" style={{ marginBottom: 0 }}>{t('appearance.themeLabel')}</span>
          <ToggleSwitch
            checked={theme === 'dark'}
            onChange={checked => onSetTheme(checked ? 'dark' : 'light')}
            iconOff="☀️"
            iconOn="🌙"
            ariaLabel={t('appearance.themeLabel')}
            title={theme === 'dark' ? t('appearance.switchToLight') : t('appearance.switchToDark')}
          />
        </div>
      ),
    },
    {
      label: t('appearance.reduceMotionLabel'),
      node: (
        <div className="settings-field settings-toggle-row" key="reduceMotion">
          <span className="settings-field-label" style={{ marginBottom: 0 }}>{t('appearance.reduceMotionLabel')}</span>
          <ToggleSwitch
            checked={reduceMotion}
            onChange={onSetReduceMotion}
            ariaLabel={t('appearance.reduceMotionLabel')}
          />
        </div>
      ),
    },
  ].filter(f => matchesQuery(f.label));

  const languageFields = [
    {
      label: t('language.label'),
      node: (
        <div className="settings-field settings-field-full" key="language">
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
      ),
    },
  ].filter(f => matchesQuery(f.label));

  const preferencesFields = [
    {
      label: t('preferences.startupTabLabel'),
      node: (
        <div className="settings-field settings-field-full" key="startupTab">
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
      ),
    },
    {
      label: t('preferences.playlistPageSizeLabel'),
      node: (
        <div className="settings-field" key="playlistLimit">
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
      ),
    },
    {
      label: t('preferences.playlistViewLabel'),
      node: (
        <div className="settings-field settings-toggle-row" key="playlistView">
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
      ),
    },
  ].filter(f => matchesQuery(f.label));

  const renderTabBadge = (count: number, restartRequired: boolean) => (
    <>
      {count > 0 && <span className="advanced-filters-count-badge settings-tab-badge">{count}</span>}
      {restartRequired && <span className="settings-tab-restart-dot" title={t('restartRequired.label')} />}
    </>
  );

  // Mirrors renderTabBadge's "no badge when zero / no marker when not
  // restart-required" logic, serialized to text since a native <option>
  // can't render a colored badge. See the "Option label encodes override
  // count and restart marker as text" design decision.
  const buildTabOptionLabel = (label: string, count: number, restartRequired: boolean) =>
    `${restartRequired ? '⚠ ' : ''}${label}${count > 0 ? ` (${count})` : ''}`;

  const settingsTabs: { id: SettingsTabId; label: string; count: number; restartRequired: boolean }[] = [
    { id: 'general', label: t('tabs.general'), count: 0, restartRequired: false },
    { id: 'integrations', label: t('tabs.integrations'), count: integrationsOverrideCount, restartRequired: integrationsRestartRequired },
    { id: 'content', label: t('tabs.content'), count: contentOverrideCount, restartRequired: false },
    { id: 'notifications', label: t('tabs.notifications'), count: notificationsOverrideCount, restartRequired: notificationsRestartRequired },
    { id: 'advanced', label: t('tabs.advanced'), count: advancedOverrideCount, restartRequired: advancedRestartRequired },
    { id: 'system', label: t('tabs.system'), count: 0, restartRequired: false },
  ];

  return (
    <div className="configuration-page">
      <div className="configuration-page-content">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0 }}>
            {t('title')}
          </h2>
          <button className="btn-secondary" onClick={onBack} style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
            {t('close')}
          </button>
        </div>

        <div className="configuration-summary-banner">
          <span className="configuration-summary-banner-text">
            {totalOverrideCount === 0
              ? t('overview.noOverrides')
              : t('overview.totalOverrides', { count: totalOverrideCount })}
          </span>
          {restartRequiredCount > 0 && (
            <div className="configuration-summary-banner-counts">
              <span className="badge badge-warning">{t('overview.restartRequired', { count: restartRequiredCount })}</span>
            </div>
          )}
        </div>

        <div className="settings-search-input-wrap">
          <Search size={16} className="settings-search-icon" aria-hidden="true" />
          <input
            type="text"
            className="custom-input settings-search-input"
            placeholder={t('search.placeholder')}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            aria-label={t('search.placeholder')}
          />
        </div>

        <Tabs.Root value={activeSettingsTab} onValueChange={v => setActiveSettingsTab(v as SettingsTabId)}>
          {isMobile ? (
            <select
              className="custom-select settings-tabs-select"
              aria-label={t('tabs.selectLabel')}
              value={activeSettingsTab}
              onChange={e => setActiveSettingsTab(e.target.value as SettingsTabId)}
            >
              {settingsTabs.map(({ id, label, count, restartRequired }) => (
                <option key={id} value={id}>{buildTabOptionLabel(label, count, restartRequired)}</option>
              ))}
            </select>
          ) : (
            <Tabs.List className="settings-tabs-list">
              {settingsTabs.map(({ id, label, count, restartRequired }) => (
                <Tabs.Trigger key={id} value={id} className="segmented-tabs-trigger">
                  {label}
                  {renderTabBadge(count, restartRequired)}
                </Tabs.Trigger>
              ))}
            </Tabs.List>
          )}

          <Tabs.Content value="general" className="settings-tab-panel">
            <div className="settings-cards-grid">
              {appearanceFields.length > 0 && (
                <div className="settings-group-card settings-group-card--compact">
                  <h4 className="settings-group-card-title">{t('sections.appearance')}</h4>
                  <div className="settings-group-card-fields">{appearanceFields.map(f => f.node)}</div>
                </div>
              )}
              {languageFields.length > 0 && (
                <div className="settings-group-card settings-group-card--compact">
                  <h4 className="settings-group-card-title">{t('sections.language')}</h4>
                  <div className="settings-group-card-fields">{languageFields.map(f => f.node)}</div>
                </div>
              )}
              {preferencesFields.length > 0 && (
                <div className="settings-group-card settings-group-card--compact">
                  <h4 className="settings-group-card-title">{t('sections.preferences')}</h4>
                  <div className="settings-group-card-fields">{preferencesFields.map(f => f.node)}</div>
                </div>
              )}
            </div>
          </Tabs.Content>

          <Tabs.Content value="integrations" className="settings-tab-panel">
            <IntegrationsSection
              isExpanded={activeSettingsTab === 'integrations'}
              settings={settings}
              loading={settingsLoading}
              onSetSetting={setSetting}
              onClearSetting={clearSetting}
              searchQuery={searchQuery}
            />
          </Tabs.Content>

          <Tabs.Content value="content" className="settings-tab-panel">
            <FiltersSection
              isExpanded={activeSettingsTab === 'content'}
              filters={filters}
              filterOrigin={filterOrigin}
              filtersLoading={filtersLoading}
              onDeleteFilter={onDeleteFilter}
              onOpenCreate={onOpenCreateFilter}
              searchQuery={searchQuery}
            />
            <div style={{ height: '1.75rem' }} />
            <M3uSourcesSection
              isExpanded={activeSettingsTab === 'content'}
              sources={sources}
              originNames={originNames}
              loading={sourcesLoading}
              onCreate={createSource}
              onUpdate={updateSource}
              onDelete={deleteSource}
              searchQuery={searchQuery}
            />
          </Tabs.Content>

          <Tabs.Content value="notifications" className="settings-tab-panel">
            <NotificationsSection
              isExpanded={activeSettingsTab === 'notifications'}
              settings={settings}
              loading={settingsLoading}
              onSetSetting={setSetting}
              onClearSetting={clearSetting}
              searchQuery={searchQuery}
            />
          </Tabs.Content>

          <Tabs.Content value="advanced" className="settings-tab-panel">
            <AdvancedSection
              isExpanded={activeSettingsTab === 'advanced'}
              settings={settings}
              loading={settingsLoading}
              onSetSetting={setSetting}
              onClearSetting={clearSetting}
              searchQuery={searchQuery}
            />
          </Tabs.Content>

          <Tabs.Content value="system" className="settings-tab-panel">
            <SystemStatusSection isExpanded={activeSettingsTab === 'system'} />

            {matchesQuery(t('sections.about')) && (
              <section className="settings-section" style={{ marginTop: '1.5rem' }}>
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
                    style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem', textDecoration: 'none', fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                  >
                    <GithubIcon size={14} />
                    {t('about.repositoryLink')}
                  </a>
                </div>
              </section>
            )}
          </Tabs.Content>
        </Tabs.Root>
      </div>
    </div>
  );
}
