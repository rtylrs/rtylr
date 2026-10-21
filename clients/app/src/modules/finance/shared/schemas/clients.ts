import { z } from 'zod';

export const clientSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    name: z.string().optional(),
    firstName: z.string().optional().nullable(),
    lastName: z.string().optional().nullable(),
    email: z.string().optional().nullable(),
    company: z.string().optional().nullable(),
  })
  .transform(({ firstName, lastName, ...client }) => ({
    ...client,
    name: client.name ?? [firstName, lastName].filter(Boolean).join(' '),
  }));

export const clientsResponseSchema = z.union([
  z.array(clientSchema),
  z.object({ customers: z.array(clientSchema).default([]) }),
]);

export const parseClientsResponse = (response: unknown) => {
  const parsed = clientsResponseSchema.parse(response);
  return Array.isArray(parsed) ? parsed : parsed.customers;
};
