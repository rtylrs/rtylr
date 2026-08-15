import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Settings } from 'lucide-react';
import { EmptyState, SectionHeader } from './ui';
import { useHRConfigStore } from '../stores/hrConfigStore';
import { isPayrollConfigured } from '../lib/payrollConfig';
import { useGetHRConfig } from '../hooks/queries/useSetup';

export function PayrollConfigRequired(): React.JSX.Element {
  const { t } = useTranslation();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title={t('payroll.title')} subtitle={t('payroll.subtitle')} />
      <EmptyState
        icon={<Settings size={20} />}
        title={t('payroll.configRequired')}
        description={t('payroll.configRequiredDescription')}
        action={
          <Link
            to="/settings/payroll-config"
            style={{ color: 'var(--hr-accent)', textDecoration: 'none', fontSize: 12 }}
          >
            {t('payroll.goToConfig')}
          </Link>
        }
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
