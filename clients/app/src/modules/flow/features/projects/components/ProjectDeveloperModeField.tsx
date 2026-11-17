import React from 'react';
import { useTranslation } from 'react-i18next';
import { Code2 } from 'lucide-react';
import GateNotice from '../../../shared/components/GateNotice';
import { useProject } from '../../projects-v2/useAllProjects';
import { useSetProjectDeveloperMode } from '../../projects-v2/useDeveloperMode';

const GATE_ID = 'flow-project-developer-gate';
const HINT_ID = 'flow-project-developer-hint';

interface ProjectDeveloperModeFieldProps {
  projectId: string | undefined;
  /** Whether the caller holds the project-admin gate the server requires. */
  canAdminister: boolean;
}

/** "Developer project" switch. Turning it off hides developer features but deletes nothing. */
const ProjectDeveloperModeField = ({
  projectId,
  canAdminister,
}: ProjectDeveloperModeFieldProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const { data: project, isLoading } = useProject(projectId);
  const setDeveloperMode = useSetProjectDeveloperMode();

  // My Tasks is never a software project; the server refuses the switch.
  if (project?.isPersonal) return null;

  const enabled = project?.developerMode === true;
  const reason = canAdminister ? '' : t('flow.projectSettings.developerModeGateReason');
  const disabled = reason !== '' || isLoading || !project || setDeveloperMode.isPending;

  return (
    <div className="mt-3">
      <label
        className="flex min-h-[44px] items-start gap-3 rounded-md px-3 py-2"
        style={{
          background: 'var(--flow-input-bg)',
          border: '1px solid var(--flow-border-control)',
          cursor: disabled ? 'not-allowed' : 'pointer',
        }}
      >
        <Code2
          size={14}
          aria-hidden
          className="mt-0.5"
          style={{ color: 'var(--flow-text-muted)' }}
        />
        <span className="flex-1">
          <span className="block text-flow-meta font-medium" style={{ color: 'var(--flow-text)' }}>
            {t('flow.projectSettings.developerMode')}
          </span>
          <span
            id={HINT_ID}
            className="block text-flow-meta leading-snug"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.projectSettings.developerModeHint')}
          </span>
        </span>
        <input
          type="checkbox"
          role="switch"
          checked={enabled}
          disabled={disabled}
          aria-describedby={reason ? `${HINT_ID} ${GATE_ID}` : HINT_ID}
          onChange={(e) => {
            if (projectId === undefined) return;
            setDeveloperMode.mutate({ id: projectId, enabled: e.target.checked });
          }}
          className="mt-1 h-4 w-4"
          style={{ accentColor: 'var(--accent)' }}
        />
      </label>
      <GateNotice id={GATE_ID} reason={reason} className="mt-1" />
    </div>
  );
};

export default ProjectDeveloperModeField;
