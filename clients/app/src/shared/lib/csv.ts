/**
 * CSV export for the portal: encoding that is safe to open in a spreadsheet,
 * and the browser download.
 *
 * Modules each grew their own encoder (hr, crm, finance, erp), with different
 * quoting and only some of them guarding against formulas. New exports use
 * this one so the formula guard and the UTF-8 byte order mark come for free.
 */

/**
 * Characters that make a spreadsheet read a cell as a formula (or, for tab
 * and carriage return, that some of them strip before doing so). A value
 * such as `=HYPERLINK(...)` must arrive as text, not run.
 */
const FORMULA_START = /^[=+\-@\t\r]/;

/**
 * One CSV field: always quoted, quotes doubled, and a value that would start
 * a formula prefixed with an apostrophe so it stays text.
 */
export function csvField(value: string | number | null | undefined): string {
  let text = value == null ? '' : String(value);
  if (FORMULA_START.test(text)) text = `'${text}`;
  return `"${text.replace(/"/g, '""')}"`;
}

/**
 * A whole document. The byte order mark makes Excel read it as UTF-8, which
 * Arabic and accented names need; CRLF is what RFC 4180 asks for.
 */
export function buildCsv(rows: ReadonlyArray<ReadonlyArray<string | number | null>>): string {
  return `﻿${rows.map((row) => row.map(csvField).join(',')).join('\r\n')}\r\n`;
}

/** Hands the CSV to the browser as a download. Nothing leaves the device. */
export function downloadCsv(fileName: string, csv: string): void {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;
  link.rel = 'noopener';
  document.body.appendChild(link);
  link.click();
  link.remove();
  // Revoked after the click has been handled, not during it.
  setTimeout(() => URL.revokeObjectURL(url), 0);
}

/**
 * Reads a CSV document into rows of cells (RFC 4180: quoted fields, doubled
 * quotes, line breaks inside quotes, CRLF or LF). A leading byte order mark is
 * dropped, and a leading apostrophe that buildCsv added to stop a formula is
 * kept, since the reader cannot tell it from a real one.
 */
export function parseCsv(text: string): string[][] {
  const input = text.startsWith('﻿') ? text.slice(1) : text;
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = '';
  let quoted = false;
  for (let i = 0; i < input.length; i++) {
    const ch = input[i];
    if (quoted) {
      if (ch === '"') {
        if (input[i + 1] === '"') {
          cell += '"';
          i++;
        } else {
          quoted = false;
        }
      } else {
        cell += ch;
      }
      continue;
    }
    if (ch === '"' && cell === '') {
      quoted = true;
    } else if (ch === ',') {
      row.push(cell);
      cell = '';
    } else if (ch === '\n' || ch === '\r') {
      if (ch === '\r' && input[i + 1] === '\n') i++;
      row.push(cell);
      rows.push(row);
      row = [];
      cell = '';
    } else {
      cell += ch;
    }
  }
  if (cell !== '' || row.length > 0) {
    row.push(cell);
    rows.push(row);
  }
  return rows.filter((r) => r.some((c) => c.trim() !== ''));
}
