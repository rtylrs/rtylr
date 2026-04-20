-- Phase: outbound email notifications on new submissions.
--
-- `notification_email` is where HR wants the "new candidate"
-- notification to go. Usually a team inbox (jobs@acme.com) rather
-- than the owner's personal login. Empty = no HR notifications.
-- `email_notifications` is a master switch so the admin can pause
-- without clearing the address.

ALTER TABLE careers_customizations ADD COLUMN notification_email VARCHAR(160) NULL;
ALTER TABLE careers_customizations ADD COLUMN email_notifications BOOLEAN NOT NULL DEFAULT true;
