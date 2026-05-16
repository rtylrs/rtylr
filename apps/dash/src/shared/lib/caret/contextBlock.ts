import { detectTimezone, timezoneOffsetLabel } from './timezone';
import { buildDashSystemContext } from './systemContext';

/**
 * Compose the full system-context block we send to the backend. Wraps
 * the per-page hint with the user's timezone + locale so the model can
 * format dates/numbers correctly.
 */
export interface FullContextInput {
  pathname: string;
  businessName?: string;
  locale?: string;
}

export function buildContextBlock(input: FullContextInput): string {
  const base = buildDashSystemContext({
    pathname: input.pathname,
    businessName: input.businessName,
  });
  const tz = detectTimezone();
  const offset = timezoneOffsetLabel();
  const locale = input.locale ?? (typeof navigator !== 'undefined' ? navigator.language : 'en-US');
  return [base, `Timezone: ${tz} (${offset}). Locale: ${locale}.`].join(' ');
}
