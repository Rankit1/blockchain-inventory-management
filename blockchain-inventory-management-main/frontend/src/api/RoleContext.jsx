import { createContext, useContext, useState } from 'react';

const ROLES = [
  'SYSTEM_ADMIN',
  'IT_ADMIN',
  'ASSET_AUDITOR',
  'AI_OPS',
  'STORE_MANAGER'
];

const RoleContext = createContext(null);

export function RoleProvider({ children }) {
  const [role, setRole] = useState('SYSTEM_ADMIN');
  return (
    <RoleContext.Provider value={{ role, setRole, ROLES }}>
      {children}
    </RoleContext.Provider>
  );
}

export function useRole() {
  return useContext(RoleContext);
}
