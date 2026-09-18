import React from 'react';
import { useShortcutStore } from '../stores/shortcutStore';

const SHORTCUTS: { keys: string; label: string }[] = [
  { keys: 'Cmd/Ctrl + K', label: 'Open command palette' },
  { keys: 'Enter', label: 'Charge current order' },
  { keys: 'Esc', label: 'Close modal / cancel' },
  { keys: 'F1', label: 'Show this help' },
  { keys: 'F2', label: 'Apply discount' },
  { keys: 'F3', label: 'Refund order' },
];

const HelpOverlay: React.FC = () => {
  const open = useShortcutStore((s) => s.helpOpen);
  const setHelp = useShortcutStore((s) => s.setHelp);
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={() => setHelp(false)}
      role="dialog"
      aria-modal="true"
    >
      <div
        className="w-full max-w-md rounded-2xl bg-pos-card border border-pos-border p-5 space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div>
          <h2 className="text-lg font-bold text-pos-text">Keyboard shortcuts</h2>
          <p className="text-xs text-pos-muted mt-1">Work faster with the keyboard.</p>
        </div>
        <ul className="space-y-1.5">
          {SHORTCUTS.map((s) => (
            <li
              key={s.keys}
              className="flex justify-between items-center p-2.5 rounded-lg bg-white/5"
            >
              <span className="text-sm text-pos-text">{s.label}</span>
              <kbd className="text-xs font-mono px-2 py-1 rounded bg-white/10 text-pos-muted border border-pos-border">
                {s.keys}
              </kbd>
            </li>
          ))}
        </ul>
        <button
          onClick={() => setHelp(false)}
          className="w-full h-11 rounded-lg bg-pos-text text-pos-bg font-bold"
        >
          Close
        </button>
      </div>
    </div>
  );
};

export default HelpOverlay;
