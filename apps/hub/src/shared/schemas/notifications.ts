import { z } from 'zod';

const appEnum = z.enum(['hub', 'pos', 'erp', 'hr', 'crm', 'finance', 'flow']);
const notifTypeEnum = z.enum(['info', 'success', 'warning', 'error']);

export const notificationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  title: z.string(),
  body: z.string(),
  app: appEnum.catch('hub'),
  type: notifTypeEnum.catch('info'),
  read: z.boolean(),
  createdAt: z.string(),
  href: z.string().optional(),
});

export type Notification = z.infer<typeof notificationSchema>;
