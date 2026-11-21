import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.bill.v1;

/** BillService defines accounts-payable bill endpoints. */
@route("/finance/bill/v1")
interface BillService {
  @get
  @route("/bills")
  listBills(...ListBillsRequest): ListBillsResponse;

  @post
  @route("/bills")
  createBill(...CreateBillRequest): CreateBillResponse;

  @put
  @route("/bills/{id}")
  updateBill(...UpdateBillRequest): UpdateBillResponse;

  @get
  @route("/bills/{id}")
  getBill(...GetBillRequest): GetBillResponse;

  @delete
  @route("/bills/{id}")
  deleteBill(...DeleteBillRequest): DeleteBillResponse;
}
