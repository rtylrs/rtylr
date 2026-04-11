import React from 'react';
import { X, Keyboard } from 'lucide-react';

interface KeyboardShortcutsModalProps {
  open: boolean;
  onClose: () => void;
}

const SHORTCUTS = [
  { keys: ['cmd', 'k'], description: 'Open command palette' },
  { keys: ['/'], description: 'Focus search' },
  { keys: ['n'], description: 'New product' },
  { keys: ['shift', 'i'], description: 'Go to Inventory' },
  { keys: ['shift', 'p'], description: 'Go to Purchasing' },
  { keys: ['shift', 's'], description: 'Go to Suppliers' },
  { keys: ['shift', 'd'], description: 'Go to Dashboard' },
  { keys: ['?'], description: 'Show this help' },
  { keys: ['esc'], description: 'Close modal or palette' },
];

const KeyboardShortcutsModal: React.FC<KeyboardShortcutsModalProps> = ({ open, onClose }) => {
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'rgba(0,0,0,0.6)' }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-lg rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <h3
            className="text-sm font-semibold flex items-center gap-2"
            style={{ color: 'var(--hub-text)' }}
          >
            <Keyboard size={14} style={{ color: '#f59e0b' }} />
            Keyboard Shortcuts
          </h3>
          <button onClick={onClose}>
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
