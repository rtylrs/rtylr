import React from 'react';
import { Info } from 'lucide-react';

/**
 * The notice that sits above a form whose writes do not reach the backend.
 *
 * Dash's menu authoring changes only the local zustand store and then reports
 * success, while `useGetMenu` refetches over the top. Someone builds a menu,
 * watches it persist through `localStorage`, and it never reaches the POS or
 * the public menu, so it is absent on another device and absent for every
 * other consumer.
 *
 * Until the mutations are wired to `pos/menu_item/v1`, saying so is the honest
 * behaviour. See rtylr#1150.
 */
export const UnavailableNotice: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div
    role="status"
    className="flex items-start gap-2 rounded-lg p-3 text-sm"
    style={{
      background: 'var(--dash-hover)',
      border: '1px solid var(--dash-border)',
      color: 'var(--dash-text-muted)',
    }}
  >
    <Info size={14} className="mt-0.5 shrink-0" aria-hidden />
    <p className="min-w-0">{children}</p>
  </div>
);

export default UnavailableNotice;
