import React from 'react';
import { useTranslation } from 'react-i18next';
import { Globe, Lock } from 'lucide-react';
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

      <div className="flex flex-col gap-1">
        {OPTIONS.map((option) => (
          <label
            key={option}
            className="flex items-center gap-2 min-h-[44px] px-2 rounded-md cursor-pointer text-meta"
            style={{ color: 'var(--flow-text)' }}
          >
            <input
              type="radio"
              name="project-visibility"
              value={option}
              checked={current === option}
              disabled={disabled}
              onChange={() => {
                if (projectId === undefined || current === option) return;
                setVisibility.mutate({ id: projectId, visibility: option });
              }}
              className="h-4 w-4"
            />
            {option === 'public' ? <Globe size={13} aria-hidden /> : <Lock size={13} aria-hidden />}
            <span>{t(`flow.visibility.${option}`)}</span>
          </label>
        ))}
      </div>

      <p className="mt-1 text-meta leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.visibility.publicHint')}
      </p>
      <GateNotice id={GATE_ID} reason={reason} className="mt-1" />
    </fieldset>
  );
};

export default ProjectVisibilityField;
