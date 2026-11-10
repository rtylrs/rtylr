//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountArticles` (Go `CountArticlesParams`).
pub struct CountArticlesParams {
    pub business_id: i64,
}

/// `CountArticles`: `one`.
pub async fn count_articles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountArticlesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_articles WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateArticle` (Go `CreateArticleParams`).
pub struct CreateArticleParams {
    pub business_id: i64,
    pub title: String,
    pub slug: String,
    pub body: String,
    pub category: Option<String>,
    pub is_published: Option<bool>,
}

/// `CreateArticle`: `execresult`.
pub async fn create_article<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateArticleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_articles (business_id, title, slug, body, category, is_published)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.title)
        .bind(&params.slug)
        .bind(&params.body)
        .bind(&params.category)
        .bind(params.is_published)
        .execute(exec)
        .await
}

/// Params for `DeleteArticle` (Go `DeleteArticleParams`).
pub struct DeleteArticleParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteArticle`: `exec`.
pub async fn delete_article<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteArticleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM support_articles WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetArticleByID` (Go `GetArticleByIDParams`).
pub struct GetArticleByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetArticleByID`: `one`.
pub async fn get_article_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetArticleByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at\nFROM support_articles\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetArticleBySlug` (Go `GetArticleBySlugParams`).
pub struct GetArticleBySlugParams {
    pub business_id: i64,
    pub slug: String,
}

/// `GetArticleBySlug`: `one`.
pub async fn get_article_by_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetArticleBySlugParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at\nFROM support_articles\nWHERE business_id = ? AND slug = ?\nORDER BY updated_at DESC, id DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `GetArticleByTitle` (Go `GetArticleByTitleParams`).
pub struct GetArticleByTitleParams {
    pub business_id: i64,
    pub title: String,
}

/// `GetArticleByTitle`: `one`.
pub async fn get_article_by_title<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetArticleByTitleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at\nFROM support_articles\nWHERE business_id = ? AND title = ?\nORDER BY updated_at DESC, id DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.title)
        .fetch_optional(exec)
        .await
}

/// Params for `ListArticles` (Go `ListArticlesParams`).
pub struct ListArticlesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListArticles`: `many`.
pub async fn list_articles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListArticlesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at\nFROM support_articles WHERE business_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateArticle` (Go `UpdateArticleParams`).
pub struct UpdateArticleParams {
    pub title: String,
    pub slug: String,
    pub body: String,
    pub category: Option<String>,
    pub is_published: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateArticle`: `exec`.
pub async fn update_article<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateArticleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE support_articles SET title = ?, slug = ?, body = ?, category = ?, is_published = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.title)
        .bind(&params.slug)
        .bind(&params.body)
        .bind(&params.category)
        .bind(params.is_published)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
