import { useTranslation } from 'react-i18next';

interface FloatingHeaderProps {
  healthStatus: 'healthy' | 'unhealthy' | 'checking';
}

const LANGUAGES: { code: 'en' | 'fr'; labelKey: string }[] = [
  { code: 'en', labelKey: 'language.en' },
  { code: 'fr', labelKey: 'language.fr' },
];

export function FloatingHeader({ healthStatus }: FloatingHeaderProps) {
  const { t, i18n } = useTranslation();
  const activeLanguage = i18n.language.startsWith('fr') ? 'fr' : 'en';

  const statusLabel = healthStatus === 'healthy'
    ? t('apiStatus.online')
    : healthStatus === 'unhealthy'
      ? t('apiStatus.error')
      : t('apiStatus.checking');

  return (
    <header className="glass-header">
      <div>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'var(--primary-slate)' }}>
          🛰️ {t('app.title')}
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem', fontWeight: 500 }}>
          {t('app.subtitle')}
        </p>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          <span className="sr-only">{t('language.label')}</span>
          <select
            aria-label={t('language.label')}
            value={activeLanguage}
            onChange={(e) => i18n.changeLanguage(e.target.value)}
            style={{
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--primary-slate)',
              background: 'var(--bg-app)',
              color: 'var(--text-secondary)',
              padding: '0.25rem 0.5rem',
              fontSize: '0.8rem',
            }}
          >
            {LANGUAGES.map(({ code, labelKey }) => (
              <option key={code} value={code}>{t(labelKey)}</option>
            ))}
          </select>
        </label>
        <span
          className={`badge ${healthStatus === 'healthy' ? 'badge-success' : healthStatus === 'unhealthy' ? 'badge-failed' : 'badge-pending'}`}
          style={{ gap: '0.5rem' }}
        >
          <span
            className="pulse-dot"
            style={{ display: healthStatus === 'healthy' ? 'inline-block' : 'none' }}
          ></span>
          API: {statusLabel}
        </span>
      </div>
    </header>
  );
}
