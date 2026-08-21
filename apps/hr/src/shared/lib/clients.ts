import { createOneKitClient } from '@voxire/api';
import { safe } from './api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { ShiftServiceClient } from '@voxire/voxie/hr/shift/v1/client';

type HrRequestConfig = {
  params?: unknown;
  data?: unknown;
};

/**
 * The HR app's REST transport.
 *
 * This used to dispatch ~40 paths onto a single `hr.compat.v1`
 * CompatibilityService, one umbrella RPC per legacy route. That contract does
 * not exist in voxie and never has, so the import could not resolve and
 * `vite build` failed for the whole app (#894).
 *
 * Backed by `safe` instead, which reaches the same onion routes these paths
 * already named. The shape is unchanged, so every `hrTransport.get/post/put/
 * delete` call site keeps working: this is the transport those paths were
 * written against before the oneKit migration introduced the shim.
 *
 * Key conversion is handled by @voxire/api's interceptors, so nothing converts
 * here. The previous shim applied `fromOneKitJson` on top, which would have
 * double-converted had it ever run.
 */
export const hrTransport = {
  get: <T = unknown>(path: string, config?: HrRequestConfig) => safe.get<T>(path, config),
  post: <T = unknown>(path: string, data?: unknown) => safe.post<T>(path, data),
  put: <T = unknown>(path: string, data?: unknown) => safe.put<T>(path, data),
  delete: <T = unknown>(path: string, config?: HrRequestConfig) => safe.delete<T>(path, config),
};

// licenseClient now lives in ./api as a REST adapter: license/v1 has no
// contract in voxie (#894).
export { licenseClient } from './api';
export const userClient = createOneKitClient(UserServiceClient, 'hr');
