import { usePOSPortal } from '../../shared/components/PortalContext';
import React from 'react';
import {
  buildCaretRouteContext,
  CaretButton,
  CaretPanel,
  useCaretSettingsStore,
  useCaretStore,
} from '@voxire/caret';
import { Moon, Power, Sun } from 'lucide-react';
import { useLocation } from 'react-router-dom';
import { useThemeStore } from '../../shared/stores/themeStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useUserStore, businessDisplayName } from '../../shared/stores/userStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { useCartStore } from '../../shared/stores/cartStore';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { useCashMovementQueueStore } from '../../shared/stores/cashMovementQueueStore';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';
import { resolveApiBaseUrl } from '../../shared/services/server/serverConfig';

interface POSLayoutProps {
  children: React.ReactNode;
}

const POS_ROUTE_LABELS: Record<string, string> = {
  '/': 'Terminal',
  '/menu': 'Menu',
  '/modifiers': 'Modifiers',
  '/combos': 'Combos',
  '/split-bill': 'Split Bill',
  '/payment': 'Payment',
  '/refund': 'Refund',
  '/discount': 'Discount',
  '/gift-cards': 'Gift Cards',
  '/kds': 'Kitchen Display',
  '/customer-lookup': 'Customer Lookup',
  '/receipts': 'Receipts',
  '/receipt': 'Receipt',
  '/day-summary': 'Day Summary',
  '/override': 'Manager Override',
  '/products/quick': 'Quick Product Edit',
  '/orders': 'Orders',
  '/tables': 'Tables',
  '/close-session': 'Close Session',
  '/offline': 'Offline Queue',
  '/settings': 'Settings',
  '/held-orders': 'Held Orders',
  '/cash-drop': 'Cash Drop',
  '/inventory': 'Inventory',
  '/staff': 'Staff',
};

const POS_ROUTE_MODULES: Record<string, string> = {
  '/': 'terminal',
  '/menu': 'menu',
  '/modifiers': 'menu',
  '/combos': 'menu',
  '/split-bill': 'orders',
  '/payment': 'checkout',
  '/refund': 'orders',
  '/discount': 'orders',
  '/gift-cards': 'customers',
  '/kds': 'kitchen',
  '/customer-lookup': 'customers',
  '/receipts': 'receipts',
  '/receipt': 'receipts',
  '/day-summary': 'reporting',
  '/override': 'security',
  '/products/quick': 'inventory',
  '/orders': 'orders',
  '/tables': 'tables',
  '/close-session': 'cash',
  '/settings': 'settings',
  '/held-orders': 'orders',
  '/cash-drop': 'cash',
  '/inventory': 'inventory',
  '/staff': 'staff',
};

function selectedOrderId(pathname: string): string | null {
  const match = pathname.match(/^\/(?:payment|refund|discount|split-bill)\/([^/?#]+)/);
  return match?.[1] ? decodeURIComponent(match[1]) : null;
}

const POSLayout: React.FC<POSLayoutProps> = ({ children }) => {
  const embedded = usePOSPortal();
  const location = useLocation();
  const { theme, toggleTheme } = useThemeStore();
  const endSession = useEmployeeSessionStore((s) => s.endSession);
  const employee = useEmployeeSessionStore((s) => s.employee);
  const employeeSessionActive = useEmployeeSessionStore((s) => s.isActive);
  const businessProfile = useUserStore((s) => s.businessProfile);
  const userIdValue = useUserStore((s) => s.userId);
  const configuredBusinessName = useSettingsStore((s) => s.business.name.trim());
  const locationId = useTerminalContextStore((s) => s.locationId);
  const terminalId = useTerminalContextStore((s) => s.terminalId);
  const cartItems = useCartStore((s) => s.items);
  const cartOrderType = useCartStore((s) => s.orderType);
  const cartTableId = useCartStore((s) => s.tableId);
  const cartCustomerId = useCartStore((s) => s.customerId);
  const cartCustomerName = useCartStore((s) => s.customerName);
  const cartHeld = useCartStore((s) => s.isHeld);
  const cartTotal = useCartStore((s) => s.total());
  const cartAmountDue = useCartStore((s) => s.amountDue());
  const offlineQueue = useOfflineQueueStore((s) => s.queue);
  const online = useOfflineQueueStore((s) => s.online);
  const cashMovementQueue = useCashMovementQueueStore((s) => s.queue);
  const businessName = configuredBusinessName || businessDisplayName(businessProfile);
  const caretOpen = useCaretStore((s) => s.isOpen);
  const caretWidth = useCaretSettingsStore((s) => s.panelWidth);
  const userId = Number(userIdValue ?? 0);
  const caretUserId = Number.isFinite(userId) && userId > 0 ? userId : 0;
  const orderId = selectedOrderId(location.pathname);
  // Single pass over each queue instead of three filters per render.
  const { pendingOfflineOrders, failedOfflineOrders } = React.useMemo(
    () => ({
      pendingOfflineOrders: offlineQueue.filter((item) => item.status === 'pending').length,
      failedOfflineOrders: offlineQueue.filter((item) => item.status === 'failed').length,
    }),
    [offlineQueue]
  );
  const pendingCashMovements = React.useMemo(
    () => cashMovementQueue.filter((item) => item.status === 'pending').length,
    [cashMovementQueue]
  );
  const caretApiBaseUrl = resolveApiBaseUrl() ?? '';
  const caretPageContext = React.useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'pos',
        appName: 'POS',
        pathname: location.pathname,
        routeLabels: POS_ROUTE_LABELS,
        routeModules: POS_ROUTE_MODULES,
        businessName: businessName ?? undefined,
        role: employee?.role,
        selected: {
          ...(orderId ? { order_id: orderId } : {}),
          ...(!orderId && location.pathname.startsWith('/tables') && cartTableId
            ? { table_id: cartTableId }
            : {}),
        },
        filters: {
          ...(locationId ? { location_id: locationId } : {}),
          ...(terminalId ? { terminal_id: terminalId } : {}),
          ...(employee?.id ? { employee_id: employee.id } : {}),
          ...(employeeSessionActive ? { employee_session: 'active' } : {}),
        },
        cacheContext: {
          cart_items: cartItems.length,
          cart_total: Number(cartTotal.toFixed(2)),
          cart_amount_due: Number(cartAmountDue.toFixed(2)),
          order_type: cartOrderType,
          cart_held: cartHeld,
          ...(cartCustomerId ? { customer_id: cartCustomerId } : {}),
          ...(cartCustomerName ? { customer_name: cartCustomerName } : {}),
          pending_offline_orders: pendingOfflineOrders,
          failed_offline_orders: failedOfflineOrders,
          pending_cash_movements: pendingCashMovements,
          online,
        },
        dataFreshness: online ? 'online' : 'offline',
        hints: [
          'current shift, open orders, tables, receipts, stock, and customers',
          pendingOfflineOrders > 0 ? `${pendingOfflineOrders} offline order sync pending` : '',
          failedOfflineOrders > 0 ? `${failedOfflineOrders} offline order sync failed` : '',
          pendingCashMovements > 0 ? `${pendingCashMovements} cash movement sync pending` : '',
        ].filter(Boolean),
      }),
    [
      businessName,
      cartAmountDue,
      cartCustomerId,
      cartCustomerName,
      cartHeld,
      cartItems.length,
      cartOrderType,
      cartTableId,
      cartTotal,
      employee?.id,
      employee?.role,
      employeeSessionActive,
      failedOfflineOrders,
      location.pathname,
      locationId,
      online,
      orderId,
      pendingCashMovements,
      pendingOfflineOrders,
      terminalId,
    ]
  );

  if (embedded) return <div className="flex h-full min-h-0 flex-col bg-pos-bg text-pos-text">{children}</div>;
  return (
    <div className="flex flex-col h-screen bg-pos-bg text-pos-text" data-theme={theme}>
      <header className="flex items-center justify-between px-5 py-3 bg-pos-card border-b border-pos-border shrink-0">
        <div className="flex items-center gap-2.5">
          <RtylrWordmark className="text-xl" />
        </div>
        <div className="flex items-center gap-1">
          {businessName && <span className="text-xs text-pos-muted mr-2">{businessName}</span>}
          <CaretButton />
          <button
            className="w-8 h-8 flex items-center justify-center rounded-lg text-pos-muted hover:bg-white/8 transition-colors"
            onClick={toggleTheme}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
              <Sun className="h-4 w-4" aria-hidden="true" />
            ) : (
              <Moon className="h-4 w-4" aria-hidden="true" />
            )}
          </button>
          <button
            className="w-8 h-8 flex items-center justify-center rounded-lg text-pos-muted hover:bg-white/8 transition-colors"
            onClick={endSession}
            aria-label="End session"
            title="End employee session"
          >
            <Power className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </header>
      <main className="flex flex-1 min-h-0 overflow-hidden">
        <div className="flex-1 min-w-0 overflow-hidden">{children}</div>
        <div
          style={{
            width: caretOpen ? `${caretWidth}px` : '0px',
            flexShrink: 0,
            overflow: 'hidden',
            transition: caretOpen ? 'none' : 'width 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
          }}
        >
          <CaretPanel
            variant="inline"
            appId="pos"
            userId={caretUserId}
            apiBaseUrl={caretApiBaseUrl}
            systemContext="User is in the POS app (terminal, orders, tables, receipts, inventory, customers, and staff workflows)."
            pageContext={caretPageContext}
          />
        </div>
      </main>
    </div>
  );
};

export default POSLayout;
