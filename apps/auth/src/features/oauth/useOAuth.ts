import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';

interface IssueRequest {
  client_id: string;
  redirect_uri: string;
  scope: string;
  state: string;
  code_challenge: string;
  code_challenge_method: string;
}

interface IssueResponse {
  redirect_url: string;
}

export const useIssueAuthCode = () =>
  useMutation({
    mutationFn: (req: IssueRequest): Promise<IssueResponse> =>
      unwrap(safe.post<IssueResponse>('/oauth/authorize/issue', req)),
    onError: () => toast.error('Authorization failed. Please try again.'),
  });
