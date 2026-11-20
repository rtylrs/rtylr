import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.supplier.v1;

/** SupplierService defines ERP supplier endpoints. */
@route("/erp/supplier/v1")
interface SupplierService {
  @get
  @route("/suppliers")
  listSuppliers(...ListSuppliersRequest): ListSuppliersResponse;

  @post
  @route("/suppliers")
  createSupplier(...CreateSupplierRequest): CreateSupplierResponse;

  @put
  @route("/suppliers/{id}")
  updateSupplier(...UpdateSupplierRequest): UpdateSupplierResponse;

  @delete
  @route("/suppliers/{id}")
  deleteSupplier(...DeleteSupplierRequest): DeleteSupplierResponse;
}
