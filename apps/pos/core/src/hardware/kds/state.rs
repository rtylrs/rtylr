//! KDS item-level state machine.
//!
//! Today the KDS surface tracks state at the ticket level. To support
//! expo, all-day mode, course firing, and recall, every line item needs
//! its own state so cooks can bump individual plates without affecting
//! the rest of the ticket.
//!
//! Transitions are intentionally narrow. The only backwards move
//! allowed is `Bumped -> Ready` via `recall`, which exists for the
//! `recall last bumped` feature. Every other illegal move returns an
//! error so frontend bugs surface at the boundary, not silently as
//! lost state.
//!
//! Bumps are idempotent: `version` increments on every transition so
//! a duplicate bump-bar tap that lands twice is a no-op rather than
//! advancing past `Bumped`.

use serde::{Deserialize, Serialize};

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ItemState {
    /// Course-fire feature: item is held server-side until the server
    /// fires the course. Newly-created items default to `Queued`; only
    /// items belonging to non-fired courses start `Held`.
    Held,
    /// In the station's incoming queue. Default starting state.
    Queued,
    /// A cook has begun preparing this item.
    Cooking,
    /// Item is plated and waiting at the station for expo/bump.
    Ready,
    /// Item has been bumped from the station (handed off to expo or guest).
    Bumped,
    /// Item was voided after firing. Terminal.
    Voided,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ItemEvent {
    /// Fire a held item into the station queue (course fire).
    Fire,
    /// A cook started working on the item.
    StartCooking,
    /// Mark the item plated and waiting.
    MarkReady,
    /// Bump the item from the station.
    Bump,
    /// Recall the most recently bumped item back to `Ready`.
    Recall,
    /// Void the item - terminal.
    Void,
}

#[derive(Debug, PartialEq, Eq)]
pub struct IllegalTransition {
    pub from: ItemState,
    pub event: ItemEvent,
}

impl std::fmt::Display for IllegalTransition {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "Illegal KDS item transition: {:?} cannot accept {:?}",
            self.from, self.event
        )
    }
}

impl std::error::Error for IllegalTransition {}

/// Pure transition function.
pub fn transition(state: ItemState, event: ItemEvent) -> Result<ItemState, IllegalTransition> {
    use ItemEvent::*;
    use ItemState::*;

    let next = match (state, event) {
        (Held,    Fire)         => Queued,
        (Queued,  StartCooking) => Cooking,
        (Queued,  MarkReady)    => Ready,   // straight-to-ready for cold items
        (Queued,  Bump)         => Bumped,  // quick-serve: skip cooking state
        (Cooking, MarkReady)    => Ready,
        (Cooking, Bump)         => Bumped,
        (Ready,   Bump)         => Bumped,
        (Bumped,  Recall)       => Ready,
        // Void is allowed from any non-terminal state
        (Held,    Void) | (Queued,  Void) | (Cooking, Void) | (Ready, Void) => Voided,
        _ => return Err(IllegalTransition { from: state, event }),
    };
    Ok(next)
}

/// A KDS item with its current state plus a monotonic version counter.
/// `version` increments on every successful transition; idempotent
/// callers (e.g. duplicate bump-bar taps) use it to dedupe.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ItemRecord {
    pub state: ItemState,
    pub version: u64,
}

impl ItemRecord {
    pub fn new_queued() -> Self {
        Self { state: ItemState::Queued, version: 0 }
    }

    pub fn new_held() -> Self {
        Self { state: ItemState::Held, version: 0 }
    }

    /// Apply an event. Returns the new state on success, leaves the
    /// record unchanged on illegal transitions.
    pub fn apply(&mut self, event: ItemEvent) -> Result<ItemState, IllegalTransition> {
        let next = transition(self.state, event)?;
        self.state = next;
        self.version += 1;
        Ok(next)
    }

    /// Idempotent apply: if `expected_version` does not match the
    /// current version, the event is ignored. Use this when the
    /// caller can re-deliver the same event (bump bar, retried IPC).
    pub fn apply_if_version(
        &mut self,
        event: ItemEvent,
        expected_version: u64,
    ) -> Result<ItemState, IllegalTransition> {
        if expected_version != self.version {
            return Ok(self.state);
        }
        self.apply(event)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use ItemEvent::*;
    use ItemState::*;

    #[test]
    fn happy_path_queued_cooking_ready_bumped() {
        assert_eq!(transition(Queued, StartCooking).unwrap(), Cooking);
        assert_eq!(transition(Cooking, MarkReady).unwrap(), Ready);
        assert_eq!(transition(Ready, Bump).unwrap(), Bumped);
    }

    #[test]
    fn held_must_be_fired_before_queued() {
        assert_eq!(transition(Held, Fire).unwrap(), Queued);
        assert!(transition(Held, StartCooking).is_err());
        assert!(transition(Held, Bump).is_err());
    }

    #[test]
    fn cold_items_can_skip_cooking_and_go_straight_to_ready() {
        // Salads, drinks, etc - fire-and-forget from queue to ready
        assert_eq!(transition(Queued, MarkReady).unwrap(), Ready);
    }

    #[test]
    fn quick_serve_can_bump_directly_from_queue_or_cooking() {
        assert_eq!(transition(Queued, Bump).unwrap(), Bumped);
        assert_eq!(transition(Cooking, Bump).unwrap(), Bumped);
    }

    #[test]
    fn recall_only_works_from_bumped() {
        assert_eq!(transition(Bumped, Recall).unwrap(), Ready);
        assert!(transition(Ready, Recall).is_err());
        assert!(transition(Queued, Recall).is_err());
    }

    #[test]
    fn void_allowed_from_any_non_terminal_state() {
        assert_eq!(transition(Held, Void).unwrap(), Voided);
        assert_eq!(transition(Queued, Void).unwrap(), Voided);
        assert_eq!(transition(Cooking, Void).unwrap(), Voided);
        assert_eq!(transition(Ready, Void).unwrap(), Voided);
    }

    #[test]
    fn void_not_allowed_from_terminal_states() {
        assert!(transition(Voided, Void).is_err());
        // Bumped is reachable-from-recall, so voiding a bumped item is
        // not a state-machine concern - refunds happen elsewhere.
        assert!(transition(Bumped, Void).is_err());
    }

    #[test]
    fn no_backwards_moves_except_recall() {
        assert!(transition(Cooking, Fire).is_err());
        assert!(transition(Ready, StartCooking).is_err());
        assert!(transition(Bumped, StartCooking).is_err());
        assert!(transition(Bumped, Fire).is_err());
    }

    #[test]
    fn record_increments_version_on_each_transition() {
        let mut r = ItemRecord::new_queued();
        assert_eq!(r.version, 0);
        r.apply(StartCooking).unwrap();
        assert_eq!(r.version, 1);
        r.apply(MarkReady).unwrap();
        assert_eq!(r.version, 2);
        r.apply(Bump).unwrap();
        assert_eq!(r.version, 3);
        assert_eq!(r.state, Bumped);
    }

    #[test]
    fn illegal_transition_does_not_advance_version() {
        let mut r = ItemRecord::new_queued();
        let res = r.apply(Recall);
        assert!(res.is_err());
        assert_eq!(r.version, 0);
        assert_eq!(r.state, Queued);
    }

    #[test]
    fn apply_if_version_is_idempotent_for_stale_callers() {
        let mut r = ItemRecord::new_queued();
        r.apply(StartCooking).unwrap();
        // Caller still thinks we're at version 0 - should no-op
        let res = r.apply_if_version(StartCooking, 0).unwrap();
        assert_eq!(res, Cooking); // returns current state
        assert_eq!(r.version, 1); // unchanged
    }

    #[test]
    fn apply_if_version_advances_when_version_matches() {
        let mut r = ItemRecord::new_queued();
        r.apply(StartCooking).unwrap();
        let res = r.apply_if_version(MarkReady, 1).unwrap();
        assert_eq!(res, Ready);
        assert_eq!(r.version, 2);
    }

    #[test]
    fn new_held_starts_in_held_state() {
        let r = ItemRecord::new_held();
        assert_eq!(r.state, Held);
        assert_eq!(r.version, 0);
    }
}
