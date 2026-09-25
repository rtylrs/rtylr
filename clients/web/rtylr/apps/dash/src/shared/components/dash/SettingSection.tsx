import React from 'react';
import { cn } from '../../lib/utils';

export interface SettingSectionProps {
  title: string;
  description?: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  className?: string;
}

export const SettingSection: React.FC<SettingSectionProps> = ({
  title,
  description,
  children,
  footer,
  className,
}) => {
  return (
    <div
      className={cn('rounded-xl', className)}
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--dash-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {title}
        </p>
        {description && (
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {description}
          </p>
        )}
      </div>
      <div className="px-5 py-4">{children}</div>
      {footer && (
        <div
          className="px-5 py-3 flex items-center justify-end gap-2"
          style={{ borderTop: '1px solid var(--dash-border)', background: 'var(--dash-hover)' }}
        >
          {footer}
        </div>
      )}
    </div>
  );
};

export default SettingSection;
