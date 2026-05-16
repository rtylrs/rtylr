import { useQuery } from '@tanstack/react-query';
import { getChatMessages, type StoredMessage } from '../../api/caret';

/** Load full message history for one chat. */
export const useCaretMessages = (chatId: string | null) =>
  useQuery<StoredMessage[]>({
    queryKey: ['caret', 'messages', chatId],
    queryFn: () => (chatId ? getChatMessages(chatId) : Promise.resolve([])),
    enabled: !!chatId,
    staleTime: 60_000,
  });
