# Shared Packages

All packages live in `rtylr/packages/` and are published as `@voxire/*` workspace packages. They are consumed by apps via pnpm workspace references.

---

## `@voxire/auth`

**Source:** `packages/auth/src/`

Provides cookie-based authentication primitives, the Zustand user store, `ProtectedRoute`, and `RoleGuard`. Every app that requires authentication depends on this package.

### Exports

```ts
// Cookie utilities
export function getCookie(name: string): string | null;
export function setCookie(name: string, value: string, options?: SetCookieOptions): void;
export function removeCookie(name: string, options?: { path?: string; domain?: string }): void;
export function getDefaultCookieDomain(): string | undefined;

// User store (Zustand)
export const useUserStore: StoreApi<UserStore>;

// Guards
export function ProtectedRoute({ appId, children }: ProtectedRouteProps): JSX.Element | null;
export function RoleGuard({ allowed, role, fallback, children }: RoleGuardProps): JSX.Element;

// Types
export type { User, UserState, UserActions, UserRole, AppId };
```

### Usage

```tsx
// Protect an entire app
import { ProtectedRoute } from '@voxire/auth';

<ProtectedRoute appId="erp">
  <App />
</ProtectedRoute>;

// Read auth state anywhere
import { useUserStore } from '@voxire/auth';

const role = useUserStore((state) => state.role);
const user = useUserStore((state) => state.user);

// Guard specific UI
import { RoleGuard } from '@voxire/auth';

<RoleGuard allowed={['owner', 'admin']} role={role}>
  <DeleteButton />
</RoleGuard>;

// Set cookie after login
import { setCookie, getDefaultCookieDomain } from '@voxire/auth';

setCookie('token', jwt, {
  domain: getDefaultCookieDomain(),
  days: 30,
});
```

---

## `@voxire/api`

**Source:** `packages/api/src/client.ts`

Pre-configured Axios instance that:

1. Reads `VITE_API_URL` (falls back to `https://api.voxire.dev/`)
2. Automatically injects `Authorization: Bearer <token>` from the `token` cookie on every request
3. On 401: clears the token cookie and redirects to `auth.rtylr.com?redirect=<appId>`

### Exports

```ts
export const api: AxiosInstance; // default client, appId="hub"
export function createApiClient(appId: string): AxiosInstance;
```

### Usage

```ts
// Use the default client (works in Hub)
import { api } from '@voxire/api';

const { data } = await api.get('/hub/users');

// Create an app-specific client (for correct redirect-back behavior)
import { createApiClient } from '@voxire/api';

const api = createApiClient('erp'); // 401 → auth.rtylr.com?redirect=erp
const { data } = await api.get('/erp/products', { params: { ownerId: 42 } });
await api.post('/erp/products', { name: 'Flour', costPrice: 1.5, salePrice: 3.0 });
```

### Environment Variable

| Variable       | Default                   | Description          |
| -------------- | ------------------------- | -------------------- |
| `VITE_API_URL` | `https://api.voxire.dev/` | Backend API base URL |

---

## `@voxire/types`

**Source:** `packages/types/src/`

TypeScript types for every domain in the platform. No runtime code - types only.

### Exports by file

**`user.ts`**

```ts
type UserRole = 'owner' | 'admin' | 'operator' | 'staff' | 'viewer';
type AppId =
  | 'hub'
  | 'auth'
  | 'pos'
  | 'erp'
  | 'hr'
  | 'crm'
  | 'finance'
  | 'flow'
  | 'menu'
  | 'landing';

interface User {
  id?: string;
  name?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  role?: UserRole;
  businessName?: string;
  apps?: AppId[]; // which apps this user has access to
  addons?: Record<string, boolean>;
}

interface AuthResponse {
  token: string;
  userId: string;
  user: User;
  visits?: any[];
}
```

**`pos.ts`**

```ts
type OrderStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';
type PaymentMethod = 'cash' | 'card' | 'online';

interface OrderItem {
  id;
  name;
  price;
  quantity;
  notes?;
}
interface Order {
  id;
  tableNumber?;
  items;
  total;
  status;
  paymentMethod?;
  createdAt;
  updatedAt?;
}
interface Table {
  id;
  number;
  capacity;
  status: 'available' | 'occupied' | 'reserved';
  currentOrderId?;
}
```

**`erp.ts`**

```ts
interface Product {
  id;
  name;
  sku?;
  category?;
  costPrice;
  salePrice;
  stock;
  lowStockThreshold?;
  unit?;
  supplierId?;
}
interface Supplier {
  id;
  name;
  email?;
  phone?;
  address?;
  products?;
}
type PurchaseOrderStatus = 'draft' | 'sent' | 'received' | 'cancelled';
interface PurchaseOrder {
  id;
  supplierId;
  items;
  total;
  status;
  expectedDelivery?;
  createdAt;
}
```

**`hr.ts`** - Staff, Shift, PayrollRecord, TimeOffRequest types

**`crm.ts`** - Customer, LoyaltyTransaction, LoyaltyConfig, Segment, Campaign types

**`finance.ts`** - Invoice, InvoiceItem, Payment, Expense, ExpenseCategory, TaxConfig types

**`flow.ts`** - Workspace, Project, Task, TaskComment, Label types

**`business.ts`** - Business-level shared types

### Usage

```ts
import type { User, UserRole, AppId } from '@voxire/types';
import type { Product, Supplier } from '@voxire/types';
import type { Order, Table } from '@voxire/types';
```

---

## `@voxire/config`

**Source:** `packages/config/src/`

Central source of truth for app URLs and configuration. Auto-switches between localhost and production based on `window.location.hostname`.

### Exports

```ts
// App base URLs keyed by AppId
export const APP_URLS: Record<AppId, string>;

// App metadata and role access keyed by AppId
export const APP_CONFIG: Record<
  AppId,
  {
    name: string;
    devPort: number;
    allowedRoles: UserRole[];
  }
>;

// Re-exported types
export type { AppId };
```

### Values

```ts
// Local dev (when hostname === 'localhost')
APP_URLS = {
  hub: 'http://localhost:5175',
  auth: 'http://localhost:5176',
  menu: 'http://localhost:5177',
  pos: 'http://localhost:5178',
  erp: 'http://localhost:5179',
  hr: 'http://localhost:5180',
  crm: 'http://localhost:5181',
  finance: 'http://localhost:5182',
  flow: 'http://localhost:5183',
  landing: 'http://localhost:4200',
};

// Production
APP_URLS = {
  hub: 'https://hub.rtylr.com',
  auth: 'https://auth.rtylr.com',
  // ...etc
};

APP_CONFIG = {
  hub: { name: 'Hub', devPort: 5175, allowedRoles: ['owner', 'admin', 'operator'] },
  auth: {
    name: 'Auth',
    devPort: 5176,
    allowedRoles: ['owner', 'admin', 'operator', 'staff', 'viewer'],
  },
  menu: { name: 'Menu', devPort: 5177, allowedRoles: [] },
  pos: { name: 'POS', devPort: 5178, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  erp: { name: 'ERP', devPort: 5179, allowedRoles: ['owner', 'admin', 'operator'] },
  hr: { name: 'HR', devPort: 5180, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  crm: { name: 'CRM', devPort: 5181, allowedRoles: ['owner', 'admin', 'operator'] },
  finance: { name: 'Finance', devPort: 5182, allowedRoles: ['owner', 'admin'] },
  flow: { name: 'Flow', devPort: 5183, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  landing: { name: 'Landing', devPort: 4200, allowedRoles: [] },
};
```

### Usage

```ts
import { APP_URLS, APP_CONFIG } from '@voxire/config';

// Navigate to another app
window.location.href = APP_URLS.hub;

// Check if a role can access an app
const canAccess = APP_CONFIG.finance.allowedRoles.includes(currentRole);

// Get the dev port for an app
const port = APP_CONFIG.erp.devPort; // 5179
```

---

## `@voxire/ui`

**Source:** `packages/ui/src/`

Shared low-level UI components used across apps (buttons, inputs, modals, cards, etc.). Import from the package directly:

```ts
import { Button, Modal, Card } from '@voxire/ui';
```

---

## `@voxire/components`

**Source:** `packages/components/src/`

Higher-level shared components including `AnimatedRoutes`, `FullScreenLoader`, and other composite UI pieces.

```ts
import { AnimatedRoutes, FullScreenLoader } from '@voxire/components';
```

---

## `@voxire/menu`

**Source:** `packages/menu/src/`

Shared menu display components used by both the customer-facing Menu app and Hub's menu editor. The Menu app imports styles via:

```ts
import '@voxire/menu/styles';
```

---

## `@voxire/assets`

**Source:** `packages/assets/src/`

Shared static assets (images, icons, SVGs). Referenced via the `assets` export in `@voxire/config`.

---

## `@voxire/utils`

**Source:** `packages/utils/src/`

Shared utility functions (formatting, validation, date helpers, etc.).

---

## `@voxire/tailwind-config`

**Source:** `packages/tailwind-config/`

Shared Tailwind CSS configuration preset. Apps extend it in their `tailwind.config.*` file.
