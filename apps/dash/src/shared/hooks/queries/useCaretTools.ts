import { useQuery } from '@tanstack/react-query';
import { listTools, type ToolDef } from '../../api/caret';

export const useCaretTools = (mode?: string) =>
  useQuery<ToolDef[]>({
    queryKey: ['caret', 'tools', mode ?? ''],
    queryFn: () => listTools(mode),
    staleTime: 5 * 60_000,
  });
