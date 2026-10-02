import React, { useId } from 'react';
import { X, Keyboard } from 'lucide-react';
import { useDialogFocus } from '../hooks/useDialogFocus';

interface KeyboardShortcutsModalProps {
  open: boolean;
  onClose: () => void;
}

const SHORTCUTS = [
  { keys: ['cmd', 'k'], description: 'Open command palette' },
  { keys: ['/erp'], description: 'Focus search' },
  { keys: ['n'], description: 'New product' },
  { keys: ['shift', 'i'], description: 'Go to Inventory' },
  { keys: ['shift', 'p'], description: 'Go to Purchasing' },
  { keys: ['shift', 's'], description: 'Go to Suppliers' },
  { keys: ['shift', 'd'], description: 'Go to Dashboard' },
  { keys: ['?'], description: 'Show this help' },
  { keys: ['esc'], description: 'Close modal or palette' },
];

const KeyboardShortcutsModal: React.FC<KeyboardShortcutsModalProps> = ({ open, onClose }) => {
  const titleId = useId();
  const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose });

  if (!open) return null;
  return (
    <div
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'var(--hub-overlay-heavy)' }}
      tabIndex={-1}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-lg rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <h3
            id={titleId}
            className="text-sm font-semibold flex items-center gap-2"
            style={{ color: 'var(--hub-text)' }}
          >
            <Keyboard size={14} style={{ color: 'var(--erp-accent)' }} />
            Keyboard Shortcuts
          </h3>
          <button type="button" onClick={onClose} aria-label="Close keyboard shortcuts">
            <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
          </button>
        </div>
        <div className="space-y-2">
          {SHORTCUTS.map((s, i) => (
            <div key={i} className="flex items-center justify-between">
              <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                {s.description}
              </span>
              <div className="flex gap-1">
                {s.keys.map((k) => (
                  <kbd
                    key={k}
                    className="text-[10px] font-mono px-1.5 py-0.5 rounded"
                    style={{
                      background: 'var(--hub-hover)',
                      border: '1px solid var(--hub-border)',
                      color: 'var(--hub-text)',
                    }}
                  >
                    {k}
                  </kbd>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default KeyboardShortcutsModal;
