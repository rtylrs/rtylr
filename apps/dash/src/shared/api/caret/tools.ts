import { createOneKitClient } from '@voxire/api';
import { CaretServiceClient } from '@voxire/voxie/hub/caret/v1/service_client';

const caretClient = createOneKitClient(CaretServiceClient, 'dash');

export interface ToolDef {
  name: string;
  label: string;
  description: string;
  category: string;
  requiresApproval: boolean;
  scope: 'read' | 'write';
}

/** Fetch the tool registry, optionally scoped to a mode. */
export async function listTools(mode?: string): Promise<ToolDef[]> {
  const res = await caretClient.listCaretTools({ mode: mode ?? '' });
  return res.tools.map((t) => ({
    name: t.name,
    label: t.label,
    description: t.description,
    category: t.category,
    requiresApproval: t.requiresApproval,
    scope: t.scope === 'write' ? 'write' : 'read',
  }));
}
