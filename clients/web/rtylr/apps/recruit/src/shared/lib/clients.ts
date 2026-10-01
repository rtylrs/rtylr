import { createOneKitClient, createRestRouteClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';

export const userClient = createOneKitClient(UserServiceClient, 'recruit');
export const rest = createRestRouteClient('recruit');

// recruit/compat/v1 is deliberately absent. That contract does not exist in
// voxie and never has, so no client could be generated for it and importing it
// broke `vite build` for this app. The admin and public hooks use the explicit
// REST fallback above for endpoints without a contract. See #894.
