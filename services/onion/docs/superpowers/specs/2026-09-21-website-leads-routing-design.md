# Website lead routing and newsletter ingestion

Sub-project 1 of 4 in the voxire.com forms overhaul. Date: 2026-09-21.

## Goal

Every form on voxire.com delivers to info@voxire.com and the named team
addresses, through infrastructure Voxire controls. The blog newsletter forms
leave Web3Forms. The visitor auto-reply speaks for the team, not one person.
The lead priority badge matches the budget bands the forms actually offer.

## Non-goals

Anything that changes what a visitor fills in (sub-project 2), the visual
redesign of the emails (sub-project 3), and CRM ingestion (sub-project 4).
Sending newsletters is out of scope: this stores subscribers and tells the
team, nothing more.

## Current state

- `POST /contact` and `POST /quote` in the `websitelead` module store a row in
  `voxire_website_leads`, then email a notification via Resend to two addresses
  hardcoded in `service.go`, email the visitor an auto-reply signed "Abed at
  Voxire", and append to Google Sheets when configured.
- The blog popup and report banner post directly to Web3Forms from the browser.
  Delivery depends on whichever inbox owns the access key, which nobody can
  verify from the repo.
- The contact page still carries a Web3Forms branch behind an env var that is
  not set on Cloudflare, so it always falls through to the API.
- The HOT/WARM badge looks for `$5k`, `$10k`, `10,000` and similar strings that
  none of the form options contain. Only `$20K` bands ever read HOT and nothing
  reads WARM.

## Design

### 1. Notification recipients move to configuration

`AppConfig` gains `LeadNotifyEmails []string`, read from `LEAD_NOTIFY_EMAILS`
as a comma-separated list, trimmed, lower-cased, de-duplicated, empties
dropped. When the variable is unset or parses to nothing, the default list
applies so a bad deploy can never send lead notifications to nobody:

```
info@voxire.com
sales@voxire.com
mohammad.saab@voxire.com
mohamad.homsi@voxire.com
amounehabed@gmail.com
```

`websitelead.NewService` receives the list instead of reading a package
variable. `.env.example` documents the variable under the email section and
`scripts/check_env_example.py` passes.

### 2. Auto-reply in team voice

| Field | Before | After |
|---|---|---|
| From | `Abed at Voxire <info@voxire.com>` | `Voxire <info@voxire.com>` |
| Reply-To | info@voxire.com | info@voxire.com |
| Heading | "Got it, Rania." | "Thanks, Rania. We have your request." |
| Body | "one of our team will get back to you" | unchanged promise: reply within 24 hours |
| WhatsApp | "reach Abed directly", personal prefill | "reach the team on WhatsApp", prefill "Hi Voxire, I just submitted a request on voxire.com" |
| Footer | unchanged | unchanged |

The notification email's WhatsApp quick action loses "this is Abed from
Voxire" and prefills "Hi {Name}, this is Voxire following up on your request."
The HTML stays inline `fmt.Sprintf` in this sub-project. Moving both emails
into the embedded template system is sub-project 3, and the template chrome
there is RTYLR-branded, which is its own design question.

### 3. Newsletter endpoint

`POST /newsletter`, anonymous, rate limited to 10 per minute per route like
the lead routes, registered through `RegisterPOSTWithValidationAndMiddleware`.

Request:

| Field | Rule |
|---|---|
| `email` | required, must match a simple `local@domain.tld` check, max 255 |
| `source` | required, one of `intel_drop`, `report_banner`, `article_pdf_mid`, `article_pdf_end` |
| `post_slug` | optional, max 255 |
| `post_title` | optional, max 255 |
| `post_category` | optional, max 100 |
| `language` | optional, `en` or `ar`, default `en` |

Response `200 {"id": n, "message": "Subscribed"}`. Validation failures return
`400` with the field named, matching the lead handlers. Storage failure
returns `500 "Failed to subscribe"`.

Storage: new table `voxire_newsletter_subscribers`.

```sql
CREATE TABLE IF NOT EXISTS voxire_newsletter_subscribers (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    source        VARCHAR(50)  NOT NULL,
    post_slug     VARCHAR(255),
    post_title    VARCHAR(255),
    post_category VARCHAR(100),
    language      VARCHAR(5)   NOT NULL DEFAULT 'en',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_newsletter_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

No unique constraint on email. The same reader can arrive twice from two
posts and both facts are useful; de-duplication belongs to whatever sends the
newsletter later. The schema lives in two places by repository convention:
`internal/modules/websitelead/sql/schema.sql` as the sqlc input and
`internal/platform/database/schemas/voxire/newsletter_subscribers.sql` as the
file `scripts/migrate.sh` applies. The `voxire` directory is already in the
migration order, so a new file there is picked up without further wiring.

After the insert, one notification goes out asynchronously with the same
detached-context pattern the lead service uses:

| Field | Value |
|---|---|
| To | `NEWSLETTER_NOTIFY_EMAILS`, default `info@voxire.com` only |
| From | `Voxire Website <noreply@voxire.com>` |
| Reply-To | the subscriber |
| Subject | `[Newsletter] rania@example.com via intel_drop` |
| Body | short card: email, source, post title and slug when present, language, time |

Newsletter signups go to info@ alone by default because five inboxes do not
need every subscriber. The variable exists so that can change without a deploy.
No email is sent to the subscriber. The popup copy already promises a first
issue "within the week", which nothing sends today. That copy is flagged for
sub-project 2, not changed here.

### 4. Budget badge matches the real bands

The two forms offer different bands today. Contact: "Under $2K", "$2K - $8K",
"$8K - $20K", "$20K+", "Let's figure it out". Quote: "Under $2k", "$2k-$5k",
"$5k-$15k", "$15k+", "Let's discuss". Unifying them is a sub-project 2 decision.
The badge has to read both, so `priorityBadge` extracts every number followed
by `K` or `k` (case-insensitive) or written with a thousands separator, takes
the largest figure in thousands of dollars, and maps:

| Largest figure in the band | Badge |
|---|---|
| 15 or more | HOT |
| 5 to 14 | WARM |
| below 5, or no figure ("Let's discuss", "Let's figure it out", empty) | NEW |

Result per band:

| Form | Band | Badge |
|---|---|---|
| Contact | Under $2K | NEW |
| Contact | $2K - $8K | WARM |
| Contact | $8K - $20K | HOT |
| Contact | $20K+ | HOT |
| Quote | Under $2k | NEW |
| Quote | $2k-$5k | WARM |
| Quote | $5k-$15k | HOT |
| Quote | $15k+ | HOT |
| Either | free text, empty | NEW |

Every row above becomes a test case.

### 5. voxire.com changes

- `NewsletterDrop` and `NewsletterReportBanner` post to `/newsletter` through
  the existing axios client in `src/lib/api.ts`, sending `email`, `source`,
  `post_slug`, `post_title`, `post_category` and `language`. Success is a 2xx
  response. The popup and banner keep their current success and error copy.
- `src/lib/web3forms.ts` is deleted. The Web3Forms branch in `pages/contact.tsx`
  is deleted so the contact form has one path. The Web3Forms section of
  `.env.local.example` is deleted.
- In local development the API base is `localhost:9999`, so both newsletter
  forms fail locally unless onion runs. That matches how the lead forms already
  behave and is documented in the site's CLAUDE.md.

## Data flow after the change

```
visitor browser
  contact / quote  ->  POST api.voxire.com/contact|/quote
  blog popup       ->  POST api.voxire.com/newsletter
  blog banner      ->  POST api.voxire.com/newsletter

onion websitelead
  /contact,/quote  -> voxire_website_leads row
                   -> notification to LEAD_NOTIFY_EMAILS (Resend)
                   -> team-voice auto-reply to visitor (Resend)
                   -> Google Sheets row when configured
  /newsletter      -> voxire_newsletter_subscribers row
                   -> notification to NEWSLETTER_NOTIFY_EMAILS (Resend)
```

## Error handling

Unchanged philosophy: the database write is the source of truth and returns
the HTTP result. Email and Sheets run detached with a 10 second timeout and
log a warning on failure. A Resend outage never turns a stored lead or
subscriber into a visitor-facing error.

## Testing

Go, in the `websitelead` package and `config`:

- `LEAD_NOTIFY_EMAILS` parsing: unset, empty, whitespace, duplicates, mixed
  case, trailing comma. Default list applies when the result is empty.
- `priorityBadge` table test over every band in both forms plus the
  free-text cases.
- `NewsletterHandler.Validate`: missing email, malformed email, unknown source,
  bad language, happy path for each source.
- `Service.Subscribe` with a mock repository: delegates, propagates error,
  and the notification goes to the configured newsletter list, not the lead
  list.
- `make test`, `make lint`, `make vet`, `make json-tags-check`,
  `make mutation-validation-check`.

voxire.com: `npm run typecheck`, `pnpm build`, and after the onion deploy one
real signup from the production popup, then confirm the row and the email in
info@voxire.com.

## Rollout order

1. Merge and deploy onion. Apply `newsletter_subscribers.sql` with
   `scripts/migrate.sh`. Verify with:

   ```bash
   curl -s -X POST -H 'Content-Type: application/json' --data '{}' https://api.voxire.com/newsletter
   ```

   Expect `400 {"message":"email is required"}`.
2. Merge and deploy voxire.com. Until this lands, the blog forms keep using
   Web3Forms and keep working.
3. Submit one signup from a live blog post and one from the banner, check the
   two rows and the two `[Newsletter]` emails in info@voxire.com.
4. Set `LEAD_NOTIFY_EMAILS` in the onion environment only if the list should
   differ from the default. The default already is the agreed list.

## Open items carried forward

- Whether Voxire exists as a business in RTYLR production. Needed by
  sub-project 4, unreachable from the development machine.
- Whether the onion `.env.example` gate and the schema drift check are part of
  the deploy pipeline or run by hand. Affects step 1 above, not the design.
