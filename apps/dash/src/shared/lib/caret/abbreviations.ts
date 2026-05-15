/** Map well-known business terms to expanded forms so dictation transcripts
 *  read more naturally. Applied only when the recognized text matches the
 *  full word. */

const EXPANSIONS: Record<string, string> = {
  pos: 'POS',
  erp: 'ERP',
  hr: 'HR',
  crm: 'CRM',
  kpi: 'KPI',
  sla: 'SLA',
  api: 'API',
  sku: 'SKU',
  sms: 'SMS',
  faq: 'FAQ',
};

export function expandAbbreviations(text: string): string {
  return text.replace(/\b([a-z]{2,4})\b/g, (m) => {
    const lower = m.toLowerCase();
    return EXPANSIONS[lower] ?? m;
  });
}
