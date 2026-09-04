import React from 'react';
import { useTranslation } from 'react-i18next';
import { Check, Globe, Lock } from 'lucide-react';
import GateNotice from '../../../shared/components/GateNotice';
import { useProject } from '../../projects-v2/useAllProjects';
import { useSetProjectVisibility } from '../../projects-v2/useVisibility';
import type { ProjectVisibility } from '../../projects-v2/useVisibility';

const GATE_ID = 'flow-project-visibility-gate';
const OPTIONS: readonly ProjectVisibility[] = ['private', 'public'];

interface ProjectVisibilityFieldProps {
  projectId: string | undefined;
  /** Whether the caller holds the project-admin gate the server requires. */
  canAdminister: boolean;
}

/**
 * Private/public control for an existing project.
 *
 * Reads the project through the v2 hook rather than flowStore because the store's
 * Project carries neither `visibility` nor `isPersonal`, and both decide what this
 * renders.
 *
 * Three states the server can refuse, each shown as a reason rather than a
 * control that fails:
 *   - not a project admin: SetProjectVisibility applies
 *     RequireProjectRole(FlowRoleAdmin)
 *   - the My Tasks bucket: the server rejects any flip on a personal project
 *   - still loading: no visibility known yet, so neither option is preselected
 */
const ProjectVisibilityField = ({
  projectId,
  canAdminister,
}: ProjectVisibilityFieldProps): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: project, isLoading } = useProject(projectId);
  const setVisibility = useSetProjectVisibility();

  const isPersonal = project?.isPersonal === true;
  const current = project?.visibility;
  const reason = isPersonal
    ? t('flow.projectSettings.visibilityPersonalGate')
    : !canAdminister
      ? t('flow.projectSettings.visibilityGateReason')
      : '';
  const disabled = reason !== '' || isLoading || setVisibility.isPending;

  return (
    <fieldset className="mt-3" disabled={disabled} aria-describedby={reason ? GATE_ID : undefined}>
      <legend className="text-meta font-medium mb-1" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.projectSettings.visibility')}
      </legend>

      {/* Option cards, not bare radios. The native input stays for its
          semantics and keyboard behaviour, visually hidden: a real
          radiogroup, arrow keys and all, rather than divs wearing
          role="radio". What it loses is the browser's own blue fill and OS
          focus ring, which were the only colours on this screen from outside
          the token set. G-C1 and G-C2. */}
      <div className="flex flex-col gap-2 sm:flex-row">
        {OPTIONS.map((option) => {
          const selected = current === option;
          const Icon = option === 'public' ? Globe : Lock;
          return (
            <label
              key={option}
              className="flex min-h-[44px] flex-1 cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-meta transition-colors focus-within:outline focus-within:outline-2 focus-within:outline-offset-2"
              style={{
                color: 'var(--flow-text)',
                background: selected ? 'var(--accent-muted)' : 'var(--flow-input-bg)',
                border: `1px solid ${selected ? 'var(--accent-border)' : 'var(--flow-border-control)'}`,
                outlineColor: 'var(--flow-focus-ring)',
                cursor: disabled ? 'not-allowed' : 'pointer',
              }}
            >
              <input
                type="radio"
                name="project-visibility"
                value={option}
                checked={selected}
                disabled={disabled}
                onChange={() => {
                  if (projectId === undefined || current === option) return;
                  setVisibility.mutate({ id: projectId, visibility: option });
                }}
                className="sr-only"
              />
              <Icon
                size={13}
                aria-hidden
                style={{ color: selected ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)' }}
              />
              <span className={selected ? 'font-semibold' : undefined}>
                {t(`flow.visibility.${option}`)}
              </span>
              {/* The selection is not carried by colour alone: WCAG 1.4.1. */}
              {selected && (
                <Check
                  size={13}
                  aria-hidden
                  className="ms-auto"
                  style={{ color: 'var(--accent-on-muted)' }}
                />
              )}
            </label>
          );
        })}
      </div>

      <p className="mt-1 text-meta leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.visibility.publicHint')}
      </p>
      <GateNotice id={GATE_ID} reason={reason} className="mt-1" />
    </fieldset>
  );
};

export default ProjectVisibilityField;
