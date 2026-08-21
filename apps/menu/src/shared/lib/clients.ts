import { createOneKitClient } from '@voxire/api';
import { BusinessServiceClient } from '@voxire/voxie/hub/business/v1/client';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';

// Menu is a public OneKit consumer. Keep the generated clients in one place so
// feature code cannot quietly grow another transport or handwritten URL.
export const businessClient = createOneKitClient(BusinessServiceClient, 'menu');
export const menuItemClient = createOneKitClient(MenuItemServiceClient, 'menu');
export const orderClient = createOneKitClient(OrderServiceClient, 'menu');
