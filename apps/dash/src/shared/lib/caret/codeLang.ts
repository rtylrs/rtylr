/** Normalize a markdown code fence language label. */
const ALIASES: Record<string, string> = {
  js: 'javascript',
  jsx: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  py: 'python',
  sh: 'bash',
  shell: 'bash',
  zsh: 'bash',
  yml: 'yaml',
  md: 'markdown',
};

const PRETTY: Record<string, string> = {
  javascript: 'JavaScript',
  typescript: 'TypeScript',
  python: 'Python',
  bash: 'Shell',
  yaml: 'YAML',
  json: 'JSON',
  go: 'Go',
  sql: 'SQL',
  html: 'HTML',
  css: 'CSS',
  rust: 'Rust',
  ruby: 'Ruby',
  java: 'Java',
};

export function normalizeLang(raw: string | undefined): string | undefined {
  if (!raw) return undefined;
  const lower = raw.toLowerCase();
  return ALIASES[lower] ?? lower;
}

export function prettyLang(raw: string | undefined): string {
  const norm = normalizeLang(raw);
  if (!norm) return 'code';
  return PRETTY[norm] ?? norm;
}
