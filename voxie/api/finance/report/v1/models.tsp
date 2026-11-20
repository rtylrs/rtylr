import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.report.v1;

// -- Tax Config --
model TaxConfig {
  @encode(string)
  id: int64;

  /** Identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  name: string;
  rate: float64;
  type: string;
  is_default: boolean;
  is_active: boolean;

  /** When the record was created. */
  created_at: string;
}

// -- Expense Categories (from invoices module) --
model ExpenseCategoryDetail {
  @encode(string)
  id: int64;

  /** Identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  name: string;
  color: string;
  icon: string;

  /** When the record was created. */
  created_at: string;
}

model GetPLReportRequest {
  /** The reporting period label. */
  @query("period")
  period: string;

  /** The report anchor date. */
  @Onkt.required
  @query("date")
  date: string;
}

model GetPLReportResponse {
  period: string;
  revenue: float64;
  cogs: float64;
  gross_profit: float64;
  operating_expenses: float64;
  net_profit: float64;
  margin_pct: float64;

  /**
   * expenses_by_category splits operating_expenses by expense category,
   * largest first.
   */
  expenses_by_category: PLCategoryTotal[];

  /**
   * pending_expenses totals the period's expenses still awaiting approval.
   * They are not in operating_expenses until approved.
   */
  pending_expenses: float64;

  /** pending_expense_count counts the expenses in pending_expenses. */
  @encode(string)
  pending_expense_count: int64;
}

/**
 * PLCategoryTotal is one expense category's share of a period's approved
 * expenses.
 */
model PLCategoryTotal {
  /** category_id is the expense category; absent for uncategorized expenses. */
  @encode(string)
  category_id?: int64;

  /** name is the category name; empty for uncategorized expenses. */
  name: string;

  /** color is the category colour as a hex string. */
  color: string;

  /** amount is the approved expense total for the category. */
  amount: float64;
}

model GetTaxReportRequest {
  /** The reporting period label. */
  @query("period")
  period: string;

  /** The report anchor date. */
  @Onkt.required
  @query("date")
  date: string;
}

model GetTaxReportResponse {
  period: string;
  tax_collected: float64;
  tax_owed: float64;
}

model ListExpenseCategoryDetailsRequest {}

model ListExpenseCategoryDetailsResponse {
  categories: ExpenseCategoryDetail[];
}

model CreateExpenseCategoryDetailRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  color: string;

  @Onkt.required
  icon: string;
}

model CreateExpenseCategoryDetailResponse {
  category: ExpenseCategoryDetail;
}

model ListTaxConfigRequest {}

model ListTaxConfigResponse {
  configs: TaxConfig[];
}

model CreateTaxConfigRequest {
  @Onkt.required
  name: string;

  @minValue(0)
  rate: float64;

  @Onkt.required
  type: string;

  is_default: boolean;
}

model CreateTaxConfigResponse {
  config: TaxConfig;
}
