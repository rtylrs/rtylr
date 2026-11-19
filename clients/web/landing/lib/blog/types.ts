/**
 * The blog's content model. Pure data and types, safe to import anywhere:
 * the loader that reads the post files from disk lives in ./posts.ts and
 * must only be reached from getStaticProps and getStaticPaths.
 */

export type BlogLocale = 'en' | 'ar';

export const BLOG_LOCALES: BlogLocale[] = ['en', 'ar'];

/** Who a post is written for. Every post belongs to exactly one. */
export type SegmentId = 'small' | 'mid' | 'large' | 'vision';

export interface Segment {
  id: SegmentId;
  /** URL segment under /blog/. A post may not take one of these as its slug. */
  slug: string;
}

export const SEGMENTS: Segment[] = [
  { id: 'small', slug: 'small-business' },
  { id: 'mid', slug: 'mid-size' },
  { id: 'large', slug: 'enterprise' },
  { id: 'vision', slug: 'vision' },
];

export const segmentBySlug = (slug: string): Segment | undefined =>
  SEGMENTS.find((s) => s.slug === slug);

export const segmentById = (id: SegmentId): Segment =>
  SEGMENTS.find((s) => s.id === id) as Segment;

/** Path segments under /blog/ that are routes of their own. */
export const RESERVED_SLUGS = [...SEGMENTS.map((s) => s.slug), 'tag', 'author'];

export type AuthorId = 'abed-el-fattah-amouneh' | 'mohammad-homsi';

export interface Author {
  id: AuthorId;
  /** Key under blog.authors in the locale files. */
  key: 'abed' | 'mohammad';
  links: { label: string; href: string }[];
}

export const AUTHORS: Author[] = [
  {
    id: 'abed-el-fattah-amouneh',
    key: 'abed',
    links: [{ label: 'Voxire', href: 'https://voxire.com' }],
  },
  {
    id: 'mohammad-homsi',
    key: 'mohammad',
    links: [{ label: 'Voxire', href: 'https://voxire.com' }],
  },
];

export const authorById = (id: string): Author | undefined => AUTHORS.find((a) => a.id === id);

export type ContentType =
  | 'guide'
  | 'comparison'
  | 'template'
  | 'rules'
  | 'essay'
  | 'shipped'
  | 'market';

/**
 * Which editorial gates a post is held to in tests/unit/blog.test.mjs.
 * `legacy` is the three posts written before the gates existed.
 */
export type PostFormat = 'search' | 'essay' | 'legacy';

/**
 * Text fields accept two inline forms: [label](href) and **strong**.
 * A href starting with / is an internal link.
 */
export type Block =
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'p'; text: string }
  | { type: 'quote'; text: string }
  | { type: 'ul'; items: string[] }
  | { type: 'ol'; items: string[] }
  | { type: 'table'; caption?: string; head: string[]; rows: string[][] }
  /** Where the first call to action sits. Placed after the first useful answer. */
  | { type: 'cta' };

export type CoverMotif = 'ledger' | 'board' | 'chart' | 'schedule' | 'stock' | 'receipt' | 'caret';

/** A small product-style panel drawn in markup. No image file behind it. */
export interface Cover {
  motif: CoverMotif;
  /** Panel title, written like a screen name: "Finance / Overview". */
  title: string;
  /** Three to five short rows. `value` and `tone` are optional. */
  rows: { label: string; value?: string; tone?: 'good' | 'warn' | 'muted' }[];
}

export interface Post {
  slug: string;
  locale: BlogLocale;
  title: string;
  metaTitle: string;
  description: string;
  excerpt: string;
  keyword: string;
  segment: SegmentId;
  type: ContentType;
  format: PostFormat;
  tags: string[];
  author: AuthorId;
  /** ISO date, YYYY-MM-DD. */
  date: string;
  dateModified?: string;
  readingMinutes: number;
  cover: Cover;
  blocks: Block[];
  faq?: { question: string; answer: string }[];
}

/** What a listing needs. Post bodies stay out of listing page props. */
export type PostSummary = Omit<Post, 'blocks' | 'faq'>;

/** A summary with its date already formatted, so server and client agree. */
export type ListedPost = PostSummary & { dateLabel: string };

export const toSummary = (post: Post): PostSummary => {
  const { blocks: _blocks, faq: _faq, ...summary } = post;
  return summary;
};

export const blogPath = (locale: BlogLocale, rest = ''): string =>
  `${locale === 'en' ? '' : `/${locale}`}/blog/${rest}`;

export const postPath = (post: Pick<Post, 'slug' | 'locale'>): string =>
  blogPath(post.locale, `${post.slug}/`);

