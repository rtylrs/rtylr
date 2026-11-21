import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.settings.v1;

/** ErpBusinessSettings stores tenant-level ERP defaults shared by operators. */
model ErpBusinessSettings {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @Onkt.required
  currency: string;

  @Onkt.required
  @minValue(0)
  @maxValue(100)
  tax_rate?: float64;

  @Onkt.required
  @minValue(0)
  low_stock_threshold?: float64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  default_warehouse_id?: int64;

  notify_on_low_stock: boolean;
  notify_on_expiry: boolean;
  created_at: string;
  updated_at: string;
}

model GetErpSettingsRequest {}

model GetErpSettingsResponse {
  @Onkt.required
  settings: ErpBusinessSettings;
}

model UpdateErpSettingsRequest {
  @Onkt.required
  currency: string;

  @Onkt.required
  @minValue(0)
  @maxValue(100)
  tax_rate?: float64;

  @Onkt.required
  @minValue(0)
  low_stock_threshold?: float64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  default_warehouse_id?: int64;

  notify_on_low_stock: boolean;
  notify_on_expiry: boolean;
}

model UpdateErpSettingsResponse {
  @Onkt.required
  settings: ErpBusinessSettings;

  message: string;
}
