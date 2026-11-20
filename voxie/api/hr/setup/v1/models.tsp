import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.setup.v1;

/** SetupWizardStepPayload is a single step in the HR setup wizard. */
model SetupWizardStepPayload {
  /** Unique identifier for the setup step record. */
  @encode(string)
  id: int64;

  /** Tenant owner this step belongs to. */
  @encode(string)
  owner_id: int64;

  /** Which setup step this represents (company_profile | work_schedule | leave_policy | org_structure | locations | roles_permissions | payroll | notifications | employees). */
  step_key: string;

  /** Unix timestamp when this step was completed; absent if not yet completed. */
  @encode(string)
  completed_at?: int64;

  /** Unix timestamp when this step record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this step record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** HRCompanyConfigPayload is the tenant-wide HR configuration set during setup. */
model HRCompanyConfigPayload {
  /** Unique identifier for the config record. */
  @encode(string)
  id: int64;

  /** Tenant owner this config belongs to. */
  @encode(string)
  owner_id: int64;

  /** Shared company display name from business settings. */
  company_name?: string;

  /** Default currency used for HR payroll and compensation. */
  default_currency?: string;

  /** Default timezone used for HR scheduling. */
  default_timezone?: string;

  /** How time is tracked (time_logging | attendance | both). */
  time_tracking_model: string;

  /** Period for timesheet submission (weekly | biweekly | monthly). */
  timesheet_period: string;

  /** Period for payroll runs (weekly | biweekly | monthly). */
  payroll_period: string;

  /** How gross pay is calculated (hourly | salary | both). */
  payroll_gross_model: string;

  /** Unix timestamp when the full setup wizard was completed; absent if still in progress. */
  @encode(string)
  setup_completed_at?: int64;

  /** Unix timestamp when this config was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this config was last updated. */
  @encode(string)
  updated_at: int64;

  /**
   * Enabled clock-in methods for attendance tracking
   * (values: "app", "pin", "gps", "kiosk"). Empty defaults to ["app"].
   */
  clock_in_methods: string[];

  /**
   * App IDs this tenant is subscribed to (values: "flow", "finance",
   * "erp", "crm", "pos", "hr"). Drives cross-app UI affordances such as
   * the "Import from Flow" button on HR timesheets.
   */
  subscribed_apps: string[];

  /**
   * Whether the company-wide staff directory is visible to employees
   * outside HR (e.g. Flow's Directory screen). Defaults to false: the
   * business owner must opt in before any employee's role/department/
   * contact info is exposed beyond HR's own admin screens.
   */
  directory_enabled: boolean;
}

/** GetSetupStatusResponse summarises setup wizard progress and current HR configuration. */
model GetSetupStatusResponse {
  /** All wizard steps and their completion state. */
  steps: SetupWizardStepPayload[];

  /** Current HR configuration for the tenant. */
  config: HRCompanyConfigPayload;

  /** Whether every setup step has been completed. */
  is_complete: boolean;
}

model CompleteSetupStepRequest {
  /** The step to mark as completed (must match a valid step_key). */
  @Onkt.required
  step_key: string;
}

model CompleteSetupStepResponse {
  /** The step after being marked complete. */
  step: SetupWizardStepPayload;
}
