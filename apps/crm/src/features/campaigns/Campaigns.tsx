import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { Plus, Mail, MessageSquare, Bell, Users, Send, Eye, TrendingUp } from 'lucide-react';
import { useCampaigns } from './useCampaigns';

const ACCENT = '#3b82f6';

type CampaignStatus = 'Draft' | 'Active' | 'Paused' | 'Completed';
type Channel = 'Email' | 'SMS' | 'Push';

interface Campaign {
  id: number;
  name: string;
  description: string;
  channel: Channel;
  status: CampaignStatus;
  audience: number;
  sent: number;
  openRate: string;
  clickRate: string;
  revenue: string;
  updatedAt: string;
}


const statusColor: Record<CampaignStatus, { bg: string; text: string; dot: string }> = {
  Draft:     { bg: 'rgba(148,163,184,0.12)', text: '#94a3b8', dot: '#94a3b8' },
  Active:    { bg: 'rgba(16,185,129,0.12)',  text: '#10b981', dot: '#10b981' },
  Paused:    { bg: 'rgba(245,158,11,0.12)',  text: '#f59e0b', dot: '#f59e0b' },
  Completed: { bg: 'rgba(59,130,246,0.12)',  text: ACCENT,    dot: ACCENT    },
};

const channelIcon: Record<Channel, React.ComponentType<{ size?: number }>> = {
  Email: Mail,
  SMS:   MessageSquare,
  Push:  Bell,
};

const channelColor: Record<Channel, string> = {
  Email: ACCENT,
  SMS:   '#10b981',
  Push:  '#8b5cf6',
};

const statusFilters: Array<'All' | CampaignStatus> = ['All', 'Active', 'Draft', 'Paused', 'Completed'];

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
      <span style={{ width: '5px', height: '5px', borderRadius: '50%', background: s.dot, display: 'inline-block' }} />
      {status}
    </span>
  );
}

function CampaignCard({ campaign, index, onClick }: { campaign: Campaign; index: number; onClick?: () => void }) {
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
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: '12px' }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px', flex: 1, minWidth: 0 }}>
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
            <ChannelIcon size={16} style={{ color: channelColor[campaign.channel] } as React.CSSProperties} />
          </div>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)', marginBottom: '2px' }}>
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
          { label: 'Audience', value: campaign.audience ? campaign.audience.toLocaleString() : '-', icon: Users },
          { label: 'Sent', value: campaign.sent ? campaign.sent.toLocaleString() : '-', icon: Send },
          { label: 'Open Rate', value: campaign.openRate, icon: Eye },
          { label: 'Click Rate', value: campaign.clickRate, icon: TrendingUp },
        ].map(({ label, value, icon: Icon }) => (
          <div
            key={label}
            style={{
              padding: '10px 12px',
              borderRadius: '8px',
              background: 'var(--hub-hover)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
              <Icon size={11} style={{ color: 'var(--hub-text-dim)' }} />
              <span style={{ fontSize: '10px', color: 'var(--hub-text-dim)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</span>
            </div>
            <span style={{ fontSize: '16px', fontWeight: 700, color: 'var(--hub-text)' }}>{value}</span>
          </div>
        ))}
      </div>

      {/* Footer */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>Updated {campaign.updatedAt}</span>
        <div style={{ display: 'flex', gap: '8px' }}>
          {campaign.status === 'Draft' && (
            <button
              style={{
                padding: '6px 14px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 600,
                background: ACCENT,
                color: '#fff',
                border: 'none',
                cursor: 'pointer',
              }}
            >
              Launch
            </button>
          )}
          {campaign.status === 'Active' && (
            <button
              style={{
                padding: '6px 14px',
                borderRadius: '7px',
                fontSize: '12px',
                fontWeight: 600,
                background: 'rgba(245,158,11,0.12)',
                color: '#f59e0b',
                border: '1px solid rgba(245,158,11,0.25)',
                cursor: 'pointer',
              }}
            >
              Pause
            </button>
          )}
          <button
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
        </div>
      </div>
    </motion.div>
  );
}

const Campaigns: React.FC = () => {
  const navigate = useNavigate();
  const { data: fetchedCampaigns } = useCampaigns();
  const campaigns: Campaign[] = (fetchedCampaigns as Campaign[]) ?? [];
  const [statusFilter, setStatusFilter] = useState<'All' | CampaignStatus>('All');

  const filtered = campaigns.filter((c) => statusFilter === 'All' || c.status === statusFilter);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>Campaigns</h2>
          <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
            {campaigns.filter(c => c.status === 'Active').length} active campaigns
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={() => navigate('/campaigns/sms')}
            style={{ display: 'flex', alignItems: 'center', gap: '6px', padding: '8px 14px', borderRadius: '8px', fontSize: '13px', fontWeight: 600, background: 'transparent', color: 'var(--hub-text)', border: '1px solid var(--hub-border)', cursor: 'pointer' }}
          >
            <MessageSquare size={14} /> New SMS
          </button>
          <button
            onClick={() => navigate('/campaigns/email')}
            style={{ display: 'flex', alignItems: 'center', gap: '6px', padding: '8px 16px', borderRadius: '8px', fontSize: '13px', fontWeight: 600, background: ACCENT, color: '#fff', border: 'none', cursor: 'pointer' }}
          >
            <Plus size={15} /> New email campaign
          </button>
        </div>
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
            key={f}
            onClick={() => setStatusFilter(f)}
            style={{
              padding: '6px 16px',
              borderRadius: '7px',
              fontSize: '12px',
              fontWeight: 500,
              cursor: 'pointer',
              border: 'none',
              background: statusFilter === f ? ACCENT : 'transparent',
              color: statusFilter === f ? '#fff' : 'var(--hub-text-muted)',
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
                {campaigns.filter(c => c.status === f).length}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Campaign cards */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        <AnimatePresence mode="popLayout">
          {filtered.map((c, i) => (
            <CampaignCard key={c.id} campaign={c} index={i} onClick={() => navigate(`/campaigns/${c.id}`)} />
          ))}
        </AnimatePresence>
        {filtered.length === 0 && (
          <div style={{ padding: '60px', textAlign: 'center', color: 'var(--hub-text-muted)', fontSize: '13px' }}>
            No campaigns in this status.
          </div>
        )}
      </div>
    </div>
  );
};

export default Campaigns;
