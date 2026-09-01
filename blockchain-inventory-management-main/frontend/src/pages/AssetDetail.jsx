import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, PriorityBadge, Notice, Field } from '../components/UI.jsx';

export default function AssetDetail() {
  const { id } = useParams();
  const { role } = useRole();
  const [asset, setAsset] = useState(null);
  const [history, setHistory] = useState([]);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [reason, setReason] = useState('');

  async function load() {
    try {
      const [a, h] = await Promise.all([api.getAsset(id), api.getAssetHistory(id)]);
      setAsset(a);
      setHistory(Array.isArray(h) ? h : h?.history || []);
    } catch (e) {
      setError(e.message);
    }
  }

  useEffect(() => { load(); }, [id]);

  async function retire() {
    setError(''); setNotice('');
    try {
      await api.retireAsset({ assetId: id, reason: reason || 'Retired via console' }, role);
      setNotice('Asset retired.');
      load();
    } catch (e) {
      setError(e.message);
    }
  }

  return (
    <>
      <PageHead eyebrow={`Asset · ${id}`} title={asset?.name || id} desc="Lifecycle detail, transaction history, and retirement control." />
      <Notice kind="err">{error}</Notice>
      <Notice kind="ok">{notice}</Notice>

      <div className="grid grid-2" style={{ marginBottom: 16 }}>
        <div className="card">
          <div className="card-title">Details</div>
          {asset ? (
            <table>
              <tbody>
                <tr><td>Department</td><td>{asset.deptId || asset.department || '—'}</td></tr>
                <tr><td>Category</td><td>{asset.category || '—'}</td></tr>
                <tr><td>Stock / Threshold</td><td>{(asset.qty ?? asset.stock ?? asset.quantity ?? '—')} / {(asset.threshold ?? '—')}</td></tr>
                <tr><td>Priority tier</td><td><PriorityBadge tier={asset.priorityTier || asset.priority} /> {asset.criticalityScore ? `(Score: ${asset.criticalityScore})` : ''}</td></tr>
                <tr><td>Lifecycle status</td><td>{asset.lifecycleState || asset.status || (asset.retired ? 'Retired' : 'Active')}</td></tr>
                <tr><td>Last audit date</td><td>{asset.lastAuditDate || '—'}</td></tr>
                <tr><td>Warranty expiry</td><td>{asset.warrantyExpiry || '—'}</td></tr>
                <tr><td>AMC expiry</td><td>{asset.amcExpiry || '—'}</td></tr>
              </tbody>
            </table>
          ) : <div className="empty">Loading…</div>}
        </div>

        <div className="card">
          <div className="card-title">Retire asset</div>
          <p style={{ fontSize: 13, color: 'var(--text-dim)', marginTop: 0 }}>
            Requires IT_ADMIN or SYSTEM_ADMIN. Permanently marks this asset retired.
          </p>
          <Field label="Reason">
            <input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="e.g. end of life" />
          </Field>
          <button className="btn btn-danger" onClick={retire}>Retire asset</button>
        </div>
      </div>

      <div className="card">
        <div className="card-title">Transaction history</div>
        <table>
          <thead><tr><th>Tx ID</th><th>Type</th><th>Dept</th><th>Time</th></tr></thead>
          <tbody>
            {history.map((h, i) => (
              <tr key={h.txId || i}>
                <td title={h.txId}>{h.txId ? (h.txId.length > 16 ? h.txId.slice(0, 16) + '…' : h.txId) : '—'}</td>
                <td>{h.type || (h.isDelete ? 'DELETE' : (h.value ? 'STATE_UPDATE' : 'TX'))}</td>
                <td>{h.deptId || h.department || h.value?.deptId || '—'}</td>
                <td>{h.timestamp || h.time || '—'}</td>
              </tr>
            ))}
            {history.length === 0 && <tr><td colSpan={4} className="empty">No transactions recorded yet.</td></tr>}
          </tbody>
        </table>
      </div>

      <p style={{ marginTop: 16 }}><Link to="/assets" style={{ color: 'var(--text-dim)', fontSize: 13 }}>← Back to all assets</Link></p>
    </>
  );
}
