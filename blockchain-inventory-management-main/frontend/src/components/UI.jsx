export function PageHead({ eyebrow, title, desc }) {
  return (
    <div className="page-head">
      {eyebrow && <div className="eyebrow">{eyebrow}</div>}
      <h1 className="page-title">{title}</h1>
      {desc && <p className="page-desc">{desc}</p>}
    </div>
  );
}

export function StatCard({ label, value, sub }) {
  return (
    <div className="card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      {sub && <div style={{ fontSize: 12, color: 'var(--text-dim)' }}>{sub}</div>}
    </div>
  );
}

export function PriorityBadge({ tier }) {
  if (!tier) return <span className="badge badge-off">—</span>;
  const cls = tier === 'P1' ? 'badge-p1' : tier === 'P2' ? 'badge-p2' : 'badge-p3';
  return <span className={`badge ${cls}`}>{tier}</span>;
}

export function Notice({ kind = 'ok', children }) {
  if (!children) return null;
  return <div className={`notice notice-${kind === 'ok' ? 'ok' : 'err'}`}>{children}</div>;
}

export function Field({ label, children }) {
  return (
    <div className="field">
      <label>{label}</label>
      {children}
    </div>
  );
}
