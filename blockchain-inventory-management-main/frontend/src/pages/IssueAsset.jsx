import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

const initial = { name: '', deptId: '', category: 'server', quantity: '', threshold: '2', replacementCost: '', leadTime: '', safetyCompliance: false };

export default function IssueAsset() {
  const { role } = useRole();
  const [form, setForm] = useState(initial);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  function set(k, v) { setForm((f) => ({ ...f, [k]: v })); }

  async function submit(e) {
    e.preventDefault();
    setError(''); setNotice(''); setBusy(true);
    try {
      const payload = {
        name: form.name,
        deptId: form.deptId,
        category: form.category,
        qty: Number(form.quantity) || 0,
        threshold: Number(form.threshold) || 2,
        replacementCost: Number(form.replacementCost) || 0,
        leadTime: Number(form.leadTime) || 0,
        safetyCompliance: form.safetyCompliance
      };
      const res = await api.issueAsset(payload, role);
      setNotice(`Issued: ${res?.assetID || res?.assetId || res?.id || 'success'}`);
      setForm(initial);
    } catch (e2) {
      setError(e2.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="Assets · POST /api/assets/issue"
        title="Issue a new asset"
        desc="Creates an inventory asset and records the issuing transaction on the ledger."
      />
      <Notice kind="err">{error}</Notice>
      <Notice kind="ok">{notice}</Notice>

      <form className="card" style={{ maxWidth: 520 }} onSubmit={submit}>
        <Field label="Asset name">
          <input required value={form.name} onChange={(e) => set('name', e.target.value)} placeholder="e.g. Infusion pump" />
        </Field>
        <Field label="Department ID">
          <input required value={form.deptId} onChange={(e) => set('deptId', e.target.value)} placeholder="e.g. Lab" />
        </Field>
        <Field label="Category">
          <select value={form.category} onChange={(e) => set('category', e.target.value)}>
            <option value="server">Server</option>
            <option value="network">Network Device</option>
            <option value="diagnostic">Diagnostic / Medical Equipment</option>
            <option value="laptop">Laptop / Computer</option>
            <option value="storage">Storage / Infrastructure</option>
            <option value="general">General Asset</option>
          </select>
        </Field>
        <Field label="Quantity">
          <input required type="number" min="0" value={form.quantity} onChange={(e) => set('quantity', e.target.value)} />
        </Field>
        <Field label="Replenishment Threshold">
          <input required type="number" min="1" value={form.threshold} onChange={(e) => set('threshold', e.target.value)} />
        </Field>
        <button className="btn" disabled={busy} type="submit">{busy ? 'Issuing…' : 'Issue asset'}</button>
      </form>
    </>
  );
}
