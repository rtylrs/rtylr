/**
 * Resolve an approval decision for a tool call. Looks at:
 *   1. The user's per-tool stored decision (cached forever within a chat).
 *   2. The auto-approve-read-only setting.
 *   3. Falls back to "pending" so the UI prompts the user.
 */
import { getApproval } from './approvals';

export type Decision = 'approved' | 'denied' | 'pending';

export interface DecisionInput {
  chatId: string;
  tool: string;
  scope: 'read' | 'write';
  autoApproveReadOnly: boolean;
}

export function resolveDecision(input: DecisionInput): Decision {
  const cached = getApproval(input.chatId, input.tool);
  if (cached) return cached.decision;
  if (input.scope === 'read' && input.autoApproveReadOnly) return 'approved';
  return 'pending';
}
