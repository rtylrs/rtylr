import React from 'react';
import { Navigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { flowApi } from '../../shared/services/flowApi';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';

/**
 * Resolves (get-or-creates) the user's personal "My Tasks" project and redirects
 * to its board, so opening Flow / clicking My Tasks lands straight on the kanban
 * instead of a separate page. Falls back to the personal workspace on error.
 */
const MyTasksRedirect = (): React.JSX.Element => {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['my-tasks-project'],
    queryFn: () => unwrap(flowApi.getMyTasksProject()),
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return (
      <div
        style={{
          padding: '32px',
          color: 'var(--flow-text-muted)',
          fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
          fontSize: '13px',
        }}
      >
        Loading…
      </div>
    );
  }

  if (isError || !data?.projectId) {
    return (
      <Navigate to="/w/$workspaceId" params={{ workspaceId: PERSONAL_WORKSPACE_ID }} replace />
    );
  }

  return (
    <Navigate
      to="/w/$workspaceId/p/$projectId"
      params={{ workspaceId: String(data.workspaceId), projectId: String(data.projectId) }}
      replace
    />
  );
};

export default MyTasksRedirect;
