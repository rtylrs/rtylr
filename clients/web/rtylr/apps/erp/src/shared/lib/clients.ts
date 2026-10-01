import { createOneKitClient, createRestRouteClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'erp');
export const rest = createRestRouteClient('erp');
export const userClient = createOneKitClient(UserServiceClient, 'erp');
