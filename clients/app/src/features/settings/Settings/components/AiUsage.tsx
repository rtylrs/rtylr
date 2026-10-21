import { useTranslation } from 'react-i18next';
import { PageHeader } from '../../../../shared/components/dash/PageHeader';
import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Users, User } from 'lucide-react';
import { Button } from '../../../../shared/components/ui';
import {
  useAiUsage,
  useAiQuotaMode,
  useUpdateAiQuotaMode,
  type AiQuotaMode,
} from '../../../../shared/hooks/queries/useAiQuota';
import { useUserStore } from '../../../../shared/stores/userStore';
import { canManageDash } from '../../../../shared/utils/canManageDash';

const fmt = (n: number): string => (n ?? 0).toLocaleString();

const MODES: {
  id: AiQuotaMode;
  labelKey: string;
  descKey: string;
  defaultLabel: string;
  defaultDesc: string;
  icon: typeof Users;
}[] = [
  {
    id: 'pool',
    labelKey: 'settings.aiUsage.poolLabel',
    descKey: 'settings.aiUsage.poolDescription',
    defaultLabel: 'Shared team pool',
    defaultDesc: 'Everyone draws from one monthly allowance.',
    icon: Users,
  },
  {
    id: 'per_person',
    labelKey: 'settings.aiUsage.perPersonLabel',
    descKey: 'settings.aiUsage.perPersonDescription',
    defaultLabel: 'Per person',
    defaultDesc: 'The allowance is split evenly across seats.',
    icon: User,
  },
];

const AiUsage: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const isOwner = canManageDash(user);
  const {
    data: usage,
    isLoading: usageLoading,
    isError: usageError,
    refetch: refetchUsage,
  } = useAiUsage();
  const {
    data: mode,
    isLoading: modeLoading,
    isError: modeError,
    refetch: refetchMode,
  } = useAiQuotaMode();
  const updateMode = useUpdateAiQuotaMode();

  const usageUnavailable = usageError || !usage;
  const modeUnavailable = modeError || !mode;
  const allowance = usage?.allowanceTokens ?? 0;
  const used = usage?.usedTokens ?? 0;
  const pct = allowance > 0 ? Math.min(100, Math.round((used / allowance) * 100)) : 0;
  const over = usage?.overageTokens ?? 0;
  const barColor = over > 0 ? '#ef4444' : pct >= 80 ? '#f59e0b' : 'var(--dash-text)';

  return (
    <>
      <HelmetContainer title={`${t('common.nav.aiUsage')} · rtylr`} />
      <div className="portal-management-page space-y-5 max-w-2xl">
        <PageHeader
          title={t('common.nav.aiUsage')}
          description={t('common.management.aiUsageDescription')}
        />

        {/* Usage */}
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {usageLoading ? (
            <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
              {t('common.loading', 'Loading…')}
            </p>
          ) : usageUnavailable ? (
            <div role="alert" className="space-y-3">
              <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                {t('settings.aiUsage.loadFailed')}
              </p>
              <Button variant="outline" size="sm" onClick={() => void refetchUsage()}>
                {t('common.refresh', 'Refresh')}
              </Button>
            </div>
          ) : (
            <>
              <div className="flex items-baseline justify-between mb-2">
                <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                  {fmt(used)} / {fmt(allowance)} {t('settings.aiUsage.tokens', 'tokens')}
                </span>
                <span
                  className="text-[11px] font-medium px-2 py-0.5 rounded-full uppercase tracking-wider"
                  style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                >
                  {usage.planSlug}
                </span>
              </div>
              <div
                className="h-2 rounded-full overflow-hidden"
                style={{ background: 'var(--dash-hover)' }}
              >
                <div style={{ width: `${pct}%`, height: '100%', background: barColor }} />
              </div>
              <div className="flex items-center justify-between mt-2 text-[11px]">
                <span style={{ color: 'var(--dash-text-dim)' }}>
                  {fmt(usage.remainingTokens)} {t('settings.aiUsage.remaining', 'remaining')}
                  {over > 0 ? ` · ${fmt(over)} ${t('settings.aiUsage.over', 'over')}` : ''}
                </span>
                {usage.periodEnd && (
                  <span style={{ color: 'var(--dash-text-dim)' }}>
                    {t('settings.aiUsage.resets', 'Resets')}{' '}
                    {new Date(usage.periodEnd).toLocaleDateString()}
                  </span>
                )}
              </div>
              {usage.blocked && (
                <p className="text-xs mt-3" style={{ color: '#ef4444' }}>
                  {t(
                    'settings.aiUsage.limitReached',
                    "You've reached this month's AI limit. Upgrade your plan to keep using Caret."
                  )}
                </p>
              )}
            </>
          )}
        </div>

        {/* Allocation mode */}
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {t('settings.aiUsage.allocation', 'Allocation')}
          </h2>
          <p className="text-xs mt-0.5 mb-3" style={{ color: 'var(--dash-text-dim)' }}>
            {isOwner
              ? t(
                  'settings.aiUsage.ownerDescription',
                  'Choose how your team consumes the monthly AI allowance.'
                )
              : t(
                  'settings.aiUsage.memberDescription',
                  'Only an owner or admin can change how AI credits are allocated.'
                )}
          </p>
          {modeLoading ? (
            <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
              {t('settings.aiUsage.loadingAllocation', 'Loading allocation…')}
            </p>
          ) : modeUnavailable ? (
            <div role="alert" className="space-y-3">
              <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                {t('settings.aiUsage.loadAllocationFailed')}
              </p>
              <Button variant="outline" size="sm" onClick={() => void refetchMode()}>
                {t('common.refresh', 'Refresh')}
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {MODES.map((m) => {
                const active = mode === m.id;
                const Icon = m.icon;
                return (
                  <button
                    type="button"
                    key={m.id}
                    disabled={
                      !isOwner || updateMode.isPending || active || usageLoading || usageUnavailable
                    }
                    onClick={() => updateMode.mutate(m.id)}
                    className="text-start rounded-lg p-3 transition-colors disabled:cursor-default"
                    style={{
                      background: active ? 'var(--dash-hover)' : 'transparent',
                      border: active
                        ? '1px solid var(--dash-text)'
                        : '1px solid var(--dash-border)',
                      opacity: !isOwner && !active ? 0.6 : 1,
                    }}
                  >
                    <div className="flex items-center gap-2">
                      <Icon
                        size={14}
                        style={{ color: active ? 'var(--dash-text)' : 'var(--dash-text-muted)' }}
                      />
                      <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                        {t(m.labelKey, m.defaultLabel)}
                      </span>
                    </div>
                    <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
                      {t(m.descKey, m.defaultDesc)}
                    </p>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </>
  );
};

export default AiUsage;
