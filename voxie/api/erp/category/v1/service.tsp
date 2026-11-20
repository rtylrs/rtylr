import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.category.v1;

@route("/erp/category/v1")
interface CategoryService {
  @get
  @route("/categories")
  listCategories(...ListCategoriesRequest): ListCategoriesResponse;

  @post
  @route("/categories")
  createCategory(...CreateCategoryRequest): CreateCategoryResponse;

  @put
  @route("/categories/{id}")
  updateCategory(...UpdateCategoryRequest): UpdateCategoryResponse;

  @delete
  @route("/categories/{id}")
  deleteCategory(...DeleteCategoryRequest): DeleteCategoryResponse;
}
