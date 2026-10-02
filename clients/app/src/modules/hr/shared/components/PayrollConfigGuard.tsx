import React from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Settings } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { SectionHeader } from './ui';
import { useHRConfigStore } from '../stores/hrConfigStore';
import { isPayrollConfigured } from '../lib/payrollConfig';
import { useGetHRConfig } from '../hooks/queries/useSetup';

export function PayrollConfigRequired(): React.JSX.Element {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title={t('hrPayroll.title')} subtitle={t('hrPayroll.subtitle')} />
      <EmptyState
        icon={Settings}
        title={t('hrPayroll.configRequired')}
        description={t('hrPayroll.configRequiredDescription')}
        action={{
          label: t('hrPayroll.goToConfig'),
          onClick: () => navigate({ to: '/settings/hr/payroll-config' }),
        }}
      />
    </div>
  );
}

interface PayrollConfigGuardProps {
  children: React.ReactNode;
}

export function usePayrollConfigured(): boolean {
  const storedConfig = useHRConfigStore((state) => state.config);
  const { data: fetchedConfig } = useGetHRConfig();

  return isPayrollConfigured(fetchedConfig ?? storedConfig);
}

export function PayrollConfigGuard({ children }: PayrollConfigGuardProps): React.JSX.Element {
  return usePayrollConfigured() ? <>{children}</> : <PayrollConfigRequired />;
}
