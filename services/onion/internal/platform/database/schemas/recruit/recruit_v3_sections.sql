-- Phase 2 — content sections on the careers page.
--
-- Each row is one section (perks grid, culture block, testimonial,
-- video embed, FAQ list, values list, CTA). `position` is a sparse
-- integer used to order sections: drag-reorder writes new positions.
-- `kind` is intentionally a VARCHAR so we can add section kinds
-- without a migration — the frontend ignores unknown kinds.
--
-- `body` holds free-form text for simple sections. Structured sections
-- (perks with icons, FAQ with many Q/A pairs, values with labels) put
-- their structured payload in `items_json` to avoid a second table.
-- Trade-off: no relational queries across items, but sections are
-- loaded in bulk per page and rendered as opaque blobs, so that's fine.

CREATE TABLE IF NOT EXISTS careers_sections (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    kind VARCHAR(32) NOT NULL,
    title VARCHAR(240) NULL,
    body TEXT NULL,
    media_url VARCHAR(500) NULL,
    items_json JSON NULL,
    position INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_careers_sections_owner (owner_id),
    INDEX idx_careers_sections_owner_pos (owner_id, position)
);
