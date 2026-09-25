import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Navigate,
  Outlet,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import { FullScreenLoader } from '@voxire/components';
import { userClient } from './shared/lib/clients';
import { z } from 'zod';
import DefaultLayout from './layouts/DefaultLayout';

const Dashboard = React.lazy(() => import('./features/dashboard/Dashboard'));
const Invoices = React.lazy(() => import('./features/invoices/Invoices'));
const InvoiceDetail = React.lazy(() => import('./features/invoices/InvoiceDetail'));
const NewInvoice = React.lazy(() => import('./features/invoices/NewInvoice'));
const RecurringInvoices = React.lazy(
  () => import('./features/recurring-invoices/RecurringInvoices')
);
const Quotes = React.lazy(() => import('./features/quotes/Quotes'));
const Expenses = React.lazy(() => import('./features/expenses/Expenses'));
const ExpenseDetail = React.lazy(() => import('./features/expenses/ExpenseDetail'));
const NewExpense = React.lazy(() => import('./features/expenses/NewExpense'));
const Categories = React.lazy(() => import('./features/categories/Categories'));
const Bills = React.lazy(() => import('./features/bills/Bills'));
const BankFeeds = React.lazy(() => import('./features/bank-feeds/BankFeeds'));
const Reconcile = React.lazy(() => import('./features/reconcile/Reconcile'));
const ProfitLoss = React.lazy(() => import('./features/profit-loss/ProfitLoss'));
const BalanceSheet = React.lazy(() => import('./features/balance-sheet/BalanceSheet'));
const CashFlow = React.lazy(() => import('./features/cash-flow/CashFlow'));
const Tax = React.lazy(() => import('./features/tax/Tax'));
const TaxReport = React.lazy(() => import('./features/tax/TaxReport'));
const Journal = React.lazy(() => import('./features/journal/Journal'));
const ChartOfAccounts = React.lazy(() => import('./features/chart-of-accounts/ChartOfAccounts'));
const Settings = React.lazy(() => import('./features/settings/Settings'));

function wrap(component: React.ReactNode) {
  return () => <Suspense fallback={<FullScreenLoader />}>{component}</Suspense>;
}

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="finance"
        fetchUser={async (): Promise<Record<string, unknown>> =>
          (await userClient.getMe({})) as unknown as Record<string, unknown>
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_FINANCE}>
          <Outlet />
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  ),
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'layout',
  component: DefaultLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: wrap(<Dashboard />),
});

const invoicesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/invoices',
  component: wrap(<Invoices />),
});

const newInvoiceRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/invoices/new',
  validateSearch: z.object({
    clone: z.string().min(1).optional().catch(undefined),
    client: z.string().min(1).optional().catch(undefined),
  }),
  component: wrap(<NewInvoice />),
});

const recurringInvoicesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/invoices/recurring',
  component: wrap(<RecurringInvoices />),
});

const invoiceDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/invoices/$id',
  component: wrap(<InvoiceDetail />),
});

const quotesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/quotes',
  component: wrap(<Quotes />),
});

const expensesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses',
  component: wrap(<Expenses />),
});

const newExpenseRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/new',
  component: wrap(<NewExpense />),
});

const expenseDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/expenses/$id',
  component: wrap(<ExpenseDetail />),
});

const categoriesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/categories',
  component: wrap(<Categories />),
});

const billsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/bills',
  component: wrap(<Bills />),
});

const bankRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/bank',
  component: wrap(<BankFeeds />),
});

const reconcileRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/reconcile',
  component: wrap(<Reconcile />),
});

const profitLossRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/pl',
  component: wrap(<ProfitLoss />),
});

const balanceSheetRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/reports/balance-sheet',
  component: wrap(<BalanceSheet />),
});

const cashFlowRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/reports/cash-flow',
  component: wrap(<CashFlow />),
});

const taxReportRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/reports/tax',
  component: wrap(<TaxReport />),
});

const journalRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/journal',
  component: wrap(<Journal />),
});

const accountsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/accounts',
  component: wrap(<ChartOfAccounts />),
});

const taxRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/tax',
  component: wrap(<Tax />),
});

const settingsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/settings',
  component: wrap(<Settings />),
});

const catchAllRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '$',
  component: () => <Navigate to="/" replace />,
});

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([
    indexRoute,
    invoicesRoute,
    newInvoiceRoute,
    recurringInvoicesRoute,
    invoiceDetailRoute,
    quotesRoute,
    expensesRoute,
    newExpenseRoute,
    expenseDetailRoute,
    categoriesRoute,
    billsRoute,
    bankRoute,
    reconcileRoute,
    profitLossRoute,
    balanceSheetRoute,
    cashFlowRoute,
    taxReportRoute,
    journalRoute,
    accountsRoute,
    taxRoute,
    settingsRoute,
    catchAllRoute,
  ]),
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
