import React from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { ShieldX } from 'lucide-react';
import { useUserStore } from '@voxire/auth';
import { EmptyState } from '@voxire/ui';
import { canHrPermission, type HRPermission } from '../lib/hrPermissions';

interface HRPermissionGuardProps {
  permission: HRPermission;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

const HRPermissionGuard = ({
  permission,
  children,
  fallback,
}: HRPermissionGuardProps): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  if (canHrPermission(permission)) return <>{children}</>;
  if (fallback) return <>{fallback}</>;

  return (
    <EmptyState
      icon={ShieldX}
      title={t('errors.unauthorized')}
      action={{ label: t('common.back'), onClick: () => navigate({ to: '/' }) }}
    />
  );
};

export default HRPermissionGuard;
