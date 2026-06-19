-- Phase 3 — custom application form fields per role + submission answers.
--
-- The careers page always captures name / email / phone / resume / cover
-- letter. These are the defaults. Anything beyond that (portfolio URL,
-- years of experience, select/multiselect, boolean consent) is a custom
-- field the HR admin adds per role.
--
-- `career_id` scopes fields to their opening so two roles can ask
-- different things. `options_json` holds the choices for single/multi
-- selects as a JSON array of strings. `kind` is opaque (VARCHAR) for
-- forward-compat.
--
-- The `answers` column on `careers_submissions` is a JSON map keyed by
-- field id → answer (string or string[]). Denormalising onto the
-- submission row means listing 500 submissions stays one query, not
-- one-plus-the-answers-per-row.

CREATE TABLE IF NOT EXISTS careers_form_fields (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    career_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    kind VARCHAR(32) NOT NULL,
    label VARCHAR(240) NOT NULL,
    placeholder VARCHAR(240) NULL,
    required BOOLEAN NOT NULL DEFAULT false,
    options_json JSON NULL,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_careers_form_fields_career (career_id),
    INDEX idx_careers_form_fields_owner (owner_id)
);

ALTER TABLE careers_submissions ADD COLUMN answers_json JSON NULL;
