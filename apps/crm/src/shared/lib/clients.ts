import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient as CrmCompatibilityServiceClient } from '@voxire/voxie/crm/compat/v1/client';
import { CustomerServiceClient } from '@voxire/voxie/crm/customer/v1/client';
import { SegmentServiceClient } from '@voxire/voxie/crm/segment/v1/client';

// licenseClient lives in ./api as a REST adapter: license/v1 has no contract
// in voxie (#894).
export { licenseClient } from './api';
export const userClient = createOneKitClient(UserServiceClient, 'crm');
export const crmCompatibilityClient = createOneKitClient(CrmCompatibilityServiceClient, 'crm');
export const crmCustomerClient = createOneKitClient(CustomerServiceClient, 'crm');
export const crmSegmentClient = createOneKitClient(SegmentServiceClient, 'crm');
