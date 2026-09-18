import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useShortcutStore } from '../stores/shortcutStore';
import { useSettingsStore } from '../stores/settingsStore';
import type { POSFeatureKey } from '../stores/settingsStore';

interface Command {
  id: string;
  label: string;
  shortcut?: string;
  action: () => void;
  category?: string;
  feature?: POSFeatureKey;
}

const CommandPalette: React.FC = () => {
  const open = useShortcutStore((s) => s.commandOpen);
  const setOpen = useShortcutStore((s) => s.setCommand);
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const features = useSettingsStore((s) => s.features);
  const businessProfile = useSettingsStore((s) => s.businessProfile);

  const commands: Command[] = useMemo(
    () => [
      { id: 'nav-pos', label: 'Go to Terminal', category: 'Navigate', action: () => navigate('/') },
      {
        id: 'nav-menu',
        label: businessProfile === 'retail' ? 'Go to Products' : 'Go to Menu Items',
        category: 'Navigate',
        action: () => navigate('/menu'),
      },
      {
        id: 'nav-orders',
        label: 'Go to Orders',
        category: 'Navigate',
        action: () => navigate('/orders'),
      },
      {
        id: 'nav-tables',
        label: 'Go to Tables',
        category: 'Navigate',
        action: () => navigate('/tables'),
        feature: 'tables',
      },
      {
        id: 'nav-combos',
        label: 'Go to Combos',
        category: 'Navigate',
        action: () => navigate('/combos'),
        feature: 'combos',
      },
      {
        id: 'nav-modifiers',
        label: 'Go to Modifiers',
        category: 'Navigate',
        action: () => navigate('/modifiers'),
        feature: 'modifiers',
      },
      {
        id: 'nav-gift',
        label: 'Go to Gift Cards',
        category: 'Navigate',
        action: () => navigate('/gift-cards'),
        feature: 'giftCards',
      },
      {
        id: 'nav-kds',
        label: 'Kitchen Display',
        category: 'Navigate',
        action: () => navigate('/kds'),
        feature: 'kitchenDisplay',
      },
      {
        id: 'nav-customer',
        label: 'Customer Lookup',
        category: 'Navigate',
        action: () => navigate('/customer-lookup'),
        feature: 'customers',
      },
      {
        id: 'nav-receipts',
        label: 'Receipts',
        category: 'Navigate',
        action: () => navigate('/receipts'),
      },
      {
        id: 'nav-day',
        label: 'Day Summary',
        category: 'Navigate',
        action: () => navigate('/day-summary'),
      },
      {
        id: 'nav-close-session',
        label: 'Close Session',
        category: 'Navigate',
        action: () => navigate('/close-session'),
      },
      {
        id: 'nav-offline',
        label: 'Offline Queue',
        category: 'Navigate',
        action: () => navigate('/offline'),
      },
      {
        id: 'nav-settings',
        label: 'Settings',
        category: 'Navigate',
        action: () => navigate('/settings'),
      },
      {
        id: 'nav-held',
        label: 'Held Orders',
        category: 'Navigate',
        action: () => navigate('/held-orders'),
      },
      {
        id: 'nav-cashdrop',
        label: 'Cash Drop / Paid Out',
        category: 'Navigate',
        action: () => navigate('/cash-drop'),
        feature: 'cashDrawer',
      },
      {
        id: 'nav-inventory',
        label: 'Inventory Management',
        category: 'Navigate',
        action: () => navigate('/inventory'),
        feature: 'inventory',
      },
      {
        id: 'nav-products',
        label: 'Product Quick Edit',
        category: 'Navigate',
        action: () => navigate('/products/quick'),
        feature: 'inventory',
      },
      {
        id: 'nav-override',
        label: 'Manager Actions',
        category: 'Navigate',
        action: () => navigate('/override'),
      },
      {
        id: 'nav-staff',
        label: 'Staff & Permissions',
        category: 'Navigate',
        action: () => navigate('/staff'),
      },
    ],
    [businessProfile, navigate]
  );

  const availableCommands = commands.filter(
    (command) => !command.feature || features[command.feature]
  );
  const filtered = query
    ? availableCommands.filter((c) => c.label.toLowerCase().includes(query.toLowerCase()))
    : availableCommands;

  useEffect(() => {
    if (!open) setQuery('');
  }, [open]);

  useEffect(() => {
    // Attach the Escape listener only while the palette is open instead of
    // keeping a global keydown handler alive for the app's lifetime.
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setOpen(false);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, setOpen]);

  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-24 bg-black/60 backdrop-blur-sm px-4"
      onClick={() => setOpen(false)}
      role="dialog"
      aria-modal="true"
    >
      <div
        className="w-full max-w-lg rounded-2xl bg-pos-card border border-pos-border overflow-hidden shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Type a command or search..."
          className="w-full h-14 px-5 bg-transparent border-b border-pos-border text-pos-text text-base outline-none"
        />
        <div className="max-h-80 overflow-y-auto">
          {filtered.length === 0 ? (
            <div className="p-6 text-center text-sm text-pos-muted">No commands found</div>
          ) : (
            filtered.map((c) => (
              <button
                key={c.id}
                className="w-full text-left px-5 py-3 hover:bg-white/8 flex items-center justify-between text-sm text-pos-text"
                onClick={() => {
                  setOpen(false);
                  c.action();
                }}
              >
                <span>{c.label}</span>
                {c.category && (
                  <span className="text-[10px] text-pos-muted uppercase">{c.category}</span>
                )}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default CommandPalette;
