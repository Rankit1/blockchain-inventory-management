import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

export default function ConsumeTransfer() {
  const { role } = useRole();
  const [mode, setMode] = useState('consume');
  const [assetId, setAssetId] = useState('');
  const [quantity, setQuantity] = useState('');
  const [toDept, setToDept] = useState('');
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  async function submit(e) {
    e.preventDefault();
    setError(''); setNotice(''); setBusy(true);
    try {
      if (mode === 'consume') {
        await api.consumeAsset({ assetId, quantity: Number(quantity) }, role);
      } else {
        await api.transferAsset({ assetId, quantity: Number(quantity), toDeptId: toDept }, role);
      }
      setNotice(`${mode === 'consume' ? 'Consumed' : 'Transferred'} stock for ${assetId}.`);
      setAssetId(''); setQuantity(''); setToDept('');
    } catch (e2) {
      setError(e2.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="Assets · POST /api/assets/consume, /transfer"
        title="Consume or transfer stock"
        desc="Draw down stock for use, or move it between departments."
      />
      <Notice kind="err">{error}</Notice>
      <Notice kind="ok">{notice}</Notice>

      <div className="card" style={{ maxWidth: 520 }}>
        <div style={{ display: 'flex', gap: 8, marginBottom: 18 }}>
          <button
            type="button"
            className={mode === 'consume' ? 'btn' : 'btn btn-outline'}
            onClick={() => setMode('consume')}
          >Consume</button>
          <button
            type="button"
            className={mode === 'transfer' ? 'btn' : 'btn btn-outline'}
            onClick={() => setMode('transfer')}
          >Transfer</button>
        </div>

        <form onSubmit={submit}>
          <Field label="Asset ID">
            <input required value={assetId} onChange={(e) => setAssetId(e.target.value)} placeholder="asset-0012" />
          </Field>
          <Field label="Quantity">
            <input required type="number" min="1" value={quantity} onChange={(e) => setQuantity(e.target.value)} />
          </Field>
          {mode === 'transfer' && (
            <Field label="Destination department">
              <input required value={toDept} onChange={(e) => setToDept(e.target.value)} placeholder="e.g. ICU" />
            </Field>
          )}
          <button className="btn" disabled={busy} type="submit">
            {busy ? 'Submitting…' : mode === 'consume' ? 'Consume stock' : 'Transfer stock'}
          </button>
        </form>
      </div>
    </>
  );
}
