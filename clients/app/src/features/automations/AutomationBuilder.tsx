import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { Plus, Mail, Trash2 } from 'lucide-react';
import {
  useCreateAutomation,
  useGetAutomation,
  useUpdateAutomation,
  actionTypeFor,
  triggerTypeFor,
} from '../../modules/crm/shared/hooks/queries/useAutomations';
import {
  Breadcrumbs,
  Button,
  Input,
  Textarea,
  PageHeader,
  SkeletonList,
} from '../../modules/crm/shared/components/ui';
import './automations.css';

interface WorkflowAction {
  id: number;
  type: string;
  config: Record<string, unknown>;
}
function actionConfig(value: unknown): Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value) ? { ...value } : {};
}

export default function AutomationBuilder() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id: automationId } = useSearch({ strict: false });
  const isEditing = automationId !== undefined;
  const create = useCreateAutomation();
  const update = useUpdateAutomation();
  const { data: existing, isLoading, refetch } = useGetAutomation(automationId);
  const isEditPending = isEditing && !existing;
  const [name, setName] = useState('');
  const [trigger, setTrigger] = useState('order_created');
  const [actions, setActions] = useState<WorkflowAction[]>([]);
  const nextId = useRef(1);
  const loadedId = useRef<number | undefined>(undefined);
  useEffect(() => {
    // Background refetches must not overwrite an in-progress edit.
    if (!existing || loadedId.current === existing.id) return;
    loadedId.current = existing.id;
    setName(existing.name);
    setTrigger(existing.triggerType ?? triggerTypeFor(existing.trigger));
    const storedActions =
      existing.actions ??
      existing.nodes
        ?.slice(1)
        .map((node) => ({ type: node.automationType ?? actionTypeFor(node.label) })) ??
      [];
    setActions(
      storedActions.map((value) => {
        const config = actionConfig(value);
        return { id: nextId.current++, type: String(actionTypeFor(config.type ?? value)), config };
      })
    );
  }, [existing]);
  const pending = create.isPending || update.isPending;
  const valid =
    name.trim().length > 0 &&
    actions.length > 0 &&
    actions.every(
      (action) =>
        action.type !== 'send_email' ||
        (typeof action.config.subject === 'string' &&
          action.config.subject.trim() &&
          typeof action.config.body === 'string' &&
          action.config.body.trim())
    );
  const setField = (id: number, key: string, value: string) =>
    setActions((current) =>
      current.map((action) =>
        action.id === id ? { ...action, config: { ...action.config, [key]: value } } : action
      )
    );
  const title = t(isEditing ? 'automations.edit' : 'crmNav.newAutomation');
  return (
    <>
      <HelmetContainer title={`${title} · rtylr`} />
      <div className="portal-automations automation-builder">
        <Breadcrumbs
          items={[{ label: t('common.nav.automations'), to: '/automations' }, { label: title }]}
        />
        <PageHeader title={title} subtitle={t('automations.setupDescription')} />
        {isEditPending ? (
          isLoading ? (
            <SkeletonList count={4} />
          ) : (
            <div role="alert">
              <p>{t('crm.automationBuilder.loadFailed')}</p>
              <Button variant="outline" onClick={() => void refetch()}>
                {t('crm.actions.refresh')}
              </Button>
            </div>
          )
        ) : (
          <form
            onSubmit={(event) => {
              event.preventDefault();
              if (!valid || pending) return;
              const payload = {
                name: name.trim(),
                triggerType: trigger,
                triggerConfig: existing?.triggerConfig ?? {},
                actions: actions.map((action) => ({ ...action.config, type: action.type })),
                isActive: existing?.isActive ?? existing?.status === 'active',
                status: existing?.status ?? ('paused' as const),
              };
              const callbacks = { onSuccess: () => navigate({ to: '/automations' }) };
              if (automationId !== undefined)
                update.mutate({ id: automationId, data: payload }, callbacks);
              else create.mutate(payload, callbacks);
            }}
          >
            <Input
              label={t('crmCommon.name')}
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={t('crm.automationBuilder.namePlaceholder')}
              required
            />
            <label className="automation-trigger">
              <span>{t('automations.when')}</span>
              <select value={trigger} onChange={(event) => setTrigger(event.target.value)}>
                <option value="order_created">{t('crm.automationTriggers.order_created')}</option>
                {trigger !== 'order_created' && (
                  <option value={trigger}>{t(`crm.automationTriggers.${trigger}`, trigger)}</option>
                )}
              </select>
            </label>
            {trigger !== 'order_created' && (
              <p className="automation-note">{t('automations.legacyTrigger')}</p>
            )}
            {actions.map((action, index) => (
              <section className="automation-step" key={action.id}>
                <div className="automation-step-heading">
                  <h3>
                    {index + 1}. {t(`crm.automationActions.${action.type}`, action.type)}
                  </h3>
                  <Button
                    variant="ghost"
                    aria-label={t('automations.removeAction', { number: index + 1 })}
                    onClick={() =>
                      setActions((current) => current.filter((entry) => entry.id !== action.id))
                    }
                  >
                    <Trash2 size={16} />
                  </Button>
                </div>
                {action.type === 'send_email' ? (
                  <>
                    <Input
                      label={t('automations.subject')}
                      value={typeof action.config.subject === 'string' ? action.config.subject : ''}
                      onChange={(event) => setField(action.id, 'subject', event.target.value)}
                      required
                    />
                    <Textarea
                      label={t('automations.body')}
                      value={typeof action.config.body === 'string' ? action.config.body : ''}
                      onChange={(event) => setField(action.id, 'body', event.target.value)}
                      required
                      rows={5}
                    />
                  </>
                ) : (
                  <p className="automation-note">{t('automations.legacyAction')}</p>
                )}
              </section>
            ))}
            <div>
              <Button
                variant="outline"
                onClick={() =>
                  setActions((current) => [
                    ...current,
                    { id: nextId.current++, type: 'send_email', config: {} },
                  ])
                }
              >
                <Plus size={16} />
                <Mail size={16} />
                {t('crm.automationActions.send_email')}
              </Button>
            </div>
            <p className="automation-note">{t('automations.saveNote')}</p>
            <div>
              <Button type="submit" disabled={!valid || pending}>
                {t(pending ? 'automations.saving' : 'automations.save')}
              </Button>
            </div>
          </form>
        )}
      </div>
    </>
  );
}
