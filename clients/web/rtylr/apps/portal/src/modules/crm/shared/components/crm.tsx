import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import {
  Star,
  X,
  TrendingUp,
  TrendingDown,
  Plus,
  Minus,
  Mail,
  MessageSquare,
  ShoppingBag,
  Gift,
  StickyNote,
  Award,
  Clock,
  Check,
  Search as SearchIcon,
  User,
} from 'lucide-react';
import { Tier, tierColor, getTier, getCustomerTier, pointsToNextTier } from '../lib/tiers';
import { getSegmentRuleOperators, SEGMENT_RULE_FIELDS } from '../lib/segmentRules';
import { ACCENT, Input, Select, Button } from './ui';
import type { TimelineEntry, Customer } from '../services/crmApi';

// ── 16. CustomerAvatar ──────────────────────────────────────────
export const CustomerAvatar: React.FC<{
  initials: string;
  points?: number;
  size?: number;
  onClick?: () => void;
}> = ({ initials, points, size = 36, onClick }) => {
  const tier = points !== undefined ? getTier(points) : null;
  const ringColor = tier ? tierColor[tier] : 'transparent';
  return (
    <div
      onClick={onClick}
      style={{
        position: 'relative',
        width: `${size}px`,
        height: `${size}px`,
        borderRadius: '50%',
        background: 'var(--dash-hover)',
        color: ACCENT,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: `${Math.max(10, size * 0.32)}px`,
        fontWeight: 700,
        border: tier ? `2px solid ${ringColor}` : '2px solid transparent',
        flexShrink: 0,
        cursor: onClick ? 'pointer' : 'default',
      }}
    >
      {initials}
    </div>
  );
};

// ── 17. TimelineEntry ───────────────────────────────────────────
export const TimelineEntryRow: React.FC<{ entry: TimelineEntry }> = ({ entry }) => {
  const icons: Record<string, React.ComponentType<{ size?: number }>> = {
    visit: User,
    purchase: ShoppingBag,
    email: Mail,
    sms: MessageSquare,
    note: StickyNote,
    points: TrendingUp,
    reward: Gift,
    adjust: TrendingUp,
    expire: Clock,
    award: Award,
    upgrade: Award,
  };
  const colors: Record<string, string> = {
    visit: '#3b82f6',
    purchase: '#10b981',
    email: '#8b5cf6',
    sms: '#f59e0b',
    note: '#94a3b8',
    points: '#22c55e',
    reward: '#e879f9',
    adjust: '#22c55e',
    expire: '#f97316',
    award: '#e879f9',
    upgrade: '#e879f9',
  };
  const Icon = icons[entry.type] || Clock;
  const color = colors[entry.type] || ACCENT;
  return (
    <div style={{ display: 'flex', gap: '12px', alignItems: 'flex-start' }}>
      <div
        style={{
          width: '32px',
          height: '32px',
          borderRadius: '8px',
          background: `${color}14`,
          color,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <Icon size={14} />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>
          {entry.title}
        </div>
        {entry.description && (
          <div style={{ fontSize: '12px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>
            {entry.description}
          </div>
        )}
        <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginTop: '3px' }}>
          {entry.at}
        </div>
      </div>
    </div>
  );
};

// ── 18. PointsBadge ─────────────────────────────────────────────
export const PointsBadge: React.FC<{ points: number; trend?: 'up' | 'down' | null }> = ({
  points,
  trend = null,
}) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: '4px',
      padding: '4px 10px',
      borderRadius: '9999px',
      background: 'var(--dash-hover)',
      color: ACCENT,
      fontSize: '12px',
      fontWeight: 600,
    }}
  >
    {points.toLocaleString()} pts
    {trend === 'up' && <TrendingUp size={11} />}
    {trend === 'down' && <TrendingDown size={11} />}
  </span>
);

// ── 19. TierBadge ───────────────────────────────────────────────
export const TierBadge: React.FC<{ tier: Tier; size?: 'sm' | 'md' }> = ({ tier, size = 'sm' }) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: '4px',
      padding: size === 'sm' ? '2px 8px' : '4px 10px',
      borderRadius: '9999px',
      fontSize: size === 'sm' ? '10px' : '11px',
      fontWeight: 600,
      background: `${tierColor[tier]}18`,
      color: tierColor[tier],
      whiteSpace: 'nowrap',
    }}
  >
    <Star size={size === 'sm' ? 9 : 10} />
    {tier}
  </span>
);

// ── 20. TagChip ─────────────────────────────────────────────────
export const TagChip: React.FC<{
  label: string;
  onRemove?: () => void;
  color?: string;
}> = ({ label, onRemove, color = ACCENT }) => {
  const { t } = useTranslation();
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        padding: '2px 8px 2px 10px',
        borderRadius: '9999px',
        fontSize: '11px',
        fontWeight: 500,
        background: `${color}14`,
        color,
        border: `1px solid ${color}33`,
      }}
    >
      {label}
      {onRemove && (
        <button
          aria-label={t('crm.leadDetail.remove')}
          type="button"
          onClick={onRemove}
          style={{
            background: 'transparent',
            border: 'none',
            cursor: 'pointer',
            color: 'currentColor',
            display: 'inline-flex',
            padding: 0,
            opacity: 0.7,
          }}
        >
          <X size={10} />
        </button>
      )}
    </span>
  );
};

// ── 21. NotesPanel ──────────────────────────────────────────────
export const NotesPanel: React.FC<{
  notes: string;
  onSave: (notes: string) => void;
  readOnly?: boolean;
}> = ({ notes, onSave, readOnly = false }) => {
  const { t } = useTranslation();
  const [value, setValue] = useState(notes);
  const [editing, setEditing] = useState(false);
  return (
    <div
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '14px',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: '8px',
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
          }}
        >
          {t('crm.leadDetail.notesLabel')}
        </h4>
        {!editing && !readOnly && (
          <button
            type="button"
            onClick={() => setEditing(true)}
            style={{
              background: 'transparent',
              border: 'none',
              color: ACCENT,
              fontSize: '12px',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            {t('crmCommon.edit')}
          </button>
        )}
      </div>
      {editing ? (
        <>
          <textarea
            value={value}
            onChange={(e) => setValue(e.target.value)}
            rows={4}
            style={{
              width: '100%',
              padding: '8px 10px',
              borderRadius: '8px',
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
              fontSize: '13px',
              fontFamily: 'inherit',
              resize: 'vertical',
              outline: 'none',
            }}
          />
          <div
            style={{ display: 'flex', gap: '6px', marginTop: '8px', justifyContent: 'flex-end' }}
          >
            <Button
              variant="outline"
              style={{ padding: '6px 12px' }}
              onClick={() => {
                setValue(notes);
                setEditing(false);
              }}
            >
              {t('crmCommon.cancel')}
            </Button>
            <Button
              variant="primary"
              style={{ padding: '6px 12px' }}
              onClick={() => {
                onSave(value);
                setEditing(false);
              }}
            >
              {t('crmCommon.save')}
            </Button>
          </div>
        </>
      ) : (
        <p
          style={{
            fontSize: '13px',
            color: notes ? 'var(--hub-text)' : 'var(--hub-text-dim)',
            margin: 0,
            lineHeight: 1.6,
            whiteSpace: 'pre-wrap',
          }}
        >
          {notes || 'No notes yet.'}
        </p>
      )}
    </div>
  );
};

// ── 22. EmailEditor ─────────────────────────────────────────────
export const EmailEditor: React.FC<{
  subject: string;
  body: string;
  onChange: (s: string, b: string) => void;
}> = ({ subject, body, onChange }) => {
  const { t } = useTranslation();
  const insertVar = (v: string) => onChange(subject, body + ` {{${v}}}`);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
      <Input
        label={t('crm.composer.subject')}
        value={subject}
        onChange={(e) => onChange(e.target.value, body)}
        placeholder={t('crm.composer.subjectPlaceholder')}
      />
      <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
        <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginRight: '4px' }}>
          {t('crm.composer.insert')}
        </span>
        {['first_name', 'last_name', 'points', 'tier', 'business_name'].map((v) => (
          <button
            type="button"
            key={v}
            onClick={() => insertVar(v)}
            style={{
              padding: '2px 8px',
              borderRadius: '6px',
              fontSize: '11px',
              background: 'rgba(59,130,246,0.1)',
              color: ACCENT,
              border: '1px solid rgba(59,130,246,0.2)',
              cursor: 'pointer',
              fontFamily: 'ui-monospace, monospace',
            }}
          >
            {`{{${v}}}`}
          </button>
        ))}
      </div>
      <label style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        <span
          style={{
            fontSize: '11px',
            color: 'var(--hub-text-dim)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontWeight: 600,
          }}
        >
          {t('crm.composer.body')}
        </span>
        <textarea
          value={body}
          onChange={(e) => onChange(subject, e.target.value)}
          rows={10}
          placeholder={t('crm.composer.bodyPlaceholder')}
          style={{
            padding: '12px',
            borderRadius: '8px',
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
            fontSize: '13px',
            fontFamily: 'inherit',
            resize: 'vertical',
            outline: 'none',
            lineHeight: 1.6,
          }}
        />
      </label>
    </div>
  );
};

// ── 23. SMSComposer ─────────────────────────────────────────────
export const SMSComposer: React.FC<{
  body: string;
  onChange: (b: string) => void;
}> = ({ body, onChange }) => {
  const { t } = useTranslation();
  const remaining = 160 - body.length;
  const tooLong = remaining < 0;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
      <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
        <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginRight: '4px' }}>
          {t('crm.composer.insert')}
        </span>
        {['first_name', 'points', 'business_name'].map((v) => (
          <button
            type="button"
            key={v}
            onClick={() => onChange(body + ` {{${v}}}`)}
            style={{
              padding: '2px 8px',
              borderRadius: '6px',
              fontSize: '11px',
              background: 'rgba(59,130,246,0.1)',
              color: ACCENT,
              border: '1px solid rgba(59,130,246,0.2)',
              cursor: 'pointer',
              fontFamily: 'ui-monospace, monospace',
            }}
          >
            {`{{${v}}}`}
          </button>
        ))}
      </div>
      <textarea
        value={body}
        onChange={(e) => onChange(e.target.value)}
        rows={5}
        maxLength={320}
        placeholder={t('crm.composer.smsPlaceholder')}
        style={{
          padding: '12px',
          borderRadius: '8px',
          background: 'var(--hub-input-bg)',
          border: `1px solid ${tooLong ? '#ef4444' : 'var(--hub-border)'}`,
          color: 'var(--hub-text)',
          fontSize: '13px',
          fontFamily: 'inherit',
          resize: 'vertical',
          outline: 'none',
          lineHeight: 1.6,
        }}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px' }}>
        <span style={{ color: 'var(--hub-text-dim)' }}>{t('crm.composer.smsHint')}</span>
        <span
          style={{
            color: tooLong ? '#ef4444' : remaining < 20 ? '#f59e0b' : 'var(--hub-text-muted)',
            fontWeight: 600,
          }}
        >
          {body.length}/160
        </span>
      </div>
    </div>
  );
};

// ── 24. SegmentRule ─────────────────────────────────────────────
// Keys, not labels: module scope runs before i18next has a language.
const OPERATOR_LABEL_KEYS: Record<string, string> = {
  '>': 'crm.segmentOperators.gt',
  '<': 'crm.segmentOperators.lt',
  '=': 'crm.segmentOperators.eq',
  '!=': 'crm.segmentOperators.ne',
};
export const SegmentRuleRow: React.FC<{
  rule: { field: string; operator: string; value: string };
  onChange: (rule: { field: string; operator: string; value: string }) => void;
  onRemove: () => void;
}> = ({ rule, onChange, onRemove }) => {
  const { t } = useTranslation();
  return (
    <div
      style={{
        display: 'flex',
        gap: '8px',
        alignItems: 'center',
        padding: '10px',
        borderRadius: '10px',
        background: 'var(--hub-hover)',
        border: '1px solid var(--hub-border)',
      }}
    >
      <Select
        value={rule.field}
        onChange={(e) => {
          const field = e.target.value;
          const operators = getSegmentRuleOperators(field);
          const operator = operators.includes(rule.operator as (typeof operators)[number])
            ? rule.operator
            : operators[0];
          onChange({ ...rule, field, operator });
        }}
        style={{ flex: 1 }}
      >
        {SEGMENT_RULE_FIELDS.map((field) => (
          <option key={field} value={field}>
            {t(`crm.segmentFields.${field}`)}
          </option>
        ))}
      </Select>
      <Select
        value={rule.operator}
        onChange={(e) => onChange({ ...rule, operator: e.target.value })}
        style={{ width: '130px' }}
      >
        {getSegmentRuleOperators(rule.field).map((operator) => (
          <option key={operator} value={operator}>
            {OPERATOR_LABEL_KEYS[operator] ? t(OPERATOR_LABEL_KEYS[operator]) : operator}
          </option>
        ))}
      </Select>
      <Input
        value={rule.value}
        onChange={(e) => onChange({ ...rule, value: e.target.value })}
        placeholder={t('crm.components.valuePlaceholder')}
        style={{ width: '140px' }}
      />
      <button
        aria-label={t('crm.components.removeTag')}
        type="button"
        onClick={onRemove}
        style={{
          background: 'transparent',
          border: '1px solid var(--hub-border)',
          borderRadius: '8px',
          width: '32px',
          height: '32px',
          cursor: 'pointer',
          color: 'var(--hub-text-muted)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <X size={14} />
      </button>
    </div>
  );
};

// ── 25. AutomationNode ──────────────────────────────────────────
export const AutomationNode: React.FC<{
  type: 'trigger' | 'action' | 'delay' | 'condition';
  label: string;
  description?: string;
  onRemove?: () => void;
}> = ({ type, label, description, onRemove }) => {
  const { t } = useTranslation();
  const colors: Record<string, string> = {
    trigger: '#10b981',
    action: '#3b82f6',
    delay: '#f59e0b',
    condition: '#8b5cf6',
  };
  const color = colors[type];
  return (
    <div
      style={{
        padding: '14px',
        borderRadius: '12px',
        background: 'var(--hub-card)',
        border: `1px solid ${color}40`,
        minWidth: '220px',
        position: 'relative',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
        <span
          style={{
            fontSize: '9px',
            fontWeight: 700,
            textTransform: 'uppercase',
            letterSpacing: '0.08em',
            color,
            background: `${color}14`,
            padding: '2px 6px',
            borderRadius: '4px',
          }}
        >
          {type}
        </span>
        {onRemove && (
          <button
            aria-label={t('crm.components.removeNode')}
            type="button"
            onClick={onRemove}
            style={{
              marginLeft: 'auto',
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-dim)',
              padding: 0,
              display: 'flex',
            }}
          >
            <X size={12} />
          </button>
        )}
      </div>
      <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>{label}</div>
      {description && (
        <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>
          {description}
        </div>
      )}
    </div>
  );
};

// ── 26. CampaignStatsCard ───────────────────────────────────────
export const CampaignStatsCard: React.FC<{
  label: string;
  value: number | string;
  subValue?: string;
  sparkline?: number[];
  color?: string;
}> = ({ label, value, subValue, sparkline, color = ACCENT }) => {
  const max = sparkline ? Math.max(...sparkline, 1) : 1;
  return (
    <div
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '14px',
      }}
    >
      <div
        style={{
          fontSize: '11px',
          color: 'var(--hub-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontWeight: 600,
          marginBottom: '6px',
        }}
      >
        {label}
      </div>
      <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--hub-text)' }}>{value}</div>
      {subValue && (
        <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>
          {subValue}
        </div>
      )}
      {sparkline && (
        <svg width="100%" height="28" style={{ marginTop: '8px', display: 'block' }}>
          {sparkline.map((v, i) => {
            const barWidth = 100 / sparkline.length;
            const h = (v / max) * 22 + 2;
            return (
              <rect
                key={i}
                x={`${i * barWidth + 0.5}%`}
                y={28 - h}
                width={`${barWidth - 1}%`}
                height={h}
                fill={color}
                opacity={0.7}
                rx={1}
              />
            );
          })}
        </svg>
      )}
    </div>
  );
};

// ── 27. LoyaltyPointsInput ──────────────────────────────────────
export const LoyaltyPointsInput: React.FC<{
  onAward: (points: number, reason: string) => void;
  onDeduct: (points: number, reason: string) => void;
}> = ({ onAward, onDeduct }) => {
  const { t } = useTranslation();
  const [points, setPoints] = useState(0);
  const [reason, setReason] = useState('');
  return (
    <div
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '14px',
        display: 'flex',
        flexDirection: 'column',
        gap: '10px',
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
        }}
      >
        {t('crm.components.adjustPoints')}
      </h4>
      <Input
        type="number"
        placeholder={t('crm.components.amount')}
        value={points || ''}
        onChange={(e) => setPoints(Number(e.target.value) || 0)}
      />
      <Input
        placeholder={t('crm.components.reasonOptional')}
        value={reason}
        onChange={(e) => setReason(e.target.value)}
      />
      <div style={{ display: 'flex', gap: '6px' }}>
        <Button
          variant="primary"
          disabled={points <= 0}
          onClick={() => {
            if (points > 0) {
              onAward(points, reason);
              setPoints(0);
              setReason('');
            }
          }}
          style={{ flex: 1, justifyContent: 'center' }}
        >
          <Plus size={13} /> Award
        </Button>
        <Button
          variant="outline"
          disabled={points <= 0}
          onClick={() => {
            if (points > 0) {
              onDeduct(points, reason);
              setPoints(0);
              setReason('');
            }
          }}
          style={{ flex: 1, justifyContent: 'center' }}
        >
          <Minus size={13} /> Deduct
        </Button>
      </div>
    </div>
  );
};

// ── 28. ConsentBanner ───────────────────────────────────────────
export const ConsentBanner: React.FC<{
  consent: Record<string, boolean>;
}> = ({ consent }) => {
  const { t } = useTranslation();
  const items = [
    { id: 'email', label: t('crmCommon.email'), allowed: consent.email === true },
    { id: 'sms', label: t('crm.components.sms'), allowed: consent.sms === true },
    {
      id: 'marketing',
      label: t('crm.createCustomer.marketing'),
      allowed: consent.marketing === true,
    },
  ];
  return (
    <div
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '12px 14px',
        display: 'flex',
        gap: '16px',
        flexWrap: 'wrap',
      }}
    >
      {items.map((it) => (
        <div key={it.id} style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <div
            style={{
              width: '16px',
              height: '16px',
              borderRadius: '50%',
              background: it.allowed ? '#10b98122' : '#ef444422',
              color: it.allowed ? '#10b981' : '#ef4444',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {it.allowed ? <Check size={10} /> : <X size={10} />}
          </div>
          <span style={{ fontSize: '12px', color: 'var(--hub-text)', fontWeight: 500 }}>
            {it.label}
          </span>
          <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
            {it.allowed ? 'opted in' : 'opted out'}
          </span>
        </div>
      ))}
    </div>
  );
};

// ── 29. CustomerQuickView ───────────────────────────────────────
export const CustomerQuickView: React.FC<{ customer: Customer }> = ({ customer }) => {
  const { t } = useTranslation();
  const tier = getCustomerTier(customer.loyaltyPoints, customer.loyaltyTier);
  const next = pointsToNextTier(customer.loyaltyPoints, undefined, customer.loyaltyTier);
  return (
    <div
      style={{
        width: '280px',
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '12px',
        padding: '14px',
        boxShadow: 'var(--hub-shadow)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '12px' }}>
        <CustomerAvatar initials={customer.initials} points={customer.loyaltyPoints} size={44} />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)' }}>
            {customer.name}
          </div>
          <TierBadge tier={tier} />
        </div>
      </div>
      <div
        style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '12px' }}
      >
        <div>
          <div
            style={{ color: 'var(--hub-text-dim)', fontSize: '10px', textTransform: 'uppercase' }}
          >
            {t('crm.components.spent')}
          </div>
          <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>
            ${customer.totalSpent.toLocaleString()}
          </div>
        </div>
        <div>
          <div
            style={{ color: 'var(--hub-text-dim)', fontSize: '10px', textTransform: 'uppercase' }}
          >
            {t('crm.customers.stats.visits')}
          </div>
          <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{customer.visits}</div>
        </div>
        <div>
          <div
            style={{ color: 'var(--hub-text-dim)', fontSize: '10px', textTransform: 'uppercase' }}
          >
            {t('crm.components.points')}
          </div>
          <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>
            {customer.loyaltyPoints.toLocaleString()}
          </div>
        </div>
        <div>
          <div
            style={{ color: 'var(--hub-text-dim)', fontSize: '10px', textTransform: 'uppercase' }}
          >
            {t('crm.components.lastSeen')}
          </div>
          <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{customer.lastVisit}</div>
        </div>
      </div>
      {next !== null && (
        <div style={{ marginTop: '10px', fontSize: '11px', color: 'var(--hub-text-muted)' }}>
          {next} points to next tier
        </div>
      )}
    </div>
  );
};

// ── 30. SearchableCustomerSelect ────────────────────────────────
export const SearchableCustomerSelect: React.FC<{
  customers: Customer[];
  value: number | null;
  onChange: (id: number | null) => void;
  placeholder?: string;
}> = ({ customers, value, onChange, placeholder }) => {
  const { t } = useTranslation();
  const placeholderText = placeholder ?? t('crm.components.selectCustomer');
  const [search, setSearch] = useState('');
  const [open, setOpen] = useState(false);
  const selected = customers.find((c) => c.id === value);
  const filtered = customers.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div style={{ position: 'relative' }}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        style={{
          width: '100%',
          height: '36px',
          padding: '0 12px',
          borderRadius: '8px',
          background: 'var(--hub-input-bg)',
          border: '1px solid var(--hub-border)',
          color: selected ? 'var(--hub-text)' : 'var(--hub-text-muted)',
          fontSize: '13px',
          textAlign: 'left',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}
      >
        <SearchIcon size={14} style={{ color: 'var(--hub-text-dim)' }} />
        {selected ? selected.name : placeholderText}
      </button>
      {open && (
        <motion.div
          initial={{ opacity: 0, y: -4 }}
          animate={{ opacity: 1, y: 0 }}
          style={{
            position: 'absolute',
            top: '40px',
            left: 0,
            right: 0,
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '10px',
            boxShadow: 'var(--hub-shadow)',
            zIndex: 10,
            maxHeight: '300px',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <input
            autoFocus
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('crm.components.searchPlaceholder')}
            style={{
              padding: '10px 12px',
              background: 'transparent',
              border: 'none',
              borderBottom: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
              fontSize: '13px',
              outline: 'none',
            }}
          />
          <div style={{ overflowY: 'auto', flex: 1, maxHeight: '240px' }}>
            {filtered.length === 0 && (
              <div
                style={{
                  padding: '20px',
                  textAlign: 'center',
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                }}
              >
                {t('crm.components.noResults')}
              </div>
            )}
            {filtered.map((c) => (
              <button
                type="button"
                key={c.id}
                onClick={() => {
                  onChange(c.id);
                  setOpen(false);
                  setSearch('');
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  width: '100%',
                  padding: '8px 12px',
                  background: 'transparent',
                  border: 'none',
                  cursor: 'pointer',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  textAlign: 'left',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'transparent';
                }}
              >
                <CustomerAvatar initials={c.initials} points={c.loyaltyPoints} size={26} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 500 }}>{c.name}</div>
                  <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>{c.email}</div>
                </div>
              </button>
            ))}
          </div>
        </motion.div>
      )}
    </div>
  );
};
