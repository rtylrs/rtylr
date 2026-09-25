import fs from 'node:fs';
import path from 'node:path';

export type Doc = {
  slug: string[];
  route: string;
  title: string;
  description: string;
  body: string;
};

export type SearchEntry = {
  title: string;
  description: string;
  route: string;
};

const docsRoot = path.join(process.cwd(), 'docs');

function stripFrontmatter(value: string) {
  return value.replace(/^---[\s\S]*?---\s*/, '').trim();
}

function titleFromContent(value: string, fallback: string) {
  const heading = stripFrontmatter(value).match(/^#\s+(.+)$/m)?.[1];
  return heading?.trim() ?? fallback;
}

function descriptionFromContent(value: string) {
  const paragraphs = stripFrontmatter(value)
    .split(/\n\s*\n/)
    .map((paragraph) => paragraph.trim())
    .filter((paragraph) => paragraph && !paragraph.startsWith('#') && !paragraph.startsWith('```'))
    .map((paragraph) => paragraph.replace(/[*_`]/g, '').trim());

  return paragraphs[0]?.slice(0, 160) ?? 'rtylr documentation.';
}

function collectContentFiles(directory: string): string[] {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) return collectContentFiles(absolutePath);
    return entry.name.endsWith('.txt') ? [absolutePath] : [];
  });
}

function routeForFile(absolutePath: string) {
  const relativePath = path.relative(docsRoot, absolutePath).replace(/\.txt$/, '');
  const segments = relativePath.split(path.sep);
  if (segments.at(-1) === 'index') segments.pop();
  return `/${segments.join('/')}`.replace(/\/$/, '') || '/';
}

function slugForRoute(route: string) {
  return route === '/' ? [] : route.slice(1).split('/');
}

function loadDocs(): Doc[] {
  return collectContentFiles(docsRoot)
    .filter((file) => !file.includes(`${path.sep}public${path.sep}`))
    .map((file) => {
      const raw = fs.readFileSync(file, 'utf8');
      const route = routeForFile(file);
      return {
        slug: slugForRoute(route),
        route,
        title: titleFromContent(raw, path.basename(file, '.txt')),
        description: descriptionFromContent(raw),
        body: stripFrontmatter(raw),
      };
    })
    .sort((a, b) => a.route.localeCompare(b.route));
}

const docs = loadDocs();

export function getAllDocs() {
  return docs;
}

export function getDoc(slug: string[]) {
  const route = slug.length ? `/${slug.join('/')}` : '/';
  return docs.find((doc) => doc.route === route);
}

export function getSearchEntries(): SearchEntry[] {
  return docs
    .filter((doc) => doc.route !== '/')
    .map(({ title, description, route }) => ({ title, description, route }));
}

export function slugifyHeading(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[`*_]/g, '')
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-');
}

export function getTableOfContents(body: string) {
  return body
    .split('\n')
    .map((line) => line.match(/^(##|###)\s+(.+)$/))
    .filter((match): match is RegExpMatchArray => Boolean(match))
    .map((match) => ({
      title: match[2].replace(/[`*_]/g, '').trim(),
      id: slugifyHeading(match[2]),
      level: match[1].length,
    }));
}
