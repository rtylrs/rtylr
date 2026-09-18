import { APP_NAMES } from './appNames';

export type PermissionLevel = 'none' | 'viewer' | 'editor';

export interface AppSection {
  /**
   * The stored permission identifier, e.g. 'erp.products'.
   *
   * This is a value, not a label: it is what a saved permission row holds, so
   * it must not change when the wording does.
   */
  key: string;
  /** i18n key for the name shown beside the permission control. */
  labelKey: string;
}

export interface AppDef {
  id: string;
  /** i18n key for the app's name. */
  labelKey: string;
  sections: AppSection[];
}

export const APP_RESOURCES: AppDef[] = [
  {
    id: 'erp',
    labelKey: APP_NAMES.erp.short,
    sections: [
      { key: 'erp.products', labelKey: 'appNames.sections.erp.products' },
      { key: 'erp.inventory', labelKey: 'appNames.sections.erp.inventory' },
      { key: 'erp.purchase-orders', labelKey: 'appNames.sections.erp.purchase-orders' },
      { key: 'erp.suppliers', labelKey: 'appNames.sections.erp.suppliers' },
      { key: 'erp.price-lists', labelKey: 'appNames.sections.erp.price-lists' },
    ],
  },
  {
    id: 'crm',
    labelKey: APP_NAMES.crm.short,
    sections: [
      { key: 'crm.customers', labelKey: 'appNames.sections.crm.customers' },
      { key: 'crm.loyalty', labelKey: 'appNames.sections.crm.loyalty' },
      { key: 'crm.campaigns', labelKey: 'appNames.sections.crm.campaigns' },
      { key: 'crm.segments', labelKey: 'appNames.sections.crm.segments' },
      { key: 'crm.automations', labelKey: 'appNames.sections.crm.automations' },
    ],
  },
  {
    id: 'finance',
    labelKey: APP_NAMES.finance.short,
    sections: [
      { key: 'finance.invoices', labelKey: 'appNames.sections.finance.invoices' },
      { key: 'finance.expenses', labelKey: 'appNames.sections.finance.expenses' },
      { key: 'finance.bills', labelKey: 'appNames.sections.finance.bills' },
      { key: 'finance.quotes', labelKey: 'appNames.sections.finance.quotes' },
      { key: 'finance.reports', labelKey: 'appNames.sections.finance.reports' },
    ],
  },
  {
    id: 'hr',
    labelKey: APP_NAMES.hr.short,
    sections: [
      { key: 'hr.staff', labelKey: 'appNames.sections.hr.staff' },
      { key: 'hr.shifts', labelKey: 'appNames.sections.hr.shifts' },
      { key: 'hr.payroll', labelKey: 'appNames.sections.hr.payroll' },
      { key: 'hr.leave', labelKey: 'appNames.sections.hr.leave' },
    ],
  },
  {
    id: 'pos',
    labelKey: APP_NAMES.pos.short,
    sections: [
      { key: 'pos.orders', labelKey: 'appNames.sections.pos.orders' },
      { key: 'pos.menu', labelKey: 'appNames.sections.pos.menu' },
      { key: 'pos.reports', labelKey: 'appNames.sections.pos.reports' },
      { key: 'pos.kds', labelKey: 'appNames.sections.pos.kds' },
    ],
  },
  {
    id: 'flow',
    labelKey: APP_NAMES.flow.short,
    sections: [
      { key: 'flow.projects', labelKey: 'appNames.sections.flow.projects' },
      { key: 'flow.inbox', labelKey: 'appNames.sections.flow.inbox' },
    ],
  },
  {
    id: 'dash',
    labelKey: 'appNames.accountHome',
    sections: [{ key: 'dash.overview', labelKey: 'appNames.sections.dash.overview' }],
  },
  {
    id: 'insights',
    labelKey: APP_NAMES.insights.short,
    sections: [
      { key: 'insights.revenue', labelKey: 'appNames.sections.insights.revenue' },
      { key: 'insights.customers', labelKey: 'appNames.sections.insights.customers' },
      { key: 'insights.products', labelKey: 'appNames.sections.insights.products' },
      { key: 'insights.operations', labelKey: 'appNames.sections.insights.operations' },
    ],
  },
];
