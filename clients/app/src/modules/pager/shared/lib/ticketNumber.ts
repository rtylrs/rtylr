export const TICKET_NUMBER_MAX_LENGTH = 12;

const ARABIC_INDIC_ZERO = 0x0660;
const EASTERN_ARABIC_INDIC_ZERO = 0x06f0;

/**
 * Shapes what is typed into what the server accepts: ASCII digits and capital
 * letters, at most twelve. Applied as the customer types, so the field always
 * shows exactly what will be sent. The server normalises again and has the
 * final word.
 */
export function normaliseTicketNumber(raw: string): string {
  let out = '';
  for (const char of raw) {
    const code = char.codePointAt(0) ?? 0;
    let ascii = char;
    if (code >= ARABIC_INDIC_ZERO && code <= ARABIC_INDIC_ZERO + 9) {
      ascii = String(code - ARABIC_INDIC_ZERO);
    } else if (code >= EASTERN_ARABIC_INDIC_ZERO && code <= EASTERN_ARABIC_INDIC_ZERO + 9) {
      ascii = String(code - EASTERN_ARABIC_INDIC_ZERO);
    }
    if (/^[A-Za-z0-9]$/.test(ascii)) out += ascii.toUpperCase();
  }
  return out.slice(0, TICKET_NUMBER_MAX_LENGTH);
}
