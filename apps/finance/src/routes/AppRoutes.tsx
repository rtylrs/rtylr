import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Invoices from '../screens/Invoices/Invoices';
import InvoiceDetail from '../screens/InvoiceDetail/InvoiceDetail';
import NewInvoice from '../screens/NewInvoice/NewInvoice';
import RecurringInvoices from '../screens/RecurringInvoices/RecurringInvoices';
import Quotes from '../screens/Quotes/Quotes';
import Expenses from '../screens/Expenses/Expenses';
import ExpenseDetail from '../screens/ExpenseDetail/ExpenseDetail';
import NewExpense from '../screens/NewExpense/NewExpense';
import Categories from '../screens/Categories/Categories';
import Bills from '../screens/Bills/Bills';
import BankFeeds from '../screens/BankFeeds/BankFeeds';
import Reconcile from '../screens/Reconcile/Reconcile';
import ProfitLoss from '../screens/ProfitLoss/ProfitLoss';
import BalanceSheet from '../screens/BalanceSheet/BalanceSheet';
import CashFlow from '../screens/CashFlow/CashFlow';
import Tax from '../screens/Tax/Tax';
import TaxReport from '../screens/TaxReport/TaxReport';
import Journal from '../screens/Journal/Journal';
import ChartOfAccounts from '../screens/ChartOfAccounts/ChartOfAccounts';
import Settings from '../screens/Settings/Settings';

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
