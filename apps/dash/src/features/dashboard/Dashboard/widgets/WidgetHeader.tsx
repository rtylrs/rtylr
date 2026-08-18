import React from 'react';
import { ArrowUpRight } from 'lucide-react';

interface WidgetHeaderAction {
  label: string;
  onClick: () => void;
}

interface WidgetHeaderProps {
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  iconColor?: string;
  title: string;
  action?: WidgetHeaderAction;
}

const WidgetHeader: React.FC<WidgetHeaderProps> = ({
  icon: Icon,
  iconColor,
  title,
  action,
}): React.JSX.Element => {
  const color = iconColor ?? 'var(--dash-text-muted)';

  return (
    <div
      className="flex items-center justify-between px-4 py-3"
      style={{ borderBottom: '1px solid var(--dash-border)' }}
    >
      <div className="flex items-center gap-2">
        <span
          className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md"
          style={{ background: `color-mix(in srgb, ${color} 12%, transparent)` }}
        >
          <Icon size={13} style={{ color }} />
        </span>
        <p className="text-sm font-semibold truncate" style={{ color: 'var(--dash-text)' }}>
          {title}
        </p>
      </div>
      {action && (
        <button
          onClick={action.onClick}
          className="flex items-center gap-1 shrink-0 text-xs font-medium transition-colors duration-150"
          style={{ color: 'var(--accent)' }}
        >
          {action.label} <ArrowUpRight size={11} aria-hidden="true" />
        </button>
      )}
    </div>
  );
};

export default WidgetHeader;
