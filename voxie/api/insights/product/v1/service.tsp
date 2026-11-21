import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.product.v1;

/** ProductService defines insights product analytics endpoints. */
@route("/insights/product/v1")
interface ProductService {
  /** Returns top-selling products. */
  @get
  @route("/best-sellers")
  getBestSellers(...GetBestSellersRequest): GetBestSellersResponse;

  /** Returns performance by category. */
  @get
  @route("/categories")
  getCategoryPerformance(
    ...GetCategoryPerformanceRequest,
  ): GetCategoryPerformanceResponse;

  /** Returns product margin analysis. */
  @get
  @route("/margins")
  getProductMargins(...GetProductMarginsRequest): GetProductMarginsResponse;

  /** Returns slow-moving products. */
  @get
  @route("/slow-movers")
  getSlowMovers(...GetSlowMoversRequest): GetSlowMoversResponse;

  /** Returns aggregate product KPI counters. */
  @get
  @route("/performance")
  getProductPerformance(
    ...GetProductPerformanceRequest,
  ): GetProductPerformanceResponse;
}
