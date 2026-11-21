/**
 * Static paths and props for every blog route, in both locales. Server side
 * only (it reads the post files): call it from getStaticPaths and
 * getStaticProps in pages/blog and pages/ar/blog, nowhere else.
 *
 * Routes under /blog/ (and /ar/blog/):
 *   /                     index
 *   /<segment>/           one of SEGMENTS
 *   /tag/<tag>/           a tag with at least MIN_POSTS_PER_TAG posts
 *   /author/<author>/     an author with at least one post
 *   /<slug>/              a post
 *
 * A page is exported only when it has something to list. That one rule is
 * also what decides hreflang: two locales are alternates exactly when both
 * exported the page, so the sitemap reads the export instead of a list.
 */
import type { GetStaticPaths, GetStaticProps } from 'next';

import {
  displayDate,
  getAllPosts,
  getPost,
  hasPosts,
  listedTags,
  postsByAuthor,
  postsInSegment,
  postsWithTag,
} from './posts';
import {
  AUTHORS,
  BLOG_LOCALES,
  RESERVED_SLUGS,
  SEGMENTS,
  authorById,
  blogPath,
  segmentBySlug,
  toSummary,
  type AuthorId,
  type BlogLocale,
  type Post,
  type ListedPost,
  type SegmentId,
} from './types';

export type BlogPageProps =
  | {
      kind: 'index';
      locale: BlogLocale;
      posts: ListedPost[];
      segments: SegmentId[];
      alternate: string | null;
    }
  | {
      kind: 'segment';
      locale: BlogLocale;
      segment: SegmentId;
      posts: ListedPost[];
      segments: SegmentId[];
      alternate: string | null;
    }
  | {
      kind: 'tag';
      locale: BlogLocale;
      tag: string;
      posts: ListedPost[];
      segments: SegmentId[];
      alternate: string | null;
    }
  | {
      kind: 'author';
      locale: BlogLocale;
      author: AuthorId;
      posts: ListedPost[];
      segments: SegmentId[];
      alternate: string | null;
    }
  | {
      kind: 'post';
      locale: BlogLocale;
      post: Post;
      publishedLabel: string;
      updatedLabel: string | null;
      related: ListedPost[];
      linkedTags: string[];
      alternate: string | null;
    };

/** Every route under /blog/ one locale exports, as path segments. */
function routesFor(locale: BlogLocale): string[][] {
  if (!hasPosts(locale)) return [];
  const routes: string[][] = [[]];
  for (const s of SEGMENTS) {
    if (postsInSegment(s.id, locale).length) routes.push([s.slug]);
  }
  for (const tag of listedTags(locale)) routes.push(['tag', tag]);
  for (const a of AUTHORS) {
    if (postsByAuthor(a.id, locale).length) routes.push(['author', a.id]);
  }
  for (const post of getAllPosts(locale)) {
    if (RESERVED_SLUGS.includes(post.slug)) {
      throw new Error(`post slug "${post.slug}" collides with a blog route`);
    }
    routes.push([post.slug]);
  }
  return routes;
}

const key = (path: string[]): string => path.join('/');

function alternateOf(path: string[], locale: BlogLocale): string | null {
  const other = BLOG_LOCALES.find((l) => l !== locale) as BlogLocale;
  const exported = new Set(routesFor(other).map(key));
  return exported.has(key(path)) ? blogPath(other, path.length ? `${key(path)}/` : '') : null;
}

/** Up to three posts from the same segment, then the newest from any. */
const listed = (post: Post): ListedPost => ({
  ...toSummary(post),
  dateLabel: displayDate(post.dateModified ?? post.date, post.locale),
});

function relatedTo(post: Post): ListedPost[] {
  const others = getAllPosts(post.locale).filter((p) => p.slug !== post.slug);
  const same = others.filter((p) => p.segment === post.segment);
  const rest = others.filter((p) => p.segment !== post.segment);
  return [...same, ...rest].slice(0, 3).map(listed);
}

function propsFor(locale: BlogLocale, path: string[]): BlogPageProps | null {
  if (!routesFor(locale).some((r) => key(r) === key(path))) return null;
  const alternate = alternateOf(path, locale);
  const list = (posts: Post[]): ListedPost[] => posts.map(listed);
  const segments = SEGMENTS.filter((s) => postsInSegment(s.id, locale).length).map((s) => s.id);

  if (path.length === 0) {
    return { kind: 'index', locale, posts: list(getAllPosts(locale)), segments, alternate };
  }
  if (path[0] === 'tag') {
    return {
      kind: 'tag',
      locale,
      tag: path[1],
      posts: list(postsWithTag(path[1], locale)),
      segments,
      alternate,
    };
  }
  if (path[0] === 'author') {
    const author = authorById(path[1]);
    if (!author) return null;
    return {
      kind: 'author',
      locale,
      author: author.id,
      posts: list(postsByAuthor(author.id, locale)),
      segments,
      alternate,
    };
  }
  const segment = segmentBySlug(path[0]);
  if (segment) {
    return {
      kind: 'segment',
      locale,
      segment: segment.id,
      posts: list(postsInSegment(segment.id, locale)),
      segments,
      alternate,
    };
  }
  const post = getPost(path[0], locale);
  if (!post) return null;
  return {
    kind: 'post',
    locale,
    post,
    publishedLabel: displayDate(post.date, locale),
    updatedLabel:
      post.dateModified && post.dateModified !== post.date
        ? displayDate(post.dateModified, locale)
        : null,
    related: relatedTo(post),
    linkedTags: listedTags(locale),
    alternate,
  };
}

export const blogStaticPaths =
  (locale: BlogLocale): GetStaticPaths =>
  () => ({
    paths: routesFor(locale).map((path) => ({ params: { path } })),
    fallback: false,
  });

export const blogStaticProps =
  (locale: BlogLocale): GetStaticProps<BlogPageProps> =>
  ({ params }) => {
    const raw = params?.path;
    const path = Array.isArray(raw) ? raw : raw ? [raw] : [];
    const props = propsFor(locale, path);
    return props ? { props } : { notFound: true };
  };
