import { useEffect, useState } from 'react';
import { api } from '../api/client.js';
import { PageHead, Notice } from '../components/UI.jsx';

function shortHash(h) {
  if (!h) return '—';
  return h.length > 14 ? `${h.slice(0, 6)}…${h.slice(-6)}` : h;
}

export default function Ledger() {
  const [blocks, setBlocks] = useState([]);
  const [verify, setVerify] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    setError('');
    try {
      const [b, v] = await Promise.all([api.ledgerBlocks(), api.ledgerVerify()]);
      setBlocks(Array.isArray(b) ? b : b?.blocks || []);
      setVerify(v);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  const valid = verify?.valid ?? verify?.ok;

  return (
    <>
      <PageHead
        eyebrow="Ledger · /api/ledger/blocks, /verify"
        title="Chain inspection"
        desc="Fabric ledger blocks and a live integrity check across the chain."
      />
      <Notice kind="err">{error}</Notice>

      <div className="card" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 4 }}>
          <div className="card-title">Chain integrity</div>
          {verify && (
            <span className={`badge ${valid ? 'badge-ok' : 'badge-p1'}`}>{valid ? 'Verified' : 'Integrity failure'}</span>
          )}
        </div>
        <p style={{ fontSize: 13, color: 'var(--text-dim)', margin: '4px 0 0' }}>
          {verify ? `Checked ${blocks.length} block${blocks.length === 1 ? '' : 's'} in the local ledger.` : loading ? 'Checking…' : 'No response yet.'}
        </p>
      </div>

      <div className="card" style={{ marginBottom: 16 }}>
        <div className="card-title" style={{ marginBottom: 10 }}>Block sequence</div>
        <div className="chain-strip">
          {blocks.map((b, i) => (
            <div key={b.index ?? b.hash ?? i} style={{ display: 'flex', alignItems: 'center' }}>
              <div className="chain-block">
                <div className="idx">BLOCK {b.index ?? i}</div>
                <div className="hash">{shortHash(b.hash)}</div>
              </div>
              {i < blocks.length - 1 && <div className="chain-link" />}
            </div>
          ))}
          {blocks.length === 0 && <div className="empty">No blocks committed yet.</div>}
        </div>
      </div>

      <div className="card">
        <div className="card-title">Block detail</div>
        <table>
          <thead><tr><th>Index</th><th>Hash</th><th>Prev hash</th><th>Tx count</th></tr></thead>
          <tbody>
            {blocks.map((b, i) => (
              <tr key={b.index ?? i}>
                <td>{b.index ?? i}</td>
                <td>{shortHash(b.hash)}</td>
                <td>{shortHash(b.prevHash || b.previousHash)}</td>
                <td>{(b.transactions || []).length}</td>
              </tr>
            ))}
            {blocks.length === 0 && <tr><td colSpan={4} className="empty">Ledger is empty.</td></tr>}
          </tbody>
        </table>
      </div>
    </>
  );
}
