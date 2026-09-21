import { z } from 'zod';

const appEnum = z.enum(['dash', 'pos', 'erp', 'hr', 'crm', 'finance', 'flow']);
const notifTypeEnum = z.enum(['info', 'success', 'warning', 'error']);

export const notificationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  title: z.string(),
  body: z.string(),
  app: appEnum.catch('dash'),
  type: notifTypeEnum.catch('info'),
  // The backend writes `read` with `omitempty`, so an unread notification
  // arrives without the key. Absence means unread, not invalid.
  read: z.boolean().default(false),
  createdAt: z.string(),
  href: z.string().optional(),
});

export type Notification = z.infer<typeof notificationSchema>;
