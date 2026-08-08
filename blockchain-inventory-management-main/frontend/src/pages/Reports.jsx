import { useEffect, useState } from 'react';
import { api } from '../api/client.js';
import { useRole } from '../api/RoleContext.jsx';
import { PageHead, Notice, StatCard } from '../components/UI.jsx';

export default function Reports() {
  const { role } = useRole();
  const [util, setUtil] = useState(null);
  const [compliance, setCompliance] = useState(null);
  const [utilError, setUtilError] = useState('');
  const [complianceError, setComplianceError] = useState('');
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    setUtilError(''); setComplianceError('');
    try {
      setUtil(await api.utilizationReport(role));
    } catch (e) {
      setUtilError(e.message);
    }
    try {
      setCompliance(await api.complianceReport(role));
    } catch (e) {
      setComplianceError(e.message);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [role]);

  const overdue = compliance?.overdueAssets || compliance?.overdue || [];

  return (
    <>
      <PageHead
        eyebrow="Reports · utilization, compliance"
        title="Reports"
        desc="Asset utilization metrics and compliance audit summary. Utilization requires STORE_MANAGER, IT_ADMIN, or SYSTEM_ADMIN; compliance requires IT_ADMIN or SYSTEM_ADMIN."
      />

      <div className="grid grid-2" style={{ marginBottom: 16 }}>
        <div className="card">
          <div className="card-title">Utilization</div>
          <Notice kind="err">{utilError}</Notice>
          {util && (
            <div className="grid grid-2" style={{ marginTop: 12 }}>
              <StatCard label="Utilization rate" value={util.utilizationRate ?? util.rate ?? '—'} />
              <StatCard label="Idle assets" value={util.idleAssets ?? util.idle ?? '—'} />
            </div>
          )}
          {!util && !utilError && !loading && <div className="empty">No data returned.</div>}
        </div>

        <div className="card">
          <div className="card-title">Compliance</div>
          <Notice kind="err">{complianceError}</Notice>
          {compliance && (
            <div className="grid grid-2" style={{ marginTop: 12 }}>
              <StatCard label="Compliant assets" value={compliance.compliantCount ?? compliance.compliant ?? '—'} />
              <StatCard label="Overdue audits" value={overdue.length || compliance.overdueCount || '—'} />
            </div>
          )}
        </div>
      </div>

      {overdue.length > 0 && (
        <div className="card">
          <div className="card-title">Overdue assets</div>
          <table>
            <thead><tr><th>Asset</th><th>Dept</th><th>Last audit</th></tr></thead>
            <tbody>
              {overdue.map((a, i) => (
                <tr key={a.id || a.assetId || i}>
                  <td>{a.id || a.assetId}</td>
                  <td>{a.deptId || a.department || '—'}</td>
                  <td>{a.lastAuditDate || a.lastAudit || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
