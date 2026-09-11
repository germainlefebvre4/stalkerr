import { useTranslation } from 'react-i18next';
import { SlidersHorizontal } from 'lucide-react';

interface FloatingHeaderProps {
  onOpenSettings: () => void;
}

export function FloatingHeader({ onOpenSettings }: FloatingHeaderProps) {
  const { t } = useTranslation();

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
          aria-label={t('settings:trigger')}
          title={t('settings:trigger')}
          onClick={onOpenSettings}
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
          }}
        >
          <SlidersHorizontal size={16} />
        </button>
      </div>
    </header>
  );
}
