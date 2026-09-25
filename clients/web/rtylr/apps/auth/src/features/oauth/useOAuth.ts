import { useMutation } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { createOneKitClient } from '@voxire/api';
import { OauthServiceClient } from '@voxire/voxie/common/oauth/v1/client';
import type { IssueRequest, IssueResponse } from '@voxire/voxie/common/oauth/v1/types';
import { toast } from 'sonner';

const oauthClient = createOneKitClient(OauthServiceClient, 'auth');

export const useIssueAuthCode = () => {
  const { t } = useTranslation();
  return useMutation<IssueResponse, Error, IssueRequest>({
    mutationFn: (req) => oauthClient.issue(req),
    onError: () => toast.error(t('auth.oauth.failed')),
  });
};
