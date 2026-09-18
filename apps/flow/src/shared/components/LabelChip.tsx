import React from 'react';
import { X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Label } from '../stores/flowStore';
import { withAlpha, onTint } from '../lib/colorUtils';

interface LabelChipProps {
  label: Label;
  small?: boolean;
  onRemove?: () => void;
}

export const LabelChip: React.FC<LabelChipProps> = ({ label, small, onRemove }) => {
  const { t } = useTranslation();
  const foreground = onTint(label.color);

  return (
    <span
      className="inline-flex items-center gap-1 rounded-full font-medium"
      style={{
        background: withAlpha(label.color, 10),
        color: foreground,
        border: `1px solid ${withAlpha(label.color, 20)}`,
        padding: small ? '1px 6px' : '2px 8px',
        fontSize: 'var(--text-meta)',
      }}
    >
      {/* Decorative: the label name is right beside it, so it carries no
          accessible name of its own. */}
      <span
        aria-hidden="true"
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: label.color }}
      />
      {label.name}
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          aria-label={t('flow.labels.removeLabel', { name: label.name })}
          className="flow-focus ml-0.5 inline-flex items-center hover:opacity-70"
          style={{ color: foreground }}
        >
          <X size={11} strokeWidth={2.5} />
        </button>
      )}
    </span>
  );
};

export default LabelChip;
