import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.payroll.v1;

model PayrollRecord {
  id: int64;
  owner_id: int64;
  staff_id: int64;
  period_start: string;
  period_end: string;
  period_label: string;
  regular_hours: float64;
  overtime_hours: float64;
  base_pay: float64;
  overtime_pay: float64;
  bonuses: float64;
  deductions: float64;
  tax_withheld: float64;
  net_pay: float64;
  status: string;
  paid_at: string;
  notes: string;
  created_at: string;
}

model ListPayrollRequest {
  @query("period_label")
  period_label: string;
}

model ListPayrollResponse {
  records: PayrollRecord[];
}

model RunPayrollRequest {
  @Onkt.required
  period_start: string;

  @Onkt.required
  period_end: string;

  @Onkt.required
  period_label: string;
}

model RunPayrollResponse {
  records_created: int32;
}

model MarkPayrollPaidRequest {
  @minValueExclusive(0)
  id: int64;
}

model MarkPayrollPaidResponse {
  success: boolean;
}
