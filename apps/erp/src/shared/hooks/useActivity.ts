import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import {
  ActivityServiceClient,
  type ActivityEntry,
} from '@voxire/voxie/hub/activity/v1/service_client';

const activityClient = createOneKitClient(ActivityServiceClient, 'hub');

export type { ActivityEntry };

export function useActivity(limit = 20) {
  return useQuery<ActivityEntry[]>({
    queryKey: ['activity', limit],
    queryFn: async () => {
      const response = await activityClient.listActivity({ limit });
      return response.activities;
    },
  });
}
