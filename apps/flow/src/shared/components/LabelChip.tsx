import React from 'react';
import { Label } from '../stores/flowStore';
import { withAlpha } from '../lib/colorUtils';

interface LabelChipProps {
  label: Label;
  small?: boolean;
  onRemove?: () => void;
}

export const LabelChip: React.FC<LabelChipProps> = ({ label, small, onRemove }) => {
  return (
    <span
      className="inline-flex items-center gap-1 rounded-full font-medium"
      style={{
        background: withAlpha(label.color, 10),
        color: label.color,
        border: `1px solid ${withAlpha(label.color, 20)}`,
        padding: small ? '1px 6px' : '2px 8px',
        fontSize: 'var(--text-meta)',
      }}
    >
      <span className="h-1.5 w-1.5 rounded-full" style={{ background: label.color }} />
      {label.name}
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          className="ml-0.5 hover:opacity-70"
          style={{ color: label.color }}
        >
          ×
        </button>
      )}
    </span>
  );
};

export default LabelChip;
