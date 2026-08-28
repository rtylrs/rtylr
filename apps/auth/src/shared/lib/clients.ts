import { createOneKitClient } from '@voxire/api';
import { FeatureServiceClient } from '@voxire/voxie/hub/feature/v1/client';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';

export const licenseClient = createOneKitClient(LicenseServiceClient, 'auth');
export const featureClient = createOneKitClient(FeatureServiceClient, 'auth');
