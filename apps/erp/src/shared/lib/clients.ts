import { createOneKitClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient as TeamCompatibilityServiceClient } from '@voxire/voxie/hub/team/compat/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'erp');
export const userClient = createOneKitClient(UserServiceClient, 'erp');
export const teamCompatibilityClient = createOneKitClient(TeamCompatibilityServiceClient, 'erp');
