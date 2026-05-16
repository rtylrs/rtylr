//! Expo-screen order assembly.
//!
//! When the router splits one order across multiple stations, expo
//! needs to know when every station has marked its portion ready so
//! the order can be released to the guest/runner. This module owns the
//! aggregation: take each station's current state for an order, decide
//! whether the order is ready as a whole, and never release a partial
//! order by accident.
//!
//! Pure logic — no I/O, no timers. Wire it to the item state machine
//! by passing each station's worst-state-among-its-items in.

use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

use crate::hardware::kds::state::ItemState;

/// What expo thinks one station's portion of an order looks like.
///
/// `worst` is the least-advanced state among the items routed to this
/// station. An order is releasable only when every station's `worst`
/// is `Ready` or `Bumped`.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct StationProgress {
    pub station: String,
    pub worst: ItemState,
    pub total_items: u32,
    pub ready_or_bumped: u32,
}

impl StationProgress {
    pub fn is_complete(&self) -> bool {
        matches!(self.worst, ItemState::Ready | ItemState::Bumped)
            && self.ready_or_bumped == self.total_items
    }
}

/// Roll up an iterator of items at a station into one `StationProgress`.
///
/// `worst` follows the natural ItemState ordering for kitchen progress:
/// Held < Queued < Cooking < Ready < Bumped, with `Voided` ignored (a
/// voided item is no longer cooking, so it doesn't block expo).
pub fn aggregate_station<I: IntoIterator<Item = ItemState>>(
    station: impl Into<String>,
    states: I,
) -> StationProgress {
    let mut total: u32 = 0;
    let mut ready_or_bumped: u32 = 0;
    let mut worst: Option<ItemState> = None;

    for s in states {
        if s == ItemState::Voided {
            continue;
        }
        total += 1;
        if matches!(s, ItemState::Ready | ItemState::Bumped) {
            ready_or_bumped += 1;
        }
        worst = Some(match worst {
            None => s,
            Some(prev) => min_state(prev, s),
        });
    }

    StationProgress {
        station: station.into(),
        worst: worst.unwrap_or(ItemState::Ready),
        total_items: total,
        ready_or_bumped,
    }
}

fn rank(s: ItemState) -> u8 {
    match s {
        ItemState::Held    => 0,
        ItemState::Queued  => 1,
        ItemState::Cooking => 2,
        ItemState::Ready   => 3,
        ItemState::Bumped  => 4,
        ItemState::Voided  => 5, // unreachable in aggregate_station
    }
}

fn min_state(a: ItemState, b: ItemState) -> ItemState {
    if rank(a) <= rank(b) { a } else { b }
}

/// One order viewed across every station it touches.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct OrderAssembly {
    pub order_number: String,
    pub stations: BTreeMap<String, StationProgress>,
}

impl OrderAssembly {
    pub fn new(order_number: impl Into<String>) -> Self {
        Self { order_number: order_number.into(), stations: BTreeMap::new() }
    }

    pub fn upsert(&mut self, progress: StationProgress) {
        self.stations.insert(progress.station.clone(), progress);
    }

    /// True only when every station's portion is complete. An empty
    /// assembly is *not* complete — releasing a no-station order would
    /// hand off nothing.
    pub fn is_releasable(&self) -> bool {
        !self.stations.is_empty()
            && self.stations.values().all(StationProgress::is_complete)
    }

    /// Stations that still have work to do.
    pub fn pending_stations(&self) -> Vec<&str> {
        self.stations
            .values()
            .filter(|p| !p.is_complete())
            .map(|p| p.station.as_str())
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use ItemState::*;

    #[test]
    fn aggregate_station_picks_least_advanced_state_as_worst() {
        let p = aggregate_station("grill", vec![Ready, Cooking, Ready]);
        assert_eq!(p.worst, Cooking);
        assert_eq!(p.total_items, 3);
        assert_eq!(p.ready_or_bumped, 2);
    }

    #[test]
    fn aggregate_station_ignores_voided_items() {
        let p = aggregate_station("grill", vec![Ready, Voided, Ready]);
        assert_eq!(p.total_items, 2);
        assert_eq!(p.ready_or_bumped, 2);
        assert_eq!(p.worst, Ready);
        assert!(p.is_complete());
    }

    #[test]
    fn aggregate_station_is_complete_when_all_ready_or_bumped() {
        let p = aggregate_station("fry", vec![Ready, Bumped, Ready]);
        assert!(p.is_complete());
    }

    #[test]
    fn aggregate_station_not_complete_with_anything_below_ready() {
        let p = aggregate_station("fry", vec![Ready, Cooking]);
        assert!(!p.is_complete());
    }

    #[test]
    fn aggregate_station_with_only_voided_items_is_complete() {
        // No active work outstanding — expo isn't blocked by a fully
        // voided station portion.
        let p = aggregate_station("cold", vec![Voided, Voided]);
        assert_eq!(p.total_items, 0);
        assert!(p.is_complete());
    }

    #[test]
    fn order_assembly_not_releasable_when_empty() {
        let a = OrderAssembly::new("T1");
        assert!(!a.is_releasable());
    }

    #[test]
    fn order_assembly_releasable_only_when_every_station_ready() {
        let mut a = OrderAssembly::new("T1");
        a.upsert(aggregate_station("grill", vec![Ready, Ready]));
        a.upsert(aggregate_station("fry", vec![Cooking]));
        assert!(!a.is_releasable());
        assert_eq!(a.pending_stations(), vec!["fry"]);

        a.upsert(aggregate_station("fry", vec![Ready]));
        assert!(a.is_releasable());
        assert!(a.pending_stations().is_empty());
    }

    #[test]
    fn order_assembly_upsert_replaces_station_progress() {
        let mut a = OrderAssembly::new("T1");
        a.upsert(aggregate_station("grill", vec![Cooking]));
        a.upsert(aggregate_station("grill", vec![Ready]));
        assert_eq!(a.stations.len(), 1);
        assert!(a.is_releasable());
    }

    #[test]
    fn pending_stations_sorted_by_name() {
        let mut a = OrderAssembly::new("T1");
        a.upsert(aggregate_station("fry", vec![Cooking]));
        a.upsert(aggregate_station("grill", vec![Cooking]));
        a.upsert(aggregate_station("cold", vec![Cooking]));
        // BTreeMap iteration is sorted, so we get a stable order
        assert_eq!(a.pending_stations(), vec!["cold", "fry", "grill"]);
    }

    #[test]
    fn order_with_one_bumped_station_and_one_ready_station_is_releasable() {
        // Expo handing off an already-bumped portion to the runner +
        // a freshly-plated portion shouldn't block release.
        let mut a = OrderAssembly::new("T1");
        a.upsert(aggregate_station("grill", vec![Bumped, Bumped]));
        a.upsert(aggregate_station("cold", vec![Ready]));
        assert!(a.is_releasable());
    }
}
