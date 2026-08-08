import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field, PriorityBadge } from '../components/UI.jsx';

export default function Classify() {
  const { role } = useRole();
  const [assetId, setAssetId] = useState('');
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const [overrideId, setOverrideId] = useState('');
  const [overrideTier, setOverrideTier] = useState('P2');
  const [overrideReason, setOverrideReason] = useState('');
  const [overrideNotice, setOverrideNotice] = useState('');
  const [overrideError, setOverrideError] = useState('');

  async function classify(e) {
    e.preventDefault();
    setError(''); setResult(null); setBusy(true);
    try {
      const res = await api.classifyAsset({ assetId }, role);
      setResult(res);
    } catch (e2) {
      setError(e2.message);
    } finally {
      setBusy(false);
    }
  }

  async function override(e) {
    e.preventDefault();
    setOverrideError(''); setOverrideNotice('');
    try {
      await api.updatePriority({ assetId: overrideId, priorityTier: overrideTier, reason: overrideReason }, role);
      setOverrideNotice(`Priority for ${overrideId} set to ${overrideTier}.`);
      setOverrideId(''); setOverrideReason('');
    } catch (e2) {
      setOverrideError(e2.message);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="GenAI Ops · POST /api/assets/classify"
        title="Classify asset priority"
        desc="Runs GenAIService across five criteria — business criticality, replacement cost, lead time, safety compliance, redundancy — to assign a P1/P2/P3 tier. Requires AI_OPS, ASSET_AUDITOR, or SYSTEM_ADMIN."
      />

      <div className="grid grid-2">
        <div className="card">
          <div className="card-title">Run classification</div>
          <Notice kind="err">{error}</Notice>
          <form onSubmit={classify}>
            <Field label="Asset ID">
              <input required value={assetId} onChange={(e) => setAssetId(e.target.value)} placeholder="asset-0012" />
            </Field>
            <button className="btn" disabled={busy} type="submit">{busy ? 'Classifying…' : 'Classify'}</button>
          </form>
          {result && (
            <div style={{ marginTop: 16 }}>
              <div style={{ marginBottom: 8 }}>Assigned tier: <PriorityBadge tier={result.priorityTier || result.priority} /></div>
              <pre style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--text-dim)', whiteSpace: 'pre-wrap' }}>
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          )}
        </div>

        <div className="card">
          <div className="card-title">Manual override</div>
          <p style={{ fontSize: 13, color: 'var(--text-dim)', marginTop: 0 }}>
            Requires ASSET_AUDITOR, IT_ADMIN, or SYSTEM_ADMIN. Always attach a justification.
          </p>
          <Notice kind="err">{overrideError}</Notice>
          <Notice kind="ok">{overrideNotice}</Notice>
          <form onSubmit={override}>
            <Field label="Asset ID">
              <input required value={overrideId} onChange={(e) => setOverrideId(e.target.value)} />
            </Field>
            <Field label="Priority tier">
              <select value={overrideTier} onChange={(e) => setOverrideTier(e.target.value)}>
                <option value="P1">P1</option>
                <option value="P2">P2</option>
                <option value="P3">P3</option>
              </select>
            </Field>
            <Field label="Justification">
              <textarea required rows={3} value={overrideReason} onChange={(e) => setOverrideReason(e.target.value)} />
            </Field>
            <button className="btn btn-outline" type="submit">Apply override</button>
          </form>
        </div>
      </div>
    </>
  );
}
