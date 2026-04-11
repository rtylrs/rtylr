import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { bookService } from '../services/bookService';

export function useBookSettings() {
  return useQuery({
    queryKey: ['book-settings'],
    queryFn: () => unwrap(bookService.getSettings()),
  });
}

export function useUpdateBookSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => unwrap(bookService.updateSettings(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['book-settings'] }),
  });
}
