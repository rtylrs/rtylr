import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { safe } from '../../lib/api';

const toolDefSchema = z.object({
  name: z.string(),
  description: z.string(),
  module: z.string(),
  scope: z.enum(['read', 'write']),
  requiresApproval: z.boolean().optional().default(false),
});

const toolsResponseSchema = z.object({
  tools: z.array(toolDefSchema),
});

export type ToolDef = z.infer<typeof toolDefSchema>;

/** Fetch the tool registry, optionally scoped to a mode. */
export async function listTools(mode?: string): Promise<ToolDef[]> {
  const res = await unwrap(safe.post('/api/hub/caret/v1/tools', { mode: mode ?? '' }));
  const parsed = toolsResponseSchema.safeParse(res);
  if (parsed.success) return parsed.data.tools;
  return [];
}
