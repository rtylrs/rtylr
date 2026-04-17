import { z } from 'zod';

export const menuItemSchema = z
  .object({
    _id: z.string(),
    name: z.string(),
  })
  .passthrough();

export const categorySchema = z
  .object({
    _id: z.string(),
    name: z.string(),
    type: z.literal('category'),
    menuItems: z.array(menuItemSchema).default([]),
  })
  .passthrough();

export const bannerSchema = z.object({
  _id: z.string(),
  bannerImage: z.string(),
  menuId: z.string(),
});

export const menuSchema = z
  .object({
    _id: z.string(),
    items: z.array(z.any()).default([]),
  })
  .passthrough();

export type MenuItem = z.infer<typeof menuItemSchema>;
export type Category = z.infer<typeof categorySchema>;
export type Banner = z.infer<typeof bannerSchema>;
export type Menu = z.infer<typeof menuSchema>;
