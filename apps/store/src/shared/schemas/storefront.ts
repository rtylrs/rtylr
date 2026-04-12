import { z } from 'zod';

export const storefrontConfigSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  themeName: z.string(),
  logoUrl: z.string().optional(),
  faviconUrl: z.string().optional(),
  primaryColor: z.string(),
  secondaryColor: z.string(),
  customDomain: z.string().optional(),
  metaTitle: z.string().optional(),
  metaDescription: z.string().optional(),
  socialLinks: z.record(z.string(), z.string()).optional(),
  announcementBar: z.string().optional(),
  footerText: z.string().optional(),
});

export type StorefrontConfigParsed = z.infer<typeof storefrontConfigSchema>;
