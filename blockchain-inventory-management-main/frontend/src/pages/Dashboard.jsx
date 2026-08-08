import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, StatCard, PriorityBadge, Notice } from '../components/UI.jsx';

export default function Dashboard() {
  const { role } = useRole();
  const [assets, setAssets] = useState([]);
  const [util, setUtil] = useState(null);
  const [ledgerOk, setLedgerOk] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [a, verify] = await Promise.all([
          api.listAssets(),
          api.ledgerVerify().catch(() => null)
        ]);
        if (cancelled) return;
        setAssets(Array.isArray(a) ? a : a?.assets || []);
        setLedgerOk(verify);
      } catch (e) {
        if (!cancelled) setError(e.message);
      }
      try {
        const u = await api.utilizationReport(role);
        if (!cancelled) setUtil(u);
      } catch {
        // report may be role-gated; ignore silently on dashboard
      }
    }
    load();
    return () => { cancelled = true; };
  }, [role]);

  const p1Count = assets.filter((a) => a.priorityTier === 'P1' || a.priority === 'P1').length;
  const deptCount = new Set(assets.map((a) => a.deptId || a.department)).size;

  return (
    <>
      <PageHead
        eyebrow="Overview"
        title="Ledger status"
        desc="Live snapshot of assets, priority exposure, and chain integrity across simulation or Fabric mode."
      />
      <Notice kind="err">{error}</Notice>

      <div className="grid grid-4" style={{ marginBottom: 16 }}>
        <StatCard label="Tracked assets" value={assets.length} />
        <StatCard label="P1 priority" value={p1Count} sub="Needs predictive audit coverage" />
        <StatCard label="Departments" value={deptCount || '—'} />
        <StatCard
          label="Chain integrity"
          value={ledgerOk === null ? '…' : ledgerOk.valid || ledgerOk.ok ? 'Valid' : 'Broken'}
        />
      </div>

      <div className="grid grid-2">
        <div className="card">
          <div className="card-title">Highest priority assets</div>
          <table>
            <thead>
              <tr><th>Asset</th><th>Dept</th><th>Tier</th></tr>
            </thead>
            <tbody>
              {assets.slice(0, 6).map((a) => (
                <tr key={a.id || a.assetId}>
                  <td><Link to={`/assets/${a.id || a.assetId}`}>{a.name || a.id || a.assetId}</Link></td>
                  <td>{a.deptId || a.department || '—'}</td>
                  <td><PriorityBadge tier={a.priorityTier || a.priority} /></td>
                </tr>
              ))}
              {assets.length === 0 && (
                <tr><td colSpan={3} className="empty">No assets returned yet — issue one to get started.</td></tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="card">
          <div className="card-title">Utilization</div>
          {util ? (
            <pre style={{ fontFamily: 'var(--font-mono)', fontSize: 12.5, color: 'var(--text-dim)', whiteSpace: 'pre-wrap' }}>
              {JSON.stringify(util, null, 2)}
            </pre>
          ) : (
            <div className="empty">Requires STORE_MANAGER, IT_ADMIN, or SYSTEM_ADMIN role.</div>
          )}
        </div>
      </div>
    </>
  );
}
