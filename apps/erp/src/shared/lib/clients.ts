import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient as TeamCompatibilityServiceClient } from '@voxire/voxie/hub/team/compat/v1/client';

// licenseClient lives in ./api as a REST adapter: license/v1 has no contract
// in voxie (#894).
export { licenseClient } from './api';
export const userClient = createOneKitClient(UserServiceClient, 'erp');
export const teamCompatibilityClient = createOneKitClient(TeamCompatibilityServiceClient, 'erp');
