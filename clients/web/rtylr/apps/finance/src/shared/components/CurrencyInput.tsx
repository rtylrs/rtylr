import React from 'react';
import { useTranslation } from 'react-i18next';
import { getCurrencySymbol, type Currency } from '../lib/money';
import { majorToMinor, minorToMajor } from '../utils/money';
import { useSettings } from '../lib/settings';
import { inputStyle } from '../lib/tokens';

interface CurrencyInputProps {
  value: number | string;
  onChange: (val: number) => void;
  currency?: Currency;
  placeholder?: string;
  disabled?: boolean;
  style?: React.CSSProperties;
}

/** The active locale's own decimal separator, read from Intl rather than guessed. */
function decimalSeparator(locale: string): string {
  try {
    const part = new Intl.NumberFormat(locale).formatToParts(1.1).find((p) => p.type === 'decimal');
    return part?.value ?? '.';
  } catch {
    return '.';
  }
}

/**
 * Keeps what an amount can be typed with: digits, one sign, and at most one of
 * each separator. A repeated `.` or `,` is dropped rather than stacked, so the
 * text left in the field always reads one way.
 */
function sanitiseAmount(raw: string): string {
  let out = '';
  let seenDot = false;
  let seenComma = false;
  for (const ch of raw.replace(/[\u2012\u2013\u2014\u2212]/g, '-').replace(/[^\d.,\-\s]/g, '')) {
    if (ch === '.') {
      if (seenDot) continue;
      seenDot = true;
    } else if (ch === ',') {
      if (seenComma) continue;
      seenComma = true;
    } else if (ch === '-' && out.trim() !== '') {
      continue;
    }
    out += ch;
  }
  return out;
}

/**
 * Reads the typed text as a decimal number.
 *
 * The locale's own separator is always the decimal point. The other one is a
 * decimal point only when it is not followed by exactly three digits, because
 * three digits after a separator is a thousands group, not a fraction. That is
 * what settles the two spellings both conventions share: "1.234" is one
 * thousand two hundred and thirty-four in de and 1.234 in en, and "12,5" is
 * twelve and a half rather than a hundred and twenty-five. Treating the
 * foreign separator as a group when it is not one is how "12,50" became 1250,
 * a hundred-fold overstatement of a real amount. Spaces, including the
 * non-breaking and narrow ones a copy carries, group everywhere.
 */
function amountToNumber(text: string, decimal: string): number {
  const sign = text.startsWith('-') ? -1 : 1;
  const body = text.replace(/-/g, '').replace(/\s/g, '');
  if (body === '') return 0;
  const at = Math.max(body.lastIndexOf('.'), body.lastIndexOf(','));
  let whole = body;
  let fraction = '';
  if (at !== -1) {
    const isGrouping = body[at] !== decimal && /^\d{3}$/.test(body.slice(at + 1));
    if (!isGrouping) {
      whole = body.slice(0, at);
      fraction = body.slice(at + 1);
    }
  }
  const n = Number(`${whole.replace(/[.,]/g, '')}.${fraction}`);
  return Number.isFinite(n) ? sign * n : 0;
}

const CurrencyInput: React.FC<CurrencyInputProps> = ({
  value,
  onChange,
  currency,
  placeholder = '0.00',
  disabled,
  style,
}) => {
  const { i18n } = useTranslation();
  const locale = i18n?.resolvedLanguage || i18n?.language || 'en';
  const decimal = React.useMemo(() => decimalSeparator(locale), [locale]);
  const settingsCurrency = useSettings((s) => s.currency);
  const c = currency ?? settingsCurrency;
  const symbol = getCurrencySymbol(c);
  // While the operator is typing, show the raw text. Rendering the parsed
  // number straight back deleted a trailing separator ("1." -> 1 -> "1"), so
  // typing "1.50" became 15 -> 150. The draft is dropped on blur, when the
  // normalised value from the prop is shown again.
  const [draft, setDraft] = React.useState<string | null>(null);

  const handle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = sanitiseAmount(e.target.value);
    setDraft(raw);
    onChange(majorToMinor(amountToNumber(raw, decimal), c));
  };

  const formatted = value === 0 ? '' : String(minorToMajor(Number(value), c));

  return (
    <div style={{ position: 'relative', width: '100%' }}>
      <span
        style={{
          position: 'absolute',
          insetInlineStart: 12,
          top: '50%',
          transform: 'translateY(-50%)',
          fontSize: 13,
          color: 'var(--hub-text-dim)',
          pointerEvents: 'none',
          fontWeight: 600,
        }}
      >
        {symbol}
      </span>
      <input
        type="text"
        inputMode="decimal"
        value={draft ?? formatted}
        onChange={handle}
        onBlur={() => setDraft(null)}
        placeholder={placeholder}
        disabled={disabled}
        style={{
          ...inputStyle,
          paddingInlineStart: 28,
          fontVariantNumeric: 'tabular-nums',
          ...style,
        }}
      />
    </div>
  );
};

export default CurrencyInput;
