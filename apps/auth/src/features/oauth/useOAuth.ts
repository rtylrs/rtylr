import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type { IssueRequest, IssueResponse } from '@voxire/contracts/common/oauth/v1/issue';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';

export const useIssueAuthCode = () =>
  useMutation({
    mutationFn: (req: IssueRequest): Promise<IssueResponse> =>
      unwrap(safe.post<IssueResponse>('/oauth/authorize/issue', req)),
    onError: () => toast.error('Authorization failed. Please try again.'),
  });
