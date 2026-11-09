//! Sanitized error writer. Mirrors `secure_errors.go` and
//! `recoverer.go`.

use serde::{Deserialize, Serialize};

/// Body written when a handler panics. Stack traces, paths, and internal
/// error details never reach the client.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct SecureErrorResponse {
    pub error: String,
    pub message: String,
    pub status: String,
}

/// The shape for any recovered panic (Go `SecureErrorResponse{...}`).
pub fn secure_error_response() -> SecureErrorResponse {
    SecureErrorResponse {
        error: "Internal Server Error".to_string(),
        message: "An unexpected error occurred. Please try again later.".to_string(),
        status: "error".to_string(),
    }
}

/// The JSON wire form of the error body.
pub fn secure_error_json() -> serde_json::Value {
    serde_json::json!({
        "error": "Internal Server Error",
        "message": "An unexpected error occurred. Please try again later.",
        "status": "error",
    })
}

/// `recoverMiddleware` writes a bare 500 header for upgrade
/// connections; the sanitized body comes from `SecureErrorMiddleware`.
/// The panic recovery logic is the same either way; this predicate
/// only selects the response shape.
pub fn recover_writes_body(is_upgrade_connection: bool) -> bool {
    !is_upgrade_connection
}

/// How a panic relates to the request: `http.ErrAbortHandler` is
/// re-raised (it is how a handler deliberately aborts a response).
pub fn is_abort_panic(is_abort_handler: bool) -> bool {
    is_abort_handler
}

/// The 413 body (Go `writeRequestEntityTooLarge` writes through the
/// shared OneKit error writer with message "Request body too large").
pub fn too_large_json() -> serde_json::Value {
    serde_json::json!({
        "message": "Request body too large",
    })
}

#[cfg(test)]
mod tests {
    use super::{recover_writes_body, secure_error_json, secure_error_response, too_large_json};

    #[test]
    fn shape() {
        let response = secure_error_response();
        assert_eq!(response.status, "error");
        assert_eq!(response.error, "Internal Server Error");
        assert_eq!(
            secure_error_json()["message"],
            "An unexpected error occurred. Please try again later."
        );
        assert!(recover_writes_body(false));
        assert!(!recover_writes_body(true));
        assert_eq!(too_large_json()["message"], "Request body too large");
    }
}
