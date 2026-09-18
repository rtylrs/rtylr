# Privacy operations and evidence templates

Status: proposed internal procedures, to be assigned, approved and exercised. Creating this file does not demonstrate an operational program. Store completed records in an access-controlled system, not a public repository; do not commit personal data, credentials, contracts or incident evidence here.

## Accountability and processing register

Assign a privacy owner, security incident lead, backup contacts and legal escalation route. Confirm monitoring and delivery of the published privacy/security inboxes. Assess whether a formal DPO or local representative is required; a contact email is not an appointment.

For each processing activity, record: service and owner; controller/processor role; purposes; subjects and data fields; source; lawful basis (and sensitive-data condition where needed); recipients; storage and support-access countries; transfer safeguards; retention and deletion; security measures; rights handling; DPIA decision; review date. Separate provider account/billing data from customer-controlled business records and optional analytics.

## Vendor register — seed candidates, not an approved subprocessor list

| Candidate evidenced in code or existing copy                           | Evidence / uncertainty                                                   | Required fields before approval                                                                                      |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| Cloudflare                                                             | Landing deployment and existing notice                                   | Legal entity, hosting/logging role, countries, data, retention, contract, safeguards                                 |
| Paddle                                                                 | Billing UI and existing notice                                           | Exact merchant entity; independent-controller versus processor duties; payment notice and contract                   |
| Google Analytics                                                       | `apps/landing/lib/consentMode.ts`                                        | Vendor entity, website analytics data, region, retention, consent settings and transfer safeguards                   |
| PostHog                                                                | `packages/observability/src/posthog.ts` uses US endpoint when configured | Actual production enablement, data fields, retention, agreement and transfer assessment                              |
| Sentry                                                                 | `packages/observability/src/sentry.ts`, configured by DSN                | Actual project/region, diagnostic payloads, lawful basis, redaction, retention and agreement                         |
| Firebase/Google Cloud, Resend/Postmark, Axiom                          | Existing privacy copy; production deployment not established here        | Confirm or remove each candidate after operations verification; do not present alternatives as a factual vendor list |
| Database, object storage, backups, support, email/SMS and integrations | Backend/vendor inventory not reviewed                                    | Discover actual recipients, including remote human access and subprocessors of providers                             |

Each approved row needs owner, date, data categories, purpose, legal role, countries, signed agreement reference, authorization and change-notice process. Publish the accurate customer-facing list and distinguish independently controlled recipients. Check whether customer integration instructions change the role.

## Rights request procedure

1. Log receipt time, requester, relevant tenant/controller, request scope, jurisdiction, owner and applicable deadline. Use proportionate identity/representative verification; do not routinely request identity documents.
2. Decide whether RTYLR handles the request as controller or assists the customer. Route processor requests securely without ignoring the request or disclosing another tenant's records.
3. Search primary stores, logs, files, exports and relevant providers. Apply access controls, third-party redactions, justified exceptions and legal holds. Record reasons.
4. Respond within the actual applicable deadline; notify any permitted extension or refusal with reasons, complaint/appeal routes where required. Do not substitute a universal “30 days” for different legal clocks.
5. Track completion across live systems, providers and backups. Retain only appropriate evidence of fulfillment. Test access/export/deletion with a synthetic tenant and document results.

## Breach response procedure

Record first awareness and facts immediately; appoint incident lead and preserve restricted evidence. Contain, investigate, assess affected data/people, document decisions and coordinate legal review. Avoid delaying initial customer notice until every fact is confirmed.

Where acting as processor under GDPR, notify the controller without undue delay after awareness. Where acting as controller, assess the Article 33 risk threshold and potentially applicable 72-hour authority deadline, and Article 34 high-risk individual notification. Other laws and contracts may have different triggers, recipients and clocks. Document non-notification decisions and reassess with new facts. Provide phased updates and a remediation review. [European Commission breach guidance](https://commission.europa.eu/law/law-topic/data-protection/information-business-and-organisations/obligations_en).

Exercise the procedure with a tabletop before claiming a response program or 24/7 coverage. Record detection, awareness, decisions, contacts, containment, notices and lessons learned.

## Retention and deletion schedule

| Dataset                           | Purpose / role                                      | Existing public commitment to verify                                                         | Owner / evidence                                            |
| --------------------------------- | --------------------------------------------------- | -------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Account/customer data             | Contract / customer instructions                    | Account life plus recovery; published 30/60-day language needs an unambiguous starting point | Unassigned / backend test required                          |
| Backups                           | Recovery                                            | Published purge within 90 days                                                               | Unassigned / backup lifecycle and restore evidence required |
| Billing records                   | Accounting/tax                                      | Published seven years; jurisdiction-specific verification required                           | Unassigned / legal schedule required                        |
| Security and audit logs           | Security/accountability                             | Published 90 days / 12 months                                                                | Unassigned / configured expiry required                     |
| Candidates, employees, payroll    | Customer instructions / applicable employment rules | No single global period should be invented                                                   | Unassigned / customer controls and schedule required        |
| Browser analytics and diagnostics | Optional analytics / assessed diagnostic purposes   | Vendor configuration not verified                                                            | Unassigned / cookie and SDK payload inventory required      |

For each row define trigger, exact period, legal reason, accessible export, automated job, provider propagation, legal-hold exception, restricted backup retention and verification. Deleting an account row alone is insufficient. Reapply deletion after restore; reconcile offline POS copies and browser storage. Do not shorten existing customer rights without the required contractual process.

## Consent and contract evidence

Keep distinct records for contract acceptance, marketing channel/purpose consent, and optional browser tracking. Record applicable notice/version and affirmative action with proportionate evidence and retention. Re-acceptance of material changes is a separate decision from editing a footer date. A `consented` boolean without notice/version/context does not establish the whole history.

Browser verification should cover new/returning users, malformed/missing cookies, accept/deny/revoke, SPA navigation, multiple tabs/subdomains, embedded/custom-domain menus and recruitment, SDK queues, session recordings and cookie expiry. Confirm third-party requests and local persistence in production configuration, not only mocked tests. No automatic Sentry Replay is enabled by this PR; diagnostic capture still needs its own necessity and payload review.

## Commercial and product controls

- Verify renewal price, tax, seat calculations, trial conversion, cancellation and refund disclosures against actual checkout and signed merchant terms. Test cancellation and retain the result.
- Review employer/recruitment notices and sensitive fields before collection; provide proper rights routing. A business-configurable empty notice URL is not a completed notice.
- Establish suppression and unsubscribe processing for every marketing channel and sender; validate provider enforcement, customer obligations and sender identity.
- Track customer content complaints and abuse reports against the existing acceptable-use terms; assess additional local reporting/takedown duties.
- Keep IP assignments and commercial/self-host grants; inventory distributed dependency/font/image licenses and required notices, including POS bundles.
- Approve new markets and regulated features against local employment, payroll, tax/fiscalization, payment and accessibility requirements. Maintain evidence rather than claiming blanket worldwide compliance.
