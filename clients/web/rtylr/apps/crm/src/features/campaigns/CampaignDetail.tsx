import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, Mail, FileQuestion } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetCampaign, useCampaignStats } from '../../shared/hooks/queries/useCampaigns';
import { Breadcrumbs, Card, SkeletonList } from '../../shared/components/ui';
import { CampaignStatsCard } from '../../shared/components/crm';

const CampaignDetail: React.FC = () => {
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const { t } = useTranslation();
  const campaignId = Number(id);
  const {
    data: fetched,
    isLoading: campaignLoading,
    error: campaignError,
  } = useGetCampaign(campaignId);
  const {
    data: fetchedStats,
    isLoading: statsLoading,
    isError: statsError,
  } = useCampaignStats(campaignId);
  const campaign = fetched;
  const stats = fetchedStats;

  const pct = (n: number, d: number) => (d > 0 ? Math.round((n / d) * 100) : 0);

  if (campaignLoading) {
    return (
      <div style={{ maxWidth: '1200px' }}>
        <SkeletonList count={6} />
      </div>
    );
  }

  if (campaignError || !campaign) {
    return (
      <div style={{ maxWidth: '1200px' }}>
        <EmptyState
          icon={FileQuestion}
          title={t('crm.campaignDetail.notFoundTitle')}
          description={t('crm.campaignDetail.notFoundDescription')}
          action={{
            label: t('crm.campaignDetail.backToCampaigns'),
            onClick: () => navigate({ to: '/campaigns' }),
          }}
        />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={`${t('crm.campaignDetail.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1200px' }}>
        <Breadcrumbs
          items={[{ label: t('nav.campaigns'), to: '/campaigns' }, { label: campaign.name }]}
        />
        <button
          type="button"
          onClick={() => navigate({ to: '/campaigns' })}
          style={{
            background: 'transparent',
            border: 'none',
            color: 'var(--hub-text-muted)',
            cursor: 'pointer',
            fontSize: '12px',
            display: 'flex',
            alignItems: 'center',
            gap: '4px',
            padding: 0,
            marginBottom: '14px',
          }}
        >
          <ArrowLeft size={12} /> Back
        </button>

        <Card style={{ marginBottom: '20px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
            <div
              style={{
                width: '48px',
                height: '48px',
                borderRadius: '12px',
                background: 'rgba(59,130,246,0.14)',
                color: '#3b82f6',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Mail size={20} />
            </div>
            <div style={{ flex: 1 }}>
              <h2
                style={{ fontSize: '20px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}
              >
                {campaign.name}
              </h2>
              <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
                {campaign.subject ?? ''}
              </p>
            </div>
            <span
              style={{
                padding: '4px 12px',
                borderRadius: '9999px',
                fontSize: '11px',
                fontWeight: 600,
                // Wire status is lowercase ('active'); comparing against the
                // capitalized label left this pill permanently gray.
                background:
                  campaign.status.toLowerCase() === 'active'
                    ? 'rgba(16,185,129,0.12)'
                    : 'rgba(148,163,184,0.12)',
                color: campaign.status.toLowerCase() === 'active' ? '#10b981' : '#94a3b8',
              }}
            >
              {campaign.status}
            </span>
          </div>
        </Card>

        {stats && (
          <>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(6, 1fr)',
                gap: '12px',
                marginBottom: '20px',
              }}
            >
              <CampaignStatsCard
                label={t('crm.campaignDetail.statSent')}
                value={stats.sent.toLocaleString()}
                subValue={t('crm.campaignDetail.subDelivered', { count: stats.delivered })}
              />
              <CampaignStatsCard
                label={t('crm.campaignDetail.statDelivered')}
                value={`${pct(stats.delivered, stats.sent)}%`}
                subValue={`${stats.delivered.toLocaleString()}`}
                color="#10b981"
              />
              <CampaignStatsCard
                label={t('crm.campaignDetail.statOpened')}
                value={`${pct(stats.opened, stats.delivered)}%`}
                subValue={t('crm.campaignDetail.subOpens', { count: stats.opened })}
                sparkline={stats.timeline.map((t) => t.opens)}
                color="#3b82f6"
              />
              <CampaignStatsCard
                label={t('crm.campaignDetail.statClicked')}
                value={`${pct(stats.clicked, stats.delivered)}%`}
                subValue={t('crm.campaignDetail.subClicks', { count: stats.clicked })}
                sparkline={stats.timeline.map((t) => t.clicks)}
                color="#8b5cf6"
              />
              <CampaignStatsCard
                label={t('crm.campaignDetail.statUnsubscribed')}
                value={stats.unsubscribed}
                subValue={`${pct(stats.unsubscribed, stats.delivered)}%`}
                color="#ef4444"
              />
              <CampaignStatsCard
                label={t('crm.campaignDetail.statRevenue')}
                value={`$${stats.revenue.toLocaleString()}`}
                subValue={t('crm.campaignDetail.subFromCampaign')}
                color="#22c55e"
              />
            </div>

            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 14px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {t('crm.campaignDetail.engagementByDay')}
              </h4>
              <div
                style={{ display: 'flex', alignItems: 'flex-end', gap: '10px', height: '160px' }}
              >
                {stats.timeline.map((t) => {
                  const max = Math.max(1, ...stats.timeline.map((x) => x.opens));
                  const openH = (t.opens / max) * 130;
                  const clickH = (t.clicks / max) * 130;
                  return (
                    <div
                      key={t.day}
                      style={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        gap: '6px',
                      }}
                    >
                      <div
                        style={{
                          display: 'flex',
                          gap: '4px',
                          alignItems: 'flex-end',
                          height: '130px',
                        }}
                      >
                        <div
                          style={{
                            width: '14px',
                            height: `${openH}px`,
                            background: '#3b82f6',
                            borderRadius: '3px 3px 0 0',
                          }}
                        />
                        <div
                          style={{
                            width: '14px',
                            height: `${clickH}px`,
                            background: '#8b5cf6',
                            borderRadius: '3px 3px 0 0',
                          }}
                        />
                      </div>
                      <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                        {t.day}
                      </span>
                    </div>
                  );
                })}
              </div>
              <div style={{ display: 'flex', gap: '16px', marginTop: '16px', fontSize: '11px' }}>
                <span
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                    color: 'var(--hub-text-muted)',
                  }}
                >
                  <span
                    style={{
                      width: '10px',
                      height: '10px',
                      background: '#3b82f6',
                      borderRadius: '2px',
                      display: 'inline-block',
                    }}
                  />{' '}
                  {t('crm.campaignDetail.opens')}
                </span>
                <span
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                    color: 'var(--hub-text-muted)',
                  }}
                >
                  <span
                    style={{
                      width: '10px',
                      height: '10px',
                      background: '#8b5cf6',
                      borderRadius: '2px',
                      display: 'inline-block',
                    }}
                  />{' '}
                  {t('crm.campaignDetail.clicks')}
                </span>
              </div>
            </Card>
          </>
        )}
        {!stats && !statsLoading && (
          <Card>
            <div style={{ color: 'var(--hub-text-muted)', fontSize: '13px' }}>
              {statsError
                ? 'Campaign statistics could not be loaded from the CRM service.'
                : 'No campaign statistics are available yet.'}
            </div>
          </Card>
        )}
      </div>
    </>
  );
};

export default CampaignDetail;
