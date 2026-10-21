import { useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';

export function useOpenTask(): (taskId: string) => void {
  const navigate = useNavigate();

  return useCallback(
    (taskId: string) => {
      void navigate({
        to: '.',
        search: (previous: Record<string, unknown>) => ({ ...previous, openTask: taskId }),
      });
    },
    [navigate]
  );
}
