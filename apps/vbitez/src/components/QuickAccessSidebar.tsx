import React from 'react';
import { NavLink } from 'react-router-dom';
import { useSettingsStore } from '../stores/settingsStore';

const LINKS: { to: string; label: string; icon: string }[] = [
  { to: '/', label: 'Terminal', icon: '🖥' },
  { to: '/menu', label: 'Menu', icon: '📋' },
  { to: '/orders', label: 'Orders', icon: '🧾' },
  { to: '/tables', label: 'Tables', icon: '🪑' },
  { to: '/kds', label: 'KDS', icon: '👨‍🍳' },
  { to: '/customer-lookup', label: 'Customers', icon: '👤' },
  { to: '/gift-cards', label: 'Gift Cards', icon: '🎁' },
  { to: '/receipts', label: 'Receipts', icon: '🧮' },
  { to: '/day-summary', label: 'Day', icon: '📈' },
  { to: '/held-orders', label: 'Held', icon: '⏸' },
  { to: '/offline', label: 'Offline', icon: '📶' },
  { to: '/settings', label: 'Settings', icon: '⚙' },
];

const QuickAccessSidebar: React.FC = () => {
  const enabled = useSettingsStore((s) => s.quickAccessSidebar);
  if (!enabled) return null;
  return (
    <aside className="hidden md:flex shrink-0 w-[4.5rem] flex-col border-r border-pos-border bg-pos-card overflow-y-auto">
      {LINKS.map((l) => (
        <NavLink
          key={l.to}
          to={l.to}
          end={l.to === '/'}
          className={({ isActive }) =>
            `flex flex-col items-center justify-center gap-0.5 py-3 text-[10px] font-medium transition-colors ${
              isActive
                ? 'text-pos-text bg-white/8 border-l-2 border-pos-text'
                : 'text-pos-muted hover:text-pos-text hover:bg-white/5'
            }`
          }
        >
          <span className="text-lg">{l.icon}</span>
          <span>{l.label}</span>
        </NavLink>
      ))}
    </aside>
  );
};

export default QuickAccessSidebar;
