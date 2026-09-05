const BASE = '';

async function request(path, { method = 'GET', body, role } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (role) headers['X-User-Role'] = role;

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined
  });

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const message = (data && (data.error || data.message)) || `Request failed (${res.status})`;
    throw new Error(message);
  }
  return data;
}

export const api = {
  health: () => request('/healthz'),

  // Asset Management
  issueAsset: (payload, role) => request('/api/assets/issue', { method: 'POST', body: payload, role }),
  consumeAsset: (payload, role) => request('/api/assets/consume', { method: 'POST', body: payload, role }),
  transferAsset: (payload, role) => request('/api/assets/transfer', { method: 'POST', body: payload, role }),
  getAsset: (id) => request(`/api/assets/${id}`),
  getAssetHistory: (id) => request(`/api/assets/${id}/history`),
  listAssets: (dept) => request(`/api/assets${dept ? `?dept=${encodeURIComponent(dept)}` : ''}`),

  // GenAI-augmented operations
  classifyAsset: (payload, role) => request('/api/assets/classify', { method: 'POST', body: payload, role }),
  updatePriority: (payload, role) => request('/api/assets/update-priority', { method: 'POST', body: payload, role }),
  scheduleAudit: (payload, role) => request('/api/assets/schedule-audit', { method: 'POST', body: payload, role }),
  recordAudit: (payload, role) => request('/api/assets/record-audit', { method: 'POST', body: payload, role }),
  retireAsset: (payload, role) => request('/api/assets/retire', { method: 'POST', body: payload, role }),

  // Replenishment
  requestReplenishment: (payload, role) => request('/api/replenish/request', { method: 'POST', body: payload, role }),

  // Reports
  utilizationReport: (role) => request('/api/reports/utilization', { role }),
  complianceReport: (role) => request('/api/reports/compliance', { role }),

  // Assistant / Admin
  assistantQuery: (payload, role) => request('/api/assistant/query', { method: 'POST', body: payload, role }),
  agentControl: (payload, role) => request('/api/admin/agents/control', { method: 'POST', body: payload, role }),
  getAgentControl: (role) => request('/api/admin/agents/control', { role }),

  // Ledger
  ledgerBlocks: () => request('/api/ledger/blocks'),
  ledgerVerify: () => request('/api/ledger/verify')
};
