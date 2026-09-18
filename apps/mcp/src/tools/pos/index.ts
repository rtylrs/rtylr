export { listOrders, getOrder, createOrder, updateOrderStatus, voidOrder } from './orders.js';
export { listMenuItems, getMenuItem, createMenuItem, updateMenuItem, toggleMenuItemAvailability } from './menu-items.js';
export { listSessions, getSession, openSession, closeSession } from './sessions.js';
export { getGiftCard, redeemGiftCard } from './gift-cards.js';
export { kdsListOrders } from './kds.js';

import { listOrders, getOrder, createOrder, updateOrderStatus, voidOrder } from './orders.js';
import { listMenuItems, getMenuItem, createMenuItem, updateMenuItem, toggleMenuItemAvailability } from './menu-items.js';
import { listSessions, getSession, openSession, closeSession } from './sessions.js';
import { getGiftCard, redeemGiftCard } from './gift-cards.js';
import { kdsListOrders } from './kds.js';

export const posTools = [
  listOrders, getOrder, createOrder, updateOrderStatus, voidOrder,
  listMenuItems, getMenuItem, createMenuItem, updateMenuItem, toggleMenuItemAvailability,
  listSessions, getSession, openSession, closeSession,
  getGiftCard, redeemGiftCard,
  kdsListOrders,
];
