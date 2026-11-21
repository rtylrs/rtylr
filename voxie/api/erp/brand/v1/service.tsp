import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.brand.v1;

@route("/erp/brand/v1")
interface BrandService {
  @get
  @route("/brands")
  listBrands(...ListBrandsRequest): ListBrandsResponse;

  @post
  @route("/brands")
  createBrand(...CreateBrandRequest): CreateBrandResponse;

  @put
  @route("/brands/{id}")
  updateBrand(...UpdateBrandRequest): UpdateBrandResponse;

  @delete
  @route("/brands/{id}")
  deleteBrand(...DeleteBrandRequest): DeleteBrandResponse;
}
