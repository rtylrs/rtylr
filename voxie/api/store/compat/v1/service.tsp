import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace store.compat.v1;

/** Legacy storefront routes, served with their existing wire format. */
@route("/store")
interface CompatibilityService {
  @delete
  @route("/categories")
  deleteCategory(...EmptyRequest): JsonResponse;

  @delete
  @route("/products")
  deleteProduct(...EmptyRequest): JsonResponse;

  @delete
  @route("/shipping")
  deleteShipping(...EmptyRequest): JsonResponse;

  @get
  @route("/categories")
  listCategories(...EmptyRequest): JsonResponse;

  @get
  @route("/orders")
  listOrders(...EmptyRequest): JsonResponse;

  @get
  @route("/orders/detail")
  getOrderDetail(...EmptyRequest): JsonResponse;

  @get
  @route("/payments")
  listPayments(...EmptyRequest): JsonResponse;

  @get
  @route("/products")
  listProducts(...EmptyRequest): JsonResponse;

  @get
  @route("/shipping")
  listShipping(...EmptyRequest): JsonResponse;

  @get
  @route("/sites")
  listSites(...EmptyRequest): JsonResponse;

  @get
  @route("/themes")
  listThemes(...EmptyRequest): JsonResponse;

  @post
  @route("/categories")
  createCategory(...BodyRequest): JsonResponse;

  @post
  @route("/checkout/complete")
  completeCheckout(...BodyRequest): JsonResponse;

  @post
  @route("/checkout/create")
  createCheckout(...BodyRequest): JsonResponse;

  @post
  @route("/orders")
  createOrder(...BodyRequest): JsonResponse;

  @post
  @route("/payments")
  createPayment(...BodyRequest): JsonResponse;

  @post
  @route("/products")
  createProduct(...BodyRequest): JsonResponse;

  @post
  @route("/shipping")
  createShipping(...BodyRequest): JsonResponse;

  @post
  @route("/sites")
  createSite(...BodyRequest): JsonResponse;

  @post
  @route("/themes")
  createTheme(...BodyRequest): JsonResponse;

  @put
  @route("/categories")
  updateCategory(...BodyRequest): JsonResponse;

  @put
  @route("/orders")
  updateOrder(...BodyRequest): JsonResponse;

  @put
  @route("/payments")
  updatePayment(...BodyRequest): JsonResponse;

  @put
  @route("/products")
  updateProduct(...BodyRequest): JsonResponse;

  @put
  @route("/shipping")
  updateShipping(...BodyRequest): JsonResponse;

  @put
  @route("/sites")
  updateSite(...BodyRequest): JsonResponse;

  @put
  @route("/themes")
  updateTheme(...BodyRequest): JsonResponse;
}
