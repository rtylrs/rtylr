import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const userClient = createOneKitClient(UserServiceClient, 'insights');
