import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type { UpdateBookSettingsRequest } from '@voxire/contracts/book/v1/settings';
import { bookService } from '../../shared/services/bookService';
import { bookSettingsSchema } from '../../shared/schemas/settings';

export function useBookSettings() {
  return useQuery({
    queryKey: ['book-settings'],
    queryFn: async () => {
      const res = await unwrap(bookService.getSettings());
      return bookSettingsSchema.parse(res);
    },
  });
}

export function useUpdateBookSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateBookSettingsRequest) => unwrap(bookService.updateSettings(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['book-settings'] }),
  });
}
