import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.settings.v1;

/** POSFeatureSettings controls which business-specific POS modules are visible. */
model POSFeatureSettings {
  /** tables enables table and floor service workflows. */
  tables: boolean;

  /** kitchen_display enables KDS and kitchen preparation workflows. */
  kitchen_display: boolean;

  /** modifiers enables product modifier workflows. */
  modifiers: boolean;

  /** combos enables bundle/combo workflows. */
  combos: boolean;

  /** tips enables tipping during payment. */
  tips: boolean;

  /** gift_cards enables gift card selling and redemption. */
  gift_cards: boolean;

  /** customers enables customer lookup and customer assignment. */
  customers: boolean;

  /** inventory enables stock/inventory surfaces. */
  inventory: boolean;

  /** cash_drawer enables cash drawer actions and cash movement surfaces. */
  cash_drawer: boolean;
}

/** POSBusinessSettings stores business-wide POS experience settings. */
model POSBusinessSettings {
  @encode(string)
  id: int64;

  /** business_id is the business id these settings belong to. */
  @encode(string)
  business_id: int64;

  /** business_profile is one of retail, restaurant, or quick_service. */
  business_profile: string;

  /** theme is the default POS theme for this business. */
  theme: string;

  /** features are the business-specific POS modules to expose. */
  features: POSFeatureSettings;

  /** receipt_footer is the default footer printed on receipts. */
  receipt_footer: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

/** GetPOSSettingsRequest reads the current business POS settings. */
model GetPOSSettingsRequest {}

/** GetPOSSettingsResponse returns the current business POS settings. */
model GetPOSSettingsResponse {
  settings: POSBusinessSettings;
}

/** UpdatePOSSettingsRequest updates the current business POS settings. */
model UpdatePOSSettingsRequest {
  /** business_profile is one of retail, restaurant, or quick_service. */
  business_profile: string;

  /** theme is the default POS theme for this business. */
  theme: string;

  /** features are the business-specific POS modules to expose. */
  features: POSFeatureSettings;

  /** receipt_footer is the default footer printed on receipts. */
  receipt_footer: string;
}

/** UpdatePOSSettingsResponse returns the saved POS settings. */
model UpdatePOSSettingsResponse {
  settings: POSBusinessSettings;
  message: string;
}
