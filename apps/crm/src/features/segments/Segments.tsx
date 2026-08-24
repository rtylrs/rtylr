import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import { useUserStore } from '@voxire/auth';
import {
  Users,
  Star,
  AlertTriangle,
  Clock,
  TrendingUp,
  Plus,
  ChevronRight,
  X,
  Filter,
  Trash2,
} from 'lucide-react';
import {
  useListSegments,
  useDeleteSegment,
  useCreateSegment,
} from '../../shared/hooks/queries/useSegments';
import { ConfirmModal } from '../../shared/components/ui';
import { canCrmPermission } from '../../shared/lib/crmPermissions';
import {
  getSegmentRuleOperators,
  SEGMENT_RULE_FIELDS,
  SEGMENT_RULE_VALUES,
  serializeSegmentRule,
  type SegmentRule,
} from '../../shared/lib/segmentRules';

const ACCENT = '#3b82f6';

interface Segment {
  id: string;
  name: string;
  description: string;
  customerCount: number;
  icon: React.ComponentType<{ size?: number }>;
  color: string;
  rules: string[];
  trend: string;
  trendUp: boolean;
}

function RuleBuilder({
  onClose,
  onSave,
}: {
  onClose: () => void;
  onSave: (name: string, rules: SegmentRule[]) => void;
}) {
  const [rules, setRules] = useState<SegmentRule[]>([
    { id: 1, field: 'totalSpent', operator: '>', value: '500' },
  ]);
  const [segmentName, setSegmentName] = useState('');

  const addRule = () => {
    setRules((prev) => [...prev, { id: Date.now(), field: 'tier', operator: '=', value: 'Gold' }]);
  };

  const removeRule = (id: number) => setRules((prev) => prev.filter((r) => r.id !== id));

  const updateRule = (id: number, key: 'field' | 'operator' | 'value', val: string) => {
    setRules((prev) =>
      prev.map((r) => {
        if (r.id !== id) return r;
        if (key === 'field') {
          const nextField = SEGMENT_RULE_FIELDS.find((field) => field.value === val)?.value;
          if (!nextField) return r;
          const nextOperators = getSegmentRuleOperators(nextField);
          return {
            ...r,
            field: nextField,
            operator: nextOperators[0],
            value: SEGMENT_RULE_VALUES[nextField][0],
          };
        }
        return { ...r, [key]: val } as SegmentRule;
      })
    );
  };

  const selectStyle: React.CSSProperties = {
    height: '32px',
    padding: '0 10px',
    borderRadius: '7px',
    border: '1px solid var(--hub-border)',
    background: 'var(--hub-input-bg)',
    color: 'var(--hub-text)',
    fontSize: '12px',
    cursor: 'pointer',
    outline: 'none',
  };

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.97 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.97 }}
      transition={{ duration: 0.16 }}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
      }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'var(--hub-surface)',
          border: '1px solid var(--hub-border)',
          borderRadius: '16px',
          padding: '24px',
          width: '540px',
          maxWidth: '90vw',
          display: 'flex',
          flexDirection: 'column',
          gap: '20px',
          boxShadow: 'var(--hub-shadow)',
        }}
      >
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Filter size={15} style={{ color: ACCENT }} />
            <span style={{ fontSize: '15px', fontWeight: 700, color: 'var(--hub-text)' }}>
              Segment Builder
            </span>
          </div>
          <button
            aria-label="Close"
            type="button"
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
              padding: '4px',
            }}
          >
            <X size={16} />
          </button>
        </div>

        {/* Name input */}
        <div>
          <label
            style={{
              display: 'block',
              fontSize: '11px',
              fontWeight: 600,
              color: 'var(--hub-text-muted)',
              marginBottom: '6px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
            }}
          >
            Segment Name
          </label>
          <input
            value={segmentName}
            onChange={(e) => setSegmentName(e.target.value)}
            placeholder="e.g. High-Value Silver Members"
            style={{
              width: '100%',
              height: '36px',
              padding: '0 12px',
              borderRadius: '8px',
              border: '1px solid var(--hub-border)',
              background: 'var(--hub-input-bg)',
              color: 'var(--hub-text)',
              fontSize: '13px',
              outline: 'none',
              boxSizing: 'border-box',
            }}
          />
        </div>

        {/* Rules */}
        <div>
          <label
            style={{
              display: 'block',
              fontSize: '11px',
              fontWeight: 600,
              color: 'var(--hub-text-muted)',
              marginBottom: '8px',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
            }}
          >
            Rules (ALL must match)
          </label>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {rules.map((rule, i) => (
              <div key={rule.id} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                {i > 0 && (
                  <span
                    style={{
                      fontSize: '11px',
                      fontWeight: 600,
                      color: ACCENT,
                      width: '28px',
                      flexShrink: 0,
                    }}
                  >
                    AND
                  </span>
                )}
                {i === 0 && <div style={{ width: '28px', flexShrink: 0 }} />}
                <select
                  style={selectStyle}
                  value={rule.field}
                  onChange={(e) => updateRule(rule.id, 'field', e.target.value)}
                >
                  {SEGMENT_RULE_FIELDS.map((f) => (
                    <option key={f.value} value={f.value}>
                      {f.label}
                    </option>
                  ))}
                </select>
                <select
                  style={selectStyle}
                  value={rule.operator}
                  onChange={(e) => updateRule(rule.id, 'operator', e.target.value)}
                >
                  {getSegmentRuleOperators(rule.field).map((op) => (
                    <option key={op}>{op}</option>
                  ))}
                </select>
                <select
                  style={selectStyle}
                  value={rule.value}
                  onChange={(e) => updateRule(rule.id, 'value', e.target.value)}
                >
                  {SEGMENT_RULE_VALUES[rule.field].map((v) => (
                    <option key={v}>{v}</option>
                  ))}
                </select>
                {rules.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeRule(rule.id)}
                    style={{
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      color: 'var(--hub-text-dim)',
                      padding: '4px',
                      flexShrink: 0,
                    }}
                  >
                    <X size={13} />
                  </button>
                )}
              </div>
            ))}
          </div>
          <button
            aria-label="Add rule"
            type="button"
            onClick={addRule}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '5px',
              marginTop: '10px',
              padding: '6px 10px',
              borderRadius: '7px',
              fontSize: '12px',
              fontWeight: 500,
              background: 'transparent',
              border: '1px dashed var(--hub-border)',
              color: 'var(--hub-text-muted)',
              cursor: 'pointer',
            }}
          >
            <Plus size={12} />
            Add Rule
          </button>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', paddingTop: '4px' }}>
          <button
            type="button"
            onClick={onClose}
            style={{
              padding: '8px 16px',
              borderRadius: '8px',
              fontSize: '13px',
              fontWeight: 500,
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              border: '1px solid var(--hub-border)',
              cursor: 'pointer',
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => onSave(segmentName, rules)}
            style={{
              padding: '8px 16px',
              borderRadius: '8px',
              fontSize: '13px',
              fontWeight: 600,
              background: ACCENT,
              color: '#fff',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            Create Segment
          </button>
        </div>
      </div>
    </motion.div>
  );
}

function SegmentCard({
  segment,
  index,
  onClick,
  onDelete,
  deletePending,
  canDelete,
}: {
  segment: Segment;
  index: number;
  onClick?: () => void;
  onDelete: () => void;
  deletePending: boolean;
  canDelete: boolean;
}) {
  const Icon = segment.icon;
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.06, duration: 0.2 }}
      onClick={onClick}
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '14px',
        cursor: onClick ? 'pointer' : 'default',
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <div
            style={{
              width: '36px',
              height: '36px',
              borderRadius: '10px',
              background: `${segment.color}15`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            <Icon size={17} {...({ style: { color: segment.color } } as any)} />
          </div>
          <div>
            <div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)' }}>
              {segment.name}
            </div>
            <div style={{ fontSize: '11px', color: segment.trendUp ? '#10b981' : '#ef4444' }}>
              {segment.trend}
            </div>
          </div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--hub-text)' }}>
            {segment.customerCount.toLocaleString()}
          </div>
          <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>customers</div>
        </div>
      </div>

      {/* Description */}
      <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', lineHeight: 1.5, margin: 0 }}>
        {segment.description}
      </p>

      {/* Rules preview */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        {segment.rules.map((rule) => (
          <div
            key={rule}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '5px 10px',
              borderRadius: '6px',
              background: 'var(--hub-hover)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <span
              style={{
                width: '4px',
                height: '4px',
                borderRadius: '50%',
                background: segment.color,
                flexShrink: 0,
              }}
            />
            <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>{rule}</span>
          </div>
        ))}
      </div>

      {/* Actions */}
      <div style={{ display: 'flex', gap: '6px' }} onClick={(e) => e.stopPropagation()}>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onClick?.();
          }}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flex: 1,
            padding: '8px 12px',
            borderRadius: '8px',
            fontSize: '12px',
            fontWeight: 600,
            background: `${segment.color}10`,
            color: segment.color,
            border: `1px solid ${segment.color}25`,
            cursor: 'pointer',
            transition: 'background 0.15s',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.background = `${segment.color}20`;
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.background = `${segment.color}10`;
          }}
        >
          <span>View Customers</span>
          <ChevronRight size={13} />
        </button>
        {canDelete && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            disabled={deletePending}
            aria-label="Delete segment"
            style={{
              padding: '6px 10px',
              borderRadius: '8px',
              background: 'rgba(239,68,68,0.08)',
              border: '1px solid rgba(239,68,68,0.25)',
              color: '#ef4444',
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
    </motion.div>
  );
}

const iconMap: Record<string, React.ComponentType<{ size?: number }>> = {
  Star,
  Users,
  AlertTriangle,
  Clock,
  TrendingUp,
};

const Segments: React.FC = () => {
  const navigate = useNavigate();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  // Segment deletion has no dedicated resource; `crm:segments` is the operator-tier
  // resource that already covers this asset, consistent with customer deletion being
  // gated at operator-tier (`crm:customers:manage`) rather than walled off to admin.
  const canDeleteSegment = canCrmPermission(role, 'crm:segments');
  const { data: fetchedSegments, isLoading } = useListSegments();
  const removeSegment = useDeleteSegment();
  const createSegment = useCreateSegment();
  const [builderOpen, setBuilderOpen] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const segments: Segment[] = useMemo(() => {
    if (!fetchedSegments) return [];
    return fetchedSegments.map((s) => ({
      ...s,
      icon: iconMap[s.icon ?? ''] ?? Users,
    }));
  }, [fetchedSegments]);

  const handleDelete = (id: string): void => {
    removeSegment.mutate(id, {
      onSuccess: () => {
        toast.success('Segment deleted');
        setConfirmDeleteId(null);
      },
      onError: () => {
        toast.error('Failed to delete segment');
        setConfirmDeleteId(null);
      },
    });
  };

  const handleCreate = (name: string, rules: SegmentRule[]): void => {
    if (createSegment.isPending) return;
    createSegment.mutate(
      { name, rules: rules.map(serializeSegmentRule) },
      {
        onSuccess: () => {
          toast.success('Segment created');
          setBuilderOpen(false);
        },
        onError: () => {
          toast.error('Failed to create segment');
        },
      }
    );
  };

  const deleteTarget = segments.find((s) => s.id === confirmDeleteId);

  return (
    <>
      <HelmetContainer title="Segments · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              Segments
            </h2>
            <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
              {segments.length} active segments ·{' '}
              {segments.reduce((sum, s) => sum + s.customerCount, 0).toLocaleString()} customers
              total
            </p>
          </div>
          <button
            type="button"
            onClick={() => setBuilderOpen(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '8px 16px',
              borderRadius: '8px',
              fontSize: '13px',
              fontWeight: 600,
              background: ACCENT,
              color: '#fff',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            <Plus size={15} />
            New Segment
          </button>
        </div>

        {/* Segment grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px' }}>
          {segments.map((segment, i) => (
            <SegmentCard
              key={segment.id}
              segment={segment}
              index={i}
              onClick={() => navigate({ to: `/segments/${segment.id}` })}
              onDelete={() => setConfirmDeleteId(segment.id)}
              deletePending={removeSegment.isPending}
              canDelete={canDeleteSegment}
            />
          ))}
        </div>

        <ConfirmModal
          open={confirmDeleteId !== null}
          onClose={() => setConfirmDeleteId(null)}
          onConfirm={() => confirmDeleteId !== null && handleDelete(confirmDeleteId)}
          title="Delete segment?"
          description={
            deleteTarget
              ? `Are you sure you want to delete "${deleteTarget.name}"? This action cannot be undone.`
              : ''
          }
          confirmLabel="Delete"
          variant="danger"
        />

        <AnimatePresence>
          {builderOpen && (
            <RuleBuilder onClose={() => setBuilderOpen(false)} onSave={handleCreate} />
          )}
        </AnimatePresence>
      </div>
    </>
  );
};

export default Segments;
