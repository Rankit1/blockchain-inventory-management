import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

export default function Replenish() {
  const { role } = useRole();
  const [assetId, setAssetId] = useState('');
  const [quantity, setQuantity] = useState('');
  const [notice, setNotice] = useState('');
  const [error, setError] = useState('');

  async function submit(e) {
    e.preventDefault();
    setError(''); setNotice('');
    try {
      await api.requestReplenishment({ assetId, quantity: Number(quantity) }, role);
      setNotice(`Replenishment requested for ${assetId}.`);
      setAssetId(''); setQuantity('');
    } catch (e2) {
      setError(e2.message);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="Replenishment · POST /api/replenish/request"
        title="Request replenishment"
        desc="Flag a low-stock asset for restocking."
      />
      <Notice kind="err">{error}</Notice>
      <Notice kind="ok">{notice}</Notice>
      <form className="card" style={{ maxWidth: 480 }} onSubmit={submit}>
        <Field label="Asset ID">
          <input required value={assetId} onChange={(e) => setAssetId(e.target.value)} />
        </Field>
        <Field label="Requested quantity">
          <input required type="number" min="1" value={quantity} onChange={(e) => setQuantity(e.target.value)} />
        </Field>
        <button className="btn" type="submit">Submit request</button>
      </form>
    </>
  );
}
