import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Navigate, useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { FolderKanban } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  requestErrorStatus,
  useDashboardProjects,
} from '../../shared/hooks/queries/useInsightsDashboard';
import ProjectDashboardPanel from '../../shared/components/ProjectDashboardPanel';
import { CardSkeleton } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

/**
 * /insights/projects/$projectId: one Work project's dashboard with a picker
 * of the projects the viewer may open. /insights/projects opens the first.
 */
const ProjectDashboardPage: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId = '' } = useParams({ strict: false });
  const { data: projects, isLoading, isError, error, refetch } = useDashboardProjects();
  const listed = !!projects && projects.some((p) => p.id === projectId);

  const header = (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.projectDashboard.title')}
        </h1>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          {t('insights.projectDashboard.subtitle')}
        </p>
      </div>
      {projects && projects.length > 0 && (
        <label className="flex flex-col gap-1 text-xs sm:w-72">
          <span style={{ color: 'var(--hub-text-muted)' }}>
            {t('insights.projectDashboard.pickerLabel')}
          </span>
          <select
            value={listed ? projectId : ''}
            onChange={(event) =>
              void navigate({
                to: '/insights/projects/$projectId',
                params: { projectId: event.target.value },
              })
            }
            className="h-9 rounded-lg px-3 text-sm"
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            {!listed && (
              <option value="" disabled>
                {t('insights.projectDashboard.pickerPlaceholder')}
              </option>
            )}
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.visibility === 'private'
                  ? t('insights.projectDashboard.privateOption', { name: project.name })
                  : project.name}
              </option>
            ))}
          </select>
        </label>
      )}
    </div>
  );

  let body: React.ReactNode;
  if (isError) {
    body =
      requestErrorStatus(error) === 403 ? (
        <EmptyState icon={FolderKanban} title={t('insights.projectDashboard.noAccess')} />
      ) : (
        <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />
      );
  } else if (isLoading || !projects) {
    body = (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4" aria-busy="true">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  } else if (projectId === '') {
    // /insights/projects: open the first project, or explain why there is none.
    if (projects.length > 0) {
      return (
        <Navigate
          to="/insights/projects/$projectId"
          params={{ projectId: projects[0].id }}
          replace
        />
      );
    }
    body = (
      <EmptyState
        icon={FolderKanban}
        title={t('insights.projectDashboard.noProjectsTitle')}
        description={t('insights.projectDashboard.noProjectsDescription')}
      />
    );
  } else {
    // The panel asks for the project itself, so a private project the viewer
    // is not a member of reads as not found even from a typed URL.
    body = <ProjectDashboardPanel projectId={projectId} />;
  }

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <HelmetContainer title={t('insights.projectDashboard.metaTitle')} />
      {header}
      {body}
    </div>
  );
};

export default ProjectDashboardPage;
