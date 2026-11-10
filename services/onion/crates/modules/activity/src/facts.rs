//! Audit-fact validation for the activity projection. Mirrors
//! `internal/infrastructure/facts.go`. The pure decision is ported here;
//! the sqlx `RecordMutation` call lands with the repository wiring.

use serde::Deserialize;
use thiserror::Error;

/// `audit.MutationRecorded`: the only event type that projects into the
/// activity log (`internal/contracts/audit`).
pub const MUTATION_RECORDED: &str = "entity_mutated";

/// Producers that publish audit facts. A fact from any other app is
/// rejected, which leaves it undispatched and retried.
pub const AUDIT_APPS: [&str; 3] = ["erp", "pos", "hub"];

#[derive(Debug, Error, PartialEq)]
pub enum FactError {
    #[error("invalid mutation fact identity")]
    InvalidIdentity,
    #[error("invalid mutation fact payload")]
    InvalidPayload,
}

#[derive(Debug, Clone, Default, Deserialize, PartialEq)]
pub struct MutationPayload {
    pub app: String,
    #[serde(default)]
    pub actor_id: i64,
    pub action: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub old_values: Option<serde_json::Value>,
    #[serde(default)]
    pub new_values: Option<serde_json::Value>,
    #[serde(default)]
    pub ip_address: String,
}

/// The identity fields a fact must carry (`eventbus.Event`).
#[derive(Debug, Clone, PartialEq)]
pub struct MutationFact {
    pub id: u64,
    pub business_id: i64,
    pub subject_type: String,
    pub subject_id: i64,
    pub payload: serde_json::Value,
}

/// `RecordFact`'s validation, before the repository writes anything.
/// `None` means "not a mutation fact" (the dispatcher skips it); `Some`
/// means the payload is ready to project.
pub fn validate_fact(
    fact_type: &str,
    fact: &MutationFact,
) -> Result<Option<MutationPayload>, FactError> {
    if fact_type != MUTATION_RECORDED {
        return Ok(None);
    }
    if fact.id == 0 || fact.business_id <= 0 || fact.subject_id <= 0 {
        return Err(FactError::InvalidIdentity);
    }
    let payload: MutationPayload =
        serde_json::from_value(fact.payload.clone()).map_err(|_| FactError::InvalidPayload)?;
    if !AUDIT_APPS.contains(&payload.app.as_str()) || payload.action.is_empty() {
        return Err(FactError::InvalidPayload);
    }
    Ok(Some(payload))
}

/// `nullableJSON`: a JSON `null` projects as SQL NULL.
pub fn nullable_json(value: &Option<serde_json::Value>) -> Option<Vec<u8>> {
    match value {
        None => None,
        Some(serde_json::Value::Null) => None,
        Some(other) => serde_json::to_vec(other).ok(),
    }
}

#[cfg(test)]
mod tests {
    use super::{nullable_json, validate_fact, FactError, MutationFact, MUTATION_RECORDED};

    fn fact(payload: serde_json::Value) -> MutationFact {
        MutationFact {
            id: 12,
            business_id: 1,
            subject_type: "order".into(),
            subject_id: 2,
            payload,
        }
    }

    #[test]
    fn non_mutation_events_are_skipped() {
        assert_eq!(
            validate_fact("order.created", &fact(serde_json::json!({}))).unwrap(),
            None
        );
    }

    #[test]
    fn identity_is_checked() {
        let mut bad = fact(serde_json::json!({"app":"pos","action":"created"}));
        bad.id = 0;
        assert_eq!(
            validate_fact(MUTATION_RECORDED, &bad),
            Err(FactError::InvalidIdentity)
        );
        let mut bad = fact(serde_json::json!({"app":"pos","action":"created"}));
        bad.business_id = 0;
        assert_eq!(
            validate_fact(MUTATION_RECORDED, &bad),
            Err(FactError::InvalidIdentity)
        );
    }

    #[test]
    fn payload_app_and_action_are_checked() {
        assert_eq!(
            validate_fact(
                MUTATION_RECORDED,
                &fact(serde_json::json!({"app":"crm","action":"created"}))
            ),
            Err(FactError::InvalidPayload)
        );
        assert_eq!(
            validate_fact(
                MUTATION_RECORDED,
                &fact(serde_json::json!({"app":"pos","action":""}))
            ),
            Err(FactError::InvalidPayload)
        );
        let ok = validate_fact(
            MUTATION_RECORDED,
            &fact(serde_json::json!({"app":"pos","action":"created","actor_id":5})),
        )
        .unwrap()
        .unwrap();
        assert_eq!(ok.app, "pos");
        assert_eq!(ok.actor_id, 5);
    }

    #[test]
    fn nullable_json_projects_null_as_none() {
        assert_eq!(nullable_json(&None), None);
        assert_eq!(nullable_json(&Some(serde_json::Value::Null)), None);
        assert_eq!(
            nullable_json(&Some(serde_json::json!({"a":1}))),
            Some(br#"{"a":1}"#.to_vec())
        );
    }
}
