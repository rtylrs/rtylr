import { createOneKitClient, createRestRouteClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const rest = createRestRouteClient('erp');
export const userClient = createOneKitClient(UserServiceClient, 'erp');
