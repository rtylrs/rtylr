import { createOneKitClient, createRestRouteClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { AiServiceClient } from '@voxire/voxie/hub/ai/v1/client';
import { AppServiceClient } from '@voxire/voxie/hub/app/v1/client';
import { BusinessServiceClient } from '@voxire/voxie/hub/business/v1/client';
import { ConfigServiceClient } from '@voxire/voxie/hub/config/v1/client';
import { DeveloperServiceClient } from '@voxire/voxie/hub/developer/v1/client';
import { IntegrationServiceClient } from '@voxire/voxie/hub/integration/v1/client';
import { NotificationServiceClient } from '@voxire/voxie/hub/notification/v1/client';
import { PermissionServiceClient } from '@voxire/voxie/hub/permission/v1/client';
import { SecurityServiceClient } from '@voxire/voxie/hub/security/v1/client';
import { SubscriptionServiceClient } from '@voxire/voxie/hub/subscription/v1/client';
import { AccessServiceClient } from '@voxire/voxie/hub/team/access/v1/client';
import { GroupsServiceClient } from '@voxire/voxie/hub/team/groups/v1/client';
import { OnboardingServiceClient } from '@voxire/voxie/hub/team/onboarding/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import { TableServiceClient } from '@voxire/voxie/pos/table/v1/client';

export const aiClient = createOneKitClient(AiServiceClient, 'dash');
export const appClient = createOneKitClient(AppServiceClient, 'dash');
export const businessClient = createOneKitClient(BusinessServiceClient, 'dash');
export const configClient = createOneKitClient(ConfigServiceClient, 'dash');
export const developerClient = createOneKitClient(DeveloperServiceClient, 'dash');
export const integrationClient = createOneKitClient(IntegrationServiceClient, 'dash');
export const notificationClient = createOneKitClient(NotificationServiceClient, 'dash');
export const permissionClient = createOneKitClient(PermissionServiceClient, 'dash');
export const securityClient = createOneKitClient(SecurityServiceClient, 'dash');
export const subscriptionClient = createOneKitClient(SubscriptionServiceClient, 'dash');
export const accessClient = createOneKitClient(AccessServiceClient, 'dash');
export const groupsClient = createOneKitClient(GroupsServiceClient, 'dash');
export const onboardingClient = createOneKitClient(OnboardingServiceClient, 'dash');
export const userClient = createOneKitClient(UserServiceClient, 'dash');
export const orderClient = createOneKitClient(OrderServiceClient, 'dash');
export const tableClient = createOneKitClient(TableServiceClient, 'dash');
export const licenseClient = createOneKitClient(LicenseServiceClient, 'dash');
export const rest = createRestRouteClient('dash');

// pos/orders/compat/v1 and hub/settings/compat/v1 have no contract in voxie
// (#894). useOrders and the SMTP test use the REST transport in ./api, which is
// what they used before the oneKit migration.
