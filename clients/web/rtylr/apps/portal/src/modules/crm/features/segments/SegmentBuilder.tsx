import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Plus, Save } from 'lucide-react';
import { useCreateSegment } from '../../shared/hooks/queries/useSegments';
import { Breadcrumbs, Button, Card, Input, Textarea, PageHeader } from '../../shared/components/ui';
import { SegmentRuleRow } from '../../shared/components/crm';
import {
  SEGMENT_RULE_FIELDS,
  SEGMENT_RULE_VALUES,
  getSegmentRuleOperators,
  serializeSegmentRule,
  type SegmentRuleField,
  type SegmentRuleOperator,
} from '../../shared/lib/segmentRules';

interface Rule {
  id: number;
  field: string;
  operator: string;
  value: string;
}

let nextRuleId = 1;

function isSegmentRuleField(value: string): value is SegmentRuleField {
  return (SEGMENT_RULE_FIELDS as readonly string[]).includes(value);
}

function isValidRule(rule: Rule): boolean {
  if (!isSegmentRuleField(rule.field)) return false;
  if (!getSegmentRuleOperators(rule.field).includes(rule.operator as SegmentRuleOperator)) {
    return false;
  }

  const value = rule.value.trim();
  if (!value) return false;
  if (rule.field === 'tier') {
    return (SEGMENT_RULE_VALUES.tier as readonly string[]).includes(value);
  }
  return /^-?(?:\d+(?:\.\d*)?|\.\d+)$/.test(value);
}

function normalizeRuleChange(current: Rule, updated: Omit<Rule, 'id'>): Omit<Rule, 'id'> {
  if (updated.field === current.field || !isSegmentRuleField(updated.field)) return updated;
  const valueIsStillValid = isValidRule({ ...current, ...updated });
  return valueIsStillValid ? updated : { ...updated, value: SEGMENT_RULE_VALUES[updated.field][0] };
}

const SegmentBuilder: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const create = useCreateSegment();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#3b82f6');
  const [rules, setRules] = useState<Rule[]>([
    { id: nextRuleId++, field: 'totalSpent', operator: '>', value: '500' },
  ]);
  const rulesAreValid = rules.length > 0 && rules.every(isValidRule);

  const updateRule = (i: number, rule: Rule) => {
    setRules((r) => r.map((x, idx) => (idx === i ? rule : x)));
  };

  const addRule = () =>
    setRules((r) => [...r, { id: nextRuleId++, field: 'visits', operator: '>', value: '5' }]);
  const removeRule = (i: number) => setRules((r) => r.filter((_, idx) => idx !== i));

  const handleSave = () => {
    if (!name.trim() || !rulesAreValid) {
      toast.error(t('crmErrors.generic'));
      return;
    }
    const ruleStrings = rules.map(serializeSegmentRule);
    create.mutate(
      {
        name,
        description,
        color,
        rules: ruleStrings,
        customerCount: 0,
        trend: '+0%',
        trendUp: true,
      },
      {
        onSuccess: () => {
          toast.success(t('crm.segments.created'));
          navigate({ to: '/crm/segments' });
        },
        onError: () => {
          toast.error(t('crm.segments.createFailed'));
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.segmentBuilder.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '840px' }}>
        <Breadcrumbs
          items={[
            { label: t('crmNav.segments'), to: '/crm/segments' },
            { label: t('crm.segments.breadcrumbNew') },
          ]}
        />
        <PageHeader
          title={t('crm.segmentBuilder.pageTitle')}
          subtitle={t('crm.segmentBuilder.subtitle')}
          actions={
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={!name.trim() || !rulesAreValid || create.isPending}
            >
              <Save size={13} />
              {create.isPending ? 'Saving…' : 'Save segment'}
            </Button>
          }
        />

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <Card>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: '2fr 1fr',
                gap: '12px',
                marginBottom: '12px',
              }}
            >
              <Input
                label={t('crmCommon.name')}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('crm.segmentBuilder.namePlaceholder')}
              />
              <div>
                <div
                  style={{
                    fontSize: '11px',
                    color: 'var(--hub-text-dim)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    fontWeight: 600,
                    marginBottom: '4px',
                  }}
                >
                  {t('crm.segmentBuilder.color')}
                </div>
                <div style={{ display: 'flex', gap: '6px' }}>
                  {['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#e879f9'].map((c) => (
                    <button
                      type="button"
                      key={c}
                      onClick={() => setColor(c)}
                      style={{
                        width: '28px',
                        height: '28px',
                        borderRadius: '8px',
                        border: color === c ? '2px solid var(--hub-text)' : '2px solid transparent',
                        background: c,
                        cursor: 'pointer',
                      }}
                    />
                  ))}
                </div>
              </div>
            </div>
            <Textarea
              label={t('crm.campaigns.descriptionLabel')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('crm.segmentBuilder.descriptionPlaceholder')}
            />
          </Card>

          <Card>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: '12px',
                marginBottom: '12px',
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
                {t('crm.segmentBuilder.rules')}
              </h4>
              <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>
                {t('crm.segmentBuilder.matchAll')}
              </span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {rules.map((r, i) => (
                <SegmentRuleRow
                  key={r.id}
                  rule={r}
                  onChange={(updated) =>
                    updateRule(i, { ...r, ...normalizeRuleChange(r, updated) })
                  }
                  onRemove={() => removeRule(i)}
                />
              ))}
            </div>
            <Button variant="outline" onClick={addRule} style={{ marginTop: '12px' }}>
              <Plus size={13} /> Add rule
            </Button>
          </Card>
        </div>
      </div>
    </>
  );
};

export default SegmentBuilder;
