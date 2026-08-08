import { NavLink, Route, Routes, Navigate } from 'react-router-dom';
import { useRole } from './api/RoleContext.jsx';

import Dashboard from './pages/Dashboard.jsx';
import Assets from './pages/Assets.jsx';
import AssetDetail from './pages/AssetDetail.jsx';
import IssueAsset from './pages/IssueAsset.jsx';
import ConsumeTransfer from './pages/ConsumeTransfer.jsx';
import Classify from './pages/Classify.jsx';
import Audits from './pages/Audits.jsx';
import Replenish from './pages/Replenish.jsx';
import Reports from './pages/Reports.jsx';
import Ledger from './pages/Ledger.jsx';
import Admin from './pages/Admin.jsx';
import Assistant from './pages/Assistant.jsx';

const NAV = [
  { group: 'Overview', links: [
    { to: '/', label: 'Dashboard', end: true }
  ]},
  { group: 'Assets', links: [
    { to: '/assets', label: 'All Assets' },
    { to: '/assets/issue', label: 'Issue Asset' },
    { to: '/assets/move', label: 'Consume / Transfer' }
  ]},
  { group: 'GenAI Ops', links: [
    { to: '/classify', label: 'Classify Priority' },
    { to: '/audits', label: 'Audits' },
    { to: '/replenish', label: 'Replenishment' },
    { to: '/assistant', label: 'Assistant' }
  ]},
  { group: 'Governance', links: [
    { to: '/reports', label: 'Reports' },
    { to: '/ledger', label: 'Ledger' },
    { to: '/admin', label: 'Agent Admin' }
  ]}
];

export default function App() {
  const { role, setRole, ROLES } = useRole();

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">InventoryChain<span className="dot">.</span></div>
        {NAV.map((section) => (
          <div key={section.group}>
            <div className="nav-group-label">{section.group}</div>
            {section.links.map((link) => (
              <NavLink
                key={link.to}
                to={link.to}
                end={link.end}
                className={({ isActive }) => 'nav-link' + (isActive ? ' active' : '')}
              >
                {link.label}
              </NavLink>
            ))}
          </div>
        ))}
        <div className="sidebar-foot">
          Simulation / Fabric hybrid ledger.<br />Port 8080.
        </div>
      </aside>

      <div>
        <div className="topbar">
          <div className="role-switcher">
            <span>Acting as</span>
            <select value={role} onChange={(e) => setRole(e.target.value)}>
              {ROLES.map((r) => (
                <option key={r} value={r}>{r}</option>
              ))}
            </select>
          </div>
        </div>

        <main className="main">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/assets" element={<Assets />} />
            <Route path="/assets/issue" element={<IssueAsset />} />
            <Route path="/assets/move" element={<ConsumeTransfer />} />
            <Route path="/assets/:id" element={<AssetDetail />} />
            <Route path="/classify" element={<Classify />} />
            <Route path="/audits" element={<Audits />} />
            <Route path="/replenish" element={<Replenish />} />
            <Route path="/assistant" element={<Assistant />} />
            <Route path="/reports" element={<Reports />} />
            <Route path="/ledger" element={<Ledger />} />
            <Route path="/admin" element={<Admin />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}
