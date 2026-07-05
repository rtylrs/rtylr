import { useMutation } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import {
  OauthServiceClient,
  type IssueRequest,
  type IssueResponse,
} from '@voxire/voxie/common/oauth/v1/service_client';
import { toast } from 'sonner';

const oauthClient = createOneKitClient(OauthServiceClient, 'auth');

export const useIssueAuthCode = () =>
  useMutation<IssueResponse, Error, IssueRequest>({
    mutationFn: (req) => oauthClient.issue(req),
    onError: () => toast.error('Authorization failed. Please try again.'),
  });
