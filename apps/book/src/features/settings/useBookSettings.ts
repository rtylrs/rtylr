import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type { UpdateBookSettingsRequest } from '@voxire/contracts/book/v1/settings';
import { bookService } from '../../shared/services/bookService';

export function useBookSettings() {
  return useQuery({
    queryKey: ['book-settings'],
    queryFn: () => unwrap(bookService.getSettings()),
  });
}

export function useUpdateBookSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateBookSettingsRequest) => unwrap(bookService.updateSettings(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['book-settings'] }),
  });
}
