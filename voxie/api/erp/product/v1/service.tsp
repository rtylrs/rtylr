import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.product.v1;

/** ProductService defines ERP product endpoints. */
@route("/erp/product/v1")
interface ProductService {
  @get
  @route("/products")
  listProducts(...ListProductsRequest): ListProductsResponse;

  @post
  @route("/products")
  createProduct(...CreateProductRequest): CreateProductResponse;

  @get
  @route("/barcodes/{code}/product")
  getProductByBarcode(
    ...GetProductByBarcodeRequest,
  ): GetProductByBarcodeResponse;

  @get
  @route("/products/{id}")
  getProduct(...GetProductRequest): GetProductResponse;

  @get
  @route("/products/{id}/full")
  getProductFull(...GetProductFullRequest): GetProductFullResponse;

  @get
  @route("/products/{id}/attributes")
  listProductAttributes(
    ...ListProductAttributesRequest,
  ): ListProductAttributesResponse;

  @put
  @route("/products/{id}")
  updateProduct(...UpdateProductRequest): UpdateProductResponse;

  @delete
  @route("/products/{id}")
  deleteProduct(...DeleteProductRequest): DeleteProductResponse;

  @get
  @route("/dashboard")
  getDashboard(...GetDashboardRequest): GetDashboardResponse;

  @get
  @route("/reorder-suggestions")
  listReorderSuggestions(
    ...ListReorderSuggestionsRequest,
  ): ListReorderSuggestionsResponse;

  @get
  @route("/expiry")
  listExpiringItems(...ListExpiringItemsRequest): ListExpiringItemsResponse;
}
