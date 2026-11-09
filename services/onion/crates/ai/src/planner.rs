//! Decision planning over scorer output: fills, checks, clicks, skips.
//! Mirrors `internal/ai/classifier/planner`.

use crate::model::{Example, Model, ModelError};
use thiserror::Error;

// Actions a field decision can carry.
pub const ACTION_FILL: &str = "fill";
pub const ACTION_CHECK: &str = "check";
pub const ACTION_CLICK: &str = "click";
pub const ACTION_SKIP: &str = "skip";

/// Minimum probability a decision needs before it leaves review.
pub const DEFAULT_THRESHOLD: f64 = 0.5;

/// One form element to fill.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Field {
    pub id: String,
    pub label: String,
    pub role: String,
    pub value: String,
}

/// One planning job: a form, its fields, and the document entities
/// extracted upstream.
#[derive(Debug, Clone, Default)]
pub struct Request {
    pub form: String,
    pub fields: Vec<Field>,
    pub entities: Vec<Entity>,
    pub threshold: f64,
    pub allow_submit: bool,
}

/// The plan for a single field.
#[derive(Debug, Clone, PartialEq)]
pub struct Decision {
    pub field_index: usize,
    pub field_id: String,
    pub field_label: String,
    pub action: String,
    pub value: String,
    pub option_index: i64,
    pub probability: f64,
    pub needs_review: bool,
    pub already_set: bool,
}

/// An ordered, reviewable set of field decisions: fills, then checks,
/// then any approved submit click, then skips.
#[derive(Debug, Clone, Default)]
pub struct Plan {
    pub decisions: Vec<Decision>,
    pub submit: Option<Decision>,
    pub entities: usize,
}

const FIXED_ACTIONS: [&str; 3] = [ACTION_CHECK, ACTION_CLICK, ACTION_SKIP];

#[derive(Debug, Error, PartialEq)]
pub enum PlannerError {
    #[error("{0}")]
    Invalid(String),
    #[error("{0}")]
    Model(#[from] ModelError),
}

/// Scoring layer: one probability per option.
pub trait Scorer {
    /// Batch logits when the scorer supports them (Go `batchScorer`).
    fn logits_batch(&self, _exs: &[Example]) -> Option<Result<Vec<Vec<f32>>, PlannerError>> {
        None
    }
    fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), PlannerError>;
}

impl Scorer for Model {
    fn logits_batch(&self, exs: &[Example]) -> Option<Result<Vec<Vec<f32>>, PlannerError>> {
        Some(self.logits_batch(exs).map_err(PlannerError::Model))
    }

    fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), PlannerError> {
        self.score(ex).map_err(PlannerError::Model)
    }
}

/// Score every field against the candidate entities and return an
/// ordered plan without mutating anything.
pub fn build(scorer: &dyn Scorer, req: &Request) -> Result<Plan, PlannerError> {
    let threshold = threshold_or(req.threshold);
    let mut plan = Plan {
        entities: req.entities.len(),
        ..Default::default()
    };
    if req.entities.is_empty() {
        for (i, field) in req.fields.iter().enumerate() {
            plan.decisions.push(skip_decision(i, field, true));
        }
        return Ok(plan);
    }
    let mut examples = Vec::with_capacity(req.fields.len());
    for field in &req.fields {
        examples.push(Example {
            context: element_context(&req.form, field),
            options: options_for(&req.entities),
        });
    }
    let rows = score_all(scorer, &examples)?;
    if rows.len() != req.fields.len() {
        return Err(PlannerError::Invalid(format!(
            "planner: scorer returned {} rows for {} fields",
            rows.len(),
            req.fields.len()
        )));
    }
    let mut decisions = Vec::with_capacity(req.fields.len());
    for (i, field) in req.fields.iter().enumerate() {
        let probs = &rows[i];
        if probs.is_empty() {
            decisions.push(skip_decision(i, field, true));
            continue;
        }
        let mut best = 0;
        for (j, prob) in probs.iter().enumerate() {
            if *prob > probs[best] {
                best = j;
            }
        }
        decisions.push(decide(i, field, &req.entities, probs, best, threshold));
    }
    let submit_index = select_submit(&mut decisions, req);
    plan.decisions = order_decisions(decisions);
    if let Some(index) = submit_index {
        for decision in &plan.decisions {
            if decision.field_index == index {
                plan.submit = Some(decision.clone());
                break;
            }
        }
    }
    Ok(plan)
}

fn decide(
    index: usize,
    field: &Field,
    entities: &[Entity],
    probs: &[f64],
    best: usize,
    threshold: f64,
) -> Decision {
    if probs.is_empty() {
        return skip_decision(index, field, true);
    }
    let mut decision = Decision {
        field_index: index,
        field_id: field.id.clone(),
        field_label: field.label.clone(),
        action: String::new(),
        value: String::new(),
        option_index: best as i64,
        probability: probs[best],
        needs_review: false,
        already_set: false,
    };
    let n = entities.len();
    match best.cmp(&n) {
        std::cmp::Ordering::Less => {
            decision.action = ACTION_FILL.to_string();
            decision.value = entities[best].value.clone();
            if !decision.value.is_empty()
                && decision
                    .value
                    .trim()
                    .eq_ignore_ascii_case(field.value.trim())
            {
                decision.action = ACTION_SKIP.to_string();
                decision.already_set = true;
                return decision;
            }
        }
        std::cmp::Ordering::Equal => decision.action = ACTION_CHECK.to_string(),
        _ => {
            decision.action = if best == n + 1 {
                ACTION_CLICK.to_string()
            } else {
                ACTION_SKIP.to_string()
            };
        }
    }
    if decision.probability < threshold {
        decision.action = ACTION_SKIP.to_string();
        decision.needs_review = true;
    }
    decision
}

fn select_submit(decisions: &mut [Decision], req: &Request) -> Option<usize> {
    let mut best_index: Option<usize> = None;
    for i in 0..decisions.len() {
        if decisions[i].action != ACTION_CLICK {
            continue;
        }
        if !is_submit_field(&req.fields[decisions[i].field_index])
            || !req.allow_submit
            || decisions[i].probability < threshold_or(req.threshold)
        {
            demote_click(&mut decisions[i]);
            continue;
        }
        if best_index.is_none_or(|best| decisions[i].probability > decisions[best].probability) {
            best_index = Some(i);
        }
    }
    let best = best_index?;
    for (i, decision) in decisions.iter_mut().enumerate() {
        if decision.action == ACTION_CLICK && i != best {
            demote_click(decision);
        }
    }
    Some(decisions[best].field_index)
}

/// Turn a disallowed or superseded click into a review skip and clear
/// the stale option metadata so clients cannot read "skip (p=0.9)".
fn demote_click(decision: &mut Decision) {
    decision.action = ACTION_SKIP.to_string();
    decision.option_index = -1;
    decision.probability = 0.0;
    decision.needs_review = true;
}

fn threshold_or(threshold: f64) -> f64 {
    if threshold <= 0.0 {
        DEFAULT_THRESHOLD
    } else if threshold > 1.0 {
        1.0
    } else {
        threshold
    }
}

fn is_submit_field(field: &Field) -> bool {
    let role = field.role.trim().to_lowercase();
    if !role.is_empty() && role != "button" && role != "axbutton" {
        return false;
    }
    let label = normalize_label(&field.label);
    label == "submit" || label == "submit form"
}

fn normalize_label(text: &str) -> String {
    let lowered = text.trim().to_lowercase();
    let lowered = lowered.strip_suffix(':').unwrap_or(&lowered);
    lowered.split_whitespace().collect::<Vec<&str>>().join(" ")
}

fn order_decisions(mut decisions: Vec<Decision>) -> Vec<Decision> {
    fn rank(action: &str) -> i32 {
        match action {
            ACTION_FILL => 0,
            ACTION_CHECK => 1,
            ACTION_CLICK => 2,
            _ => 3,
        }
    }
    decisions.sort_by_key(|decision| rank(&decision.action));
    decisions
}

fn skip_decision(index: usize, field: &Field, needs_review: bool) -> Decision {
    Decision {
        field_index: index,
        field_id: field.id.clone(),
        field_label: field.label.clone(),
        action: ACTION_SKIP.to_string(),
        value: String::new(),
        option_index: -1,
        probability: 0.0,
        needs_review,
        already_set: false,
    }
}

fn options_for(entities: &[Entity]) -> Vec<String> {
    let mut out = Vec::with_capacity(entities.len() + FIXED_ACTIONS.len());
    for entity in entities {
        out.push(format!("fill {}: {}", entity.label, entity.value));
    }
    out.extend(FIXED_ACTIONS.iter().map(|action| (*action).to_string()));
    out
}

fn element_context(form: &str, field: &Field) -> String {
    let role = field.role.trim();
    let role = if role.is_empty() { "Edit" } else { role };
    let mut name = normalize_form(form);
    if name.is_empty() {
        name = "Form".to_string();
    }
    format!(
        "TASK fill the form from the document, then submit\nFORM {}\nELEMENT {} \"{}\" value=\"{}\"",
        clean_field(&name),
        clean_field(role),
        clean_field(&field.label),
        clean_field(&field.value),
    )
}

fn clean_field(text: &str) -> String {
    let replaced: String = text
        .chars()
        .map(|ch| match ch {
            '"' => '\'',
            '\n' | '\r' | '\t' => ' ',
            other => other,
        })
        .collect();
    replaced.split_whitespace().collect::<Vec<&str>>().join(" ")
}

fn score_all(scorer: &dyn Scorer, examples: &[Example]) -> Result<Vec<Vec<f64>>, PlannerError> {
    if let Some(batch) = scorer.logits_batch(examples) {
        let logits = batch?;
        let mut rows = Vec::with_capacity(logits.len());
        for row in logits {
            rows.push(softmax(&row));
        }
        return Ok(rows);
    }
    let mut rows = Vec::with_capacity(examples.len());
    for ex in examples {
        let (probs, _) = scorer.score(ex)?;
        rows.push(probs);
    }
    Ok(rows)
}

fn softmax(logits: &[f32]) -> Vec<f64> {
    if logits.is_empty() {
        return Vec::new();
    }
    let mut max_logit = logits[0];
    for value in logits.iter().skip(1) {
        if *value > max_logit {
            max_logit = *value;
        }
    }
    let mut out = Vec::with_capacity(logits.len());
    let mut sum = 0.0f64;
    for logit in logits {
        let mut d = f64::from(*logit - max_logit);
        if d.is_nan() {
            d = 0.0;
        }
        let e = d.exp();
        out.push(e);
        sum += e;
    }
    if sum == 0.0 {
        let uniform = 1.0 / out.len() as f64;
        for value in out.iter_mut() {
            *value = uniform;
        }
        return out;
    }
    for value in out.iter_mut() {
        *value /= sum;
    }
    out
}

/// The deterministic fallback used when no classifier checkpoint is
/// configured. It scores one example at a time.
#[derive(Debug, Clone, Copy, Default)]
pub struct LexicalScorer;

const LEXICAL_TEMPERATURE: f64 = 0.35;

impl Scorer for LexicalScorer {
    fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), PlannerError> {
        let label = element_label(&ex.context);
        let scores: Vec<f64> = ex
            .options
            .iter()
            .map(|option| lexical_score(&label, option))
            .collect();
        let mut best = 0;
        for (i, score) in scores.iter().enumerate() {
            if *score > scores[best] {
                best = i;
            }
        }
        let mut probs: Vec<f64> = scores
            .iter()
            .map(|score| (score / LEXICAL_TEMPERATURE).exp())
            .collect();
        let sum: f64 = probs.iter().sum();
        for prob in probs.iter_mut() {
            *prob /= sum;
        }
        Ok((probs, best))
    }
}

fn lexical_score(field_label: &str, option: &str) -> f64 {
    match option {
        ACTION_CHECK => 1.2 * boolean_field_prior(field_label),
        ACTION_CLICK => {
            if is_submit_label(field_label) {
                1.2
            } else {
                0.05
            }
        }
        ACTION_SKIP => 0.4,
        _ => {
            if !option.starts_with("fill ") {
                return 0.0;
            }
            let rest = option.strip_prefix("fill ").unwrap_or("");
            let Some((label, _)) = rest.split_once(':') else {
                return 0.0;
            };
            label_similarity(field_label, label)
        }
    }
}

fn label_similarity(a: &str, b: &str) -> f64 {
    let na = normalize_prompt_label(a);
    let nb = normalize_prompt_label(b);
    if na.is_empty() || nb.is_empty() {
        return 0.0;
    }
    if na == nb {
        return 1.6;
    }
    if na.contains(&nb) || nb.contains(&na) {
        return 1.0;
    }
    let tokens_a: Vec<&str> = na.split_whitespace().collect();
    let tokens_b: Vec<&str> = nb.split_whitespace().collect();
    if tokens_a.is_empty() || tokens_b.is_empty() {
        return 0.0;
    }
    let mut shared = 0usize;
    for ta in &tokens_a {
        if tokens_b.iter().any(|tb| tb == ta) {
            shared += 1;
        }
    }
    if shared == 0 {
        return 0.0;
    }
    let overlap = shared as f64 / tokens_a.len() as f64;
    0.7 * overlap
}

fn boolean_field_prior(label: &str) -> f64 {
    let n = normalize_prompt_label(label);
    for token in n.split_whitespace() {
        match token {
            "yes" | "no" | "has" | "is" | "active" | "agree" | "agreed" | "consent"
            | "subscribe" | "enable" | "enabled" | "include" | "applicable" | "checked"
            | "unchecked" | "already" | "same" => return 1.0,
            _ => {}
        }
    }
    for phrase in ["opt in", "opt out"] {
        if n.contains(phrase) {
            return 1.0;
        }
    }
    0.15
}

fn is_submit_label(label: &str) -> bool {
    let normalized = normalize_label(label);
    normalized == "submit" || normalized == "submit form"
}

fn normalize_prompt_label(text: &str) -> String {
    let replaced: String = text
        .trim()
        .to_lowercase()
        .chars()
        .map(|ch| match ch {
            '_' | '-' | '.' | ',' => ' ',
            other => other,
        })
        .collect();
    replaced.split_whitespace().collect::<Vec<&str>>().join(" ")
}

fn element_label(context: &str) -> String {
    for line in context.lines() {
        let line = line.trim();
        let Some(rest) = line.strip_prefix("ELEMENT ") else {
            continue;
        };
        if let Some(start) = rest.find(" \"") {
            let remainder = &rest[start + 2..];
            if let Some(end) = remainder.find('"') {
                if end > 0 {
                    return remainder[..end].to_string();
                }
            }
        }
        if let Some((_, after)) = rest.split_once(' ') {
            let (label, _) = after.split_once(" value=").unwrap_or((after, ""));
            return label.trim().to_string();
        }
        return rest.trim().to_string();
    }
    String::new()
}

/// One "Label: value" pair extracted from a source document.
#[derive(Debug, Clone, PartialEq)]
pub struct Entity {
    pub label: String,
    pub value: String,
}

/// Caps how many candidates travel into scoring.
pub const EXTRACT_ENTITY_LIMIT: usize = 128;

const EXTRACT_LABEL_BYTES: usize = 80;
const EXTRACT_VALUE_BYTES: usize = 200;
const FIELD_LABEL_BYTES: usize = 80;
const FIELD_VALUE_BYTES: usize = 200;
const FIELD_ROLE_BYTES: usize = 32;
const FIELD_ID_BYTES: usize = 64;
const FORM_NAME_BYTES: usize = 200;

/// Parse "Label: value" lines, strip bullet prefixes, deduplicate
/// pairs, and cap the result at `limit` (or `EXTRACT_ENTITY_LIMIT`
/// when `limit <= 0`).
pub fn extract_entities(document: &str, limit: usize) -> Vec<Entity> {
    let limit = if limit == 0 {
        EXTRACT_ENTITY_LIMIT
    } else {
        limit
    };
    let mut seen = std::collections::HashSet::new();
    let mut out = Vec::new();
    for raw_line in document.split('\n') {
        let line = raw_line.strip_suffix('\r').unwrap_or(raw_line);
        let line = line.trim();
        let line = line.trim_start_matches(['-', '*', '•']);
        let line = line.trim();
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        let Some((label_raw, value_raw)) = line.split_once(':') else {
            continue;
        };
        let label = trim_to_bytes(label_raw.trim(), EXTRACT_LABEL_BYTES);
        let value = trim_to_bytes(value_raw.trim(), EXTRACT_VALUE_BYTES);
        if is_url_noise(&label, &value) {
            continue;
        }
        if label.is_empty() || value.is_empty() || !has_letter(&label) {
            continue;
        }
        let key = format!("{}\0{}", label.to_lowercase(), value.to_lowercase());
        if !seen.insert(key) {
            continue;
        }
        out.push(Entity { label, value });
        if out.len() >= limit {
            break;
        }
    }
    out
}

pub fn normalize_entity(entity: Entity) -> Option<Entity> {
    let entity = Entity {
        label: trim_to_bytes(entity.label.trim(), EXTRACT_LABEL_BYTES),
        value: trim_to_bytes(entity.value.trim(), EXTRACT_VALUE_BYTES),
    };
    if is_url_noise(&entity.label, &entity.value) {
        return None;
    }
    if entity.label.is_empty() || entity.value.is_empty() || !has_letter(&entity.label) {
        return None;
    }
    Some(entity)
}

pub fn normalize_form(name: &str) -> String {
    trim_to_bytes(name.trim(), FORM_NAME_BYTES)
}

pub fn normalize_field(field: Field) -> Option<Field> {
    let field = Field {
        id: trim_to_bytes(field.id.trim(), FIELD_ID_BYTES),
        label: trim_to_bytes(field.label.trim(), FIELD_LABEL_BYTES),
        role: trim_to_bytes(field.role.trim(), FIELD_ROLE_BYTES),
        value: trim_to_bytes(field.value.trim(), FIELD_VALUE_BYTES),
    };
    if field.label.is_empty() {
        return None;
    }
    Some(field)
}

/// Report whether a label/value pair is a link fragment or prose that
/// embeds a URL. A clean URL value (no surrounding prose) is allowed.
pub fn is_url_noise(label: &str, value: &str) -> bool {
    let lower = label.to_lowercase();
    if value.starts_with("//") || lower.ends_with("http") || lower.ends_with("https") {
        return true;
    }
    value.contains("://") && value.chars().any(|c| c == ' ' || c == '\t')
}

fn trim_to_bytes(text: &str, limit: usize) -> String {
    if text.len() <= limit {
        return text.to_string();
    }
    let mut cut = limit;
    while cut > 0 && !text.is_char_boundary(cut) {
        cut -= 1;
    }
    text[..cut].trim().to_string()
}

fn has_letter(text: &str) -> bool {
    text.chars().any(|ch| ch.is_alphabetic())
}

#[cfg(test)]
mod tests {
    use super::{
        extract_entities, normalize_field, Field, LexicalScorer, Request, Scorer, ACTION_SKIP,
    };

    #[test]
    fn extract_and_normalize() {
        let entities = extract_entities(
            "- Full name: Jane Doe\n# hi\nLink: https://x.example and more\nCity: Paris",
            0,
        );
        assert_eq!(entities.len(), 2);
        assert_eq!(entities[0].label, "Full name");
        let field = Field {
            id: " f ".into(),
            label: " Name ".into(),
            role: " ".into(),
            value: " ".into(),
        };
        let normalized = normalize_field(field).unwrap_or_default();
        assert_eq!(normalized.id, "f");
    }

    #[test]
    fn empty_entities_skip_all() {
        let req = Request {
            form: "F".into(),
            fields: vec![Field {
                id: "1".into(),
                label: "L".into(),
                role: "Edit".into(),
                value: String::new(),
            }],
            ..Default::default()
        };
        let plan = super::build(&LexicalScorer, &req).unwrap_or_default();
        assert_eq!(plan.decisions.len(), 1);
        assert_eq!(plan.decisions[0].action, ACTION_SKIP);
    }

    #[test]
    fn lexical_scores() {
        let scorer = LexicalScorer;
        let (probs, _) = scorer
            .score(&crate::model::Example {
                context: "TASK x\nFORM F\nELEMENT Edit \"Name\" value=\"\"".into(),
                options: vec!["fill Name: Jane".into(), "skip".into()],
            })
            .unwrap_or_default();
        assert!(probs.len() == 2);
    }
}
