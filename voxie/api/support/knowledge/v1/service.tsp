import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.knowledge.v1;

@route("/support/knowledge/v1")
interface KnowledgeService {
  @get
  @route("/articles")
  listArticles(...ListArticlesRequest): ListArticlesResponse;

  @post
  @route("/articles")
  createArticle(...CreateArticleRequest): CreateArticleResponse;

  @put
  @route("/articles/{id}")
  updateArticle(...UpdateArticleRequest): UpdateArticleResponse;

  @delete
  @route("/articles/{id}")
  deleteArticle(...DeleteArticleRequest): DeleteArticleResponse;
}
