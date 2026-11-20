import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.tax.v1;

/**
 * Renamed from TaxConfig to avoid a name collision with the distinct
 * finance.report.v1.TaxConfig - onk message names are unique across the
 * whole schema tree, unlike proto's per-package scoping.
 */
model TaxConfig {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  tax_name: string;
  tax_rate: float64;
  is_inclusive: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model GetTaxConfigRequest {}

model GetTaxConfigResponse {
  config: TaxConfig;
}

model UpdateTaxConfigRequest {
  @Onkt.required
  tax_name: string;

  @minValue(0)
  tax_rate: float64;

  is_inclusive: boolean;
  is_active: boolean;
}

model UpdateTaxConfigResponse {
  message: string;
}

/**
 * Renamed from GetPLReportRequest to avoid a name collision with the
 * distinct finance.report.v1.GetPLReportRequest.
 */
model GetPLReportRequest {}

/**
 * Renamed from GetPLReportResponse to avoid a name collision with the
 * distinct finance.report.v1.GetPLReportResponse.
 */
model GetPLReportResponse {
  total_revenue: float64;
  total_expenses: float64;
  gross_profit: float64;
  net_profit: float64;
}
