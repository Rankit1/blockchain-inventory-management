import { useState, useEffect } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, Field } from '../components/UI.jsx';

const AGENTS = [
  { key: 'genai', label: 'GenAIService', desc: 'Auto-classifies missing or stale priority tiers.' },
  { key: 'predictive', label: 'PredictiveAgent', desc: 'Schedules audits for P1 assets with missing audit history.' },
  { key: 'vision', label: 'VisionAgent', desc: 'Simulated visual asset auditing.' },
  { key: 'document', label: 'DocumentAgent', desc: 'Simulated document intelligence and warranty enrichment.' }
];

export default function Admin() {
  const { role } = useRole();
  const [enabled, setEnabled] = useState({ genai: true, predictive: true, vision: true, document: true });
  const [busyKey, setBusyKey] = useState('');
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const [llmProvider, setLlmProvider] = useState('dummy');
  const [ocrProvider, setOcrProvider] = useState('dummy');
  const [providerNotice, setProviderNotice] = useState('');
  const [providerError, setProviderError] = useState('');

  useEffect(() => {
    async function loadStatus() {
      try {
        const res = await api.getAgentControl(role);
        if (res.llmProvider) setLlmProvider(res.llmProvider);
        if (res.ocrProvider) setOcrProvider(res.ocrProvider);
        if (res.agents) setEnabled((prev) => ({ ...prev, ...res.agents }));
      } catch (err) {
        // Non-fatal if role not yet admin
      }
    }
    loadStatus();
  }, [role]);

  async function toggle(key) {
    setError(''); setNotice(''); setBusyKey(key);
    const next = !enabled[key];
    try {
      await api.agentControl({ agent: key, enabled: next }, role);
      setEnabled((e) => ({ ...e, [key]: next }));
      setNotice(`${key} agent ${next ? 'enabled' : 'disabled'}.`);
    } catch (e2) {
      setError(e2.message);
    } finally {
      setBusyKey('');
    }
  }

  async function updateProviders(e) {
    e.preventDefault();
    setProviderError(''); setProviderNotice('');
    try {
      await api.agentControl({ llmProvider, ocrProvider }, role);
      setProviderNotice('Model providers updated.');
    } catch (e2) {
      setProviderError(e2.message);
    }
  }

  return (
    <>
      <PageHead
        eyebrow="Admin · POST /api/admin/agents/control"
        title="Agent administration"
        desc="Runtime toggle for GenAI automation agents and active model provider selection. Requires IT_ADMIN or SYSTEM_ADMIN."
      />
      <Notice kind="err">{error}</Notice>
      <Notice kind="ok">{notice}</Notice>

      <div className="card" style={{ marginBottom: 16 }}>
        <div className="card-title" style={{ marginBottom: 4 }}>Runtime agents</div>
        {AGENTS.map((a) => (
          <div className="toggle-row" key={a.key}>
            <div>
              <div style={{ fontSize: 14, fontWeight: 500 }}>{a.label}</div>
              <div style={{ fontSize: 12.5, color: 'var(--text-dim)' }}>{a.desc}</div>
            </div>
            <button
              className={`switch ${enabled[a.key] ? 'on' : ''}`}
              disabled={busyKey === a.key}
              onClick={() => toggle(a.key)}
              aria-label={`Toggle ${a.label}`}
            >
              <span className="knob" />
            </button>
          </div>
        ))}
      </div>

      <div className="card" style={{ maxWidth: 460 }}>
        <div className="card-title">Model providers</div>
        <Notice kind="err">{providerError}</Notice>
        <Notice kind="ok">{providerNotice}</Notice>
        <form onSubmit={updateProviders}>
          <Field label="LLM provider">
            <select value={llmProvider} onChange={(e) => setLlmProvider(e.target.value)}>
              <option value="dummy">dummy</option>
              <option value="openai">openai</option>
              <option value="mistral">mistral</option>
              <option value="gemini">gemini</option>
            </select>
          </Field>
          <Field label="OCR provider">
            <select value={ocrProvider} onChange={(e) => setOcrProvider(e.target.value)}>
              <option value="dummy">dummy</option>
              <option value="tesseract">tesseract</option>
              <option value="azure">azure</option>
            </select>
          </Field>
          <button className="btn btn-outline" type="submit">Update providers</button>
        </form>
      </div>
    </>
  );
}
