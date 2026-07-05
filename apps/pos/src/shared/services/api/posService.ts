import { ResultAsync, unwrap, type ApiError } from '@voxire/api';
import { AccessServiceClient } from '@voxire/voxie/pos/access/v1/service_client';
import { AuditServiceClient } from '@voxire/voxie/pos/audit/v1/service_client';
import { CashServiceClient } from '@voxire/voxie/pos/cash/v1/service_client';
import { ComboServiceClient } from '@voxire/voxie/pos/combo/v1/service_client';
import { CustomerServiceClient } from '@voxire/voxie/pos/customer/v1/service_client';
import { DaySummaryServiceClient } from '@voxire/voxie/pos/day_summary/v1/service_client';
import { FloorServiceClient } from '@voxire/voxie/pos/floor/v1/service_client';
import { GiftCardServiceClient } from '@voxire/voxie/pos/gift_card/v1/service_client';
import { KDSServiceClient } from '@voxire/voxie/pos/kds/v1/service_client';
import { LocationServiceClient } from '@voxire/voxie/pos/location/v1/service_client';
import { LoyaltyServiceClient } from '@voxire/voxie/pos/loyalty/v1/service_client';
import { ManagerServiceClient } from '@voxire/voxie/pos/manager/v1/service_client';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/service_client';
import { ModifierServiceClient } from '@voxire/voxie/pos/modifier/v1/service_client';
import { OrderDiscountServiceClient } from '@voxire/voxie/pos/orders/discounts/v1/service_client';
import { OrderHoldServiceClient } from '@voxire/voxie/pos/orders/holds/v1/service_client';
import { OrderItemServiceClient } from '@voxire/voxie/pos/orders/items/v1/service_client';
import { OrderPaymentServiceClient } from '@voxire/voxie/pos/orders/payments/v1/service_client';
import { OrderRefundServiceClient } from '@voxire/voxie/pos/orders/refunds/v1/service_client';
import { OrderSplitServiceClient } from '@voxire/voxie/pos/orders/splits/v1/service_client';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/service_client';
import { ReceiptServiceClient } from '@voxire/voxie/pos/receipt/v1/service_client';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/service_client';
import { SettingsServiceClient } from '@voxire/voxie/pos/settings/v1/service_client';
import { StockServiceClient } from '@voxire/voxie/pos/stock/v1/service_client';
import { TableServiceClient } from '@voxire/voxie/pos/table/v1/service_client';
import { TerminalServiceClient } from '@voxire/voxie/pos/terminal/v1/service_client';
import { UserServiceClient } from '@voxire/voxie/pos/user/v1/service_client';
import { safe } from './api';
import type { Order, PaymentEntry } from '../../types/pos';

// ── Request body types ─────────────────────────────────────────────────────

export interface BranchPayload {
  name: string;
  address?: string;
  phone?: string;
  [key: string]: unknown;
}

export interface FloorPayload {
  name: string;
  branchId?: number;
  [key: string]: unknown;
}

export interface TerminalPayload {
  name: string;
  branchId?: number;
  branch_id?: number;
  deviceId?: string | null;
  device_id?: string | null;
  is_active?: boolean;
  [key: string]: unknown;
}

export interface TablePayload {
  name?: string;
  label?: string;
  location_id?: number;
  locationId?: number;
  floorId?: number;
  seats?: number;
  capacity?: number;
  [key: string]: unknown;
}

/** Line item matching the backend CreateOrderItemInput contract (snake_case). */
export interface CreateOrderItemPayload {
  menu_item_id: number;
  name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  discount_pct?: number;
  tax_rate?: number;
  notes?: string;
}

/** Matches the backend CreateOrderRequest contract (snake_case). */
export interface CreateOrderPayload {
  session_id?: number;
  table_id?: number;
  location_id?: number;
  type?: string;
  customer_id?: number;
  customer_name?: string;
  payment_method?: string;
  notes?: string;
  kitchen_notes?: string;
  items: CreateOrderItemPayload[];
}

export type OrderPayload = CreateOrderPayload;

/** Matches the backend ProcessPaymentRequest contract. */
export interface PaymentPayload {
  payment_method: string;
  amount_paid: number;
}

/** Matches the backend DiscountOrderRequest contract. */
export interface DiscountPayload {
  discount_type: 'flat' | 'percent';
  discount_value: number;
  reason?: string;
}

export interface SplitPayload {
  parts: number;
  splits?: number[][];
  [key: string]: unknown;
}

export interface HoldPayload {
  reason?: string;
  [key: string]: unknown;
}

export interface VoidRefundPayload {
  reason?: string;
  [key: string]: unknown;
}

export interface ModifierPayload {
  name: string;
  options?: Array<{ name: string; price: number }>;
  [key: string]: unknown;
}

export interface ComboPayload {
  name: string;
  items?: number[];
  price?: number;
  [key: string]: unknown;
}

/** Matches the backend OpenSessionRequest contract. */
export interface OpenSessionPayload {
  location_id?: number;
  staff_id?: number | null;
  opening_float?: number;
}

/** Matches the backend CloseSessionRequest contract (id goes in the URL). */
export interface CloseSessionPayload {
  closing_float: number;
}

/** Matches the backend cash movement contracts (drop/pickup/adjust). */
export interface CashMovementPayload {
  branch_id: number;
  terminal_id?: number | null;
  session_id: number;
  employee_id: number;
  amount: number;
  reason?: string | null;
}

export interface GiftCardPayload {
  amount: number;
  code?: string;
  [key: string]: unknown;
}

export interface GiftCardRedeemPayload {
  amount: number;
  orderId?: number;
  [key: string]: unknown;
}

export interface CustomerPayload {
  name: string;
  email?: string;
  phone?: string;
  [key: string]: unknown;
}

export interface LoyaltyConfigPayload {
  pointsPerDollar?: number;
  redeemRate?: number;
  [key: string]: unknown;
}

export interface LoyaltyPayload {
  customer_id: number;
  points: number;
  order_id?: number;
  [key: string]: unknown;
}

export type POSBusinessProfile = 'retail' | 'restaurant' | 'quick_service';
export type POSTheme = 'dark' | 'light' | 'midnight' | 'slate' | 'aurora';

export interface POSFeatureSettingsPayload {
  tables: boolean;
  kitchenDisplay: boolean;
  modifiers: boolean;
  combos: boolean;
  tips: boolean;
  giftCards: boolean;
  customers: boolean;
  inventory: boolean;
  cashDrawer: boolean;
}

export interface POSBusinessSettingsPayload {
  id?: number;
  business_id?: number;
  business_profile: POSBusinessProfile;
  theme: POSTheme;
  features: POSFeatureSettingsPayload;
  receipt_footer: string;
  created_at?: string;
  updated_at?: string;
}

export interface POSBusinessSettingsResponse {
  settings: POSBusinessSettingsPayload;
  message?: string;
}

export interface POSBusinessSettingsUpdatePayload {
  business_profile?: POSBusinessProfile;
  theme?: POSTheme;
  features?: POSFeatureSettingsPayload;
  receipt_footer?: string;
}

/** Matches the backend stock movement contracts (receive/adjust/waste). */
export interface StockPayload {
  branch_id: number;
  product_id: number;
  quantity: number;
  note?: string | null;
  employee_id?: number | null;
}

/** Matches the backend POS stock transfer contract. */
export interface StockTransferPayload {
  from_branch_id: number;
  to_branch_id: number;
  product_id: number;
  quantity: number;
  note?: string | null;
  employee_id?: number | null;
}

/** Matches the backend CreatePOSUserRequest contract (snake_case). */
export interface POSUserPayload {
  name: string;
  pin: string;
  location_id: number;
  role?: string;
  hourly_rate?: number;
  max_discount_pct?: number;
  can_void?: boolean;
  can_refund?: boolean;
  can_open_drawer?: boolean;
  can_apply_discount?: boolean;
  [key: string]: unknown;
}

export interface MemberLocationPayload {
  userId?: number;
  user_id?: number;
  locationId: number;
  canManagePosUsers?: boolean;
  canManageTerminals?: boolean;
  canViewReports?: boolean;
  [key: string]: unknown;
}

export interface MenuItemUpdatePayload {
  name?: string;
  price?: number;
  available?: boolean;
  [key: string]: unknown;
}

export interface OrderFilterParams {
  status?: string;
  search?: string;
  page?: number;
  page_size?: number;
  date?: string;
  tableId?: number;
  [key: string]: unknown;
}

export interface MenuItemFilterParams {
  categoryId?: number;
  available?: boolean;
  [key: string]: unknown;
}

export interface ReceiptFilterParams {
  date?: string;
  orderId?: number;
  [key: string]: unknown;
}

export interface AuditFilterParams {
  startDate?: string;
  endDate?: string;
  type?: string;
  [key: string]: unknown;
}

function toTerminalPayload(data: Partial<TerminalPayload>) {
  const deviceId = data.device_id ?? data.deviceId;
  return {
    name: data.name,
    branch_id: data.branch_id ?? data.branchId,
    ...(deviceId ? { device_id: deviceId } : {}),
    ...(typeof data.is_active === 'boolean' ? { is_active: data.is_active } : {}),
  };
}

function toTablePayload(data: Partial<TablePayload>) {
  return {
    label: data.label ?? data.name,
    location_id: data.location_id ?? data.locationId ?? data.floorId,
    seats: data.seats ?? data.capacity,
  };
}

const ONEKIT_BASE_URL = 'https://onekit.local';

function toApiError(error: unknown): ApiError {
  if (error && typeof error === 'object') {
    const record = error as Record<string, unknown>;
    const status =
      typeof record.status === 'number'
        ? record.status
        : typeof record.statusCode === 'number'
          ? record.statusCode
          : 0;
    return {
      status,
      code: status >= 500 ? 'SERVER_ERROR' : status === 403 ? 'FORBIDDEN' : status === 401 ? 'UNAUTHORIZED' : 'UNKNOWN',
      message: typeof record.message === 'string' ? record.message : 'An unexpected error occurred.',
      cause: error,
    };
  }
  return {
    status: 0,
    code: 'UNKNOWN',
    message: 'An unexpected error occurred.',
    cause: error,
  };
}

function paramsFromUrl(url: URL): Record<string, string> | undefined {
  const params = Object.fromEntries(url.searchParams.entries());
  return Object.keys(params).length > 0 ? params : undefined;
}

function bodyFromInit(init?: RequestInit): unknown {
  if (typeof init?.body !== 'string' || init.body.length === 0) return undefined;
  try {
    return JSON.parse(init.body);
  } catch {
    return init.body;
  }
}

function responseFromJson(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data ?? {}), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

async function onekitSafeFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const url = new URL(
    typeof input === 'string' || input instanceof URL ? String(input) : input.url,
    ONEKIT_BASE_URL
  );
  const path = url.pathname;
  const params = paramsFromUrl(url);
  const body = bodyFromInit(init);
  const method = (init?.method ?? 'GET').toUpperCase();

  try {
    switch (method) {
      case 'GET':
        return responseFromJson(await unwrap(safe.get(path, params ? { params } : undefined)));
      case 'POST':
        return responseFromJson(await unwrap(safe.post(path, body)));
      case 'PUT':
        return responseFromJson(await unwrap(safe.put(path, body)));
      case 'PATCH':
        return responseFromJson(await unwrap(safe.patch(path, body)));
      case 'DELETE':
        return responseFromJson(await unwrap(safe.delete(path, params ? { params } : undefined)));
      default:
        throw new Error(`Unsupported one-kit method ${method}`);
    }
  } catch (error) {
    const apiError = toApiError(error);
    if (apiError.status <= 0) throw error;
    return responseFromJson(
      {
        error: {
          code: apiError.code,
          message: apiError.message,
        },
        status: 'error',
      },
      apiError.status
    );
  }
}

const onekitOptions = { fetch: onekitSafeFetch };
const accessClient = new AccessServiceClient(ONEKIT_BASE_URL, onekitOptions);
const auditClient = new AuditServiceClient(ONEKIT_BASE_URL, onekitOptions);
const cashClient = new CashServiceClient(ONEKIT_BASE_URL, onekitOptions);
const comboClient = new ComboServiceClient(ONEKIT_BASE_URL, onekitOptions);
const customerClient = new CustomerServiceClient(ONEKIT_BASE_URL, onekitOptions);
const daySummaryClient = new DaySummaryServiceClient(ONEKIT_BASE_URL, onekitOptions);
const floorClient = new FloorServiceClient(ONEKIT_BASE_URL, onekitOptions);
const giftCardClient = new GiftCardServiceClient(ONEKIT_BASE_URL, onekitOptions);
const kdsClient = new KDSServiceClient(ONEKIT_BASE_URL, onekitOptions);
const locationClient = new LocationServiceClient(ONEKIT_BASE_URL, onekitOptions);
const loyaltyClient = new LoyaltyServiceClient(ONEKIT_BASE_URL, onekitOptions);
const managerClient = new ManagerServiceClient(ONEKIT_BASE_URL, onekitOptions);
const menuItemClient = new MenuItemServiceClient(ONEKIT_BASE_URL, onekitOptions);
const modifierClient = new ModifierServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderClient = new OrderServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderDiscountClient = new OrderDiscountServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderHoldClient = new OrderHoldServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderItemClient = new OrderItemServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderPaymentClient = new OrderPaymentServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderRefundClient = new OrderRefundServiceClient(ONEKIT_BASE_URL, onekitOptions);
const orderSplitClient = new OrderSplitServiceClient(ONEKIT_BASE_URL, onekitOptions);
const receiptClient = new ReceiptServiceClient(ONEKIT_BASE_URL, onekitOptions);
const sessionClient = new SessionServiceClient(ONEKIT_BASE_URL, onekitOptions);
const settingsClient = new SettingsServiceClient(ONEKIT_BASE_URL, onekitOptions);
const stockClient = new StockServiceClient(ONEKIT_BASE_URL, onekitOptions);
const tableClient = new TableServiceClient(ONEKIT_BASE_URL, onekitOptions);
const terminalClient = new TerminalServiceClient(ONEKIT_BASE_URL, onekitOptions);
const userClient = new UserServiceClient(ONEKIT_BASE_URL, onekitOptions);

function onekit<T>(promise: Promise<T>): ResultAsync<any, ApiError> {
  return ResultAsync.fromPromise(promise, toApiError) as ResultAsync<any, ApiError>;
}

function id(value: unknown): string {
  return value == null || value === '' ? '0' : String(value);
}

function text(value: unknown): string {
  return typeof value === 'string' ? value : '';
}

function numberValue(value: unknown): number {
  return typeof value === 'number' ? value : 0;
}

function boolValue(value: unknown): boolean {
  return typeof value === 'boolean' ? value : false;
}

function field(data: Record<string, unknown>, name: string): unknown {
  return data[name];
}

function toSettingsUpdate(data: POSBusinessSettingsUpdatePayload) {
  return {
    businessProfile: data.business_profile ?? '',
    theme: data.theme ?? '',
    features: data.features,
    receiptFooter: data.receipt_footer ?? '',
  };
}

function toCreateOrder(data: OrderPayload) {
  return {
    sessionId: id(data.session_id),
    tableId: id(data.table_id),
    locationId: id(data.location_id),
    type: data.type ?? '',
    customerId: id(data.customer_id),
    customerName: data.customer_name ?? '',
    paymentMethod: data.payment_method ?? '',
    notes: data.notes ?? '',
    kitchenNotes: data.kitchen_notes ?? '',
    items: data.items.map((item) => ({
      menuItemId: id(item.menu_item_id),
      name: item.name,
      sku: item.sku ?? '',
      quantity: item.quantity,
      unitPrice: item.unit_price,
      discountPct: item.discount_pct ?? 0,
      taxRate: item.tax_rate ?? 0,
      notes: item.notes ?? '',
    })),
  };
}

function toCashMovement(data: CashMovementPayload) {
  return {
    branchId: id(data.branch_id),
    terminalId: id(data.terminal_id),
    sessionId: id(data.session_id),
    employeeId: id(data.employee_id),
    amount: data.amount,
    reason: data.reason ?? '',
  };
}

function toStock(data: StockPayload) {
  return {
    branchId: id(data.branch_id),
    productId: id(data.product_id),
    quantity: data.quantity,
    note: data.note ?? '',
    employeeId: id(data.employee_id),
  };
}

function toLocation(data: BranchPayload) {
  return {
    name: data.name,
    address: text(data.address),
    phone: text(data.phone),
    isActive: typeof data.is_active === 'boolean' ? data.is_active : null,
  };
}

function toFloor(data: FloorPayload) {
  const record = data as Record<string, unknown>;
  return {
    branchId: id(data.branchId ?? field(record, 'branch_id')),
    name: data.name,
    sortOrder: numberValue(field(record, 'sortOrder') ?? field(record, 'sort_order')),
  };
}

function toTerminal(data: Partial<TerminalPayload>) {
  const mapped = toTerminalPayload(data);
  return {
    branchId: id(mapped.branch_id),
    name: text(mapped.name),
    deviceId: text(mapped.device_id),
    isActive: typeof mapped.is_active === 'boolean' ? mapped.is_active : true,
  };
}

function toTableBody(data: Partial<TablePayload>, status = '') {
  const mapped = toTablePayload(data);
  return {
    locationId: mapped.location_id == null ? null : id(mapped.location_id),
    label: text(mapped.label),
    seats: numberValue(mapped.seats),
    status,
  };
}

// ── Service ────────────────────────────────────────────────────────────────

export const posService = {
  // ── Business POS settings ─────────────────────────────────────
  getPOSSettings: () => onekit(settingsClient.getPOSSettings({})),
  updatePOSSettings: (data: POSBusinessSettingsUpdatePayload) =>
    onekit(settingsClient.updatePOSSettings(toSettingsUpdate(data))),

  // ── Branches (locations) ──────────────────────────────────────
  getBranches: () => onekit(locationClient.listLocations({})),
  createBranch: (data: BranchPayload) => onekit(locationClient.createLocation(toLocation(data))),
  updateBranch: (id: number, data: Partial<BranchPayload>) =>
    onekit(locationClient.updateLocation({ id: String(id), body: toLocation(data as BranchPayload) })),

  // ── Floors ────────────────────────────────────────────────────
  getFloors: (branchId: number) =>
    onekit(floorClient.listFloors({ branchId: String(branchId) })),
  createFloor: (data: FloorPayload) => onekit(floorClient.createFloor(toFloor(data))),
  updateFloor: (id: number, data: Partial<FloorPayload>) =>
    onekit(
      floorClient.updateFloor({
        id: String(id),
        body: { name: text(data.name), sortOrder: numberValue(data.sortOrder ?? data.sort_order) },
      })
    ),
  deleteFloor: (id: number) => onekit(floorClient.deleteFloor({ id: String(id) })),

  // ── Terminals ─────────────────────────────────────────────────
  getTerminals: () => onekit(terminalClient.listTerminals({})),
  createTerminal: (data: TerminalPayload) =>
    onekit(terminalClient.createTerminal(toTerminal(data))),
  updateTerminal: (id: number, data: Partial<TerminalPayload>) =>
    onekit(terminalClient.updateTerminal({ id: String(id), body: toTerminal(data) })),
  heartbeat: (terminalId: number) =>
    onekit(terminalClient.heartbeatTerminal({ terminalId: String(terminalId) })),

  // ── Tables ────────────────────────────────────────────────────
  getTables: () => onekit(tableClient.listTables({})),
  createTable: (data: TablePayload) => onekit(tableClient.createTable(toTableBody(data))),
  updateTable: (id: number, data: Partial<TablePayload>) =>
    onekit(tableClient.updateTable({ id: String(id), body: toTableBody(data) })),
  updateTableStatus: (id: number, status: string) =>
    onekit(tableClient.updateTable({ id: String(id), body: toTableBody({}, status) })),
  deleteTable: (id: number) => onekit(tableClient.deleteTable({ id: String(id) })),

  // ── Orders ────────────────────────────────────────────────────
  getOrders: (params?: OrderFilterParams) =>
    onekit(
      orderClient.listOrders({
        sessionId: id(params?.sessionId),
        status: params?.status ?? '',
        dateFrom: text(params?.dateFrom ?? params?.date),
        dateTo: text(params?.dateTo),
        search: params?.search ?? '',
        page: numberValue(params?.page),
        pageSize: numberValue(params?.page_size ?? params?.pageSize),
      })
    ),
  getOrder: (id: number) => onekit(orderClient.getOrderDetail({ orderId: String(id) })),
  createOrder: (data: OrderPayload) => onekit(orderClient.createOrder(toCreateOrder(data))),
  updateOrderStatus: (id: number, status: string) =>
    onekit(orderClient.updateOrderStatus({ orderId: String(id), body: { status } })),
  chargeOrder: (id: number, data: PaymentPayload) =>
    onekit(
      orderPaymentClient.processPayment({
        orderId: String(id),
        paymentMethod: data.payment_method,
        amountPaid: data.amount_paid,
      })
    ),
  voidOrder: (id: number, data: VoidRefundPayload) =>
    onekit(orderClient.voidOrder({ orderId: String(id), reason: data.reason ?? '' })),
  refundOrder: (id: number, data: VoidRefundPayload) =>
    onekit(
      orderRefundClient.processRefund({
        orderId: String(id),
        amount: numberValue(data.amount),
        reason: data.reason ?? '',
      })
    ),
  reopenOrder: (id: number) =>
    onekit(orderClient.updateOrderStatus({ orderId: String(id), body: { status: 'pending' } })),
  applyDiscount: (id: number, data: DiscountPayload) =>
    onekit(
      orderDiscountClient.discountOrder({
        orderId: String(id),
        discountType: data.discount_type,
        discountValue: data.discount_value,
        reason: data.reason ?? '',
      })
    ),
  splitOrder: (id: number, data: SplitPayload) =>
    onekit(
      orderSplitClient.splitOrder({
        orderId: String(id),
        splits: Array.isArray(data.splits)
          ? data.splits.map((itemIds) => ({ itemIds: itemIds.map(String) }))
          : [],
      })
    ),
  holdOrder: (id: number, data: HoldPayload) =>
    onekit(orderHoldClient.holdOrder({ orderId: String(id), reason: data.reason ?? '' })),
  getHeldOrders: () => onekit(orderHoldClient.listHeldOrders({})),
  sendToKitchen: (id: number) => onekit(orderClient.updateOrderStatus({ orderId: String(id), body: { status: 'preparing' } })),
  transferTable: (id: number, newTableId: number) =>
    safe.post(`/api/pos/orders/v1/orders/${id}/transfer-table`, { new_table_id: newTableId }),
  voidOrderItem: (orderId: number, itemId: number) =>
    onekit(
      orderItemClient.updateOrderItemStatus({
        orderId: String(orderId),
        itemId: String(itemId),
        body: { status: 'voided' },
      })
    ),

  // ── Menu / Products ───────────────────────────────────────────
  getMenuItems: (params?: MenuItemFilterParams) =>
    params ? safe.get('/api/pos/menu-item/v1/menu-items', { params }) : onekit(menuItemClient.listMenuItems({})),
  updateMenuItem: (id: number, data: MenuItemUpdatePayload) =>
    safe.put(`/api/pos/menu-item/v1/menu-items/${id}`, data),
  getModifiers: () => onekit(modifierClient.listModifiers({})),
  createModifier: (data: ModifierPayload) =>
    onekit(
      modifierClient.createModifier({
        name: data.name,
        priceDelta: numberValue(data.priceDelta ?? data.price_delta),
        groupName: text(data.groupName ?? data.group_name),
        isRequired: boolValue(data.isRequired ?? data.is_required),
      })
    ),
  updateModifier: (id: number, data: Partial<ModifierPayload>) =>
    onekit(
      modifierClient.updateModifier({
        id: String(id),
        body: {
          name: text(data.name),
          priceDelta: numberValue(data.priceDelta ?? data.price_delta),
          groupName: text(data.groupName ?? data.group_name),
          isRequired: boolValue(data.isRequired ?? data.is_required),
        },
      })
    ),
  deleteModifier: (id: number) => onekit(modifierClient.deleteModifier({ id: String(id) })),
  getCombos: () => onekit(comboClient.listCombos({})),
  createCombo: (data: ComboPayload) =>
    onekit(
      comboClient.createCombo({
        name: data.name,
        description: text(data.description),
        price: numberValue(data.price),
        items: Array.isArray(data.items) ? JSON.stringify(data.items) : text(data.items),
        isActive: typeof data.isActive === 'boolean' ? data.isActive : true,
      })
    ),
  updateCombo: (id: number, data: Partial<ComboPayload>) =>
    onekit(
      comboClient.updateCombo({
        id: String(id),
        body: {
          name: text(data.name),
          description: text(data.description),
          price: numberValue(data.price),
          items: Array.isArray(data.items) ? JSON.stringify(data.items) : text(data.items),
          isActive: typeof data.isActive === 'boolean' ? data.isActive : true,
        },
      })
    ),
  deleteCombo: (id: number) => onekit(comboClient.deleteCombo({ id: String(id) })),

  // ── Sessions ──────────────────────────────────────────────────
  getSessions: () => onekit(sessionClient.listSessions({})),
  getCurrentSession: () => safe.get('/api/pos/session/v1/sessions/current'),
  openSession: (data: OpenSessionPayload) =>
    onekit(
      sessionClient.openSession({
        locationId: id(data.location_id),
        staffId: id(data.staff_id),
        openingFloat: data.opening_float ?? 0,
      })
    ),
  closeSession: (id: number, data: CloseSessionPayload) =>
    onekit(sessionClient.closeSession({ id: String(id), closingFloat: data.closing_float })),

  // ── Cash movements ────────────────────────────────────────────
  getCashMovements: (sessionId: number) =>
    onekit(cashClient.listCashMovements({ sessionId: String(sessionId) })),
  cashDrop: (data: CashMovementPayload) => onekit(cashClient.cashDrop(toCashMovement(data))),
  cashPickup: (data: CashMovementPayload) => onekit(cashClient.cashPickup(toCashMovement(data))),
  cashAdjust: (data: CashMovementPayload) => onekit(cashClient.cashAdjust(toCashMovement(data))),

  // ── Gift cards ────────────────────────────────────────────────
  getGiftCards: () => onekit(giftCardClient.listGiftCards({})),
  createGiftCard: (data: GiftCardPayload) =>
    onekit(
      giftCardClient.createGiftCard({
        code: data.code ?? '',
        initialAmount: data.amount,
        customerName: text(data.customerName ?? data.customer_name),
        customerEmail: text(data.customerEmail ?? data.customer_email),
        expiresAt: text(data.expiresAt ?? data.expires_at),
      })
    ),
  redeemGiftCard: (code: string, data: GiftCardRedeemPayload) =>
    onekit(giftCardClient.redeemGiftCard({ code, amount: data.amount })),
  getGiftCardBalance: (code: string) => safe.get(`/api/pos/gift-card/v1/gift-cards/${code}`),

  // ── KDS ────────────────────────────────────────────────────────
  getKDSOrders: (station?: string) =>
    onekit(kdsClient.listKDSOrders({})),
  updateKDSItemStatus: (itemId: number, status: string) =>
    safe.put(`/api/pos/kds/v1/items/${itemId}`, { status }),

  // ── Customers ─────────────────────────────────────────────────
  searchCustomers: (q: string) => onekit(customerClient.searchCustomers({ query: q })),
  createCustomer: (data: CustomerPayload) => safe.post('/api/pos/customer/v1/customers', data),
  getCustomer: (id: number) => safe.get(`/api/pos/customer/v1/customers/${id}`),

  // ── Receipts ──────────────────────────────────────────────────
  getReceipts: (params?: ReceiptFilterParams) =>
    onekit(
      receiptClient.listReceipts({
        from: text(params?.from ?? params?.date),
        to: text(params?.to ?? params?.date),
      })
    ),
  emailReceipt: (id: number, email: string) =>
    safe.post(`/api/pos/receipt/v1/receipts/${id}/email`, { email }),

  // ── Day summary / Reports ─────────────────────────────────────
  getDaySummary: (date?: string) =>
    onekit(daySummaryClient.daySummary({ date: date ?? '' })),

  // ── Manager override ──────────────────────────────────────────
  verifyManagerPin: (pin: string) => onekit(managerClient.verifyPIN({ pin })),

  // ── Loyalty ───────────────────────────────────────────────────
  getLoyaltyConfig: () => onekit(loyaltyClient.getLoyaltyConfig({})),
  updateLoyaltyConfig: (data: LoyaltyConfigPayload) =>
    onekit(
      loyaltyClient.updateLoyaltyConfig({
        isEnabled: typeof data.isEnabled === 'boolean' ? data.isEnabled : true,
        pointsPerDollar: numberValue(data.pointsPerDollar ?? data.points_per_dollar),
        redemptionRate: numberValue(data.redemptionRate ?? data.redeemRate ?? data.redeem_rate),
        minRedeemPoints: numberValue(data.minRedeemPoints ?? data.min_redeem_points),
      })
    ),
  getLoyaltyBalance: (customerId: number) =>
    onekit(loyaltyClient.getLoyaltyBalance({ customerId: String(customerId) })),
  earnLoyaltyPoints: (data: LoyaltyPayload) =>
    onekit(
      loyaltyClient.earnPoints({
        customerId: String(data.customer_id),
        points: data.points,
        orderId: id(data.order_id),
      })
    ),
  redeemLoyaltyPoints: (data: LoyaltyPayload) =>
    onekit(
      loyaltyClient.redeemPoints({
        customerId: String(data.customer_id),
        points: data.points,
      })
    ),

  // ── Stock / Inventory ─────────────────────────────────────────
  getBranchInventory: (branchId: number) =>
    onekit(stockClient.listInventory({ branchId: String(branchId) })),
  getLowStock: (branchId: number) =>
    onekit(stockClient.lowStock({ branchId: String(branchId) })),
  receiveStock: (data: StockPayload) => onekit(stockClient.receiveStock(toStock(data))),
  adjustStock: (data: StockPayload) => onekit(stockClient.adjustStock(toStock(data))),
  wasteStock: (data: StockPayload) => onekit(stockClient.wasteStock(toStock(data))),
  transferStock: (data: StockTransferPayload) =>
    onekit(
      stockClient.transferStock({
        fromBranchId: String(data.from_branch_id),
        toBranchId: String(data.to_branch_id),
        productId: String(data.product_id),
        quantity: data.quantity,
        note: data.note ?? '',
        employeeId: id(data.employee_id),
      })
    ),

  // ── Audit ─────────────────────────────────────────────────────
  getAuditLog: (params?: AuditFilterParams) =>
    onekit(
      auditClient.listAudit({
        action: text(params?.action),
        entityType: text(params?.entityType ?? params?.entity_type),
        employeeId: id(params?.employeeId ?? params?.employee_id),
        startDate: text(params?.startDate),
        endDate: text(params?.endDate),
      })
    ),

  // ── POS Users (cashiers/staff - PIN login) ────────────────────
  getPOSUsers: (locationId?: number) =>
    onekit(userClient.listPOSUsers({ locationId: id(locationId) })),
  createPOSUser: (data: POSUserPayload) =>
    onekit(
      userClient.createPOSUser({
        locationId: String(data.location_id),
        name: data.name,
        pin: data.pin,
        role: data.role ?? '',
        hourlyRate: numberValue(data.hourly_rate),
        maxDiscountPct: numberValue(data.max_discount_pct),
        canVoid: Boolean(data.can_void),
        canRefund: Boolean(data.can_refund),
        canOpenDrawer: Boolean(data.can_open_drawer),
        canApplyDiscount: Boolean(data.can_apply_discount),
      })
    ),
  updatePOSUser: (data: Partial<POSUserPayload> & { id: number }) =>
    onekit(
      userClient.updatePOSUser({
        id: String(data.id),
        name: text(data.name),
        pin: text(data.pin),
        role: text(data.role),
        isActive: typeof data.is_active === 'boolean' ? data.is_active : true,
        hourlyRate: numberValue(data.hourly_rate),
        maxDiscountPct: numberValue(data.max_discount_pct),
        canVoid: Boolean(data.can_void),
        canRefund: Boolean(data.can_refund),
        canOpenDrawer: Boolean(data.can_open_drawer),
        canApplyDiscount: Boolean(data.can_apply_discount),
      })
    ),
  deactivatePOSUser: (id: number) => onekit(userClient.deactivatePOSUser({ id: String(id) })),
  verifyPIN: (locationId: number, userId: number, pin: string) =>
    onekit(
      userClient.verifyPOSUserPIN({
        locationId: String(locationId),
        userId: String(userId),
        pin,
      })
    ),
  setPIN: (data: { userId: number; locationId: number; name: string; pin: string }) =>
    safe.post('/api/pos/user/v1/users/set-pin', {
      user_id: data.userId,
      location_id: data.locationId,
      name: data.name,
      pin: data.pin,
    }),

  // ── Member Location Access ────────────────────────────────────
  getMyLocations: () => onekit(accessClient.getMyLocations({})),
  getLocationRoster: (locationId: number) =>
    safe.get('/api/pos/access/v1/location-roster', {
      params: { location_id: locationId },
    }),
  assignMemberToLocation: (data: MemberLocationPayload) =>
    onekit(
      accessClient.assignLocation({
        userId: id(data.user_id ?? data.userId),
        locationId: String(data.locationId),
        canManagePosUsers: Boolean(data.canManagePosUsers),
        canManageTerminals: Boolean(data.canManageTerminals),
        canViewReports: Boolean(data.canViewReports),
      })
    ),
  removeMemberFromLocation: (id: number) =>
    onekit(accessClient.removeLocation({ id: String(id) })),
};
