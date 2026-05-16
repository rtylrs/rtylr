/** Drop empty/null values from a query-string param map. */
export function trimQueryParams(
  params: Record<string, string | number | null | undefined>
): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(params)) {
    if (v == null) continue;
    const s = String(v).trim();
    if (s.length === 0) continue;
    out[k] = s;
  }
  return out;
}
