/**
 * Reads the post files from disk. Server side only: import this from
 * getStaticProps and getStaticPaths, never from a component, so no post body
 * reaches a bundle for a page that does not render it.
 *
 * One post is one folder, content/blog/posts/<slug>/, holding en.json and,
 * when the Arabic version exists, ar.json. The pair shares its slug, which is
 * what makes the hreflang set derivable instead of maintained.
 */
import fs from 'node:fs';
import path from 'node:path';

import {
  BLOG_LOCALES,
  type AuthorId,
  type BlogLocale,
  type Post,
  type SegmentId,
} from './types';

const POSTS_DIR = path.join(process.cwd(), 'content/blog/posts');

/** Tags with fewer posts than this get no listing page: one card is not a page. */
export const MIN_POSTS_PER_TAG = 2;

let cache: Post[] | undefined;

function readAll(): Post[] {
  if (cache) return cache;
  const posts: Post[] = [];
  for (const slug of fs.readdirSync(POSTS_DIR).sort()) {
    const dir = path.join(POSTS_DIR, slug);
    if (!fs.statSync(dir).isDirectory()) continue;
    for (const locale of BLOG_LOCALES) {
      const file = path.join(dir, `${locale}.json`);
      if (!fs.existsSync(file)) continue;
      const post = JSON.parse(fs.readFileSync(file, 'utf8')) as Post;
      if (post.slug !== slug || post.locale !== locale) {
        throw new Error(`${file} declares ${post.slug}/${post.locale}, which is not where it lives`);
      }
      posts.push(post);
    }
  }
  cache = posts;
  return posts;
}

const newestFirst = (a: Post, b: Post): number =>
  a.date === b.date ? a.slug.localeCompare(b.slug) : a.date < b.date ? 1 : -1;

export function getAllPosts(locale: BlogLocale): Post[] {
  return readAll()
    .filter((p) => p.locale === locale)
    .sort(newestFirst);
}

export function getPost(slug: string, locale: BlogLocale): Post | undefined {
  return readAll().find((p) => p.slug === slug && p.locale === locale);
}

/** The locales one post exists in. A post in one locale has no alternates. */
export function localesOfPost(slug: string): BlogLocale[] {
  return BLOG_LOCALES.filter((locale) => getPost(slug, locale));
}

export const postsInSegment = (segment: SegmentId, locale: BlogLocale): Post[] =>
  getAllPosts(locale).filter((p) => p.segment === segment);

export const postsWithTag = (tag: string, locale: BlogLocale): Post[] =>
  getAllPosts(locale).filter((p) => p.tags.includes(tag));

export const postsByAuthor = (author: AuthorId, locale: BlogLocale): Post[] =>
  getAllPosts(locale).filter((p) => p.author === author);

/** Tags that earn a listing page in this locale, most used first. */
export function listedTags(locale: BlogLocale): string[] {
  const counts = new Map<string, number>();
  for (const post of getAllPosts(locale)) {
    for (const tag of post.tags) counts.set(tag, (counts.get(tag) ?? 0) + 1);
  }
  return [...counts]
    .filter(([, n]) => n >= MIN_POSTS_PER_TAG)
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([tag]) => tag);
}

/** Whether the blog has anything to serve in this locale at all. */
export const hasPosts = (locale: BlogLocale): boolean => getAllPosts(locale).length > 0;

/**
 * The date as a reader sees it. Formatted here, at build time, so the server
 * render and the hydrated page cannot disagree about a timezone.
 */
export function displayDate(iso: string, locale: BlogLocale): string {
  return new Intl.DateTimeFormat(locale === 'ar' ? 'ar-u-nu-latn' : 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(`${iso}T00:00:00Z`));
}
