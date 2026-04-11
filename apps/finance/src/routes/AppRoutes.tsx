import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../features/dashboard/Dashboard';
import Invoices from '../features/invoices/Invoices';
import InvoiceDetail from '../features/invoices/InvoiceDetail';
import NewInvoice from '../features/invoices/NewInvoice';
import RecurringInvoices from '../features/recurring-invoices/RecurringInvoices';
import Quotes from '../features/quotes/Quotes';
import Expenses from '../features/expenses/Expenses';
import ExpenseDetail from '../features/expenses/ExpenseDetail';
import NewExpense from '../features/expenses/NewExpense';
import Categories from '../features/categories/Categories';
import Bills from '../features/bills/Bills';
import BankFeeds from '../features/bank-feeds/BankFeeds';
import Reconcile from '../features/reconcile/Reconcile';
import ProfitLoss from '../features/profit-loss/ProfitLoss';
import BalanceSheet from '../features/balance-sheet/BalanceSheet';
import CashFlow from '../features/cash-flow/CashFlow';
import Tax from '../features/tax/Tax';
import TaxReport from '../features/tax/TaxReport';
import Journal from '../features/journal/Journal';
import ChartOfAccounts from '../features/chart-of-accounts/ChartOfAccounts';
import Settings from '../features/settings/Settings';

interface RouteChild {
  index?: boolean;
  path?: string;
  element: React.ReactElement;
}

interface AppRoute {
  path: string;
  element: React.ReactElement;
  children?: RouteChild[];
}

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true,                              element: <Dashboard /> },

      // Invoices
      { path: 'invoices',                         element: <Invoices /> },
      { path: 'invoices/new',                     element: <NewInvoice /> },
      { path: 'invoices/recurring',               element: <RecurringInvoices /> },
      { path: 'invoices/:id',                     element: <InvoiceDetail /> },
      { path: 'quotes',                           element: <Quotes /> },

      // Expenses
      { path: 'expenses',                         element: <Expenses /> },
      { path: 'expenses/new',                     element: <NewExpense /> },
      { path: 'expenses/:id',                     element: <ExpenseDetail /> },
      { path: 'categories',                       element: <Categories /> },

      // Bills & bank
      { path: 'bills',                            element: <Bills /> },
      { path: 'bank',                             element: <BankFeeds /> },
      { path: 'reconcile',                        element: <Reconcile /> },

      // Reports
      { path: 'pl',                               element: <ProfitLoss /> },
      { path: 'reports/balance-sheet',            element: <BalanceSheet /> },
      { path: 'reports/cash-flow',                element: <CashFlow /> },
      { path: 'reports/tax',                      element: <TaxReport /> },

      // Accounting
      { path: 'journal',                          element: <Journal /> },
      { path: 'accounts',                         element: <ChartOfAccounts /> },

      // Tax & settings
      { path: 'tax',                              element: <Tax /> },
      { path: 'settings',                         element: <Settings /> },
    ],
  },
];

export default AppRoutes;
