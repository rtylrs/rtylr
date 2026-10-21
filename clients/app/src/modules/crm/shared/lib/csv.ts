/**
 * A small RFC 4180 reader: quoted fields may hold the delimiter, doubled
 * quotes and line breaks. The delimiter is detected from the header line, so
 * spreadsheets exported with semicolons (common in French locales) or tabs
 * read correctly. Blank lines are skipped.
 */
export function parseCsv(input: string, delimiter = detectDelimiter(input)): string[][] {
  const text = input.charCodeAt(0) === 0xfeff ? input.slice(1) : input;
  const records: string[][] = [];
  let record: string[] = [];
  let field = '';
  let quoted = false;
  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    if (quoted) {
      if (char === '"') {
        if (text[i + 1] === '"') {
          field += '"';
          i += 1;
        } else {
          quoted = false;
        }
      } else {
        field += char;
      }
      continue;
    }
    if (char === '"' && field.trim() === '') {
      field = '';
      quoted = true;
      continue;
    }
    if (char === delimiter) {
      record.push(field.trim());
      field = '';
      continue;
    }
    if (char === '\n' || char === '\r') {
      if (char === '\r' && text[i + 1] === '\n') i += 1;
      record.push(field.trim());
      field = '';
      if (record.some((value) => value !== '')) records.push(record);
      record = [];
      continue;
    }
    field += char;
  }
  if (field !== '' || record.length > 0) {
    record.push(field.trim());
    if (record.some((value) => value !== '')) records.push(record);
  }
  return records;
}

/** The delimiter used most outside quotes on the first line. */
export function detectDelimiter(input: string): string {
  const end = input.search(/\r?\n/);
  const firstLine = end === -1 ? input : input.slice(0, end);
  let quoted = false;
  const counts: Record<string, number> = { ',': 0, ';': 0, '\t': 0 };
  for (const char of firstLine) {
    if (char === '"') quoted = !quoted;
    else if (!quoted && char in counts) counts[char] += 1;
  }
  const [best, count] = Object.entries(counts).sort((a, b) => b[1] - a[1])[0];
  return count > 0 ? best : ',';
}

/** Quotes a value for a CSV cell when it needs it. */
export function csvCell(value: string): string {
  return /[",\r\n]/.test(value) ? `"${value.replace(/"/g, '""')}"` : value;
}

export function toCsv(rows: string[][]): string {
  return rows.map((row) => row.map(csvCell).join(',')).join('\r\n');
}
