/**
 * Feature matrix: rtylr POS vs Omega POS.
 *
 * Each entry declares whether rtylr and Omega support a feature,
 * and optionally where the rtylr implementation lives.
 */

export type SupportLevel = 'yes' | 'no' | 'partial';

export interface FeatureEntry {
  category: string;
  feature: string;
  rtylr: SupportLevel;
  omega: SupportLevel;
  rtylrNote?: string;
}

export const FEATURE_MATRIX: FeatureEntry[] = [
  // ── Payment ────────────────────────────────────────────────────────────────
  { category: 'Payment', feature: 'Cash payment', rtylr: 'yes', omega: 'yes' },
  { category: 'Payment', feature: 'Card payment', rtylr: 'yes', omega: 'yes' },
  { category: 'Payment', feature: 'Gift card payment', rtylr: 'yes', omega: 'partial', rtylrNote: 'Multi-card split supported' },
  { category: 'Payment', feature: 'Split tender (cash + card)', rtylr: 'yes', omega: 'partial' },
  { category: 'Payment', feature: 'Offline queue (ECONNREFUSED fallback)', rtylr: 'yes', omega: 'no', rtylrNote: 'Auto-syncs on reconnect' },
  { category: 'Payment', feature: 'Tip entry', rtylr: 'yes', omega: 'yes' },

  // ── Discounts ──────────────────────────────────────────────────────────────
  { category: 'Discounts', feature: 'Flat dollar discount', rtylr: 'yes', omega: 'yes' },
  { category: 'Discounts', feature: 'Percent discount (capped at 100%)', rtylr: 'yes', omega: 'yes' },
  { category: 'Discounts', feature: 'Manager PIN override to reopen order', rtylr: 'yes', omega: 'partial', rtylrNote: 'PIN verified server-side' },

  // ── Loyalty ────────────────────────────────────────────────────────────────
  { category: 'Loyalty', feature: 'Earn points on sale', rtylr: 'yes', omega: 'yes' },
  { category: 'Loyalty', feature: 'Redeem points at checkout', rtylr: 'yes', omega: 'yes' },
  { category: 'Loyalty', feature: 'Points capped at order total', rtylr: 'yes', omega: 'no' },
  { category: 'Loyalty', feature: 'Gift card issuance', rtylr: 'yes', omega: 'partial' },

  // ── Hardware ───────────────────────────────────────────────────────────────
  { category: 'Hardware', feature: 'ESC/POS thermal receipt printer', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'ESC/POS protocol confirmation (DLE EOT)', rtylr: 'yes', omega: 'no', rtylrNote: 'Byte-level handshake' },
  { category: 'Hardware', feature: 'Star Micronics native (9101/9102)', rtylr: 'yes', omega: 'partial' },
  { category: 'Hardware', feature: 'Star CloudPRNT (4242)', rtylr: 'yes', omega: 'no' },
  { category: 'Hardware', feature: 'Zebra ZPL label printer (6101)', rtylr: 'yes', omega: 'no', rtylrNote: '~HI handshake confirmed' },
  { category: 'Hardware', feature: 'Epson ePOS-Print HTTP (8008/8043)', rtylr: 'yes', omega: 'no' },
  { category: 'Hardware', feature: 'Network LPR printer (515)', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'IPP network printer (631)', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'Cash drawer (printer kick pin)', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'Cash drawer (serial)', rtylr: 'yes', omega: 'partial' },
  { category: 'Hardware', feature: 'Cash drawer (network)', rtylr: 'yes', omega: 'no' },
  { category: 'Hardware', feature: 'KDS network display', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'KDS serial display', rtylr: 'yes', omega: 'partial' },
  { category: 'Hardware', feature: 'Subnet hardware discovery (parallel)', rtylr: 'yes', omega: 'no', rtylrNote: '254 hosts × 12 ports in parallel' },
  { category: 'Hardware', feature: 'Retry on flaky connection', rtylr: 'yes', omega: 'no', rtylrNote: 'probe_with_retry wrapper' },
  { category: 'Hardware', feature: 'Windows printer spooler (OpenPrinterW)', rtylr: 'yes', omega: 'yes' },
  { category: 'Hardware', feature: 'Customer-facing kiosk display (9999)', rtylr: 'yes', omega: 'no' },

  // ── KDS / Kitchen ──────────────────────────────────────────────────────────
  { category: 'Kitchen', feature: 'Send ticket to KDS on sale', rtylr: 'yes', omega: 'yes' },
  { category: 'Kitchen', feature: 'Update item status on KDS', rtylr: 'yes', omega: 'partial' },
  { category: 'Kitchen', feature: 'Skip KDS when transport is none', rtylr: 'yes', omega: 'yes' },

  // ── Orders ─────────────────────────────────────────────────────────────────
  { category: 'Orders', feature: 'Hold / resume order', rtylr: 'yes', omega: 'yes' },
  { category: 'Orders', feature: 'Refund (full)', rtylr: 'yes', omega: 'yes' },
  { category: 'Orders', feature: 'Refund (partial)', rtylr: 'yes', omega: 'partial' },
  { category: 'Orders', feature: 'Reopen closed order (PIN required)', rtylr: 'yes', omega: 'no' },

  // ── Inventory ──────────────────────────────────────────────────────────────
  { category: 'Inventory', feature: 'Receive stock', rtylr: 'yes', omega: 'yes' },
  { category: 'Inventory', feature: 'Waste / shrink adjustment', rtylr: 'yes', omega: 'partial' },
  { category: 'Inventory', feature: 'Manual quantity adjust with reason', rtylr: 'yes', omega: 'partial' },
  { category: 'Inventory', feature: 'Auto-decrement on sale', rtylr: 'yes', omega: 'yes' },

  // ── Session ────────────────────────────────────────────────────────────────
  { category: 'Session', feature: 'Open/close cash session', rtylr: 'yes', omega: 'yes' },
  { category: 'Session', feature: 'Cash drop during session', rtylr: 'yes', omega: 'partial' },
  { category: 'Session', feature: 'Cash pickup during session', rtylr: 'yes', omega: 'no' },
  { category: 'Session', feature: 'Closing summary with expected vs actual', rtylr: 'yes', omega: 'yes' },
];

/** Returns features where rtylr is ahead of Omega. */
export function getRtylrAdvantages(): FeatureEntry[] {
  return FEATURE_MATRIX.filter(
    (f) => f.rtylr === 'yes' && f.omega !== 'yes',
  );
}

/** Counts features by support level for each system. */
export function getSupportSummary() {
  const count = (system: 'rtylr' | 'omega', level: SupportLevel) =>
    FEATURE_MATRIX.filter((f) => f[system] === level).length;

  return {
    rtylr: { yes: count('rtylr', 'yes'), partial: count('rtylr', 'partial'), no: count('rtylr', 'no') },
    omega: { yes: count('omega', 'yes'), partial: count('omega', 'partial'), no: count('omega', 'no') },
  };
}
