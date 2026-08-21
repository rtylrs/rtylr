import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const userClient = createOneKitClient(UserServiceClient, 'recruit');

// recruit/compat/v1 is deliberately absent. That contract does not exist in
// voxie and never has, so no client could be generated for it and importing it
// broke `vite build` for this app. The admin and public hooks use the REST
// transport in ./api, which is what they used before the oneKit migration
// introduced this shim. See #894.
