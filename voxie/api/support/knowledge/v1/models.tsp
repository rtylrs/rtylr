import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.knowledge.v1;

model Article {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  title: string;
  slug: string;
  body: string;
  category: string;
  is_published: boolean;
  views: int32;
  created_at: string;
  updated_at: string;
}

model ListArticlesRequest {
  @query("page")
  page: int32;

  @query("page_size")
  page_size: int32;
}

model ListArticlesResponse {
  articles: Article[];
  total: int32;
  page: int32;
  page_size: int32;
}

model CreateArticleRequest {
  @Onkt.required
  title: string;

  @Onkt.required
  slug: string;

  body: string;
  category: string;
  is_published: boolean;
}

model CreateArticleResponse {
  @encode(string)
  article_id: int64;

  message: string;
}

model UpdateArticleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  title: string;
  slug: string;
  body: string;
  category: string;
  is_published: boolean;
}

model UpdateArticleResponse {
  message: string;
}

model DeleteArticleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteArticleResponse {
  message: string;
}
