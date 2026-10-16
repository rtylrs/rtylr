const FORMULA_PREFIX = /^[=+\-@\t\r]/;

export function escapeCsvValue(val: unknown): string {
  let str = val == null ? '' : String(val);
  if (typeof val === 'string' && FORMULA_PREFIX.test(str)) {
    str = `'${str}`;
  }
  if (/[",\r\n]/.test(str)) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}
