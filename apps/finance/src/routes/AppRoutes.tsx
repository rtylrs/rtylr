import React, { Suspense } from 'react';
import { FullScreenLoader } from '@voxire/components';
import DefaultLayout from '../layouts/DefaultLayout';

// All route-level components are lazy-loaded per CLAUDE.md mandate. Layout
// components stay eager (always needed).
const Dashboard = React.lazy(() => import('../features/dashboard/Dashboard'));
const Invoices = React.lazy(() => import('../features/invoices/Invoices'));
const InvoiceDetail = React.lazy(() => import('../features/invoices/InvoiceDetail'));
const NewInvoice = React.lazy(() => import('../features/invoices/NewInvoice'));
const RecurringInvoices = React.lazy(
  () => import('../features/recurring-invoices/RecurringInvoices')
);
const Quotes = React.lazy(() => import('../features/quotes/Quotes'));
const Expenses = React.lazy(() => import('../features/expenses/Expenses'));
const ExpenseDetail = React.lazy(() => import('../features/expenses/ExpenseDetail'));
const NewExpense = React.lazy(() => import('../features/expenses/NewExpense'));
const Categories = React.lazy(() => import('../features/categories/Categories'));
const Bills = React.lazy(() => import('../features/bills/Bills'));
const BankFeeds = React.lazy(() => import('../features/bank-feeds/BankFeeds'));
const Reconcile = React.lazy(() => import('../features/reconcile/Reconcile'));
const ProfitLoss = React.lazy(() => import('../features/profit-loss/ProfitLoss'));
const BalanceSheet = React.lazy(() => import('../features/balance-sheet/BalanceSheet'));
const CashFlow = React.lazy(() => import('../features/cash-flow/CashFlow'));
const Tax = React.lazy(() => import('../features/tax/Tax'));
const TaxReport = React.lazy(() => import('../features/tax/TaxReport'));
const Journal = React.lazy(() => import('../features/journal/Journal'));
const ChartOfAccounts = React.lazy(() => import('../features/chart-of-accounts/ChartOfAccounts'));
const Settings = React.lazy(() => import('../features/settings/Settings'));

interface RouteChild {
  index?: boolean;
  path?: string;
  element: React.ReactNode;
}

interface AppRoute {
  path: string;
  element: React.ReactNode;
  children?: RouteChild[];
}

function lazy(element: React.ReactNode): React.ReactNode {
  return <Suspense fallback={<FullScreenLoader />}>{element}</Suspense>;
}

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true, element: lazy(<Dashboard />) },

      // Invoices
      { path: 'invoices', element: lazy(<Invoices />) },
      { path: 'invoices/new', element: lazy(<NewInvoice />) },
      { path: 'invoices/recurring', element: lazy(<RecurringInvoices />) },
      { path: 'invoices/:id', element: lazy(<InvoiceDetail />) },
      { path: 'quotes', element: lazy(<Quotes />) },

      // Expenses
      { path: 'expenses', element: lazy(<Expenses />) },
      { path: 'expenses/new', element: lazy(<NewExpense />) },
      { path: 'expenses/:id', element: lazy(<ExpenseDetail />) },
      { path: 'categories', element: lazy(<Categories />) },

      // Bills & bank
      { path: 'bills', element: lazy(<Bills />) },
      { path: 'bank', element: lazy(<BankFeeds />) },
      { path: 'reconcile', element: lazy(<Reconcile />) },

      // Reports
      { path: 'pl', element: lazy(<ProfitLoss />) },
      { path: 'reports/balance-sheet', element: lazy(<BalanceSheet />) },
      { path: 'reports/cash-flow', element: lazy(<CashFlow />) },
      { path: 'reports/tax', element: lazy(<TaxReport />) },

      // Accounting
      { path: 'journal', element: lazy(<Journal />) },
      { path: 'accounts', element: lazy(<ChartOfAccounts />) },

      // Tax & settings
      { path: 'tax', element: lazy(<Tax />) },
      { path: 'settings', element: lazy(<Settings />) },
    ],
  },
];

export default AppRoutes;
