import { createOneKitClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient as CrmCompatibilityServiceClient } from '@voxire/voxie/crm/compat/v1/client';
import { CustomerServiceClient } from '@voxire/voxie/crm/customer/v1/client';
import { SegmentServiceClient } from '@voxire/voxie/crm/segment/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'crm');
export const userClient = createOneKitClient(UserServiceClient, 'crm');
export const crmCompatibilityClient = createOneKitClient(CrmCompatibilityServiceClient, 'crm');
export const crmCustomerClient = createOneKitClient(CustomerServiceClient, 'crm');
export const crmSegmentClient = createOneKitClient(SegmentServiceClient, 'crm');
