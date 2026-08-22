import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Plus, Save } from 'lucide-react';
import { useCreateSegment } from '../../shared/hooks/queries/useSegments';
import { Breadcrumbs, Button, Card, Input, Textarea, PageHeader } from '../../shared/components/ui';
import { SegmentRuleRow } from '../../shared/components/crm';

interface Rule {
  field: string;
  operator: string;
  value: string;
}

const SegmentBuilder: React.FC = () => {
  const navigate = useNavigate();
  const create = useCreateSegment();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#3b82f6');
  const [rules, setRules] = useState<Rule[]>([
    { field: 'totalSpent', operator: '>', value: '500' },
  ]);
  const [matchAll, setMatchAll] = useState(true);

  const updateRule = (i: number, rule: Rule) => {
    setRules((r) => r.map((x, idx) => (idx === i ? rule : x)));
  };

  const addRule = () => setRules((r) => [...r, { field: 'visits', operator: '>', value: '5' }]);
  const removeRule = (i: number) => setRules((r) => r.filter((_, idx) => idx !== i));

  const handleSave = () => {
    const ruleStrings = rules.map((r) => `${r.field} ${r.operator} ${r.value}`);
    create.mutate(
      {
        name: name.trim(),
        description: description.trim() || name.trim(),
        color,
        rules: ruleStrings,
        matchAll,
        customerCount: 0,
        trend: '+0%',
        trendUp: true,
      },
      {
        onSuccess: () => {
          navigate({ to: '/segments' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="New segment · rtylr" />
      <div style={{ maxWidth: '840px' }}>
        <Breadcrumbs items={[{ label: 'Segments', to: '/segments' }, { label: 'New' }]} />
        <PageHeader
          title="New segment"
          subtitle="Build dynamic rules to target customers"
          actions={
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={!name.trim() || create.isPending}
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
                label="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="High Value Customers"
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
                  Color
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
              label="Description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What defines this segment?"
            />
          </Card>

          <Card>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
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
                Rules
              </h4>
              <div
                style={{
                  display: 'flex',
                  borderRadius: '8px',
                  background: 'var(--hub-hover)',
                  padding: '3px',
                  border: '1px solid var(--hub-border)',
                }}
              >
                {[
                  { key: true, label: 'Match all' },
                  { key: false, label: 'Match any' },
                ].map((o) => (
                  <button
                    type="button"
                    key={String(o.key)}
                    onClick={() => setMatchAll(o.key)}
                    style={{
                      padding: '4px 12px',
                      borderRadius: '6px',
                      fontSize: '11px',
                      fontWeight: 600,
                      cursor: 'pointer',
                      border: 'none',
                      background: matchAll === o.key ? '#3b82f6' : 'transparent',
                      color: matchAll === o.key ? '#fff' : 'var(--hub-text-muted)',
                    }}
                  >
                    {o.label}
                  </button>
                ))}
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {rules.map((r, i) => (
                <SegmentRuleRow
                  key={i}
                  rule={r}
                  onChange={(updated) => updateRule(i, updated)}
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
