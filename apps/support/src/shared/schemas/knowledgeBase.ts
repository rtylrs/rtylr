import { z } from 'zod';

export const articleSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  title: z.string(),
  body: z.string().optional(),
  category: z.string().optional(),
  author: z.string().optional(),
  authorId: z.number().optional(),
  published: z.boolean().optional(),
  viewCount: z.number().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export const articleCategorySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  slug: z.string().optional(),
  articleCount: z.number().optional(),
  description: z.string().optional(),
});

export type Article = z.infer<typeof articleSchema>;
export type ArticleCategory = z.infer<typeof articleCategorySchema>;
