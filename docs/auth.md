# Authentication

rtylr uses cookie-based SSO across all apps. A single JWT stored in a cookie on the `.rtylr.com` domain is shared by every subdomain. No per-app login sessions exist — authenticate once, access everything your role permits.

---

## The Token Cookie

| Property | Value |
|----------|-------|
| Name | `token` |
| Domain | `.rtylr.com` (or `undefined` on localhost) |
| Path | `/` |
| SameSite | `lax` |
| Secure | `true` in production |
| Encoding | URL-encoded JWT string |

The cookie is set by the Auth app after a successful login (`POST /login`). It is read by every other app to authenticate API calls.

---

## Full Auth Flow

```
1. User visits any app (e.g., pos.rtylr.com)

2. ProtectedRoute checks for `token` cookie
   → cookie absent: redirect to auth.rtylr.com?redirect=pos
   → cookie present: render children

3. auth.rtylr.com shows the login form

4. User submits credentials
   → POST https://api.voxire.dev/login
   → Response: { token, userId, user, visits }

5. Auth app calls setCookie("token", jwt, { domain: ".rtylr.com", days: N })

6. Auth app reads ?redirect query param and navigates to:
   → e.g., https://pos.rtylr.com  (or http://localhost:5178 in dev)

7. pos.rtylr.com loads, ProtectedRoute finds the cookie, renders the app

8. Every API request sends:
   Authorization: Bearer <token>

9. On 401 response:
   → @voxire/api client clears the cookie
   → Redirects to auth.rtylr.com?redirect=<current-appId>
```

---

## Backend Token Validation

The Go backend validates the `Authorization: Bearer <token>` header on all non-anonymous routes. The auth middleware is applied globally:

```go
http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver)
```

On successful validation the middleware injects `AuthUser` into the request context:

```go
type AuthUser struct {
    UserID         int64
    Platform       string
    PlatformUserID int64
    PlanID         int64
    Token          string
    IP             string
}

// Read in a handler:
authUser, ok := authPkg.GetAuthUser(ctx)
```

Anonymous routes bypass the middleware. A module declares them via:

```go
func (m *Module) GetAnonymousRoutes() []string {
    return []string{"/login", "/register"}
}
```

The `user` module marks `/login` and `/register` as anonymous. All other routes require a valid token.

---

## User Roles

Roles are stored in `user_app_permissions.role` and embedded in the JWT claims.

| Role | Description |
|------|-------------|
| `owner` | Full access to everything. The business account creator. |
| `admin` | Full access to all apps except some billing/account-level operations. |
| `operator` | Access to operational apps: Hub, POS, ERP, HR, CRM, Flow. |
| `staff` | Access to POS, HR (own data), Flow. |
| `viewer` | Read-only access to Auth. |

App-level access matrix (from `APP_CONFIG`):

| App | owner | admin | operator | staff | viewer |
|-----|-------|-------|----------|-------|--------|
| Hub | ✓ | ✓ | ✓ | | |
| Auth | ✓ | ✓ | ✓ | ✓ | ✓ |
| Menu | (public) | (public) | (public) | (public) | (public) |
| POS | ✓ | ✓ | ✓ | ✓ | |
| ERP | ✓ | ✓ | ✓ | | |
| HR | ✓ | ✓ | ✓ | ✓ | |
| CRM | ✓ | ✓ | ✓ | | |
| Finance | ✓ | ✓ | | | |
| Flow | ✓ | ✓ | ✓ | ✓ | |
| Landing | (public) | (public) | (public) | (public) | (public) |

---

## `ProtectedRoute`

Wraps any app or subtree that requires authentication. If no `token` cookie is found, it immediately redirects to the Auth app.

```tsx
import { ProtectedRoute } from '@voxire/auth';

// In your app's root (e.g., main.tsx or App.tsx)
<ProtectedRoute appId="pos">
  <App />
</ProtectedRoute>
```

**Props:**

| Prop | Type | Description |
|------|------|-------------|
| `appId` | `AppId` | Used to build the redirect URL: `auth.rtylr.com?redirect=<appId>` |
| `children` | `ReactNode` | Rendered only when the token cookie is present |

**Behavior:**
- Reads the `token` cookie synchronously on render
- If absent: `window.location.href = ${AUTH_URL}?redirect=${appId}` and returns `null`
- If present: renders `children` immediately (no network call — the API will 401 if the token is expired)

---

## `RoleGuard`

Conditionally renders UI based on the current user's role. Does not redirect — it simply shows or hides content.

```tsx
import { RoleGuard } from '@voxire/auth';
import { useUserStore } from '@voxire/auth';

function AdminPanel() {
  const role = useUserStore((state) => state.role);

  return (
    <RoleGuard allowed={['owner', 'admin']} role={role} fallback={<AccessDenied />}>
      <SensitiveSettings />
    </RoleGuard>
  );
}
```

**Props:**

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `allowed` | `UserRole[]` | — | Roles that may see `children` |
| `role` | `UserRole \| null` | — | The current user's role from the store |
| `fallback` | `ReactNode` | `null` | Rendered when role is not in `allowed` |
| `children` | `ReactNode` | — | Rendered when role is allowed |

---

## User Store (`useUserStore`)

Zustand store from `@voxire/auth`. Persisted to `localStorage` under the key `rtylr:voxire:user`. The `token` is intentionally excluded from persistence — it lives only in the cookie.

```ts
import { useUserStore } from '@voxire/auth';

const { user, role, isAuthenticated, login, logout, setToken } = useUserStore();
```

**State:**

| Field | Type | Description |
|-------|------|-------------|
| `user` | `User \| null` | Full user object |
| `token` | `string \| null` | JWT (in-memory only, not persisted) |
| `isAuthenticated` | `boolean` | Whether the user is logged in |
| `role` | `UserRole \| null` | Current role |
| `userId` | `string \| null` | User ID |
| `addons` | `Record<string, boolean>` | Enabled add-ons |

**Actions:**

| Action | Signature | Description |
|--------|-----------|-------------|
| `login` | `(user: User) => void` | Sets user, role, userId, addons |
| `logout` | `() => void` | Clears state and removes `token` cookie |
| `setToken` | `(token: string) => void` | Stores token in memory |
| `setUser` | `(user: User) => void` | Updates user and derived fields |
| `setRole` | `(role: UserRole) => void` | Updates role |

---

## Cookie Helpers

From `@voxire/auth`:

```ts
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';

// Read a cookie
const token = getCookie('token'); // string | null

// Write a cookie (on .rtylr.com in production, undefined on localhost)
setCookie('token', jwtString, {
  domain: getDefaultCookieDomain(), // '.rtylr.com' or undefined
  days: 30,
  secure: true,   // auto-set in production
  sameSite: 'lax',
});

// Remove a cookie
removeCookie('token', { domain: getDefaultCookieDomain() });
```

`getDefaultCookieDomain()` returns:
- `'.rtylr.com'` when on `rtylr.com` or any subdomain
- `undefined` on localhost or IP addresses (so the cookie is domain-less)
