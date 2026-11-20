import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.variant.v1;

@route("/erp/variant/v1")
interface VariantService {
  @get
  @route("/variants")
  listVariants(...ListVariantsRequest): ListVariantsResponse;

  @post
  @route("/variants")
  createVariant(...CreateVariantRequest): CreateVariantResponse;

  @put
  @route("/variants/{id}")
  updateVariant(...UpdateVariantRequest): UpdateVariantResponse;

  @delete
  @route("/variants/{id}")
  deleteVariant(...DeleteVariantRequest): DeleteVariantResponse;
}
