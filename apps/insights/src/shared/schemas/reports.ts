import { z } from 'zod';

// reportTemplateSchema validates GET /insights/reports/templates
// (insights ReportTemplatePayload). The backend returns a static catalog of
// templates where every entry carries a non-pointer id, name, and description
// — all three are always present, so they are required here. The backend does
// NOT send category/params/data; those fields were speculative and have been
// removed so the schema fails loudly if the contract drifts.
export const reportTemplateSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
});

export type ReportTemplate = z.infer<typeof reportTemplateSchema>;
