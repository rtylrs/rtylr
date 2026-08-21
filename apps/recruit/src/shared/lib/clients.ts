import { createOneKitClient, fromOneKitJson, oneKitResult, toOneKitJson } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient } from '@voxire/voxie/recruit/compat/v1/client';

export const userClient = createOneKitClient(UserServiceClient, 'recruit');
export const recruitCompatibilityClient = createOneKitClient(CompatibilityServiceClient, 'recruit');

export function recruitResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((value) => fromOneKitJson<T>(value));
}

export function recruitBody(value: unknown) {
  return { body: toOneKitJson(value ?? {}) };
}
