import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

export default function Assistant() {
  const { role } = useRole();
  const [query, setQuery] = useState('');
  const [log, setLog] = useState([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  async function ask(e) {
    e.preventDefault();
    if (!query.trim()) return;
    setError(''); setBusy(true);
    const q = query;
    setQuery('');
    try {
      const res = await api.assistantQuery({ query: q }, role);
      setLog((l) => [...l, { q, a: res?.answer || res?.response || JSON.stringify(res) }]);
    } catch (e2) {
      setError(e2.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="Assistant · POST /api/assistant/query"
        title="Assistant"
        desc="Currently a stub — returns operational guidance on asset, audit, and priority workflows rather than a production conversational model."
      />
      <Notice kind="err">{error}</Notice>

      <div className="card" style={{ maxWidth: 640 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 16 }}>
          {log.length === 0 && <div className="empty">Ask about asset priority, audit scheduling, or reports.</div>}
          {log.map((entry, i) => (
            <div key={i}>
              <div style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 4 }}>You: {entry.q}</div>
              <div style={{ fontSize: 13.5, background: 'var(--bg-raised)', border: '1px solid var(--border)', borderRadius: 8, padding: 10 }}>
                {entry.a}
              </div>
            </div>
          ))}
        </div>
        <form onSubmit={ask} style={{ display: 'flex', gap: 8 }}>
          <div style={{ flex: 1 }}>
            <Field label="Message">
              <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="How do I schedule an audit for a P1 asset?" />
            </Field>
          </div>
          <button className="btn" style={{ height: 40, marginTop: 22 }} disabled={busy} type="submit">
            {busy ? '…' : 'Send'}
          </button>
        </form>
      </div>
    </>
  );
}
