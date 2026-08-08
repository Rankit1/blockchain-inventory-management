import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client.js';
import { PageHead, PriorityBadge, Notice, Field } from '../components/UI.jsx';

export default function Assets() {
  const [assets, setAssets] = useState([]);
  const [dept, setDept] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  async function load(d) {
    setLoading(true);
    setError('');
    try {
      const res = await api.listAssets(d || undefined);
      setAssets(Array.isArray(res) ? res : res?.assets || []);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  return (
    <>
      <PageHead
        eyebrow="Assets · GET /api/assets"
        title="All assets"
        desc="Every tracked asset across departments, with current stock and priority tier."
      />
      <Notice kind="err">{error}</Notice>

      <div className="card" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-end' }}>
          <div style={{ flex: 1 }}>
            <Field label="Filter by department">
              <input value={dept} onChange={(e) => setDept(e.target.value)} placeholder="e.g. RADIOLOGY" />
            </Field>
          </div>
          <button className="btn" style={{ marginBottom: 14 }} onClick={() => load(dept)}>Apply filter</button>
          <button className="btn btn-outline" style={{ marginBottom: 14 }} onClick={() => { setDept(''); load(); }}>Clear</button>
        </div>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr><th>ID</th><th>Name</th><th>Dept</th><th>Stock</th><th>Tier</th><th>Status</th></tr>
          </thead>
          <tbody>
            {assets.map((a) => (
              <tr key={a.id || a.assetId}>
                <td><Link to={`/assets/${a.id || a.assetId}`}>{a.id || a.assetId}</Link></td>
                <td style={{ fontFamily: 'var(--font-body)' }}>{a.name || '—'}</td>
                <td>{a.deptId || a.department || '—'}</td>
                <td>{a.stock ?? a.quantity ?? '—'}</td>
                <td><PriorityBadge tier={a.priorityTier || a.priority} /></td>
                <td>{a.status || a.retired ? 'Retired' : 'Active'}</td>
              </tr>
            ))}
            {!loading && assets.length === 0 && (
              <tr><td colSpan={6} className="empty">No assets found for this filter.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </>
  );
}
