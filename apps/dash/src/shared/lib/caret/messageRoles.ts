import type { MessageRole } from '@voxire/caret';

/** Human-readable label for each message role. */
export function roleLabel(role: MessageRole): string {
  switch (role) {
    case 'user':
      return 'You';
    case 'assistant':
      return 'Caret';
    case 'tool':
      return 'Tool';
    case 'system':
      return 'System';
  }
}
