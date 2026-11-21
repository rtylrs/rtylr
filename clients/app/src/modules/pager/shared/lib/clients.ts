import { createOneKitClient } from '@voxire/api';
import { PagerServiceClient } from '@voxire/voxie/pager/v1/client';

// The customer pages are anonymous and sit beside the public menu; the staff
// board belongs to the signed-in Portal. Each gets the transport of its side.
export const publicPagerClient = createOneKitClient(PagerServiceClient, 'menu');
export const staffPagerClient = createOneKitClient(PagerServiceClient, 'dash');
