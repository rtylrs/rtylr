import { safe } from '../lib/api';

// -- Types --

export interface StoreProduct {
  id: number;
  name: string;
  sku: string;
  category: string;
  categoryId?: number;
  price: number;
  compareAtPrice?: number;
  stock: number;
  status: 'active' | 'draft' | 'archived';
  imageUrl?: string;
  description?: string;
  tags?: string[];
  weight?: number;
  weightUnit?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProductInput {
  name: string;
  sku: string;
  categoryId?: number;
  price: number;
  compareAtPrice?: number;
  stock: number;
  status?: StoreProduct['status'];
  description?: string;
  tags?: string[];
  weight?: number;
  weightUnit?: string;
}

export interface UpdateProductInput extends Partial<CreateProductInput> {
  id: number;
}

export interface StoreOrder {
  id: number;
  orderNumber: string;
  customerName: string;
  customerEmail: string;
  status: 'pending' | 'confirmed' | 'shipped' | 'delivered' | 'cancelled' | 'refunded';
  total: number;
  subtotal: number;
  shippingCost: number;
  tax: number;
  itemCount: number;
  paymentMethod: string;
  shippingMethod?: string;
  trackingNumber?: string;
  createdAt: string;
  updatedAt: string;
}

export interface StoreCategory {
  id: number;
  name: string;
  slug: string;
  description?: string;
  parentId?: number | null;
  productCount: number;
  imageUrl?: string;
  sortOrder: number;
}

export interface CreateCategoryInput {
  name: string;
  slug?: string;
  description?: string;
  parentId?: number | null;
  sortOrder?: number;
}

export interface ShippingZone {
  id: number;
  name: string;
  regions: string[];
  methods: ShippingMethod[];
}

export interface ShippingMethod {
  id: number;
  name: string;
  type: 'flat_rate' | 'free' | 'weight_based' | 'price_based';
  cost: number;
  minOrderAmount?: number;
  estimatedDays: string;
  enabled: boolean;
}

export interface CreateShippingZoneInput {
  name: string;
  regions: string[];
}

export interface PaymentMethod {
  id: number;
  name: string;
  provider: string;
  type: 'credit_card' | 'paypal' | 'apple_pay' | 'google_pay' | 'cod' | 'bank_transfer';
  enabled: boolean;
  testMode: boolean;
  configuredAt?: string;
}

export interface StorefrontConfig {
  id: number;
  themeName: string;
  logoUrl?: string;
  faviconUrl?: string;
  primaryColor: string;
  secondaryColor: string;
  customDomain?: string;
  metaTitle?: string;
  metaDescription?: string;
  socialLinks?: Record<string, string>;
  announcementBar?: string;
  footerText?: string;
}

export interface StoreSettings {
  id: number;
  storeName: string;
  currency: string;
  taxRate: number;
  taxInclusive: boolean;
  orderPrefix: string;
  lowStockThreshold: number;
  emailNotifications: boolean;
  orderConfirmationEmail: boolean;
  shippingNotificationEmail: boolean;
  timezone: string;
}

export interface DashboardStats {
  totalProducts: number;
  ordersToday: number;
  revenueToday: number;
  activeVisitors: number;
}

// -- API --

export const storeApi = {
  // Dashboard
  getDashboardStats: () => safe.get<DashboardStats>('/store/dashboard'),

  // Products
  listProducts: (params?: {
    search?: string;
    categoryId?: number;
    status?: string;
    limit?: number;
    offset?: number;
  }) => safe.get<StoreProduct[]>('/store/products', { params }),
  getProduct: (id: number) => safe.get<StoreProduct>(`/store/products/${id}`),
  createProduct: (data: CreateProductInput) => safe.post<StoreProduct>('/store/products', data),
  updateProduct: (data: UpdateProductInput) => safe.post<StoreProduct>('/store/products', data),
  deleteProduct: (id: number) =>
    safe.post<{ success: boolean }>('/store/products', { id, _action: 'delete' }),

  // Orders
  listOrders: (params?: { status?: string; search?: string; limit?: number; offset?: number }) =>
    safe.get<StoreOrder[]>('/store/orders', { params }),
  getOrder: (id: number) => safe.get<StoreOrder>(`/store/orders/${id}`),
  updateOrderStatus: (id: number, status: StoreOrder['status']) =>
    safe.post<StoreOrder>('/store/orders/status', { id, status }),

  // Categories
  listCategories: () => safe.get<StoreCategory[]>('/store/categories'),
  createCategory: (data: CreateCategoryInput) =>
    safe.post<StoreCategory>('/store/categories', data),
  updateCategory: (data: Partial<CreateCategoryInput> & { id: number }) =>
    safe.post<StoreCategory>('/store/categories', data),
  deleteCategory: (id: number) =>
    safe.post<{ success: boolean }>('/store/categories', { id, _action: 'delete' }),

  // Shipping
  listShippingZones: () => safe.get<ShippingZone[]>('/store/shipping/zones'),
  createShippingZone: (data: CreateShippingZoneInput) =>
    safe.post<ShippingZone>('/store/shipping/zones', data),
  updateShippingZone: (data: Partial<CreateShippingZoneInput> & { id: number }) =>
    safe.post<ShippingZone>('/store/shipping/zones', data),
  listShippingMethods: () => safe.get<ShippingMethod[]>('/store/shipping/methods'),

  // Payments
  listPaymentMethods: () => safe.get<PaymentMethod[]>('/store/payments/methods'),
  updatePaymentMethod: (data: Partial<PaymentMethod> & { id: number }) =>
    safe.post<PaymentMethod>('/store/payments/methods', data),

  // Storefront
  getStorefrontConfig: () => safe.get<StorefrontConfig>('/store/storefront'),
  updateStorefrontConfig: (data: Partial<StorefrontConfig>) =>
    safe.post<StorefrontConfig>('/store/storefront', data),

  // Settings
  getSettings: () => safe.get<StoreSettings>('/store/settings'),
  updateSettings: (data: Partial<StoreSettings>) =>
    safe.post<StoreSettings>('/store/settings', data),
};
