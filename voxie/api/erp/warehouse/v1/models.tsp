import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.warehouse.v1;

model Warehouse {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  address: string;
  city: string;
  country: string;
  manager_name: string;

  @encode(string)
  capacity: int64;

  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model CreateWarehouseRequest {
  @Onkt.required
  name: string;

  address: string;
  city: string;
  country: string;
  manager_name: string;

  @encode(string)
  capacity: int64;

  is_default: boolean;
  is_active: boolean;
}

model CreateWarehouseResponse {
  @encode(string)
  warehouse_id: int64;

  message: string;
}

model DeleteWarehouseRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteWarehouseResponse {
  message: string;
}

model ListWarehousesRequest {
  /** Maximum number of warehouses to return. */
  @query("limit")
  limit: int64;

  /** Number of warehouses to skip. */
  @query("offset")
  offset: int64;
}

model ListWarehousesResponse {
  warehouses: Warehouse[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}

model UpdateWarehouseBody {
  @Onkt.required
  name: string;

  address: string;
  city: string;
  country: string;
  manager_name: string;

  @encode(string)
  capacity: int64;

  is_default: boolean;
  is_active: boolean;
}

model UpdateWarehouseRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  body: UpdateWarehouseBody;
}

model UpdateWarehouseResponse {
  message: string;
}
