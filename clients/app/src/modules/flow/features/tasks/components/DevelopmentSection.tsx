import React from 'react';
import { useTranslation } from 'react-i18next';
import { Code2 } from 'lucide-react';

/**
 * Development section of the task panel, shown only on developer projects.
 * Branches, pull requests, deploys and the dev fields are filled in by the
 * Git integration and dev-planning work that follows; until then the section
 * shows where they will appear.
 */
const DevelopmentSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const fields = [
    { key: 'type', label: t('flow.taskDetail.devType') },
    { key: 'points', label: t('flow.taskDetail.devPoints') },
    { key: 'branch', label: t('flow.taskDetail.devBranch') },
  ];
  return (
    <section className="mb-8" aria-labelledby="flow-task-development-heading">
      <div className="flex items-center gap-2 mb-3">
        <Code2 size={12} aria-hidden style={{ color: 'var(--flow-text-muted)' }} />
        <h3
          id="flow-task-development-heading"
          className="text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.taskDetail.development')}
        </h3>
        <span
          className="text-flow-meta px-1.5 rounded-sm"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {t('flowCommon.comingSoon')}
        </span>
      </div>
      <dl className="grid grid-cols-3 gap-2 mb-3">
        {fields.map((field) => (
          <div
            key={field.key}
            className="rounded-md px-3 py-2"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
            }}
          >
            <dt className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {field.label}
            </dt>
            <dd className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.taskDetail.devNotSet')}
            </dd>
          </div>
        ))}
      </dl>
      <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.taskDetail.developmentEmpty')}
      </p>
    </section>
  );
};

export default DevelopmentSection;
