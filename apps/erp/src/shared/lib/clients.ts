import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

// licenseClient lives in ./api as a REST adapter: license/v1 has no contract
// in voxie (#894).
export { licenseClient } from './api';
export const userClient = createOneKitClient(UserServiceClient, 'erp');
// hub/team/compat/v1 has no contract in voxie (#894). useErpTeam uses the REST
// transport in ./api, which is what it used before the oneKit migration.
// hub.team.members.v1 is not a substitute: it has listTeam but no inviteMember,
// and removeMember is a delete rather than a deactivate.
