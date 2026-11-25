import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import GateNotice from '../../../shared/components/GateNotice';
import { useSetTaskTemplate, useTaskTemplates } from '../../tasks/useTaskTemplates';

// Epics describe a goal rather than a piece of work, so they get no template.
const TEMPLATE_TYPES = ['bug', 'feature', 'task', 'chore', 'spike'] as const;
const GATE_ID = 'flow-project-templates-gate';

const fieldStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

/**
 * Project Settings > Templates (T-380): the description each task type starts
 * with in the New task form. Admins edit; everyone else reads.
 */
const ProjectTemplatesSettings: React.FC<{ projectId: string; canManage: boolean }> = ({
  projectId,
  canManage,
}) => {
  const { t } = useTranslation();
  const { data: templates = [] } = useTaskTemplates(projectId);
  return (
    <div className="max-w-[640px] space-y-4">
      <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.taskTemplates.intro')}
      </p>
      <GateNotice id={GATE_ID} reason={canManage ? '' : t('flow.taskTemplates.gateReason')} />
      {TEMPLATE_TYPES.map((type) => (
        <TemplateEditor
          key={type}
          projectId={projectId}
          type={type}
          saved={templates.find((template) => template.taskType === type)?.body ?? ''}
          canManage={canManage}
        />
      ))}
    </div>
  );
};

const TemplateEditor: React.FC<{
  projectId: string;
  type: string;
  saved: string;
  canManage: boolean;
}> = ({ projectId, type, saved, canManage }) => {
  const { t } = useTranslation();
  const id = useId();
  const setTemplate = useSetTaskTemplate();
  const [body, setBody] = useState(saved);
  useEffect(() => setBody(saved), [saved]);
  const dirty = body !== saved;

  return (
    <section
      className="rounded-lg p-4 space-y-2"
      style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
    >
      <div className="flex items-center justify-between gap-2">
        <label
          htmlFor={id}
          className="text-flow-body font-semibold"
          style={{ color: 'var(--flow-text)' }}
        >
          {t(`flow.taskType.${type}`)}
        </label>
        {!saved && (
          <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {type === 'bug' ? t('flow.taskTemplates.usingDefault') : t('flow.taskTemplates.none')}
          </span>
        )}
      </div>
      <textarea
        id={id}
        value={body}
        onChange={(e) => setBody(e.target.value)}
        rows={6}
        dir="auto"
        readOnly={!canManage}
        aria-describedby={canManage ? undefined : GATE_ID}
        placeholder={
          type === 'bug' ? t('flow.taskTemplates.defaultBug') : t('flow.taskTemplates.placeholder')
        }
        className="flow-focus w-full rounded-md px-3 py-2 text-flow-meta font-mono"
        style={fieldStyle}
      />
      {canManage && (
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            disabled={!dirty || setTemplate.isPending}
            onClick={() => setTemplate.mutate({ projectId, taskType: type, body })}
            className="px-3 py-1.5 rounded-sm text-flow-meta font-semibold disabled:opacity-50"
            style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
          >
            {t('flow.taskTemplates.save')}
          </button>
          {type === 'bug' && !body && (
            <button
              type="button"
              onClick={() => setBody(t('flow.taskTemplates.defaultBug'))}
              className="px-3 py-1.5 rounded-sm text-flow-meta"
              style={{ border: '1px solid var(--flow-border-control)', color: 'var(--flow-text)' }}
            >
              {t('flow.taskTemplates.startFromDefault')}
            </button>
          )}
          {saved && (
            <button
              type="button"
              disabled={setTemplate.isPending}
              onClick={() => setTemplate.mutate({ projectId, taskType: type, body: '' })}
              className="px-3 py-1.5 rounded-sm text-flow-meta"
              style={{ color: 'var(--flow-danger-fg)' }}
            >
              {t('flow.taskTemplates.remove')}
            </button>
          )}
        </div>
      )}
    </section>
  );
};

export default ProjectTemplatesSettings;
