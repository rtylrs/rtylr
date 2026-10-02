import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
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
import {
  Breadcrumbs,
  Button,
  Card,
  Input,
  PageHeader,
  SkeletonList,
} from '../../shared/components/ui';
import { AutomationNode } from '../../shared/components/crm';

interface Node {
  id: string;
  type: 'trigger' | 'action' | 'delay' | 'condition';
  label: string;
  automationType: string;
  description?: string;
}

/**
 * `label` here is the canonical English and is deliberately not translated: it
 * is persisted as the automation's `trigger`, and `triggerTypeFor` matches it
 * against rows already in the database that predate `automationType`. Translating
 * it would write Arabic into that column and stop old automations resolving
 * their type.
 *
 * What the builder shows comes from the type instead, via
 * crm.automationTriggers.<type> and its description twin.
 */
const triggers = [
  { type: 'customer_signup', label: 'New customer' },
  { type: 'tag_added', label: 'Tag added' },
  { type: 'points_threshold', label: 'Points threshold' },
] as const;
const actions = [
  { type: 'send_email', label: 'Send email', icon: Mail },
  { type: 'send_sms', label: 'Send SMS', icon: MessageSquare },
  { type: 'award_points', label: 'Award points', icon: Gift },
  { type: 'tag_add', label: 'Add tag', icon: Plus },
] as const;

const AutomationBuilder: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { id?: number };
  const automationId = search?.id ? Number(search.id) : undefined;
  const isEditing = !!automationId;
  const create = useCreateAutomation();
  const update = useUpdateAutomation();
  const { data: existing, isLoading } = useGetAutomation(automationId);
  // An edit must not render (or save) the create-form defaults while the
  // detail query is still loading or has failed: the payload would replace the
  // stored trigger and actions with those defaults (B25-R06).
  const isEditPending = isEditing && !existing;
  const [name, setName] = useState('');
  const [nodes, setNodes] = useState<Node[]>([
    {
      id: 'n1',
      type: 'trigger',
      automationType: 'customer_signup',
      label: 'New customer',
    },
  ]);
  // Monotonic id source so removed nodes never free up an id that a later
  // add would reuse (duplicate keys / removeNode deleting both).
  const nextNodeId = React.useRef(2);

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
    // Seed the id counter past any loaded node ids.
    const maxLoaded =
      existing.nodes?.reduce((max, n) => Math.max(max, Number(n.id.replace('n', '')) || 0), 0) ?? 0;
    nextNodeId.current = Math.max(nextNodeId.current, maxLoaded + 1);
  }, [existing]);

  const addTrigger = (trigger: (typeof triggers)[number]) => {
    setNodes([
      {
        id: 'n1',
        type: 'trigger',
        automationType: trigger.type,
        label: trigger.label,
      },
    ]);
  };

  const addAction = (action: (typeof actions)[number]) => {
    setNodes((n) => [
      ...n,
      {
        id: 'n' + nextNodeId.current++,
        type: 'action',
        automationType: action.type,
        label: action.label,
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
      // Preserve the automation's current run state: saving an edit must not
      // silently pause an active workflow.
      isActive: existing?.isActive ?? false,
      status: existing?.status ?? ('paused' as const),
    };
    const callbacks = {
      onSuccess: () => {
        toast.success(t('crm.automations.saved'));
        navigate({ to: '/crm/automations' });
      },
      onError: () => {
        toast.error(t('crm.automations.saveFailed'));
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
            { label: t('crmNav.automations'), to: '/crm/automations' },
            {
              label: isEditing
                ? t('crm.automationBuilder.breadcrumbEdit')
                : t('crm.automationBuilder.breadcrumbNew'),
            },
          ]}
        />
        <PageHeader
          title={isEditing ? 'Edit automation' : 'New automation'}
          subtitle={t('crm.automationBuilder.subtitle')}
          actions={
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={!name || isEditPending || create.isPending || update.isPending}
            >
              <Save size={13} />
              {create.isPending || update.isPending ? 'Saving…' : 'Save'}
            </Button>
          }
        />

        {!isEditPending && (
          <Input
            label={t('crmCommon.name')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('crm.automationBuilder.namePlaceholder')}
            style={{ maxWidth: '320px', marginBottom: '16px' }}
          />
        )}

        {isEditPending ? (
          <Card style={{ minHeight: '500px' }}>
            {isLoading ? (
              <SkeletonList count={4} />
            ) : (
              <p style={{ color: 'var(--hub-text-muted)', fontSize: '13px', margin: 0 }}>
                {t('crm.automationBuilder.loadFailed')}
              </p>
            )}
          </Card>
        ) : (
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
                  {t('crm.automationBuilder.triggers')}
                </h4>
                {triggers.map((trigger) => (
                  <button
                    type="button"
                    key={trigger.type}
                    onClick={() => addTrigger(trigger)}
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
                    {t(`crm.automationTriggers.${trigger.type}`)}
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
                  {t('crmCommon.actions')}
                </h4>
                {actions.map((action) => {
                  const Icon = action.icon;
                  return (
                    <button
                      type="button"
                      key={action.type}
                      onClick={() => addAction(action)}
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
                      <Icon size={12} style={{ color: 'var(--crm-accent)', flexShrink: 0 }} />
                      {t(`crm.automationActions.${action.type}`)}
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
                {t('crm.automationBuilder.workflow')}
              </h4>
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: '8px',
                }}
              >
                {nodes.map((node, i) => (
                  <React.Fragment key={node.id}>
                    <AutomationNode
                      type={node.type}
                      label={
                        node.type === 'trigger'
                          ? t(`crm.automationTriggers.${node.automationType}`, node.label)
                          : t(`crm.automationActions.${node.automationType}`, node.label)
                      }
                      description={
                        node.type === 'trigger'
                          ? t(`crm.automationTriggerDescriptions.${node.automationType}`, '')
                          : t(`crm.automationActionDescriptions.${node.automationType}`, '')
                      }
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
        )}
      </div>
    </>
  );
};

export default AutomationBuilder;
