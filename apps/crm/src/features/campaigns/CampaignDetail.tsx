import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { ArrowLeft, Bell, Circle, Mail, MessageSquare } from 'lucide-react';
import { useGetCampaign, useCampaignStats } from '../../shared/hooks/queries/useCampaigns';
import { CAMPAIGN_STATUS_LABELS, isCampaignStatus } from '../../shared/schemas/campaigns';
import { Breadcrumbs, Card, SkeletonList, EmptyState } from '../../shared/components/ui';
import { CampaignStatsCard } from '../../shared/components/crm';
import { normalizeCampaignChannel } from '../../shared/lib/campaignUtils';

const CampaignDetail: React.FC = () => {
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
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
  const channel = normalizeCampaignChannel(campaign?.type);
  const ChannelIcon =
    channel === 'sms'
      ? MessageSquare
      : channel === 'push'
        ? Bell
        : channel === 'unknown'
          ? Circle
          : Mail;
  const channelLabel =
    channel === 'sms'
      ? 'SMS'
      : channel === 'push'
        ? 'Push'
        : channel === 'unknown'
          ? 'Unknown'
          : 'Email';
  const normalizedStatus = campaign?.status.toLowerCase() ?? '';
  const campaignIsActive = normalizedStatus === 'active';
  const campaignStatusLabel = isCampaignStatus(normalizedStatus)
    ? CAMPAIGN_STATUS_LABELS[normalizedStatus]
    : campaign?.status;

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
          icon={Mail}
          title="Campaign not found"
          description="This campaign could not be loaded."
        />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Campaign detail · rtylr" />
      <div style={{ maxWidth: '1200px' }}>
        <Breadcrumbs items={[{ label: 'Campaigns', to: '/campaigns' }, { label: campaign.name }]} />
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
          <div
            className="crm-entity-header"
            style={{ display: 'flex', alignItems: 'center', gap: '14px' }}
          >
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
              <ChannelIcon size={20} aria-hidden="true" />
            </div>
            <div className="crm-entity-header__content" style={{ flex: 1 }}>
              <h2
                style={{ fontSize: '20px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}
              >
                {campaign.name}
              </h2>
              <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
                {channelLabel}
                {campaign.subject ? ` · ${campaign.subject}` : ''}
              </p>
            </div>
            <span
              style={{
                padding: '4px 12px',
                borderRadius: '9999px',
                fontSize: '11px',
                fontWeight: 600,
                background: campaignIsActive ? 'rgba(16,185,129,0.12)' : 'rgba(148,163,184,0.12)',
                color: campaignIsActive ? '#10b981' : '#94a3b8',
              }}
            >
              {campaignStatusLabel}
            </span>
          </div>
        </Card>

        {stats && (
          <>
            <div
              className="crm-stats-grid--six"
              style={{ display: 'grid', gap: '12px', marginBottom: '20px' }}
            >
              <CampaignStatsCard
                label="Sent"
                value={stats.sent.toLocaleString()}
                subValue={`${stats.delivered.toLocaleString()} delivered`}
              />
              <CampaignStatsCard
                label="Delivered"
                value={`${pct(stats.delivered, stats.sent)}%`}
                subValue={`${stats.delivered.toLocaleString()}`}
                color="#10b981"
              />
              <CampaignStatsCard
                label="Opened"
                value={`${pct(stats.opened, stats.delivered)}%`}
                subValue={`${stats.opened.toLocaleString()} opens`}
                sparkline={stats.timeline.map((t) => t.opens)}
                color="#3b82f6"
              />
              <CampaignStatsCard
                label="Clicked"
                value={`${pct(stats.clicked, stats.delivered)}%`}
                subValue={`${stats.clicked.toLocaleString()} clicks`}
                sparkline={stats.timeline.map((t) => t.clicks)}
                color="#8b5cf6"
              />
              <CampaignStatsCard
                label="Unsubscribed"
                value={
                  stats.unsubscribed === undefined
                    ? 'Unavailable'
                    : stats.unsubscribed.toLocaleString()
                }
                subValue={
                  stats.unsubscribed === undefined
                    ? 'Data unavailable'
                    : `${pct(stats.unsubscribed, stats.delivered)}%`
                }
                color="#ef4444"
              />
              <CampaignStatsCard
                label="Revenue"
                value={
                  stats.revenue === undefined ? 'Unavailable' : `$${stats.revenue.toLocaleString()}`
                }
                subValue={stats.revenue === undefined ? 'Data unavailable' : 'from campaign'}
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
                Engagement by day
              </h4>
              {stats.timeline.length === 0 ? (
                <p
                  style={{
                    color: 'var(--hub-text-muted)',
                    fontSize: '13px',
                    margin: 0,
                    padding: '36px 0',
                    textAlign: 'center',
                  }}
                >
                  Day-by-day engagement data is not available yet.
                </p>
              ) : (
                <>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'flex-end',
                      gap: '10px',
                      minHeight: '160px',
                      overflowX: 'auto',
                    }}
                  >
                    {stats.timeline.map((t, i) => {
                      const max = Math.max(1, ...stats.timeline.map((x) => x.opens));
                      const openH = (t.opens / max) * 130;
                      const clickH = (t.clicks / max) * 130;
                      return (
                        <div
                          key={[t.day, i].join('-')}
                          style={{
                            flex: '1 0 42px',
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
                  <div
                    style={{ display: 'flex', gap: '16px', marginTop: '16px', fontSize: '11px' }}
                  >
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
                      Opens
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
                      Clicks
                    </span>
                  </div>
                </>
              )}
            </Card>
          </>
        )}
        {statsLoading && (
          <Card>
            <SkeletonList count={3} />
          </Card>
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
