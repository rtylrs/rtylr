import postsData from './posts.json';

export type Block =
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'p'; text: string }
  | { type: 'quote'; text: string }
  | { type: 'ul'; items: string[] };

export interface Post {
  slug: string;
  title: string;
  metaTitle: string;
  description: string;
  excerpt: string;
  keyword: string;
  category: string;
  date: string;
  dateDisplay: string;
  readingMinutes: number;
  blocks: Block[];
}

const posts = postsData as unknown as Post[];

export function getAllPosts(): Post[] {
  return [...posts].sort((a, b) => (a.date < b.date ? 1 : -1));
}

export function getPostBySlug(slug: string): Post | undefined {
  return posts.find((p) => p.slug === slug);
}
