import type { HRCompanyConfigPayload } from '@voxire/voxie/hr/setup/v1/types';

/**
 * The setup contract predates the payroll capability flags, so deployed API
 * responses can contain either camelCase or wire-format snake_case fields.
 */
export type PayrollConfigSource = HRCompanyConfigPayload & {
  payrollEnabled?: boolean | string;
  payrollConfigured?: boolean | string;
  payroll_enabled?: boolean | string;
  payroll_configured?: boolean | string;
  payroll_period?: string;
};

const readBoolean = (value: boolean | string | undefined): boolean | undefined => {
  if (typeof value === 'boolean') return value;
  if (value === 'true') return true;
  if (value === 'false') return false;
  return undefined;
};

/**
 * Payroll is usable only when it has a period and is not explicitly disabled
 * or unconfigured. Fully populated capability flags are accepted even when an
 * older response omits the legacy period field.
 */
export function isPayrollConfigured(
  config: HRCompanyConfigPayload | PayrollConfigSource | null | undefined
): boolean {
  if (!config) return false;

  const source = config as PayrollConfigSource;
  const enabled = readBoolean(source.payrollEnabled ?? source.payroll_enabled);
  const configured = readBoolean(source.payrollConfigured ?? source.payroll_configured);

  if (enabled === false || configured === false) return false;
  if (enabled === true && configured === true) return true;

  return Boolean(source.payrollPeriod ?? source.payroll_period);
}
