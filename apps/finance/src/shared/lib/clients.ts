import { createOneKitClient, fromOneKitJson, oneKitResult } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient as TeamCompatibilityServiceClient } from '@voxire/voxie/hub/team/compat/v1/client';
import { DashboardServiceClient as FinanceDashboardServiceClient } from '@voxire/voxie/finance/dashboard/v1/client';
import { CompatibilityServiceClient as FinancialCompatibilityServiceClient } from '@voxire/voxie/insights/financial/compat/v1/client';
import { CompatibilityServiceClient as FinanceCompatibilityServiceClient } from '@voxire/voxie/finance/compat/v1/client';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { CompatibilityServiceClient as CrmCompatibilityServiceClient } from '@voxire/voxie/crm/compat/v1/client';

// licenseClient lives in ./api as a REST adapter: license/v1 has no contract
// in voxie (#894).
export { licenseClient } from './api';
export const userClient = createOneKitClient(UserServiceClient, 'finance');
export const teamCompatibilityClient = createOneKitClient(
  TeamCompatibilityServiceClient,
  'finance'
);
export const financeDashboardClient = createOneKitClient(FinanceDashboardServiceClient, 'finance');
export const financialCompatibilityClient = createOneKitClient(
  FinancialCompatibilityServiceClient,
  'finance'
);
export const financeCompatibilityClient = createOneKitClient(
  FinanceCompatibilityServiceClient,
  'finance'
);
export const invoicesClient = createOneKitClient(InvoicesServiceClient, 'finance');
export const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');
export const crmCompatibilityClient = createOneKitClient(CrmCompatibilityServiceClient, 'finance');

export const listCrmCustomers = () =>
  oneKitResult(() => crmCompatibilityClient.listCustomers({})).map((value) =>
    fromOneKitJson(value)
  );
