import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Save, Plus, Zap, Mail, MessageSquare, Gift } from 'lucide-react';
import {
  useCreateAutomation,
  useGetAutomation,
  useUpdateAutomation,
  actionTypeFor,
  triggerTypeFor,
} from '../../shared/hooks/queries/useAutomations';
import { Breadcrumbs, Button, Card, Input, PageHeader } from '../../shared/components/ui';
import { AutomationNode } from '../../shared/components/crm';

interface Node {
  id: string;
  type: 'trigger' | 'action' | 'delay' | 'condition';
  label: string;
  automationType: string;
  description?: string;
}

const triggers = [
  { type: 'customer_signup', label: 'New customer', description: 'When a customer is created' },
  { type: 'tag_added', label: 'Tag added', description: 'When a tag is added to a customer' },
  {
    type: 'points_threshold',
    label: 'Points threshold',
    description: 'When customer points reach target',
  },
];
const actions = [
  { type: 'send_email', label: 'Send email', description: 'Sends an email campaign', icon: Mail },
  { type: 'send_sms', label: 'Send SMS', description: 'Sends an SMS message', icon: MessageSquare },
  { type: 'award_points', label: 'Award points', description: 'Adds loyalty points', icon: Gift },
  { type: 'tag_add', label: 'Add tag', description: 'Adds a tag to the customer', icon: Plus },
];

const AutomationBuilder: React.FC = () => {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { id?: number };
  const automationId = search?.id ? Number(search.id) : undefined;
  const isEditing = !!automationId;
  const create = useCreateAutomation();
  const update = useUpdateAutomation();
  const { data: existing } = useGetAutomation(automationId);
  const [name, setName] = useState('');
  const [nodes, setNodes] = useState<Node[]>([
    {
      id: 'n1',
      type: 'trigger',
      automationType: 'customer_signup',
      label: 'New customer',
      description: 'When a customer is created',
    },
  ]);

  React.useEffect(() => {
    if (!existing) return;
    setName(existing.name);
    if (existing.nodes?.length) {
      setNodes(
        existing.nodes.map((node) => ({
          id: node.id,
          type: node.type as Node['type'],
          label: node.label,
          automationType:
            node.automationType ??
            (node.type === 'trigger'
              ? triggerTypeFor(node.label)
              : String(actionTypeFor(node.label))),
        }))
      );
    }
  }, [existing]);

  const addTrigger = (t: (typeof triggers)[number]) => {
    setNodes([
      {
        id: 'n1',
        type: 'trigger',
        automationType: t.type,
        label: t.label,
        description: t.description,
      },
    ]);
  };

  const addAction = (a: (typeof actions)[number]) => {
    setNodes((n) => [
      ...n,
      {
        id: 'n' + (n.length + 1),
        type: 'action',
        automationType: a.type,
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
    const payload = {
      name,
      trigger: nodes[0]?.label,
      triggerType: nodes[0]?.automationType,
      actions: nodes.slice(1).map((node) => ({ type: node.automationType })),
      nodes: nodes.map((node) => ({ ...node, x: 0, y: 0 })),
      isActive: false,
      status: 'paused' as const,
    };
    const callbacks = {
      onSuccess: () => {
        toast.success('Automation saved');
        navigate({ to: '/automations' });
      },
      onError: () => {
        toast.error('Failed to save automation');
      },
    };
    if (isEditing) {
      update.mutate({ id: automationId!, data: payload }, callbacks);
    } else {
      create.mutate(payload, callbacks);
    }
  };

  return (
    <>
      <HelmetContainer title={`${isEditing ? 'Edit' : 'New'} automation · rtylr`} />
      <div style={{ maxWidth: '1200px' }}>
        <Breadcrumbs
          items={[
            { label: 'Automations', to: '/automations' },
            { label: isEditing ? 'Edit' : 'New' },
          ]}
        />
        <PageHeader
          title={isEditing ? 'Edit automation' : 'New automation'}
          subtitle="Drag triggers and actions to build a workflow"
          actions={
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={!name || create.isPending || update.isPending}
            >
              <Save size={13} />
              {create.isPending || update.isPending ? 'Saving…' : 'Save'}
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
