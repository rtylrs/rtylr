import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import {
  ArrowLeft,
  Mail,
  Phone,
  Calendar,
  Tag as TagIcon,
  Trash2,
  Download,
  Users,
} from 'lucide-react';
import {
  useGetCustomerFull,
  useCustomerTimeline,
  useUpdateCustomer,
  useDeleteCustomer,
  useExportCustomers,
} from '../../shared/hooks/queries/useCustomers';
import {
  useAwardPoints,
  useRedeemPoints,
  useLoyaltyConfig,
} from '../../shared/hooks/queries/useLoyalty';
import { Breadcrumbs, Button, Card, ConfirmModal, SkeletonList } from '../../shared/components/ui';
import {
  CustomerAvatar,
  TierBadge,
  PointsBadge,
  TagChip,
  NotesPanel,
  TimelineEntryRow,
  ConsentBanner,
  LoyaltyPointsInput,
} from '../../shared/components/crm';
import { getTier, pointsToNextTier } from '../../shared/lib/tiers';

const CustomerProfile: React.FC = () => {
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const customerId = Number(id);
  const { data: fetched, isLoading, isError } = useGetCustomerFull(customerId);
  const [timelineCursor, setTimelineCursor] = useState<string>();
  const { data: fetchedTimeline } = useCustomerTimeline(customerId, timelineCursor);
  const [timelineEntries, setTimelineEntries] = useState<
    NonNullable<typeof fetchedTimeline>['entries']
  >([]);
  const update = useUpdateCustomer();
  const remove = useDeleteCustomer();
  const award = useAwardPoints();
  const redeem = useRedeemPoints();
  const { data: loyaltyConfig } = useLoyaltyConfig();
  const exportData = useExportCustomers();
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [newTag, setNewTag] = useState('');

  const customer = fetched;
  useEffect(() => {
    if (!fetchedTimeline) return;
    setTimelineEntries((entries) =>
      timelineCursor ? [...entries, ...fetchedTimeline.entries] : fetchedTimeline.entries
    );
  }, [fetchedTimeline, timelineCursor]);
  useEffect(() => {
    setTimelineCursor(undefined);
    setTimelineEntries([]);
  }, [customerId]);
  const timeline = timelineEntries;

  if (isLoading) {
    return (
      <div style={{ maxWidth: '1100px' }}>
        <SkeletonList count={6} />
      </div>
    );
  }

  if (isError) {
    return (
      <div style={{ maxWidth: '1100px', textAlign: 'center', padding: '60px 20px' }}>
        <p style={{ fontSize: '14px', color: 'var(--hub-text-muted)' }}>
          Unable to load this customer. Please try again.
        </p>
      </div>
    );
  }

  if (!customer) {
    return (
      <div style={{ maxWidth: '1100px', textAlign: 'center', padding: '60px 20px' }}>
        <p style={{ fontSize: '14px', color: 'var(--hub-text-muted)' }}>Customer not found.</p>
      </div>
    );
  }

  const tier = getTier(customer.loyaltyPoints);
  const next = pointsToNextTier(customer.loyaltyPoints);
  const tags = customer.tags || [];
  // Missing consent means unknown, never opt-in. The backend now persists this map for new records.
  const consent = customer.consent || { email: false, sms: false, marketing: false };

  const handleAddTag = () => {
    if (!newTag.trim()) {
      toast.error('Tag name is required');
      return;
    }
    const tagName = newTag.trim();
    const updated = [...tags, tagName];
    update.mutate(
      { id: customer.id, data: { tags: updated } },
      {
        onSuccess: () => toast.success(`Tag "${tagName}" added`),
        onError: () => toast.error('Failed to add tag'),
      }
    );
    setNewTag('');
  };

  const handleRemoveTag = (tag: string) => {
    const updated = tags.filter((t) => t !== tag);
    update.mutate(
      { id: customer.id, data: { tags: updated } },
      {
        onSuccess: () => toast.success(`Tag "${tag}" removed`),
        onError: () => toast.error('Failed to remove tag'),
      }
    );
  };

  const handleNotesSave = (notes: string) => {
    update.mutate(
      { id: customer.id, data: { notes } },
      {
        onSuccess: () => toast.success('Notes saved'),
        onError: () => toast.error('Failed to save notes'),
      }
    );
  };

  const handleAward = (points: number, reason: string) => {
    if (!loyaltyConfig?.programEnabled) {
      toast.error('Loyalty program is currently disabled');
      return;
    }
    award.mutate(
      { customerId: customer.id, points, reason },
      {
        onSuccess: () => toast.success(`Awarded ${points} points`),
        onError: () => toast.error('Failed to award points'),
      }
    );
  };

  const handleDeduct = (points: number, reason: string) => {
    if (!loyaltyConfig?.programEnabled) {
      toast.error('Loyalty program is currently disabled');
      return;
    }
    redeem.mutate(
      { customerId: customer.id, points, reason },
      {
        onSuccess: () => toast.success(`Deducted ${points} points`),
        onError: () => toast.error('Failed to deduct points'),
      }
    );
  };

  const handleDelete = () => {
    remove.mutate(customer.id, {
      onSuccess: () => {
        toast.success('Customer deleted');
        navigate({ to: '/customers' });
      },
      onError: () => {
        toast.error('Failed to delete customer');
      },
    });
  };

  const handleExport = () => {
    exportData.mutate(customer.id, {
      onSuccess: (data) => {
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `customer-${customer.id}-export.json`;
        link.click();
        URL.revokeObjectURL(url);
        toast.success('Customer data exported');
      },
      onError: () => toast.error('Failed to export customer data'),
    });
  };

  return (
    <>
      <HelmetContainer title="Customer profile · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs items={[{ label: 'Customers', to: '/customers' }, { label: customer.name }]} />
        <button
          onClick={() => navigate({ to: '/customers' })}
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

        {/* Header card */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '14px',
            padding: '22px',
            marginBottom: '20px',
            display: 'flex',
            alignItems: 'center',
            gap: '18px',
          }}
        >
          <CustomerAvatar initials={customer.initials} points={customer.loyaltyPoints} size={72} />
          <div style={{ flex: 1 }}>
            <h2
              style={{
                fontSize: '22px',
                fontWeight: 700,
                color: 'var(--hub-text)',
                margin: '0 0 4px',
              }}
            >
              {customer.name}
            </h2>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '16px',
                color: 'var(--hub-text-muted)',
                fontSize: '13px',
                flexWrap: 'wrap',
              }}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <Mail size={12} />
                {customer.email}
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <Phone size={12} />
                {customer.phone}
              </span>
              {customer.createdAt && (
                <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <Calendar size={12} />
                  Since {customer.createdAt}
                </span>
              )}
            </div>
            <div style={{ display: 'flex', gap: '8px', marginTop: '10px' }}>
              <TierBadge tier={tier} size="md" />
              <PointsBadge points={customer.loyaltyPoints} trend="up" />
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            <Button variant="outline" onClick={handleExport} disabled={exportData.isPending}>
              <Download size={13} />
              {exportData.isPending ? 'Exporting…' : 'Export GDPR'}
            </Button>
            <Button
              variant="outline"
              onClick={() => navigate({ to: '/customers/merge?primary=' + customer.id })}
            >
              <Users size={13} />
              Merge
            </Button>
            <Button
              variant="outline"
              onClick={() => setConfirmDelete(true)}
              disabled={remove.isPending}
              style={{ color: '#ef4444', borderColor: 'rgba(239,68,68,0.3)' }}
            >
              <Trash2 size={13} />
              {remove.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </div>
        </motion.div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '20px' }}>
          {/* Left column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {/* Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px' }}>
              {[
                { label: 'Total Spent', value: `$${customer.totalSpent.toLocaleString()}` },
                { label: 'Visits', value: customer.visits },
                { label: 'Loyalty Points', value: customer.loyaltyPoints.toLocaleString() },
                { label: 'Last Visit', value: customer.lastVisit },
              ].map((s) => (
                <Card key={s.label}>
                  <div
                    style={{
                      fontSize: '10px',
                      color: 'var(--hub-text-dim)',
                      textTransform: 'uppercase',
                      letterSpacing: '0.05em',
                      fontWeight: 700,
                    }}
                  >
                    {s.label}
                  </div>
                  <div
                    style={{
                      fontSize: '18px',
                      fontWeight: 700,
                      color: 'var(--hub-text)',
                      marginTop: '4px',
                    }}
                  >
                    {s.value}
                  </div>
                </Card>
              ))}
            </div>

            {/* Tags */}
            <Card>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '10px',
                }}
              >
                <h4
                  style={{
                    fontSize: '12px',
                    fontWeight: 600,
                    color: 'var(--hub-text)',
                    margin: 0,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                  }}
                >
                  <TagIcon size={12} /> Tags
                </h4>
              </div>
              <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap', marginBottom: '10px' }}>
                {tags.map((t) => (
                  <TagChip key={t} label={t} onRemove={() => handleRemoveTag(t)} />
                ))}
                {tags.length === 0 && (
                  <span style={{ fontSize: '12px', color: 'var(--hub-text-dim)' }}>
                    No tags yet
                  </span>
                )}
              </div>
              <div style={{ display: 'flex', gap: '6px' }}>
                <input
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddTag()}
                  placeholder="Add tag..."
                  style={{
                    flex: 1,
                    height: '32px',
                    padding: '0 10px',
                    borderRadius: '6px',
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                    fontSize: '12px',
                    outline: 'none',
                  }}
                />
                <Button variant="primary" onClick={handleAddTag} style={{ padding: '6px 12px' }}>
                  Add
                </Button>
              </div>
            </Card>

            {/* Timeline */}
            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 12px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Timeline
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                {timeline.map((t) => (
                  <TimelineEntryRow key={t.id} entry={t} />
                ))}
              </div>
              {fetchedTimeline?.nextCursor && (
                <Button
                  variant="outline"
                  onClick={() => setTimelineCursor(fetchedTimeline.nextCursor)}
                  style={{ marginTop: '14px' }}
                >
                  Load more
                </Button>
              )}
            </Card>
          </div>

          {/* Right column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 8px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Tier Progress
              </h4>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '6px',
                  fontSize: '12px',
                }}
              >
                <span style={{ color: 'var(--hub-text)' }}>
                  {customer.loyaltyPoints.toLocaleString()} pts
                </span>
                <span style={{ color: 'var(--hub-text-muted)' }}>
                  {next !== null ? `${next} to next` : 'Max tier'}
                </span>
              </div>
              <div
                style={{
                  height: '6px',
                  background: 'var(--hub-hover)',
                  borderRadius: '3px',
                  overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    width:
                      next === null
                        ? '100%'
                        : `${Math.min(
                            100,
                            (customer.loyaltyPoints / (customer.loyaltyPoints + (next || 1))) * 100
                          )}%`,
                    height: '100%',
                    background: '#3b82f6',
                  }}
                />
              </div>
            </Card>

            {loyaltyConfig?.programEnabled && (
              <LoyaltyPointsInput onAward={handleAward} onDeduct={handleDeduct} />
            )}
            <ConsentBanner consent={consent} />
            <NotesPanel notes={customer.notes || ''} onSave={handleNotesSave} />
          </div>
        </div>

        <ConfirmModal
          open={confirmDelete}
          onClose={() => setConfirmDelete(false)}
          onConfirm={handleDelete}
          title="Delete customer?"
          description={`Are you sure you want to delete ${customer.name}? This action cannot be undone.`}
          confirmLabel="Delete"
          variant="danger"
        />
      </div>
    </>
  );
};

export default CustomerProfile;
