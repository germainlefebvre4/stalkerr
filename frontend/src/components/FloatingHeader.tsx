interface FloatingHeaderProps {
  healthStatus: 'healthy' | 'unhealthy' | 'checking';
}

export function FloatingHeader({ healthStatus }: FloatingHeaderProps) {
  return (
    <header className="glass-header">
      <div>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'var(--primary-slate)' }}>
          🛰️ Stalkeer Dashboard
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem', fontWeight: 500 }}>
          IHM de pilotage, logs, filtres et réorganisation de médias
        </p>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <span 
          className={`badge ${healthStatus === 'healthy' ? 'badge-success' : healthStatus === 'unhealthy' ? 'badge-failed' : 'badge-pending'}`} 
          style={{ gap: '0.5rem' }}
        >
          <span 
            className="pulse-dot" 
            style={{ display: healthStatus === 'healthy' ? 'inline-block' : 'none' }}
          ></span>
          API: {healthStatus === 'healthy' ? 'En ligne' : healthStatus === 'unhealthy' ? 'Erreur' : 'Vérification...'}
        </span>
      </div>
    </header>
  );
}
