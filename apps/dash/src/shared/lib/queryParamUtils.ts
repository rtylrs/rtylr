/**
 * Query parameter utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type ParamValue = string | number | boolean | null | undefined;

export type ParamMap = Record<string, ParamValue>;

/** Serialize a ParamMap to a URL query string (without leading "?"). */
export function toQueryString(params: ParamMap): string {
  const parts: string[] = [];
  for (const [key, value] of Object.entries(params)) {
    if (value == null) continue;
    parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
  }
  return parts.join('&');
}

/** Parse a URL query string (with or without "?") into a plain object. */
export function fromQueryString(qs: string): Record<string, string> {
  const result: Record<string, string> = {};
  const cleaned = qs.startsWith('?') ? qs.slice(1) : qs;
  if (!cleaned) return result;
  for (const part of cleaned.split('&')) {
    const [rawKey, rawValue = ''] = part.split('=');
    if (!rawKey) continue;
    result[decodeURIComponent(rawKey)] = decodeURIComponent(rawValue);
  }
  return result;
}

/** Return a new param map with the given key set to a value. */
export function setParam(params: ParamMap, key: string, value: ParamValue): ParamMap {
  return { ...params, [key]: value };
}

/** Return a new param map with the given key removed. */
export function removeParam(params: ParamMap, key: string): ParamMap {
  const next = { ...params };
  delete next[key];
  return next;
}

/** Return the string value for a key, or null when not present. */
export function getStringParam(params: ParamMap, key: string): string | null {
  const val = params[key];
  if (val == null) return null;
  return String(val);
}

/** Return the numeric value for a key, or null when not a valid number. */
export function getNumberParam(params: ParamMap, key: string): number | null {
  const val = params[key];
  if (val == null) return null;
  const n = Number(val);
  return isNaN(n) ? null : n;
}

/** Return true when the param map has no non-null entries. */
export function isEmpty(params: ParamMap): boolean {
  return Object.values(params).every((v) => v == null);
}
