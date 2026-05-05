-- Phase 1 — visual-depth columns on careers_customizations.
--
-- Extends the single primary_color into a full palette, adds a font
-- pairing key (the frontend maps this to a Google Fonts bundle), and
-- room for a hero image + favicon.  Everything is nullable; defaults
-- live in the template so a freshly-created row still renders.

ALTER TABLE careers_customizations ADD COLUMN accent_color VARCHAR(16) NULL;
ALTER TABLE careers_customizations ADD COLUMN bg_color VARCHAR(16) NULL;
ALTER TABLE careers_customizations ADD COLUMN text_color VARCHAR(16) NULL;
ALTER TABLE careers_customizations ADD COLUMN muted_text_color VARCHAR(16) NULL;
-- font_pairing is a small enum-ish string (e.g. 'jakarta-fraunces');
-- keeping it opaque means we can add new pairings without a migration.
ALTER TABLE careers_customizations ADD COLUMN font_pairing VARCHAR(32) NULL;
ALTER TABLE careers_customizations ADD COLUMN hero_image_url VARCHAR(500) NULL;
ALTER TABLE careers_customizations ADD COLUMN favicon_url VARCHAR(500) NULL;
