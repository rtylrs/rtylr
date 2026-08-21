import { createOneKitClient } from '@voxire/api';
import { FeatureServiceClient } from '@voxire/voxie/hub/feature/v1/client';

// licenseClient lives in ./api as a REST adapter: license/v1 has no contract
// in voxie (#894).
export { licenseClient } from './api';
export const featureClient = createOneKitClient(FeatureServiceClient, 'auth');
