import { useTranslation } from 'react-i18next';

const LANGUAGES: { code: 'en' | 'fr'; labelKey: string }[] = [
  { code: 'en', labelKey: 'language.en' },
  { code: 'fr', labelKey: 'language.fr' },
];

interface FloatingHeaderProps {
  onOpenSystemStatus: () => void;
}

export function FloatingHeader({ onOpenSystemStatus }: FloatingHeaderProps) {
  const { t, i18n } = useTranslation();
  const activeLanguage = i18n.language.startsWith('fr') ? 'fr' : 'en';

  return (
    <header className="glass-header">
      <div>
        <h1 className="app-title">
          🛰️ {t('app.title')}
        </h1>
        <p className="app-subtitle">
          {t('app.subtitle')}
        </p>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <button
          type="button"
          aria-label={t('systemStatus.iconLabel')}
          title={t('systemStatus.iconLabel')}
          onClick={onOpenSystemStatus}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '2rem',
            height: '2rem',
            borderRadius: 'var(--radius-sm)',
            border: '1px solid var(--primary-slate)',
            background: 'var(--bg-app)',
            color: 'var(--text-secondary)',
            fontSize: '1rem',
            lineHeight: 1,
          }}
        >
          📶
        </button>
        <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
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
      </div>
    </header>
  );
}
