import { createOneKitClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'insights');
export const userClient = createOneKitClient(UserServiceClient, 'insights');
