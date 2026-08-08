import { useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

export default function Audits() {
  const { role } = useRole();

  const [schedId, setSchedId] = useState('');
  const [schedDate, setSchedDate] = useState('');
  const [schedNotice, setSchedNotice] = useState('');
  const [schedError, setSchedError] = useState('');

  const [recId, setRecId] = useState('');
  const [recResult, setRecResult] = useState('pass');
  const [recNotes, setRecNotes] = useState('');
  const [recNotice, setRecNotice] = useState('');
  const [recError, setRecError] = useState('');

  async function schedule(e) {
    e.preventDefault();
    setSchedError(''); setSchedNotice('');
    try {
      await api.scheduleAudit({ assetId: schedId, scheduledDate: schedDate }, role);
      setSchedNotice(`Audit scheduled for ${schedId}.`);
      setSchedId(''); setSchedDate('');
    } catch (e2) {
      setSchedError(e2.message);
    }
  }

  async function record(e) {
    e.preventDefault();
    setRecError(''); setRecNotice('');
    try {
      await api.recordAudit({ assetId: recId, result: recResult, notes: recNotes }, role);
      setRecNotice(`Audit result recorded for ${recId}.`);
      setRecId(''); setRecNotes('');
    } catch (e2) {
      setRecError(e2.message);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="GenAI Ops · schedule-audit, record-audit"
        title="Audits"
        desc="Schedule audits manually or via PredictiveAgent, then record results and update audit metadata. Requires ASSET_AUDITOR or SYSTEM_ADMIN."
      />

      <div className="grid grid-2">
        <div className="card">
          <div className="card-title">Schedule audit</div>
          <Notice kind="err">{schedError}</Notice>
          <Notice kind="ok">{schedNotice}</Notice>
          <form onSubmit={schedule}>
            <Field label="Asset ID">
              <input required value={schedId} onChange={(e) => setSchedId(e.target.value)} />
            </Field>
            <Field label="Scheduled date">
              <input required type="date" value={schedDate} onChange={(e) => setSchedDate(e.target.value)} />
            </Field>
            <button className="btn" type="submit">Schedule</button>
          </form>
        </div>

        <div className="card">
          <div className="card-title">Record audit result</div>
          <Notice kind="err">{recError}</Notice>
          <Notice kind="ok">{recNotice}</Notice>
          <form onSubmit={record}>
            <Field label="Asset ID">
              <input required value={recId} onChange={(e) => setRecId(e.target.value)} />
            </Field>
            <Field label="Result">
              <select value={recResult} onChange={(e) => setRecResult(e.target.value)}>
                <option value="pass">Pass</option>
                <option value="fail">Fail</option>
                <option value="flagged">Flagged</option>
              </select>
            </Field>
            <Field label="Notes">
              <textarea rows={3} value={recNotes} onChange={(e) => setRecNotes(e.target.value)} />
            </Field>
            <button className="btn btn-outline" type="submit">Record result</button>
          </form>
        </div>
      </div>
    </>
  );
}
