import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.warehouse.v1;

@route("/erp/warehouse/v1")
interface WarehouseService {
  @get
  @route("/warehouses")
  listWarehouses(...ListWarehousesRequest): ListWarehousesResponse;

  @post
  @route("/warehouses")
  createWarehouse(...CreateWarehouseRequest): CreateWarehouseResponse;

  @put
  @route("/warehouses/{id}")
  updateWarehouse(...UpdateWarehouseRequest): UpdateWarehouseResponse;

  @delete
  @route("/warehouses/{id}")
  deleteWarehouse(...DeleteWarehouseRequest): DeleteWarehouseResponse;
}
