export type PermissionLevel = 'none' | 'viewer' | 'editor';

export interface AppSection {
  key: string; // e.g. 'erp.products'
  label: string;
}

export interface AppDef {
  id: string;
  label: string;
  sections: AppSection[];
}

export const APP_RESOURCES: AppDef[] = [
  {
    id: 'erp',
    label: 'ERP',
    sections: [
      { key: 'erp.products', label: 'Products' },
      { key: 'erp.inventory', label: 'Inventory & Stock' },
      { key: 'erp.purchase-orders', label: 'Purchase Orders' },
      { key: 'erp.suppliers', label: 'Suppliers' },
      { key: 'erp.price-lists', label: 'Price Lists' },
    ],
  },
  {
    id: 'crm',
    label: 'CRM',
    sections: [
      { key: 'crm.customers', label: 'Customers' },
      { key: 'crm.loyalty', label: 'Loyalty Program' },
      { key: 'crm.campaigns', label: 'Campaigns' },
      { key: 'crm.segments', label: 'Segments' },
      { key: 'crm.automations', label: 'Automations' },
    ],
  },
  {
    id: 'finance',
    label: 'Finance',
    sections: [
      { key: 'finance.invoices', label: 'Invoices' },
      { key: 'finance.expenses', label: 'Expenses' },
      { key: 'finance.bills', label: 'Bills' },
      { key: 'finance.quotes', label: 'Quotes' },
      { key: 'finance.reports', label: 'Reports (P&L, Balance Sheet)' },
    ],
  },
  {
    id: 'hr',
    label: 'HR',
    sections: [
      { key: 'hr.staff', label: 'Staff' },
      { key: 'hr.shifts', label: 'Shifts' },
      { key: 'hr.payroll', label: 'Payroll' },
      { key: 'hr.leave', label: 'Leave Management' },
    ],
  },
  {
    id: 'pos',
    label: 'POS',
    sections: [
      { key: 'pos.orders', label: 'Orders' },
      { key: 'pos.menu', label: 'Menu' },
      { key: 'pos.reports', label: 'Reports' },
      { key: 'pos.kds', label: 'Kitchen Display' },
    ],
  },
  {
    id: 'flow',
    label: 'Flow',
    sections: [
      { key: 'flow.projects', label: 'Projects' },
      { key: 'flow.inbox', label: 'Inbox' },
    ],
  },
  {
    id: 'dash',
    label: 'Account Home',
    sections: [{ key: 'dash.overview', label: 'Overview' }],
  },
  {
    id: 'insights',
    label: 'Insights',
    sections: [
      { key: 'insights.revenue', label: 'Revenue' },
      { key: 'insights.customers', label: 'Customers' },
      { key: 'insights.products', label: 'Products' },
      { key: 'insights.operations', label: 'Operations' },
    ],
  },
];
