import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { BarChart3 } from 'lucide-react';
import { Select } from '../../shared/components/ui';
import { useMenuAnalyticsSummary } from '../../shared/hooks/queries/useMenuBuilder';
import { useMenuBranchProfile } from './useMenuBranchProfile';

const MenuAnalyticsPage: React.FC = () => {
  const { t } = useTranslation();
  const { branches, selectedLocationId, setSelectedLocationId } = useMenuBranchProfile();
  const [days, setDays] = useState<number>(30);
  const { data: summary } = useMenuAnalyticsSummary(selectedLocationId, days);

  const steps = [
    { key: 'qrScans', label: t('design.studio.metricScans'), value: summary?.qrScans ?? 0 },
    { key: 'menuViews', label: t('design.studio.metricViews'), value: summary?.menuViews ?? 0 },
    {
      key: 'itemViews',
      label: t('design.menuApp.metricItemViews'),
      value: summary?.itemViews ?? 0,
    },
    { key: 'addToCarts', label: t('design.studio.metricCarts'), value: summary?.addToCarts ?? 0 },
    {
      key: 'checkoutStarts',
      label: t('design.menuApp.metricCheckouts'),
      value: summary?.checkoutStarts ?? 0,
    },
    {
      key: 'ordersPlaced',
      label: t('design.studio.metricOrders'),
      value: summary?.ordersPlaced ?? 0,
    },
  ];
  const maxVal = Math.max(1, ...steps.map((s) => s.value));

  return (
    <>
      <HelmetContainer title={`${t('common.nav.menuAnalytics')} · rtylr`} />
      <div className="w-full max-w-5xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <BarChart3 size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuAnalytics')}</h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.menuApp.analyticsSubtitle')}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {branches.length > 0 && (
              <Select
                aria-label={t('design.studio.branchLabel')}
                value={String(selectedLocationId)}
                onChange={(e) => setSelectedLocationId(Number(e.target.value) || 0)}
                className="h-8 text-xs w-44"
              >
                <option value="0">{t('design.studio.defaultBranch')}</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={String(branch.id)}>
                    {branch.name}
                  </option>
                ))}
              </Select>
            )}
            <Select
              aria-label={t('design.studio.analyticsTitle')}
              value={String(days)}
              onChange={(e) => setDays(Number(e.target.value) || 30)}
              className="h-8 text-xs w-32"
            >
              <option value="7">7d</option>
              <option value="30">30d</option>
              <option value="90">90d</option>
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
          {steps.map((step) => (
            <div
              key={step.key}
              className="rounded-xl border p-4"
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <div className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                {step.label}
              </div>
              <div className="mt-1 text-2xl font-semibold">{step.value}</div>
            </div>
          ))}
        </div>

        <section
          className="rounded-xl border p-5 space-y-3"
          style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
        >
          <h2 className="text-sm font-semibold">{t('design.menuApp.funnelTitle')}</h2>
          <div className="space-y-3">
            {steps.map((step) => {
              const pct = Math.round((step.value / maxVal) * 100);
              return (
                <div key={step.key} className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-medium">{step.label}</span>
                    <span style={{ color: 'var(--dash-text-muted)' }}>{step.value}</span>
                  </div>
                  <div
                    className="h-2.5 w-full rounded-full overflow-hidden"
                    style={{ background: 'var(--dash-hover)' }}
                  >
                    <div
                      className="h-full rounded-full transition-all"
                      style={{
                        width: `${step.value > 0 ? Math.max(4, pct) : 0}%`,
                        background: 'var(--dash-accent, #4583C4)',
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </section>
      </div>
    </>
  );
};

export default MenuAnalyticsPage;
