import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.supplier.v1;

/** Supplier is a business's registered supplier record. */
model Supplier {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** email is the email address used for lookup or contact. */
  email: string;

  /** phone is the phone value for this contract. */
  phone: string;

  /** address is the address value for this contract. */
  address: string;

  /** city is the city value for this contract. */
  city: string;

  /** country is the country value for this contract. */
  country: string;

  /** tax_id is the tax id value for this contract. */
  tax_id: string;

  /** payment_terms is the payment terms value for this contract. */
  payment_terms: string;

  /** notes is the notes value for this contract. */
  notes: string;

  /** is_active is the is active value for this contract. */
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateSupplierRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** email is the email address used for lookup or contact. */
  @Onkt.required
  @format("email")
  email: string;

  /** phone is the phone value for this contract. */
  phone: string;

  /** address is the address value for this contract. */
  address: string;

  /** city is the city value for this contract. */
  @Onkt.required
  city: string;

  /** country is the country value for this contract. */
  @Onkt.required
  country: string;

  /** tax_id is the tax id value for this contract. */
  tax_id: string;

  /** payment_terms is the payment terms value for this contract. */
  payment_terms: string;

  /** notes is the notes value for this contract. */
  notes: string;

  /** is_active is the is active value for this contract. */
  is_active: boolean;
}

model CreateSupplierResponse {
  /** supplier_id is the supplier id value for this contract. */
  @encode(string)
  supplier_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model DeleteSupplierRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteSupplierResponse {
  message: string;
}

model ListSuppliersRequest {
  /** page is the 1-based page number. */
  @query("page")
  page: int32;

  /** page_size is the number of rows to return. */
  @query("page_size")
  page_size: int32;
}

model ListSuppliersResponse {
  /** suppliers is the suppliers value for this contract. */
  suppliers: Supplier[];

  /** has_more is true when another page is available. */
  has_more: boolean;

  /** page is the returned 1-based page number. */
  page: int32;

  /** page_size is the returned page size. */
  page_size: int32;
}

model UpdateSupplierBody {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** email is the email address used for lookup or contact. */
  @Onkt.required
  @format("email")
  email: string;

  /** phone is the phone value for this contract. */
  phone: string;

  /** address is the address value for this contract. */
  address: string;

  /** city is the city value for this contract. */
  @Onkt.required
  city: string;

  /** country is the country value for this contract. */
  @Onkt.required
  country: string;

  /** tax_id is the tax id value for this contract. */
  tax_id: string;

  /** payment_terms is the payment terms value for this contract. */
  payment_terms: string;

  /** notes is the notes value for this contract. */
  notes: string;

  /** is_active is the is active value for this contract. */
  is_active: boolean;
}

model UpdateSupplierRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable supplier fields bound from the HTTP payload. */
  body: UpdateSupplierBody;
}

model UpdateSupplierResponse {
  /** message is a human-readable result message. */
  message: string;
}
