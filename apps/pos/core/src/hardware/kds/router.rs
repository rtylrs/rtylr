//! Station routing for KDS tickets.
//!
//! A POS order typically contains items from multiple stations (grill,
//! fry, cold, expo). The router takes one incoming `KdsTicket` and
//! produces N `StationTicket`s — one per station the order touches —
//! all sharing the parent order number so expo can re-assemble them.
//!
//! Rule lookup is deterministic and total: every item resolves to
//! exactly one station, with `expo` (or whatever the configured
//! fallback is) as the catch-all. This is intentional — failing to
//! route an item silently is far worse than over-routing.

use serde::{Deserialize, Serialize};

use crate::hardware::kds::formatter::{KdsItem, KdsTicket};

/// A station identifier (e.g. "grill", "fry", "cold", "expo").
pub type Station = String;

/// What field on an item the rule matches against.
#[derive(Clone, Debug, Deserialize, Serialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum RuleKey {
    /// Match on a normalized item name (lowercase, trimmed). Exact match.
    ItemName(String),
    /// Match if the item's name *contains* this substring (case-insensitive).
    ItemNameContains(String),
    /// Match on category tag passed alongside the item. (See `RouterInput`.)
    Category(String),
}

#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct RouterRule {
    pub key: RuleKey,
    pub station: Station,
}

/// Routing configuration. Rules are evaluated in order; the first
/// match wins. If no rule matches, `fallback_station` is used.
#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct RouterConfig {
    pub rules: Vec<RouterRule>,
    pub fallback_station: Station,
}

impl RouterConfig {
    pub fn new<S: Into<String>>(fallback_station: S) -> Self {
        Self { rules: Vec::new(), fallback_station: fallback_station.into() }
    }

    pub fn with_rule(mut self, key: RuleKey, station: impl Into<String>) -> Self {
        self.rules.push(RouterRule { key, station: station.into() });
        self
    }
}

/// One item paired with its optional category tag. The category isn't
/// on `KdsItem` today; callers that want category-based routing supply
/// it via this wrapper rather than baking it into the formatter type.
#[derive(Clone, Debug)]
pub struct RoutableItem<'a> {
    pub item: &'a KdsItem,
    pub category: Option<&'a str>,
}

/// A ticket destined for a single station. `parent_order_number` lets
/// expo correlate all sibling tickets back to the original order.
#[derive(Clone, Debug)]
pub struct StationTicket {
    pub station: Station,
    pub parent_order_number: String,
    pub items: Vec<KdsItem>,
}

/// Resolve a single item to its station.
pub fn route_item(config: &RouterConfig, item: &RoutableItem<'_>) -> Station {
    for rule in &config.rules {
        if rule_matches(&rule.key, item) {
            return rule.station.clone();
        }
    }
    config.fallback_station.clone()
}

fn rule_matches(key: &RuleKey, item: &RoutableItem<'_>) -> bool {
    match key {
        RuleKey::ItemName(name) => normalize(&item.item.name) == normalize(name),
        RuleKey::ItemNameContains(needle) => {
            normalize(&item.item.name).contains(&normalize(needle))
        }
        RuleKey::Category(cat) => item
            .category
            .map(|c| normalize(c) == normalize(cat))
            .unwrap_or(false),
    }
}

fn normalize(s: &str) -> String {
    s.trim().to_lowercase()
}

/// Split a ticket into per-station tickets. Items are grouped into the
/// returned vec in the order their stations were first encountered, so
/// the result is deterministic for a given input + config.
pub fn route_ticket(
    config: &RouterConfig,
    ticket: &KdsTicket,
    categories: &[Option<String>],
) -> Vec<StationTicket> {
    let mut groups: Vec<(Station, Vec<KdsItem>)> = Vec::new();

    for (i, item) in ticket.items.iter().enumerate() {
        let category = categories.get(i).and_then(|c| c.as_deref());
        let station = route_item(config, &RoutableItem { item, category });

        if let Some(group) = groups.iter_mut().find(|(s, _)| s == &station) {
            group.1.push(item.clone());
        } else {
            groups.push((station, vec![item.clone()]));
        }
    }

    groups
        .into_iter()
        .map(|(station, items)| StationTicket {
            station,
            parent_order_number: ticket.order_number.clone(),
            items,
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::hardware::kds::formatter::KdsItem;

    fn item(name: &str) -> KdsItem {
        KdsItem {
            name: name.to_string(),
            quantity: 1.0,
            notes: None,
            modifiers: vec![],
        }
    }

    fn ticket(items: Vec<KdsItem>) -> KdsTicket {
        KdsTicket {
            order_number: "T1".to_string(),
            station: None,
            order_type: None,
            table_number: None,
            created_at: None,
            items,
        }
    }

    fn cfg() -> RouterConfig {
        RouterConfig::new("expo")
            .with_rule(RuleKey::ItemNameContains("burger".to_string()), "grill")
            .with_rule(RuleKey::ItemNameContains("fries".to_string()), "fry")
            .with_rule(RuleKey::Category("salad".to_string()), "cold")
    }

    #[test]
    fn route_item_uses_first_matching_rule() {
        let cfg = cfg();
        let i = item("Cheeseburger");
        let r = RoutableItem { item: &i, category: None };
        assert_eq!(route_item(&cfg, &r), "grill");
    }

    #[test]
    fn route_item_falls_back_when_no_rule_matches() {
        let cfg = cfg();
        let i = item("Mystery Box");
        let r = RoutableItem { item: &i, category: None };
        assert_eq!(route_item(&cfg, &r), "expo");
    }

    #[test]
    fn route_item_is_case_insensitive() {
        let cfg = cfg();
        let i = item("BURGER deluxe");
        let r = RoutableItem { item: &i, category: None };
        assert_eq!(route_item(&cfg, &r), "grill");
    }

    #[test]
    fn route_item_matches_category_when_name_misses() {
        let cfg = cfg();
        let i = item("Caesar");
        let r = RoutableItem { item: &i, category: Some("salad") };
        assert_eq!(route_item(&cfg, &r), "cold");
    }

    #[test]
    fn first_match_wins_when_multiple_rules_could_apply() {
        let cfg = RouterConfig::new("expo")
            .with_rule(RuleKey::ItemNameContains("burger".to_string()), "grill")
            .with_rule(RuleKey::ItemNameContains("cheese".to_string()), "cold");
        let i = item("Cheeseburger");
        let r = RoutableItem { item: &i, category: None };
        assert_eq!(route_item(&cfg, &r), "grill");
    }

    #[test]
    fn route_ticket_groups_items_per_station() {
        let cfg = cfg();
        let t = ticket(vec![
            item("Cheeseburger"),
            item("Fries"),
            item("Mystery"),
            item("Bacon Burger"),
        ]);
        let result = route_ticket(&cfg, &t, &[None, None, None, None]);

        assert_eq!(result.len(), 3);
        let grill = result.iter().find(|s| s.station == "grill").unwrap();
        assert_eq!(grill.items.len(), 2);
        assert_eq!(grill.items[0].name, "Cheeseburger");
        assert_eq!(grill.items[1].name, "Bacon Burger");

        let fry = result.iter().find(|s| s.station == "fry").unwrap();
        assert_eq!(fry.items.len(), 1);

        let expo = result.iter().find(|s| s.station == "expo").unwrap();
        assert_eq!(expo.items.len(), 1);
    }

    #[test]
    fn route_ticket_preserves_parent_order_number_on_every_sub_ticket() {
        let cfg = cfg();
        let t = ticket(vec![item("Cheeseburger"), item("Fries")]);
        let result = route_ticket(&cfg, &t, &[None, None]);
        for st in &result {
            assert_eq!(st.parent_order_number, "T1");
        }
    }

    #[test]
    fn route_ticket_uses_category_when_provided() {
        let cfg = cfg();
        let t = ticket(vec![item("Caesar"), item("Cheeseburger")]);
        let cats = vec![Some("salad".to_string()), None];
        let result = route_ticket(&cfg, &t, &cats);

        let stations: Vec<_> = result.iter().map(|s| s.station.as_str()).collect();
        assert!(stations.contains(&"cold"));
        assert!(stations.contains(&"grill"));
    }

    #[test]
    fn route_ticket_with_no_rules_routes_everything_to_fallback() {
        let cfg = RouterConfig::new("expo");
        let t = ticket(vec![item("Anything"), item("Else")]);
        let result = route_ticket(&cfg, &t, &[None, None]);
        assert_eq!(result.len(), 1);
        assert_eq!(result[0].station, "expo");
        assert_eq!(result[0].items.len(), 2);
    }

    #[test]
    fn route_ticket_with_empty_items_returns_empty() {
        let cfg = cfg();
        let t = ticket(vec![]);
        let result = route_ticket(&cfg, &t, &[]);
        assert!(result.is_empty());
    }

    #[test]
    fn route_ticket_is_deterministic_in_station_order() {
        // Stations come out in first-encounter order — important for
        // tests and for UI ordering.
        let cfg = cfg();
        let t = ticket(vec![item("Fries"), item("Cheeseburger"), item("Fries")]);
        let result = route_ticket(&cfg, &t, &[None, None, None]);
        assert_eq!(result[0].station, "fry");
        assert_eq!(result[1].station, "grill");
        assert_eq!(result[0].items.len(), 2);
    }
}
