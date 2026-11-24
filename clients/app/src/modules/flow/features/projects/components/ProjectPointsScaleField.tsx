import React from 'react';
import { useTranslation } from 'react-i18next';
import FlowSelect from '../../../shared/components/FlowSelect';
import GateNotice from '../../../shared/components/GateNotice';
import {
  useIsDeveloperProject,
  useProjectPointsScale,
  useSetProjectPointsScale,
} from '../../projects-v2/useDeveloperMode';
import { POINTS_SCALES, type PointsScale } from '../../../shared/lib/storyPoints';

const GATE_ID = 'flow-project-points-scale-gate';

/** Story points scale for a developer project (T-381). */
const ProjectPointsScaleField = ({
  projectId,
  canAdminister,
}: {
  projectId: string | undefined;
  canAdminister: boolean;
}): React.JSX.Element | null => {
  const { t } = useTranslation();
  const isDeveloperProject = useIsDeveloperProject(projectId);
  const scale = useProjectPointsScale(projectId);
  const setScale = useSetProjectPointsScale();
  if (!isDeveloperProject || projectId === undefined) return null;
  const reason = canAdminister ? '' : t('flow.projectSettings.developerModeGateReason');

  return (
    <div className="mt-3 space-y-1">
      <span className="block text-flow-meta font-medium" style={{ color: 'var(--flow-text)' }}>
        {t('flow.projectSettings.pointsScale')}
      </span>
      <span className="block text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.projectSettings.pointsScaleHint')}
      </span>
      <div className="max-w-[280px]">
        <FlowSelect
          label={t('flow.projectSettings.pointsScale')}
          value={scale}
          disabled={reason !== '' || setScale.isPending}
          options={POINTS_SCALES.map((value) => ({
            value,
            label: t(`flow.projectSettings.pointsScales.${value}`),
          }))}
          onChange={(value) => setScale.mutate({ id: projectId, scale: value as PointsScale })}
        />
      </div>
      <GateNotice id={GATE_ID} reason={reason} className="mt-1" />
    </div>
  );
};

export default ProjectPointsScaleField;
