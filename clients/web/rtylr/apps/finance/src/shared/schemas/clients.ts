import { z } from 'zod';

export const clientSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default(''),
  email: z.string().optional().nullable(),
  company: z.string().optional().nullable(),
});

export const clientsResponseSchema = z.union([
  z.array(clientSchema),
  z.object({ customers: z.array(clientSchema).default([]) }),
]);

export type ClientsResponse = z.infer<typeof clientsResponseSchema>;

export const parseClientsResponse = (response: unknown) => {
  const parsed = clientsResponseSchema.parse(response);
  return Array.isArray(parsed) ? parsed : parsed.customers;
};
