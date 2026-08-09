import { useMutation } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { OauthServiceClient } from '@voxire/voxie/common/oauth/v1/client';
import type { IssueRequest, IssueResponse } from '@voxire/voxie/common/oauth/v1/types';
import { toast } from 'sonner';

const oauthClient = createOneKitClient(OauthServiceClient, 'auth');

export const useIssueAuthCode = () =>
  useMutation<IssueResponse, Error, IssueRequest>({
    mutationFn: (req) => oauthClient.issue(req),
    onError: () => toast.error('Authorization failed. Please try again.'),
  });
