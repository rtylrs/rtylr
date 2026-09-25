-- Phase 4 — SEO, legal, analytics on careers_customizations.
--
-- meta_title / meta_description override the default tags injected by
-- the frontend. privacy_url / terms_url show up in the apply form's
-- consent row. ga_measurement_id / meta_pixel_id, if set, cause the
-- frontend to load GA4 / Meta Pixel tags on the public pages.
-- require_consent toggles the explicit GDPR consent checkbox on apply.

ALTER TABLE careers_customizations ADD COLUMN meta_title VARCHAR(240) NULL;
ALTER TABLE careers_customizations ADD COLUMN meta_description VARCHAR(320) NULL;
ALTER TABLE careers_customizations ADD COLUMN privacy_url VARCHAR(500) NULL;
ALTER TABLE careers_customizations ADD COLUMN terms_url VARCHAR(500) NULL;
ALTER TABLE careers_customizations ADD COLUMN ga_measurement_id VARCHAR(32) NULL;
ALTER TABLE careers_customizations ADD COLUMN meta_pixel_id VARCHAR(32) NULL;
ALTER TABLE careers_customizations ADD COLUMN require_consent BOOLEAN NOT NULL DEFAULT false;
