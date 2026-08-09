export interface ConsentRecord {
  email: boolean;
  sms: boolean;
  marketing: boolean;
}
export function hasAllConsents(c: ConsentRecord): boolean {
  return c.email && c.sms && c.marketing;
}
export function hasAnyConsent(c: ConsentRecord): boolean {
  return c.email || c.sms || c.marketing;
}
export function consentedChannels(c: ConsentRecord): string[] {
  const r: string[] = [];
  if (c.email) r.push('email');
  if (c.sms) r.push('sms');
  if (c.marketing) r.push('marketing');
  return r;
}
export function canContact(c: ConsentRecord, channel: 'email' | 'sms'): boolean {
  return channel === 'email' ? c.email : c.sms;
}
export function consentSummary(c: ConsentRecord): string {
  const ch = consentedChannels(c);
  return ch.length === 0 ? 'No consents' : ch.join(', ');
}
export function emptyConsent(): ConsentRecord {
  return { email: false, sms: false, marketing: false };
}
