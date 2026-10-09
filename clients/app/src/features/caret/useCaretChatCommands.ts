import { useCaretSessions } from '@voxire/caret';
import { resolveApiBase } from '@voxire/config';
import { MessageSquare } from 'lucide-react';
import type { StoredUser } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import type { PortalCommand } from '../../shell/command-palette/commands';

export function useCaretChatCommands(user: StoredUser | null, enabled: boolean) {
  const allowed = enabled && canAccessPortalPath(user, '/caret');
  const sessions = useCaretSessions({
    userId: Number(user?.id) || 0,
    apiBaseUrl: resolveApiBase(),
    limit: 100,
    enabled: allowed,
  });
  const commands: PortalCommand[] = allowed
    ? sessions.sessions.map((chat) => ({
        id: `caret-chat-${chat.id}`,
        label: chat.title || chat.id,
        group: 'chats',
        icon: MessageSquare,
        destination: { openCaretChat: chat.id, title: chat.title, projectId: chat.projectId },
      }))
    : [];
  return {
    commands,
    loading: allowed && sessions.isLoading,
    failed: allowed && Boolean(sessions.error),
  };
}
