import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.payroll.v1;

@route("/hr/payroll/v1")
interface PayrollService {
  @get
  @route("/payroll")
  listPayroll(...ListPayrollRequest): ListPayrollResponse;

  @post
  @route("/payroll/run")
  runPayroll(...RunPayrollRequest): RunPayrollResponse;

  @post
  @route("/payroll/{id}/mark-paid")
  markPayrollPaid(...MarkPayrollPaidRequest): MarkPayrollPaidResponse;
}
