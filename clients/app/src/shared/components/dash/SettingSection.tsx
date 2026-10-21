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
    <section className={cn('portal-settings-section', className)}>
      <div className="portal-settings-section-header">
        <h2>{title}</h2>
        {description && <p>{description}</p>}
      </div>
      <div className="portal-settings-section-body">{children}</div>
      {footer && <div className="portal-settings-section-footer">{footer}</div>}
    </section>
  );
};

export default SettingSection;
