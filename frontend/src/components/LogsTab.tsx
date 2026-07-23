import * as Tabs from '@radix-ui/react-tabs';
import { ProcessingLog } from '../types';

interface LogsTabProps {
  logs: ProcessingLog[];
  logsLoading: boolean;
  onFetchLogs: () => void;
}

export function LogsTab({
  logs,
  logsLoading,
  onFetchLogs
}: LogsTabProps) {
  return (
    <Tabs.Content value="logs" className="card" style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>Dernières Exécutions de Tâches</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Historique des rafraîchissements M3U, nettoyages de base et enrichissements</p>
        </div>
        <button onClick={onFetchLogs} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>Actualiser ↻</button>
      </div>

      <div style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
        <table className="custom-table">
          <thead>
            <tr>
              <th>Action exécutée</th>
              <th>Éléments Traités</th>
              <th>Statut Pipeline</th>
              <th>Lancement</th>
              <th>Détails d'Erreur</th>
            </tr>
          </thead>
          <tbody>
            {logsLoading && logs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Chargement en cours...</td>
              </tr>
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Aucun log enregistré en base de données</td>
              </tr>
            ) : (
              logs.map(log => (
                <tr key={log.id}>
                  <td style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{log.action}</td>
                  <td style={{ fontWeight: 600 }}>{log.item_count} items</td>
                  <td>
                    <span className={`badge ${log.status === 'success' ? 'badge-success' : log.status === 'failed' ? 'badge-failed' : 'badge-progress'}`}>
                      {log.status === 'in_progress' ? 'En cours' : log.status === 'success' ? 'Réussi' : 'Échoué'}
                    </span>
                  </td>
                  <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{new Date(log.started_at).toLocaleString()}</td>
                  <td style={{ maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {log.error_message ? (
                      <span style={{ color: 'var(--status-failed-text)', fontWeight: 600 }} title={log.error_message}>{log.error_message}</span>
                    ) : (
                      <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>Aucune erreur</span>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </Tabs.Content>
  );
}
