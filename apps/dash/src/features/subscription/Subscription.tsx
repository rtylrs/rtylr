import React from 'react';
import { useTranslation } from '@voxire/i18n';
import { CheckCircle2, Clock, XCircle, AlertCircle, Box } from 'lucide-react';
import { APP_CONFIG, APP_URLS } from '@voxire/config';
import type { AppId } from '@voxire/config';
import { SettingSection, Breadcrumbs } from '../../shared/components/dash';

import {
  useMySubscription,
  useListPlans,
  type SubscriptionPayload,
  type PlanPayload,
} from '../../shared/hooks/queries/useSubscription';

// ── Status helpers ────────────────────────────────────────────────────────────

type SubStatus = SubscriptionPayload['status'];

interface StatusConfig {
  label: string;
  color: string;
  bg: string;
  Icon: React.ComponentType<{ size?: number }>;
}

const STATUS_MAP: Record<SubStatus, StatusConfig> = {
  active: {
    label: 'Active',
    color: 'var(--status-active)',
    bg: 'var(--status-active-bg)',
    Icon: CheckCircle2,
  },
  trial: {
    label: 'Trial',
    color: 'var(--status-trial)',
    bg: 'var(--status-trial-bg)',
    Icon: Clock,
  },
  cancelled: {
    label: 'Cancelled',
    color: 'var(--status-cancelled)',
    bg: 'var(--status-cancelled-bg)',
    Icon: XCircle,
  },
  expired: {
    label: 'Expired',
    color: 'var(--status-expired)',
    bg: 'var(--status-expired-bg)',
    Icon: AlertCircle,
  },
};

const StatusBadge: React.FC<{ status: SubStatus }> = ({ status }) => {
  const cfg = STATUS_MAP[status] ?? STATUS_MAP.expired;
  const { Icon } = cfg;
  return (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold"
      style={{ color: cfg.color, background: cfg.bg }}
    >
      <Icon size={10} />
      {cfg.label}
    </span>
  );
};

// ── App pill ─────────────────────────────────────────────────────────────────

const AppPill: React.FC<{ appId: string }> = ({ appId }) => {
  const id = appId as AppId;
  const config = APP_CONFIG[id];
  const name = config?.name ?? appId;
  const url = APP_URLS[id];

  const pill = (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium"
      style={{
        background: 'var(--accent-muted)',
        color: 'var(--accent)',
        border: '1px solid var(--accent-border)',
      }}
    >
      <Box size={10} />
      {name}
    </span>
  );

  if (url) {
    return (
      <a href={url} target="_blank" rel="noreferrer noopener">
        {pill}
      </a>
    );
  }
  return pill;
};

// ── Current plan card ─────────────────────────────────────────────────────────

const CurrentPlanCard: React.FC<{ sub: SubscriptionPayload }> = ({ sub }) => {
  const { t } = useTranslation();
  return (
    <SettingSection
      title={t('subscription.currentPlan')}
      description={t('subscription.currentPlanDesc')}
    >
      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-base font-bold" style={{ color: 'var(--dash-text)' }}>
              {sub.planName}
            </p>
            <p className="text-xs mt-0.5 font-mono" style={{ color: 'var(--dash-text-dim)' }}>
              {sub.planSlug}
            </p>
          </div>
          <StatusBadge status={sub.status} />
        </div>

        <div>
          <p className="text-xs font-medium mb-2" style={{ color: 'var(--dash-text-muted)' }}>
            {t('subscription.includedApps')}
          </p>
          <div className="flex flex-wrap gap-1.5">
            {sub.apps.map((app) => (
              <AppPill key={app} appId={app} />
            ))}
          </div>
        </div>
      </div>
    </SettingSection>
  );
};

// ── Plan card (for the "all plans" list) ─────────────────────────────────────

const PlanCard: React.FC<{ plan: PlanPayload; isCurrent: boolean }> = ({ plan, isCurrent }) => {
  const { t } = useTranslation();
  return (
    <div
      className="rounded-xl p-4 space-y-3"
      style={{
        border: isCurrent ? '1.5px solid var(--accent-border)' : '1px solid var(--dash-border)',
        background: isCurrent ? 'var(--accent-muted)' : 'var(--dash-card)',
      }}
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {plan.name}
        </p>
        {isCurrent && (
          <span
            className="text-[10px] font-bold px-1.5 py-0.5 rounded"
            style={{ background: 'var(--accent)', color: '#fff' }}
          >
            {t('subscription.current')}
          </span>
        )}
      </div>
      <div className="flex flex-wrap gap-1">
        {plan.apps.map((app) => (
          <AppPill key={app} appId={app} />
        ))}
      </div>
    </div>
  );
};

// ── Main page ─────────────────────────────────────────────────────────────────

const Subscription: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();

  return (
    <div className="space-y-6 max-w-2xl">
      <Breadcrumbs />

      {/* Current plan */}
      {subLoading ? (
        <div
          className="h-32 rounded-xl animate-pulse"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        />
      ) : sub ? (
        <CurrentPlanCard sub={sub} />
      ) : (
        <SettingSection
          title={t('subscription.currentPlan')}
          description={t('subscription.currentPlanDesc')}
        >
          <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
            {t('subscription.noPlan')}
          </p>
        </SettingSection>
      )}

      {/* All plans */}
      <SettingSection
        title={t('subscription.allPlans')}
        description={t('subscription.allPlansDesc')}
      >
        {plansLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-20 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : (
          <div className="space-y-3">
            {plans.map((plan) => (
              <PlanCard key={plan.id} plan={plan} isCurrent={plan.slug === sub?.planSlug} />
            ))}
          </div>
        )}
      </SettingSection>
    </div>
  );
};

export default Subscription;
