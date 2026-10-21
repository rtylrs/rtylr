import type {
  ImportLeadRow,
  ImportLeadRowResult,
  ImportLeadRowsBody,
  ImportLeadRowsResponse,
  LeadFields,
  LeadImport,
} from '@voxire/voxie/crm/sales/v1/types';

/** Columns a spreadsheet can map onto. Full name is split into first and last. */
export const IMPORT_FIELDS = [
  'fullName',
  'firstName',
  'lastName',
  'jobTitle',
  'email',
  'phone',
  'companyName',
  'website',
  'linkedin',
  'industry',
  'size',
  'city',
  'country',
  'source',
  'value',
  'notes',
] as const;
export type ImportField = (typeof IMPORT_FIELDS)[number];

/** Field -> column index in the file; unmapped fields are absent. */
export type ColumnMapping = Partial<Record<ImportField, number>>;

// Header spellings people use, normalized (see normalizeHeader), in English,
// French and Arabic. Order matters: earlier fields claim a column first.
const SYNONYMS: Record<ImportField, string[]> = {
  firstName: ['firstname', 'first', 'givenname', 'forename', 'prenom', 'الاسمالاول', 'الاسمالأول'],
  lastName: [
    'lastname',
    'last',
    'surname',
    'familyname',
    'nomdefamille',
    'اسمالعائلة',
    'العائلة',
    'الكنية',
  ],
  fullName: [
    'name',
    'fullname',
    'contact',
    'contactname',
    'person',
    'nomcomplet',
    'nom',
    'الاسم',
    'الاسمالكامل',
  ],
  jobTitle: [
    'title',
    'jobtitle',
    'position',
    'role',
    'poste',
    'fonction',
    'titre',
    'المسمىالوظيفي',
    'الوظيفة',
  ],
  email: [
    'email',
    'emailaddress',
    'mail',
    'courriel',
    'adresseemail',
    'البريد',
    'البريدالالكتروني',
    'البريدالإلكتروني',
  ],
  phone: [
    'phone',
    'phonenumber',
    'mobile',
    'tel',
    'telephone',
    'cell',
    'cellphone',
    'whatsapp',
    'portable',
    'الهاتف',
    'رقمالهاتف',
    'الجوال',
    'الموبايل',
  ],
  companyName: [
    'company',
    'companyname',
    'organization',
    'organisation',
    'business',
    'account',
    'entreprise',
    'societe',
    'الشركة',
    'المؤسسة',
  ],
  website: ['website', 'web', 'url', 'site', 'domain', 'siteweb', 'siteinternet', 'الموقع'],
  linkedin: ['linkedin', 'linkedinurl', 'linkedinprofile', 'profillinkedin'],
  industry: ['industry', 'sector', 'secteur', 'industrie', 'القطاع', 'المجال'],
  size: [
    'size',
    'companysize',
    'employees',
    'headcount',
    'taille',
    'effectif',
    'الحجم',
    'عددالموظفين',
  ],
  city: ['city', 'town', 'ville', 'المدينة'],
  country: ['country', 'pays', 'البلد', 'الدولة'],
  source: ['source', 'leadsource', 'origine', 'المصدر'],
  value: [
    'value',
    'dealvalue',
    'amount',
    'budget',
    'revenue',
    'montant',
    'valeur',
    'القيمة',
    'المبلغ',
  ],
  notes: ['notes', 'note', 'comments', 'comment', 'description', 'remarques', 'ملاحظات'],
};

/** Lowercase, accents dropped, letters and digits only. */
export function normalizeHeader(header: string): string {
  return header
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]/gu, '');
}

const NORMALIZED_SYNONYMS = Object.fromEntries(
  Object.entries(SYNONYMS).map(([field, names]) => [field, new Set(names.map(normalizeHeader))])
) as Record<ImportField, Set<string>>;

const GUESS_ORDER: ImportField[] = [
  'firstName',
  'lastName',
  'email',
  'phone',
  'companyName',
  'jobTitle',
  'website',
  'linkedin',
  'industry',
  'size',
  'city',
  'country',
  'source',
  'value',
  'notes',
  'fullName',
];

/** Maps each field to the first unclaimed column whose header names it. */
export function guessMapping(headers: string[]): ColumnMapping {
  const normalized = headers.map(normalizeHeader);
  const claimed = new Set<number>();
  const mapping: ColumnMapping = {};
  for (const field of GUESS_ORDER) {
    if (field === 'fullName' && mapping.firstName !== undefined) continue;
    const index = normalized.findIndex(
      (header, column) => !claimed.has(column) && NORMALIZED_SYNONYMS[field].has(header)
    );
    if (index >= 0) {
      mapping[field] = index;
      claimed.add(index);
    }
  }
  return mapping;
}

/** Whether the mapping can name a lead: a company or a person. */
export function mappingNamesLead(mapping: ColumnMapping): boolean {
  return (
    mapping.companyName !== undefined ||
    mapping.fullName !== undefined ||
    mapping.firstName !== undefined ||
    mapping.lastName !== undefined
  );
}

function amountCell(value: string): string | undefined {
  const cleaned = value
    .replace(/[^\d.,-]/g, '')
    .replace(/,(?=\d{3}\b)/g, '')
    .replace(',', '.');
  if (!/^\d+(\.\d+)?$/.test(cleaned)) return undefined;
  return Number(cleaned).toFixed(2);
}

/** One file row as lead fields. An unreadable value is dropped, not fatal. */
export function rowToLead(row: string[], mapping: ColumnMapping, billing: string): LeadFields {
  const cell = (field: ImportField) => {
    const index = mapping[field];
    return index === undefined ? '' : (row[index] ?? '').trim();
  };
  let firstName = cell('firstName');
  let lastName = cell('lastName');
  const fullName = cell('fullName');
  if (fullName && !firstName && !lastName) {
    const space = fullName.indexOf(' ');
    firstName = space === -1 ? fullName : fullName.slice(0, space);
    lastName = space === -1 ? '' : fullName.slice(space + 1).trim();
  }
  return {
    firstName,
    lastName,
    jobTitle: cell('jobTitle'),
    email: cell('email'),
    phone: cell('phone'),
    companyName: cell('companyName'),
    website: cell('website'),
    linkedin: cell('linkedin'),
    industry: cell('industry'),
    size: cell('size'),
    city: cell('city'),
    country: cell('country'),
    source: cell('source'),
    value: amountCell(cell('value')),
    billing,
    notes: cell('notes'),
  };
}

export interface ImportOptions {
  fileName: string;
  duplicates: 'skip' | 'fill';
  assignedTo?: string;
  source: string;
  billing: string;
}

export interface ImportProgress {
  importId?: string;
  sent: number;
  total: number;
  created: number;
  updated: number;
  skipped: number;
  failed: number;
}

export interface ImportFailure {
  /** The row's line in the file (the header is line 1). */
  line: number;
  error: string;
}

export interface ImportTransport {
  start: (fileName: string, totalRows: number) => Promise<LeadImport>;
  send: (id: string, body: ImportLeadRowsBody) => Promise<ImportLeadRowsResponse>;
  finish: (id: string) => Promise<unknown>;
}

export const CHUNK_SIZE = 500;
const RETRIES = 3;

async function withRetries<T>(run: () => Promise<T>, signal?: AbortSignal): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt < RETRIES; attempt += 1) {
    if (signal?.aborted) throw new DOMException('Import cancelled', 'AbortError');
    try {
      return await run();
    } catch (error) {
      lastError = error;
      const status = (error as { statusCode?: number } | null)?.statusCode;
      // A rejected request will be rejected again; only transient failures retry.
      if (status !== undefined && status < 500 && status !== 429) throw error;
      await new Promise((resolve) => setTimeout(resolve, 500 * 2 ** attempt));
    }
  }
  throw lastError;
}

/**
 * Imports a file's data rows in chunks, reporting progress after each one and
 * collecting the rows the server rejected. Stops between chunks when the
 * signal aborts; the rows already sent stay imported.
 */
export async function runLeadImport(
  rows: string[][],
  mapping: ColumnMapping,
  options: ImportOptions,
  transport: ImportTransport,
  onProgress: (progress: ImportProgress) => void,
  signal?: AbortSignal
): Promise<{ progress: ImportProgress; failures: ImportFailure[] }> {
  const progress: ImportProgress = {
    sent: 0,
    total: rows.length,
    created: 0,
    updated: 0,
    skipped: 0,
    failed: 0,
  };
  const failures: ImportFailure[] = [];
  const started = await withRetries(() => transport.start(options.fileName, rows.length), signal);
  const id = started.id ?? '';
  progress.importId = id;
  onProgress({ ...progress });

  for (let offset = 0; offset < rows.length; offset += CHUNK_SIZE) {
    if (signal?.aborted) break;
    const chunk: ImportLeadRow[] = rows.slice(offset, offset + CHUNK_SIZE).map((row, index) => ({
      row: offset + index + 2,
      lead: rowToLead(row, mapping, options.billing),
    }));
    const response = await withRetries(
      () =>
        transport.send(id, {
          rows: chunk,
          duplicates: options.duplicates,
          assignedTo: options.assignedTo,
          source: options.source,
        }),
      signal
    );
    for (const result of response.results ?? []) countResult(progress, failures, result);
    progress.sent = Math.min(rows.length, offset + chunk.length);
    onProgress({ ...progress });
  }
  await withRetries(() => transport.finish(id));
  return { progress, failures };
}

function countResult(
  progress: ImportProgress,
  failures: ImportFailure[],
  result: ImportLeadRowResult
): void {
  switch (result.outcome) {
    case 'created':
      progress.created += 1;
      break;
    case 'updated':
      progress.updated += 1;
      break;
    case 'skipped':
      progress.skipped += 1;
      break;
    default:
      progress.failed += 1;
      failures.push({ line: result.row ?? 0, error: result.error || 'rejected' });
  }
}
