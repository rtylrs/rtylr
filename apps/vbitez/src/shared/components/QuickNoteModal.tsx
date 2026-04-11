import React, { useState } from 'react';

interface QuickNoteModalProps {
  open: boolean;
  initialValue?: string;
  title?: string;
  onSave: (value: string) => void;
  onCancel: () => void;
}

const QUICK_CHIPS = ['No onions', 'Extra spicy', 'Gluten free', 'Allergy', 'To go', 'Rush'];

const QuickNoteModal: React.FC<QuickNoteModalProps> = ({
  open,
  initialValue = '',
  title = 'Add note',
  onSave,
  onCancel,
}) => {
  const [value, setValue] = useState(initialValue);
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      role="dialog"
      aria-modal="true"
    >
      <div className="w-full max-w-md rounded-2xl bg-pos-card border border-pos-border p-5 space-y-4">
        <h2 className="text-lg font-bold text-pos-text">{title}</h2>
        <div className="flex flex-wrap gap-2">
          {QUICK_CHIPS.map((chip) => (
            <button
              key={chip}
              onClick={() => setValue((v) => (v ? `${v}, ${chip}` : chip))}
              className="px-3 h-9 rounded-full bg-white/8 text-xs font-semibold text-pos-text hover:bg-white/14"
            >
              {chip}
            </button>
          ))}
        </div>
        <textarea
          rows={4}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Type a note..."
          className="w-full px-3 py-2 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25 resize-none"
        />
        <div className="grid grid-cols-2 gap-2">
          <button
            className="h-12 rounded-lg bg-white/10 text-pos-text font-semibold"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            className="h-12 rounded-lg bg-pos-text text-pos-bg font-bold"
            onClick={() => onSave(value)}
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
};

export default QuickNoteModal;
