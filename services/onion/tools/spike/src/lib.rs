//! Unit 0.6 spike: one `@http` raw route and one contract service trait,
//! end to end over axum.
//!
//! Raw route: `GET|POST /crm/public/unsubscribe` (CRM compat), ported
//! 1:1 from `campaign_send_handler.go` (headers, status, HTML template).
//! The spike runs with no delivery secret / DB, which in Go is the
//! `delivery == nil` case: every POST is an invalid link (400).
//!
//! Service: `hub.activity.v1.ActivityService::list_activity` (1 method),
//! stub returning an empty page. DB, tenant resolution and cache are
//! Phase 3 (`activity` module); this proves trait + router + validation.

use std::sync::Arc;

use axum::{
    http::{HeaderMap, StatusCode},
    response::IntoResponse,
    routing::get,
    Router,
};
use contracts::generated::hub::activity::v1::{
    server::RequestContext, types::ListActivityResponse,
};

/// Empty-page implementation of the hub activity feed.
pub struct EmptyActivityService;

pub async fn list_activity_stub(
    _context: RequestContext,
    _req: contracts::generated::hub::activity::v1::types::ListActivityRequest,
) -> Result<
    ListActivityResponse,
    contracts::generated::hub::activity::v1::server::ActivityServiceListActivityServerError,
> {
    Ok(ListActivityResponse {
        activities: Vec::new(),
        total: 0,
    })
}

pub struct SpikeActivityService;

impl contracts::generated::hub::activity::v1::server::ActivityService for SpikeActivityService {
    async fn list_activity(
        &self,
        context: RequestContext,
        req: contracts::generated::hub::activity::v1::types::ListActivityRequest,
    ) -> Result<
        ListActivityResponse,
        contracts::generated::hub::activity::v1::server::ActivityServiceListActivityServerError,
    > {
        list_activity_stub(context, req).await
    }
}

fn escape_html(value: &str) -> String {
    let mut out = String::with_capacity(value.len());
    for ch in value.chars() {
        match ch {
            '&' => out.push_str("&amp;"),
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '"' => out.push_str("&#34;"),
            '\'' => out.push_str("&#39;"),
            _ => out.push(ch),
        }
    }
    out
}

fn unsubscribe_page(
    status: StatusCode,
    title: &str,
    message: &str,
    confirm_token: &str,
) -> axum::response::Response {
    let form = if confirm_token.is_empty() {
        String::new()
    } else {
        format!(
            r#"<form method="post" action="?t={}"><button type="submit" style="margin-top:16px;padding:10px 18px;border:0;border-radius:8px;background:#111;color:#fff;font-size:15px;cursor:pointer">Unsubscribe</button></form>"#,
            escape_html(confirm_token)
        )
    };
    let body = format!(
        "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>{}</title></head><body style=\"margin:0;padding:48px 24px;font-family:Helvetica,Arial,sans-serif;background:#f6f6f6;color:#111\"><main style=\"max-width:440px;margin:0 auto;background:#fff;border-radius:8px;padding:28px\"><h1 style=\"font-size:20px;margin:0 0 8px\">{}</h1><p style=\"margin:0;line-height:1.5;color:#444\">{}</p>{}</main></body></html>",
        escape_html(title),
        escape_html(title),
        escape_html(message),
        form
    );
    (
        status,
        [
            (axum::http::header::CONTENT_TYPE, "text/html; charset=utf-8"),
            (axum::http::header::CACHE_CONTROL, "no-store"),
            (
                axum::http::header::HeaderName::from_static("x-robots-tag"),
                "noindex",
            ),
        ],
        body,
    )
        .into_response()
}

async fn unsubscribe_get(
    headers: HeaderMap,
    axum::extract::Query(query): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> impl IntoResponse {
    let _ = headers;
    let token = query.get("t").map(String::as_str).unwrap_or_default();
    unsubscribe_page(
        StatusCode::OK,
        "Unsubscribe from marketing email?",
        "You will stop receiving marketing email from this business. Order and account messages are not affected.",
        token,
    )
}

async fn unsubscribe_post() -> impl IntoResponse {
    // Spike has no delivery secret / DB (`delivery == nil` in Go), so
    // `Unsubscribe` returns `ErrInvalidUnsubscribe` for every token.
    unsubscribe_page(
        StatusCode::BAD_REQUEST,
        "This link is not valid",
        "The unsubscribe link is incomplete or has been changed. Use the link from the email.",
        "",
    )
}

/// Router for the spike: one contract router + one raw route.
pub fn spike_router() -> Router {
    let activity = contracts::generated::hub::activity::v1::server::activity_service_router(
        Arc::new(SpikeActivityService),
    );
    activity.route(
        "/crm/public/unsubscribe",
        get(unsubscribe_get).post(unsubscribe_post),
    )
}

#[cfg(test)]
mod tests {
    use super::spike_router;
    use axum::http::Request;
    use tower::ServiceExt;

    async fn body_string(response: axum::response::Response) -> String {
        let parts = axum::body::to_bytes(response.into_body(), 1 << 20)
            .await
            .unwrap();
        String::from_utf8(parts.to_vec()).unwrap()
    }

    #[tokio::test]
    async fn activity_empty_page() {
        let response = spike_router()
            .oneshot(
                Request::builder()
                    .uri("/hub/activity/v1/activity?limit=10")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), axum::http::StatusCode::OK);
        let body = body_string(response).await;
        // Generated type skips empty `activities` on the wire.
        let value = serde_json::from_str::<serde_json::Value>(&body).unwrap();
        assert_eq!(value["total"], serde_json::json!(0));
    }

    #[tokio::test]
    async fn unsubscribe_get_confirm_page() {
        let response = spike_router()
            .oneshot(
                Request::builder()
                    .uri("/crm/public/unsubscribe?t=abc123")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), axum::http::StatusCode::OK);
        let body = body_string(response).await;
        assert!(
            body.contains("Unsubscribe from marketing email?"),
            "{}",
            body
        );
        assert!(body.contains("?t=abc123"), "{}", body);
    }

    #[tokio::test]
    async fn unsubscribe_post_invalid_without_delivery() {
        let response = spike_router()
            .oneshot(
                Request::builder()
                    .method("POST")
                    .uri("/crm/public/unsubscribe?t=abc123")
                    .body(axum::body::Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), axum::http::StatusCode::BAD_REQUEST);
        let body = body_string(response).await;
        assert!(body.contains("This link is not valid"), "{}", body);
    }
}
