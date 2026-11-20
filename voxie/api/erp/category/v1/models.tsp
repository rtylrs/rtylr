import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.category.v1;

model Category {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

model CreateCategoryRequest {
  @Onkt.required
  name: string;

  description: string;
}

model CreateCategoryResponse {
  @encode(string)
  category_id: int64;

  message: string;
}

model UpdateCategoryBody {
  @Onkt.required
  name: string;

  description: string;
}

model UpdateCategoryRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateCategoryBody;
}

model UpdateCategoryResponse {
  message: string;
}

model DeleteCategoryRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteCategoryResponse {
  message: string;
}

model ListCategoriesRequest {
  /** Maximum number of categories to return. */
  @query("limit")
  limit: int64;

  /** Number of categories to skip. */
  @query("offset")
  offset: int64;
}

model ListCategoriesResponse {
  categories: Category[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}
