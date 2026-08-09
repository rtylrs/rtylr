import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Save, Plus, Zap, Mail, Clock, GitBranch, MessageSquare, Gift } from 'lucide-react';
import { useCreateAutomation } from '../../shared/hooks/queries/useAutomations';
import { Breadcrumbs, Button, Card, Input, PageHeader } from '../../shared/components/ui';
import { AutomationNode } from '../../shared/components/crm';

interface Node {
  id: string;
  type: 'trigger' | 'action' | 'delay' | 'condition';
  label: string;
  description?: string;
}

const triggers = [
  { label: 'New customer', description: 'When a customer is created' },
  { label: 'Birthday', description: "On the customer's birthday" },
  { label: 'Inactive 60 days', description: 'When inactive for 60+ days' },
  { label: 'Post visit', description: 'After a customer visit' },
  { label: 'Tier upgrade', description: 'When tier changes' },
];
const actions = [
  { label: 'Send email', description: 'Sends an email template', icon: Mail },
  { label: 'Send SMS', description: 'Sends an SMS message', icon: MessageSquare },
  { label: 'Award points', description: 'Adds loyalty points', icon: Gift },
  { label: 'Wait 3 days', description: 'Delays next step', icon: Clock, type: 'delay' as const },
  {
    label: 'If opted in',
    description: 'Branch on condition',
    icon: GitBranch,
    type: 'condition' as const,
  },
];

const AutomationBuilder: React.FC = () => {
  const navigate = useNavigate();
  const create = useCreateAutomation();
  const [name, setName] = useState('');
  const [nodes, setNodes] = useState<Node[]>([
    { id: 'n1', type: 'trigger', label: 'New customer', description: 'When a customer is created' },
  ]);

  const addTrigger = (t: (typeof triggers)[number]) => {
    setNodes([{ id: 'n1', type: 'trigger', label: t.label, description: t.description }]);
  };

  const addAction = (a: (typeof actions)[number]) => {
    setNodes((n) => [
      ...n,
      {
        id: 'n' + (n.length + 1),
        type: (a as any).type || 'action',
        label: a.label,
        description: a.description,
      },
    ]);
  };

  const removeNode = (id: string) => {
    if (id === 'n1') return;
    setNodes((n) => n.filter((x) => x.id !== id));
  };

  const handleSave = () => {
    create.mutate(
      {
        name,
        description: `${nodes[0]?.label} -> ${nodes.length - 1} step(s)`,
        trigger: nodes[0]?.label,
        status: 'draft',
        enrolled: 0,
        completed: 0,
      },
      {
        onSuccess: () => {
          toast.success('Automation saved');
          navigate({ to: '/automations' });
        },
        onError: () => {
          toast.error('Failed to save automation');
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="New automation · rtylr" />
      <div style={{ maxWidth: '1200px' }}>
        <Breadcrumbs items={[{ label: 'Automations', to: '/automations' }, { label: 'New' }]} />
        <PageHeader
          title="New automation"
          subtitle="Drag triggers and actions to build a workflow"
          actions={
            <Button variant="primary" onClick={handleSave} disabled={!name || create.isPending}>
              <Save size={13} />
              {create.isPending ? 'Saving…' : 'Save'}
            </Button>
          }
        />

        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Welcome series"
          style={{ maxWidth: '320px', marginBottom: '16px' }}
        />

        <div style={{ display: 'grid', gridTemplateColumns: '260px 1fr', gap: '16px' }}>
          {/* Palette */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <Card>
              <h4
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                  margin: '0 0 10px',
                }}
              >
                Triggers
              </h4>
              {triggers.map((t) => (
                <button
                  key={t.label}
                  onClick={() => addTrigger(t)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    width: '100%',
                    padding: '8px 10px',
                    borderRadius: '8px',
                    background: 'var(--hub-hover)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                    fontSize: '12px',
                    cursor: 'pointer',
                    marginBottom: '6px',
                    textAlign: 'left',
                  }}
                >
                  <Zap size={12} style={{ color: '#10b981', flexShrink: 0 }} />
                  {t.label}
                </button>
              ))}
            </Card>
            <Card>
              <h4
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 700,
                  margin: '0 0 10px',
                }}
              >
                Actions
              </h4>
              {actions.map((a) => {
                const Icon = a.icon;
                return (
                  <button
                    key={a.label}
                    onClick={() => addAction(a)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '8px',
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: '8px',
                      background: 'var(--hub-hover)',
                      border: '1px solid var(--hub-border)',
                      color: 'var(--hub-text)',
                      fontSize: '12px',
                      cursor: 'pointer',
                      marginBottom: '6px',
                      textAlign: 'left',
                    }}
                  >
                    <Icon size={12} style={{ color: '#3b82f6', flexShrink: 0 }} />
                    {a.label}
                  </button>
                );
              })}
            </Card>
          </div>

          {/* Canvas */}
          <Card style={{ minHeight: '500px' }}>
            <h4
              style={{
                fontSize: '11px',
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                fontWeight: 700,
                margin: '0 0 14px',
              }}
            >
              Workflow
            </h4>
            <div
              style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px' }}
            >
              {nodes.map((node, i) => (
                <React.Fragment key={node.id}>
                  <AutomationNode
                    type={node.type}
                    label={node.label}
                    description={node.description}
                    onRemove={i > 0 ? () => removeNode(node.id) : undefined}
                  />
                  {i < nodes.length - 1 && (
                    <div
                      style={{ width: '2px', height: '18px', background: 'var(--hub-border)' }}
                    />
                  )}
                </React.Fragment>
              ))}
              <div style={{ width: '2px', height: '18px', background: 'var(--hub-border)' }} />
              <div
                style={{
                  padding: '10px 18px',
                  border: '1px dashed var(--hub-border)',
                  borderRadius: '12px',
                  color: 'var(--hub-text-muted)',
                  fontSize: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                }}
              >
                <Plus size={12} /> Click an action to add
              </div>
            </div>
          </Card>
        </div>
      </div>
    </>
  );
};

export default AutomationBuilder;
