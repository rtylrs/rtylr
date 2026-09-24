/**
 * Post a lead to the worker's /api/v1/lead endpoint.
 *
 * Both forms used to set `window.location.href` to a `mailto:` URL. That has
 * three problems this replaces: it cannot tell whether anything was sent, it
 * depends on the visitor having a configured mail client, and Cloudflare's
 * Email Address Obfuscation rewrites the address for any client that does not
 * run JavaScript.
 *
 * The success state is shown only on a 2xx. Anything else surfaces an error
 * with the visible address as the fallback, because a form that says "sent"
 * while the lead is gone is worse than the link it replaced.
 */
export type LeadForm = 'demo' | 'contact';

export interface LeadPayload {
  form: LeadForm;
  email: string;
  name?: string;
  company?: string;
  country?: string;
  companySize?: string;
  reason?: string;
  message?: string;
  /** Never shown, never filled by a human. See validateLead in _worker.js. */
  honeypot?: string;
}

export async function submitLead(payload: LeadPayload): Promise<void> {
  const res = await fetch('/api/v1/lead', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    throw new Error(`lead submission failed with ${res.status}`);
  }
}
