import React from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../lib/utils';

export interface RoleBadgeProps {
  role: string;
  className?: string;
}

export const RoleBadge: React.FC<RoleBadgeProps> = ({ role, className }) => {
  const { t } = useTranslation();
  const key = role.toLowerCase();
  const label = t(`common.roleLabels.${key}`, { defaultValue: role });
  return (
    <span
      className={cn('inline-flex items-center rounded-md px-2 py-1 text-xs font-medium', className)}
      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
    >
      {label}
    </span>
  );
};

export default RoleBadge;
