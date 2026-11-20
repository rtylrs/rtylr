import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.employees.payroll.runs.v1;

/** RunsService defines HR payroll run endpoints. */
@route("/hr/employees/payroll/runs/v1")
interface RunsService {
  /** ListPayrollRuns returns payroll runs. */
  @get
  @route("/runs")
  listPayrollRuns(...ListPayrollRunsRequest): ListPayrollRunsResponse;

  /** TriggerPayrollRun triggers a new payroll run. */
  @post
  @route("/runs/trigger")
  triggerPayrollRun(...TriggerPayrollRunRequest): TriggerPayrollRunResponse;

  /** ApprovePayrollRun approves a payroll run. */
  @patch
  @route("/runs/{id}/approve")
  approvePayrollRun(...ApprovePayrollRunRequest): ApprovePayrollRunResponse;

  /** ExportPayroll exports payroll data. */
  @post
  @route("/runs/export")
  exportPayroll(...ExportPayrollRequest): ExportPayrollResponse;
}
