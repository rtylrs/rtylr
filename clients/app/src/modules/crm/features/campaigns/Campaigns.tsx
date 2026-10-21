import React, { useEffect, useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import {
  Plus,
  Mail,
  MessageSquare,
  Bell,
  Megaphone,
  Users,
  Send,
  Eye,
  TrendingUp,
  Trash2,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListCampaigns,
  useUpdateCampaignStatus,
  useDeleteCampaign,
} from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import {
  calculateOpenRate,
  type Campaign as CampaignPayload,
} from '../../shared/schemas/campaigns';
import { ConfirmModal } from '../../shared/components/ui';
import { canCrmPermission } from '../../shared/lib/crmPermissions';
import { canEditCampaign, hasVerifiedCampaignAudience } from './campaignActions';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';

const ACCENT = 'var(--crm-accent)';

type CampaignStatus =
  'Draft' | 'Scheduled' | 'Active' | 'Paused' | 'Completed' | 'Sent' | 'Cancelled';
type Channel = 'Email' | 'SMS' | 'Push';

interface Campaign {
  id: number;
  name: string;
  description: string;
  channel: Channel;
  status: CampaignStatus;
  audienceSegmentId?: number | null;
  audience: number;
  sent: number;
  openRate: string;
  clickRate: string;
  revenue: string;
  updatedAt: string;
}

const TYPE_TO_CHANNEL: Record<string, Channel> = {
  email: 'Email',
  sms: 'SMS',
  push: 'Push',
};

const STATUS_TO_DISPLAY: Record<string, CampaignStatus> = {
  draft: 'Draft',
  active: 'Active',
  paused: 'Paused',
  completed: 'Completed',
  scheduled: 'Scheduled',
  sent: 'Sent',
  cancelled: 'Cancelled',
  canceled: 'Cancelled',
  DRAFT: 'Draft',
  ACTIVE: 'Active',
  PAUSED: 'Paused',
  COMPLETED: 'Completed',
  SCHEDULED: 'Scheduled',
  SENT: 'Sent',
  CANCELLED: 'Cancelled',
  CANCELED: 'Cancelled',
};

/** Projects a backend campaign payload onto the local display shape. */
function toDisplayCampaign(c: CampaignPayload, audienceBySegment: Map<number, number>): Campaign {
  const segmentAudience =
    c.audienceSegmentId !== null && c.audienceSegmentId !== undefined
      ? audienceBySegment.get(c.audienceSegmentId)
      : undefined;
  return {
    id: c.id,
    name: c.name,
    description: c.subject ?? '',
    channel: TYPE_TO_CHANNEL[c.type] ?? 'Email',
    status: STATUS_TO_DISPLAY[c.status] ?? 'Draft',
    audienceSegmentId: c.audienceSegmentId,
    audience: segmentAudience ?? c.audienceCount ?? c.sentCount,
    sent: c.sentCount,
    openRate: `${Math.round(calculateOpenRate(c))}%`,
    clickRate: '-',
    revenue: '-',
    updatedAt: c.updatedAt ?? '',
  };
}

const statusColor: Record<CampaignStatus, { bg: string; text: string; dot: string }> = {
  Draft: { bg: 'rgba(148,163,184,0.12)', text: '#94a3b8', dot: '#94a3b8' },
  Scheduled: { bg: 'rgba(139,92,246,0.12)', text: '#8b5cf6', dot: '#8b5cf6' },
  Active: { bg: 'rgba(16,185,129,0.12)', text: '#10b981', dot: '#10b981' },
  Paused: { bg: 'rgba(245,158,11,0.12)', text: '#f59e0b', dot: '#f59e0b' },
  Completed: { bg: 'rgba(59,130,246,0.12)', text: ACCENT, dot: ACCENT },
  Sent: { bg: 'rgba(16,185,129,0.12)', text: '#10b981', dot: '#10b981' },
  Cancelled: { bg: 'rgba(239,68,68,0.12)', text: '#ef4444', dot: '#ef4444' },
};

const channelIcon: Record<
  Channel,
  React.ComponentType<{ size?: number; style?: React.CSSProperties }>
> = {
  Email: Mail,
  SMS: MessageSquare,
  Push: Bell,
};

const channelColor: Record<Channel, string> = {
  Email: ACCENT,
  SMS: '#10b981',
  Push: '#8b5cf6',
};

const statusFilters: Array<'All' | CampaignStatus> = [
  'All',
  'Active',
  'Draft',
  'Scheduled',
  'Paused',
  'Completed',
  'Sent',
  'Cancelled',
];

function StatusBadge({ status }: { status: CampaignStatus }) {
  const s = statusColor[status];
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '5px',
        padding: '2px 10px',
        borderRadius: '9999px',
        fontSize: '11px',
        fontWeight: 600,
        background: s.bg,
        color: s.text,
      }}
    >
      <span
        style={{
          width: '5px',
          height: '5px',
          borderRadius: '50%',
          background: s.dot,
          display: 'inline-block',
        }}
      />
      {status}
    </span>
  );
}

function CampaignCard({
  campaign,
  index,
  onClick,
  onLaunch,
  onPause,
  onEdit,
  onDelete,
  statusPending,
  deletePending,
  canSend,
  canManage,
}: {
  campaign: Campaign;
  index: number;
  onClick?: () => void;
  onLaunch: () => void;
  onPause: () => void;
  onEdit: () => void;
  onDelete: () => void;
  statusPending: boolean;
  deletePending: boolean;
  canSend: boolean;
  canManage: boolean;
}) {
  const { t } = useTranslation();
  const ChannelIcon = channelIcon[campaign.channel];

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      transition={{ delay: index * 0.05, duration: 0.2 }}
      onClick={onClick}
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
        cursor: onClick ? 'pointer' : 'default',
      }}
    >
      {/* Card header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
          gap: '12px',
        }}
      >
        <div
          style={{ display: 'flex', alignItems: 'flex-start', gap: '12px', flex: 1, minWidth: 0 }}
        >
          <div
            style={{
              width: '36px',
              height: '36px',
              borderRadius: '10px',
              background: `${channelColor[campaign.channel]}15`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            <ChannelIcon size={16} style={{ color: channelColor[campaign.channel] }} />
          </div>
          <div style={{ minWidth: 0 }}>
            <div
              style={{
                fontSize: '14px',
                fontWeight: 600,
                color: 'var(--hub-text)',
                marginBottom: '2px',
              }}
            >
              {campaign.name}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--hub-text-muted)', lineHeight: 1.4 }}>
              {campaign.description}
            </div>
          </div>
        </div>
        <StatusBadge status={campaign.status} />
      </div>

      {/* Stats row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px' }}>
        {[
          {
            id: 'audience',
            label: t('crm.campaigns.colAudience'),
            value: campaign.audience ? campaign.audience.toLocaleString() : '-',
            icon: Users,
          },
          {
            id: 'sent',
            label: t('crm.campaigns.colSent'),
            value: campaign.sent ? campaign.sent.toLocaleString() : '-',
            icon: Send,
          },
          {
            id: 'openRate',
            label: t('crm.campaigns.colOpenRate'),
            value: campaign.openRate,
            icon: Eye,
          },
          {
            id: 'clickRate',
            label: t('crm.campaigns.colClickRate'),
            value: campaign.clickRate,
            icon: TrendingUp,
          },
        ].map(({ id, label, value, icon: Icon }) => (
          <div
            key={id}
            style={{
              padding: '10px 12px',
              borderRadius: '8px',
              background: 'var(--hub-hover)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
              <Icon size={11} style={{ color: 'var(--hub-text-dim)' }} />
              <span
                style={{
                  fontSize: '10px',
                  color: 'var(--hub-text-dim)',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {label}
              </span>
            </div>
            <span style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>
              {value}
            </span>
          </div>
        ))}
      </div>

      {/* Footer */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>
          Updated {campaign.updatedAt}
        </span>
        <div style={{ display: 'flex', gap: '8px' }} onClick={(e) => e.stopPropagation()}>
          {campaign.status === 'Draft' && canSend && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onLaunch();
              }}
              disabled={statusPending}
              style={{
                padding: '6px 14px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 600,
                background: 'var(--crm-action)',
                color: 'var(--crm-on-action)',
                border: 'none',
                cursor: statusPending ? 'not-allowed' : 'pointer',
                opacity: statusPending ? 0.6 : 1,
              }}
            >
              {statusPending ? 'Activating…' : 'Activate'}
            </button>
          )}
          {campaign.status === 'Active' && canManage && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onPause();
              }}
              disabled={statusPending}
              style={{
                padding: '6px 14px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 600,
                background: 'rgba(245,158,11,0.12)',
                color: '#f59e0b',
                border: '1px solid rgba(245,158,11,0.25)',
                cursor: statusPending ? 'not-allowed' : 'pointer',
                opacity: statusPending ? 0.6 : 1,
              }}
            >
              {statusPending ? 'Pausing…' : 'Pause'}
            </button>
          )}
          {/* Viewing the report of a finished campaign is read-only; only editing a live one is gated. */}
          {(campaign.status === 'Completed' ||
            campaign.status === 'Sent' ||
            campaign.status === 'Cancelled' ||
            canManage) && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onEdit();
              }}
              style={{
                padding: '6px 14px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 500,
                background: 'var(--hub-hover)',
                color: 'var(--hub-text-muted)',
                border: '1px solid var(--hub-border)',
                cursor: 'pointer',
              }}
            >
              {campaign.status === 'Completed' ? 'View Report' : 'Edit'}
            </button>
          )}
          {canManage && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
              disabled={deletePending}
              aria-label={t('crm.campaigns.deleteCampaign')}
              style={{
                padding: '6px 10px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 500,
                background: 'rgba(239,68,68,0.08)',
                color: '#ef4444',
                border: '1px solid rgba(239,68,68,0.25)',
                cursor: deletePending ? 'not-allowed' : 'pointer',
                opacity: deletePending ? 0.6 : 1,
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <Trash2 size={12} />
            </button>
          )}
        </div>
      </div>
    </motion.div>
  );
}

const Campaigns: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const canCreateCampaigns = canCrmPermission('crm:campaigns:create');
  const canManageCampaigns = canCrmPermission('crm:campaigns:manage');
  const canSendCampaigns = canCrmPermission('crm:campaigns:send');
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  // Set once a "Load more" page comes back empty.
  const [exhausted, setExhausted] = useState(false);
  const [loadedCampaigns, setLoadedCampaigns] = useState<CampaignPayload[]>([]);
  const campaignsQuery = useListCampaigns(pageSize, offset);
  const { data: fetchedCampaigns, isFetching, refetch } = campaignsQuery;
  const isError = hasQueryFailed(campaignsQuery);
  const campaignsLoading = campaignsQuery.isPending && !isError;
  const segmentsQuery = useListSegments();
  const { data: fetchedSegments } = segmentsQuery;
  const segmentsError = hasQueryFailed(segmentsQuery);
  const updateStatus = useUpdateCampaignStatus();
  const removeCampaign = useDeleteCampaign();
  const resetPagination = (): void => {
    setOffset(0);
    setExhausted(false);
    setLoadedCampaigns([]);
  };
  useEffect(() => {
    if (!fetchedCampaigns) return;
    if (offset > 0 && fetchedCampaigns.length === 0) {
      // Exact page-size multiple: the extra page came back empty — stop
      // offering "Load more".
      setExhausted(true);
    }
    setLoadedCampaigns((previous) => {
      if (offset === 0) return fetchedCampaigns;
      const byId = new Map(previous.map((campaign) => [campaign.id, campaign]));
      fetchedCampaigns.forEach((campaign) => byId.set(campaign.id, campaign));
      return Array.from(byId.values());
    });
  }, [fetchedCampaigns, offset]);

  const hasMore =
    !exhausted && !isError && loadedCampaigns.length > 0 && loadedCampaigns.length % pageSize === 0;
  const audienceBySegment = useMemo(
    () =>
      new Map(
        (fetchedSegments ?? []).map((segment) => [Number(segment.id), segment.customerCount])
      ),
    [fetchedSegments]
  );
  const campaigns: Campaign[] = useMemo(
    () => loadedCampaigns.map((campaign) => toDisplayCampaign(campaign, audienceBySegment)),
    [audienceBySegment, loadedCampaigns]
  );
  const [statusFilter, setStatusFilter] = useState<'All' | CampaignStatus>('All');
  const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);

  const filtered = campaigns.filter((c) => statusFilter === 'All' || c.status === statusFilter);

  const handleLaunch = (id: number): void => {
    const campaign = campaigns.find((c) => c.id === id);
    if (segmentsError || fetchedSegments === undefined) {
      toast.error(
        segmentsError
          ? 'Cannot verify the campaign audience right now'
          : 'Audience data is still loading'
      );
      return;
    }
    if (!campaign || !hasVerifiedCampaignAudience(campaign, audienceBySegment)) {
      toast.error(t('crm.campaigns.noEligibleRecipients'));
      return;
    }
    updateStatus.mutate({ id, status: 'active' }, { onSuccess: resetPagination });
  };

  const handlePause = (id: number): void => {
    updateStatus.mutate({ id, status: 'paused' }, { onSuccess: resetPagination });
  };

  const handleEdit = (c: Campaign): void => {
    if (!canEditCampaign(c)) {
      navigate({ to: `/crm/campaigns/${c.id}` });
      return;
    }
    const target = c.channel === 'SMS' ? '/crm/campaigns/sms' : '/crm/campaigns/email';
    navigate({ to: `${target}?id=${c.id}` });
  };

  const handleDelete = (id: number): void => {
    removeCampaign.mutate(id, {
      onSuccess: () => {
        resetPagination();
        setConfirmDeleteId(null);
      },
      onError: () => setConfirmDeleteId(null),
    });
  };

  const deleteTarget = campaigns.find((c) => c.id === confirmDeleteId);

  return (
    <>
      <HelmetContainer title={`${t('crm.campaigns.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              {t('crm.campaigns.pageTitle')}
            </h2>
            <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
              {t('crm.campaigns.activeCount', {
                count: campaigns.filter((c) => c.status === 'Active').length,
              })}
            </p>
          </div>
          {canCreateCampaigns && (
            <div style={{ display: 'flex', gap: '8px' }}>
              {/* SMS campaigns return once an SMS provider is connected;
                  until then nothing could deliver them. */}
              <button
                type="button"
                onClick={() => navigate({ to: '/crm/campaigns/email' })}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                  padding: '8px 16px',
                  borderRadius: '8px',
                  fontSize: '13px',
                  fontWeight: 600,
                  background: 'var(--crm-action)',
                  color: 'var(--crm-on-action)',
                  border: 'none',
                  cursor: 'pointer',
                }}
              >
                <Plus size={15} /> New email campaign
              </button>
            </div>
          )}
        </div>

        {/* Filter tabs */}
        <div
          style={{
            display: 'flex',
            gap: '4px',
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '10px',
            padding: '4px',
            width: 'fit-content',
          }}
        >
          {statusFilters.map((f) => (
            <button
              type="button"
              key={f}
              onClick={() => setStatusFilter(f)}
              style={{
                padding: '6px 16px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 500,
                cursor: 'pointer',
                border: 'none',
                background: statusFilter === f ? 'var(--crm-action)' : 'transparent',
                color: statusFilter === f ? 'var(--crm-on-action)' : 'var(--hub-text-muted)',
                transition: 'all 0.15s',
              }}
            >
              {f}
              {f !== 'All' && (
                <span
                  style={{
                    marginLeft: '6px',
                    padding: '1px 5px',
                    borderRadius: '9999px',
                    fontSize: '10px',
                    background: statusFilter === f ? 'rgba(255,255,255,0.25)' : 'var(--hub-hover)',
                  }}
                >
                  {campaigns.filter((c) => c.status === f).length}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Campaign cards */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {isError && (
            <div
              role="alert"
              style={{
                padding: '16px',
                textAlign: 'center',
                color: 'var(--hub-text-muted)',
                border: '1px solid var(--hub-border)',
                borderRadius: '8px',
              }}
            >
              <div>{t('crm.campaigns.loadFailed')}</div>
              <button
                type="button"
                onClick={() => void refetch()}
                style={{
                  marginTop: '8px',
                  padding: '6px 12px',
                  borderRadius: '7px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-card)',
                  color: 'var(--hub-text)',
                  cursor: 'pointer',
                }}
              >
                {t('crm.actions.refresh')}
              </button>
            </div>
          )}
          {campaignsLoading && campaigns.length === 0 && (
            <div style={{ padding: '40px', textAlign: 'center', color: 'var(--hub-text-muted)' }}>
              Loading campaigns…
            </div>
          )}
          <AnimatePresence mode="popLayout">
            {filtered.map((c, i) => (
              <CampaignCard
                key={c.id}
                campaign={c}
                index={i}
                onClick={() => navigate({ to: `/crm/campaigns/${c.id}` })}
                onLaunch={() => handleLaunch(c.id)}
                onPause={() => handlePause(c.id)}
                onEdit={() => handleEdit(c)}
                onDelete={() => setConfirmDeleteId(c.id)}
                statusPending={updateStatus.isPending}
                deletePending={removeCampaign.isPending}
                canSend={canSendCampaigns}
                canManage={canManageCampaigns}
              />
            ))}
          </AnimatePresence>
          {filtered.length === 0 && !campaignsLoading && !isError && (
            <EmptyState
              icon={Megaphone}
              title={
                hasMore
                  ? t('crm.campaigns.loadMoreToSearch')
                  : statusFilter !== 'All'
                    ? t('crm.campaigns.emptyStatus')
                    : t('crm.campaigns.emptyAllTitle')
              }
              description={
                !hasMore && statusFilter === 'All'
                  ? t('crm.campaigns.emptyAllDescription')
                  : undefined
              }
              action={
                hasMore
                  ? {
                      label: t('crm.campaigns.loadMore'),
                      onClick: () => setOffset((current) => current + pageSize),
                    }
                  : statusFilter !== 'All'
                    ? {
                        label: t('crm.campaigns.viewAllCampaigns'),
                        onClick: () => setStatusFilter('All'),
                      }
                    : canCreateCampaigns
                      ? {
                          label: t('crm.campaigns.newEmailCampaign'),
                          onClick: () => navigate({ to: '/crm/campaigns/email' }),
                        }
                      : { label: t('crm.actions.refresh'), onClick: () => refetch() }
              }
            />
          )}
          {hasMore && (
            <div style={{ display: 'flex', justifyContent: 'center', padding: '4px 0 8px' }}>
              <button
                type="button"
                onClick={() => setOffset((current) => current + pageSize)}
                disabled={isFetching}
                style={{
                  padding: '8px 14px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-card)',
                  color: 'var(--hub-text)',
                  cursor: isFetching ? 'not-allowed' : 'pointer',
                  opacity: isFetching ? 0.6 : 1,
                }}
              >
                {isFetching ? 'Loading…' : 'Load more campaigns'}
              </button>
            </div>
          )}
        </div>

        <ConfirmModal
          open={confirmDeleteId !== null}
          onClose={() => setConfirmDeleteId(null)}
          onConfirm={() => confirmDeleteId !== null && handleDelete(confirmDeleteId)}
          title={t('crm.campaigns.deleteConfirmTitle')}
          description={
            deleteTarget
              ? t('crm.campaigns.deleteConfirmDescription', { name: deleteTarget.name })
              : ''
          }
          confirmLabel={t('crmCommon.delete')}
          variant="danger"
        />
      </div>
    </>
  );
};

export default Campaigns;
