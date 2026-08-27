import { createOneKitClient, createRestRouteClient } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'erp');
export const rest = createRestRouteClient('erp');
export const userClient = createOneKitClient(UserServiceClient, 'erp');
// hub/team/compat/v1 has no contract in voxie (#894). useErpTeam uses the REST
// transport in ./api, which is what it used before the oneKit migration.
// hub.team.members.v1 is not a substitute: it has listTeam but no inviteMember,
// and removeMember is a delete rather than a deactivate.
