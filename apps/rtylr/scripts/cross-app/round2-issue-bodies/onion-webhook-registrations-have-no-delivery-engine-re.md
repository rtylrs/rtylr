## Benefit

Tenants who register a webhook start receiving events, instead of configuring integrations that silently never fire.

## Why

Trigger and automation (4). The webhooks module is CRUD over registrations exposed through the hubdeveloper port (`main.go:732`), but a repo-wide sweep finds exactly one outbound POST site, Paddles client talking to Paddle (`internal/platform/paddle/client.go`); nothing delivers tenant webhook events (`backend_events.csv`, emission probe). Dash sells a developer surface whose subscriptions are inert. Filed in onion because filling the gap is backend platform work; it also gates any future shape-3 design that assumed webhooks exist.

## ROI

| Item                   | Value                                                                         |
| ---------------------- | ----------------------------------------------------------------------------- |
| Affected tenants       | those using api_keys/webhooks developer features                              |
| Manual work today      | tenants poll or hand-sync; or do not integrate at all                         |
| Modelled monthly value | unquantified enablement; small now, grows with API adoption                   |
| Build cost             | 5 to 8 developer days (worker-driven delivery, retry/backoff, signing, dedup) |
| Payback                | strategic; negative short-term on direct ROI alone                            |
| Confidence             | Modelled from assumption                                                      |

## What exists, what is missing

Registration CRUD: webhooks/internal/application/service.go:46-100. Provider port: platform/hubdeveloper/developer.go:34. Delivery: absent (search across repo). Note onion#755 covers Paddle inbound security, not this outbound engine.

## How to build it

Worker-driven dispatcher reading due registrations, POSTing signed payloads with retries and dead-lettering, per-event-type filters; reuse emailqueue patterns (workerqueue). Decide event vocabulary after HR->HUB-01 and FIN->ERP-02 needs are known; do not invent events speculatively.

## Coupling cost

Platform capability, no new domain edge until events exist. Owner: webhooks module + worker. Tenant predicate business_id scoping throughout. Degrade: absence is todays behaviour; additive.

Docs: docs/cross-app/16-backend-matrix.md async backbone; 17-backend-plumbing.md shape B4 option 3.

Labelled audit:cross-app-round2-2026-08-22, area:webhooks, type:gap, severity:medium.
